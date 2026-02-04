// Package zettelkasten implements a file-based memory provider using Markdown
// files with TOML frontmatter, following the Zettelkasten methodology.
package zettelkasten

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexcabrera/ayo/internal/paths"
)

// Directory names for memory categories
const (
	DirFacts       = "facts"
	DirPreferences = "preferences"
	DirCorrections = "corrections"
	DirPatterns    = "patterns"
	DirTopics      = "topics"

	// IndexFile is the SQLite index filename (derived, rebuildable)
	IndexFile = ".index.sqlite"

	// IndexMDFile is the optional markdown overview file
	IndexMDFile = "index.md"
)

// DefaultMemoryDir returns the default memory directory path.
func DefaultMemoryDir() string {
	return filepath.Join(paths.DataDir(), "memory")
}

// Structure represents the Zettelkasten directory layout.
type Structure struct {
	// Root is the base directory for all memories
	Root string

	// Facts contains factual information memories
	Facts string

	// Preferences contains user preference memories
	Preferences string

	// Corrections contains behavior correction memories
	Corrections string

	// Patterns contains observed pattern memories
	Patterns string

	// Topics contains topic-based symlinks
	Topics string

	// IndexDB is the path to the SQLite index
	IndexDB string
}

// NewStructure creates a Structure pointing to the given root directory.
func NewStructure(root string) *Structure {
	if root == "" {
		root = DefaultMemoryDir()
	}
	return &Structure{
		Root:        root,
		Facts:       filepath.Join(root, DirFacts),
		Preferences: filepath.Join(root, DirPreferences),
		Corrections: filepath.Join(root, DirCorrections),
		Patterns:    filepath.Join(root, DirPatterns),
		Topics:      filepath.Join(root, DirTopics),
		IndexDB:     filepath.Join(root, IndexFile),
	}
}

// Initialize creates the directory structure if it doesn't exist.
func (s *Structure) Initialize() error {
	dirs := []string{
		s.Root,
		s.Facts,
		s.Preferences,
		s.Corrections,
		s.Patterns,
		s.Topics,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Exists checks if the memory directory structure exists.
func (s *Structure) Exists() bool {
	info, err := os.Stat(s.Root)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CategoryDir returns the directory path for a given category.
func (s *Structure) CategoryDir(category string) string {
	switch category {
	case "fact":
		return s.Facts
	case "preference":
		return s.Preferences
	case "correction":
		return s.Corrections
	case "pattern":
		return s.Patterns
	default:
		return s.Facts // Default to facts
	}
}

// MemoryPath returns the full path for a memory file given its ID and category.
func (s *Structure) MemoryPath(id, category string) string {
	return filepath.Join(s.CategoryDir(category), id+".md")
}

// TopicDir returns the path to a topic's directory under topics/.
func (s *Structure) TopicDir(topic string) string {
	return filepath.Join(s.Topics, topic)
}

// EnsureTopicDir creates a topic directory if it doesn't exist.
func (s *Structure) EnsureTopicDir(topic string) error {
	dir := s.TopicDir(topic)
	return os.MkdirAll(dir, 0o755)
}

// LinkToTopic creates a symlink in the topic directory pointing to a memory.
func (s *Structure) LinkToTopic(memoryID, category, topic string) error {
	if err := s.EnsureTopicDir(topic); err != nil {
		return err
	}

	// Source: the actual memory file
	source := s.MemoryPath(memoryID, category)

	// Target: the symlink in the topic directory
	target := filepath.Join(s.TopicDir(topic), memoryID+".md")

	// Calculate relative path from topic dir to memory file
	relPath, err := filepath.Rel(s.TopicDir(topic), source)
	if err != nil {
		return fmt.Errorf("calculate relative path: %w", err)
	}

	// Remove existing symlink if any
	os.Remove(target)

	return os.Symlink(relPath, target)
}

// UnlinkFromTopic removes a symlink from a topic directory.
func (s *Structure) UnlinkFromTopic(memoryID, topic string) error {
	target := filepath.Join(s.TopicDir(topic), memoryID+".md")
	return os.Remove(target)
}

// ListMemories lists all memory files in a category directory.
func (s *Structure) ListMemories(category string) ([]string, error) {
	dir := s.CategoryDir(category)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var memories []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".md" {
			memories = append(memories, filepath.Join(dir, entry.Name()))
		}
	}

	return memories, nil
}

// ListAllMemories lists all memory files across all categories.
func (s *Structure) ListAllMemories() ([]string, error) {
	categories := []string{"fact", "preference", "correction", "pattern"}
	var all []string

	for _, cat := range categories {
		memories, err := s.ListMemories(cat)
		if err != nil {
			return nil, err
		}
		all = append(all, memories...)
	}

	return all, nil
}

// ListTopics returns all topic directory names.
func (s *Structure) ListTopics() ([]string, error) {
	entries, err := os.ReadDir(s.Topics)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var topics []string
	for _, entry := range entries {
		if entry.IsDir() {
			topics = append(topics, entry.Name())
		}
	}

	return topics, nil
}

// Clean removes empty topic directories.
func (s *Structure) Clean() error {
	topics, err := s.ListTopics()
	if err != nil {
		return err
	}

	for _, topic := range topics {
		dir := s.TopicDir(topic)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		if len(entries) == 0 {
			os.Remove(dir)
		}
	}

	return nil
}
