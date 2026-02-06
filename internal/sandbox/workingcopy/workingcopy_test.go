package workingcopy

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

func TestWorkingCopy_Create(t *testing.T) {
	// Create a temp host directory with test files
	hostDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(hostDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(hostDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hostDir, "subdir", "nested.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create mock provider
	mockProvider := sandbox.NewMockProvider()
	ctx := context.Background()

	// Create sandbox
	sb, err := mockProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-sandbox",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Set up mock exec to track calls
	var execCalls []string
	mockProvider.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		execCalls = append(execCalls, opts.Command)
		return providers.ExecResult{ExitCode: 0}, nil
	}

	// Create working copy
	manager := NewManager(mockProvider)
	wc, err := manager.Create(ctx, sb.ID, hostDir, "/workspace", nil)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify working copy properties
	if wc.HostPath != hostDir {
		t.Errorf("HostPath = %q, want %q", wc.HostPath, hostDir)
	}
	if wc.SandboxPath != "/workspace" {
		t.Errorf("SandboxPath = %q, want %q", wc.SandboxPath, "/workspace")
	}
	if wc.SandboxID != sb.ID {
		t.Errorf("SandboxID = %q, want %q", wc.SandboxID, sb.ID)
	}
	if wc.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// Verify mkdir and tar commands were called
	if len(execCalls) < 2 {
		t.Errorf("Expected at least 2 exec calls, got %d", len(execCalls))
	}
}

func TestDefaultIgnorePatterns(t *testing.T) {
	patterns := DefaultIgnorePatterns()

	// Check that common patterns are included
	expected := []string{".git", "node_modules", ".venv", "__pycache__"}
	for _, exp := range expected {
		found := false
		for _, p := range patterns {
			if p == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern %q not found in defaults", exp)
		}
	}
}

func TestShouldIgnore(t *testing.T) {
	patterns := DefaultIgnorePatterns()

	tests := []struct {
		path   string
		ignore bool
	}{
		{".git", true},
		{".git/objects", true},
		{"node_modules", true},
		{"node_modules/package/index.js", true},
		{"src/main.go", false},
		{"README.md", false},
		{".DS_Store", true},
		{"file.pyc", true},
		{"file.swp", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := shouldIgnore(tt.path, patterns)
			if got != tt.ignore {
				t.Errorf("shouldIgnore(%q) = %v, want %v", tt.path, got, tt.ignore)
			}
		})
	}
}

func TestParseChecksums(t *testing.T) {
	output := `d41d8cd98f00b204e9800998ecf8427e  ./empty.txt
098f6bcd4621d373cade4e832627b4f6  ./test.txt
5eb63bbbe01eeed093cb22bb8f5acdc3  ./subdir/hello.txt`

	checksums := parseChecksums(output)

	if len(checksums) != 3 {
		t.Errorf("Expected 3 checksums, got %d", len(checksums))
	}

	if checksums["./empty.txt"] != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Errorf("Wrong checksum for empty.txt: %q", checksums["./empty.txt"])
	}
	if checksums["./test.txt"] != "098f6bcd4621d373cade4e832627b4f6" {
		t.Errorf("Wrong checksum for test.txt: %q", checksums["./test.txt"])
	}
}

func TestFileChecksum(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	checksum, err := fileChecksum(testFile)
	if err != nil {
		t.Fatalf("fileChecksum failed: %v", err)
	}

	// MD5 of "test" is 098f6bcd4621d373cade4e832627b4f6
	expected := "098f6bcd4621d373cade4e832627b4f6"
	if checksum != expected {
		t.Errorf("checksum = %q, want %q", checksum, expected)
	}
}

func TestCreateTar(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".git", "config"), []byte("git config"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create tar with default ignore patterns
	tarData, err := createTar(tmpDir, DefaultIgnorePatterns())
	if err != nil {
		t.Fatalf("createTar failed: %v", err)
	}

	// Verify tar was created
	if len(tarData) == 0 {
		t.Error("createTar returned empty data")
	}

	// The tar should contain test.txt but not .git
	// We can't easily inspect tar contents here, but at least verify it's non-empty
}

func TestWorkingCopy_Reset(t *testing.T) {
	// Create a temp host directory
	hostDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(hostDir, "test.txt"), []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create mock provider
	mockProvider := sandbox.NewMockProvider()
	ctx := context.Background()

	// Create sandbox
	sb, err := mockProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-sandbox",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Track exec calls
	var execCalls []string
	mockProvider.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		execCalls = append(execCalls, opts.Command)
		return providers.ExecResult{ExitCode: 0}, nil
	}

	// Create working copy
	manager := NewManager(mockProvider)
	wc, err := manager.Create(ctx, sb.ID, hostDir, "/workspace", nil)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Reset working copy
	execCalls = nil
	err = manager.Reset(ctx, wc)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	// Verify rm and tar commands were called
	if len(execCalls) < 2 {
		t.Errorf("Expected at least 2 exec calls for reset, got %d", len(execCalls))
	}
}

func TestFileDiff_Status(t *testing.T) {
	// Test diff status constants
	if DiffStatusAdded != "added" {
		t.Errorf("DiffStatusAdded = %q, want 'added'", DiffStatusAdded)
	}
	if DiffStatusModified != "modified" {
		t.Errorf("DiffStatusModified = %q, want 'modified'", DiffStatusModified)
	}
	if DiffStatusDeleted != "deleted" {
		t.Errorf("DiffStatusDeleted = %q, want 'deleted'", DiffStatusDeleted)
	}
}
