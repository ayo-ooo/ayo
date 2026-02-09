package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/cli"
	"github.com/alexcabrera/ayo/internal/daemon"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox/workingcopy"
)

// Ensure cli package is used
var _ = cli.Output{}

func newSandboxCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Manage agent sandboxes",
		Long: `Manage sandboxed execution environments for agents.

Sandboxes are isolated Linux containers where agents execute commands.
They provide security isolation and reproducible environments.

Examples:
  ayo sandbox service start           Start the sandbox service
  ayo sandbox service status          Show service status
  ayo sandbox list                    List active sandboxes
  ayo sandbox show [id]               Show sandbox details
  ayo sandbox exec [id] <cmd>         Run command in sandbox
  ayo sandbox login [id]              Open interactive shell
  ayo sandbox push <id> <src> <dest>  Copy file to sandbox
  ayo sandbox pull <id> <src> <dest>  Copy file from sandbox
  ayo sandbox stop [id]               Stop a sandbox
  ayo sandbox prune                   Remove stopped sandboxes`,
	}

	cmd.AddCommand(newSandboxServiceCmd(cfgPath))
	cmd.AddCommand(newSandboxListCmd())
	cmd.AddCommand(newSandboxShowCmd())
	cmd.AddCommand(newSandboxExecCmd())
	cmd.AddCommand(newSandboxShellCmd())
	cmd.AddCommand(newSandboxLoginCmd())
	cmd.AddCommand(newSandboxPushCmd())
	cmd.AddCommand(newSandboxPullCmd())
	cmd.AddCommand(newSandboxLogsCmd())
	cmd.AddCommand(newSandboxStartCmd())
	cmd.AddCommand(newSandboxStopCmd())
	cmd.AddCommand(newSandboxPruneCmd())
	cmd.AddCommand(newSandboxStatsCmd())
	cmd.AddCommand(newSandboxSyncCmd())
	cmd.AddCommand(newSandboxDiffCmd())
	cmd.AddCommand(newSandboxJoinCmd())
	cmd.AddCommand(newSandboxUsersCmd())

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
				if globalOutput.JSON {
					globalOutput.PrintData([]struct{}{}, "")
					return nil
				}
				if !globalOutput.Quiet {
					fmt.Println("No active sandboxes")
				}
				return nil
			}

			// JSON output
			if globalOutput.JSON {
				type sandboxJSON struct {
					ID        string    `json:"id"`
					Name      string    `json:"name"`
					Status    string    `json:"status"`
					Image     string    `json:"image,omitempty"`
					CreatedAt time.Time `json:"created_at"`
				}
				var out []sandboxJSON
				for _, sb := range sandboxes {
					out = append(out, sandboxJSON{
						ID:        sb.ID,
						Name:      sb.Name,
						Status:    string(sb.Status),
						Image:     sb.Image,
						CreatedAt: sb.CreatedAt,
					})
				}
				globalOutput.PrintData(out, "")
				return nil
			}

			// Quiet mode: just list IDs
			if globalOutput.Quiet {
				for _, sb := range sandboxes {
					fmt.Println(sb.ID)
				}
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
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show sandbox details",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			sb, err := provider.Get(cmd.Context(), sandboxID)
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

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")

	return cmd
}

