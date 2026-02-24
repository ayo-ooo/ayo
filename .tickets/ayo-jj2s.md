---
id: ayo-jj2s
status: closed
deps: [ayo-q841]
links: []
created: 2026-02-23T22:16:02Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-sqad
tags: [triggers, cron]
---
# Polish cron trigger configuration

Clean up the trigger engine's cron handling. Support friendly schedule syntax aliases, add validation, and improve error messages.

## Context

After migrating to gocron v2 (ayo-q841), polish the cron trigger experience with better syntax support and validation.

## Friendly Aliases

Support common aliases in addition to cron expressions:

| Alias | Cron Expression | Description |
|-------|-----------------|-------------|
| `@hourly` | `0 * * * *` | Every hour |
| `@daily` | `0 0 * * *` | Every day at midnight |
| `@weekly` | `0 0 * * 0` | Every Sunday at midnight |
| `@monthly` | `0 0 1 * *` | First of month at midnight |
| `@yearly` | `0 0 1 1 *` | January 1st at midnight |
| `@weekdays` | `0 9 * * 1-5` | 9am on weekdays |
| `@weekends` | `0 9 * * 0,6` | 9am on weekends |

## Enhanced Syntax

Support human-readable additions:

```yaml
# Standard cron
schedule:
  cron: "0 9 * * *"

# With alias
schedule:
  cron: "@daily"

# With timezone
schedule:
  cron: "0 9 * * *"
  timezone: "America/New_York"
```

## Validation

Validate cron expressions before registering:

```go
// internal/daemon/cron_parser.go
func ValidateCronExpression(expr string) error {
    // Check for alias
    if strings.HasPrefix(expr, "@") {
        if _, ok := cronAliases[expr]; !ok {
            return fmt.Errorf("unknown cron alias: %s", expr)
        }
        return nil
    }
    
    // Parse cron expression
    _, err := gocron.CronJob(expr, false)
    if err != nil {
        return fmt.Errorf("invalid cron expression '%s': %w", expr, err)
    }
    
    return nil
}
```

## Error Messages

Improve error messages with suggestions:

```
Error: Invalid cron expression '0 9 * *'
       Expected 5 fields (minute hour day month weekday), got 4
       
       Example: '0 9 * * *' runs at 9:00 AM every day
       
       See 'ayo help cron' for syntax reference
```

## Help Text

Add comprehensive cron help:

```bash
ayo help cron
# Cron Expression Syntax
# ======================
# 
# Format: minute hour day month weekday
# 
# Field     Allowed Values
# -----     --------------
# minute    0-59
# hour      0-23
# day       1-31
# month     1-12 or JAN-DEC
# weekday   0-6 or SUN-SAT (0=Sunday)
# 
# Special Characters
# ------------------
# *   Any value
# ,   List separator (1,3,5)
# -   Range (1-5)
# /   Step (/15 = every 15)
# 
# Aliases
# -------
# @hourly    Every hour
# @daily     Every day at midnight
# @weekly    Every Sunday at midnight
# @monthly   First of month
# @weekdays  9am Mon-Fri
```

## Files to Create/Modify

1. **`internal/daemon/cron_parser.go`** (new) - Alias expansion, validation
2. **`internal/daemon/trigger_engine.go`** - Use new parser
3. **`cmd/ayo/help_cron.go`** (new) - Cron help command
4. **`docs/triggers.md`** - Document cron syntax

## Acceptance Criteria

- [ ] All aliases expand correctly
- [ ] Timezone support works
- [ ] Invalid expressions show clear error
- [ ] Error messages include suggestions
- [ ] `ayo help cron` shows syntax reference
- [ ] Existing cron triggers still work

## Testing

- Test all alias expansions
- Test timezone handling
- Test invalid expression errors
- Test error message quality
- Test help command output
