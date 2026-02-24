---
id: ayo-snst
status: closed
deps: [ayo-q841]
links: []
created: 2026-02-23T23:15:51Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [scheduler]
---
# Implement duration/interval jobs

Add support for duration-based scheduling using gocron's `DurationJob`. Jobs execute at regular intervals (e.g., every 5 minutes).

## Context

After migrating to gocron v2 (ayo-q841), we can support interval-based jobs. This is simpler than cron for common patterns:
- "Run every 5 minutes"
- "Run every hour"
- "Run once a day"

## Trigger Configuration

```yaml
# ~/.config/ayo/triggers/health-check.yaml
name: health-check
type: interval
agent: "@monitor"
schedule:
  every: "5m"         # Duration: s, m, h, d
options:
  singleton: true     # Prevent overlapping runs
  start_immediately: false  # Wait for first interval
prompt: "Check system health and report issues"
```

## Duration Units

| Unit | Example | Description |
|------|---------|-------------|
| `s` | `30s` | Seconds |
| `m` | `5m` | Minutes |
| `h` | `1h` | Hours |
| `d` | `1d` | Days (24h) |

Combined: `1h30m`, `2d12h`

## gocron Implementation

```go
// internal/daemon/trigger_engine.go
func (te *TriggerEngine) createIntervalJob(cfg TriggerConfig) error {
    duration, err := time.ParseDuration(cfg.Schedule.Every)
    if err != nil {
        return fmt.Errorf("invalid schedule.every: %w", err)
    }
    
    opts := []gocron.JobOption{}
    if cfg.Options.Singleton {
        opts = append(opts, gocron.WithSingletonMode(gocron.LimitModeReschedule))
    }
    
    jobDef := gocron.DurationJob(duration)
    if cfg.Options.StartImmediately {
        opts = append(opts, gocron.WithStartAt(gocron.WithStartImmediately()))
    }
    
    job, err := te.scheduler.NewJob(
        jobDef,
        gocron.NewTask(func() { te.executeAgent(cfg) }),
        opts...,
    )
    if err != nil {
        return err
    }
    
    return te.storeJobInDB(cfg, job.ID())
}
```

## Singleton Mode

When `singleton: true`:
- If job is still running when next interval fires, skip the new run
- Prevents overlapping executions for long-running agents
- Uses `LimitModeReschedule` to reschedule the missed run

## Files to Modify

1. **`internal/daemon/trigger_engine.go`** - Add interval job creation
2. **`internal/daemon/trigger_config.go`** - Add `Schedule.Every` and `Options` fields
3. **`cmd/ayo/trigger.go`** - Update help text with interval examples

## Acceptance Criteria

- [ ] Interval jobs execute at regular intervals
- [ ] Duration parsing works for s, m, h, d units
- [ ] Combined durations work (`1h30m`)
- [ ] Singleton mode prevents overlapping runs
- [ ] `start_immediately` option works
- [ ] Jobs survive daemon restart
- [ ] Invalid durations show clear error

## Testing

- Test various duration formats
- Test singleton mode prevents overlap
- Test start_immediately vs waiting for first interval
- Test daemon restart preserves interval jobs
- Test invalid duration error messages
