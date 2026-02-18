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

	// Should have a default model (dynamically determined based on credentials)
	if cfg.DefaultModel == "" {
		t.Error("expected non-empty default model")
	}
}

func TestDefaultProvidersConfig(t *testing.T) {
	cfg := Default()

	// Verify default active providers
	if cfg.Providers.Active["memory"] != "zettelkasten" {
		t.Errorf("expected memory provider 'zettelkasten', got %q", cfg.Providers.Active["memory"])
	}
	if cfg.Providers.Active["sandbox"] != "none" {
		t.Errorf("expected sandbox provider 'none', got %q", cfg.Providers.Active["sandbox"])
	}
	if cfg.Providers.Active["embedding"] != "ollama" {
		t.Errorf("expected embedding provider 'ollama', got %q", cfg.Providers.Active["embedding"])
	}
	if cfg.Providers.Active["observer"] != "memory-extractor" {
		t.Errorf("expected observer provider 'memory-extractor', got %q", cfg.Providers.Active["observer"])
	}

	// Verify memory config
	if !cfg.Providers.Memory.AutoMerge {
		t.Error("expected memory auto-merge to be enabled")
	}

	// Verify sandbox config
	if cfg.Providers.Sandbox.Image != "busybox:latest" {
		t.Errorf("expected sandbox image 'busybox:latest', got %q", cfg.Providers.Sandbox.Image)
	}
	if cfg.Providers.Sandbox.Pool.MinSize != 1 {
		t.Errorf("expected sandbox pool min size 1, got %d", cfg.Providers.Sandbox.Pool.MinSize)
	}
	if cfg.Providers.Sandbox.Pool.MaxSize != 4 {
		t.Errorf("expected sandbox pool max size 4, got %d", cfg.Providers.Sandbox.Pool.MaxSize)
	}
	if cfg.Providers.Sandbox.Pool.IdleTimeout != "30m" {
		t.Errorf("expected sandbox idle timeout '30m', got %q", cfg.Providers.Sandbox.Pool.IdleTimeout)
	}
	if cfg.Providers.Sandbox.Network.Enabled == nil || !*cfg.Providers.Sandbox.Network.Enabled {
		t.Error("expected sandbox network to be enabled by default")
	}
	if cfg.Providers.Sandbox.Resources.CPUs != 2 {
		t.Errorf("expected sandbox CPUs 2, got %d", cfg.Providers.Sandbox.Resources.CPUs)
	}
	if cfg.Providers.Sandbox.Resources.MemoryMB != 2048 {
		t.Errorf("expected sandbox memory 2048 MB, got %d", cfg.Providers.Sandbox.Resources.MemoryMB)
	}

	// Verify embedding config
	if cfg.Providers.Embedding.Model != "nomic-embed-text" {
		t.Errorf("expected embedding model 'nomic-embed-text', got %q", cfg.Providers.Embedding.Model)
	}

	// Verify observer config
	if cfg.Providers.Observer.Enabled == nil || !*cfg.Providers.Observer.Enabled {
		t.Error("expected observer to be enabled by default")
	}
	if cfg.Providers.Observer.BatchSize != 10 {
		t.Errorf("expected observer batch size 10, got %d", cfg.Providers.Observer.BatchSize)
	}
	if cfg.Providers.Observer.DebounceMs != 1000 {
		t.Errorf("expected observer debounce 1000ms, got %d", cfg.Providers.Observer.DebounceMs)
	}
}

