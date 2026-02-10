---
id: ase-gbox
status: closed
deps: [ase-5xlc, ase-1f71]
links: []
created: 2026-02-10T01:34:22Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-8d04
---
# Implement 'ayo share rm' command

Implement the share remove command to delete shares.

## File to Modify
- cmd/ayo/share.go

## Dependencies
- ase-5xlc (share.go exists)
- ase-1f71 (ShareService.Remove() exists)

## Implementation

```go
func newShareRmCmd() *cobra.Command {
    var removeAll bool

    cmd := &cobra.Command{
        Use:     "rm [name|path]",
        Aliases: []string{"remove"},
        Short:   "Remove a share",
        Long: `Remove a share from the workspace.

Accepts either the share name (as shown in 'ayo share list') or the original
host path. Use --all to remove all shares at once.

The symlink is removed immediately from /workspace/.

Examples:
  ayo share rm project              Remove by name
  ayo share rm ~/Code/project       Remove by original path
  ayo share rm --all                Remove all shares`,
        Args: cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            service := share.NewShareService()
            if err := service.Load(); err != nil {
                return fmt.Errorf("load shares: %w", err)
            }

            if removeAll {
                shares := service.List()
                count := len(shares)

                for _, s := range shares {
                    if err := service.Remove(s.Name); err != nil {
                        return fmt.Errorf("remove %s: %w", s.Name, err)
                    }
                }

                if globalOutput.JSON {
                    return globalOutput.Print(map[string]int{"removed": count})
                }

                successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
                fmt.Printf("%s Removed %d share(s)\n", successStyle.Render("✓"), count)
                return nil
            }

            if len(args) == 0 {
                return fmt.Errorf("share name or path required (or use --all)")
            }

            nameOrPath := args[0]

            // Expand ~/ if present (in case user provides path)
            if len(nameOrPath) >= 2 && nameOrPath[:2] == "~/" {
                home, err := os.UserHomeDir()
                if err != nil {
                    return fmt.Errorf("expand home: %w", err)
                }
                nameOrPath = filepath.Join(home, nameOrPath[2:])
            }

            // Try to find by name first, then by path
            s := service.Get(nameOrPath)
            if s == nil {
                // Might be a path - resolve and try by path
                absPath, _ := filepath.Abs(nameOrPath)
                s = service.GetByPath(absPath)
            }

            if s == nil {
                warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
                fmt.Fprintf(os.Stderr, "%s Share not found: %s\n", warnStyle.Render("!"), nameOrPath)
                return nil
            }

            name := s.Name
            if err := service.Remove(name); err != nil {
                return fmt.Errorf("remove share: %w", err)
            }

            if globalOutput.JSON {
                return globalOutput.Print(map[string]string{"removed": name})
            }

            successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
            fmt.Printf("%s Removed share '%s'\n", successStyle.Render("✓"), name)
            return nil
        },
    }

    cmd.Flags().BoolVar(&removeAll, "all", false, "remove all shares")

    return cmd
}
```

## Register Command
Add to newShareCmd():
```go
cmd.AddCommand(newShareRmCmd())
```

## Output Format
Success: ✓ Removed share 'project'
Not found: ! Share not found: nonexistent
All removed: ✓ Removed 3 share(s)

## Acceptance Criteria

- [ ] 'ayo share rm <name>' removes by name
- [ ] 'ayo share rm <path>' removes by host path
- [ ] 'ayo share remove' alias works
- [ ] --all flag removes all shares
- [ ] ~/ paths are expanded for path lookup
- [ ] Warning (not error) if share not found
- [ ] Success message shows share name
- [ ] --json output works
- [ ] Symlink removed immediately from workspace

