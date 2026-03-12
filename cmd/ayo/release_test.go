package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/build"
)

func TestCreateGitTag(t *testing.T) {
	// Test in a non-git directory (should error)
	tmpDir := t.TempDir()

	err := createGitTag(tmpDir, "v1.0.0", "Test release")
	if err == nil {
		t.Error("createGitTag should error in non-git directory")
	}

	// The error should mention git failure
	if !strings.Contains(err.Error(), "git") {
		t.Errorf("expected git-related error, got: %v", err)
	}
}

func TestRunRelease(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
description = "Test agent for release"
version = "1.0.0"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test bumping patch version
	err := runRelease(tmpDir, "patch", "", "")
	if err != nil {
		t.Fatalf("runRelease with patch bump failed: %v", err)
	}

	// Verify version was updated in config
	config, _, err := build.LoadConfigFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config after release: %v", err)
	}

	if config.Agent.Version != "1.0.1" {
		t.Errorf("expected version 1.0.1, got %s", config.Agent.Version)
	}
}

func TestRunReleaseMinor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
description = "Test agent for release"
version = "1.0.0"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test bumping minor version
	err := runRelease(tmpDir, "minor", "", "")
	if err != nil {
		t.Fatalf("runRelease with minor bump failed: %v", err)
	}

	// Verify version was updated
	config, _, err := build.LoadConfigFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config after release: %v", err)
	}

	if config.Agent.Version != "1.1.0" {
		t.Errorf("expected version 1.1.0, got %s", config.Agent.Version)
	}
}

func TestRunReleaseMajor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
description = "Test agent for release"
version = "1.0.0"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test bumping major version
	err := runRelease(tmpDir, "major", "", "")
	if err != nil {
		t.Fatalf("runRelease with major bump failed: %v", err)
	}

	// Verify version was updated
	config, _, err := build.LoadConfigFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config after release: %v", err)
	}

	if config.Agent.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %s", config.Agent.Version)
	}
}

func TestRunReleasePreRelease(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
description = "Test agent for release"
version = "1.0.0"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
 if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test bumping with pre-release
	err := runRelease(tmpDir, "patch", "beta", "")
	if err != nil {
		t.Fatalf("runRelease with pre-release failed: %v", err)
	}

	// Verify version was updated
	config, _, err := build.LoadConfigFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config after release: %v", err)
	}

	if config.Agent.Version != "1.0.1-beta" {
		t.Errorf("expected version 1.0.1-beta, got %s", config.Agent.Version)
	}

	// Verify CHANGELOG was created (without git tag since it's a pre-release)
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		t.Error("CHANGELOG.md should be created even for pre-releases")
	}
}

func TestRunReleaseBuildMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
description = "Test agent for release"
version = "1.0.0"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test bumping with build metadata
	err := runRelease(tmpDir, "patch", "", "exp.sha.5114f85")
	if err != nil {
		t.Fatalf("runRelease with build metadata failed: %v", err)
	}

	// Verify version was updated
	config, _, err := build.LoadConfigFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config after release: %v", err)
	}

	if config.Agent.Version != "1.0.1+exp.sha.5114f85" {
		t.Errorf("expected version 1.0.1+exp.sha.5114f85, got %s", config.Agent.Version)
	}
}

func TestRunReleaseNoBumpType(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
description = "Test agent for release"
version = "1.0.0"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test without bump type (should error)
	err := runRelease(tmpDir, "", "", "")
	if err == nil {
		t.Error("runRelease should error when no bump type is specified")
	}

	if !strings.Contains(err.Error(), "please specify a version or use --bump flag") {
		t.Errorf("expected 'please specify a version' error, got: %v", err)
	}
}

func TestRunReleaseInvalidBumpType(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
description = "Test agent for release"
version = "1.0.0"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with invalid bump type
	err := runRelease(tmpDir, "invalid", "", "")
	if err == nil {
		t.Error("runRelease should error with invalid bump type")
	}

	if !strings.Contains(err.Error(), "invalid bump type") {
		t.Errorf("expected 'invalid bump type' error, got: %v", err)
	}
}

func TestRunReleaseDefaultVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config.toml without version
	configContent := `[agent]
name = "test-agent"
description = "Test agent for release"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test bumping when version is not set (should default to 0.0.0)
	err := runRelease(tmpDir, "patch", "", "")
	if err != nil {
		t.Fatalf("runRelease with default version failed: %v", err)
	}

	// Verify version was updated from 0.0.0
	config, _, err := build.LoadConfigFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config after release: %v", err)
	}

	if config.Agent.Version != "0.0.1" {
		t.Errorf("expected version 0.0.1, got %s", config.Agent.Version)
	}
}

func TestRunReleaseNonExistentDirectory(t *testing.T) {
	// Test with non-existent directory
	err := runRelease("/nonexistent/directory", "patch", "", "")
	if err == nil {
		t.Error("runRelease should error for non-existent directory")
	}
}

func TestRunReleaseInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid config.toml
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("invalid config"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with invalid config
	err := runRelease(tmpDir, "patch", "", "")
	if err == nil {
		t.Error("runRelease should error for invalid config")
	}
}

func TestIncrementVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"increment zero", "0", "1"},
		{"increment one", "1", "2"},
		{"increment large", "99", "100"},
		{"increment very large", "999", "1000"},
		{"non-numeric defaults to 1", "abc", "1"},
		{"empty defaults to 1", "", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := incrementVersion(tt.input)
			if result != tt.expected {
				t.Errorf("incrementVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
