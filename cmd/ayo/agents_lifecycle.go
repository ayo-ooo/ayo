package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	// TODO: Re-implement db for build system
	// "github.com/alexcabrera/ayo/internal/db" - Removed as part of framework cleanup
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
				// Database persistence has been removed as part of framework cleanup
				// Agents are now standalone projects created with 'ayo fresh'
				return fmt.Errorf("agent promotion is no longer supported in the build system. Use 'ayo fresh' to create new agents")
			}), nil
		}),
	}

	return cmd
}

func archiveAgentCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <handle>",
		Short: "Archive an @ayo-created agent",
		Long: `Archive an @ayo-created agent to hide it from listings.

Archived agents are not deleted and can be restored later.

Examples:
  ayo agents archive @unused-agent`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Database persistence has been removed as part of framework cleanup
				// Agents are now standalone projects created with 'ayo fresh'
				return fmt.Errorf("agent archiving is no longer supported in the build system. Use 'ayo fresh' to create new agents")
			}), nil
		}),
	}

	return cmd
}

func unarchiveAgentCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unarchive <handle>",
		Short: "Unarchive an @ayo-created agent",
		Long: `Restore an archived @ayo-created agent.

This makes the agent visible again in listings and available for use.

Examples:
  ayo agents unarchive @restored-agent`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Database persistence has been removed as part of framework cleanup
				// Agents are now standalone projects created with 'ayo fresh'
				return fmt.Errorf("agent unarchiving is no longer supported in the build system. Use 'ayo fresh' to create new agents")
			}), nil
		}),
	}

	return cmd
}

func refineAgentCmd(cfgPath *string) *cobra.Command {
	var note string

	cmd := &cobra.Command{
		Use:   "refine <handle>",
		Short: "Refine an @ayo-created agent",
		Long: `Create a refined version of an @ayo-created agent with improvements.

This creates a new agent based on the original but with refinements.
The original agent remains unchanged.

Examples:
  ayo agents refine @original-agent --note "Improved reasoning capabilities"
  ayo agents refine science-researcher --note "Added domain-specific knowledge"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])
			if note == "" {
				return fmt.Errorf("--note is required to explain the refinement reason")
			}

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Database persistence has been removed as part of framework cleanup
				// Agents are now standalone projects created with 'ayo fresh'
				return fmt.Errorf("agent refinement is no longer supported in the build system. Use 'ayo fresh' to create new agents")
			}), nil
		}),
	}

	cmd.Flags().StringVar(&note, "note", "", "Explanation of the refinement (required)")
	cmd.MarkFlagRequired("note")

	return cmd
}