package daemon

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestFileWatcher_MatchesPatterns(t *testing.T) {
	fw := &FileWatcher{
		config: WatchConfig{
			Patterns: []string{"*.go", "*.ts"},
		},
	}

	tests := []struct {
		path    string
		matches bool
	}{
		{"/path/to/file.go", true},
		{"/path/to/file.ts", true},
		{"/path/to/file.js", false},
		{"/path/to/file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := fw.matchesPatterns(tt.path); got != tt.matches {
				t.Errorf("matchesPatterns(%q) = %v, want %v", tt.path, got, tt.matches)
			}
		})
	}
}

func TestFileWatcher_MatchesExclude(t *testing.T) {
	fw := &FileWatcher{
		config: WatchConfig{
			Exclude: []string{"*_test.go", "**/node_modules/**"},
		},
	}

	tests := []struct {
		path    string
		matches bool
	}{
		{"/path/to/file_test.go", true},
		{"/path/to/file.go", false},
		{"/path/to/node_modules/foo/bar.js", true},
		{"/path/to/src/bar.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := fw.matchesExclude(tt.path); got != tt.matches {
				t.Errorf("matchesExclude(%q) = %v, want %v", tt.path, got, tt.matches)
			}
		})
	}
}

func TestFileWatcher_MatchesEventType(t *testing.T) {
	fw := &FileWatcher{
		config: WatchConfig{
			Events: []string{"create", "modify"},
		},
	}

	tests := []struct {
		eventType string
		matches   bool
	}{
		{"create", true},
		{"modify", true},
		{"delete", false},
		{"rename", false},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			if got := fw.matchesEventType(tt.eventType); got != tt.matches {
				t.Errorf("matchesEventType(%q) = %v, want %v", tt.eventType, got, tt.matches)
			}
		})
	}
}

func TestFileWatcher_DeduplicateEvents(t *testing.T) {
	fw := &FileWatcher{}

	events := []WatchEvent{
		{Path: "/a/b.go", EventType: "modify"},
		{Path: "/a/b.go", EventType: "modify"},
		{Path: "/a/c.go", EventType: "create"},
		{Path: "/a/b.go", EventType: "delete"},
	}

	result := fw.deduplicateEvents(events)
	if len(result) != 3 {
		t.Errorf("expected 3 deduplicated events, got %d", len(result))
	}
}

func TestFileWatcher_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tmpDir := t.TempDir()

	var mu sync.Mutex
	var receivedEvents []WatchEvent

	fw, err := NewFileWatcher(WatchConfig{
		Paths:     []string{tmpDir},
		Patterns:  []string{"*.go"},
		Exclude:   []string{"*_test.go"},
		Events:    []string{"create", "modify"},
		Recursive: true,
		Debounce:  50 * time.Millisecond,
	}, func(events []WatchEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, events...)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("NewFileWatcher: %v", err)
	}

	if err := fw.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer fw.Stop()

	// Give watcher time to initialize
	time.Sleep(100 * time.Millisecond)

	// Create a .go file - should trigger
	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Create a _test.go file - should NOT trigger (excluded)
	testFile := filepath.Join(tmpDir, "main_test.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Create a .txt file - should NOT trigger (not matching pattern)
	txtFile := filepath.Join(tmpDir, "readme.txt")
	if err := os.WriteFile(txtFile, []byte("readme"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Wait for debounce
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	count := len(receivedEvents)
	mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 event (main.go), got %d events", count)
	}

	if count > 0 {
		mu.Lock()
		first := receivedEvents[0]
		mu.Unlock()
		if filepath.Base(first.Path) != "main.go" {
			t.Errorf("expected main.go event, got %s", first.Path)
		}
	}
}

func TestCollectChangedFiles(t *testing.T) {
	events := []WatchEvent{
		{Path: "/a/b.go"},
		{Path: "/a/c.go"},
		{Path: "/a/d.go"},
	}

	files := CollectChangedFiles(events)
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}
	if files[0] != "/a/b.go" {
		t.Errorf("expected /a/b.go, got %s", files[0])
	}
}

func TestWatchConfig_Defaults(t *testing.T) {
	fw, err := NewFileWatcher(WatchConfig{}, nil)
	if err != nil {
		t.Fatalf("NewFileWatcher: %v", err)
	}

	if fw.config.Debounce != 100*time.Millisecond {
		t.Errorf("expected default debounce 100ms, got %v", fw.config.Debounce)
	}

	if len(fw.config.Events) != 3 {
		t.Errorf("expected 3 default events, got %d", len(fw.config.Events))
	}

	fw.watcher.Close()
}
