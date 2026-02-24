package daemon

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
)

// WatchConfig configures file watching behavior.
type WatchConfig struct {
	// Paths are the directories or files to watch.
	Paths []string `json:"paths,omitempty"`

	// Patterns are glob patterns to include (e.g., "*.go", "**/*.ts").
	Patterns []string `json:"patterns,omitempty"`

	// Exclude are glob patterns to exclude (e.g., "*_test.go", "node_modules/**").
	Exclude []string `json:"exclude,omitempty"`

	// Events specifies which event types to trigger on.
	// Default: ["create", "modify", "delete"]
	Events []string `json:"events,omitempty"`

	// Recursive enables watching subdirectories. Default: true.
	Recursive bool `json:"recursive,omitempty"`

	// Debounce is the time to wait for burst of changes.
	Debounce time.Duration `json:"debounce,omitempty"`

	// Singleton if true, only one run at a time.
	Singleton bool `json:"singleton,omitempty"`
}

// WatchEvent represents a file change event.
type WatchEvent struct {
	// Path is the full path to the changed file.
	Path string

	// EventType is one of: create, modify, delete, rename.
	EventType string

	// WatchPath is the root path being watched.
	WatchPath string
}

// FileWatcher provides debounced, filtered file watching.
type FileWatcher struct {
	config   WatchConfig
	watcher  *fsnotify.Watcher
	callback func([]WatchEvent)

	mu             sync.Mutex
	pendingEvents  []WatchEvent
	debounceTimer  *time.Timer
	running        bool
	runningJob     bool // for singleton mode
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

// NewFileWatcher creates a new file watcher with the given configuration.
func NewFileWatcher(cfg WatchConfig, callback func([]WatchEvent)) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Debounce == 0 {
		cfg.Debounce = 100 * time.Millisecond
	}
	if len(cfg.Events) == 0 {
		cfg.Events = []string{"create", "modify", "delete"}
	}

	return &FileWatcher{
		config:   cfg,
		watcher:  watcher,
		callback: callback,
		stopCh:   make(chan struct{}),
	}, nil
}

// Start begins watching for file changes.
func (fw *FileWatcher) Start() error {
	fw.mu.Lock()
	if fw.running {
		fw.mu.Unlock()
		return nil
	}
	fw.running = true
	fw.mu.Unlock()

	// Add all paths
	for _, path := range fw.config.Paths {
		// Expand home directory
		if len(path) > 0 && path[0] == '~' {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, path[1:])
		}

		// Expand to absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		if fw.config.Recursive {
			if err := fw.addRecursive(absPath); err != nil {
				return err
			}
		} else {
			if err := fw.watcher.Add(absPath); err != nil {
				return err
			}
		}
	}

	// Start watch loop
	fw.wg.Add(1)
	go fw.watchLoop()

	return nil
}

// Stop stops watching for file changes.
func (fw *FileWatcher) Stop() {
	fw.mu.Lock()
	if !fw.running {
		fw.mu.Unlock()
		return
	}
	fw.running = false
	close(fw.stopCh)
	fw.mu.Unlock()

	fw.watcher.Close()
	fw.wg.Wait()
}

// addRecursive adds a directory and all subdirectories to the watcher.
func (fw *FileWatcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if d.IsDir() {
			// Skip excluded directories
			if fw.matchesExclude(path) {
				return filepath.SkipDir
			}
			return fw.watcher.Add(path)
		}
		return nil
	})
}

