package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	"github.com/alexcabrera/ayo/internal/paths"
)

// TriggerYAML represents a trigger configuration in YAML format.
type TriggerYAML struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"` // cron, interval, watch, once, daily, weekly, monthly
	Agent   string `yaml:"agent"`
	Enabled *bool  `yaml:"enabled,omitempty"` // Default: true

	Schedule TriggerScheduleYAML `yaml:"schedule,omitempty"`
	Options  TriggerOptionsYAML  `yaml:"options,omitempty"`
	Prompt   string              `yaml:"prompt,omitempty"`
	Env      map[string]string   `yaml:"env,omitempty"`
}

// TriggerScheduleYAML contains type-specific schedule configuration.
type TriggerScheduleYAML struct {
	// For cron type
	Cron string `yaml:"cron,omitempty"`

	// For interval type
	Every string `yaml:"every,omitempty"`

	// For once type
	At string `yaml:"at,omitempty"`

	// For daily/weekly/monthly types
	Times []string `yaml:"times,omitempty"`
	Days  []string `yaml:"days,omitempty"` // For weekly

	// For watch type
	Path      string   `yaml:"path,omitempty"`
	Pattern   string   `yaml:"pattern,omitempty"`
	Patterns  []string `yaml:"patterns,omitempty"`
	Events    []string `yaml:"events,omitempty"` // create, modify, delete
	Recursive bool     `yaml:"recursive,omitempty"`
}

// TriggerOptionsYAML contains trigger execution options.
type TriggerOptionsYAML struct {
	Singleton bool   `yaml:"singleton,omitempty"` // Prevent overlapping runs
	Timeout   string `yaml:"timeout,omitempty"`   // Max execution time
	Retry     int    `yaml:"retry,omitempty"`     // Retry on failure
}

// IsEnabled returns true if the trigger is enabled.
func (t *TriggerYAML) IsEnabled() bool {
	if t.Enabled == nil {
		return true // Default to enabled
	}
	return *t.Enabled
}

// TriggerLoader loads and watches trigger YAML configurations.
type TriggerLoader struct {
	mu       sync.RWMutex
	triggers map[string]*TriggerYAML // filename -> trigger
	engine   *TriggerEngine
	watcher  *fsnotify.Watcher
	stopCh   chan struct{}
	logger   Logger
}

