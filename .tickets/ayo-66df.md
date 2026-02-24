---
id: ayo-66df
status: closed
deps: [ayo-kkxg]
links: []
created: 2026-02-23T22:15:34Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-whmn
tags: [sandbox, filesystem]
---
# Implement /output safe write zone

Create a "safe write zone" where agents can write freely without approval.

## Context

Agents need somewhere to put generated artifacts (reports, code, data) without going through file_request approval for each file. The `/output/` directory provides this.

## Design

```
SANDBOX:
/output/
└── {session_id}/      # Created per-session
    ├── report.md      # Agent-generated files
    ├── analysis.json
    └── code/
        └── main.go

HOST (auto-synced):
~/.local/share/ayo/output/
└── {session_id}/      # Mirror of sandbox /output/
    └── ...
```

## Behavior

1. Session starts → `/output/{session_id}/` created in sandbox
2. Agent writes files to `/output/` freely (no approval needed)
3. ayod syncs changes to host automatically (via inotify/fswatch)
4. Session ends → output persists on host

## Implementation

### ayod Side (in sandbox)

```go
// cmd/ayod/sync.go
func (s *Server) StartOutputSync(sessionID string) {
    outputDir := fmt.Sprintf("/output/%s", sessionID)
    os.MkdirAll(outputDir, 0755)
    
    // Watch for changes
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add(outputDir)
    
    // Notify host daemon of changes
    for event := range watcher.Events {
        s.notifyHost("output_changed", event.Name)
    }
}
```

### Host Side

```go
// internal/daemon/output_sync.go
func (d *Daemon) HandleOutputChange(sandboxID, path string) {
    // Copy from sandbox to host
    hostPath := filepath.Join(paths.OutputDir(), sessionID, relativePath)
    d.sandbox.CopyFrom(sandboxID, path, hostPath)
}
```

## Environment Variables

Agent sees:
- `$OUTPUT=/output/{session_id}` - Where to write output
- `$SESSION_ID={session_id}` - Current session identifier

## Files to Create/Modify

- Add sync logic to `cmd/ayod/`
- Create `internal/daemon/output_sync.go`
- Update sandbox creation to mount output directory
- Update session start to create output dir

## Testing

- Write file in /output/, verify appears on host
- Verify no approval prompt for /output/ writes
- Verify output persists after session ends
- Verify $OUTPUT env var is set correctly
