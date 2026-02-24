---
id: ayo-y0x1
status: open
deps: [ayo-q841]
links: []
created: 2026-02-23T23:15:56Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [scheduler]
---
# Implement daily/weekly/monthly jobs

Add support for friendly schedule syntax using gocron's fluent API (`DailyJob`, `WeeklyJob`, `MonthlyJob`).

## Context

After migrating to gocron v2 (ayo-q841), we can offer user-friendly schedule syntax instead of cron expressions. This is more readable for common patterns.

## Trigger Configuration

### Daily Jobs

```yaml
name: morning-report
type: daily
agent: "@reporter"
schedule:
  times: ["09:00", "17:00"]  # Multiple times per day
  timezone: "America/New_York"
```

### Weekly Jobs

```yaml
name: weekly-summary
type: weekly
agent: "@summarizer"
schedule:
  days: [monday, friday]
  times: ["09:00"]
  timezone: "America/Los_Angeles"
```

### Monthly Jobs

```yaml
name: monthly-audit
type: monthly
agent: "@auditor"
schedule:
  days_of_month: [1, 15]  # 1st and 15th
  times: ["10:00"]
```

## Day Names

Supported day names (case-insensitive):
- `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`, `sunday`
- Shorthand: `mon`, `tue`, `wed`, `thu`, `fri`, `sat`, `sun`

## gocron Implementation

```go
// internal/daemon/trigger_engine.go
func (te *TriggerEngine) createDailyJob(cfg TriggerConfig) error {
    times := make([]gocron.AtTime, len(cfg.Schedule.Times))
    for i, t := range cfg.Schedule.Times {
        parsed, _ := time.Parse("15:04", t)
        times[i] = gocron.NewAtTime(uint(parsed.Hour()), uint(parsed.Minute()), 0)
    }
    
    job, err := te.scheduler.NewJob(
        gocron.DailyJob(1, gocron.NewAtTimes(times...)),
        gocron.NewTask(func() { te.executeAgent(cfg) }),
    )
    return te.storeJobInDB(cfg, job.ID())
}

func (te *TriggerEngine) createWeeklyJob(cfg TriggerConfig) error {
    days := parseDayNames(cfg.Schedule.Days)
    times := parseTimes(cfg.Schedule.Times)
    
    job, err := te.scheduler.NewJob(
        gocron.WeeklyJob(1, gocron.NewWeekdays(days...), gocron.NewAtTimes(times...)),
        gocron.NewTask(func() { te.executeAgent(cfg) }),
    )
    return te.storeJobInDB(cfg, job.ID())
}
```

## Files to Modify

1. **`internal/daemon/trigger_engine.go`** - Add daily/weekly/monthly job creation
2. **`internal/daemon/trigger_config.go`** - Add schedule fields for times, days, etc.
3. **`internal/daemon/schedule_parser.go`** (new) - Parse day names, times, timezones

## Acceptance Criteria

- [ ] Daily jobs run at specified times
- [ ] Weekly jobs run on specified days at specified times
- [ ] Monthly jobs run on specified days of month
- [ ] Multiple times per day supported
- [ ] Timezone configuration works
- [ ] Day name parsing is case-insensitive
- [ ] Invalid day names show clear error
- [ ] Jobs survive daemon restart

## Testing

- Test daily job with multiple times
- Test weekly job with multiple days
- Test monthly job with multiple days of month
- Test timezone handling
- Test invalid day name error
- Test daemon restart preserves jobs
