// Package flows provides trigger extraction from YAML flow files.
package flows

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ExtractedTrigger represents a trigger extracted from a flow file.
type ExtractedTrigger struct {
	ID                  string
	Type                FlowTriggerType
	FlowName            string
	FlowPath            string
	Schedule            string // for cron
	Path                string // for watch
	Patterns            []string
	Recursive           bool
	Events              []string
	Params              map[string]any
	RunsBeforePermanent int
	Enabled             bool
}

// TriggerLoader loads triggers from flow files and watches for changes.
type TriggerLoader struct {
	dirs           []string
	watcher        *fsnotify.Watcher
	logger         *slog.Logger
	mu             sync.RWMutex
	triggers       map[string]*ExtractedTrigger // trigger ID -> trigger
	flowTriggers   map[string][]string          // flow path -> trigger IDs
	stopCh         chan struct{}
	wg             sync.WaitGroup
	running        bool

	// Callback when triggers are updated
	OnTriggersChanged func(added, removed []*ExtractedTrigger)
}

// TriggerLoaderConfig configures the trigger loader.
type TriggerLoaderConfig struct {
	Dirs              []string
	Logger            *slog.Logger
	OnTriggersChanged func(added, removed []*ExtractedTrigger)
}

// NewTriggerLoader creates a new trigger loader.
func NewTriggerLoader(cfg TriggerLoaderConfig) *TriggerLoader {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &TriggerLoader{
		dirs:              cfg.Dirs,
		logger:            cfg.Logger,
		triggers:          make(map[string]*ExtractedTrigger),
		flowTriggers:      make(map[string][]string),
		stopCh:            make(chan struct{}),
		OnTriggersChanged: cfg.OnTriggersChanged,
	}
}

// Start begins loading triggers and watching for changes.
func (l *TriggerLoader) Start(ctx context.Context) error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return fmt.Errorf("trigger loader already running")
	}
	l.running = true
	l.mu.Unlock()

	// Create file watcher
	var err error
	l.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	// Initial load
	l.loadAllFlows()

	// Watch directories
	for _, dir := range l.dirs {
		if err := l.watchDir(dir); err != nil {
			l.logger.Warn("failed to watch directory", "dir", dir, "error", err)
		}
	}

	// Start watch loop
	l.wg.Add(1)
	go l.watchLoop()

	l.logger.Info("flow trigger loader started", "dirs", l.dirs)
	return nil
}

// Stop stops the trigger loader.
func (l *TriggerLoader) Stop(ctx context.Context) error {
	l.mu.Lock()
	if !l.running {
		l.mu.Unlock()
		return nil
	}
	l.running = false
	close(l.stopCh)
	l.mu.Unlock()

	if l.watcher != nil {
		l.watcher.Close()
	}

	l.wg.Wait()

	l.logger.Info("flow trigger loader stopped")
	return nil
}

// GetTriggers returns all loaded triggers.
func (l *TriggerLoader) GetTriggers() []*ExtractedTrigger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]*ExtractedTrigger, 0, len(l.triggers))
	for _, t := range l.triggers {
		result = append(result, t)
	}
	return result
}

// loadAllFlows scans all directories and loads triggers.
func (l *TriggerLoader) loadAllFlows() {
	var added []*ExtractedTrigger

	for _, dir := range l.dirs {
		triggers := l.loadFlowsFromDir(dir)
		added = append(added, triggers...)
	}

	if len(added) > 0 && l.OnTriggersChanged != nil {
		l.OnTriggersChanged(added, nil)
	}
}

// loadFlowsFromDir loads YAML flows from a directory and extracts triggers.
func (l *TriggerLoader) loadFlowsFromDir(dir string) []*ExtractedTrigger {
	var result []*ExtractedTrigger

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !IsYAMLFlowFile(entry.Name()) {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		triggers := l.loadFlowTriggers(path)
		result = append(result, triggers...)
	}

	return result
}

