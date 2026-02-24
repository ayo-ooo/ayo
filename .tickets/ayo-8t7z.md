---
id: ayo-8t7z
status: closed
deps: [ayo-899j]
links: []
created: 2026-02-23T22:16:02Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [triggers, notification]
---
# Add trigger notification system

When ambient agents complete triggered work, notify users appropriately. Support multiple notification channels and persistent notification storage.

## Context

Ambient agents run in the background via triggers. When they complete, users need to know:
- What work was done
- Whether it succeeded or failed
- Where to find outputs

## Notification Channels

### Terminal (Active Session)

If user has an active ayo session:
```
┌─ Trigger Completed ─────────────────────────┐
│ ✓ health-check completed (12s)              │
│                                             │
│ No issues found                             │
│                                             │
│ [View Output]  [Dismiss]                    │
└─────────────────────────────────────────────┘
```

### Stored Notifications

If no active session, store for next time:
```bash
ayo
# 📬 2 notifications
# - health-check completed 5m ago ✓
# - weekly-summary failed 2h ago ✗
# 
# Run 'ayo notifications' to view details
```

### System Notifications (Optional)

On macOS, use `terminal-notifier` or native notifications:
```bash
# If enabled in config
osascript -e 'display notification "health-check completed" with title "ayo"'
```

## Notification Storage

```sql
-- ~/.local/share/ayo/ayo.db
CREATE TABLE notifications (
    id INTEGER PRIMARY KEY,
    trigger_name TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    read_at DATETIME,
    status TEXT NOT NULL,  -- 'success', 'failed'
    summary TEXT,
    output_path TEXT,
    FOREIGN KEY (trigger_name) REFERENCES jobs(name)
);
```

## CLI Commands

```bash
# List unread notifications
ayo notifications
# 📬 2 unread notifications
# 
# 1. ✓ health-check (5m ago)
#    No issues found
# 
# 2. ✗ weekly-summary (2h ago)
#    Error: Connection timeout

# Mark as read
ayo notifications clear

# View notification details
ayo notifications show 1
```

## Configuration

```json
// ~/.config/ayo/config.json
{
  "notifications": {
    "terminal": true,      // Show in-terminal notifications
    "system": false,       // Use system notifications
    "store": true,         // Store for later viewing
    "sound": false         // Play sound on completion
  }
}
```

## Implementation

```go
// internal/daemon/notifications.go
type NotificationService struct {
    db     *sql.DB
    config NotificationConfig
}

func (ns *NotificationService) Notify(trigger string, status string, summary string, outputPath string) error {
    // Store in DB
    if ns.config.Store {
        ns.storeNotification(trigger, status, summary, outputPath)
    }
    
    // Terminal notification (via RPC to active sessions)
    if ns.config.Terminal {
        ns.notifyActiveSessions(trigger, status, summary)
    }
    
    // System notification
    if ns.config.System {
        ns.systemNotify(trigger, status, summary)
    }
    
    return nil
}
```

## Files to Create/Modify

1. **`internal/daemon/notifications.go`** (new) - Notification service
2. **`internal/daemon/db.go`** - Add notifications table
3. **`cmd/ayo/notifications.go`** (new) - CLI commands
4. **`internal/config/config.go`** - Add notification config
5. **`internal/ui/notification.go`** (new) - Terminal notification UI

## Acceptance Criteria

- [ ] Notifications shown in active terminal sessions
- [ ] Notifications stored in DB when no active session
- [ ] Unread count shown on ayo startup
- [ ] `notifications` command lists stored notifications
- [ ] `notifications clear` marks as read
- [ ] System notifications work on macOS (if enabled)
- [ ] Output path included for viewing results

## Testing

- Test terminal notification display
- Test notification storage
- Test unread count on startup
- Test clear command
- Test system notifications (macOS)
