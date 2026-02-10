---
id: ase-doxk
status: closed
deps: [ase-1f71]
links: []
created: 2026-02-10T01:34:40Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-8d04
---
# Implement session share cleanup

Ensure session shares are automatically removed when their associated session ends.

## Context
Session shares (created with --session flag) should be cleaned up when the session terminates. This requires hooking into the session lifecycle.

## Files to Modify
- internal/share/share.go (add cleanup method)
- Session management code (TBD - need to find session lifecycle hooks)

## Dependencies
- ase-1f71 (ShareService with Session support)

## Implementation

### Add cleanup method to ShareService
```go
// RemoveSessionShares removes all shares associated with a session ID.
func (s *ShareService) RemoveSessionShares(sessionID string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.shares == nil {
        return nil
    }

    // Find and remove session shares
    var toRemove []string
    for _, share := range s.shares.Shares {
        if share.Session && share.SessionID == sessionID {
            toRemove = append(toRemove, share.Name)
        }
    }

    // Remove symlinks and update list
    for _, name := range toRemove {
        symlinkPath := filepath.Join(sync.WorkspaceDir(), name)
        os.Remove(symlinkPath) // Ignore errors - may already be gone
        
        // Remove from list
        for i, share := range s.shares.Shares {
            if share.Name == name {
                s.shares.Shares = append(s.shares.Shares[:i], s.shares.Shares[i+1:]...)
                break
            }
        }
    }

    return s.saveUnlocked()
}
```

### Hook into session lifecycle
Find where sessions are cleaned up and call:
```go
shareService := share.NewShareService()
if err := shareService.Load(); err == nil {
    shareService.RemoveSessionShares(sessionID)
}
```

## Investigation Required
1. Search for session cleanup/termination code:
   - internal/server/session_manager.go
   - Look for Sleep(), Close(), or similar methods
   
2. Find where session ID is available:
   - AgentSession struct has SessionID field
   
3. Determine the right hook point:
   - When session ends normally
   - When session times out
   - When daemon stops

## Search Hints
- grep for "session.*end" or "session.*close"
- Look at SessionManager.Sleep() method
- Check daemon shutdown handling

## Acceptance Criteria

- [ ] RemoveSessionShares(sessionID) method exists
- [ ] Method removes all shares with matching sessionID
- [ ] Symlinks deleted from workspace directory
- [ ] shares.json updated
- [ ] Session cleanup calls RemoveSessionShares
- [ ] Works for normal session end
- [ ] Works for daemon shutdown
- [ ] Unit tests for cleanup method

