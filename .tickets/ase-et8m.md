---
id: ase-et8m
status: closed
deps: [ase-5xlc, ase-1f71]
links: []
created: 2026-02-10T01:33:35Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-8d04
---
# Implement 'ayo share' and 'ayo share add' commands

Implement the share add command which creates a new share (symlink to host path).

## Context
'ayo share <path>' is the primary way users share directories. It should be simple and provide immediate feedback that no sandbox restart is needed.

## File to Modify
- cmd/ayo/share.go

## Dependencies
- ase-5xlc (share.go exists)
- ase-1f71 (ShareService.Add() exists)

## Implementation

### Default behavior
When 'ayo share <path>' is run without a subcommand, treat it as 'ayo share add <path>'.

### Flags
- --as <name>: Custom name for the share (default: derived from path basename)
- --session: Mark as session-only share (removed when session ends)
- --json: Output in JSON format (inherited from root)

### Command Implementation
```go
func newShareAddCmd() *cobra.Command {
    var asName string
    var session bool

    cmd := &cobra.Command{
        Use:   "add <path>",
        Short: "Share a host directory",
        Long: `Share a host directory with sandboxed agents.

The path is immediately accessible at /workspace/{name} inside any sandbox.
No sandbox restart required.

Path can be relative, absolute, or use ~/. Name is derived from the path
basename unless --as is specified.

Examples:
  ayo share add ~/Code/myproject
  ayo share add . --as project
  ayo share add /tmp/data --session`,
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            path := args[0]

            // Expand ~/ if present
            if len(path) >= 2 && path[:2] == "~/" {
                home, err := os.UserHomeDir()
                if err != nil {
                    return fmt.Errorf("expand home: %w", err)
                }
                path = filepath.Join(home, path[2:])
            }

            // Load share service
            service := share.NewShareService()
            if err := service.Load(); err != nil {
                return fmt.Errorf("load shares: %w", err)
            }

            // Add share (service handles path resolution, validation, symlink)
            if err := service.Add(path, asName, session, ""); err != nil {
                return err
            }

            // Get the share we just added for output
            absPath, _ := filepath.Abs(path)
            name := asName
            if name == "" {
                name = filepath.Base(absPath)
            }

            if globalOutput.JSON {
                return globalOutput.Print(map[string]any{
                    "name": name,
                    "path": absPath,
                    "workspace_path": "/workspace/" + name,
                })
            }

            successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
            fmt.Printf("%s Shared %s → /workspace/%s\n", 
                successStyle.Render("✓"), absPath, name)
            return nil
        },
    }

    cmd.Flags().StringVar(&asName, "as", "", "custom name for the share")
    cmd.Flags().BoolVar(&session, "session", false, "remove share when session ends")

    return cmd
}
```

### Make 'ayo share <path>' work as default
Set the parent command's RunE to delegate to add when args are provided:
```go
cmd := &cobra.Command{
    Use:   "share [path]",
    // ...
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) > 0 {
            // Delegate to add command
            addCmd := newShareAddCmd()
            addCmd.SetArgs(args)
            return addCmd.Execute()
        }
        return cmd.Help()
    },
}
```

## Output Format
Success: ✓ Shared /Users/alex/Code/project → /workspace/project
Error: error: path does not exist: /nonexistent

## Import Requirements
```go
import (
    "github.com/alexcabrera/ayo/internal/share"
)
```

## Acceptance Criteria

- [ ] 'ayo share <path>' works as shorthand
- [ ] 'ayo share add <path>' works explicitly
- [ ] --as flag allows custom name
- [ ] --session flag marks share as session-only
- [ ] ~/ paths are expanded
- [ ] Relative paths work
- [ ] Success message shows path and workspace location
- [ ] Error shown if path doesn't exist
- [ ] Error shown if name conflicts
- [ ] --json output works
- [ ] No restart message (key differentiator from mount)

