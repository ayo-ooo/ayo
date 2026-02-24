---
id: ayo-zn5p
status: open
deps: [ayo-899j, ayo-7xsf]
links: []
created: 2026-02-23T22:16:10Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [triggers, cli]
---
# Add trigger management CLI commands

Polish the `ayo trigger` command family for complete trigger lifecycle management.

## Context

After trigger YAML config (ayo-7xsf) and monitoring (ayo-899j), users need CLI commands to manage triggers without editing YAML files directly.

## Commands

### Create Trigger

```bash
# Interactive creation
ayo trigger create

# One-liner
ayo trigger create health-check \
  --type interval \
  --agent @monitor \
  --every 5m \
  --prompt "Check system health"

# Creates ~/.config/ayo/triggers/health-check.yaml
```

### List Triggers

```bash
ayo trigger list
# NAME              TYPE       STATUS     NEXT RUN           LAST RUN
# health-check      interval   active     2026-02-24 14:05   14:00 ✓
# morning-report    daily      active     2026-02-25 09:00   today ✓
# weekly-summary    weekly     disabled   -                  -
```

### Show Trigger Details

```bash
ayo trigger show health-check
# (detailed output from ayo-899j)
```

### Enable/Disable

```bash
ayo trigger disable morning-report
# ✓ Trigger 'morning-report' disabled

ayo trigger enable morning-report
# ✓ Trigger 'morning-report' enabled
```

### Delete Trigger

```bash
ayo trigger delete weekly-summary
# Are you sure? [y/N] y
# ✓ Trigger 'weekly-summary' deleted

# Force without confirmation
ayo trigger delete weekly-summary --force
```

### Test Trigger

```bash
# Manually fire a trigger
ayo trigger test health-check
# ⏳ Running trigger 'health-check'...
# ✓ Completed in 12s
# Output: No issues found
```

### View History

```bash
ayo trigger history health-check
# (from ayo-899j)
```

## Implementation

### Create Command

```go
// cmd/ayo/trigger_create.go
var triggerCreateCmd = &cobra.Command{
    Use:   "create [name]",
    Short: "Create a new trigger",
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg := TriggerConfig{
            Name:   args[0],
            Type:   flags.Type,
            Agent:  flags.Agent,
            // ... other fields
        }
        
        // Validate
        if err := cfg.Validate(); err != nil {
            return err
        }
        
        // Write YAML
        path := filepath.Join(triggersDir, cfg.Name+".yaml")
        return writeTriggerYAML(path, cfg)
    },
}
```

### Enable/Disable

Modifies the `enabled` field in the YAML file:

```go
func toggleTrigger(name string, enabled bool) error {
    cfg, err := loadTrigger(name)
    if err != nil {
        return err
    }
    cfg.Enabled = enabled
    return saveTrigger(name, cfg)
}
```

## Files to Create/Modify

1. **`cmd/ayo/trigger_create.go`** (new) - Create command
2. **`cmd/ayo/trigger_delete.go`** (new) - Delete command
3. **`cmd/ayo/trigger_enable.go`** (new) - Enable/disable commands
4. **`cmd/ayo/trigger_test.go`** (new) - Manual trigger firing
5. **`cmd/ayo/trigger.go`** - Parent command, register subcommands

## Acceptance Criteria

- [ ] `trigger create` creates YAML file with all options
- [ ] `trigger list` shows all triggers with status
- [ ] `trigger show` displays detailed info
- [ ] `trigger enable/disable` toggles without deleting
- [ ] `trigger delete` removes with confirmation
- [ ] `trigger test` manually fires trigger
- [ ] `trigger history` shows run history
- [ ] All commands support `--json` output

## Testing

- Test create with various options
- Test enable/disable modifies YAML
- Test delete with and without --force
- Test test command fires trigger
- Test list/show output formatting
