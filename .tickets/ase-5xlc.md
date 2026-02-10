---
id: ase-5xlc
status: closed
deps: []
links: []
created: 2026-02-10T01:33:15Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-8d04
---
# Create cmd/ayo/share.go with base command structure

Create the base 'ayo share' command with subcommand structure and register it in root.go.

## Context
The share command will have subcommands: add (default), list, rm. This ticket creates the file structure and base command.

## Files to Create/Modify
- cmd/ayo/share.go (new)
- cmd/ayo/root.go (add share command)

## Implementation

### share.go
```go
package main

import (
    "github.com/spf13/cobra"
)

func newShareCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "share",
        Short: "Share host directories with sandboxed agents",
        Long: `Share host directories with sandboxed agents.

Shares create symlinks in a workspace directory that is mounted into sandboxes.
Unlike the deprecated 'ayo mount' command, shares take effect immediately
without requiring sandbox restart.

Shared directories appear at /workspace/{name} inside the sandbox.

Examples:
  ayo share ~/Code/myproject           Share with auto-generated name
  ayo share . --as project             Share current directory as 'project'
  ayo share ~/data --session           Share for this session only
  ayo share list                       List all shares
  ayo share rm project                 Remove a share`,
    }

    // Add subcommands (will be implemented in separate tickets)
    // cmd.AddCommand(newShareAddCmd())
    // cmd.AddCommand(newShareListCmd())
    // cmd.AddCommand(newShareRmCmd())

    return cmd
}
```

### root.go modification
Find where commands are added (search for 'AddCommand') and add:
```go
cmd.AddCommand(newShareCmd())
```

## Reference
Look at cmd/ayo/mount.go for the pattern:
- newMountCmd() creates the parent command
- Subcommands are added via cmd.AddCommand()
- Help text includes examples

## Command Structure (for reference)
Final structure will be:
- ayo share <path>        (add share, default action)
- ayo share add <path>    (explicit add)
- ayo share list          (list shares)
- ayo share ls            (alias for list)
- ayo share rm <name>     (remove share)

## Acceptance Criteria

- [ ] cmd/ayo/share.go created
- [ ] newShareCmd() function defined
- [ ] Command has helpful long description
- [ ] Command includes usage examples
- [ ] Command registered in root.go
- [ ] 'ayo share --help' works
- [ ] Build succeeds

