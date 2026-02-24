package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

func squadShellCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell <squad> [agent]",
		Short: "Open an interactive shell in a squad sandbox",
		Long: `Open an interactive shell session inside a squad's sandbox container.

If an agent is specified, the shell runs as that agent's Unix user.
If no agent is specified, the shell runs as the squad's lead agent.

Examples:
  # Shell as the lead agent in dev-team squad
  ayo squad shell dev-team

  # Shell as a specific agent
  ayo squad shell dev-team frontend

  # Using @ prefix for agent name
  ayo squad shell dev-team @backend`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			squadName := strings.TrimPrefix(args[0], "#")

			// Determine agent
			var agentName string
			if len(args) > 1 {
				agentName = strings.TrimPrefix(args[1], "@")
			}

			// If no agent specified, use the lead agent from config
			if agentName == "" {
				cfg, err := config.LoadSquadConfig(squadName)
				if err != nil {
					return fmt.Errorf("load squad config: %w", err)
				}
				if cfg.Lead != "" {
					agentName = strings.TrimPrefix(cfg.Lead, "@")
				} else if len(cfg.Agents) > 0 {
					// Use the first agent as default
					agentName = strings.TrimPrefix(cfg.Agents[0], "@")
				} else {
					// Default to "ayo" user
					agentName = "ayo"
				}
			}

			return runSquadShell(ctx, squadName, agentName)
		},
	}

	return cmd
}

// runSquadShell opens an interactive shell in a squad sandbox as the specified agent.
func runSquadShell(ctx context.Context, squadName, agentName string) error {
	containerName := sandbox.SquadSandboxName(squadName)

	// Build the exec command
	// Using -it for interactive terminal with PTY
	args := []string{
		"exec",
		"-it",
		"--user", agentName,
		"--workdir", "/workspace",
		containerName,
		"sh",
	}

	cmd := exec.CommandContext(ctx, "container", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the shell - this blocks until the shell exits
	if err := cmd.Run(); err != nil {
		// Exit with the same code as the shell if it was an exit error
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("exec shell: %w", err)
	}

	return nil
}

func init() {
	// This will be added to the squad command in squad.go
}
