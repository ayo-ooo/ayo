---
id: ayo-q841
status: open
deps: []
links: []
created: 2026-02-23T23:15:39Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-sqad
tags: [scheduler, daemon]
---
# Replace robfig/cron with gocron v2

Migrate from robfig/cron to go-co-op/gocron/v2 for advanced scheduling.

## Why gocron v2?

| Feature | robfig/cron | gocron v2 |
|---------|-------------|-----------|
| Cron expressions | ✓ | ✓ |
| Duration-based (`10s`, `5m`) | ✗ | ✓ |
| One-time jobs | ✗ | ✓ (`OneTimeJob`) |
| Daily/Weekly/Monthly | ✗ | ✓ (fluent API) |
| Singleton mode | ✗ | ✓ |
| Event listeners | ✗ | ✓ |
| Distributed locking | ✗ | ✓ |

## Implementation

### Files to Modify

1. `go.mod` - Add `github.com/go-co-op/gocron/v2`
2. `internal/daemon/trigger_engine.go` - Replace scheduler

### Current Code (robfig/cron)

```go
import "github.com/robfig/cron/v3"

c := cron.New(cron.WithSeconds())
c.AddFunc(schedule, func() { ... })
c.Start()
```

### New Code (gocron v2)

```go
import "github.com/go-co-op/gocron/v2"

s, _ := gocron.NewScheduler()
s.NewJob(
    gocron.CronJob(schedule, false),
    gocron.NewTask(func() { ... }),
)
s.Start()
```

### Migration Steps

1. Add gocron/v2 to go.mod
2. Create new scheduler initialization in trigger engine
3. Migrate existing cron triggers to gocron CronJob
4. Remove robfig/cron from go.mod
5. Test all existing trigger functionality

### Backward Compatibility

- Existing cron expressions (`0 9 * * *`) work unchanged
- Existing trigger YAML files work unchanged
- No user-facing changes in this ticket

## Testing

- Verify all existing cron triggers work
- Test scheduler start/stop lifecycle
- Test multiple concurrent jobs
- Verify daemon restart doesn't lose jobs (covered by persistence ticket)
