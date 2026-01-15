package config

import (
	"os"
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/paths"
)

func TestDefaultPaths(t *testing.T) {
	cfg := Default()

	// Verify paths use the platform-specific data directory
	if cfg.AgentsDir != paths.AgentsDir() {
		t.Fatalf("agents dir mismatch: got %s, want %s", cfg.AgentsDir, paths.AgentsDir())
	}
	if cfg.SkillsDir != paths.SkillsDir() {
		t.Fatalf("skills dir mismatch: got %s, want %s", cfg.SkillsDir, paths.SkillsDir())
	}

	// System prompts are now resolved at load time via paths.FindPromptFile
	// Default config has empty strings for SystemPrefix and SystemSuffix
	if cfg.SystemPrefix != "" {
		t.Fatalf("expected empty SystemPrefix, got %s", cfg.SystemPrefix)
	}
	if cfg.SystemSuffix != "" {
		t.Fatalf("expected empty SystemSuffix, got %s", cfg.SystemSuffix)
	}

	// All paths should contain "ayo"
	if !strings.Contains(cfg.AgentsDir, "ayo") {
		t.Fatalf("agents dir should contain 'ayo': %s", cfg.AgentsDir)
	}
}

func TestDefaultCatwalkURLFromEnv(t *testing.T) {
	t.Setenv("CATWALK_URL", "https://catwalk.example")
	cfg := Default()
	if cfg.CatwalkBaseURL != "https://catwalk.example" {
		t.Fatalf("expected catwalk base URL from env, got %q", cfg.CatwalkBaseURL)
	}
}

func TestDefaultCatwalkURLFallback(t *testing.T) {
	t.Setenv("CATWALK_URL", "")
	cfg := Default()
	if cfg.CatwalkBaseURL == "" {
		t.Fatalf("expected default catwalk URL to be set")
	}
}

func mustUserHome(t *testing.T) string {
	t.Helper()
	h, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("home: %v", err)
	}
	return h
}

func TestLoadJSONConfig(t *testing.T) {
	// Create a temp config file
	tmpFile, err := os.CreateTemp("", "ayo-config-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write JSON config
	configJSON := `{
		"$schema": "./ayo-schema.json",
		"default_model": "gpt-4-test",
		"agents_dir": "/custom/agents",
		"provider": {
			"name": "anthropic",
			"id": "anthropic",
			"api_endpoint": "https://api.anthropic.com/v1"
		}
	}`
	if _, err := tmpFile.WriteString(configJSON); err != nil {
		t.Fatalf("write config: %v", err)
	}
	tmpFile.Close()

	// Load config
	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	// Verify loaded values
	if cfg.DefaultModel != "gpt-4-test" {
		t.Errorf("expected model 'gpt-4-test', got %q", cfg.DefaultModel)
	}
	if cfg.AgentsDir != "/custom/agents" {
		t.Errorf("expected agents_dir '/custom/agents', got %q", cfg.AgentsDir)
	}
	if cfg.Provider.Name != "anthropic" {
		t.Errorf("expected provider name 'anthropic', got %q", cfg.Provider.Name)
	}
	if cfg.Schema != "./ayo-schema.json" {
		t.Errorf("expected $schema './ayo-schema.json', got %q", cfg.Schema)
	}
}

func TestLoadMissingConfig(t *testing.T) {
	// Load from non-existent file should return defaults
	cfg, err := Load("/nonexistent/path/ayo.json")
	if err != nil {
		t.Fatalf("load missing config should not error: %v", err)
	}

	// Should have default values
	if cfg.DefaultModel != "gpt-4.1" {
		t.Errorf("expected default model 'gpt-4.1', got %q", cfg.DefaultModel)
	}
}
