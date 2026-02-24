# Trigger Configuration Guide

Complete reference for configuring ayo trigger.

## Trigger Types

| Type | Description | Use Case |
|------|-------------|----------|
| `cron` | Time-based schedule | Daily reports, hourly checks |
| `interval` | Recurring interval | Every 30 minutes |
| `once` | One-time execution | Future reminders |
| `watch` | File system changes | Auto-review on save |
| `daily` | Daily at specific time | Morning standup |
| `weekly` | Weekly on specific day | Weekly reports |
| `monthly` | Monthly on specific day | Monthly summaries |

## CLI Reference

### Create Triggers

```bash
# Cron trigger
ayo trigger schedule morning-standup \
  --cron "0 9 * * MON-FRI" \
  --agent @standup \
  --prompt "Run daily standup"

# Interval trigger
ayo trigger schedule health-check \
  --interval 30m \
  --agent @monitor \
  --prompt "Check system health"

# One-time trigger
ayo trigger schedule reminder \
  --once "2024-12-25 09:00" \
  --agent @ayo \
  --prompt "Send holiday message"

# Watch trigger
ayo trigger schedule code-watcher \
  --watch ./src \
  --pattern "*.go" \
  --agent @reviewer \
  --prompt "Review changed files"

# Daily trigger
ayo trigger schedule daily-backup \
  --daily 02:00 \
  --agent @backup \
  --prompt "Run nightly backup"

# Weekly trigger
ayo trigger schedule weekly-report \
  --weekly monday 09:00 \
  --agent @reporter \
  --prompt "Generate weekly report"
```

### Manage Triggers

```bash
ayo trigger list                 # List all triggers
ayo trigger show <name>          # Show trigger details
ayo trigger enable <name>        # Enable trigger
ayo trigger disable <name>       # Disable trigger
ayo trigger remove <name>        # Remove trigger
ayo trigger fire <name>          # Manual execution
ayo trigger history [name]       # View execution history
```

## Cron Expressions

### Standard Format

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-6, Sun=0)
│ │ │ │ │
* * * * *
```

### Extended Format (with seconds)

```
┌───────────── second (0-59)
│ ┌───────────── minute (0-59)
│ │ ┌───────────── hour (0-23)
│ │ │ ┌───────────── day of month (1-31)
│ │ │ │ ┌───────────── month (1-12)
│ │ │ │ │ ┌───────────── day of week (0-6, Sun=0)
│ │ │ │ │ │
* * * * * *
```

### Special Characters

| Character | Description | Example |
|-----------|-------------|---------|
| `*` | Any value | `* * * * *` (every minute) |
| `,` | Value list | `0,30 * * * *` (0 and 30) |
| `-` | Range | `9-17 * * * *` (9am-5pm) |
| `/` | Step | `*/15 * * * *` (every 15 min) |
| `L` | Last | `0 0 L * *` (last day of month) |
| `W` | Weekday | `0 0 15W * *` (nearest weekday to 15th) |

### Aliases

| Alias | Equivalent | Description |
|-------|------------|-------------|
| `@yearly` | `0 0 1 1 *` | January 1st, midnight |
| `@monthly` | `0 0 1 * *` | 1st of month, midnight |
| `@weekly` | `0 0 * * 0` | Sunday, midnight |
| `@daily` | `0 0 * * *` | Every day, midnight |
| `@hourly` | `0 * * * *` | Every hour |
| `@every 5m` | - | Every 5 minutes |
| `@every 1h30m` | - | Every 1.5 hours |

### Day Names

Use day names instead of numbers:

```
0 9 * * MON-FRI    # Weekdays at 9am
0 10 * * SUN       # Sundays at 10am
```

### Examples

```bash
# Every 15 minutes
--cron "*/15 * * * *"

# Every hour at minute 30
--cron "30 * * * *"

# Daily at 9am
--cron "0 9 * * *"

# Weekdays at 9am
--cron "0 9 * * MON-FRI"

# First Monday of month at 9am
--cron "0 9 1-7 * MON"

# Every 6 hours
--cron "0 */6 * * *"
```

## Watch Triggers

### Basic Configuration

```bash
ayo trigger schedule file-watcher \
  --watch ./src \
  --agent @watcher \
  --prompt "Files changed"
```

### Advanced Options

```bash
ayo trigger schedule code-watcher \
  --watch ./src \
  --pattern "*.go,*.ts" \
  --exclude "*.test.*" \
  --events create,modify \
  --debounce 5s \
  --singleton \
  --agent @reviewer \
  --prompt "Review: {{files}}"
