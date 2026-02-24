---
id: ayo-158p
status: closed
deps: [ayo-ao4q]
links: []
created: 2026-02-23T22:15:54Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-xfu3
tags: [squads, shell]
---
# Add squad agent shell access

Enable `ayo squad shell` command to drop into a shell session as a specific agent inside a squad sandbox. Useful for debugging and experimentation.

## Context

After the shared sandbox with real Unix users is implemented (ayo-ao4q), users need a way to access the sandbox as specific agents for debugging.

## Command Syntax

```bash
# Shell as specific agent in squad
ayo squad shell #dev-team @frontend
# Drops into shell as 'frontend' user in dev-team sandbox

# Shell as lead (default)
ayo squad shell #dev-team
# Drops into shell as lead agent

# Short form (if in squad context)
ayo shell @frontend
```

## Implementation

```go
// cmd/ayo/squad_shell.go
var squadShellCmd = &cobra.Command{
    Use:   "shell [squad] [agent]",
    Short: "Open shell in squad sandbox as agent",
    RunE: func(cmd *cobra.Command, args []string) error {
        squad := args[0]
        agent := args[1] // or config.Lead if not specified
        
        // Connect to daemon
        client := daemon.NewClient()
        
        // Request shell session
        session, err := client.SquadShell(squad, agent)
        if err != nil {
            return err
        }
        
        // Connect PTY to terminal
        return session.Attach(os.Stdin, os.Stdout, os.Stderr)
    },
}
```

## ayod Integration

The shell request goes through ayod (ayo-kkxg):

```go
// internal/ayod/shell.go
func (d *Daemon) HandleShell(agent string) (*pty.Pty, error) {
    // Verify agent exists
    if !d.hasUser(agent) {
        return nil, fmt.Errorf("agent %s not found", agent)
    }
    
    // Start shell as agent user
    cmd := exec.Command("su", "-", agent, "-c", "/bin/sh")
    
    // Create PTY
    ptmx, err := pty.Start(cmd)
    if err != nil {
        return nil, err
    }
    
    return ptmx, nil
}
```

## Features

### Environment Setup

Shell inherits agent environment:
- `$USER` = agent name
- `$HOME` = `/home/{agent}`
- `$PWD` = `/workspace`

### Working Directory

Start in workspace:

```bash
ayo squad shell #dev-team @frontend
frontend@dev-team:/workspace$ pwd
/workspace
```

### Agent Identity

Commands run as the agent:

```bash
frontend@dev-team:/workspace$ whoami
frontend
frontend@dev-team:/workspace$ ls -la
# Files show frontend ownership for their changes
```

## Files to Create/Modify

1. **`cmd/ayo/squad_shell.go`** (new) - CLI command
2. **`internal/daemon/shell.go`** - Shell session handling
3. **`internal/ayod/shell.go`** - In-sandbox shell execution
4. **`internal/ui/pty.go`** - PTY attachment

## Acceptance Criteria

- [ ] `squad shell #team @agent` opens shell as agent
- [ ] Default to lead if no agent specified
- [ ] Shell runs as correct Unix user
- [ ] Environment variables set correctly
- [ ] Working directory is /workspace
- [ ] PTY works correctly (colors, line editing)
- [ ] Ctrl+C, Ctrl+D work as expected

## Testing

- Test shell as different agents
- Test default to lead
- Test environment variables
- Test file ownership in shell
- Test PTY functionality
