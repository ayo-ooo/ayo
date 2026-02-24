package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MigrateToAyoJSON migrates an agent from config.json to ayo.json format.
// Returns nil if already migrated or no config exists.
func MigrateToAyoJSON(agentDir string) error {
	ayoPath := filepath.Join(agentDir, "ayo.json")
	configPath := filepath.Join(agentDir, "config.json")

	// Skip if ayo.json already exists
	if _, err := os.Stat(ayoPath); err == nil {
		return nil
	}

	// Check if config.json exists
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to migrate
		}
		return err
	}

	// Parse legacy config
	var legacy Config
	if err := json.Unmarshal(data, &legacy); err != nil {
		return fmt.Errorf("parse config.json: %w", err)
	}

	// Create ayo.json with namespaced structure
	ayo := AyoConfig{
		Schema:  "https://ayo.dev/schemas/agent.json",
		Version: "1",
		Agent:   &legacy,
	}

	// Write new format
	newData, err := json.MarshalIndent(ayo, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ayo.json: %w", err)
	}

	if err := os.WriteFile(ayoPath, newData, 0644); err != nil {
		return fmt.Errorf("write ayo.json: %w", err)
	}

	return nil
}

// MigrateAllAgents migrates all agents in a directory from config.json to ayo.json.
// Returns the number of agents migrated and any error.
func MigrateAllAgents(agentsDir string) (int, error) {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	migrated := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		agentDir := filepath.Join(agentsDir, entry.Name())
		configPath := filepath.Join(agentDir, "config.json")
		ayoPath := filepath.Join(agentDir, "ayo.json")

		// Skip if no config.json or ayo.json already exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}
		if _, err := os.Stat(ayoPath); err == nil {
			continue
		}

		if err := MigrateToAyoJSON(agentDir); err != nil {
			return migrated, fmt.Errorf("migrate %s: %w", entry.Name(), err)
		}
		migrated++
	}

	return migrated, nil
}
