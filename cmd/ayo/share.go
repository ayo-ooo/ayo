package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/share"
)

func newShareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share [path]",
		Short: "Share host directories with sandboxed agents",
		Long: `Share host directories with sandboxed agents.

Shares create symlinks in a workspace directory that is mounted into sandboxes.
Changes take effect immediately without requiring sandbox restart.

Shared directories appear at /workspace/{name} inside the sandbox.

Examples:
  ayo share ~/Code/myproject           Share with auto-generated name
  ayo share . --as project             Share current directory as 'project'
  ayo share ~/data --session           Share for this session only
  ayo share list                       List all shares
  ayo share rm project                 Remove a share`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If path provided, delegate to add
			if len(args) > 0 {
				addCmd := newShareAddCmd()
				addCmd.SetArgs(args)
				// Pass flags
				if asName, _ := cmd.Flags().GetString("as"); asName != "" {
					addCmd.Flags().Set("as", asName)
				}
				if session, _ := cmd.Flags().GetBool("session"); session {
					addCmd.Flags().Set("session", "true")
				}
				return addCmd.Execute()
			}
			return cmd.Help()
		},
	}

	// Flags for default share behavior
	cmd.Flags().String("as", "", "custom name for the share")
	cmd.Flags().Bool("session", false, "remove share when session ends")

	cmd.AddCommand(newShareAddCmd())
	cmd.AddCommand(newShareListCmd())
	cmd.AddCommand(newShareRmCmd())

	return cmd
}

func newShareAddCmd() *cobra.Command {
	var asName string
	var session bool

	cmd := &cobra.Command{
		Use:   "add <path>",
		Short: "Share a host directory",
		Long: `Share a host directory with sandboxed agents.

The directory is immediately accessible at /workspace/{name} inside any sandbox.
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

			// Resolve to absolute
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			// Load share service
			service := share.NewService()
			if err := service.Load(); err != nil {
				return fmt.Errorf("load shares: %w", err)
			}

			// Add share
			if err := service.Add(absPath, asName, session, ""); err != nil {
				return err
			}

			// Determine display name
			name := asName
			if name == "" {
				name = filepath.Base(absPath)
			}

			if globalOutput.JSON {
				globalOutput.PrintData(map[string]any{
					"name":           name,
					"path":           absPath,
					"workspace_path": "/workspace/" + name,
				}, "")
				return nil
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

func newShareListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all shares",
		Long: `List all shared host directories.

Shows each share with its host path and workspace location.
Session shares (temporary) are marked with ○, permanent shares with ●.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service := share.NewService()
			if err := service.Load(); err != nil {
				return fmt.Errorf("load shares: %w", err)
			}

			shares := service.List()

			if globalOutput.JSON {
				globalOutput.PrintData(shares, "")
				return nil
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

func newShareRmCmd() *cobra.Command {
	var removeAll bool

	cmd := &cobra.Command{
		Use:     "rm [name|path]",
		Aliases: []string{"remove"},
		Short:   "Remove a share",
		Long: `Remove a share from the workspace.

Accepts either the share name or the original host path.
The symlink is removed immediately from /workspace/.

Examples:
  ayo share rm project              Remove by name
  ayo share rm ~/Code/project       Remove by path
  ayo share rm --all                Remove all shares`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := share.NewService()
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
					globalOutput.PrintData(map[string]int{"removed": count}, "")
					return nil
				}

				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Printf("%s Removed %d share(s)\n", successStyle.Render("✓"), count)
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("share name or path required (or use --all)")
			}

			nameOrPath := args[0]

			// Expand ~/ if present
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
				globalOutput.PrintData(map[string]string{"removed": name}, "")
				return nil
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Printf("%s Removed share '%s'\n", successStyle.Render("✓"), name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&removeAll, "all", false, "remove all shares")

	return cmd
}

// shareTimeAgo returns a human-readable relative time string.
func shareTimeAgo(t time.Time) string {
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
