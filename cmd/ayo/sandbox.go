package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/providers"
)

func newSandboxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Manage agent sandboxes",
		Long: `Manage sandboxed execution environments for agents.

Sandboxes are isolated Linux containers where agents execute commands.
They provide security isolation and reproducible environments.

Examples:
  ayo sandbox list                    List active sandboxes
  ayo sandbox show <id>               Show sandbox details
  ayo sandbox exec <id> <cmd>         Run command in sandbox
  ayo sandbox shell <id>              Open shell in sandbox
  ayo sandbox logs <id>               View sandbox logs
  ayo sandbox stop <id>               Stop a sandbox
  ayo sandbox prune                   Remove stopped sandboxes`,
	}

	cmd.AddCommand(newSandboxListCmd())
	cmd.AddCommand(newSandboxShowCmd())
	cmd.AddCommand(newSandboxExecCmd())
	cmd.AddCommand(newSandboxShellCmd())
	cmd.AddCommand(newSandboxLogsCmd())
	cmd.AddCommand(newSandboxStopCmd())
	cmd.AddCommand(newSandboxPruneCmd())

	return cmd
}

func newSandboxListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List active sandboxes",
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			sandboxes, err := provider.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list sandboxes: %w", err)
			}

			if len(sandboxes) == 0 {
				fmt.Println("No active sandboxes")
				return nil
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
			statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
			timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

			fmt.Println()
			fmt.Println(headerStyle.Render("  Sandboxes"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 70)))
			fmt.Println()

			for _, sb := range sandboxes {
				status := string(sb.Status)
				if sb.Status == providers.SandboxStatusRunning {
					statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
				} else if sb.Status == providers.SandboxStatusStopped {
					statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
				} else {
					statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
				}

				age := formatAge(sb.CreatedAt)

				fmt.Printf("  %s  %s  %s  %s\n",
					idStyle.Render(fmt.Sprintf("%-10s", truncate(sb.ID, 10))),
					nameStyle.Render(fmt.Sprintf("%-30s", truncate(sb.Name, 30))),
					statusStyle.Render(fmt.Sprintf("%-10s", status)),
					timeStyle.Render(age),
				)
			}
			fmt.Println()

			return nil
		},
	}

	return cmd
}

func newSandboxShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show sandbox details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			sb, err := provider.Get(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("failed to get sandbox: %w", err)
			}

			// Styles
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(16)
			valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))

			fmt.Println()
			fmt.Println(headerStyle.Render("  Sandbox Details"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 50)))
			fmt.Println()

			fmt.Printf("  %s %s\n", labelStyle.Render("ID:"), valueStyle.Render(sb.ID))
			fmt.Printf("  %s %s\n", labelStyle.Render("Name:"), valueStyle.Render(sb.Name))
			fmt.Printf("  %s %s\n", labelStyle.Render("Status:"), valueStyle.Render(string(sb.Status)))
			fmt.Printf("  %s %s\n", labelStyle.Render("Image:"), valueStyle.Render(sb.Image))
			if sb.User != "" {
				fmt.Printf("  %s %s\n", labelStyle.Render("User:"), valueStyle.Render(sb.User))
			}
			fmt.Printf("  %s %s\n", labelStyle.Render("Created:"), valueStyle.Render(sb.CreatedAt.Format(time.RFC3339)))

			if len(sb.Mounts) > 0 {
				fmt.Println()
				fmt.Println(headerStyle.Render("  Mounts"))
				for _, m := range sb.Mounts {
					ro := ""
					if m.ReadOnly {
						ro = " (ro)"
					}
					fmt.Printf("    %s -> %s%s\n", m.Source, m.Destination, ro)
				}
			}

			fmt.Println()

			return nil
		},
	}

	return cmd
}