func newSandboxExecCmd() *cobra.Command {
	var user string
	var workdir string
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "exec <command> [args...]",
		Short: "Run command in sandbox",
		Long: `Execute a command inside a running sandbox.

Flags (--id, --user, --workdir) must come before the command.
After the first non-flag argument, everything is passed to the command.

Examples:
  ayo sandbox exec ls -la
  ayo sandbox exec --id abc123 cat /etc/os-release
  ayo sandbox exec --user ayo whoami
  ayo sandbox exec sh -c "echo hello"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			execCommand := args[0]
			var cmdArgs []string
			if len(args) > 1 {
				cmdArgs = args[1:]
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
			hasOutput := false
			if result.Stdout != "" {
				fmt.Print(result.Stdout)
				hasOutput = true
			}
			if result.Stderr != "" {
				fmt.Fprint(os.Stderr, result.Stderr)
				hasOutput = true
			}

			if result.ExitCode != 0 {
				os.Exit(result.ExitCode)
			}

			// Show success indicator for commands with no output
			if !hasOutput {
				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Printf("%s Command completed (exit code 0)\n", successStyle.Render("✓"))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")
	cmd.Flags().StringVarP(&user, "user", "u", "", "Run as specified user")
	cmd.Flags().StringVarP(&workdir, "workdir", "w", "", "Working directory inside container")

	// Stop parsing flags after the first positional argument (the command)
	// This allows: ayo sandbox exec sh -c "echo hi" without needing --
	cmd.Flags().SetInterspersed(false)

	return cmd
}

func newSandboxShellCmd() *cobra.Command {
	var asAgent string
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Open shell in sandbox",
		Long: `Open an interactive shell inside a running sandbox.

Note: Due to Apple Container limitations, full TTY support may not be
available. The shell will operate in line mode.

Examples:
  ayo sandbox shell
  ayo sandbox shell --id abc123
  ayo sandbox shell --as @ayo`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

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

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")
	cmd.Flags().StringVar(&asAgent, "as", "", "Run shell as agent user (e.g., --as @ayo)")

	return cmd
}

func newSandboxLoginCmd() *cobra.Command {
	var asAgent string
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Open interactive shell in sandbox (human use only)",
		Long: `Open an interactive shell inside a running sandbox with full PTY support.

This command allocates a pseudo-terminal and provides a real interactive
shell experience with job control, tab completion, and proper terminal
handling.

Note: This command is for human use only. Agents should use 'sandbox exec'
or 'sandbox shell' instead.

Examples:
  ayo sandbox login                    Login to most recent sandbox as root
  ayo sandbox login --id abc123        Login to sandbox abc123 as root
  ayo sandbox login --as @ayo          Login as @ayo user
  ayo sandbox login --id abc123 --as @crush`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			// Get sandbox details
			target, err := provider.Get(cmd.Context(), sandboxID)
			if err != nil {
				return fmt.Errorf("failed to get sandbox: %w", err)
			}

			// Check sandbox is running
			if target.Status != providers.SandboxStatusRunning {
				return fmt.Errorf("sandbox %s is not running (status: %s)", target.ID, target.Status)
			}

			// Determine user
			user := ""
			homeDir := "/root"
			if asAgent != "" {
				user = strings.TrimPrefix(asAgent, "@")
				homeDir = "/home/" + user

				// Ensure user exists before login
				if err := provider.EnsureAgentUser(cmd.Context(), target.ID, user, ""); err != nil {
					return fmt.Errorf("failed to ensure user exists: %w", err)
				}
			}

			// Build container exec command with PTY
			containerArgs := []string{"exec", "-it"}
			if user != "" {
				containerArgs = append(containerArgs, "--user", user)
			}

			// Set environment variables
			containerArgs = append(containerArgs, "--env", "TERM="+os.Getenv("TERM"))
			containerArgs = append(containerArgs, "--env", "HOME="+homeDir)
			if user != "" {
				containerArgs = append(containerArgs, "--env", "USER="+user)
			} else {
				containerArgs = append(containerArgs, "--env", "USER=root")
			}

			containerArgs = append(containerArgs, target.ID)

			// Determine shell to use - prefer bash, then ash, then sh
			shell := detectShell(cmd.Context(), provider, target.ID)
			containerArgs = append(containerArgs, shell, "-l")

			// Execute with direct PTY passthrough
			execCmd := exec.CommandContext(cmd.Context(), "container", containerArgs...)
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			// Run the command - this takes over the terminal
			if err := execCmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					os.Exit(exitErr.ExitCode())
				}
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")
	cmd.Flags().StringVar(&asAgent, "as", "", "Login as agent user (e.g., --as @ayo)")

	return cmd
}

// detectShell attempts to find the best available shell in the container.
func detectShell(ctx context.Context, provider providers.SandboxProvider, sandboxID string) string {
	// Check for bash first
	result, err := provider.Exec(ctx, sandboxID, providers.ExecOptions{
		Command: "test",
		Args:    []string{"-x", "/bin/bash"},
	})
	if err == nil && result.ExitCode == 0 {
		return "/bin/bash"
	}

	// Check for ash (Alpine default)
	result, err = provider.Exec(ctx, sandboxID, providers.ExecOptions{
		Command: "test",
		Args:    []string{"-x", "/bin/ash"},
	})
	if err == nil && result.ExitCode == 0 {
		return "/bin/ash"
	}

	// Fall back to sh
	return "/bin/sh"
}

func newSandboxLogsCmd() *cobra.Command {
	var follow bool
	var tail int
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "View sandbox logs",
		Long: `Fetch logs from a sandbox container.

Note: Logs are retrieved via 'container logs' command.

Examples:
  ayo sandbox logs
  ayo sandbox logs --id abc123
  ayo sandbox logs --follow
  ayo sandbox logs --tail 50`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sandboxID == "" {
				provider := selectSandboxProvider()
				if provider == nil {
					return errors.New("no sandbox provider available on this platform")
				}
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			// Use container CLI directly since provider interface doesn't have Logs
			containerArgs := []string{"logs"}
			if follow {
				containerArgs = append(containerArgs, "--follow")
			}
			if tail > 0 {
				containerArgs = append(containerArgs, "--tail", fmt.Sprintf("%d", tail))
			}
			containerArgs = append(containerArgs, sandboxID)

			execCmd := exec.CommandContext(cmd.Context(), "container", containerArgs...)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().IntVarP(&tail, "tail", "n", 0, "Number of lines to show from end")

	return cmd
}

func newSandboxStartCmd() *cobra.Command {
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a stopped sandbox",
		Long: `Start a stopped sandbox container.

Examples:
  ayo sandbox start
  ayo sandbox start --id abc123`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox to start:")
				if err != nil {
					return err
				}
			}

			if err := provider.Start(cmd.Context(), sandboxID); err != nil {
				return fmt.Errorf("failed to start sandbox: %w", err)
			}

			fmt.Printf("Sandbox %s started\n", sandboxID)
			return nil
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")

	return cmd
}

func newSandboxStopCmd() *cobra.Command {
	var force bool
	var timeout int
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a running sandbox",
		Long: `Stop a running sandbox container.

Examples:
  ayo sandbox stop
  ayo sandbox stop --id abc123
  ayo sandbox stop --force
  ayo sandbox stop --timeout 30`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox to stop:")
				if err != nil {
					return err
				}
			}

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

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force kill immediately")
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 10, "Seconds to wait before force kill")

	return cmd
}

func newSandboxPruneCmd() *cobra.Command {
	var force bool
	var all bool
	var homes bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove stopped sandboxes",
		Long: `Remove all stopped sandbox containers.

Examples:
  ayo sandbox prune
  ayo sandbox prune --force
  ayo sandbox prune --all
  ayo sandbox prune --homes   # Also remove persistent agent home directories`,
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

			// Remove agent home directories if --homes
			if homes {
				homesDir := paths.AgentHomesDir()
				entries, err := os.ReadDir(homesDir)
				if err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("read agent homes directory: %w", err)
				}

				if len(entries) > 0 {
					if !force {
						fmt.Printf("\nThis will also remove %d agent home director(ies):\n", len(entries))
						for _, e := range entries {
							fmt.Printf("  - %s\n", e.Name())
						}
						fmt.Print("\nContinue? [y/N] ")

						var response string
						fmt.Scanln(&response)
						if strings.ToLower(response) != "y" {
							fmt.Println("Skipping home directories")
							return nil
						}
					}

					homesRemoved := 0
					for _, e := range entries {
						entryPath := filepath.Join(homesDir, e.Name())
						if err := os.RemoveAll(entryPath); err != nil {
							fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", entryPath, err)
						} else {
							homesRemoved++
						}
					}
					fmt.Printf("Removed %d agent home director(ies)\n", homesRemoved)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Also stop and remove running sandboxes")
	cmd.Flags().BoolVar(&homes, "homes", false, "Also remove persistent agent home directories")

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

func newSandboxPushCmd() *cobra.Command {
	var user string
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "push <local-path> <container-path>",
		Short: "Copy file or directory to sandbox",
		Long: `Copy a file or directory from the host to a sandbox container.

Examples:
  ayo sandbox push ./data.txt /tmp/data.txt
  ayo sandbox push ./mydir /home/ayo/mydir --id abc123
  ayo sandbox push ./script.sh /home/ayo/script.sh --user ayo`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			localPath := args[0]
			containerPath := args[1]

			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			// Get sandbox ID from flag or picker
			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			// Resolve sandbox ID
			sandboxes, err := provider.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list sandboxes: %w", err)
			}

			var target *providers.Sandbox
			for _, sb := range sandboxes {
				if sb.ID == sandboxID || strings.HasPrefix(sb.ID, sandboxID) || sb.Name == sandboxID {
					target = &sb
					break
				}
			}
			if target == nil {
				return fmt.Errorf("sandbox not found: %s", sandboxID)
			}

			// Read local file/directory
			info, err := os.Stat(localPath)
			if err != nil {
				return fmt.Errorf("cannot access local path: %w", err)
			}

			// Show what we're doing
			srcDesc := localPath
			if info.IsDir() {
				srcDesc = localPath + "/"
			}
			fmt.Printf("Pushing %s to %s:%s\n", srcDesc, target.ID[:8], containerPath)

			// Create tar archive with progress tracking
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)
			var fileCount int

			if info.IsDir() {
				err = filepath.Walk(localPath, func(path string, fi os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					relPath, err := filepath.Rel(localPath, path)
					if err != nil {
						return err
					}

					header, err := tar.FileInfoHeader(fi, "")
					if err != nil {
						return err
					}
					header.Name = filepath.Join(filepath.Base(containerPath), relPath)

					if err := tw.WriteHeader(header); err != nil {
						return err
					}

					if !fi.IsDir() {
						data, err := os.ReadFile(path)
						if err != nil {
							return err
						}
						if _, err := tw.Write(data); err != nil {
							return err
						}
						fileCount++
					}
					return nil
				})
			} else {
				header, err := tar.FileInfoHeader(info, "")
				if err != nil {
					return fmt.Errorf("create tar header: %w", err)
				}
				header.Name = filepath.Base(containerPath)

				if err := tw.WriteHeader(header); err != nil {
					return fmt.Errorf("write tar header: %w", err)
				}

				data, err := os.ReadFile(localPath)
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}

				if _, err := tw.Write(data); err != nil {
					return fmt.Errorf("write tar data: %w", err)
				}
				fileCount = 1
			}

			if err := tw.Close(); err != nil {
				return fmt.Errorf("close tar: %w", err)
			}

			// Create destination directory in container if needed
			destDir := filepath.Dir(containerPath)
			if destDir != "/" && destDir != "." {
				mkdirResult, err := provider.Exec(cmd.Context(), target.ID, providers.ExecOptions{
					Command:    "mkdir",
					Args:       []string{"-p", destDir},
					User:       user,
					WorkingDir: "/",
				})
				if err != nil {
					return fmt.Errorf("create destination directory: %w", err)
				}
				if mkdirResult.ExitCode != 0 {
					return fmt.Errorf("mkdir failed: %s", mkdirResult.Stderr)
				}
			}

			// Extract in container
			result, err := provider.Exec(cmd.Context(), target.ID, providers.ExecOptions{
				Command:    "tar",
				Args:       []string{"-xf", "-", "-C", destDir},
				Stdin:      buf.Bytes(),
				User:       user,
				WorkingDir: "/",
			})
			if err != nil {
				return fmt.Errorf("exec tar in sandbox: %w", err)
			}
			if result.ExitCode != 0 {
				return fmt.Errorf("tar failed: %s", result.Stderr)
			}

			// Success message
			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			sizeStr := formatBytes(int64(buf.Len()))
			if fileCount == 1 {
				fmt.Printf("%s Copied 1 file (%s)\n", successStyle.Render("✓"), sizeStr)
			} else {
				fmt.Printf("%s Copied %d files (%s)\n", successStyle.Render("✓"), fileCount, sizeStr)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")
	cmd.Flags().StringVarP(&user, "user", "u", "", "Run as specific user")

	return cmd
}

