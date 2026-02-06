package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/cli"
	"github.com/alexcabrera/ayo/internal/daemon"
)

// Ensure cli package is used
var _ = cli.Output{}

func newTriggersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "triggers",
		Short:   "Manage triggers",
		Aliases: []string{"trigger"},
		Long: `Manage triggers that wake agents on events.

Triggers can be:
  cron    Schedule-based triggers (e.g., "every day at 9am")
  watch   File system triggers (e.g., "when a file changes")

Examples:
  # List all triggers
  ayo triggers list

  # Add a cron trigger
  ayo triggers add --type cron --agent @backup --schedule "0 0 2 * * *"

  # Add a watch trigger
  ayo triggers add --type watch --agent @build --path ./src --patterns "*.go"

  # Test a trigger
  ayo triggers test trig_123456789

  # Remove a trigger
  ayo triggers rm trig_123456789`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTriggersCmd().RunE(cmd, args)
		},
	}

	cmd.AddCommand(listTriggersCmd())
	cmd.AddCommand(showTriggerCmd())
	cmd.AddCommand(addTriggerCmd())
	cmd.AddCommand(removeTriggerCmd())
	cmd.AddCommand(testTriggerCmd())

	return cmd
}

func listTriggersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all registered triggers",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := daemon.ConnectOrStart(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			result, err := client.TriggerList(ctx)
			if err != nil {
				return fmt.Errorf("list triggers: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result.Triggers)
			}

			if len(result.Triggers) == 0 {
				if !globalOutput.Quiet {
					fmt.Println("No triggers registered")
					fmt.Println()
					fmt.Println("Add one with: ayo triggers add --type cron --agent @name --schedule \"0 0 * * * *\"")
				}
				return nil
			}

			// Quiet mode: just list IDs
			if globalOutput.Quiet {
				for _, t := range result.Triggers {
					fmt.Println(t.ID)
				}
				return nil
			}

			// Styles
			purple := lipgloss.Color("#a78bfa")
			cyan := lipgloss.Color("#67e8f9")
			muted := lipgloss.Color("#6b7280")
			subtle := lipgloss.Color("#374151")
			green := lipgloss.Color("#34d399")

			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
			dividerStyle := lipgloss.NewStyle().Foreground(subtle)
			idStyle := lipgloss.NewStyle().Foreground(cyan)
			typeStyle := lipgloss.NewStyle().Foreground(green)
			mutedStyle := lipgloss.NewStyle().Foreground(muted)

			fmt.Println()
			fmt.Println(headerStyle.Render("  Triggers"))
			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 70)))
			fmt.Println()

			// Header row
			fmt.Printf("  %-22s %-8s %-15s %-25s\n",
				mutedStyle.Render("ID"),
				mutedStyle.Render("Type"),
				mutedStyle.Render("Agent"),
				mutedStyle.Render("Config"))

			for _, t := range result.Triggers {
				config := ""
				if t.Type == "cron" {
					config = t.Schedule
				} else if t.Type == "watch" {
					config = t.Path
					if len(t.Patterns) > 0 {
						config += " (" + strings.Join(t.Patterns, ", ") + ")"
					}
				}
				if len(config) > 25 {
					config = config[:22] + "..."
				}

				fmt.Printf("  %-22s %-8s %-15s %-25s\n",
					idStyle.Render(t.ID),
					typeStyle.Render(t.Type),
					t.Agent,
					config)
			}

			fmt.Println()
			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 70)))
			fmt.Println(mutedStyle.Render(fmt.Sprintf("  %d triggers", len(result.Triggers))))
			fmt.Println()

			return nil
		},
	}

	return cmd
}

func showTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show trigger details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			ctx := cmd.Context()

			client, err := daemon.ConnectOrStart(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			result, err := client.TriggerGet(ctx, id)
			if err != nil {
				return fmt.Errorf("get trigger: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result.Trigger)
			}

			t := result.Trigger

			// Styles
			purple := lipgloss.Color("#a78bfa")
			cyan := lipgloss.Color("#67e8f9")
			muted := lipgloss.Color("#6b7280")
			subtle := lipgloss.Color("#374151")
			text := lipgloss.Color("#e5e7eb")

			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
			dividerStyle := lipgloss.NewStyle().Foreground(subtle)
			iconStyle := lipgloss.NewStyle().Foreground(cyan)
			labelStyle := lipgloss.NewStyle().Foreground(muted)
			valueStyle := lipgloss.NewStyle().Foreground(text)

			fmt.Println()
			fmt.Println("  " + iconStyle.Render("◆") + " " + headerStyle.Render(t.ID))
			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))

			fmt.Printf("  %s   %s\n", labelStyle.Render("Type:"), valueStyle.Render(t.Type))
			fmt.Printf("  %s  %s\n", labelStyle.Render("Agent:"), valueStyle.Render(t.Agent))
			fmt.Printf("  %s %s\n", labelStyle.Render("Source:"), valueStyle.Render(t.Source))

			if t.Type == "cron" {
				fmt.Printf("  %s %s\n", labelStyle.Render("Schedule:"), valueStyle.Render(t.Schedule))
			} else if t.Type == "watch" {
				fmt.Printf("  %s   %s\n", labelStyle.Render("Path:"), valueStyle.Render(t.Path))
				if len(t.Patterns) > 0 {
					fmt.Printf("  %s %s\n", labelStyle.Render("Patterns:"), valueStyle.Render(strings.Join(t.Patterns, ", ")))
				}
				if len(t.Events) > 0 {
					fmt.Printf("  %s %s\n", labelStyle.Render("Events:"), valueStyle.Render(strings.Join(t.Events, ", ")))
				}
			}

			if t.Prompt != "" {
				fmt.Printf("  %s %s\n", labelStyle.Render("Prompt:"), valueStyle.Render(t.Prompt))
			}

			fmt.Println()

			return nil
		},
	}

	return cmd
}

