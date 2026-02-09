package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/paths"
)

func promoteAgentCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "promote <old-handle> <new-handle>",
		Short: "Promote an @ayo-created agent to user-owned",
		Long: `Promote an agent that was created by @ayo to a new, user-owned handle.

This copies the agent configuration to a new handle that you control.
The original @ayo-created agent is marked as promoted but remains functional.

Examples:
  ayo agents promote science-researcher my-science-helper
  ayo agents promote @ayo-created-agent @my-agent`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldHandle := agent.NormalizeHandle(args[0])
			newHandle := agent.NormalizeHandle(args[1])

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Open database
				database, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
				if err != nil {
					return fmt.Errorf("open database: %w", err)
				}
				defer database.Close()

				// Check if agent is @ayo-created
				if !agent.IsAyoCreated(cmd.Context(), queries, oldHandle) {
					return fmt.Errorf("%s is not an @ayo-created agent; only @ayo-created agents can be promoted", oldHandle)
				}

				// Promote agent
				if err := agent.PromoteAgent(cmd.Context(), cfg, queries, oldHandle, newHandle); err != nil {
					return err
				}

				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ Promoted %s to %s", oldHandle, newHandle)))
				fmt.Println("  You now own this agent and can modify it freely.")
				return nil
			})
		},
	}

	return cmd
}

func archiveAgentCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <handle>",
		Short: "Archive an @ayo-created agent",
		Long: `Archive an @ayo-created agent to hide it from listings.

Archived agents:
- Are hidden from 'ayo agents list' (use --archived to see them)
- Still exist on disk and can be used if you know the handle
- Are NOT included in capability search
- Can be unarchived later with 'ayo agents unarchive'

Examples:
  ayo agents archive old-helper
  ayo agents archive @unused-agent`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])

			// Open database
			database, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer database.Close()

			// Check if agent is @ayo-created
			if !agent.IsAyoCreated(cmd.Context(), queries, handle) {
				return fmt.Errorf("%s is not an @ayo-created agent; only @ayo-created agents can be archived", handle)
			}

			// Archive agent
			if err := agent.ArchiveAgent(cmd.Context(), queries, handle); err != nil {
				return err
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Archived %s", handle)))
			fmt.Println("  Use 'ayo agents unarchive' to restore it.")
			return nil
		},
	}

	return cmd
}

func unarchiveAgentCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unarchive <handle>",
		Short: "Unarchive an @ayo-created agent",
		Long: `Restore an archived @ayo-created agent.

This reverses the archive operation, making the agent visible again
in listings and capability searches.

Examples:
  ayo agents unarchive old-helper
  ayo agents unarchive @restored-agent`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])

			// Open database
			database, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer database.Close()

			// Unarchive agent
			if err := agent.UnarchiveAgent(cmd.Context(), queries, handle); err != nil {
				return err
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Unarchived %s", handle)))
			return nil
		},
	}

	return cmd
}

func refineAgentCmd(cfgPath *string) *cobra.Command {
	var (
		appendSystem string
		replaceWith  string
		note         string
	)

	cmd := &cobra.Command{
		Use:   "refine <handle>",
		Short: "Refine an @ayo-created agent's system prompt",
		Long: `Refine an agent that was created by @ayo.

This allows you to modify the agent's system prompt by either appending
new instructions or replacing the entire prompt. All refinements are
tracked in the database with their reasons.

Examples:
  # Append instructions to existing prompt
  ayo agents refine @science-researcher \
    --append "When discussing biology, always cite recent papers." \
    --note "User prefers academic sources"

  # Replace the entire prompt
  ayo agents refine @helper \
    --replace "You are a specialized helper for..." \
    --note "Complete rewrite for new use case"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])

			// Validate flags
			if appendSystem == "" && replaceWith == "" {
				return fmt.Errorf("either --append or --replace is required")
			}
			if appendSystem != "" && replaceWith != "" {
				return fmt.Errorf("use only one of --append or --replace, not both")
			}
			if note == "" {
				return fmt.Errorf("--note is required to explain the refinement reason")
			}

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Open database
				database, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
				if err != nil {
					return fmt.Errorf("open database: %w", err)
				}
				defer database.Close()

				// Check if agent is @ayo-created
				if !agent.IsAyoCreated(cmd.Context(), queries, handle) {
					return fmt.Errorf("%s is not an @ayo-created agent; only @ayo-created agents can be refined", handle)
				}

				// Refine agent
				opts := agent.RefinementOptions{
					AgentHandle:  handle,
					Reason:       note,
					UpdateOnDisk: true,
				}
				if appendSystem != "" {
					opts.AppendPrompt = appendSystem
				} else {
					opts.NewPrompt = replaceWith
				}

				if err := agent.RefineAgent(cmd.Context(), cfg, queries, opts); err != nil {
					return err
				}

				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ Refined %s", handle)))
				if appendSystem != "" {
					fmt.Println("  Appended to system prompt")
				} else {
					fmt.Println("  Replaced system prompt")
				}
				fmt.Printf("  Reason: %s\n", note)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&appendSystem, "append", "", "text to append to the existing system prompt")
	cmd.Flags().StringVar(&replaceWith, "replace", "", "new system prompt to replace the existing one")
	cmd.Flags().StringVar(&note, "note", "", "explanation for this refinement (required)")

	return cmd
}