// Logger is a leveled logger interface for trigger loading events.
// NOTE: This differs from audit.Logger intentionally. audit.Logger is for
// structured audit trails with Query support, while this provides standard
// Info/Warn/Error leveled logging for operational diagnostics.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// NewTriggerLoader creates a new trigger loader.
func NewTriggerLoader(engine *TriggerEngine, logger Logger) *TriggerLoader {
	return &TriggerLoader{
		triggers: make(map[string]*TriggerYAML),
		engine:   engine,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// LoadAll loads all trigger YAML files from the config directory.
func (tl *TriggerLoader) LoadAll(ctx context.Context) error {
	triggersDir := paths.TriggersConfigDir()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(triggersDir, 0o755); err != nil {
		return fmt.Errorf("create triggers directory: %w", err)
	}

	// Find all YAML files
	entries, err := os.ReadDir(triggersDir)
	if err != nil {
		return fmt.Errorf("read triggers directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := filepath.Join(triggersDir, name)
		if err := tl.loadTrigger(path); err != nil {
			tl.logger.Warn("failed to load trigger", "file", name, "error", err)
		}
	}

	return nil
}

// StartWatching starts watching the triggers directory for changes.
func (tl *TriggerLoader) StartWatching(ctx context.Context) error {
	triggersDir := paths.TriggersConfigDir()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	if err := watcher.Add(triggersDir); err != nil {
		watcher.Close()
		return fmt.Errorf("watch triggers directory: %w", err)
	}

	tl.watcher = watcher

	go tl.watchLoop()

	tl.logger.Info("watching triggers directory", "path", triggersDir)
	return nil
}

// Stop stops the trigger loader.
func (tl *TriggerLoader) Stop() {
	close(tl.stopCh)
	if tl.watcher != nil {
		tl.watcher.Close()
	}
}

func (tl *TriggerLoader) watchLoop() {
	for {
		select {
		case <-tl.stopCh:
			return
		case event, ok := <-tl.watcher.Events:
			if !ok {
				return
			}
			tl.handleFileEvent(event)
		case err, ok := <-tl.watcher.Errors:
			if !ok {
				return
			}
			tl.logger.Warn("watcher error", "error", err)
		}
	}
}

func (tl *TriggerLoader) handleFileEvent(event fsnotify.Event) {
	name := filepath.Base(event.Name)
	if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
		return
	}

	switch {
	case event.Op&fsnotify.Create != 0, event.Op&fsnotify.Write != 0:
		if err := tl.loadTrigger(event.Name); err != nil {
			tl.logger.Warn("failed to load trigger", "file", name, "error", err)
		} else {
			tl.logger.Info("loaded trigger", "file", name)
		}
	case event.Op&fsnotify.Remove != 0, event.Op&fsnotify.Rename != 0:
		tl.unloadTrigger(event.Name)
		tl.logger.Info("unloaded trigger", "file", name)
	}
}

func (tl *TriggerLoader) loadTrigger(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var cfg TriggerYAML
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse YAML: %w", err)
	}

	// Validate
	if err := tl.validateTrigger(&cfg); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	// Convert to engine trigger
	trigger, err := tl.toEngineTrigger(&cfg, path)
	if err != nil {
		return fmt.Errorf("convert trigger: %w", err)
	}

	// Unload existing trigger with same source
	tl.unloadTrigger(path)

	// Register with engine
	if cfg.IsEnabled() {
		if err := tl.engine.Register(trigger); err != nil {
			return fmt.Errorf("register trigger: %w", err)
		}
	}

	tl.mu.Lock()
	tl.triggers[path] = &cfg
	tl.mu.Unlock()

	return nil
}

func (tl *TriggerLoader) unloadTrigger(path string) {
	tl.mu.Lock()
	cfg, exists := tl.triggers[path]
	if exists {
		delete(tl.triggers, path)
	}
	tl.mu.Unlock()

	if exists && cfg != nil {
		triggerID := tl.triggerID(cfg.Name, path)
		tl.engine.Unregister(triggerID)
	}
}

func (tl *TriggerLoader) validateTrigger(cfg *TriggerYAML) error {
	if cfg.Name == "" {
		return fmt.Errorf("name is required")
	}
	if cfg.Type == "" {
		return fmt.Errorf("type is required")
	}
	if cfg.Agent == "" {
		return fmt.Errorf("agent is required")
	}

	// Type-specific validation
	switch cfg.Type {
	case "cron":
		if cfg.Schedule.Cron == "" {
			return fmt.Errorf("schedule.cron is required for cron type")
		}
	case "interval":
		if cfg.Schedule.Every == "" {
			return fmt.Errorf("schedule.every is required for interval type")
		}
	case "watch":
		if cfg.Schedule.Path == "" {
			return fmt.Errorf("schedule.path is required for watch type")
		}
	case "once":
		if cfg.Schedule.At == "" {
			return fmt.Errorf("schedule.at is required for once type")
		}
	case "daily", "weekly", "monthly":
		// Times is optional, uses sensible defaults
	default:
		return fmt.Errorf("unknown trigger type: %s", cfg.Type)
	}

	return nil
}

func (tl *TriggerLoader) toEngineTrigger(cfg *TriggerYAML, source string) (*Trigger, error) {
	triggerID := tl.triggerID(cfg.Name, source)

	var triggerType TriggerType
	var triggerConfig TriggerConfig

	switch cfg.Type {
	case "cron":
		triggerType = TriggerTypeCron
		triggerConfig.Schedule = cfg.Schedule.Cron
	case "interval":
		triggerType = TriggerTypeCron
		// Convert interval to cron-like (gocron handles this)
		triggerConfig.Schedule = "@every " + cfg.Schedule.Every
	case "watch":
		triggerType = TriggerTypeWatch
		triggerConfig.Path = cfg.Schedule.Path
		triggerConfig.Patterns = cfg.Schedule.Patterns
		if cfg.Schedule.Pattern != "" && len(triggerConfig.Patterns) == 0 {
			triggerConfig.Patterns = []string{cfg.Schedule.Pattern}
		}
		triggerConfig.Events = cfg.Schedule.Events
		triggerConfig.Recursive = cfg.Schedule.Recursive
	default:
		// For other types (once, daily, weekly, monthly), convert to cron
		triggerType = TriggerTypeCron
		triggerConfig.Schedule = tl.convertToCron(cfg)
	}

	return &Trigger{
		ID:      triggerID,
		Type:    triggerType,
		Agent:   cfg.Agent,
		Config:  triggerConfig,
		Prompt:  cfg.Prompt,
		Source:  source,
		Enabled: cfg.IsEnabled(),
	}, nil
}

func (tl *TriggerLoader) triggerID(name, source string) string {
	// Use filename as part of ID to ensure uniqueness
	base := filepath.Base(source)
	base = strings.TrimSuffix(base, ".yaml")
	base = strings.TrimSuffix(base, ".yml")
	return fmt.Sprintf("%s_%s", base, name)
}

func (tl *TriggerLoader) convertToCron(cfg *TriggerYAML) string {
	// Basic conversion - can be enhanced
	switch cfg.Type {
	case "daily":
		if len(cfg.Schedule.Times) > 0 {
			// Parse first time, e.g., "09:00"
			parts := strings.Split(cfg.Schedule.Times[0], ":")
			if len(parts) == 2 {
				return fmt.Sprintf("%s %s * * *", parts[1], parts[0])
			}
		}
		return "0 9 * * *" // Default: 9 AM
	case "weekly":
		if len(cfg.Schedule.Days) > 0 {
			day := tl.dayToNumber(cfg.Schedule.Days[0])
			return fmt.Sprintf("0 9 * * %d", day)
		}
		return "0 9 * * 1" // Default: Monday 9 AM
	case "monthly":
		return "0 9 1 * *" // Default: 1st of month, 9 AM
	default:
		return "0 * * * *" // Default: every hour
	}
}

func (tl *TriggerLoader) dayToNumber(day string) int {
	days := map[string]int{
		"sunday": 0, "sun": 0,
		"monday": 1, "mon": 1,
		"tuesday": 2, "tue": 2,
		"wednesday": 3, "wed": 3,
		"thursday": 4, "thu": 4,
		"friday": 5, "fri": 5,
		"saturday": 6, "sat": 6,
	}
	if n, ok := days[strings.ToLower(day)]; ok {
		return n
	}
	return 1 // Default to Monday
}

// ListLoaded returns all loaded triggers.
func (tl *TriggerLoader) ListLoaded() []*TriggerYAML {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	result := make([]*TriggerYAML, 0, len(tl.triggers))
	for _, t := range tl.triggers {
		result = append(result, t)
	}
	return result
}