```

### Watch Options

| Option | Description | Example |
|--------|-------------|---------|
| `--watch` | Directory to watch | `./src` |
| `--pattern` | Glob patterns to include | `*.go,*.ts` |
| `--exclude` | Glob patterns to exclude | `*.test.*` |
| `--events` | Event types | `create,modify,delete` |
| `--debounce` | Wait for changes to settle | `5s` |
| `--singleton` | Prevent overlapping runs | - |

### Event Types

| Event | Description |
|-------|-------------|
| `create` | File created |
| `modify` | File modified |
| `delete` | File deleted |
| `rename` | File renamed |
| `chmod` | Permissions changed |

### Template Variables

Use variables in prompts:

```bash
--prompt "Review {{file}} (event: {{event}})"
```

| Variable | Description |
|----------|-------------|
| `{{file}}` | Changed file path |
| `{{files}}` | All changed files (comma-separated) |
| `{{event}}` | Event type |
| `{{dir}}` | Watched directory |

## YAML Configuration

### Trigger File

`~/.config/ayo/triggers/daily-standup.yaml`:

```yaml
name: daily-standup
type: cron
enabled: true
agent: "@standup"
prompt: "Run daily standup"
config:
  schedule: "0 9 * * MON-FRI"
  singleton: true
```

### Watch Trigger YAML

```yaml
name: code-watcher
type: watch
enabled: true
agent: "@reviewer"
prompt: "Review changed files: {{files}}"
config:
  path: ./src
  patterns:
    - "*.go"
    - "*.ts"
  exclude:
    - "*_test.go"
    - "*.test.ts"
  events:
    - create
    - modify
  debounce: 5s
  singleton: true
```

### Full Schema

```yaml
name: string                # Trigger identifier
type: string                # cron, watch, interval, once
enabled: boolean            # Enable/disable (default: true)
agent: string               # Agent handle
prompt: string              # Prompt to send
config:
  # For cron:
  schedule: string          # Cron expression
  
  # For interval:
  interval: string          # Duration (e.g., "30m", "1h")
  
  # For once:
  at: string                # ISO 8601 datetime
  
  # For watch:
  path: string              # Directory to watch
  patterns: string[]        # Include patterns
  exclude: string[]         # Exclude patterns
  events: string[]          # Event types
  debounce: string          # Debounce duration
  
  # Common:
  singleton: boolean        # Prevent overlapping
  timeout: string           # Max execution time
  retries: integer          # Retry on failure
```

## Agent-Level Triggers

Define triggers in agent config.json:

```json
{
  "triggers": [
    {
      "name": "morning-check",
      "type": "cron",
      "schedule": "0 9 * * *",
      "prompt": "Run morning checks"
    },
    {
      "name": "file-watch",
      "type": "watch",
      "path": "./src",
      "pattern": "*.go",
      "prompt": "Review changes"
    }
  ]
}
```

## Squad-Level Triggers

Define triggers in squad ayo.json:

```json
{
  "triggers": [
    {
      "name": "daily-standup",
      "type": "cron",
      "schedule": "0 9 * * MON-FRI",
      "agent": "@lead",
      "prompt": "Run daily standup"
    }
  ]
}
```

## Execution Behavior

### Singleton Mode

Prevent overlapping executions:

```yaml
config:
  singleton: true
```

If previous execution is still running, new trigger is skipped.

### Timeout

Limit execution time:

```yaml
config:
  timeout: "5m"
```

Agent is terminated after timeout.

### Retries

Retry on failure:

```yaml
config:
  retries: 3
  retry_delay: "30s"
```

### Error Handling

Failed triggers:
- Logged to daemon log
- Visible in `ayo trigger history --status failed`
- Don't block future executions

## Persistence

### SQLite Storage

Triggers persist in:
```
~/.local/share/ayo/daemon.db
```

Tables:
- `triggers` - Trigger configuration
- `trigger_history` - Execution history

### Hot Reload

YAML triggers in `~/.config/ayo/triggers/` are watched:
- New files: Trigger added
- Modified files: Trigger updated
- Deleted files: Trigger removed

## Troubleshooting

### Trigger not firing

```bash
# Check daemon running
ayo service status

# Check trigger enabled
ayo trigger show <name> | grep enabled

# Check schedule
ayo trigger show <name> | grep schedule

# View logs
tail -f ~/.local/share/ayo/daemon.log
```

### Watch trigger missing events

File system watch limitations:
- Some editors use atomic saves
- Network filesystems may not trigger
- Recursive depth may be limited

Try increasing debounce:
```bash
ayo trigger schedule watcher --watch ./src --debounce 3s ...
```

### Cron trigger wrong timezone

Uses system timezone:
```bash
date +%Z
```

### Overlapping executions

Enable singleton mode:
```bash
ayo trigger schedule job --singleton ...
```
