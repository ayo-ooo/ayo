# Scheduled Pattern

The scheduled pattern runs agents at specific times or intervals. This is ideal for periodic tasks like reports, backups, or maintenance.

## Overview

```
Time/Interval → Trigger → Agent Runs → Output
                   ↑            |
                   └────────────┘
                    (repeats)
```

## Basic Setup

```bash
# Run every day at 9 AM
ayo triggers schedule @reporter "0 9 * * *" \
  --prompt "Generate a daily summary of project activity"
```

## Cron Syntax

```
┌───────────── minute (0-59)
│ ┌─────────── hour (0-23)
│ │ ┌───────── day of month (1-31)
│ │ │ ┌─────── month (1-12)
│ │ │ │ ┌───── day of week (0-6, Sunday=0)
│ │ │ │ │
* * * * *
```

### Common Patterns

| Pattern | Description |
|---------|-------------|
| `0 9 * * *` | Daily at 9:00 AM |
| `0 9,17 * * *` | Daily at 9 AM and 5 PM |
| `0 9 * * 1-5` | Weekdays at 9 AM |
| `0 * * * *` | Every hour |
| `*/15 * * * *` | Every 15 minutes |
| `0 0 * * 0` | Weekly on Sunday midnight |
| `0 0 1 * *` | Monthly on the 1st |

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--timeout` | Maximum run time | `30m` |
| `--singleton` | Prevent concurrent runs | `false` |
| `--timezone` | Timezone for schedule | System default |

## Examples

### Morning Briefing

Generate a summary every morning:

```bash
ayo triggers schedule @briefer "0 8 * * 1-5" \
  --prompt "Generate a morning briefing including:
    - Git commits from yesterday
    - Open pull requests
    - Calendar events for today
    - Any failing CI builds"
```

### End-of-Day Summary

Wrap up at end of workday:

```bash
ayo triggers schedule @summarizer "0 17 * * 1-5" \
  --prompt "Generate an end-of-day summary:
    - Work completed today
    - Open items to continue tomorrow
    - Any blockers to flag"
```

### Weekly Report

Create weekly status reports:

```bash
ayo triggers schedule @reporter "0 9 * * 1" \
  --prompt "Generate a weekly status report covering:
    - Completed features
    - Open issues
    - Test coverage changes
    - Performance metrics"
```

### Hourly Health Check

Monitor system health:

```bash
ayo triggers schedule @monitor "0 * * * *" \
  --prompt "Check system health and report only if issues found:
    - Disk space below 20%
    - Memory usage above 90%
    - Services not responding"
```

### Backup Reminder

Daily backup prompt:

```bash
ayo triggers schedule @backup "0 23 * * *" \
  --prompt "Verify today's backup completed successfully.
    Check backup logs and report any failures."
```

### Code Quality Scan

Weekly deep analysis:

```bash
ayo triggers schedule @analyzer "0 2 * * 0" \
  --prompt "Run comprehensive code analysis:
    - Security vulnerability scan
    - Dependency updates available
    - Code complexity metrics
    - Test coverage gaps
    Generate a report for Monday review." \
  --timeout 1h
```

## Timezone Handling

Specify timezone for consistent scheduling:

```bash
ayo triggers schedule @reporter "0 9 * * *" \
  --prompt "Daily report" \
  --timezone "America/New_York"
```

## Singleton Mode

For long-running scheduled tasks, use singleton to prevent overlap:

```bash
ayo triggers schedule @analyzer "*/30 * * * *" \
  --prompt "Deep analysis that might take 20+ minutes" \
  --singleton
```

## Agent Prompt Best Practices

### Be Time-Aware

```
Good: "Generate a summary of activity since the last report"
Bad:  "Generate a summary"
```

### Specify Output Format

```
Good: "Create a markdown report and save to workspace/reports/"
Bad:  "Make a report"
```

### Handle No-Activity Cases

```
Good: "Report on changes. If no changes, respond with 'No activity'"
Bad:  "Report on changes" (might generate empty output)
```

## Combining with Other Features

### With Memory

Store recurring context:

```bash
# Store project facts
ayo memory store "Weekly reports go to #status-updates channel"
ayo memory store "Critical issues should mention @oncall"

# Scheduled task uses memory context
ayo triggers schedule @reporter "0 9 * * 1" \
  --prompt "Generate weekly report following team conventions"
```

### With Squads

Coordinate multiple agents:

```bash
# Create reporting squad
ayo squad create reporters -a @code-analyst,@security-scanner,@metrics

# Scheduled dispatch
ayo triggers schedule "#reporters" "0 6 * * 1" \
  --prompt "Generate comprehensive weekly report"
```

## Troubleshooting

### Trigger Not Running at Expected Time

1. Check timezone settings:
   ```bash
   ayo triggers show <id>
   ```

2. Verify daemon is running:
   ```bash
   ayo sandbox service status
   ```

3. Check system time:
   ```bash
   date
   ```

### Runs Taking Too Long

- Add timeout to prevent runaway tasks
- Use singleton mode to prevent overlap
- Break into smaller, more frequent tasks

### Missing Reports

Check session history:
```bash
ayo session list --agent @reporter
```

## Best Practices

1. **Use singleton mode** for tasks over 5 minutes
2. **Set appropriate timeouts** (default 30m may be too long)
3. **Schedule during low-activity hours** for intensive tasks
4. **Include timestamp** in output for tracking
5. **Handle edge cases** (no data, errors, timeouts)

## See Also

- [Triggers Guide](../guides/triggers.md)
- [Watcher Pattern](watcher.md)
- [Ticket Worker Pattern](ticket-worker.md)
