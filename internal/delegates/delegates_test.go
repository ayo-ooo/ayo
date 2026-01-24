package delegates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexcabrera/ayo/internal/config"
)

func TestResolve(t *testing.T) {
	// Create temp directory for directory config
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	tests := []struct {
		name           string
		taskType       TaskType
		agentDelegates map[string]string
		globalDelegates map[string]string
		dirDelegates   map[string]string
		wantAgent      string
		wantSource     ResolutionSource
	}{
		{
			name:       "no delegation configured",
			taskType:   TaskTypeCoding,
			wantAgent:  "",
			wantSource: SourceNone,
		},
		{
			name:       "global delegation",
			taskType:   TaskTypeCoding,
			globalDelegates: map[string]string{"coding": "@crush"},
			wantAgent:  "@crush",
			wantSource: SourceGlobal,
		},
		{
			name:           "agent overrides global",
			taskType:       TaskTypeCoding,
			agentDelegates: map[string]string{"coding": "@agent-crush"},
			globalDelegates: map[string]string{"coding": "@global-crush"},
			wantAgent:      "@agent-crush",
			wantSource:     SourceAgent,
		},
		{
			name:           "directory overrides agent",
			taskType:       TaskTypeCoding,
			dirDelegates:   map[string]string{"coding": "@dir-crush"},
			agentDelegates: map[string]string{"coding": "@agent-crush"},
			wantAgent:      "@dir-crush",
			wantSource:     SourceDirectory,
		},
		{
			name:           "normalize handle without @",
			taskType:       TaskTypeCoding,
			agentDelegates: map[string]string{"coding": "crush"},
			wantAgent:      "@crush",
			wantSource:     SourceAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up directory config if needed
			if tt.dirDelegates != nil {
				dirCfg := DirectoryConfig{Delegates: tt.dirDelegates}
				data, _ := json.Marshal(dirCfg)
				os.WriteFile(filepath.Join(dir, ".ayo.json"), data, 0o644)
			} else {
				os.Remove(filepath.Join(dir, ".ayo.json"))
			}

			globalCfg := config.Config{Delegates: tt.globalDelegates}

			res := Resolve(tt.taskType, tt.agentDelegates, globalCfg)

			if res.Agent != tt.wantAgent {
				t.Errorf("Agent = %q, want %q", res.Agent, tt.wantAgent)
			}
			if res.Source != tt.wantSource {
				t.Errorf("Source = %v, want %v", res.Source, tt.wantSource)
			}
		})
	}
}

func TestResolveWithFallback(t *testing.T) {
	globalCfg := config.Config{}

	res := ResolveWithFallback(TaskTypeCoding, nil, globalCfg, "@default")

	if res.Agent != "@default" {
		t.Errorf("Agent = %q, want @default", res.Agent)
	}
	if res.Source != SourceNone {
		t.Errorf("Source = %v, want SourceNone", res.Source)
	}
}

func TestGetAllDelegates(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	// Set up directory config
	dirCfg := DirectoryConfig{Delegates: map[string]string{"coding": "@dir"}}
	data, _ := json.Marshal(dirCfg)
	os.WriteFile(filepath.Join(dir, ".ayo.json"), data, 0o644)

	agentDelegates := map[string]string{"research": "@agent"}
	globalCfg := config.Config{Delegates: map[string]string{"debug": "@global"}}

	all := GetAllDelegates(agentDelegates, globalCfg)

	if all["coding"] != "@dir" {
		t.Errorf("coding = %q, want @dir", all["coding"])
	}
	if all["research"] != "@agent" {
		t.Errorf("research = %q, want @agent", all["research"])
	}
	if all["debug"] != "@global" {
		t.Errorf("debug = %q, want @global", all["debug"])
	}
}

func TestLoadDirectoryConfig(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	// No config file
	cfg, path := LoadDirectoryConfig()
	if cfg != nil {
		t.Error("Expected nil config when no file exists")
	}
	if path != "" {
		t.Error("Expected empty path when no file exists")
	}

	// Create config file
	dirCfg := DirectoryConfig{
		Delegates: map[string]string{"coding": "@crush"},
		Model:     "gpt-4",
		Agent:     "@custom",
	}
	data, _ := json.Marshal(dirCfg)
	os.WriteFile(filepath.Join(dir, ".ayo.json"), data, 0o644)

	cfg, path = LoadDirectoryConfig()
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if cfg.Delegates["coding"] != "@crush" {
		t.Errorf("coding = %q, want @crush", cfg.Delegates["coding"])
	}
	if cfg.Model != "gpt-4" {
		t.Errorf("Model = %q, want gpt-4", cfg.Model)
	}
}

func TestSaveDirectoryConfig(t *testing.T) {
	dir := t.TempDir()

	cfg := &DirectoryConfig{
		Delegates: map[string]string{"coding": "@crush"},
		Agent:     "@custom",
	}

	if err := SaveDirectoryConfig(dir, cfg); err != nil {
		t.Fatalf("SaveDirectoryConfig failed: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(filepath.Join(dir, ".ayo.json"))
	if err != nil {
		t.Fatalf("Read config failed: %v", err)
	}

	var loaded DirectoryConfig
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if loaded.Delegates["coding"] != "@crush" {
		t.Errorf("coding = %q, want @crush", loaded.Delegates["coding"])
	}
}
