package zettelkasten

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStructure(t *testing.T) {
	root := "/custom/memory"
	s := NewStructure(root)

	if s.Root != root {
		t.Errorf("Root = %q, want %q", s.Root, root)
	}
	if s.Facts != filepath.Join(root, "facts") {
		t.Errorf("Facts = %q, want %q", s.Facts, filepath.Join(root, "facts"))
	}
	if s.IndexDB != filepath.Join(root, ".index.sqlite") {
		t.Errorf("IndexDB = %q, want %q", s.IndexDB, filepath.Join(root, ".index.sqlite"))
	}
}

func TestNewStructureDefaultRoot(t *testing.T) {
	s := NewStructure("")
	if s.Root == "" {
		t.Error("Root should not be empty when created with empty string")
	}
}

func TestStructureInitialize(t *testing.T) {
	dir := t.TempDir()
	s := NewStructure(dir)

	if err := s.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Verify directories exist
	dirs := []string{s.Facts, s.Preferences, s.Corrections, s.Patterns, s.Topics}
	for _, d := range dirs {
		info, err := os.Stat(d)
		if err != nil {
			t.Errorf("Directory %q not created: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", d)
		}
	}
}

func TestStructureExists(t *testing.T) {
	// Non-existent directory
	s := NewStructure("/nonexistent/path")
	if s.Exists() {
		t.Error("Exists() = true for non-existent directory")
	}

	// Existing directory
	dir := t.TempDir()
	s = NewStructure(dir)
	if !s.Exists() {
		t.Error("Exists() = false for existing directory")
	}
}

func TestStructureCategoryDir(t *testing.T) {
	s := NewStructure("/test")

	tests := []struct {
		category string
		want     string
	}{
		{"fact", s.Facts},
		{"preference", s.Preferences},
		{"correction", s.Corrections},
		{"pattern", s.Patterns},
		{"unknown", s.Facts}, // Defaults to facts
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := s.CategoryDir(tt.category)
			if got != tt.want {
				t.Errorf("CategoryDir(%q) = %q, want %q", tt.category, got, tt.want)
			}
		})
	}
}

func TestStructureMemoryPath(t *testing.T) {
	s := NewStructure("/test")

	path := s.MemoryPath("mem_01HX123", "fact")
	want := filepath.Join(s.Facts, "mem_01HX123.md")
	if path != want {
		t.Errorf("MemoryPath() = %q, want %q", path, want)
	}
}

func TestStructureTopicOperations(t *testing.T) {
	dir := t.TempDir()
	s := NewStructure(dir)
	s.Initialize()

	// Create a memory file first
	memID := "mem_test123"
	memPath := s.MemoryPath(memID, "fact")
	os.WriteFile(memPath, []byte("test content"), 0o644)

	// Link to topic
	topic := "go-testing"
	if err := s.LinkToTopic(memID, "fact", topic); err != nil {
		t.Fatalf("LinkToTopic() error = %v", err)
	}

	// Verify topic dir was created
	topicDir := s.TopicDir(topic)
	if _, err := os.Stat(topicDir); err != nil {
		t.Errorf("Topic directory not created: %v", err)
	}

	// Verify symlink exists
	linkPath := filepath.Join(topicDir, memID+".md")
	linkInfo, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("Symlink not created: %v", err)
	}
	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected symlink, got regular file")
	}

	// Unlink from topic
	if err := s.UnlinkFromTopic(memID, topic); err != nil {
		t.Fatalf("UnlinkFromTopic() error = %v", err)
	}

	// Verify symlink removed
	if _, err := os.Stat(linkPath); !os.IsNotExist(err) {
		t.Error("Symlink should be removed")
	}
}

func TestStructureListMemories(t *testing.T) {
	dir := t.TempDir()
	s := NewStructure(dir)
	s.Initialize()

	// Create some memory files
	ids := []string{"mem_1", "mem_2", "mem_3"}
	for _, id := range ids {
		path := s.MemoryPath(id, "fact")
		os.WriteFile(path, []byte("content"), 0o644)
	}

	// Also create a non-md file (should be ignored)
	os.WriteFile(filepath.Join(s.Facts, "ignore.txt"), []byte("ignore"), 0o644)

	// List memories
	memories, err := s.ListMemories("fact")
	if err != nil {
		t.Fatalf("ListMemories() error = %v", err)
	}

	if len(memories) != 3 {
		t.Errorf("ListMemories() returned %d, want 3", len(memories))
	}
}

func TestStructureListAllMemories(t *testing.T) {
	dir := t.TempDir()
	s := NewStructure(dir)
	s.Initialize()

	// Create memories in different categories
	os.WriteFile(s.MemoryPath("mem_fact1", "fact"), []byte("c"), 0o644)
	os.WriteFile(s.MemoryPath("mem_pref1", "preference"), []byte("c"), 0o644)
	os.WriteFile(s.MemoryPath("mem_corr1", "correction"), []byte("c"), 0o644)

	// List all
	all, err := s.ListAllMemories()
	if err != nil {
		t.Fatalf("ListAllMemories() error = %v", err)
	}

	if len(all) != 3 {
		t.Errorf("ListAllMemories() returned %d, want 3", len(all))
	}
}

func TestStructureListTopics(t *testing.T) {
	dir := t.TempDir()
	s := NewStructure(dir)
	s.Initialize()

	// Create some topic directories
	s.EnsureTopicDir("go")
	s.EnsureTopicDir("testing")
	s.EnsureTopicDir("style")

	topics, err := s.ListTopics()
	if err != nil {
		t.Fatalf("ListTopics() error = %v", err)
	}

	if len(topics) != 3 {
		t.Errorf("ListTopics() returned %d, want 3", len(topics))
	}
}

func TestStructureClean(t *testing.T) {
	dir := t.TempDir()
	s := NewStructure(dir)
	s.Initialize()

	// Create some topic directories, some empty
	s.EnsureTopicDir("empty1")
	s.EnsureTopicDir("empty2")
	s.EnsureTopicDir("notempty")

	// Add a file to notempty
	os.WriteFile(filepath.Join(s.TopicDir("notempty"), "link.md"), []byte("c"), 0o644)

	// Clean
	if err := s.Clean(); err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	// Verify empty dirs removed
	if _, err := os.Stat(s.TopicDir("empty1")); !os.IsNotExist(err) {
		t.Error("empty1 should be removed")
	}
	if _, err := os.Stat(s.TopicDir("empty2")); !os.IsNotExist(err) {
		t.Error("empty2 should be removed")
	}

	// Verify non-empty dir kept
	if _, err := os.Stat(s.TopicDir("notempty")); err != nil {
		t.Error("notempty should NOT be removed")
	}
}

func TestListMemoriesNonExistent(t *testing.T) {
	s := NewStructure("/nonexistent")

	memories, err := s.ListMemories("fact")
	if err != nil {
		t.Fatalf("ListMemories() on nonexistent should return nil error, got %v", err)
	}
	if memories != nil {
		t.Errorf("ListMemories() on nonexistent should return nil, got %v", memories)
	}
}
