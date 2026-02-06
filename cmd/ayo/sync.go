package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/sync"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Manage sandbox synchronization",
		Long: `Manage sandbox synchronization via git.

Sync keeps sandbox state in a git repository that can be pushed to a remote
for backup and cross-machine synchronization.

Examples:
  ayo sync init              Initialize sync (git repo)
  ayo sync status            Show sync state
  ayo sync remote <url>      Set sync remote
  ayo sync push              Push to remote
  ayo sync pull              Pull from remote`,
	}

	cmd.AddCommand(newSyncInitCmd())
	cmd.AddCommand(newSyncStatusCmd())
	cmd.AddCommand(newSyncRemoteCmd())
	cmd.AddCommand(newSyncPushCmd())
	cmd.AddCommand(newSyncPullCmd())

	return cmd
}

func newSyncInitCmd() *cobra.Command {
	var jsonOutput bool
	var createBranch bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize sync repository",
		Long: `Initialize the sandbox directory as a git repository.

Creates the directory structure and initial commit. Optionally creates
a machine-specific branch (machines/{hostname}).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := sync.Init(); err != nil {
				return fmt.Errorf("init sync: %w", err)
			}

			result := map[string]any{
				"initialized": true,
				"path":        sync.SandboxDir(),
			}

			if createBranch {
				branch, err := sync.CreateMachineBranch()
				if err != nil {
					return fmt.Errorf("create machine branch: %w", err)
				}
				result["branch"] = branch
			} else {
				branch, _ := sync.GetBranch()
				result["branch"] = branch
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render("Sync initialized"))
			fmt.Printf("  Path:   %s\n", result["path"])
			fmt.Printf("  Branch: %s\n", result["branch"])
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&createBranch, "branch", false, "create machine-specific branch")

	return cmd
}

func newSyncStatusCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !sync.IsInitialized() {
				if jsonOutput {
					result := map[string]any{"initialized": false}
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(result)
				}
				dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				fmt.Println(dimStyle.Render("Sync not initialized. Use 'ayo sync init' to initialize."))
				return nil
			}

			status, err := sync.Status()
			if err != nil {
				return fmt.Errorf("get status: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(status)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			// Sync enabled
			enabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Fprintf(w, "Sync:\t%s\n", enabledStyle.Render("enabled"))

			// Branch
			fmt.Fprintf(w, "Branch:\t%s\n", status.LocalBranch)

			// Remote
			if status.RemoteConfigured {
				fmt.Fprintf(w, "Remote:\t%s (%s)\n", status.RemoteURL, status.RemoteName)

				// Ahead/Behind
				if status.Ahead > 0 || status.Behind > 0 {
					var statusParts []string
					if status.Ahead > 0 {
						statusParts = append(statusParts, fmt.Sprintf("%d ahead", status.Ahead))
					}
					if status.Behind > 0 {
						statusParts = append(statusParts, fmt.Sprintf("%d behind", status.Behind))
					}
					fmt.Fprintf(w, "Status:\t%s\n", joinStrings(statusParts, ", "))
				} else {
					fmt.Fprintf(w, "Status:\tup to date\n")
				}
			} else {
				dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				fmt.Fprintf(w, "Remote:\t%s\n", dimStyle.Render("not configured"))
			}

			// Changes
			if status.HasChanges {
				warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
				fmt.Fprintf(w, "Changes:\t%s\n", warnStyle.Render("uncommitted changes"))
			} else {
				fmt.Fprintf(w, "Changes:\tclean\n")
			}

			// Last commit
			if status.LastCommit != "" {
				fmt.Fprintf(w, "Last commit:\t%s (%s)\n", status.LastCommit, status.LastCommitTime)
			}

			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func newSyncRemoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote",
		Short: "Manage sync remote",
	}

	cmd.AddCommand(newSyncRemoteAddCmd())
	cmd.AddCommand(newSyncRemoteShowCmd())

	return cmd
}

func newSyncRemoteAddCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Add or update sync remote",
		Long: `Add a git remote for syncing.

If a remote named 'origin' already exists, it will be updated with the new URL.

Examples:
  ayo sync remote add git@github.com:user/ayo-sync.git
  ayo sync remote add https://github.com/user/ayo-sync.git`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			if !sync.IsInitialized() {
				return fmt.Errorf("sync not initialized, run 'ayo sync init' first")
			}

			// Check if remote exists
			remoteName, _, err := sync.GetRemote()
			if err != nil {
				return fmt.Errorf("check remote: %w", err)
			}

			if remoteName != "" {
				// Update existing remote
				if err := sync.SetRemoteURL(remoteName, url); err != nil {
					return fmt.Errorf("update remote: %w", err)
				}
			} else {
				// Add new remote
				if err := sync.AddRemote("origin", url); err != nil {
					return fmt.Errorf("add remote: %w", err)
				}
				remoteName = "origin"
			}

			if jsonOutput {
				result := map[string]any{
					"name": remoteName,
					"url":  url,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render(fmt.Sprintf("Remote configured: %s", url)))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func newSyncRemoteShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show configured remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !sync.IsInitialized() {
				return fmt.Errorf("sync not initialized")
			}

			name, url, err := sync.GetRemote()
			if err != nil {
				return fmt.Errorf("get remote: %w", err)
			}

			if name == "" {
				if jsonOutput {
					result := map[string]any{"configured": false}
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(result)
				}
				dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				fmt.Println(dimStyle.Render("No remote configured. Use 'ayo sync remote add <url>' to add one."))
				return nil
			}

			if jsonOutput {
				result := map[string]any{
					"configured": true,
					"name":       name,
					"url":        url,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Printf("Remote: %s\n", name)
			fmt.Printf("  URL: %s\n", url)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func newSyncPushCmd() *cobra.Command {
	var message string
	var jsonOutput bool
	var force bool

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push changes to remote",
		Long: `Push local changes to the remote repository.

Commits any uncommitted changes with the specified message (or a default)
and pushes to the configured remote.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !sync.IsInitialized() {
				return fmt.Errorf("sync not initialized, run 'ayo sync init' first")
			}

			if !sync.HasRemote() {
				return fmt.Errorf("no remote configured, run 'ayo sync remote add <url>' first")
			}

			// For force push, we'd need to add a ForcePush function
			// For now, use regular push
			_ = force

			if err := sync.Push(message); err != nil {
				return fmt.Errorf("push: %w", err)
			}

			if jsonOutput {
				result := map[string]any{"pushed": true}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render("Pushed to remote"))
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "commit message")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&force, "force", false, "force push (not implemented)")

	return cmd
}

func newSyncPullCmd() *cobra.Command {
	var jsonOutput bool
	var force bool

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull changes from remote",
		Long: `Pull changes from the remote repository.

Commits any local changes first to avoid losing work, then fetches
and merges from the remote with rebase.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !sync.IsInitialized() {
				return fmt.Errorf("sync not initialized, run 'ayo sync init' first")
			}

			if !sync.HasRemote() {
				return fmt.Errorf("no remote configured, run 'ayo sync remote add <url>' first")
			}

			// For force pull, we'd need to add a ForcePull function
			// For now, use regular pull
			_ = force

			if err := sync.Pull(); err != nil {
				return fmt.Errorf("pull: %w", err)
			}

			if jsonOutput {
				result := map[string]any{"pulled": true}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render("Pulled from remote"))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&force, "force", false, "force pull (not implemented)")

	return cmd
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
