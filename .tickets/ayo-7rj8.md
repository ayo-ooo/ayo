---
id: ayo-7rj8
status: closed
deps: []
links: []
created: 2026-02-23T22:15:54Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-xfu3
tags: [squads, workspace]
---
# Implement squad workspace initialization

When creating a squad, properly initialize the workspace directory. Support multiple initialization methods and set up correct permissions.

## Context

Squads have a shared workspace at `{squad}/workspace/`. This ticket implements initialization options when creating a squad.

## Initialization Options

### Empty Workspace

```bash
ayo squad create dev-team
# Creates empty workspace/
```

### Git Clone

```bash
ayo squad create dev-team --git https://github.com/user/repo.git
# Clones repo into workspace/
```

### Copy from Host

```bash
ayo squad create dev-team --copy ~/Projects/myapp
# Copies directory into workspace/
```

### Link from Host

```bash
ayo squad create dev-team --link ~/Projects/myapp
# Mounts host directory at workspace/ (read-write)
```

## Implementation

```go
// internal/squads/workspace.go
type WorkspaceInit struct {
    Type   string // "empty", "git", "copy", "link"
    Source string // URL or path
}

func (s *SquadService) InitWorkspace(squadDir string, init WorkspaceInit) error {
    workspaceDir := filepath.Join(squadDir, "workspace")
    
    switch init.Type {
    case "empty":
        return os.MkdirAll(workspaceDir, 0755)
        
    case "git":
        return exec.Command("git", "clone", init.Source, workspaceDir).Run()
        
    case "copy":
        return copyDir(init.Source, workspaceDir)
        
    case "link":
        // Configure as share mount
        return s.shares.Add(init.Source, workspaceDir, true)
        
    default:
        return fmt.Errorf("unknown workspace init type: %s", init.Type)
    }
}
```

## Permissions

All agents in the squad need read-write access:

```go
func (s *SquadService) SetWorkspacePermissions(squadDir string, agents []string) error {
    workspaceDir := filepath.Join(squadDir, "workspace")
    
    // Create group for squad
    gid := s.createSquadGroup(squadDir)
    
    // Add all agent users to group
    for _, agent := range agents {
        s.addUserToGroup(agent, gid)
    }
    
    // Set workspace ownership
    return os.Chown(workspaceDir, -1, gid)
}
```

## Git Configuration

For git clones, configure git for agents:

```go
func configureGit(workspaceDir string, leadAgent string) error {
    // Set git user for commits
    exec.Command("git", "-C", workspaceDir, 
        "config", "user.name", leadAgent).Run()
    exec.Command("git", "-C", workspaceDir,
        "config", "user.email", fmt.Sprintf("%s@ayo.local", leadAgent)).Run()
    return nil
}
```

## Files to Create/Modify

1. **`internal/squads/workspace.go`** (new) - Workspace initialization
2. **`cmd/ayo/squad_create.go`** - Add workspace flags
3. **`internal/squads/service.go`** - Call workspace init

## Acceptance Criteria

- [ ] Empty workspace creates directory
- [ ] Git clone works with public repos
- [ ] Copy transfers files correctly
- [ ] Link mounts host directory
- [ ] All agents have write access
- [ ] Git configured with lead identity
- [ ] Errors are clear for invalid sources

## Testing

- Test each initialization type
- Test permission setup
- Test git configuration
- Test error handling for invalid sources
