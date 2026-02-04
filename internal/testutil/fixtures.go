// Package testutil provides test utilities and fixtures for ayo tests.
package testutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/providers"
)

// TempDir creates a temporary directory for testing and returns a cleanup function.
func TempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// TestAgentConfig returns a minimal agent configuration for testing.
func TestAgentConfig() agent.Config {
	return agent.Config{
		Description:  "Test agent for unit tests",
		AllowedTools: []string{"bash"},
	}
}

// TestAgent creates and returns a minimal test agent directory.
// Returns the agent directory path (caller should use agent.Load with config).
func TestAgent(t *testing.T, dir string) string {
	t.Helper()

	// Create agent directory
	agentDir := filepath.Join(dir, "@test-agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("create agent dir: %v", err)
	}

	// Write config.json
	configPath := filepath.Join(agentDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{
		"description": "Test agent",
		"allowed_tools": ["bash"]
	}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Write system.md
	systemPath := filepath.Join(agentDir, "system.md")
	if err := os.WriteFile(systemPath, []byte("You are a test agent."), 0644); err != nil {
		t.Fatalf("write system.md: %v", err)
	}

	return agentDir
}

// TestMemory returns a test memory object.
func TestMemory() *providers.Memory {
	now := time.Now()
	return &providers.Memory{
		ID:              "mem_test123",
		Content:         "Test memory content",
		Category:        providers.MemoryCategoryFact,
		Topics:          []string{"test", "fixture"},
		CreatedAt:       now,
		UpdatedAt:       now,
		SourceSessionID: "sess_test",
		Status:          providers.MemoryStatusActive,
	}
}

// TestSandbox returns a test sandbox object.
func TestSandbox() *providers.Sandbox {
	return &providers.Sandbox{
		ID:        "sb_test123",
		Name:      "test-sandbox",
		Status:    providers.SandboxStatusRunning,
		Image:     "busybox:latest",
		Agents:    []string{"@test"},
		CreatedAt: time.Now(),
	}
}

// TestMount returns a test mount configuration.
func TestMount() providers.Mount {
	return providers.Mount{
		Source:      "/host/path",
		Destination: "/container/path",
		ReadOnly:    false,
	}
}

// TestExecOptions returns test execution options.
func TestExecOptions() providers.ExecOptions {
	return providers.ExecOptions{
		Command:    "echo hello",
		WorkingDir: "/",
		Timeout:    30 * time.Second,
		Env:        map[string]string{"TEST": "1"},
	}
}

// TestExecResult returns a successful test execution result.
func TestExecResult() *providers.ExecResult {
	return &providers.ExecResult{
		Stdout:   "hello\n",
		Stderr:   "",
		ExitCode: 0,
		Duration: 50 * time.Millisecond,
		TimedOut: false,
	}
}

// WithTimeout returns a context with timeout for tests.
func WithTimeout(t *testing.T, d time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), d)
}

// CreateTestAgentDir creates a complete test agent directory structure.
func CreateTestAgentDir(t *testing.T, baseDir, handle string, cfg agent.Config, systemPrompt string) string {
	t.Helper()

	agentDir := filepath.Join(baseDir, handle)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("create agent dir: %v", err)
	}

	// Write config.json
	configJSON := []byte(`{
		"description": "` + cfg.Description + `",
		"allowed_tools": ["bash"]
	}`)
	if err := os.WriteFile(filepath.Join(agentDir, "config.json"), configJSON, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Write system.md
	if err := os.WriteFile(filepath.Join(agentDir, "system.md"), []byte(systemPrompt), 0644); err != nil {
		t.Fatalf("write system.md: %v", err)
	}

	return agentDir
}

// CreateTestSkillDir creates a test skill directory.
func CreateTestSkillDir(t *testing.T, baseDir, name, description, content string) string {
	t.Helper()

	skillDir := filepath.Join(baseDir, name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}

	skillMD := []byte(`---
name: ` + name + `
description: ` + description + `
---

` + content)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillMD, 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	return skillDir
}
