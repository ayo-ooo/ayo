---
id: ayo-rptd
status: open
deps: [ayo-q841]
links: []
created: 2026-02-23T23:15:47Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [scheduler]
---
# Implement one-time job support

Add support for one-time scheduled jobs using gocron's `OneTimeJob`. These jobs execute at a specific time and are automatically removed after execution.

## Context

After migrating to gocron v2 (ayo-q841), we can support one-time jobs. This is useful for:
- Scheduled reminders
- Delayed task execution
- "Run this agent tomorrow at 9am"

## Trigger Configuration

```yaml
# ~/.config/ayo/triggers/reminder.yaml
name: daily-standup-reminder
type: once
agent: "@reminder"
schedule:
  at: "2026-02-24T14:00:00Z"  # ISO 8601 format
prompt: "Remind me about the daily standup"
```

## gocron Implementation

```go
// internal/daemon/trigger_engine.go
func (te *TriggerEngine) createOneTimeJob(cfg TriggerConfig) error {
    runTime, err := time.Parse(time.RFC3339, cfg.Schedule.At)
    if err != nil {
        return fmt.Errorf("invalid schedule.at: %w", err)
    }
    
    job, err := te.scheduler.NewJob(
        gocron.OneTimeJob(
            gocron.OneTimeJobStartDateTime(runTime),
        ),
        gocron.NewTask(func() {
            te.executeAgent(cfg)
            // Job auto-removes, but we still need to clean up DB
            te.removeJobFromDB(cfg.Name)
        }),
    )
    if err != nil {
        return err
    }
    
    return te.storeJobInDB(cfg, job.ID())
}
```

## CLI Command

```bash
# Create a one-time job
ayo trigger once "@reminder" --at "2026-02-24T14:00:00" --prompt "Check on PR status"

# Or with relative time
ayo trigger once "@backup" --in "2h" --prompt "Run backup"

# List shows pending one-time jobs
ayo trigger list
# NAME                 TYPE     SCHEDULE              STATUS
# daily-standup        once     2026-02-24 14:00      pending
```

## Files to Modify

1. **`internal/daemon/trigger_engine.go`** - Add one-time job creation
2. **`internal/daemon/trigger_config.go`** - Add `Schedule.At` field
3. **`cmd/ayo/trigger_once.go`** (new) - CLI subcommand
4. **`internal/daemon/db.go`** - Handle job cleanup after execution

## Acceptance Criteria

- [ ] One-time jobs execute at specified time
- [ ] Jobs are removed from scheduler after execution
- [ ] Jobs are removed from persistence DB after execution
- [ ] CLI `trigger once` creates one-time jobs
- [ ] Relative time parsing works (`--in 2h`)
- [ ] Past times are rejected with clear error
- [ ] Jobs survive daemon restart (if not yet executed)

## Testing

- Test job executes at correct time
- Test job is removed after execution
- Test past time is rejected
- Test relative time parsing (`1h`, `30m`, `2d`)
- Test daemon restart preserves pending one-time jobs