func newSandboxPullCmd() *cobra.Command {
	var user string
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "pull <container-path> <local-path>",
		Short: "Copy file or directory from sandbox",
		Long: `Copy a file or directory from a sandbox container to the host.

Examples:
  ayo sandbox pull /tmp/output.txt ./output.txt
  ayo sandbox pull /home/ayo/results ./results --id abc123
  ayo sandbox pull /var/log/app.log ./app.log --user root`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			containerPath := args[0]
			localPath := args[1]

			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			// Get sandbox ID from flag or picker
			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			// Resolve sandbox ID
			sandboxes, err := provider.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list sandboxes: %w", err)
			}

			var target *providers.Sandbox
			for _, sb := range sandboxes {
				if sb.ID == sandboxID || strings.HasPrefix(sb.ID, sandboxID) || sb.Name == sandboxID {
					target = &sb
					break
				}
			}
			if target == nil {
				return fmt.Errorf("sandbox not found: %s", sandboxID)
			}

			// Show progress message
			fmt.Printf("Pulling %s:%s to %s\n", target.ID[:8], containerPath, localPath)

			// Create tar in container and get output
			result, err := provider.Exec(cmd.Context(), target.ID, providers.ExecOptions{
				Command:    "tar",
				Args:       []string{"-cf", "-", "-C", filepath.Dir(containerPath), filepath.Base(containerPath)},
				User:       user,
				WorkingDir: "/",
			})
			if err != nil {
				return fmt.Errorf("exec tar in sandbox: %w", err)
			}
			if result.ExitCode != 0 {
				return fmt.Errorf("tar failed: %s", result.Stderr)
			}

			// Extract tar on host
			tr := tar.NewReader(bytes.NewReader([]byte(result.Stdout)))
			var fileCount int
			var totalBytes int64

			for {
				header, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("read tar: %w", err)
				}

				// Determine destination path
				destPath := localPath
				if header.Name != filepath.Base(containerPath) {
					// Subdirectory or file within a directory
					relPath := strings.TrimPrefix(header.Name, filepath.Base(containerPath)+"/")
					if relPath != header.Name {
						destPath = filepath.Join(localPath, relPath)
					}
				}

				switch header.Typeflag {
				case tar.TypeDir:
					if err := os.MkdirAll(destPath, os.FileMode(header.Mode)); err != nil {
						return fmt.Errorf("create directory: %w", err)
					}
				case tar.TypeReg:
					if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
						return fmt.Errorf("create parent directory: %w", err)
					}
					outFile, err := os.Create(destPath)
					if err != nil {
						return fmt.Errorf("create file: %w", err)
					}
					written, err := io.Copy(outFile, tr)
					outFile.Close()
					if err != nil {
						return fmt.Errorf("write file: %w", err)
					}
					if err := os.Chmod(destPath, os.FileMode(header.Mode)); err != nil {
						return fmt.Errorf("set permissions: %w", err)
					}
					fileCount++
					totalBytes += written
				}
			}

			// Success message
			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			sizeStr := formatBytes(totalBytes)
			if fileCount == 1 {
				fmt.Printf("%s Copied 1 file (%s)\n", successStyle.Render("✓"), sizeStr)
			} else {
				fmt.Printf("%s Copied %d files (%s)\n", successStyle.Render("✓"), fileCount, sizeStr)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")
	cmd.Flags().StringVarP(&user, "user", "u", "", "Run as specific user")

	return cmd
}

