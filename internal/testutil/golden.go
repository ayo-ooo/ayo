// Package testutil provides test utilities and fixtures for ayo tests.
package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GoldenDir is the default directory for golden files relative to a test package.
const GoldenDir = "testdata"

// Golden provides golden file testing utilities.
type Golden struct {
	t       *testing.T
	dir     string
	update  bool
	ext     string
	marshal func(any) ([]byte, error)
}

// NewGolden creates a new golden file helper for the given test.
// By default, golden files are stored in testdata/ with .golden extension.
func NewGolden(t *testing.T) *Golden {
	t.Helper()
	return &Golden{
		t:      t,
		dir:    GoldenDir,
		update: os.Getenv("UPDATE_GOLDEN") == "1",
		ext:    ".golden",
		marshal: func(v any) ([]byte, error) {
			return json.MarshalIndent(v, "", "  ")
		},
	}
}

// WithDir sets a custom directory for golden files.
func (g *Golden) WithDir(dir string) *Golden {
	g.dir = dir
	return g
}

// WithExtension sets a custom extension for golden files.
func (g *Golden) WithExtension(ext string) *Golden {
	g.ext = ext
	return g
}

// WithMarshal sets a custom marshal function for converting values to bytes.
func (g *Golden) WithMarshal(fn func(any) ([]byte, error)) *Golden {
	g.marshal = fn
	return g
}

// path returns the full path for a golden file.
func (g *Golden) path(name string) string {
	if !strings.HasSuffix(name, g.ext) {
		name = name + g.ext
	}
	return filepath.Join(g.dir, name)
}

// Assert compares the given data with the golden file.
// If UPDATE_GOLDEN=1 is set, it updates the golden file instead.
func (g *Golden) Assert(name string, got []byte) {
	g.t.Helper()
	goldenPath := g.path(name)

	if g.update {
		// Update mode - write the new golden file
		if err := os.MkdirAll(g.dir, 0755); err != nil {
			g.t.Fatalf("create golden dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, got, 0644); err != nil {
			g.t.Fatalf("write golden file: %v", err)
		}
		return
	}

	// Compare mode
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		g.t.Fatalf("read golden file %s: %v", goldenPath, err)
	}

	if string(got) != string(want) {
		g.t.Errorf("golden file mismatch for %s:\n\nGot:\n%s\n\nWant:\n%s", name, got, want)
	}
}

// AssertString compares the given string with the golden file.
func (g *Golden) AssertString(name string, got string) {
	g.t.Helper()
	g.Assert(name, []byte(got))
}

// AssertValue marshals the value and compares with the golden file.
func (g *Golden) AssertValue(name string, got any) {
	g.t.Helper()
	data, err := g.marshal(got)
	if err != nil {
		g.t.Fatalf("marshal value: %v", err)
	}
	g.Assert(name, data)
}

// AssertJSON is a shortcut for comparing JSON data.
func (g *Golden) AssertJSON(name string, got any) {
	g.t.Helper()
	data, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		g.t.Fatalf("marshal JSON: %v", err)
	}
	g.Assert(name, data)
}

// Read reads the content of a golden file.
func (g *Golden) Read(name string) []byte {
	g.t.Helper()
	data, err := os.ReadFile(g.path(name))
	if err != nil {
		g.t.Fatalf("read golden file: %v", err)
	}
	return data
}

// ReadString reads the content of a golden file as a string.
func (g *Golden) ReadString(name string) string {
	g.t.Helper()
	return string(g.Read(name))
}

// Exists checks if a golden file exists.
func (g *Golden) Exists(name string) bool {
	_, err := os.Stat(g.path(name))
	return err == nil
}

// Update returns true if golden files should be updated.
func (g *Golden) Update() bool {
	return g.update
}

// AssertLines compares line by line, ignoring trailing whitespace.
func (g *Golden) AssertLines(name string, got string) {
	g.t.Helper()
	goldenPath := g.path(name)

	if g.update {
		if err := os.MkdirAll(g.dir, 0755); err != nil {
			g.t.Fatalf("create golden dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
			g.t.Fatalf("write golden file: %v", err)
		}
		return
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		g.t.Fatalf("read golden file %s: %v", goldenPath, err)
	}

	gotLines := strings.Split(strings.TrimSpace(got), "\n")
	wantLines := strings.Split(strings.TrimSpace(string(wantBytes)), "\n")

	if len(gotLines) != len(wantLines) {
		g.t.Errorf("line count mismatch for %s: got %d, want %d", name, len(gotLines), len(wantLines))
		return
	}

	for i := range gotLines {
		gotLine := strings.TrimRight(gotLines[i], " \t")
		wantLine := strings.TrimRight(wantLines[i], " \t")
		if gotLine != wantLine {
			g.t.Errorf("line %d mismatch for %s:\n  got:  %q\n  want: %q", i+1, name, gotLine, wantLine)
		}
	}
}
