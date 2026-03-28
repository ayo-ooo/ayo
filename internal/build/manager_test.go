package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ayo-ooo/ayo/internal/project"
	"github.com/ayo-ooo/ayo/internal/schema"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Error("NewManager() should not return nil")
	}
}

func TestGetSchema_Nil(t *testing.T) {
	result := getSchema(nil)
	if result != nil {
		t.Error("getSchema(nil) should return nil")
	}
}

func TestGetSchema_WithParsedSchema(t *testing.T) {
	parsed := &schema.ParsedSchema{
		Type: "object",
		Properties: map[string]schema.Property{
			"name": {Type: "string"},
		},
	}

	s := &project.Schema{
		Content: []byte(`{"type":"object"}`),
		Parsed:  parsed,
	}

	result := getSchema(s)
	if result == nil {
		t.Error("getSchema() should return parsed schema")
	}
	if result.Type != "object" {
		t.Errorf("getSchema() returned wrong type: %s", result.Type)
	}
}

func TestGetSchema_WithWrongType(t *testing.T) {
	s := &project.Schema{
		Content: []byte(`{"type":"object"}`),
		Parsed:  "not a schema", // wrong type
	}

	result := getSchema(s)
	if result != nil {
		t.Error("getSchema() should return nil for wrong type")
	}
}

func TestCopyDirectory(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source structure
	srcFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	srcSubdir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(srcSubdir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	srcSubFile := filepath.Join(srcSubdir, "nested.txt")
	if err := os.WriteFile(srcSubFile, []byte("nested content"), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Copy directory
	if err := copyDirectory(srcDir, filepath.Join(dstDir, "copied")); err != nil {
		t.Fatalf("copyDirectory() error = %v", err)
	}

	// Verify copy
	copiedFile := filepath.Join(dstDir, "copied", "test.txt")
	data, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("Copied file content = %q, want %q", string(data), "test content")
	}

	copiedNested := filepath.Join(dstDir, "copied", "subdir", "nested.txt")
	data, err = os.ReadFile(copiedNested)
	if err != nil {
		t.Fatalf("Failed to read nested file: %v", err)
	}
	if string(data) != "nested content" {
		t.Errorf("Nested file content = %q, want %q", string(data), "nested content")
	}
}

func TestCopyDirectory_NonExistent(t *testing.T) {
	err := copyDirectory("/nonexistent/path", t.TempDir())
	if err == nil {
		t.Error("copyDirectory() should return error for non-existent source")
	}
}


func TestManager_Cleanup(t *testing.T) {
	m := NewManager()

	// Create a temp dir manually
	m.buildDir = t.TempDir()

	// Create a file in it
	testFile := filepath.Join(m.buildDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Cleanup
	m.Cleanup()

	// Verify directory is removed
	if _, err := os.Stat(m.buildDir); !os.IsNotExist(err) {
		t.Error("Cleanup() should remove build directory")
	}

	// Cleanup with empty buildDir should not panic
	m.buildDir = ""
	m.Cleanup() // Should not panic
}