// loadFlowTriggers loads triggers from a single flow file.
func (l *TriggerLoader) loadFlowTriggers(path string) []*ExtractedTrigger {
	flow, err := ParseYAMLFlow(path)
	if err != nil {
		l.logger.Debug("failed to parse flow", "path", path, "error", err)
		return nil
	}

	if err := flow.Validate(); err != nil {
		l.logger.Debug("invalid flow", "path", path, "error", err)
		return nil
	}

	if len(flow.Triggers) == 0 {
		return nil
	}

	var result []*ExtractedTrigger
	var triggerIDs []string

	for _, ft := range flow.Triggers {
		trigger := &ExtractedTrigger{
			ID:       fmt.Sprintf("flow:%s:%s", flow.Name, ft.ID),
			FlowName: flow.Name,
			FlowPath: path,
			Enabled:  ft.IsEnabled(),
		}

		if ft.RunsBeforePermanent != nil {
			trigger.RunsBeforePermanent = *ft.RunsBeforePermanent
		}

		if ft.Params != nil {
			trigger.Params = ft.Params
		}

		switch ft.Type {
		case FlowTriggerTypeCron:
			trigger.Type = FlowTriggerTypeCron
			trigger.Schedule = ft.Schedule
		case FlowTriggerTypeWatch:
			trigger.Type = FlowTriggerTypeWatch
			trigger.Path = ft.Path
			trigger.Patterns = ft.Patterns
			trigger.Recursive = ft.Recursive
			trigger.Events = ft.Events
		default:
			continue
		}

		l.mu.Lock()
		l.triggers[trigger.ID] = trigger
		l.mu.Unlock()

		triggerIDs = append(triggerIDs, trigger.ID)
		result = append(result, trigger)
	}

	l.mu.Lock()
	l.flowTriggers[path] = triggerIDs
	l.mu.Unlock()

	return result
}

// reloadFlow reloads triggers from a flow file.
func (l *TriggerLoader) reloadFlow(path string) {
	l.mu.Lock()
	// Remove existing triggers for this flow
	var removed []*ExtractedTrigger
	if triggerIDs, ok := l.flowTriggers[path]; ok {
		for _, id := range triggerIDs {
			if t, exists := l.triggers[id]; exists {
				removed = append(removed, t)
				delete(l.triggers, id)
			}
		}
		delete(l.flowTriggers, path)
	}
	l.mu.Unlock()

	// Load new triggers
	added := l.loadFlowTriggers(path)

	// Notify callback
	if l.OnTriggersChanged != nil && (len(added) > 0 || len(removed) > 0) {
		l.OnTriggersChanged(added, removed)
	}
}

// removeFlow removes triggers for a deleted flow file.
func (l *TriggerLoader) removeFlow(path string) {
	l.mu.Lock()
	var removed []*ExtractedTrigger
	if triggerIDs, ok := l.flowTriggers[path]; ok {
		for _, id := range triggerIDs {
			if t, exists := l.triggers[id]; exists {
				removed = append(removed, t)
				delete(l.triggers, id)
			}
		}
		delete(l.flowTriggers, path)
	}
	l.mu.Unlock()

	if l.OnTriggersChanged != nil && len(removed) > 0 {
		l.OnTriggersChanged(nil, removed)
	}
}

// watchDir adds a directory to the watcher.
func (l *TriggerLoader) watchDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip
	}
	return l.watcher.Add(dir)
}

// watchLoop handles file system events.
func (l *TriggerLoader) watchLoop() {
	defer l.wg.Done()

	// Debounce timer for rapid changes
	var debounceTimer *time.Timer
	pendingPaths := make(map[string]bool)
	var pendingMu sync.Mutex

	processPath := func(path string) {
		if !IsYAMLFlowFile(path) {
			return
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			l.removeFlow(path)
		} else {
			l.reloadFlow(path)
		}
	}

	for {
		select {
		case <-l.stopCh:
			return

		case event, ok := <-l.watcher.Events:
			if !ok {
				return
			}

			// Only care about YAML files
			if !IsYAMLFlowFile(event.Name) {
				continue
			}

			// Debounce: accumulate changes for 100ms
			pendingMu.Lock()
			pendingPaths[event.Name] = true
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(100*time.Millisecond, func() {
				pendingMu.Lock()
				paths := make([]string, 0, len(pendingPaths))
				for p := range pendingPaths {
					paths = append(paths, p)
				}
				pendingPaths = make(map[string]bool)
				pendingMu.Unlock()

				for _, p := range paths {
					processPath(p)
				}
			})
			pendingMu.Unlock()

		case err, ok := <-l.watcher.Errors:
			if !ok {
				return
			}
			l.logger.Warn("watcher error", "error", err)
		}
	}
}