func TestLoadProvidersConfig(t *testing.T) {
	// Create a temp config file with providers config
	tmpFile, err := os.CreateTemp("", "ayo-config-providers-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write JSON config with custom providers
	configJSON := `{
		"providers": {
			"active": {
				"memory": "custom-memory",
				"sandbox": "apple-container"
			},
			"memory": {
				"directory": "/custom/memory",
				"auto_merge": false
			},
			"sandbox": {
				"image": "custom:latest",
				"pool": {
					"min_size": 3,
					"max_size": 6,
					"idle_timeout": "1h"
				},
				"resources": {
					"cpus": 4,
					"memory_mb": 4096
				}
			}
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

	// Verify custom active providers
	if cfg.Providers.Active["memory"] != "custom-memory" {
		t.Errorf("expected memory provider 'custom-memory', got %q", cfg.Providers.Active["memory"])
	}
	if cfg.Providers.Active["sandbox"] != "apple-container" {
		t.Errorf("expected sandbox provider 'apple-container', got %q", cfg.Providers.Active["sandbox"])
	}

	// Verify custom memory config
	if cfg.Providers.Memory.Directory != "/custom/memory" {
		t.Errorf("expected memory directory '/custom/memory', got %q", cfg.Providers.Memory.Directory)
	}
	if cfg.Providers.Memory.AutoMerge {
		t.Error("expected memory auto-merge to be disabled")
	}

	// Verify custom sandbox config
	if cfg.Providers.Sandbox.Image != "custom:latest" {
		t.Errorf("expected sandbox image 'custom:latest', got %q", cfg.Providers.Sandbox.Image)
	}
	if cfg.Providers.Sandbox.Pool.MinSize != 3 {
		t.Errorf("expected sandbox pool min size 3, got %d", cfg.Providers.Sandbox.Pool.MinSize)
	}
	if cfg.Providers.Sandbox.Pool.MaxSize != 6 {
		t.Errorf("expected sandbox pool max size 6, got %d", cfg.Providers.Sandbox.Pool.MaxSize)
	}
	if cfg.Providers.Sandbox.Resources.CPUs != 4 {
		t.Errorf("expected sandbox CPUs 4, got %d", cfg.Providers.Sandbox.Resources.CPUs)
	}
}

func TestPlannersConfig_WithDefaults(t *testing.T) {
	tests := []struct {
		name string
		cfg  PlannersConfig
		want PlannersConfig
	}{
		{
			name: "empty config gets defaults",
			cfg:  PlannersConfig{},
			want: PlannersConfig{NearTerm: "ayo-todos", LongTerm: "ayo-tickets"},
		},
		{
			name: "partial config gets partial defaults",
			cfg:  PlannersConfig{NearTerm: "custom-todos"},
			want: PlannersConfig{NearTerm: "custom-todos", LongTerm: "ayo-tickets"},
		},
		{
			name: "full config unchanged",
			cfg:  PlannersConfig{NearTerm: "custom-todos", LongTerm: "custom-tickets"},
			want: PlannersConfig{NearTerm: "custom-todos", LongTerm: "custom-tickets"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.WithDefaults()
			if got != tt.want {
				t.Errorf("PlannersConfig.WithDefaults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlannersConfig_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		cfg  PlannersConfig
		want bool
	}{
		{
			name: "empty config",
			cfg:  PlannersConfig{},
			want: true,
		},
		{
			name: "near term only",
			cfg:  PlannersConfig{NearTerm: "ayo-todos"},
			want: false,
		},
		{
			name: "long term only",
			cfg:  PlannersConfig{LongTerm: "ayo-tickets"},
			want: false,
		},
		{
			name: "both set",
			cfg:  PlannersConfig{NearTerm: "a", LongTerm: "b"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsEmpty(); got != tt.want {
				t.Errorf("PlannersConfig.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadPlannersConfig(t *testing.T) {
	// Create a temp config file
	tmpFile, err := os.CreateTemp("", "ayo-config-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write JSON config with planners
	configJSON := `{
		"planners": {
			"near_term": "custom-todos",
			"long_term": "custom-tickets"
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
	if cfg.Planners.NearTerm != "custom-todos" {
		t.Errorf("expected near_term 'custom-todos', got %q", cfg.Planners.NearTerm)
	}
	if cfg.Planners.LongTerm != "custom-tickets" {
		t.Errorf("expected long_term 'custom-tickets', got %q", cfg.Planners.LongTerm)
	}
}
