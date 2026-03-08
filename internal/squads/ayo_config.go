package squads

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/pelletier/go-toml/v2"
)

// AyoConfig represents the top-level ayo.json structure for squads.
type AyoConfig struct {
	// Schema enables IDE validation/autocomplete.
	Schema string `json:"$schema,omitempty"`

	// Version is the configuration version.
	Version string `json:"version,omitempty"`

	// Squad contains squad-specific configuration.
	Squad *config.SquadConfig `json:"squad,omitempty"`
}

// LoadAyoConfig loads the ayo.json configuration from a squad directory.
// Returns nil if no ayo.json exists.
func LoadAyoConfig(squadName string) (*AyoConfig, error) {
	squadDir := paths.SquadDir(squadName)
	ayoPath := filepath.Join(squadDir, "ayo.json")

	data, err := os.ReadFile(ayoPath)
	if err != nil {
		if os.IsNotExist(err) {
			debug.Log("no ayo.json found for squad", "squad", squadName)
			return nil, nil // No config is not an error
		}
		return nil, fmt.Errorf("read ayo.json: %w", err)
	}

	var cfg AyoConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse ayo.json: %w", err)
	}

	debug.Log("loaded ayo.json for squad", "squad", squadName)
	return &cfg, nil
}

// LoadSquadConfigFromAyo loads squad configuration from ayo.json.
// Falls back to existing config.LoadSquadConfig if no ayo.json exists.
func LoadSquadConfigFromAyo(squadName string) (config.SquadConfig, error) {
	// Try to load from ayo.json first
	ayoCfg, err := LoadAyoConfig(squadName)
	if err != nil {
		return config.SquadConfig{}, err
	}

	if ayoCfg != nil && ayoCfg.Squad != nil {
		// Use ayo.json config, but ensure Name is set
		cfg := *ayoCfg.Squad
		if cfg.Name == "" {
			cfg.Name = squadName
		}
		return cfg, nil
	}

	// Fall back to existing config loading
	return config.LoadSquadConfig(squadName)
}

// SaveAyoConfig saves squad configuration to ayo.json.
func SaveAyoConfig(squadName string, squadCfg *config.SquadConfig) error {
	squadDir := paths.SquadDir(squadName)
	ayoPath := filepath.Join(squadDir, "ayo.json")

	// Ensure directory exists
	if err := os.MkdirAll(squadDir, 0o755); err != nil {
		return fmt.Errorf("create squad directory: %w", err)
	}

	cfg := AyoConfig{
		Schema:  "https://ayo.dev/schemas/ayo.json",
		Version: "1",
		Squad:   squadCfg,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ayo.json: %w", err)
	}

	if err := os.WriteFile(ayoPath, data, 0o644); err != nil {
		return fmt.Errorf("write ayo.json: %w", err)
	}

	debug.Log("saved ayo.json for squad", "squad", squadName)
	return nil
}

// AyoConfigExists returns true if ayo.json exists for a squad.
func AyoConfigExists(squadName string) bool {
	squadDir := paths.SquadDir(squadName)
	ayoPath := filepath.Join(squadDir, "ayo.json")
	_, err := os.Stat(ayoPath)
	return err == nil
}

// TeamConfigExists returns true if team.toml exists in a directory.
func TeamConfigExists(teamDir string) bool {
	teamPath := filepath.Join(teamDir, "team.toml")
	_, err := os.Stat(teamPath)
	return err == nil
}

// LoadTeamConfigFromTOML loads team configuration from team.toml file.
// This is the new team project format for the build system.
func LoadTeamConfigFromTOML(teamDir string) (*TeamConfig, error) {
	teamPath := filepath.Join(teamDir, "team.toml")
	data, err := os.ReadFile(teamPath)
	if err != nil {
		if os.IsNotExist(err) {
			debug.Log("no team.toml found", "dir", teamDir)
			return nil, nil // No config is not an error
		}
		return nil, fmt.Errorf("read team.toml: %w", err)
	}

	var config TeamConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse team.toml: %w", err)
	}

	debug.Log("loaded team.toml", "dir", teamDir, "team", config.Team.Name)
	return &config, nil
}

// TeamConfig represents the team.toml configuration structure
type TeamConfig struct {
	Team struct {
		Name        string `toml:"name"`
		Description string `toml:"description"`
		Coordination string `toml:"coordination"`
	} `toml:"team"`
	Agents map[string]struct {
		Path string `toml:"path"`
	} `toml:"agents"`
	Workspace struct {
		SharedPath string `toml:"shared_path"`
		OutputPath string `toml:"output_path"`
	} `toml:"workspace"`
	Coordination struct {
		Strategy      string `toml:"strategy"`
		MaxIterations int    `toml:"max_iterations"`
	} `toml:"coordination"`
}
