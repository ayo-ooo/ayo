package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTempDir(t *testing.T) {
	dir := TempDir(t)
	if dir == "" {
		t.Error("TempDir returned empty string")
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat temp dir: %v", err)
	}
	if !info.IsDir() {
		t.Error("TempDir did not return a directory")
	}
}

func TestTestAgentConfig(t *testing.T) {
	cfg := TestAgentConfig()
	if cfg.Description == "" {
		t.Error("Config should have description")
	}
	if len(cfg.AllowedTools) == 0 {
		t.Error("Config should have allowed tools")
	}
}

func TestTestAgent(t *testing.T) {
	dir := TempDir(t)
	agentDir := TestAgent(t, dir)

	if agentDir == "" {
		t.Error("TestAgent returned empty path")
	}

	// Verify files exist
	configPath := filepath.Join(agentDir, "config.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Error("config.json not created")
	}

	systemPath := filepath.Join(agentDir, "system.md")
	if _, err := os.Stat(systemPath); err != nil {
		t.Error("system.md not created")
	}
}

func TestTestMemory(t *testing.T) {
	mem := TestMemory()
	if mem.ID == "" {
		t.Error("Memory should have ID")
	}
	if mem.Content == "" {
		t.Error("Memory should have content")
	}
	if mem.Category == "" {
		t.Error("Memory should have category")
	}
}

func TestTestSandbox(t *testing.T) {
	sb := TestSandbox()
	if sb.ID == "" {
		t.Error("Sandbox should have ID")
	}
	if sb.Name == "" {
		t.Error("Sandbox should have name")
	}
}

func TestTestMount(t *testing.T) {
	mount := TestMount()
	if mount.Source == "" {
		t.Error("Mount should have source")
	}
	if mount.Destination == "" {
		t.Error("Mount should have destination")
	}
}

func TestTestExecOptions(t *testing.T) {
	opts := TestExecOptions()
	if opts.Command == "" {
		t.Error("ExecOptions should have command")
	}
	if opts.Timeout == 0 {
		t.Error("ExecOptions should have timeout")
	}
}

func TestTestExecResult(t *testing.T) {
	result := TestExecResult()
	if result.ExitCode != 0 {
		t.Error("TestExecResult should have exit code 0")
	}
}

func TestWithTimeout(t *testing.T) {
	ctx, cancel := WithTimeout(t, 1*time.Second)
	defer cancel()
	if ctx == nil {
		t.Error("WithTimeout should return non-nil context")
	}
}

func TestCreateTestAgentDir(t *testing.T) {
	dir := TempDir(t)
	cfg := TestAgentConfig()
	agentDir := CreateTestAgentDir(t, dir, "@custom", cfg, "Custom prompt")

	if agentDir == "" {
		t.Error("CreateTestAgentDir returned empty path")
	}

	systemPath := filepath.Join(agentDir, "system.md")
	content, err := os.ReadFile(systemPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Custom prompt" {
		t.Errorf("system.md content = %q, want 'Custom prompt'", content)
	}
}

func TestCreateTestSkillDir(t *testing.T) {
	dir := TempDir(t)
	skillDir := CreateTestSkillDir(t, dir, "test-skill", "A test skill", "Skill content here")

	if skillDir == "" {
		t.Error("CreateTestSkillDir returned empty path")
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Error("SKILL.md not created")
	}
}
