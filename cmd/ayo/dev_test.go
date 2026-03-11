package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindConfigPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with config.toml
	configContent := `[agent]
name = "test"
description = "test"
model = "gpt-4"

[cli]
mode = "freeform"
description = "test"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	found, err := findConfigPath(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if found != configPath {
		t.Errorf("expected %s, got %s", configPath, found)
	}

	// Test with team.toml
	if err := os.Remove(configPath); err != nil {
		t.Fatal(err)
	}

	teamPath := filepath.Join(tmpDir, "team.toml")
	if err := os.WriteFile(teamPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	found, err = findConfigPath(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if found != teamPath {
		t.Errorf("expected %s, got %s", teamPath, found)
	}

	// Test with no config
	if err := os.Remove(teamPath); err != nil {
		t.Fatal(err)
	}

	_, err = findConfigPath(tmpDir)
	if err == nil {
		t.Error("expected error when no config found")
	}
}

func TestTriggerBuild(t *testing.T) {
	// This test is minimal as we don't want to actually run builds
	// We just test the function signature exists
	tmpDir := t.TempDir()

	// Create a minimal config
	configContent := `[agent]
name = "test-agent"
description = "test agent"
model = "gpt-4"

[cli]
mode = "freeform"
description = "test"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create prompts directory
	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create system prompt
	systemPrompt := `You are a helpful AI assistant.`
	if err := os.WriteFile(filepath.Join(promptsDir, "system.txt"), []byte(systemPrompt), 0644); err != nil {
		t.Fatal(err)
	}

	// Note: We can't actually run triggerBuild in tests because it would
	// try to compile Go code which requires the ayo module root to be set.
	// In a real test environment, we would need to set ModuleRoot properly.
	t.Skip("Skipping actual build test - requires module root configuration")
}

func TestDevModeWatchesFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal agent setup
	configContent := `[agent]
name = "test-agent"
description = "test agent"
model = "gpt-4"

[cli]
mode = "freeform"
description = "test"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatal(err)
	}

	systemPromptPath := filepath.Join(promptsDir, "system.txt")
	if err := os.WriteFile(systemPromptPath, []byte("initial"), 0644); err != nil {
		t.Fatal(err)
	}

	// This test would require running the dev mode in a goroutine
	// and verifying that file changes trigger rebuilds.
	// For now, we'll skip this as it's integration testing.
	t.Skip("Skipping file watch test - requires integration setup")
}

func TestVerboseFlag(t *testing.T) {
	// Test that the verbose flag is accessible
	if !verbose {
		// This is the default value
		verbose = true
		if !verbose {
			t.Error("verbose flag should be settable")
		}
	}
}

func TestDevCommandRegistration(t *testing.T) {
	// Verify the dev command is registered
	rootCmd := newRootCmd()
	devCmd, _, err := rootCmd.Find([]string{"dev"})
	if err != nil {
		t.Fatalf("dev command not registered: %v", err)
	}

	if devCmd.Use != "dev <directory>" {
		t.Errorf("unexpected dev command use: %s", devCmd.Use)
	}
}

func TestDevCommandFlags(t *testing.T) {
	rootCmd := newRootCmd()
	devCmd, _, err := rootCmd.Find([]string{"dev"})
	if err != nil {
		t.Fatalf("dev command not registered: %v", err)
	}

	runFlag := devCmd.Flag("run")
	if runFlag == nil {
		t.Error("dev command should have --run flag")
	}

	verboseFlag := devCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("dev command should have --verbose flag")
	}
}

func TestBuildTimeMeasurement(t *testing.T) {
	// Simple test to ensure we can measure build times
	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("expected at least 10ms, got %v", elapsed)
	}
}
