---
id: ase-ku4z
status: closed
deps: [ase-5xlc]
links: []
created: 2026-02-10T01:37:38Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-8d04
---
# Add comprehensive help text to share commands

Ensure all share commands have comprehensive, helpful text.

## File to Modify
- cmd/ayo/share.go

## Dependencies
- ase-5xlc (share.go exists)

## Requirements

### Parent command (ayo share --help)
Should explain:
- What shares are
- How they differ from (deprecated) mounts
- Key concept: /workspace/{name} path
- No restart required
- List all subcommands with brief descriptions

### Each subcommand
Should have:
- Clear short description (one line)
- Long description with context
- Usage examples
- Flag descriptions

## Content Review

### ayo share
```
Share host directories with sandboxed agents.

Shares create symlinks in a workspace directory that is mounted into sandboxes.
Changes take effect immediately without requiring sandbox restart.

Shared directories appear at /workspace/{name} inside the sandbox.

Usage:
  ayo share [path] [flags]
  ayo share [command]

Examples:
  ayo share ~/Code/myproject           Share with auto-generated name
  ayo share . --as project             Share current directory as 'project'
  ayo share ~/data --session           Share for this session only

Available Commands:
  add         Share a host directory
  list        List all shares
  rm          Remove a share
  migrate     Migrate from deprecated mount system

Flags:
  -h, --help   help for share

Use "ayo share [command] --help" for more information about a command.
```

### ayo share add
```
Share a host directory with sandboxed agents.

The directory is immediately accessible at /workspace/{name} inside any sandbox.
No sandbox restart is required.

Usage:
  ayo share add <path> [flags]

Examples:
  ayo share add ~/Code/myproject
  ayo share add . --as project
  ayo share add /tmp/data --session

Flags:
      --as string    Custom name for the share (default: directory basename)
      --session      Remove share when session ends
  -h, --help         help for add
```

### ayo share list
```
List all shared host directories.

Shows each share with its host path and workspace location.
Session shares (temporary) are marked with ○, permanent shares with ●.

Usage:
  ayo share list [flags]

Aliases:
  list, ls

Flags:
      --json   Output in JSON format
  -h, --help   help for list
```

### ayo share rm
```
Remove a share from the workspace.

Accepts either the share name or the original host path.
The symlink is removed immediately from /workspace/.

Usage:
  ayo share rm [name|path] [flags]

Aliases:
  rm, remove

Examples:
  ayo share rm project              Remove by name
  ayo share rm ~/Code/project       Remove by path
  ayo share rm --all                Remove all shares

Flags:
      --all    Remove all shares
  -h, --help   help for rm
```

## Acceptance Criteria

- [ ] Parent command has clear long description
- [ ] Each subcommand has examples
- [ ] /workspace/ path mentioned in help
- [ ] No-restart benefit mentioned
- [ ] Flag descriptions are clear
- [ ] Aliases documented (ls, remove)

