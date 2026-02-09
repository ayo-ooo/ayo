package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSandboxDir(t *testing.T) {
	dir := SandboxDir()
	if dir == "" {
		t.Error("SandboxDir returned empty string")
	}
	// Should end with "sandbox"
	if filepath.Base(dir) != "sandbox" {
		t.Errorf("SandboxDir should end with 'sandbox', got %s", dir)
	}
}

func TestSubdirectories(t *testing.T) {
	tests := []struct {
		name string
		fn   func() string
		base string
	}{
		{"HomesDir", HomesDir, "homes"},
		{"SharedDir", SharedDir, "shared"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.fn()
			if filepath.Base(dir) != tt.base {
				t.Errorf("%s should end with '%s', got %s", tt.name, tt.base, filepath.Base(dir))
			}
			// Should be under SandboxDir
			if filepath.Dir(dir) != SandboxDir() {
				t.Errorf("%s should be under SandboxDir", tt.name)
			}
		})
	}
}

func TestInit(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ayo-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the sandbox dir for testing
	origSandboxDir := sandboxDirOverride
	sandboxDirOverride = tmpDir
	defer func() { sandboxDirOverride = origSandboxDir }()

	// Test initialization
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Should now be initialized
	if !IsInitialized() {
		t.Error("IsInitialized should return true after Init")
	}

	// Check directories were created
	dirs := []string{"homes", "shared"}
	for _, dir := range dirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", dir)
		}
		// Check .gitkeep exists
		gitkeep := filepath.Join(path, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			t.Errorf(".gitkeep not created in %s", dir)
		}
	}

	// Check .gitignore was created
	gitignore := filepath.Join(tmpDir, ".gitignore")
	if _, err := os.Stat(gitignore); os.IsNotExist(err) {
		t.Error(".gitignore was not created")
	}

	// Check .git directory exists
	gitDir := filepath.Join(tmpDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory was not created")
	}
}

func TestCommit(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ayo-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the sandbox dir for testing
	origSandboxDir := sandboxDirOverride
	sandboxDirOverride = tmpDir
	defer func() { sandboxDirOverride = origSandboxDir }()

	// Initialize first
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "homes", "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Commit should succeed
	if err := Commit("Test commit"); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Commit with no changes should succeed (no-op)
	if err := Commit("No changes"); err != nil {
		t.Fatalf("Commit with no changes failed: %v", err)
	}
}

func TestGetBranch(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ayo-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the sandbox dir for testing
	origSandboxDir := sandboxDirOverride
	sandboxDirOverride = tmpDir
	defer func() { sandboxDirOverride = origSandboxDir }()

	// Initialize first
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Get branch - should be main or master
	branch, err := GetBranch()
	if err != nil {
		t.Fatalf("GetBranch failed: %v", err)
	}

	// Git default branch could be main or master depending on git version
	if branch != "main" && branch != "master" {
		t.Errorf("Expected branch to be 'main' or 'master', got '%s'", branch)
	}
}

func TestHasChanges(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ayo-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the sandbox dir for testing
	origSandboxDir := sandboxDirOverride
	sandboxDirOverride = tmpDir
	defer func() { sandboxDirOverride = origSandboxDir }()

	// Initialize first
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Should have no changes after init
	hasChanges, err := HasChanges()
	if err != nil {
		t.Fatalf("HasChanges failed: %v", err)
	}
	if hasChanges {
		t.Error("Expected no changes after init")
	}

	// Create a new file
	testFile := filepath.Join(tmpDir, "homes", "new.txt")
	if err := os.WriteFile(testFile, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should now have changes
	hasChanges, err = HasChanges()
	if err != nil {
		t.Fatalf("HasChanges failed: %v", err)
	}
	if !hasChanges {
		t.Error("Expected changes after creating file")
	}
}

func TestCreateMachineBranch(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ayo-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the sandbox dir for testing
	origSandboxDir := sandboxDirOverride
	sandboxDirOverride = tmpDir
	defer func() { sandboxDirOverride = origSandboxDir }()

	// Initialize first
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create machine branch
	branchName, err := CreateMachineBranch()
	if err != nil {
		t.Fatalf("CreateMachineBranch failed: %v", err)
	}

	// Should start with "machines/"
	if len(branchName) < 10 || branchName[:9] != "machines/" {
		t.Errorf("Expected branch name to start with 'machines/', got '%s'", branchName)
	}

	// Current branch should now be the machine branch
	currentBranch, err := GetBranch()
	if err != nil {
		t.Fatalf("GetBranch failed: %v", err)
	}
	if currentBranch != branchName {
		t.Errorf("Expected current branch to be '%s', got '%s'", branchName, currentBranch)
	}

	// Creating again should succeed (checkout existing)
	branchName2, err := CreateMachineBranch()
	if err != nil {
		t.Fatalf("CreateMachineBranch second call failed: %v", err)
	}
	if branchName2 != branchName {
		t.Errorf("Expected same branch name on second call, got '%s' vs '%s'", branchName, branchName2)
	}
}

func TestIsInitialized_NotInitialized(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ayo-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the sandbox dir for testing
	origSandboxDir := sandboxDirOverride
	sandboxDirOverride = tmpDir
	defer func() { sandboxDirOverride = origSandboxDir }()

	// Should not be initialized
	if IsInitialized() {
		t.Error("IsInitialized should return false for uninitialized directory")
	}
}
