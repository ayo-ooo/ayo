---
id: ayo-pltg
status: open
deps: [ayo-sqad]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-plug
tags: [plugins, triggers]
---
# Task: Trigger Plugin Architecture

## Summary

Refactor the trigger system to support pluggable trigger types. Currently triggers (cron, watch, webhook) are hardcoded in `internal/daemon/trigger_engine.go`. This task creates a plugin interface for custom trigger types like IMAP, calendar events, database changes, etc.

## Context

### Current Trigger Architecture

```
internal/daemon/trigger_engine.go:
├── TriggerEngine struct
├── RegisterTrigger() - adds to triggers map
├── cronRunner() - hardcoded cron handling
├── watchRunner() - hardcoded file watch handling
└── webhookRunner() - hardcoded webhook handling
```

All trigger types are implemented inline with no extension point.

### Desired Architecture

```
internal/triggers/
├── interface.go      # TriggerPlugin interface
├── registry.go       # Trigger type registry
├── builtin/
│   ├── cron.go       # Cron trigger (refactored)
│   ├── watch.go      # File watch trigger (refactored)
│   └── webhook.go    # Webhook trigger (refactored)
└── loader.go         # External trigger loading
```

## Technical Approach

### Trigger Plugin Interface

```go
package triggers

type TriggerType string

const (
    TriggerTypePoll  TriggerType = "poll"   // Periodic polling
    TriggerTypePush  TriggerType = "push"   // Event-driven
    TriggerTypeWatch TriggerType = "watch"  // File system watching
)

type TriggerPlugin interface {
    // Metadata
    Name() string
    Type() TriggerType
    Description() string
    
    // Lifecycle
    Init(ctx context.Context, config map[string]any) error
    Start(ctx context.Context, callback EventCallback) error
    Stop() error
    
    // Health
    Status() TriggerStatus
}

type EventCallback func(event TriggerEvent) error

type TriggerEvent struct {
    TriggerName string
    Type        string
    Payload     map[string]any
    Timestamp   time.Time
}

type TriggerStatus struct {
    Running    bool
    LastEvent  time.Time
    ErrorCount int
    LastError  string
}
```

### Plugin Manifest Entry

```json
{
  "components": {
    "triggers": {
      "imap": {
        "path": "bin/imap-trigger",
        "type": "poll",
        "config_schema": {
          "server": { "type": "string", "required": true },
          "username": { "type": "string", "required": true },
          "password": { "type": "string", "secret": true },
          "folder": { "type": "string", "default": "INBOX" },
          "poll_interval": { "type": "duration", "default": "1m" }
        },
        "description": "Triggers on new IMAP emails"
      }
    }
  }
}
```

### Trigger Configuration

```yaml
# ~/.config/ayo/triggers/email-responder.yaml
name: email-responder
type: imap  # Plugin-provided trigger type
config:
  server: imap.gmail.com
  username: ${IMAP_USER}
  password: ${IMAP_PASSWORD}
  folder: INBOX
  filter:
    subject_contains: "urgent"
agent: "@email-handler"
prompt_template: |
  New email received:
  From: {{.From}}
  Subject: {{.Subject}}
  Body: {{.Body}}
  
  Please draft a response.
```

### Trigger Registry

```go
type TriggerRegistry struct {
    builtins map[string]TriggerFactory
    external map[string]TriggerFactory
}

type TriggerFactory func(config map[string]any) (TriggerPlugin, error)

func (r *TriggerRegistry) Register(name string, factory TriggerFactory) error
func (r *TriggerRegistry) Create(name string, config map[string]any) (TriggerPlugin, error)
func (r *TriggerRegistry) List() []string
```

## Implementation Steps

1. [ ] Create `internal/triggers/` package with interface
2. [ ] Refactor cron trigger to implement interface
3. [ ] Refactor watch trigger to implement interface
4. [ ] Refactor webhook trigger to implement interface
5. [ ] Create trigger registry
6. [ ] Update TriggerEngine to use registry
7. [ ] Implement external trigger loading (exec-based)
8. [ ] Add trigger config schema validation
9. [ ] Update CLI commands (`ayo trigger add/list/remove`)
10. [ ] Add tests for trigger plugin loading
11. [ ] Document trigger plugin development

## Dependencies

- Depends on: `ayo-sqad` (scheduler improvements)
- Blocks: `ayo-tgimap`, `ayo-tgcal` (specific trigger plugins)

## Acceptance Criteria

- [ ] Existing triggers (cron, watch, webhook) work unchanged
- [ ] Plugins can register new trigger types
- [ ] Trigger configs validate against plugin schemas
- [ ] External triggers communicate via stdin/stdout JSON
- [ ] `ayo trigger types` lists available trigger types
- [ ] Documentation covers trigger plugin development

## Files to Create

- `internal/triggers/interface.go`
- `internal/triggers/registry.go`
- `internal/triggers/builtin/cron.go`
- `internal/triggers/builtin/watch.go`
- `internal/triggers/builtin/webhook.go`
- `internal/triggers/loader.go`

## Files to Modify

- `internal/daemon/trigger_engine.go` - Use registry
- `internal/plugins/manifest.go` - Add triggers component
- `internal/plugins/registry.go` - Register triggers
- `cmd/ayo/trigger.go` - Add types command

## External Trigger Protocol

For cross-platform compatibility, external triggers use exec + JSON:

```
Host → Trigger:  {"action": "init", "config": {...}}
Trigger → Host:  {"status": "ready"}

Host → Trigger:  {"action": "start"}
Trigger → Host:  {"event": {...}}  # On each trigger
Trigger → Host:  {"event": {...}}  # On each trigger

Host → Trigger:  {"action": "stop"}
Trigger → Host:  {"status": "stopped"}
```

This avoids Go plugin limitations (.so files, same Go version).

---

*Created: 2026-02-23*
