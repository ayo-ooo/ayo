---
id: ayo-6lcg
status: open
deps: [ayo-q841, ayo-wt6w, ayo-rptd, ayo-snst, ayo-y0x1, ayo-899j]
links: []
created: 2026-02-24T01:02:53Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-sqad
tags: [verification, e2e]
---
# Phase 4 E2E verification

End-to-end verification of advanced scheduler with gocron v2.

## Prerequisites

All Phase 4 tickets complete:
- gocron migration (ayo-q841)
- SQLite persistence (ayo-wt6w)
- One-time jobs (ayo-rptd)
- Interval jobs (ayo-snst)
- Daily/weekly/monthly (ayo-y0x1)
- Job monitoring (ayo-899j)

## Verification Checklist

### Cron Jobs
- [ ] Standard cron expression works (`0 9 * * *`)
- [ ] Job fires at scheduled time
- [ ] Multiple cron jobs run independently

### One-Time Jobs
- [ ] `ayo trigger once --at "2026-02-24T14:00:00Z" ...` creates job
- [ ] Job fires at specified time
- [ ] Job removed from DB after execution

### Interval Jobs
- [ ] `every: 5m` syntax works
- [ ] Job fires at specified interval
- [ ] `singleton: true` prevents overlap

### Daily/Weekly/Monthly
- [ ] `times: ["09:00"]` syntax works
- [ ] `days: [monday, friday]` works
- [ ] Weekly/monthly variants work

### Persistence
- [ ] Jobs stored in `~/.local/share/ayo/jobs.db`
- [ ] Daemon restart restores all jobs
- [ ] Jobs resume at correct next run time

### Monitoring
- [ ] `ayo trigger list` shows all jobs
- [ ] `ayo trigger status <id>` shows details
- [ ] Last run time and status visible
- [ ] Run history accessible

### File Watch Triggers
- [ ] File changes trigger agent execution
- [ ] Debouncing prevents rapid re-triggers
- [ ] Glob patterns work

## Acceptance Criteria

Ambient agents work reliably. Jobs survive daemon restarts.
