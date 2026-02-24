package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
)

// TriggerType represents the type of trigger.
type TriggerType string

const (
	TriggerTypeCron  TriggerType = "cron"
	TriggerTypeWatch TriggerType = "watch"
)

// Trigger represents a trigger configuration.
type Trigger struct {
	ID       string         `json:"id"`
	Type     TriggerType    `json:"type"`
	Agent    string         `json:"agent"`
	Config   TriggerConfig  `json:"config"`
	Prompt   string         `json:"prompt,omitempty"`
	Source   string         `json:"source"` // Path to config that defined this trigger
	Enabled  bool           `json:"enabled"`
}

// TriggerConfig holds type-specific trigger configuration.
type TriggerConfig struct {
	// Cron configuration
	Schedule string `json:"schedule,omitempty"` // cron expression

	// Watch configuration
	Path      string   `json:"path,omitempty"`      // path to watch
	Patterns  []string `json:"patterns,omitempty"`  // glob patterns to match
	Recursive bool     `json:"recursive,omitempty"` // watch subdirectories
	Events    []string `json:"events,omitempty"`    // create, modify, delete
}

// TriggerEvent represents a triggered event.
type TriggerEvent struct {
	TriggerID string            `json:"trigger_id"`
	FiredAt   time.Time         `json:"fired_at"`
	Context   map[string]any    `json:"context,omitempty"`
	Agent     string            `json:"agent"`
	Prompt    string            `json:"prompt,omitempty"`
}

// TriggerCallback is called when a trigger fires.
type TriggerCallback func(event TriggerEvent)

// TriggerEngine manages all triggers.
type TriggerEngine struct {
	mu        sync.RWMutex
	triggers  map[string]*Trigger
	cronJobs  map[string]cron.EntryID
	cron      *cron.Cron
	watcher   *fsnotify.Watcher
	watchDirs map[string][]string // dir -> trigger IDs
	callback  TriggerCallback
	logger    *slog.Logger
	stopCh    chan struct{}
	wg        sync.WaitGroup
	running   bool
}

// TriggerEngineConfig configures the trigger engine.
type TriggerEngineConfig struct {
	Logger   *slog.Logger
	Callback TriggerCallback
}

// NewTriggerEngine creates a new trigger engine.
func NewTriggerEngine(cfg TriggerEngineConfig) *TriggerEngine {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &TriggerEngine{
		triggers:  make(map[string]*Trigger),
		cronJobs:  make(map[string]cron.EntryID),
		watchDirs: make(map[string][]string),
		callback:  cfg.Callback,
		logger:    cfg.Logger,
		stopCh:    make(chan struct{}),
	}
}

// Start starts the trigger engine.
func (e *TriggerEngine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("trigger engine already running")
	}
	e.running = true
	e.mu.Unlock()

	// Initialize cron scheduler
	e.cron = cron.New(cron.WithSeconds())
	e.cron.Start()

	// Initialize file watcher
	var err error
	e.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		e.cron.Stop()
		return fmt.Errorf("create file watcher: %w", err)
	}

	// Start watch event loop
	e.wg.Add(1)
	go e.watchLoop()

	e.logger.Info("trigger engine started")
	return nil
}

// Stop stops the trigger engine.
func (e *TriggerEngine) Stop(ctx context.Context) error {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return nil
	}
	e.running = false
	close(e.stopCh)
	e.mu.Unlock()

	// Stop cron scheduler
	if e.cron != nil {
		cronCtx := e.cron.Stop()
		<-cronCtx.Done()
	}

	// Stop file watcher
	if e.watcher != nil {
		e.watcher.Close()
	}

	// Wait for goroutines
	e.wg.Wait()

	e.logger.Info("trigger engine stopped")
	return nil
}

// Register registers a new trigger.
func (e *TriggerEngine) Register(trigger *Trigger) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !trigger.Enabled {
		return nil
	}

	// Remove existing trigger with same ID
	if _, exists := e.triggers[trigger.ID]; exists {
		e.unregisterLocked(trigger.ID)
	}

	e.triggers[trigger.ID] = trigger

	switch trigger.Type {
	case TriggerTypeCron:
		return e.registerCronTrigger(trigger)
	case TriggerTypeWatch:
		return e.registerWatchTrigger(trigger)
	default:
		return fmt.Errorf("unknown trigger type: %s", trigger.Type)
	}
}

