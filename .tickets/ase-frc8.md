---
id: ase-frc8
status: closed
deps: [ase-5xlc, ase-ionw]
links: []
created: 2026-02-10T01:34:00Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-8d04
---
# Implement 'ayo share list' command

Implement the share list command to display all current shares.

## File to Modify
- cmd/ayo/share.go

## Dependencies
- ase-5xlc (share.go exists)
- ase-ionw (ShareService types exist)

## Implementation

```go
func newShareListCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:     "list",
        Aliases: []string{"ls"},
        Short:   "List all shares",
        Long: `List all shared host directories.

Shows a table of shares with their host path and workspace location.
Use --json for machine-readable output.

Session shares are marked with ○ (temporary), permanent shares with ● (persistent).`,
        RunE: func(cmd *cobra.Command, args []string) error {
            service := share.NewShareService()
            if err := service.Load(); err != nil {
                return fmt.Errorf("load shares: %w", err)
            }

            shares := service.List()

            if globalOutput.JSON {
                return globalOutput.Print(shares)
            }

            if len(shares) == 0 {
                dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
                fmt.Println(dimStyle.Render("No shares configured. Use 'ayo share <path>' to add one."))
                return nil
            }

            // Styles
            headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
            nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)
            pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
            timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
            sessionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
            permStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

            fmt.Println()
            fmt.Println(headerStyle.Render("  Shares"))
            fmt.Println(headerStyle.Render("  " + strings.Repeat("─", 50)))
            fmt.Println()

            for _, s := range shares {
                icon := permStyle.Render("●")
                sessionInfo := ""
                if s.Session {
                    icon = sessionStyle.Render("○")
                    sessionInfo = sessionStyle.Render(" (session)")
                }

                age := shareTimeAgo(s.SharedAt)

                fmt.Printf("  %s %s → /workspace/%s%s\n",
                    icon,
                    nameStyle.Render(s.Name),
                    s.Name,
                    sessionInfo,
                )
                fmt.Printf("    %s  %s\n",
                    pathStyle.Render(s.Path),
                    timeStyle.Render(age),
                )
            }
            fmt.Println()
            fmt.Println(timeStyle.Render("  Access at /workspace/{name} inside sandbox"))
            fmt.Println()

            return nil
        },
    }

    return cmd
}

// shareTimeAgo returns human-readable time (copy from mount.go or extract)
func shareTimeAgo(t time.Time) string {
    // Same implementation as mountTimeAgo in mount.go
    d := time.Since(t)
    switch {
    case d < time.Minute:
        return "just now"
    case d < time.Hour:
        m := int(d.Minutes())
        if m == 1 {
            return "1 minute ago"
        }
        return fmt.Sprintf("%d minutes ago", m)
    case d < 24*time.Hour:
        h := int(d.Hours())
        if h == 1 {
            return "1 hour ago"
        }
        return fmt.Sprintf("%d hours ago", h)
    default:
        days := int(d.Hours() / 24)
        if days == 1 {
            return "yesterday"
        }
        return fmt.Sprintf("%d days ago", days)
    }
}
```

## Output Format (Human)
```
  Shares
  ──────────────────────────────────────────────────

  ● project → /workspace/project
    /Users/alex/Code/project  2 hours ago

  ○ temp-data → /workspace/temp-data (session)
    /tmp/data  5 minutes ago

  Access at /workspace/{name} inside sandbox
```

## Output Format (JSON)
```json
[
  {
    "name": "project",
    "path": "/Users/alex/Code/project",
    "session": false,
    "shared_at": "2025-02-09T12:00:00Z"
  }
]
```

## Register Command
Add to newShareCmd():
```go
cmd.AddCommand(newShareListCmd())
```

## Acceptance Criteria

- [ ] 'ayo share list' shows all shares
- [ ] 'ayo share ls' alias works
- [ ] Permanent shares shown with ● icon
- [ ] Session shares shown with ○ and (session) label
- [ ] Host path and workspace path shown
- [ ] Time ago shown for each share
- [ ] Empty state has helpful message
- [ ] --json outputs array of shares
- [ ] Footer shows /workspace/{name} reminder

