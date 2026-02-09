package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/sandbox/mounts"
)

func newMountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mount",
		Short: "Manage persistent filesystem access",
		Long: `Manage persistent filesystem access for sandboxed agents.

Grants persist across sessions and allow agents to access host filesystem paths.
Project-level mounts (.ayo.json) and session mounts (--mount flag) can only 
restrict access to paths already granted here—they cannot grant new access.

Mount Hierarchy:
  1. Global grants (ayo mount add) - Maximum accessible paths
  2. Project mounts (.ayo.json)    - Restricts to project-relevant paths  
  3. Session mounts (--mount)      - Further restricts for specific sessions

Examples:
  ayo mount add .                  Grant readwrite access to current directory
  ayo mount add ~/Documents --ro   Grant readonly access to Documents
  ayo mount list                   List all grants (table or --json)
  ayo mount rm ~/Documents         Remove specific grant
  ayo mount rm --all               Remove all grants`,
	}

	cmd.AddCommand(newMountAddCmd())
	cmd.AddCommand(newMountListCmd())
	cmd.AddCommand(newMountRmCmd())

	return cmd
}

func newMountAddCmd() *cobra.Command {
	var readonly bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "add <path>",
		Aliases: []string{"grant"},
		Short:   "Grant filesystem access",
		Long: `Grant persistent filesystem access for sandboxed agents.

By default grants readwrite access. Use --ro for read-only access.
Path can be relative, absolute, or use ~/. Paths are resolved to absolute paths.

If the path doesn't exist, a warning is shown but the grant is still created
(useful for paths that will be created later).

Examples:
  ayo mount add .                  # Current directory, read-write
  ayo mount add ~/Documents --ro   # Home subdirectory, read-only
  ayo mount add /tmp/project       # Absolute path`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Expand home directory
			if len(path) >= 2 && path[:2] == "~/" {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("expand home directory: %w", err)
				}
				path = filepath.Join(home, path[2:])
			}

			// Resolve to absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			// Check if path exists (warn if not)
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
				fmt.Fprintf(os.Stderr, "%s Path does not exist: %s\n", warnStyle.Render("!"), absPath)
			}

			// Load grants service
			service := mounts.NewGrantService()
			if err := service.Load(); err != nil {
				return fmt.Errorf("load grants: %w", err)
			}

			// Determine mode
			mode := mounts.GrantModeReadWrite
			if readonly {
				mode = mounts.GrantModeReadOnly
			}

			// Grant access
			if err := service.Grant(absPath, mode); err != nil {
				return fmt.Errorf("grant access: %w", err)
			}

			// Save grants
			if err := service.Save(); err != nil {
				return fmt.Errorf("save grants: %w", err)
			}

			if jsonOutput {
				result := map[string]string{
					"path": absPath,
					"mode": string(mode),
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Printf("%s Granted %s access to %s\n", successStyle.Render("✓"), mode, absPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&readonly, "ro", false, "grant read-only access")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func newMountListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all filesystem grants",
		Long: `List all persistent filesystem grants.

Shows a table of granted paths with their access mode and grant date.
Use --json for machine-readable output.

Output columns:
  PATH     - Absolute path to granted directory/file
  MODE     - Access mode (readonly or readwrite)
  GRANTED  - Date the grant was created`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load grants service
			service := mounts.NewGrantService()
			if err := service.Load(); err != nil {
				return fmt.Errorf("load grants: %w", err)
			}

			grants := service.List()

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(grants)
			}

			if len(grants) == 0 {
				dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				fmt.Println(dimStyle.Render("No grants configured. Use 'ayo mount add <path>' to add one."))
				return nil
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
			rwStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
			roStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

			fmt.Println()
			fmt.Println(headerStyle.Render("  Filesystem Grants"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 60)))
			fmt.Println()

			for _, g := range grants {
				modeStyle := rwStyle
				modeIcon := "●"
				if g.Mode == mounts.GrantModeReadOnly {
					modeStyle = roStyle
					modeIcon = "○"
				}

				age := mountTimeAgo(g.GrantedAt)

				fmt.Printf("  %s %s  %s\n",
					modeStyle.Render(modeIcon),
					pathStyle.Render(g.Path),
					timeStyle.Render(fmt.Sprintf("(%s, %s)", g.Mode, age)),
				)
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func newMountRmCmd() *cobra.Command {
	var revokeAll bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "rm [path]",
		Aliases: []string{"revoke"},
		Short:   "Remove filesystem access",
		Long: `Remove persistent filesystem access.

Removes a previously granted path from sandbox access. Use --all to remove
all grants at once.

Path can be relative, absolute, or use ~/. Paths are resolved to absolute paths.
If the path wasn't granted, a warning is shown (not an error).

Examples:
  ayo mount rm ~/Documents         # Remove specific grant
  ayo mount rm .                    # Remove current directory grant
  ayo mount rm --all                # Remove all grants`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load grants service
			service := mounts.NewGrantService()
			if err := service.Load(); err != nil {
				return fmt.Errorf("load grants: %w", err)
			}

			if revokeAll {
				// Get current count
				grants := service.List()
				count := len(grants)

				// Revoke all
				for _, g := range grants {
					if err := service.Revoke(g.Path); err != nil {
						return fmt.Errorf("revoke %s: %w", g.Path, err)
					}
				}

				// Save
				if err := service.Save(); err != nil {
					return fmt.Errorf("save grants: %w", err)
				}

				if jsonOutput {
					result := map[string]int{"revoked": count}
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(result)
				}

				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Printf("%s Revoked %d grant(s)\n", successStyle.Render("✓"), count)
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("path required (or use --all)")
			}

			path := args[0]

			// Expand home directory
			if len(path) >= 2 && path[:2] == "~/" {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("expand home directory: %w", err)
				}
				path = filepath.Join(home, path[2:])
			}

			// Resolve to absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			// Check if granted
			grant := service.GetGrant(absPath)
			if grant == nil {
				warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
				fmt.Fprintf(os.Stderr, "%s Path not granted: %s\n", warnStyle.Render("!"), absPath)
				return nil
			}

			// Revoke
			if err := service.Revoke(absPath); err != nil {
				return fmt.Errorf("revoke access: %w", err)
			}

			// Save
			if err := service.Save(); err != nil {
				return fmt.Errorf("save grants: %w", err)
			}

			if jsonOutput {
				result := map[string]string{"revoked": absPath}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Printf("%s Revoked access to %s\n", successStyle.Render("✓"), absPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&revokeAll, "all", false, "revoke all grants")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

// mountTimeAgo returns a human-readable relative time string
func mountTimeAgo(t time.Time) string {
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
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}
