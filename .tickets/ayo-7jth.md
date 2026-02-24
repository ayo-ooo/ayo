---
id: ayo-7jth
status: open
deps: [ayo-nqyv]
links: []
created: 2026-02-23T23:13:19Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-pv3a
tags: [squads, config]
---
# Implement ayo.json loader for squads

Update squad loading to read configuration from `ayo.json` instead of SQUAD.md frontmatter. SQUAD.md becomes pure markdown documentation (the "constitution").

## Context

This ticket implements the loader for the squad schema defined in ayo-nqyv. Currently, squad configuration is embedded in SQUAD.md as YAML frontmatter, which is awkward and mixes configuration with documentation. This ticket separates them:

- **ayo.json**: Machine-readable configuration
- **SQUAD.md**: Human-readable constitution (injected into agent prompts)

## Current State

```
~/.local/share/ayo/sandboxes/squads/{name}/
├── SQUAD.md        # Has YAML frontmatter with config
└── workspace/
```

SQUAD.md currently:
```markdown
---
lead: "@architect"
agents: ["@frontend", "@backend"]
planners:
  near_term: ayo-todos
---
# Mission
Build awesome stuff
```

## Target State

```
~/.local/share/ayo/sandboxes/squads/{name}/
├── ayo.json        # Configuration
├── SQUAD.md        # Pure markdown constitution
└── workspace/
```

ayo.json:
```json
{
  "version": "1",
  "squad": {
    "lead": "@architect",
    "agents": ["@frontend", "@backend"],
    "planners": { "near_term": "ayo-todos" }
  }
}
```

SQUAD.md (no frontmatter):
```markdown
# Mission
Build awesome stuff
```

## Files to Modify

1. **`internal/squads/context.go`** - Update `LoadConstitution()` to not parse frontmatter
2. **`internal/squads/config.go`** (new) - Add ayo.json loader for squads
3. **`internal/squads/service.go`** - Update squad initialization to load from ayo.json

## Implementation

```go
// internal/squads/config.go
func LoadSquadConfig(squadDir string) (*SquadConfig, error) {
    ayoPath := filepath.Join(squadDir, "ayo.json")
    
    data, err := os.ReadFile(ayoPath)
    if err != nil {
        return nil, fmt.Errorf("missing ayo.json in squad: %w", err)
    }
    
    var cfg AyoConfig
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("invalid ayo.json: %w", err)
    }
    
    if cfg.Squad == nil {
        return nil, fmt.Errorf("ayo.json missing 'squad' section")
    }
    
    return cfg.Squad, nil
}

// internal/squads/context.go  
func LoadConstitution(squadDir string) (string, error) {
    mdPath := filepath.Join(squadDir, "SQUAD.md")
    data, err := os.ReadFile(mdPath)
    if err != nil {
        return "", err
    }
    
    // No more frontmatter parsing - return raw markdown
    return string(data), nil
}
```

## Acceptance Criteria

- [ ] Squad loads configuration from ayo.json
- [ ] SQUAD.md is read as pure markdown (no frontmatter parsing)
- [ ] Error message is clear when ayo.json missing
- [ ] Existing frontmatter in SQUAD.md is ignored (not an error)
- [ ] Squad lead, agents, planners all load correctly from new format

## Testing

- Test loading squad with ayo.json
- Test loading squad without ayo.json (error)
- Test SQUAD.md with old frontmatter still loads (frontmatter passed through as text)
- Test SQUAD.md without frontmatter loads correctly
- Test all squad config fields parse correctly
