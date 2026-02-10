---
id: ase-0r4v
status: closed
deps: [ase-et8m, ase-625q]
links: []
created: 2026-02-10T01:36:32Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-8d04
---
# Implement 'ayo share migrate' command

Create a migration command that converts existing mount grants to shares.

## Context
Users with existing grants in mounts.json need a way to migrate to the share system without manually recreating each grant.

## File to Modify
- cmd/ayo/share.go

## Dependencies
- ase-et8m (share add works)
- ase-625q (mount deprecation in place)

## Implementation

```go
func newShareMigrateCmd() *cobra.Command {
    var dryRun bool
    var removeOld bool

    cmd := &cobra.Command{
        Use:   "migrate",
        Short: "Migrate grants from 'ayo mount' to shares",
        Long: `Migrate existing grants from 'ayo mount' to the share system.

Reads grants from mounts.json and creates equivalent shares. Use --dry-run
to preview changes without applying them.

Note: Readonly grants will be migrated as read-write shares. The share system
doesn't support read-only mode; use file permissions instead if needed.

Examples:
  ayo share migrate --dry-run    Preview migration
  ayo share migrate              Apply migration
  ayo share migrate --remove-old Remove old grants after migration`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Load existing grants
            grantService := mounts.NewGrantService()
            if err := grantService.Load(); err != nil {
                return fmt.Errorf("load grants: %w", err)
            }

            grants := grantService.List()
            if len(grants) == 0 {
                fmt.Println("No grants to migrate.")
                return nil
            }

            // Load share service
            shareService := share.NewShareService()
            if err := shareService.Load(); err != nil {
                return fmt.Errorf("load shares: %w", err)
            }

            if dryRun {
                fmt.Println("Would migrate the following grants:")
                for _, g := range grants {
                    name := filepath.Base(g.Path)
                    fmt.Printf("  %s → /workspace/%s\n", g.Path, name)
                }
                fmt.Println("\nRun without --dry-run to apply.")
                return nil
            }

            // Migrate each grant
            var migrated, skipped int
            for _, g := range grants {
                name := filepath.Base(g.Path)
                
                // Check if already shared
                if existing := shareService.Get(name); existing != nil {
                    fmt.Printf("  Skip %s (share '%s' already exists)\n", g.Path, name)
                    skipped++
                    continue
                }

                // Check if path exists
                if _, err := os.Stat(g.Path); os.IsNotExist(err) {
                    fmt.Printf("  Skip %s (path no longer exists)\n", g.Path)
                    skipped++
                    continue
                }

                // Add share
                if err := shareService.Add(g.Path, name, false, ""); err != nil {
                    fmt.Printf("  Error migrating %s: %v\n", g.Path, err)
                    skipped++
                    continue
                }

                fmt.Printf("  ✓ Migrated %s → /workspace/%s\n", g.Path, name)
                migrated++
            }

            // Optionally remove old grants
            if removeOld && migrated > 0 {
                for _, g := range grants {
                    grantService.Revoke(g.Path)
                }
                if err := grantService.Save(); err != nil {
                    fmt.Fprintf(os.Stderr, "Warning: failed to remove old grants: %v\n", err)
                }
                fmt.Println("Old grants removed.")
            }

            successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
            fmt.Printf("\n%s Migrated %d grant(s)", successStyle.Render("✓"), migrated)
            if skipped > 0 {
                fmt.Printf(", skipped %d", skipped)
            }
            fmt.Println()

            if !removeOld && migrated > 0 {
                fmt.Println("\nUse --remove-old to remove the old grants.")
            }

            return nil
        },
    }

    cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be migrated without making changes")
    cmd.Flags().BoolVar(&removeOld, "remove-old", false, "remove old grants after successful migration")

    return cmd
}
```

## Register Command
Add to newShareCmd():
```go
cmd.AddCommand(newShareMigrateCmd())
```

## Edge Cases
1. Grant path no longer exists → skip with warning
2. Share name already exists → skip with warning
3. Name conflicts (two grants with same basename) → second one skipped, suggest --as

## Output Example
```
  ✓ Migrated /Users/alex/Code/project → /workspace/project
  ✓ Migrated /Users/alex/Documents → /workspace/Documents
  Skip /tmp/old (path no longer exists)

✓ Migrated 2 grant(s), skipped 1
```

## Acceptance Criteria

- [ ] --dry-run shows preview without changes
- [ ] Migrates each grant to equivalent share
- [ ] Skips paths that no longer exist
- [ ] Skips shares that already exist
- [ ] --remove-old removes grants after migration
- [ ] Summary shows migrated and skipped counts
- [ ] Works with empty grants list
- [ ] JSON output option (optional)

