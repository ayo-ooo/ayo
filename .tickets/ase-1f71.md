---
id: ase-1f71
status: closed
deps: [ase-ionw, ase-uwnw]
links: []
created: 2026-02-10T01:32:33Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-8d04
---
# Implement Add/Remove methods with symlink operations

Add the Add() and Remove() methods to ShareService that create/delete symlinks in the workspace directory.

## Context
When a user shares a host path, a symlink is created in the workspace directory pointing to that path. The workspace directory is mounted into the container, so symlinks are followed by the bind mount.

## File to Modify
- internal/share/share.go

## Dependencies
- ase-ionw (Share types exist)
- ase-uwnw (WorkspaceDir() exists)

## Methods to Implement

### Add(path, name string, session bool, sessionID string) error
1. Resolve path to absolute using filepath.Abs()
2. Validate path exists on host (os.Stat), return error if not
3. If name is empty, derive from filepath.Base(path)
4. Validate name doesn't exist in shares (return error with suggestion)
5. Validate name is safe (no path separators, special chars)
6. Create symlink: os.Symlink(absPath, filepath.Join(sync.WorkspaceDir(), name))
7. Add Share to internal list
8. Call Save() to persist

```go
func (s *ShareService) Add(path, name string, session bool, sessionID string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Resolve absolute path
    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("resolve path: %w", err)
    }
    
    // Validate path exists
    if _, err := os.Stat(absPath); err != nil {
        return fmt.Errorf("path does not exist: %s", absPath)
    }
    
    // Generate name if empty
    if name == "" {
        name = filepath.Base(absPath)
    }
    
    // Validate name is safe
    if err := validateShareName(name); err != nil {
        return err
    }
    
    // Check for existing share with same name
    for _, share := range s.shares.Shares {
        if share.Name == name {
            return fmt.Errorf("share '%s' already exists, use --as to specify a different name", name)
        }
    }
    
    // Create symlink in workspace directory
    symlinkPath := filepath.Join(sync.WorkspaceDir(), name)
    if err := os.Symlink(absPath, symlinkPath); err != nil {
        return fmt.Errorf("create symlink: %w", err)
    }
    
    // Add to shares list
    s.shares.Shares = append(s.shares.Shares, Share{
        Name:      name,
        Path:      absPath,
        Session:   session,
        SessionID: sessionID,
        SharedAt:  time.Now(),
    })
    
    return s.saveUnlocked()
}
```

### Remove(nameOrPath string) error
1. Find share by name first, then by path
2. If found, delete symlink from workspace directory
3. Remove from internal list
4. Call Save() to persist
5. Return nil if not found (idempotent)

### validateShareName(name string) error (private)
- Reject empty names
- Reject names with / or \
- Reject . and ..
- Reject names starting with -

## Import Requirements
Add import for sync package (use alias to avoid conflict):
```go
import (
    ayosync "github.com/alexcabrera/ayo/internal/sync"
)
```

## Error Messages
- "path does not exist: %s"
- "share '%s' already exists, use --as to specify a different name"
- "invalid share name: %s (reason)"
- "create symlink: %w"
- "remove symlink: %w"

## Acceptance Criteria

- [ ] Add() validates path exists
- [ ] Add() auto-generates name from basename
- [ ] Add() validates name is safe
- [ ] Add() detects name conflicts with clear error
- [ ] Add() creates symlink in WorkspaceDir()
- [ ] Add() persists to shares.json
- [ ] Remove() finds by name or path
- [ ] Remove() deletes symlink
- [ ] Remove() is idempotent (no error if not found)
- [ ] Remove() persists changes
- [ ] validateShareName rejects unsafe names
- [ ] Unit tests for all scenarios

