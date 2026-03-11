package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/version"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ayo",
		Short: "Build system for AI agents",
		Long: `ayo - Build system for creating standalone AI agent executables

Ayo compiles agent definitions into self-contained, distributable binaries.
No runtime framework required - just pure agent executables.

Usage:
  ayo fresh <name>        Create a new agent project
  ayo build <directory>  Build agent executable
  ayo checkit <directory> Validate configuration
  ayo add-agent <team> <name> Add agent to team
  ayo clean [directory]  Clean build artifacts and cache

Examples:
  ayo fresh my-agent
  ayo build my-agent
  ayo checkit my-agent
  ayo add-agent my-team reviewer
  ayo clean my-agent
  ayo clean --cache

For more information, visit: https://github.com/alexcabrera/ayo`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			fmt.Printf("ayo: unknown command '%s'\n", args[0])
			fmt.Println("Run 'ayo --help' for usage.")
		},
	}

	// Add subcommands
	cmd.AddCommand(newFreshCmd())
	cmd.AddCommand(newBuildCmd())
	cmd.AddCommand(newCheckitCmd())
	cmd.AddCommand(newAddAgentCmd())
	cmd.AddCommand(newCleanCmd())

	// Version flag
	cmd.Version = version.Version

	return cmd
}