func newSandboxExecCmd() *cobra.Command {
	var user string
	var workdir string

	cmd := &cobra.Command{
		Use:   "exec <id> <command> [args...]",
		Short: "Run command in sandbox",
		Long: `Execute a command inside a running sandbox.

Examples:
  ayo sandbox exec abc123 ls -la
  ayo sandbox exec abc123 --user ayo whoami
  ayo sandbox exec abc123 --workdir /workspace pwd`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			sandboxID := args[0]
			execCommand := args[1]
			cmdArgs := []string{}
			if len(args) > 2 {
				cmdArgs = args[2:]
			}

			opts := providers.ExecOptions{
				Command:    execCommand,
				Args:       cmdArgs,
				User:       user,
				WorkingDir: workdir,
			}

			result, err := provider.Exec(cmd.Context(), sandboxID, opts)
			if err != nil {
				return fmt.Errorf("exec failed: %w", err)
			}

			// Print output
			if result.Stdout != "" {
				fmt.Print(result.Stdout)
			}
			if result.Stderr != "" {
				fmt.Fprint(os.Stderr, result.Stderr)
			}

			if result.ExitCode != 0 {
				os.Exit(result.ExitCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&user, "user", "u", "", "Run as specified user")
	cmd.Flags().StringVarP(&workdir, "workdir", "w", "", "Working directory inside container")

	return cmd
}

func newSandboxShellCmd() *cobra.Command {
	var asAgent string

	cmd := &cobra.Command{
		Use:   "shell <id>",
		Short: "Open shell in sandbox",
		Long: `Open an interactive shell inside a running sandbox.

Note: Due to Apple Container limitations, full TTY support may not be
available. The shell will operate in line mode.

Examples:
  ayo sandbox shell abc123
  ayo sandbox shell abc123 --as @ayo`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			sandboxID := args[0]

			// Get sandbox info first
			sb, err := provider.Get(cmd.Context(), sandboxID)
			if err != nil {
				return fmt.Errorf("failed to get sandbox: %w", err)
			}

			fmt.Printf("Connecting to sandbox %s (%s)...\n", sb.ID, sb.Name)
			if asAgent != "" {
				fmt.Printf("Running as: %s\n", asAgent)
			}
			fmt.Println("Type 'exit' to leave the shell.")
			fmt.Println()

			// Determine user
			user := ""
			if asAgent != "" {
				// Strip @ prefix for username
				user = strings.TrimPrefix(asAgent, "@")
			}

			// Run sh in interactive mode
			// Note: Due to Apple Container limitations, we run non-interactively
			// and read commands from stdin line by line
			_ = user // used in loop below

			// Simple shell loop - read commands from stdin
			scanner := bufio.NewScanner(os.Stdin)
			for {
				fmt.Print("$ ")
				if !scanner.Scan() {
					break
				}
				line := scanner.Text()
				if line == "exit" || line == "quit" {
					break
				}
				if line == "" {
					continue
				}

				execOpts := providers.ExecOptions{
					Command: "sh",
					Args:    []string{"-c", line},
					User:    user,
				}
				result, err := provider.Exec(cmd.Context(), sandboxID, execOpts)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					continue
				}
				if result.Stdout != "" {
					fmt.Print(result.Stdout)
				}
				if result.Stderr != "" {
					fmt.Fprint(os.Stderr, result.Stderr)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&asAgent, "as", "", "Run shell as agent user (e.g., --as @ayo)")

	return cmd
}

func newSandboxLogsCmd() *cobra.Command {
	var follow bool
	var tail int

	cmd := &cobra.Command{
		Use:   "logs <id>",
		Short: "View sandbox logs",
		Long: `Fetch logs from a sandbox container.

Note: Logs are retrieved via 'container logs' command.

Examples:
  ayo sandbox logs abc123
  ayo sandbox logs abc123 --follow
  ayo sandbox logs abc123 --tail 50`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use container CLI directly since provider interface doesn't have Logs
			containerArgs := []string{"logs"}
			if follow {
				containerArgs = append(containerArgs, "--follow")
			}
			if tail > 0 {
				containerArgs = append(containerArgs, "--tail", fmt.Sprintf("%d", tail))
			}
			containerArgs = append(containerArgs, args[0])

			exec := exec.CommandContext(cmd.Context(), "container", containerArgs...)
			exec.Stdout = os.Stdout
			exec.Stderr = os.Stderr
			return exec.Run()
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().IntVarP(&tail, "tail", "n", 0, "Number of lines to show from end")

	return cmd
}

func newSandboxStopCmd() *cobra.Command {
	var force bool
	var timeout int

	cmd := &cobra.Command{
		Use:   "stop <id>",
		Short: "Stop a running sandbox",
		Long: `Stop a running sandbox container.

Examples:
  ayo sandbox stop abc123
  ayo sandbox stop abc123 --force
  ayo sandbox stop abc123 --timeout 30`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			sandboxID := args[0]

			opts := providers.SandboxStopOptions{}
			if force {
				opts.Signal = "SIGKILL"
			}
			if timeout > 0 {
				opts.Timeout = time.Duration(timeout) * time.Second
			}

			if err := provider.Stop(cmd.Context(), sandboxID, opts); err != nil {
				return fmt.Errorf("failed to stop sandbox: %w", err)
			}

			fmt.Printf("Sandbox %s stopped\n", sandboxID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force kill immediately")
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 10, "Seconds to wait before force kill")

	return cmd
}

func newSandboxPruneCmd() *cobra.Command {
	var force bool
	var all bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove stopped sandboxes",
		Long: `Remove all stopped sandbox containers.

Examples:
  ayo sandbox prune
  ayo sandbox prune --force
  ayo sandbox prune --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			sandboxes, err := provider.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list sandboxes: %w", err)
			}

			// Filter to stopped sandboxes (or all if --all)
			var toRemove []providers.Sandbox
			for _, sb := range sandboxes {
				if all || sb.Status == providers.SandboxStatusStopped {
					toRemove = append(toRemove, sb)
				}
			}

			if len(toRemove) == 0 {
				fmt.Println("No sandboxes to remove")
				return nil
			}

			// Confirm unless --force
			if !force {
				fmt.Printf("This will remove %d sandbox(es):\n", len(toRemove))
				for _, sb := range toRemove {
					fmt.Printf("  - %s (%s)\n", sb.ID, sb.Name)
				}
				fmt.Print("\nContinue? [y/N] ")

				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					fmt.Println("Aborted")
					return nil
				}
			}

			// Stop running sandboxes if --all
			if all {
				for _, sb := range toRemove {
					if sb.Status == providers.SandboxStatusRunning {
						if err := provider.Stop(cmd.Context(), sb.ID, providers.SandboxStopOptions{}); err != nil {
							fmt.Fprintf(os.Stderr, "Warning: failed to stop %s: %v\n", sb.ID, err)
						}
					}
				}
			}

			// Remove sandboxes
			removed := 0
			for _, sb := range toRemove {
				if err := provider.Delete(cmd.Context(), sb.ID, true); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", sb.ID, err)
				} else {
					removed++
				}
			}

			fmt.Printf("Removed %d sandbox(es)\n", removed)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Also stop and remove running sandboxes")

	return cmd
}

// Helper functions

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
