// Package integration provides integration test utilities for ayo.
// These tests validate end-to-end functionality across multiple packages.
package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
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

	// Use none provider by default (no container isolation)
	env.SandboxProvider = sandbox.NewNoneProvider()

	return env
}

// WithDocker configures the test environment to use Docker sandbox.
// Skips the test if Docker is not available.
func (e *TestEnv) WithDocker() *TestEnv {
	e.t.Helper()

	docker := sandbox.NewDockerProvider()
	if !docker.IsAvailable() {
		e.t.Skip("Docker not available")
	}

	e.SandboxProvider = docker
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

// Exec executes a command in the sandbox and returns the result.
func (e *TestEnv) Exec(ctx context.Context, command string) (providers.ExecResult, error) {
	// Create a sandbox
	sb, err := e.SandboxProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-" + e.t.Name(),
	})
	if err != nil {
		return providers.ExecResult{}, err
	}
	defer e.SandboxProvider.Delete(ctx, sb.ID, true)

	// Execute command
	return e.SandboxProvider.Exec(ctx, sb.ID, providers.ExecOptions{
		Command:    command,
		Timeout:    30 * time.Second,
		WorkingDir: e.baseDir,
	})
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
	Provider string            // "none" or "docker"
	Image    string            // Container image (for docker)
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

// DockerSandboxConfig returns a Docker sandbox configuration.
func DockerSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Provider: "docker",
		Image:    "busybox:stable",
		Network:  true,
	}
}