// Unregister removes a trigger.
func (e *TriggerEngine) Unregister(triggerID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.unregisterLocked(triggerID)
}

func (e *TriggerEngine) unregisterLocked(triggerID string) error {
	trigger, exists := e.triggers[triggerID]
	if !exists {
		return nil
	}

	switch trigger.Type {
	case TriggerTypeCron:
		if entryID, ok := e.cronJobs[triggerID]; ok {
			e.cron.Remove(entryID)
			delete(e.cronJobs, triggerID)
		}
	case TriggerTypeWatch:
		// Remove from watchDirs
		for dir, ids := range e.watchDirs {
			newIDs := make([]string, 0, len(ids))
			for _, id := range ids {
				if id != triggerID {
					newIDs = append(newIDs, id)
				}
			}
			if len(newIDs) == 0 {
				e.watcher.Remove(dir)
				delete(e.watchDirs, dir)
			} else {
				e.watchDirs[dir] = newIDs
			}
		}
	}

	delete(e.triggers, triggerID)
	return nil
}

// List returns all registered triggers.
func (e *TriggerEngine) List() []*Trigger {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*Trigger, 0, len(e.triggers))
	for _, t := range e.triggers {
		result = append(result, t)
	}
	return result
}

// Get returns a trigger by ID.
func (e *TriggerEngine) Get(id string) (*Trigger, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if t, ok := e.triggers[id]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("trigger not found: %s", id)
}

// SetEnabled enables or disables a trigger.
func (e *TriggerEngine) SetEnabled(id string, enabled bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	trigger, exists := e.triggers[id]
	if !exists {
		return fmt.Errorf("trigger not found: %s", id)
	}

	if trigger.Enabled == enabled {
		return nil // No change needed
	}

	trigger.Enabled = enabled

	if enabled {
		// Re-register the trigger to activate it
		switch trigger.Type {
		case TriggerTypeCron:
			return e.registerCronTrigger(trigger)
		case TriggerTypeWatch:
			return e.registerWatchTrigger(trigger)
		}
	} else {
		// Unregister without removing from map
		switch trigger.Type {
		case TriggerTypeCron:
			if entryID, ok := e.cronJobs[id]; ok {
				e.cron.Remove(entryID)
				delete(e.cronJobs, id)
			}
		case TriggerTypeWatch:
			for dir, ids := range e.watchDirs {
				newIDs := make([]string, 0, len(ids))
				for _, tid := range ids {
					if tid != id {
						newIDs = append(newIDs, tid)
					}
				}
				if len(newIDs) == 0 {
					e.watcher.Remove(dir)
					delete(e.watchDirs, dir)
				} else {
					e.watchDirs[dir] = newIDs
				}
			}
		}
	}

	e.logger.Info("trigger enabled state changed", "id", id, "enabled", enabled)
	return nil
}

// normalizeCronSchedule converts a 5-field standard cron expression to
// a 6-field expression (with seconds) expected by robfig/cron.
// If already 6 fields, returns unchanged.
func normalizeCronSchedule(schedule string) string {
	fields := strings.Fields(schedule)
	if len(fields) == 5 {
		// Standard cron (minute hour day month weekday) -> add "0" for seconds
		return "0 " + schedule
	}
	return schedule
}

func (e *TriggerEngine) registerCronTrigger(trigger *Trigger) error {
	if trigger.Config.Schedule == "" {
		return fmt.Errorf("cron trigger requires schedule")
	}

	// Normalize schedule to 6-field cron (with seconds)
	schedule := normalizeCronSchedule(trigger.Config.Schedule)

	entryID, err := e.cron.AddFunc(schedule, func() {
		e.fireTrigger(trigger.ID, map[string]any{
			"scheduled_at": time.Now(),
		})
	})
	if err != nil {
		return fmt.Errorf("add cron job: %w", err)
	}

	e.cronJobs[trigger.ID] = entryID
	e.logger.Info("registered cron trigger", "id", trigger.ID, "schedule", schedule, "agent", trigger.Agent)
	return nil
}

