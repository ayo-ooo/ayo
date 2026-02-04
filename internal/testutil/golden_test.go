package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGolden_Assert(t *testing.T) {
	dir := t.TempDir()
	g := NewGolden(t).WithDir(dir)

	// Test update mode first
	os.Setenv("UPDATE_GOLDEN", "1")
	defer os.Unsetenv("UPDATE_GOLDEN")
	g = NewGolden(t).WithDir(dir)

	g.Assert("test1", []byte("hello world"))

	// Verify file was created
	content, err := os.ReadFile(filepath.Join(dir, "test1.golden"))
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("got %q, want %q", content, "hello world")
	}
}

func TestGolden_AssertString(t *testing.T) {
	dir := t.TempDir()

	// Create golden file
	if err := os.WriteFile(filepath.Join(dir, "test.golden"), []byte("expected"), 0644); err != nil {
		t.Fatal(err)
	}

	g := NewGolden(t).WithDir(dir)
	g.AssertString("test", "expected")
}

func TestGolden_AssertJSON(t *testing.T) {
	dir := t.TempDir()

	// Create golden file with expected JSON
	if err := os.WriteFile(filepath.Join(dir, "json_test.golden"), []byte(`{
  "name": "test",
  "value": 42
}`), 0644); err != nil {
		t.Fatal(err)
	}

	g := NewGolden(t).WithDir(dir)
	g.AssertJSON("json_test", map[string]any{"name": "test", "value": 42})
}

func TestGolden_Read(t *testing.T) {
	dir := t.TempDir()

	// Create golden file
	if err := os.WriteFile(filepath.Join(dir, "read_test.golden"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	g := NewGolden(t).WithDir(dir)
	content := g.ReadString("read_test")
	if content != "content" {
		t.Errorf("got %q, want %q", content, "content")
	}
}

func TestGolden_Exists(t *testing.T) {
	dir := t.TempDir()

	g := NewGolden(t).WithDir(dir)

	if g.Exists("nonexistent") {
		t.Error("expected Exists to return false for nonexistent file")
	}

	// Create file
	if err := os.WriteFile(filepath.Join(dir, "exists.golden"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if !g.Exists("exists") {
		t.Error("expected Exists to return true for existing file")
	}
}

func TestGolden_WithExtension(t *testing.T) {
	dir := t.TempDir()

	os.Setenv("UPDATE_GOLDEN", "1")
	defer os.Unsetenv("UPDATE_GOLDEN")

	g := NewGolden(t).WithDir(dir).WithExtension(".txt")
	g.Assert("custom_ext", []byte("data"))

	// Verify file was created with custom extension
	if _, err := os.Stat(filepath.Join(dir, "custom_ext.txt")); err != nil {
		t.Errorf("expected file with .txt extension: %v", err)
	}
}

func TestGolden_AssertLines(t *testing.T) {
	dir := t.TempDir()

	// Create golden file with trailing whitespace
	if err := os.WriteFile(filepath.Join(dir, "lines.golden"), []byte("line1  \nline2\nline3"), 0644); err != nil {
		t.Fatal(err)
	}

	g := NewGolden(t).WithDir(dir)
	// Should match even with different trailing whitespace
	g.AssertLines("lines", "line1\nline2  \nline3")
}
