---
id: ayo-mp44
status: open
deps: [ayo-7jth]
links: []
created: 2026-02-23T23:13:09Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-pv3a
tags: [squads, config, migration]
---
# Migrate SQUAD.md frontmatter to ayo.json

Create a migration tool that converts existing SQUAD.md frontmatter configuration to ayo.json format, then strips the frontmatter from SQUAD.md.

## Context

After ayo-7jth implements the new ayo.json loader, existing squads still have configuration in SQUAD.md frontmatter. This ticket provides automatic migration to convert them to the new format.

## Migration Process

1. Parse SQUAD.md frontmatter (YAML)
2. Generate ayo.json with squad namespace
3. Strip frontmatter from SQUAD.md (keep markdown body)
4. Display deprecation warning for unmigrated squads

## Before Migration

```
squad-dir/
├── SQUAD.md (with frontmatter)
└── workspace/
```

SQUAD.md:
```markdown
---
lead: "@architect"
agents: ["@frontend", "@backend"]
planners:
  near_term: ayo-todos
  long_term: ayo-tickets
input_accepts: "@planner"
---
# Mission
Build the auth system...
```

## After Migration

```
squad-dir/
├── ayo.json (new)
├── SQUAD.md (stripped)
└── workspace/
```

ayo.json:
```json
{
  "version": "1",
  "squad": {
    "lead": "@architect",
    "agents": ["@frontend", "@backend"],
    "planners": {
      "near_term": "ayo-todos",
      "long_term": "ayo-tickets"
    },
    "input_accepts": "@planner"
  }
}
```

SQUAD.md (no frontmatter):
```markdown
# Mission
Build the auth system...
```

## Files to Create/Modify

1. **`internal/squads/migration.go`** (new) - Migration logic
2. **`cmd/ayo/migrate.go`** (new) - CLI command for migration
3. **`internal/squads/service.go`** - Add deprecation warning on load

## Implementation

```go
// internal/squads/migration.go
func MigrateSquadConfig(squadDir string) error {
    squadMD := filepath.Join(squadDir, "SQUAD.md")
    ayoJSON := filepath.Join(squadDir, "ayo.json")
    
    // Skip if already migrated
    if _, err := os.Stat(ayoJSON); err == nil {
        return nil
    }
    
    // Read and parse frontmatter
    data, err := os.ReadFile(squadMD)
    if err != nil {
        return err
    }
    
    frontmatter, body, err := parseFrontmatter(data)
    if err != nil {
        return err  // No frontmatter = already clean
    }
    
    // Convert to new format
    cfg := AyoConfig{
        Version: "1",
        Squad: &SquadConfig{
            Lead:         frontmatter.Lead,
            Agents:       frontmatter.Agents,
            InputAccepts: frontmatter.InputAccepts,
            Planners:     frontmatter.Planners,
        },
    }
    
    // Write ayo.json
    cfgData, _ := json.MarshalIndent(cfg, "", "  ")
    if err := os.WriteFile(ayoJSON, cfgData, 0644); err != nil {
        return err
    }
    
    // Strip frontmatter from SQUAD.md
    return os.WriteFile(squadMD, []byte(body), 0644)
}
```

## CLI Command

```bash
# Migrate single squad
ayo migrate squad auth-team

# Migrate all squads
ayo migrate squads --all

# Dry run (show what would change)
ayo migrate squads --all --dry-run
```

## Deprecation Warning

When loading a squad that still has frontmatter but no ayo.json:

```
⚠️  Squad 'auth-team' uses deprecated SQUAD.md frontmatter.
    Run 'ayo migrate squad auth-team' to update.
```

## Acceptance Criteria

- [ ] Migration parses YAML frontmatter correctly
- [ ] Migration generates valid ayo.json
- [ ] Migration strips frontmatter from SQUAD.md
- [ ] Migration preserves markdown body exactly
- [ ] CLI command works for single squad and all squads
- [ ] Dry run mode shows changes without applying
- [ ] Deprecation warning shown for unmigrated squads
- [ ] Migration is idempotent (safe to run multiple times)

## Testing

- Test migration with various frontmatter configurations
- Test migration preserves markdown formatting
- Test dry run doesn't modify files
- Test idempotency (running twice produces same result)
- Test error handling for invalid frontmatter
