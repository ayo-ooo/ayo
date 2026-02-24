---
id: ayo-o841
status: open
deps: [ayo-q841]
links: []
created: 2026-02-23T22:16:02Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-sqad
tags: [triggers, watch]
---
# Implement file watch triggers

Polish fsnotify-based file watching in the trigger engine. Support glob patterns, recursive watching, debouncing, and configurable event types.

## Context

File watch triggers invoke agents when files change. This is useful for:
- Auto-linting on save
- Test runner on code changes
- Documentation generation on markdown updates

## Trigger Configuration

```yaml
name: auto-lint
type: watch
agent: "@linter"
watch:
  paths:
    - "~/Projects/myapp/src"
  patterns:
    - "*.go"
    - "*.ts"
  exclude:
    - "*_test.go"
    - "node_modules/**"
  events: [create, modify]  # Default: [create, modify, delete]
  recursive: true           # Default: true

options:
  debounce: "500ms"         # Wait for burst of changes
  singleton: true           # Only one run at a time

prompt: |
  Lint the changed files: {{.ChangedFiles}}
```

## Event Types

| Event | Description |
|-------|-------------|
| `create` | New file created |
| `modify` | Existing file modified |
| `delete` | File deleted |
| `rename` | File renamed |

## Debouncing

When multiple files change rapidly (e.g., git checkout), debounce to avoid multiple triggers:

```go
type FileWatcher struct {
    debounceTimer *time.Timer
    pendingEvents []fsnotify.Event
    debounceWait  time.Duration
}

func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
    fw.mu.Lock()
    defer fw.mu.Unlock()
    
    fw.pendingEvents = append(fw.pendingEvents, event)
    
    if fw.debounceTimer != nil {
        fw.debounceTimer.Stop()
    }
    
    fw.debounceTimer = time.AfterFunc(fw.debounceWait, func() {
        fw.trigger(fw.pendingEvents)
        fw.pendingEvents = nil
    })
}
```

## Glob Pattern Matching

Use `doublestar` library for glob patterns:

```go
import "github.com/bmatcuk/doublestar/v4"

func (fw *FileWatcher) matchesPattern(path string) bool {
    for _, pattern := range fw.config.Watch.Patterns {
        if matched, _ := doublestar.Match(pattern, filepath.Base(path)); matched {
            return true
        }
    }
    return false
}
```

## Template Variables

Pass changed file info to agent prompt:

| Variable | Description |
|----------|-------------|
| `{{.ChangedFiles}}` | List of changed file paths |
| `{{.EventType}}` | create, modify, delete |
| `{{.WatchPath}}` | Root watch path |

## Files to Create/Modify

1. **`internal/daemon/file_watcher.go`** (new) - File watching implementation
2. **`internal/daemon/trigger_engine.go`** - Integrate file watcher
3. **`internal/daemon/trigger_config.go`** - Add Watch config struct

## Implementation

```go
// internal/daemon/file_watcher.go
type FileWatcher struct {
    watcher  *fsnotify.Watcher
    config   WatchConfig
    engine   *TriggerEngine
    debounce time.Duration
}

func (fw *FileWatcher) Start() error {
    for _, path := range fw.config.Paths {
        if fw.config.Recursive {
            filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
                if d.IsDir() {
                    fw.watcher.Add(p)
                }
                return nil
            })
        } else {
            fw.watcher.Add(path)
        }
    }
    
    go fw.watchLoop()
    return nil
}
```

## Acceptance Criteria

- [ ] Watch triggers fire on file changes
- [ ] Glob patterns filter which files trigger
- [ ] Exclude patterns work
- [ ] Recursive watching works
- [ ] Event type filtering works
- [ ] Debouncing prevents burst triggers
- [ ] Singleton mode prevents overlapping runs
- [ ] Changed file list passed to agent prompt

## Testing

- Test single file change triggers
- Test glob pattern matching
- Test exclude patterns
- Test recursive vs non-recursive
- Test debouncing with rapid changes
- Test event type filtering
- Test template variable substitution
