// Package integration provides integration test utilities for ayo.
// These tests validate end-to-end functionality across multiple packages.
package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
	// Removed sandbox import during sandbox infrastructure removal
	// "github.com/alexcabrera/ayo/internal/sandbox"
)

// TestEnv holds a complete test environment for integration tests.
type TestEnv struct {
	t       *testing.T
	baseDir string

	// Directories
	ConfigDir  string
	DataDir    string
	AgentsDir  string
	SkillsDir  string
	SessionDir string
	MemoryDir  string

	// Providers
	SandboxProvider providers.SandboxProvider
}

// NewTestEnv creates a new test environment with isolated directories.
func NewTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	baseDir := t.TempDir()

	env := &TestEnv{
		t:          t,
		baseDir:    baseDir,
		ConfigDir:  filepath.Join(baseDir, ".config", "ayo"),
		DataDir:    filepath.Join(baseDir, ".local", "share", "ayo"),
		AgentsDir:  filepath.Join(baseDir, ".config", "ayo", "agents"),
		SkillsDir:  filepath.Join(baseDir, ".config", "ayo", "skills"),
		SessionDir: filepath.Join(baseDir, ".local", "share", "ayo", "sessions"),
		MemoryDir:  filepath.Join(baseDir, ".local", "share", "ayo", "memory"),
	}

	// Create directories
	dirs := []string{
		env.ConfigDir,
		env.DataDir,
		env.AgentsDir,
		env.SkillsDir,
		env.SessionDir,
		env.MemoryDir,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("create dir %s: %v", dir, err)
		}
	}

	// Sandbox infrastructure removed - set to nil
	env.SandboxProvider = nil

	return env
}

// WithAppleContainer is disabled during sandbox infrastructure removal.
func (e *TestEnv) WithAppleContainer() *TestEnv {
	e.t.Helper()
	e.t.Skip("Apple Container provider disabled during sandbox infrastructure removal")
	return e
}

// WithLinuxContainer is disabled during sandbox infrastructure removal.
func (e *TestEnv) WithLinuxContainer() *TestEnv {
	e.t.Helper()
	e.t.Skip("Linux container provider disabled during sandbox infrastructure removal")
	return e
}

// CreateAgent creates a test agent in the environment.
func (e *TestEnv) CreateAgent(handle, description, systemPrompt string) string {
	e.t.Helper()

	agentDir := filepath.Join(e.AgentsDir, handle)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		e.t.Fatalf("create agent dir: %v", err)
	}

	// Write config.json
	config := []byte(`{
	"description": "` + description + `",
	"allowed_tools": ["bash"]
}`)
	if err := os.WriteFile(filepath.Join(agentDir, "config.json"), config, 0644); err != nil {
		e.t.Fatalf("write config: %v", err)
	}

	// Write system.md
	if err := os.WriteFile(filepath.Join(agentDir, "system.md"), []byte(systemPrompt), 0644); err != nil {
		e.t.Fatalf("write system.md: %v", err)
	}

	return agentDir
}

// CreateSkill creates a test skill in the environment.
func (e *TestEnv) CreateSkill(name, description, content string) string {
	e.t.Helper()

	skillDir := filepath.Join(e.SkillsDir, name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		e.t.Fatalf("create skill dir: %v", err)
	}

	skillMD := []byte(`---
name: ` + name + `
description: ` + description + `
---

` + content)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillMD, 0644); err != nil {
		e.t.Fatalf("write SKILL.md: %v", err)
	}

	return skillDir
}

// Exec executes a command and returns result.
// Sandbox execution disabled during sandbox infrastructure removal.
func (e *TestEnv) Exec(ctx context.Context, command string) (providers.ExecResult, error) {
	// Sandbox infrastructure removed - skip execution
	return providers.ExecResult{}, fmt.Errorf("sandbox execution disabled during infrastructure removal")
}

// Cleanup performs test cleanup.
func (e *TestEnv) Cleanup() {
	// Cleanup is automatic with t.TempDir()
}

// Context returns a test context with timeout.
func (e *TestEnv) Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 60*time.Second)
}

// SandboxConfig provides common sandbox test configurations.
type SandboxConfig struct {
	Provider string            // "none", "apple-container", or "systemd-nspawn"
	Image    string            // Container image (for container providers)
	Mounts   []providers.Mount // Mount points
	Network  bool              // Enable networking
}

// DefaultSandboxConfig returns a minimal sandbox configuration.
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Provider: "none",
		Image:    "busybox:stable",
		Network:  true,
	}
}

// AppleContainerSandboxConfig returns an Apple Container sandbox configuration.
func AppleContainerSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Provider: "apple-container",
		Image:    "busybox:stable",
		Network:  true,
	}
}

// LinuxSandboxConfig returns a Linux container sandbox configuration.
func LinuxSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Provider: "systemd-nspawn",
		Image:    "",
		Network:  true,
	}
}
