---
id: ase-ionw
status: closed
deps: [ase-uwnw]
links: []
created: 2026-02-10T01:32:12Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-8d04
---
# Create internal/share package with Share types

Create the foundational internal/share/ package with core types for managing shares. This ticket creates types only - symlink operations are in a separate ticket.

## Context
Shares are symlinks from the workspace directory to host paths. The share service manages a shares.json file that persists share metadata.

## Files to Create
- internal/share/share.go (new)
- internal/share/share_test.go (new)

## Type Definitions

```go
package share

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
    
    "github.com/alexcabrera/ayo/internal/paths"
)

// Share represents a single shared host path.
type Share struct {
    Name      string    `json:"name"`       // Name in /workspace/
    Path      string    `json:"path"`       // Absolute host path
    Session   bool      `json:"session"`    // If true, removed when session ends
    SessionID string    `json:"session_id,omitempty"` // Session ID if session share
    SharedAt  time.Time `json:"shared_at"`
}

// SharesFile represents the shares.json file structure.
type SharesFile struct {
    Version int     `json:"version"`
    Shares  []Share `json:"shares"`
}

// ShareService manages filesystem shares.
type ShareService struct {
    mu       sync.RWMutex
    filePath string
    shares   *SharesFile
}
```

## Methods to Implement

1. NewShareService() *ShareService
   - Returns new service with filePath from sharesFilePath()

2. sharesFilePath() string (private)
   - Returns filepath.Join(paths.DataDir(), "shares.json")

3. Load() error
   - Read shares.json, handle missing file (init empty)
   - Use same pattern as internal/sandbox/mounts/grants.go:58-83

4. Save() error
   - Write shares.json with 0644 permissions
   - Use same pattern as grants.go:85-110

5. List() []Share
   - Return copy of all shares

6. Get(name string) *Share
   - Return share by workspace name

7. GetByPath(path string) *Share
   - Return share by original host path

## Reference Implementation
Follow the pattern in internal/sandbox/mounts/grants.go for:
- Mutex usage (RWMutex for thread safety)
- JSON marshaling with indentation
- Error handling patterns
- File permission patterns

## Acceptance Criteria

- [ ] Share struct defined with all fields
- [ ] SharesFile struct defined
- [ ] ShareService struct with mutex
- [ ] NewShareService() constructor
- [ ] Load() handles missing file gracefully
- [ ] Save() creates parent directory if missing
- [ ] List() returns copy of shares
- [ ] Get() returns share by name or nil
- [ ] GetByPath() returns share by path or nil
- [ ] Uses paths.DataDir() for file location
- [ ] Unit tests for all methods