func addTriggerCmd() *cobra.Command {
	var (
		triggerType string
		agent       string
		prompt      string

		// Cron options
		schedule string

		// Watch options
		path      string
		patterns  []string
		recursive bool
		events    []string

		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new trigger",
		Long: `Add a new trigger to wake an agent.

Cron triggers use standard cron syntax with seconds:
  seconds minutes hours day-of-month month day-of-week

Examples:
  # Every hour
  ayo triggers add --type cron --agent @backup --schedule "0 0 * * * *"

  # Every day at 2am
  ayo triggers add --type cron --agent @reports --schedule "0 0 2 * * *"

  # Watch for Go file changes
  ayo triggers add --type watch --agent @build --path ./src --patterns "*.go"

  # Watch with custom prompt
  ayo triggers add --type watch --agent @review \
    --path ./src --patterns "*.go" \
    --prompt "Review the changed files"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Validation
			if triggerType == "" {
				return fmt.Errorf("--type is required (cron or watch)")
			}
			if agent == "" {
				return fmt.Errorf("--agent is required")
			}
			if triggerType == "cron" && schedule == "" {
				return fmt.Errorf("--schedule is required for cron triggers")
			}
			if triggerType == "watch" && path == "" {
				return fmt.Errorf("--path is required for watch triggers")
			}

			client, err := daemon.ConnectOrStart(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			params := daemon.TriggerRegisterParams{
				Type:      triggerType,
				Agent:     agent,
				Prompt:    prompt,
				Schedule:  schedule,
				Path:      path,
				Patterns:  patterns,
				Recursive: recursive,
				Events:    events,
			}

			result, err := client.TriggerRegister(ctx, params)
			if err != nil {
				return fmt.Errorf("register trigger: %w", err)
			}

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(result.Trigger)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render("✓ Created trigger: " + result.Trigger.ID))
			fmt.Printf("  Type:  %s\n", result.Trigger.Type)
			fmt.Printf("  Agent: %s\n", result.Trigger.Agent)

			return nil
		},
	}

	cmd.Flags().StringVarP(&triggerType, "type", "t", "", "trigger type: cron or watch (required)")
	cmd.Flags().StringVarP(&agent, "agent", "a", "", "agent handle to wake (required)")
	cmd.Flags().StringVarP(&prompt, "prompt", "p", "", "prompt to pass to agent when triggered")

	// Cron options
	cmd.Flags().StringVar(&schedule, "schedule", "", "cron schedule (required for cron type)")

	// Watch options
	cmd.Flags().StringVar(&path, "path", "", "path to watch (required for watch type)")
	cmd.Flags().StringSliceVar(&patterns, "patterns", nil, "file patterns to match (e.g., *.go)")
	cmd.Flags().BoolVar(&recursive, "recursive", false, "watch subdirectories")
	cmd.Flags().StringSliceVar(&events, "events", nil, "events to trigger on: create, modify, delete")

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}

func removeTriggerCmd() *cobra.Command {
	var (
		force bool
		quiet bool
	)

	cmd := &cobra.Command{
		Use:     "rm <id>",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove a trigger",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			ctx := cmd.Context()

			client, err := daemon.ConnectOrStart(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			// Confirm unless forced
			if !force {
				fmt.Printf("Remove trigger %s? (y/N): ", id)
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					return fmt.Errorf("cancelled")
				}
			}

			if err := client.TriggerRemove(ctx, id); err != nil {
				return fmt.Errorf("remove trigger: %w", err)
			}

			if !quiet {
				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("✓ Removed trigger: " + id))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")

	return cmd
}

func testTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <id>",
		Short: "Fire a trigger manually",
		Long: `Fire a trigger manually for testing purposes.

This will wake the associated agent just as if the trigger
had fired naturally.

Examples:
  ayo triggers test trig_123456789`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			ctx := cmd.Context()

			client, err := daemon.ConnectOrStart(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.TriggerTest(ctx, id); err != nil {
				return fmt.Errorf("test trigger: %w", err)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render("✓ Trigger fired: " + id))

			return nil
		},
	}

	return cmd
}
