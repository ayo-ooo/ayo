# ayo trigger

Manage triggers—scheduled tasks and file watchers that invoke agents.

## Synopsis

```
ayo trigger <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List all triggers |
| `create` | Create a trigger |
| `show` | Show trigger details |
| `enable` | Enable a trigger |
| `disable` | Disable a trigger |
| `delete` | Delete a trigger |
| `run` | Run a trigger manually |

---

## ayo trigger list

List all triggers.

```bash
$ ayo trigger list
NAME            TYPE      STATUS    NEXT RUN
daily-backup    cron      enabled   2026-02-13 02:00:00
file-sync       watch     enabled   (watching)
weekly-report   cron      disabled  -
```

---

## ayo trigger create

Create a new trigger.

### Cron Trigger

```bash
$ ayo trigger create <name> --cron <schedule> --agent <agent> --prompt <prompt>
```

### File Watch Trigger

```bash
$ ayo trigger create <name> --watch <path> --agent <agent> --prompt <prompt>
```

### Examples

```bash
# Daily at 9am
$ ayo trigger create morning-standup \
  --cron "0 9 * * *" \
  --agent @assistant \
  --prompt "Summarize today's calendar and priorities"

# Watch for new files
$ ayo trigger create process-uploads \
  --watch ~/Downloads/*.csv \
  --agent @analyzer \
  --prompt "Process the new CSV file: {{.path}}"
```

---

## ayo trigger show

Show trigger details.

```bash
$ ayo trigger show <name>
```

### Example

```bash
$ ayo trigger show daily-backup
Name:      daily-backup
Type:      cron
Schedule:  0 2 * * *
Agent:     @backup
Prompt:    Run daily backup procedure
Status:    enabled
Last Run:  2026-02-12T02:00:00Z
Next Run:  2026-02-13T02:00:00Z
```

---

## ayo trigger enable/disable

Enable or disable a trigger.

```bash
$ ayo trigger enable <name>
$ ayo trigger disable <name>
```

---

## ayo trigger run

Run a trigger manually (for testing).

```bash
$ ayo trigger run <name>
```

---

## ayo trigger delete

Delete a trigger.

```bash
$ ayo trigger delete <name>
```

---

## Cron Schedule Format

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-6, Sun=0)
│ │ │ │ │
* * * * *
```

Examples:
- `0 9 * * *` - 9am daily
- `0 */2 * * *` - Every 2 hours
- `0 9 * * 1-5` - 9am weekdays
- `0 0 1 * *` - Midnight on 1st of month

## Watch Variables

File watch triggers have access to:

| Variable | Description |
|----------|-------------|
| `{{.path}}` | Full path to changed file |
| `{{.name}}` | File name only |
| `{{.dir}}` | Directory containing file |
| `{{.event}}` | Event type (create, modify, delete) |
