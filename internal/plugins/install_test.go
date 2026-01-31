package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunPostInstallHook(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ayo-test-plugin-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple post-install script that creates a marker file
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}

	markerFile := filepath.Join(tmpDir, "installed.marker")
	scriptContent := `#!/bin/sh
touch "$1/installed.marker"
`
	scriptPath := filepath.Join(scriptsDir, "post-install.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	// Run the hook
	if err := runPostInstallHook(tmpDir, "scripts/post-install.sh"); err != nil {
		t.Errorf("runPostInstallHook() error = %v", err)
	}

	// Verify the marker file was created
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error("Post-install hook did not create expected marker file")
	}
}

func TestRunPostInstallHook_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-test-plugin-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Try to run non-existent hook
	err = runPostInstallHook(tmpDir, "scripts/nonexistent.sh")
	if err == nil {
		t.Error("expected error for non-existent script")
	}
}

func TestRunPostInstallHook_IsDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-test-plugin-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a directory where the script should be
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}

	// Try to run a directory as script
	err = runPostInstallHook(tmpDir, "scripts")
	if err == nil {
		t.Error("expected error when hook path is a directory")
	}
}