func newSandboxStatsCmd() *cobra.Command {
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show resource usage statistics for a sandbox",
		Long: `Display resource usage statistics for a running sandbox.

Shows CPU usage, memory consumption, disk usage, network I/O,
process count, and uptime.

Examples:
  ayo sandbox stats
  ayo sandbox stats --id abc123`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			// Resolve sandbox ID
			sandboxes, err := provider.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list sandboxes: %w", err)
			}

			var target *providers.Sandbox
			for _, sb := range sandboxes {
				if sb.ID == sandboxID || strings.HasPrefix(sb.ID, sandboxID) || strings.HasPrefix(sb.Name, sandboxID) {
					target = &sb
					break
				}
			}

			if target == nil {
				return fmt.Errorf("sandbox not found: %s", sandboxID)
			}

			stats, err := provider.Stats(cmd.Context(), target.ID)
			if err != nil {
				return fmt.Errorf("failed to get stats: %w", err)
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(20)
			valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))

			fmt.Println()
			fmt.Println(headerStyle.Render("  Sandbox Stats: " + target.Name))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 50)))
			fmt.Println()

			// CPU
			fmt.Printf("  %s %s\n",
				labelStyle.Render("CPU:"),
				valueStyle.Render(fmt.Sprintf("%.1f%%", stats.CPUPercent)))

			// Memory
			memUsage := formatBytes(stats.MemoryUsageBytes)
			if stats.MemoryLimitBytes > 0 {
				memLimit := formatBytes(stats.MemoryLimitBytes)
				pct := float64(stats.MemoryUsageBytes) / float64(stats.MemoryLimitBytes) * 100
				fmt.Printf("  %s %s\n",
					labelStyle.Render("Memory:"),
					valueStyle.Render(fmt.Sprintf("%s / %s (%.1f%%)", memUsage, memLimit, pct)))
			} else {
				fmt.Printf("  %s %s\n",
					labelStyle.Render("Memory:"),
					valueStyle.Render(memUsage))
			}

			// Disk
			if stats.DiskUsageBytes > 0 {
				fmt.Printf("  %s %s\n",
					labelStyle.Render("Disk:"),
					valueStyle.Render(formatBytes(stats.DiskUsageBytes)))
			}

			// Network
			if stats.NetworkRxBytes > 0 || stats.NetworkTxBytes > 0 {
				fmt.Printf("  %s %s\n",
					labelStyle.Render("Network RX:"),
					valueStyle.Render(formatBytes(stats.NetworkRxBytes)))
				fmt.Printf("  %s %s\n",
					labelStyle.Render("Network TX:"),
					valueStyle.Render(formatBytes(stats.NetworkTxBytes)))
			}

			// Processes
			if stats.ProcessCount > 0 {
				fmt.Printf("  %s %s\n",
					labelStyle.Render("Processes:"),
					valueStyle.Render(fmt.Sprintf("%d", stats.ProcessCount)))
			}

			// Uptime
			fmt.Printf("  %s %s\n",
				labelStyle.Render("Uptime:"),
				valueStyle.Render(formatDuration(stats.Uptime)))

			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")

	return cmd
}

