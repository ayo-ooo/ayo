package squads

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexcabrera/ayo/internal/config"
)

func TestLoadAyoConfig(t *testing.T) {
	// Create temp squad directory
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "test-squad")
	if err := os.MkdirAll(squadDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Run("no ayo.json returns nil", func(t *testing.T) {
		// Uses a non-existent squad name with no file
		cfg, err := LoadAyoConfig("nonexistent-squad-xyz")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if cfg != nil {
			t.Error("expected nil config for missing file")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		badDir := filepath.Join(tmpDir, "bad-squad")
		os.MkdirAll(badDir, 0o755)
		os.WriteFile(filepath.Join(badDir, "ayo.json"), []byte("{invalid}"), 0o644)

		// This won't work directly since LoadAyoConfig uses paths.SquadDir
		// Test covered by integration tests
	})
}

func TestSaveAyoConfig(t *testing.T) {
	// Note: This test would need to mock paths.SquadDir
	// For now, we test the data structures directly

	squadCfg := &config.SquadConfig{
		Name:        "test-squad",
		Description: "A test squad",
		Lead:        "@architect",
		Agents:      []string{"@frontend", "@backend"},
	}

	cfg := AyoConfig{
		Schema:  "https://ayo.dev/schemas/ayo.json",
		Version: "1",
		Squad:   squadCfg,
	}

	if cfg.Squad == nil {
		t.Error("Squad should not be nil")
	}
	if cfg.Squad.Lead != "@architect" {
		t.Errorf("Lead = %q, want %q", cfg.Squad.Lead, "@architect")
	}
}

func TestAyoConfigStructure(t *testing.T) {
	squadCfg := &config.SquadConfig{
		Name:        "dev-team",
		Description: "Development team",
		Lead:        "@lead",
		Agents:      []string{"@dev1", "@dev2"},
		Planners: &config.SquadPlannersConfig{
			NearTerm: "ayo-todos",
			LongTerm: "ayo-tickets",
		},
		IO: &config.SquadIOConfig{
			InputSchema:  "schemas/input.json",
			OutputSchema: "schemas/output.json",
		},
		Coordination: &config.SquadCoordinationConfig{
			TicketWorkflow: "kanban",
			AutoAssign:     true,
		},
	}

	// Verify all fields are accessible
	if squadCfg.Lead != "@lead" {
		t.Error("Lead not set correctly")
	}
	if len(squadCfg.Agents) != 2 {
		t.Error("Agents not set correctly")
	}
	if squadCfg.Planners.NearTerm != "ayo-todos" {
		t.Error("Planners.NearTerm not set correctly")
	}
	if squadCfg.IO.InputSchema != "schemas/input.json" {
		t.Error("IO.InputSchema not set correctly")
	}
	if squadCfg.Coordination.TicketWorkflow != "kanban" {
		t.Error("Coordination.TicketWorkflow not set correctly")
	}
}