func (e *TriggerEngine) registerWatchTrigger(trigger *Trigger) error {
	if trigger.Config.Path == "" {
		return fmt.Errorf("watch trigger requires path")
	}

	// Expand path
	path := trigger.Config.Path
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}

	// Ensure path exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("watch path: %w", err)
	}

	if !info.IsDir() {
		path = filepath.Dir(path)
	}

	// Add to watcher
	if err := e.watcher.Add(path); err != nil {
		return fmt.Errorf("add watch: %w", err)
	}

	e.watchDirs[path] = append(e.watchDirs[path], trigger.ID)
	e.logger.Info("registered watch trigger", "id", trigger.ID, "path", path, "agent", trigger.Agent)
	return nil
}

func (e *TriggerEngine) watchLoop() {
	defer e.wg.Done()

	for {
		select {
		case <-e.stopCh:
			return
		case event, ok := <-e.watcher.Events:
			if !ok {
				return
			}
			e.handleWatchEvent(event)
		case err, ok := <-e.watcher.Errors:
			if !ok {
				return
			}
			e.logger.Warn("watch error", "error", err)
		}
	}
}

func (e *TriggerEngine) handleWatchEvent(event fsnotify.Event) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	dir := filepath.Dir(event.Name)
	triggerIDs, ok := e.watchDirs[dir]
	if !ok {
		e.logger.Debug("watch event for unregistered directory", "dir", dir, "file", event.Name, "op", event.Op.String())
		return
	}

	e.logger.Debug("handling watch event", "dir", dir, "file", event.Name, "op", event.Op.String(), "triggers", len(triggerIDs))

	for _, triggerID := range triggerIDs {
		trigger, ok := e.triggers[triggerID]
		if !ok {
			continue
		}

		// Check if event matches patterns
		if !e.matchesWatchEvent(trigger, event) {
			e.logger.Debug("watch event did not match trigger", "trigger", triggerID, "file", event.Name)
			continue
		}

		e.fireWatchTrigger(triggerID, event)
	}
}

func (e *TriggerEngine) matchesWatchEvent(trigger *Trigger, event fsnotify.Event) bool {
	// Check event type
	if len(trigger.Config.Events) > 0 {
		eventType := ""
		if event.Op&fsnotify.Create != 0 {
			eventType = "create"
		} else if event.Op&fsnotify.Write != 0 {
			eventType = "modify"
		} else if event.Op&fsnotify.Remove != 0 {
			eventType = "delete"
		} else if event.Op&fsnotify.Rename != 0 {
			eventType = "rename"
		}

		found := false
		for _, e := range trigger.Config.Events {
			if e == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check patterns
	if len(trigger.Config.Patterns) > 0 {
		filename := filepath.Base(event.Name)
		matched := false
		for _, pattern := range trigger.Config.Patterns {
			if ok, _ := filepath.Match(pattern, filename); ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func (e *TriggerEngine) fireWatchTrigger(triggerID string, event fsnotify.Event) {
	eventType := "unknown"
	if event.Op&fsnotify.Create != 0 {
		eventType = "create"
	} else if event.Op&fsnotify.Write != 0 {
		eventType = "modify"
	} else if event.Op&fsnotify.Remove != 0 {
		eventType = "delete"
	} else if event.Op&fsnotify.Rename != 0 {
		eventType = "rename"
	}

	e.fireTrigger(triggerID, map[string]any{
		"file_path":   event.Name,
		"file_name":   filepath.Base(event.Name),
		"event_type":  eventType,
		"triggered_at": time.Now(),
	})
}

func (e *TriggerEngine) fireTrigger(triggerID string, ctx map[string]any) {
	trigger, ok := e.triggers[triggerID]
	if !ok {
		return
	}

	event := TriggerEvent{
		TriggerID: triggerID,
		FiredAt:   time.Now(),
		Context:   ctx,
		Agent:     trigger.Agent,
		Prompt:    trigger.Prompt,
	}

	e.logger.Info("trigger fired", "id", triggerID, "agent", trigger.Agent)

	if e.callback != nil {
		go e.callback(event)
	}
}

// GenerateTriggerID generates a unique trigger ID.
func GenerateTriggerID() string {
	return fmt.Sprintf("trig_%d", time.Now().UnixNano())
}
