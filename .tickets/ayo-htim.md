---
id: ayo-htim
status: open
deps: [ayo-hscm]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-hitl
tags: [human-in-the-loop, timeout]
---
# Task: Timeout and Fallback Handling

## Summary

Implement timeout handling for human input requests. When a human doesn't respond in time, the agent needs graceful fallback options rather than hard failure.

## Timeout Configuration

```json
{
  "timeout": "1h",
  "fallback": {
    "action": "default|error|retry|escalate",
    "default_values": {...},
    "escalation_recipient": "..."
  }
}
```

### Fallback Actions

| Action | Behavior |
|--------|----------|
| `error` | Return error to agent (default) |
| `default` | Use default values from schema |
| `retry` | Re-send request (with limit) |
| `escalate` | Send to different recipient |
| `skip` | Skip the request, continue |

## Implementation

### TimeoutHandler

```go
type TimeoutHandler struct {
    defaultTimeout time.Duration
}

func (h *TimeoutHandler) Run(ctx context.Context, req *InputRequest, renderer FormRenderer) (*InputResponse, error) {
    timeout := req.Timeout
    if timeout == 0 {
        timeout = h.defaultTimeout
    }
    
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    resp, err := renderer.Render(ctx, req)
    if errors.Is(err, context.DeadlineExceeded) {
        return h.handleTimeout(req)
    }
    return resp, err
}

func (h *TimeoutHandler) handleTimeout(req *InputRequest) (*InputResponse, error) {
    switch req.Fallback.Action {
    case "default":
        return buildDefaultResponse(req)
    case "escalate":
        return h.escalate(req)
    case "retry":
        return nil, &RetryError{...}
    default:
        return nil, &TimeoutError{...}
    }
}
```

### Reminder System

For long timeouts (>1h), send reminders:
- 50% of timeout: First reminder
- 90% of timeout: Final reminder

```go
func (h *TimeoutHandler) scheduleReminders(req *InputRequest) {
    halfTime := req.Timeout / 2
    nearEnd := req.Timeout * 9 / 10
    
    time.AfterFunc(halfTime, func() {
        h.sendReminder(req, "halfway")
    })
    time.AfterFunc(nearEnd, func() {
        h.sendReminder(req, "final")
    })
}
```

## Files to Create

- `internal/hitl/timeout.go` - Timeout handler
- `internal/hitl/timeout_test.go` - Tests

## Acceptance Criteria

- [ ] Timeout respects configured duration
- [ ] Default fallback uses schema defaults
- [ ] Escalation sends to new recipient
- [ ] Retry respects max attempts
- [ ] Reminders sent for long timeouts
- [ ] Context cancellation handled properly
