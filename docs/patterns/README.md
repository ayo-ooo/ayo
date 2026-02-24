# Ambient Agent Patterns

Ayo supports "ambient agents" - agents that run automatically in response to events, schedules, or conditions. This section documents common patterns for setting up effective ambient agents.

## What Are Ambient Agents?

Unlike interactive agents that respond to direct prompts, ambient agents operate autonomously:

- **Triggered automatically** by file changes, schedules, or events
- **Run in the background** without user intervention
- **Perform routine tasks** like monitoring, formatting, or reporting

## Core Patterns

| Pattern | Use Case | Trigger Type |
|---------|----------|--------------|
| [Watcher](watcher.md) | React to file changes | `watch` |
| [Scheduled](scheduled.md) | Run at specific times | `schedule` |
| [Ticket Worker](ticket-worker.md) | Process work queues | `interval` |
| [Monitor](monitor.md) | System health checks | `interval` |

## Quick Examples

### File Watcher
```bash
ayo triggers watch ~/Code/myproject @linter \
  --prompt "Lint changed files and fix issues" \
  --pattern "*.go"
```

### Scheduled Report
```bash
ayo triggers schedule @reporter "0 9 * * *" \
  --prompt "Generate morning summary of git activity"
```

### Ticket Processor
```bash
ayo triggers schedule @worker "*/5 * * * *" \
  --prompt "Check for ready tickets and work on highest priority"
```

## Key Concepts

### Singleton Mode

Prevents multiple instances of a trigger from running simultaneously:

```bash
ayo triggers schedule @worker "* * * * *" \
  --prompt "Process queue" \
  --singleton
```

Without singleton mode, if a trigger takes longer than its interval, multiple instances may overlap.

### Debouncing

For file watchers, debouncing groups rapid changes into a single trigger:

```bash
ayo triggers watch ~/Code @formatter \
  --prompt "Format changed files" \
  --debounce 500ms
```

This prevents triggering on every keystroke during active editing.

### Timeouts

Set maximum execution time to prevent runaway agents:

```bash
ayo triggers schedule @analyzer "0 * * * *" \
  --prompt "Analyze codebase" \
  --timeout 10m
```

## Best Practices

1. **Use singleton mode** for tasks that shouldn't overlap
2. **Set appropriate timeouts** to prevent resource exhaustion
3. **Use debouncing** for file watchers (500ms-2s typical)
4. **Output to workspace** not to shared directories
5. **Log progress** so you can troubleshoot issues
6. **Start simple** and add complexity as needed

## Managing Triggers

```bash
# List all triggers
ayo triggers list

# View trigger details
ayo triggers show <id>

# Disable without removing
ayo triggers disable <id>

# Remove trigger
ayo triggers rm <id>
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Trigger not firing | Check `ayo triggers list` status |
| Too many executions | Increase debounce/interval |
| Agent errors | Check session history for details |
| High resource usage | Reduce trigger frequency or add singleton |

## Next Steps

- [Watcher Pattern](watcher.md) - React to file changes
- [Scheduled Pattern](scheduled.md) - Time-based automation
- [Ticket Worker Pattern](ticket-worker.md) - Process work queues
- [Monitor Pattern](monitor.md) - System health monitoring
