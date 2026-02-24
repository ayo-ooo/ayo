---
id: ayo-1v23
status: open
deps: [ayo-o841, ayo-jj2s, ayo-zn5p]
links: []
created: 2026-02-23T22:16:10Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [triggers, docs]
---
# Document ambient agent patterns

Create comprehensive documentation for common ambient agent patterns. Include example configurations, best practices, and troubleshooting guides.

## Context

After all trigger types are implemented, document how to use them effectively. This helps users understand the "ambient agent" concept and implement common patterns.

## Patterns to Document

### 1. Watcher Agent

Auto-run on file changes:

```yaml
# Example: Auto-linter
name: auto-lint
type: watch
agent: "@linter"
watch:
  paths: ["~/Projects/myapp/src"]
  patterns: ["*.go", "*.ts"]
options:
  debounce: "500ms"
  singleton: true
prompt: "Lint the changed files and fix any issues"
```

Use cases:
- Code formatting on save
- Test runner on code changes
- Documentation generation

### 2. Scheduled Reporter

Generate periodic reports:

```yaml
# Example: Daily summary
name: daily-summary
type: daily
agent: "@summarizer"
schedule:
  times: ["09:00", "17:00"]
  timezone: "America/New_York"
prompt: |
  Generate a summary of:
  - Git commits since last report
  - Open PRs and issues
  - Calendar for today
```

Use cases:
- Morning briefings
- EOD summaries
- Weekly status reports

### 3. Ticket-Driven Worker

Process tickets from a queue:

```yaml
# Example: Ticket processor
name: ticket-worker
type: interval
agent: "@worker"
schedule:
  every: "1m"
options:
  singleton: true
prompt: |
  Check for open tickets assigned to you.
  Pick the highest priority one and work on it.
  When complete, close the ticket and commit your changes.
```

Use cases:
- Background code review
- Automated bug fixes
- Documentation updates

### 4. Health Monitor

Continuous system monitoring:

```yaml
# Example: System health
name: health-check
type: interval
agent: "@monitor"
schedule:
  every: "5m"
options:
  singleton: true
  timeout: "2m"
prompt: |
  Check system health:
  - CPU and memory usage
  - Disk space
  - Running services
  Report only if issues found.
```

## Documentation Structure

```
docs/
├── patterns/
│   ├── README.md          # Overview of patterns
│   ├── watcher.md         # File watcher pattern
│   ├── scheduled.md       # Scheduled jobs pattern
│   ├── ticket-worker.md   # Ticket processing pattern
│   └── monitor.md         # Monitoring pattern
```

## Best Practices Section

- **Singleton mode**: When to use to prevent overlap
- **Debouncing**: How to choose debounce duration
- **Timeouts**: Setting appropriate limits
- **Output handling**: Where agents should write results
- **Error handling**: What happens on failure
- **Resource usage**: Managing multiple ambient agents

## Troubleshooting Guide

Common issues:
- Trigger not firing → check `ayo trigger list`
- Agent errors → check `ayo trigger history`
- Too many runs → increase debounce/interval
- Missing files → verify watch paths

## Files to Create

1. **`docs/patterns/README.md`** - Pattern overview
2. **`docs/patterns/watcher.md`** - Watcher pattern
3. **`docs/patterns/scheduled.md`** - Scheduled pattern
4. **`docs/patterns/ticket-worker.md`** - Ticket worker pattern
5. **`docs/patterns/monitor.md`** - Monitor pattern
6. **`docs/triggers.md`** - Update with links to patterns

## Acceptance Criteria

- [ ] All four patterns documented with examples
- [ ] Best practices section included
- [ ] Troubleshooting guide included
- [ ] Example YAML configs are valid and tested
- [ ] Cross-references to other docs
- [ ] Linked from main docs index