func (fw *FileWatcher) watchLoop() {
	defer fw.wg.Done()

	for {
		select {
		case <-fw.stopCh:
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)
		case _, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	// Convert fsnotify event to our event type
	eventType := fw.getEventType(event.Op)
	if eventType == "" {
		return
	}

	// Check if event type is in our filter
	if !fw.matchesEventType(eventType) {
		return
	}

	// Check if file matches patterns
	if !fw.matchesPatterns(event.Name) {
		return
	}

	// Check if file is excluded
	if fw.matchesExclude(event.Name) {
		return
	}

	// Get watch path (find which config path this belongs to)
	watchPath := fw.findWatchPath(event.Name)

	watchEvent := WatchEvent{
		Path:      event.Name,
		EventType: eventType,
		WatchPath: watchPath,
	}

	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Add to pending events
	fw.pendingEvents = append(fw.pendingEvents, watchEvent)

	// Reset debounce timer
	if fw.debounceTimer != nil {
		fw.debounceTimer.Stop()
	}

	fw.debounceTimer = time.AfterFunc(fw.config.Debounce, func() {
		fw.triggerCallback()
	})
}

func (fw *FileWatcher) triggerCallback() {
	fw.mu.Lock()
	events := fw.pendingEvents
	fw.pendingEvents = nil

	// Singleton mode check
	if fw.config.Singleton && fw.runningJob {
		fw.mu.Unlock()
		return
	}
	if fw.config.Singleton {
		fw.runningJob = true
	}
	fw.mu.Unlock()

	if len(events) == 0 {
		return
	}

	// Deduplicate events for same file
	events = fw.deduplicateEvents(events)

	if fw.callback != nil {
		fw.callback(events)
	}

	if fw.config.Singleton {
		fw.mu.Lock()
		fw.runningJob = false
		fw.mu.Unlock()
	}
}

func (fw *FileWatcher) getEventType(op fsnotify.Op) string {
	if op&fsnotify.Create != 0 {
		return "create"
	}
	if op&fsnotify.Write != 0 {
		return "modify"
	}
	if op&fsnotify.Remove != 0 {
		return "delete"
	}
	if op&fsnotify.Rename != 0 {
		return "rename"
	}
	return ""
}

func (fw *FileWatcher) matchesEventType(eventType string) bool {
	for _, et := range fw.config.Events {
		if et == eventType {
			return true
		}
	}
	return false
}

func (fw *FileWatcher) matchesPatterns(path string) bool {
	if len(fw.config.Patterns) == 0 {
		return true // No patterns = match all
	}

	for _, pattern := range fw.config.Patterns {
		// Use doublestar for glob matching
		matched, _ := doublestar.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
		// Also try matching against full path for ** patterns
		matched, _ = doublestar.Match(pattern, path)
		if matched {
			return true
		}
	}
	return false
}

func (fw *FileWatcher) matchesExclude(path string) bool {
	for _, pattern := range fw.config.Exclude {
		// Try matching against filename
		matched, _ := doublestar.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
		// Try matching against full path
		matched, _ = doublestar.Match(pattern, path)
		if matched {
			return true
		}
	}
	return false
}

func (fw *FileWatcher) findWatchPath(eventPath string) string {
	for _, path := range fw.config.Paths {
		// Expand home directory
		expanded := path
		if len(expanded) > 0 && expanded[0] == '~' {
			home, _ := os.UserHomeDir()
			expanded = filepath.Join(home, expanded[1:])
		}
		abs, _ := filepath.Abs(expanded)
		if hasPathPrefix(eventPath, abs) {
			return path
		}
	}
	return ""
}

func (fw *FileWatcher) deduplicateEvents(events []WatchEvent) []WatchEvent {
	seen := make(map[string]bool)
	result := make([]WatchEvent, 0, len(events))

	for _, e := range events {
		key := e.Path + ":" + e.EventType
		if !seen[key] {
			seen[key] = true
			result = append(result, e)
		}
	}
	return result
}

// hasPathPrefix checks if path has the given prefix.
func hasPathPrefix(path, prefix string) bool {
	if len(path) < len(prefix) {
		return false
	}
	if path[:len(prefix)] != prefix {
		return false
	}
	if len(path) == len(prefix) {
		return true
	}
	return path[len(prefix)] == filepath.Separator
}

// CollectChangedFiles returns all file paths from a slice of events.
func CollectChangedFiles(events []WatchEvent) []string {
	files := make([]string, len(events))
	for i, e := range events {
		files[i] = e.Path
	}
	return files
}
