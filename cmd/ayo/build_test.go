package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildBasicAgent(t *testing.T) {
	// Skip if ModuleRoot is not set (ayo binary wasn't built with ldflags)
	if ModuleRoot == "" {
		t.Skip("ModuleRoot not set, skipping build test")
	}

	// Create a temporary directory for the test agent
	tempDir := t.TempDir()

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
description = "A test agent"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test agent CLI"
`
	if err := os.WriteFile(filepath.Join(tempDir, "config.toml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Create prompts directory with system.md
	promptsDir := filepath.Join(tempDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("Failed to create prompts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(promptsDir, "system.md"), []byte("# System Prompt\n\nYou are a test agent."), 0644); err != nil {
		t.Fatalf("Failed to create system prompt: %v", err)
	}

	// Run build
	outputPath := filepath.Join(tempDir, "test-agent")
	err := runBuild(tempDir, outputPath, "", "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify output binary exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output binary not created at %s", outputPath)
	}
}

func TestBuildAgentWithSkills(t *testing.T) {
	if ModuleRoot == "" {
		t.Skip("ModuleRoot not set, skipping build test")
	}

	tempDir := t.TempDir()

	// Create config
	configContent := `[agent]
name = "skilled-agent"
description = "Agent with skills"
model = "gpt-4"

[cli]
mode = "hybrid"
description = "Agent with skills"
`
	if err := os.WriteFile(filepath.Join(tempDir, "config.toml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Create skills directory with skill files
	skillsDir := filepath.Join(tempDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "coding.md"), []byte("# Coding\n\nYou can code."), 0644); err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "writing.md"), []byte("# Writing\n\nYou can write."), 0644); err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	// Run build
	outputPath := filepath.Join(tempDir, "skilled-agent")
	err := runBuild(tempDir, outputPath, "", "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify output
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output binary not created")
	}
}

func TestBuildAgentWithTools(t *testing.T) {
	if ModuleRoot == "" {
		t.Skip("ModuleRoot not set, skipping build test")
	}

	tempDir := t.TempDir()

	// Create config
	configContent := `[agent]
name = "tool-agent"
description = "Agent with tools"
model = "gpt-4"

[cli]
mode = "structured"
description = "Agent with tools"
`
	if err := os.WriteFile(filepath.Join(tempDir, "config.toml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Create tools directory
	toolsDir := filepath.Join(tempDir, "tools")
	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		t.Fatalf("Failed to create tools dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "tool1.txt"), []byte("Tool 1"), 0644); err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Run build
	outputPath := filepath.Join(tempDir, "tool-agent")
	err := runBuild(tempDir, outputPath, "", "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify output
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output binary not created")
	}
}

func TestBuildAgentMissingConfig(t *testing.T) {
	if ModuleRoot == "" {
		t.Skip("ModuleRoot not set, skipping build test")
	}

	tempDir := t.TempDir()

	// Don't create config.toml

	// Run build - should fail
	err := runBuild(tempDir, filepath.Join(tempDir, "output"), "", "")
	if err == nil {
		t.Error("Expected build to fail with missing config")
	}

	// Verify error message
	if err != nil && !os.IsNotExist(err) {
		// Expected error
	}
}

func TestBuildCrossPlatform(t *testing.T) {
	if ModuleRoot == "" {
		t.Skip("ModuleRoot not set, skipping build test")
	}

	tempDir := t.TempDir()

	// Create minimal config
	configContent := `[agent]
name = "cross-platform"
description = "Test cross-platform build"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test"
`
	if err := os.WriteFile(filepath.Join(tempDir, "config.toml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Test Linux build
	outputLinux := filepath.Join(tempDir, "cross-platform-linux")
	err := runBuild(tempDir, outputLinux, "linux", "amd64")
	if err != nil {
		t.Fatalf("Linux build failed: %v", err)
	}
	if _, err := os.Stat(outputLinux); os.IsNotExist(err) {
		t.Errorf("Linux binary not created")
	}

	// Test Windows build
	outputWindows := filepath.Join(tempDir, "cross-platform-windows")
	err = runBuild(tempDir, outputWindows, "windows", "amd64")
	if err != nil {
		t.Fatalf("Windows build failed: %v", err)
	}
	if _, err := os.Stat(outputWindows); os.IsNotExist(err) {
		t.Errorf("Windows binary not created")
	}
}

func TestFindModuleRoot(t *testing.T) {
	// Test that we can find module root when ModuleRoot is set via ldflags
	if ModuleRoot == "" {
		t.Log("ModuleRoot not set (expected when not built with ldflags)")
	} else {
		t.Logf("ModuleRoot: %s", ModuleRoot)

		// Verify the path exists
		if _, err := os.Stat(ModuleRoot); os.IsNotExist(err) {
			t.Errorf("ModuleRoot path does not exist: %s", ModuleRoot)
		}
	}
}
