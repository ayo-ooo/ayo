package zettelkasten

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Common parse errors
var (
	ErrNoFrontmatter    = errors.New("no TOML frontmatter found")
	ErrMissingDelimiter = errors.New("missing closing +++ delimiter")
	ErrMissingID        = errors.New("memory ID is required")
	ErrMissingCreated   = errors.New("created timestamp is required")
	ErrMissingCategory  = errors.New("category is required")
	ErrInvalidCategory  = errors.New("invalid category: must be fact, preference, correction, or pattern")
)

// Frontmatter represents the TOML metadata in a memory file.
type Frontmatter struct {
	ID         string    `toml:"id"`
	Created    time.Time `toml:"created"`
	Updated    time.Time `toml:"updated,omitempty"`
	Category   string    `toml:"category"`
	Status     string    `toml:"status,omitempty"`
	Topics     []string  `toml:"topics,omitempty"`
	Confidence float64   `toml:"confidence,omitempty"`

	Source      SourceSection      `toml:"source,omitempty"`
	Scope       ScopeSection       `toml:"scope,omitempty"`
	Access      AccessSection      `toml:"access,omitempty"`
	Supersession SupersessionSection `toml:"supersession,omitempty"`
	Links       LinksSection       `toml:"links,omitempty"`
	Unclear     UnclearSection     `toml:"unclear,omitempty"`
}

// SourceSection identifies where the memory was formed.
type SourceSection struct {
	SessionID string `toml:"session_id,omitempty"`
	MessageID string `toml:"message_id,omitempty"`
}

// ScopeSection defines the memory's applicability.
type ScopeSection struct {
	Agent string `toml:"agent,omitempty"` // Agent handle (empty = global)
	Path  string `toml:"path,omitempty"`  // Directory path (empty = all paths)
}

// AccessSection tracks memory usage.
type AccessSection struct {
	LastAccessed time.Time `toml:"last_accessed,omitempty"`
	AccessCount  int64     `toml:"access_count,omitempty"`
}

// SupersessionSection tracks memory replacement.
type SupersessionSection struct {
	Supersedes   string `toml:"supersedes,omitempty"`    // ID of memory this replaces
	SupersededBy string `toml:"superseded_by,omitempty"` // ID of memory that replaced this
	Reason       string `toml:"reason,omitempty"`
}

// LinksSection contains related memories.
type LinksSection struct {
	Related []string `toml:"related,omitempty"`
}

// UnclearSection tracks memories needing clarification.
type UnclearSection struct {
	Flagged bool   `toml:"flagged,omitempty"`
	Reason  string `toml:"reason,omitempty"`
}

// MemoryFile represents a complete memory file (frontmatter + content).
type MemoryFile struct {
	Frontmatter Frontmatter
	Content     string
}

// ParseFile reads and parses a memory file from disk.
func ParseFile(path string) (*MemoryFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return Parse(data)
}

// Parse parses a memory file from bytes.
func Parse(data []byte) (*MemoryFile, error) {
	frontmatter, content, err := splitFrontmatter(data)
	if err != nil {
		return nil, err
	}

	var fm Frontmatter
	if err := toml.Unmarshal(frontmatter, &fm); err != nil {
		return nil, fmt.Errorf("parse TOML: %w", err)
	}

	// Set defaults
	if fm.Status == "" {
		fm.Status = "active"
	}
	if fm.Confidence == 0 {
		fm.Confidence = 1.0
	}
	if fm.Updated.IsZero() {
		fm.Updated = fm.Created
	}

	// Validate required fields
	if err := fm.Validate(); err != nil {
		return nil, err
	}

	return &MemoryFile{
		Frontmatter: fm,
		Content:     strings.TrimSpace(content),
	}, nil
}

// splitFrontmatter extracts TOML frontmatter from a markdown file.
// The frontmatter must be delimited by +++ at the start and end.
func splitFrontmatter(data []byte) (frontmatter []byte, content string, err error) {
	reader := bufio.NewReader(bytes.NewReader(data))

	// Read first line - must be +++
	firstLine, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, "", err
	}
	if strings.TrimSpace(firstLine) != "+++" {
		return nil, "", ErrNoFrontmatter
	}

	// Read until closing +++
	var fmBuf bytes.Buffer
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return nil, "", ErrMissingDelimiter
		}
		if err != nil {
			return nil, "", err
		}

		if strings.TrimSpace(line) == "+++" {
			break
		}
		fmBuf.WriteString(line)
	}

	// Rest is content
	var contentBuf bytes.Buffer
	io.Copy(&contentBuf, reader)

	return fmBuf.Bytes(), contentBuf.String(), nil
}

// Validate checks that the frontmatter has all required fields.
func (fm *Frontmatter) Validate() error {
	if fm.ID == "" {
		return ErrMissingID
	}
	if fm.Created.IsZero() {
		return ErrMissingCreated
	}
	if fm.Category == "" {
		return ErrMissingCategory
	}

	// Validate category
	switch fm.Category {
	case "fact", "preference", "correction", "pattern":
		// Valid
	default:
		return fmt.Errorf("%w: got %q", ErrInvalidCategory, fm.Category)
	}

	return nil
}

// Marshal converts a MemoryFile back to bytes.
func (mf *MemoryFile) Marshal() ([]byte, error) {
	var buf bytes.Buffer

	// Write opening delimiter
	buf.WriteString("+++\n")

	// Marshal frontmatter
	fmData, err := toml.Marshal(mf.Frontmatter)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}
	buf.Write(fmData)

	// Write closing delimiter
	buf.WriteString("+++\n\n")

	// Write content
	buf.WriteString(mf.Content)
	buf.WriteString("\n")

	return buf.Bytes(), nil
}

// WriteFile writes a memory file to disk.
func (mf *MemoryFile) WriteFile(path string) error {
	data, err := mf.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// NewMemoryFile creates a new MemoryFile with default values.
func NewMemoryFile(id, category, content string) *MemoryFile {
	now := time.Now().UTC()
	return &MemoryFile{
		Frontmatter: Frontmatter{
			ID:         id,
			Created:    now,
			Updated:    now,
			Category:   category,
			Status:     "active",
			Confidence: 1.0,
		},
		Content: content,
	}
}

// ValidCategories returns the list of valid memory categories.
func ValidCategories() []string {
	return []string{"fact", "preference", "correction", "pattern"}
}

// IsValidCategory checks if a category string is valid.
func IsValidCategory(category string) bool {
	for _, c := range ValidCategories() {
		if c == category {
			return true
		}
	}
	return false
}
