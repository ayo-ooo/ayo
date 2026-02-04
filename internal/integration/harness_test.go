package integration

import (
	"os"
	"testing"

	"github.com/alexcabrera/ayo/internal/providers"
)

func TestHarness_CreateEnv(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Verify directories exist
	dirs := []string{
		env.ConfigDir,
		env.DataDir,
		env.AgentsDir,
		env.SkillsDir,
		env.SessionDir,
		env.MemoryDir,
	}
	for _, dir := range dirs {
		if !dirExists(dir) {
			t.Errorf("directory not created: %s", dir)
		}
	}
}

func TestHarness_CreateAgent(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	agentDir := env.CreateAgent("@test-agent", "Test agent", "You are a test agent.")

	// Verify files exist
	if !fileExists(agentDir + "/config.json") {
		t.Error("config.json not created")
	}
	if !fileExists(agentDir + "/system.md") {
		t.Error("system.md not created")
	}
}

func TestHarness_CreateSkill(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	skillDir := env.CreateSkill("test-skill", "A test skill", "# Test Skill\n\nThis is a test.")

	if !fileExists(skillDir + "/SKILL.md") {
		t.Error("SKILL.md not created")
	}
}

func TestHarness_ExecNone(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	ctx, cancel := env.Context()
	defer cancel()

	result, err := env.Exec(ctx, "echo hello")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("exit code: got %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("stdout: got %q, want %q", result.Stdout, "hello\n")
	}
}

func TestHarness_ExecDocker(t *testing.T) {
	env := NewTestEnv(t).WithDocker()
	defer env.Cleanup()

	ctx, cancel := env.Context()
	defer cancel()

	result, err := env.Exec(ctx, "echo hello from docker")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("exit code: got %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello from docker\n" {
		t.Errorf("stdout: got %q, want %q", result.Stdout, "hello from docker\n")
	}
}

func TestHarness_SandboxConfig(t *testing.T) {
	cfg := DefaultSandboxConfig()
	if cfg.Provider != "none" {
		t.Errorf("provider: got %q, want %q", cfg.Provider, "none")
	}

	cfg = DockerSandboxConfig()
	if cfg.Provider != "docker" {
		t.Errorf("provider: got %q, want %q", cfg.Provider, "docker")
	}
}

func TestHarness_ProviderInterface(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Verify sandbox provider implements interface
	var _ providers.SandboxProvider = env.SandboxProvider
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
