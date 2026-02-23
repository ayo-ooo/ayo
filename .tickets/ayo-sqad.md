---
id: ayo-sqad
status: open
deps: [ayo-pv3a]
links: []
created: 2026-02-23T22:14:55Z
type: epic
priority: 0
assignee: Alex Cabrera
tags: [gtm, phase4]
---
# Phase 4: Advanced Scheduler

Migrate from robfig/cron to go-co-op/gocron v2 for powerful, persistent scheduling.

## Goals

- Replace robfig/cron with gocron v2 in trigger engine
- Add job persistence in SQLite (survive daemon restarts)
- Support one-time jobs (`OneTimeJob`)
- Support duration-based jobs ("every 5m")
- Support daily/weekly/monthly with friendly syntax
- Add job monitoring and history
- Add file watch triggers
- Add trigger notification system

## Why gocron v2?

| Feature | robfig/cron | gocron v2 |
|---------|-------------|------------|
| Cron expressions | ✓ | ✓ |
| Duration-based | ✗ | ✓ |
| One-time jobs | ✗ | ✓ |
| Daily/Weekly/Monthly | ✗ | ✓ (fluent API) |
| Singleton mode | ✗ | ✓ |
| Event listeners | ✗ | ✓ |
| Distributed locking | ✗ | ✓ |

## New Trigger Types

```yaml
# One-time job
type: once
schedule:
  at: "2026-02-24T14:00:00Z"

# Duration-based
type: interval
schedule:
  every: 5m
singleton: true

# Daily with friendly syntax
type: daily
schedule:
  times: ["09:00"]
  days: [monday, tuesday, wednesday, thursday, friday]
```

## Child Tickets

- Replace robfig/cron with gocron v2
- Add job persistence in SQLite
- Implement one-time job support
- Implement duration/interval jobs
- Implement daily/weekly/monthly jobs
- Add job monitoring interface
- `ayo-o841`: Implement file watch triggers
- `ayo-8t7z`: Add trigger notification system
- `ayo-zn5p`: Add trigger management CLI commands
- `ayo-jj2s`: Polish cron trigger configuration
- `ayo-7xsf`: Add trigger YAML configuration

