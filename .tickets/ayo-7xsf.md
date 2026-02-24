---
id: ayo-7xsf
status: open
deps: [ayo-q841]
links: []
created: 2026-02-23T22:16:10Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [triggers, config]
---
# Add trigger YAML configuration

Support defining triggers in YAML files at `~/.config/ayo/triggers/*.yaml`. Auto-discover and register triggers on daemon start. Support hot-reload when files change.

## Context

Currently triggers must be registered via CLI commands. This ticket adds declarative YAML configuration, allowing users to define triggers in config files that are automatically loaded.

## Trigger Directory

```
~/.config/ayo/triggers/
├── health-check.yaml
├── morning-report.yaml
└── file-watcher.yaml
```

## YAML Schema

```yaml
# ~/.config/ayo/triggers/health-check.yaml
name: health-check
type: interval          # cron, interval, once, daily, weekly, monthly, watch
agent: "@monitor"
enabled: true           # Default: true

# Type-specific schedule
schedule:
  every: "5m"           # For interval
  # cron: "0 9 * * *"   # For cron
  # at: "2026-02-24..."  # For once
  # times: ["09:00"]    # For daily/weekly/monthly
  # days: [monday]      # For weekly
  # pattern: "*.go"     # For watch

options:
  singleton: true       # Prevent overlapping runs
  timeout: "30m"        # Max execution time
  retry: 3              # Retry on failure

prompt: |
  Check system health and report any issues.
  Focus on CPU, memory, and disk usage.

# Environment variables passed to agent
env:
  ALERT_THRESHOLD: "80"
```

## Auto-Discovery

On daemon start:
1. Scan `~/.config/ayo/triggers/` for `*.yaml` files
2. Parse and validate each trigger definition
3. Register with gocron scheduler
4. Watch directory for changes

## Hot-Reload

```go
// internal/daemon/trigger_loader.go
func (tl *TriggerLoader) watchConfigDir() {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add(triggersDir)
    
    for event := range watcher.Events {
        switch {
        case event.Op&fsnotify.Create != 0:
            tl.loadTrigger(event.Name)
        case event.Op&fsnotify.Write != 0:
            tl.reloadTrigger(event.Name)
        case event.Op&fsnotify.Remove != 0:
            tl.unloadTrigger(event.Name)
        }
    }
}
```

## Files to Create/Modify

1. **`internal/daemon/trigger_loader.go`** (new) - YAML loading and file watching
2. **`internal/daemon/trigger_config.go`** - Unified trigger config struct
3. **`internal/daemon/daemon.go`** - Call trigger loader on start
4. **`docs/triggers.md`** - Document YAML schema

## Validation

On load, validate:
- Required fields (name, type, agent)
- Type-specific schedule fields present
- Agent exists
- Schedule is valid

## Acceptance Criteria

- [ ] YAML files in triggers/ are auto-loaded on daemon start
- [ ] New files are auto-registered (hot-reload)
- [ ] Modified files are reloaded
- [ ] Removed files are unregistered
- [ ] Invalid YAML shows clear error in logs
- [ ] `enabled: false` disables trigger without removing
- [ ] All trigger types work from YAML

## Testing

- Test auto-discovery on daemon start
- Test hot-reload for create/modify/delete
- Test validation error messages
- Test enabled/disabled toggle
- Test all trigger types from YAML