func newSandboxSyncCmd() *cobra.Command {
	var dryRun bool
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "sync <sandbox-path> <host-path>",
		Short: "Sync working copy from sandbox back to host",
		Long: `Synchronize changes from a sandbox working copy back to the host filesystem.

This command copies modified files from the sandbox's working directory
back to the original host project directory.

Examples:
  ayo sandbox sync /workspace /path/to/project
  ayo sandbox sync /app . --dry-run
  ayo sandbox sync /workspace . --id abc123`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			sandboxPath := args[0]
			hostPath := args[1]

			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			// Resolve host path
			absHostPath, err := filepath.Abs(hostPath)
			if err != nil {
				return fmt.Errorf("resolve host path: %w", err)
			}

			// Resolve sandbox ID
			sandboxes, err := provider.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list sandboxes: %w", err)
			}

			var target *providers.Sandbox
			for _, sb := range sandboxes {
				if sb.ID == sandboxID || strings.HasPrefix(sb.ID, sandboxID) || strings.HasPrefix(sb.Name, sandboxID) {
					target = &sb
					break
				}
			}

			if target == nil {
				return fmt.Errorf("sandbox not found: %s", sandboxID)
			}

			// Create working copy for sync
			wc := &workingcopy.WorkingCopy{
				HostPath:       absHostPath,
				SandboxPath:    sandboxPath,
				SandboxID:      target.ID,
				IgnorePatterns: workingcopy.DefaultIgnorePatterns(),
			}

			manager := workingcopy.NewManager(provider)

			if dryRun {
				// Show diff instead of syncing
				diffs, err := manager.Diff(cmd.Context(), wc)
				if err != nil {
					return fmt.Errorf("diff failed: %w", err)
				}

				if len(diffs) == 0 {
					fmt.Println("No changes to sync")
					return nil
				}

				fmt.Println("Changes to be synced:")
				for _, d := range diffs {
					symbol := "M"
					switch d.Status {
					case workingcopy.DiffStatusAdded:
						symbol = "A"
					case workingcopy.DiffStatusDeleted:
						symbol = "D"
					}
					fmt.Printf("  %s %s\n", symbol, d.Path)
				}
				return nil
			}

			// Perform sync
			changedFiles, err := manager.Sync(cmd.Context(), wc)
			if err != nil {
				return fmt.Errorf("sync failed: %w", err)
			}

			if len(changedFiles) == 0 {
				fmt.Println("No files changed")
			} else {
				fmt.Printf("Synced %d files from %s:%s to %s\n", len(changedFiles), target.ID[:8], sandboxPath, absHostPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")

	return cmd
}

func newSandboxDiffCmd() *cobra.Command {
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "diff <sandbox-path> <host-path>",
		Short: "Show differences between sandbox and host",
		Long: `Compare files in a sandbox working copy with the host filesystem.

Shows which files have been added, modified, or deleted in the sandbox
compared to the host project.

Examples:
  ayo sandbox diff /workspace /path/to/project
  ayo sandbox diff /app . --id abc123`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := selectSandboxProvider()
			if provider == nil {
				return errors.New("no sandbox provider available on this platform")
			}

			sandboxPath := args[0]
			hostPath := args[1]

			if sandboxID == "" {
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			// Resolve host path
			absHostPath, err := filepath.Abs(hostPath)
			if err != nil {
				return fmt.Errorf("resolve host path: %w", err)
			}

			// Resolve sandbox ID
			sandboxes, err := provider.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list sandboxes: %w", err)
			}

			var target *providers.Sandbox
			for _, sb := range sandboxes {
				if sb.ID == sandboxID || strings.HasPrefix(sb.ID, sandboxID) || strings.HasPrefix(sb.Name, sandboxID) {
					target = &sb
					break
				}
			}

			if target == nil {
				return fmt.Errorf("sandbox not found: %s", sandboxID)
			}

			// Create working copy for diff
			wc := &workingcopy.WorkingCopy{
				HostPath:       absHostPath,
				SandboxPath:    sandboxPath,
				SandboxID:      target.ID,
				IgnorePatterns: workingcopy.DefaultIgnorePatterns(),
			}

			manager := workingcopy.NewManager(provider)

			diffs, err := manager.Diff(cmd.Context(), wc)
			if err != nil {
				return fmt.Errorf("diff failed: %w", err)
			}

			if len(diffs) == 0 {
				fmt.Println("No differences found")
				return nil
			}

			// Styles
			addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
			modifiedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			deletedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

			fmt.Printf("Comparing %s:%s with %s\n\n", target.ID[:8], sandboxPath, absHostPath)

			for _, d := range diffs {
				switch d.Status {
				case workingcopy.DiffStatusAdded:
					fmt.Printf("  %s %s\n", addedStyle.Render("A"), d.Path)
				case workingcopy.DiffStatusModified:
					fmt.Printf("  %s %s\n", modifiedStyle.Render("M"), d.Path)
				case workingcopy.DiffStatusDeleted:
					fmt.Printf("  %s %s\n", deletedStyle.Render("D"), d.Path)
				}
			}

			fmt.Printf("\n%d file(s) differ\n", len(diffs))

			return nil
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")

	return cmd
}

func newSandboxJoinCmd() *cobra.Command {
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "join <agent>",
		Short: "Add an agent to a sandbox",
		Long: `Add an agent to an existing sandbox for multi-agent collaboration.

The agent will get its own user account in the sandbox but shares the workspace.

Examples:
  ayo sandbox join @reviewer
  ayo sandbox join @tester --id abc123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agent := args[0]

			if sandboxID == "" {
				provider := selectSandboxProvider()
				if provider == nil {
					return errors.New("no sandbox provider available on this platform")
				}
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			client, err := daemon.ConnectOrStart(cmd.Context())
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.SandboxJoin(cmd.Context(), sandboxID, agent); err != nil {
				return fmt.Errorf("join sandbox: %w", err)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Agent %s joined sandbox %s", agent, sandboxID)))

			return nil
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")

	return cmd
}

func newSandboxUsersCmd() *cobra.Command {
	var sandboxID string

	cmd := &cobra.Command{
		Use:   "users",
		Short: "List agents in a sandbox",
		Long: `List all agents currently using a sandbox.

Examples:
  ayo sandbox users
  ayo sandbox users --id abc123`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sandboxID == "" {
				provider := selectSandboxProvider()
				if provider == nil {
					return errors.New("no sandbox provider available on this platform")
				}
				var err error
				sandboxID, err = pickSandbox(cmd.Context(), provider, "Select a sandbox:")
				if err != nil {
					return err
				}
			}

			client, err := daemon.ConnectOrStart(cmd.Context())
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			agents, err := client.SandboxAgents(cmd.Context(), sandboxID)
			if err != nil {
				return fmt.Errorf("get sandbox agents: %w", err)
			}

			if len(agents) == 0 {
				dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				fmt.Println(dimStyle.Render("No agents in this sandbox"))
				return nil
			}

			fmt.Printf("Agents in sandbox %s:\n", sandboxID)
			for _, agent := range agents {
				fmt.Printf("  %s\n", agent)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sandboxID, "id", "", "Sandbox ID (uses picker if not specified)")

	return cmd
}

// pickSandbox selects a sandbox - auto-selects if only one, shows picker if multiple.
// Returns the sandbox ID or an error if no sandboxes available or user cancelled.
func pickSandbox(ctx context.Context, provider providers.SandboxProvider, title string) (string, error) {
	sandboxes, err := provider.List(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list sandboxes: %w", err)
	}

	if len(sandboxes) == 0 {
		return "", errors.New("no active sandboxes")
	}

	// Auto-select if only one sandbox
	if len(sandboxes) == 1 {
		return sandboxes[0].ID, nil
	}

	// Build options for picker
	options := make([]huh.Option[string], len(sandboxes))
	for i, sb := range sandboxes {
		age := formatAge(sb.CreatedAt)
		status := string(sb.Status)
		label := fmt.Sprintf("%s  %s  %s  %s", sb.ID, truncate(sb.Name, 25), status, age)
		options[i] = huh.NewOption(label, sb.ID)
	}

	var selectedID string
	err = huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(&selectedID).
		Run()
	if err != nil {
		return "", err
	}

	return selectedID, nil
}
