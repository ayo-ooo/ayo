---
id: ayo-la11
status: open
deps: [ayo-7dui]
links: []
created: 2026-02-23T23:13:19Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-pv3a
tags: [agents, config]
---
# Implement ayo.json loader for agents

Update agent loading to read `ayo.json` instead of `config.json`. Support both formats during migration period with deprecation warning for `config.json`.

## Context

This ticket implements the loader for the agent schema defined in ayo-7dui. The loader must:
1. Check for `ayo.json` first (new format)
2. Fall back to `config.json` (legacy format) with deprecation warning
3. Parse the new nested schema (`agent.model` instead of `model`)

## Current State

```
~/.config/ayo/agents/{name}/
├── config.json     # Old format: flat structure
├── AGENT.md
└── skills/
```

## Target State

```
~/.config/ayo/agents/{name}/
├── ayo.json        # New format: namespaced structure
├── AGENT.md
└── skills/
```

## Files to Modify

1. **`internal/agent/agent.go`** - Update `loadAgentConfig()` function
2. **`internal/agent/config.go`** - Add AyoConfig struct with agent namespace
3. **`internal/agent/migration.go`** (new) - Add migration helper

## Implementation

```go
// internal/agent/agent.go
func loadAgentConfig(agentDir string) (*AgentConfig, error) {
    // Try new format first
    ayoPath := filepath.Join(agentDir, "ayo.json")
    if _, err := os.Stat(ayoPath); err == nil {
        return loadAyoJSON(ayoPath)
    }
    
    // Fall back to legacy format
    legacyPath := filepath.Join(agentDir, "config.json")
    if _, err := os.Stat(legacyPath); err == nil {
        log.Warn().Str("agent", agentDir).
            Msg("config.json is deprecated, migrate to ayo.json")
        return loadLegacyConfig(legacyPath)
    }
    
    return nil, fmt.Errorf("no config found in %s", agentDir)
}

func loadAyoJSON(path string) (*AgentConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var cfg AyoConfig
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("invalid ayo.json: %w", err)
    }
    
    if cfg.Agent == nil {
        return nil, fmt.Errorf("ayo.json missing 'agent' section")
    }
    
    return cfg.Agent, nil
}
```

## Migration Helper

```go
// internal/agent/migration.go
func MigrateConfigToAyo(agentDir string) error {
    legacyPath := filepath.Join(agentDir, "config.json")
    ayoPath := filepath.Join(agentDir, "ayo.json")
    
    // Read legacy config
    legacy, err := loadLegacyConfig(legacyPath)
    if err != nil {
        return err
    }
    
    // Wrap in new structure
    ayo := AyoConfig{
        Version: "1",
        Agent:   legacy,
    }
    
    // Write new format
    data, _ := json.MarshalIndent(ayo, "", "  ")
    return os.WriteFile(ayoPath, data, 0644)
}
```

## Acceptance Criteria

- [ ] Agent loads from ayo.json when present
- [ ] Agent loads from config.json with deprecation warning when ayo.json missing
- [ ] Error message is clear when neither file exists
- [ ] New AyoConfig struct correctly parses nested agent section
- [ ] Migration helper converts old to new format
- [ ] All existing tests still pass

## Testing

- Test loading agent with ayo.json (new format)
- Test loading agent with config.json (legacy format, check warning)
- Test loading agent with both files (ayo.json takes precedence)
- Test loading agent with neither file (error)
- Test migration helper creates valid ayo.json
