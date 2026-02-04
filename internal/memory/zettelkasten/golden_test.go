package zettelkasten

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/testutil"
)

func TestGolden_ParsePreference(t *testing.T) {
	mem, err := ParseFile("testdata/preference.md")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if mem.Frontmatter.ID != "mem_test001" {
		t.Errorf("id: got %q, want %q", mem.Frontmatter.ID, "mem_test001")
	}
	if mem.Frontmatter.Category != "preference" {
		t.Errorf("category: got %q, want %q", mem.Frontmatter.Category, "preference")
	}
	if len(mem.Frontmatter.Topics) != 2 {
		t.Errorf("topics: got %d, want %d", len(mem.Frontmatter.Topics), 2)
	}
	if mem.Content != "User strongly prefers table-driven tests in Go." {
		t.Errorf("content: got %q", mem.Content)
	}
}

func TestGolden_ParseFact(t *testing.T) {
	mem, err := ParseFile("testdata/fact.md")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if mem.Frontmatter.ID != "mem_test002" {
		t.Errorf("id: got %q, want %q", mem.Frontmatter.ID, "mem_test002")
	}
	if mem.Frontmatter.Category != "fact" {
		t.Errorf("category: got %q, want %q", mem.Frontmatter.Category, "fact")
	}
	if mem.Frontmatter.Scope.Path != "/Users/alex/projects/myapi" {
		t.Errorf("scope.path: got %q", mem.Frontmatter.Scope.Path)
	}
	// Multi-line content
	if len(mem.Content) < 50 {
		t.Errorf("content too short: got %d chars", len(mem.Content))
	}
}

func TestGolden_ParseCorrection(t *testing.T) {
	mem, err := ParseFile("testdata/correction.md")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if mem.Frontmatter.ID != "mem_test003" {
		t.Errorf("id: got %q, want %q", mem.Frontmatter.ID, "mem_test003")
	}
	if mem.Frontmatter.Category != "correction" {
		t.Errorf("category: got %q, want %q", mem.Frontmatter.Category, "correction")
	}
	if mem.Frontmatter.Confidence != 0.9 {
		t.Errorf("confidence: got %f, want %f", mem.Frontmatter.Confidence, 0.9)
	}
	if mem.Frontmatter.Supersession.Supersedes != "mem_old001" {
		t.Errorf("supersedes: got %q", mem.Frontmatter.Supersession.Supersedes)
	}
}

func TestGolden_WriteAndReadRoundtrip(t *testing.T) {
	dir := t.TempDir()

	original := &MemoryFile{
		Frontmatter: Frontmatter{
			ID:         "mem_roundtrip",
			Created:    parseTime(t, "2024-01-15T12:00:00Z"),
			Category:   "pattern",
			Topics:     []string{"testing", "roundtrip"},
			Confidence: 0.85,
		},
		Content: "This is a roundtrip test memory.",
	}

	// Write
	path := filepath.Join(dir, "roundtrip.md")
	if err := original.WriteFile(path); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read back
	parsed, err := ParseFile(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Compare
	if parsed.Frontmatter.ID != original.Frontmatter.ID {
		t.Errorf("id: got %q, want %q", parsed.Frontmatter.ID, original.Frontmatter.ID)
	}
	if parsed.Frontmatter.Category != original.Frontmatter.Category {
		t.Errorf("category: got %q, want %q", parsed.Frontmatter.Category, original.Frontmatter.Category)
	}
	if parsed.Content != original.Content {
		t.Errorf("content: got %q, want %q", parsed.Content, original.Content)
	}
}

func TestGolden_FrontmatterJSON(t *testing.T) {
	g := testutil.NewGolden(t).WithDir("testdata")

	fm := Frontmatter{
		ID:         "mem_json",
		Created:    parseTime(t, "2024-01-15T12:00:00Z"),
		Category:   "preference",
		Topics:     []string{"go", "json"},
		Confidence: 1.0,
		Source: SourceSection{
			SessionID: "sess_001",
		},
	}

	data, err := json.MarshalIndent(fm, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	g.Assert("frontmatter", data)
}

func TestGolden_AllTestFiles(t *testing.T) {
	// Ensure all golden test files parse correctly
	files := []string{"preference.md", "fact.md", "correction.md"}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			path := filepath.Join("testdata", file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("test file not found: %s", path)
			}

			mem, err := ParseFile(path)
			if err != nil {
				t.Fatalf("parse %s: %v", file, err)
			}

			if mem.Frontmatter.ID == "" {
				t.Error("missing ID")
			}
			if mem.Frontmatter.Category == "" {
				t.Error("missing category")
			}
			if mem.Content == "" {
				t.Error("missing content")
			}
		})
	}
}

func parseTime(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return tm
}
