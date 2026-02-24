package main

import (
	"fmt"

	"github.com/alexcabrera/ayo/internal/daemon"
	"github.com/spf13/cobra"
)

func newHelpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help",
		Short: "Help about any command",
		Long:  "Help provides help for any command in the application.\n\nAdditional topics: cron",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	
	cmd.AddCommand(newHelpCronCmd())
	return cmd
}

func newHelpCronCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cron",
		Short: "Show cron expression syntax reference",
		Long:  "Display help for cron expression syntax including field formats, special characters, and aliases.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(daemon.CronHelp())
			return nil
		},
	}
}
