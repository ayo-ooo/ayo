package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/cli"
	"github.com/alexcabrera/ayo/internal/daemon"
)

// Ensure cli package is used
var _ = cli.Output{}

// connectToDaemon connects to the daemon with a spinner if auto-starting.
func connectToDaemon(ctx context.Context) (*daemon.Client, error) {
	client := daemon.NewClient()

	// Try to connect first (fast path)
	if err := client.Connect(ctx); err == nil {
		if err := client.Ping(ctx); err == nil {
			return client, nil
		}
		client.Close()
	}

	// Daemon not running - start with spinner feedback
	if globalOutput.JSON || globalOutput.Quiet {
		// No spinner for JSON/quiet mode
		return daemon.ConnectOrStart(ctx)
	}

	// Start daemon in background
	if err := daemon.StartDaemonBackground(); err != nil {
		return nil, fmt.Errorf("start daemon: %w", err)
	}

	// Wait for daemon with spinner
	var result *daemon.Client
	var connectErr error

	spinErr := spinner.New().
		Title("Starting service...").
		Type(spinner.Dots).
		Style(lipgloss.NewStyle().Foreground(lipgloss.Color("212"))).
		ActionWithErr(func(_ context.Context) error {
			deadline := time.Now().Add(45 * time.Second) // Longer timeout for sandbox creation
			for time.Now().Before(deadline) {
				time.Sleep(200 * time.Millisecond)
				client := daemon.NewClient()
				if err := client.Connect(ctx); err == nil {
					if err := client.Ping(ctx); err == nil {
						result = client
						return nil
					}
					client.Close()
				}
			}
			connectErr = fmt.Errorf("daemon started but not responding")
			return connectErr
		}).
		Run()

	if spinErr != nil {
		return nil, spinErr
	}
	if connectErr != nil {
		return nil, connectErr
	}
	return result, nil
}

func newTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "trigger",
		Short:   "Manage triggers",
		Aliases: []string{"triggers"}, // Hidden backwards-compat alias
		Long: `Manage triggers that wake agents on events.

Triggers can be:
  cron    Schedule-based triggers (e.g., "every day at 9am")
  watch   File system triggers (e.g., "when a file changes")

Examples:
  # Create a scheduled trigger
  ayo trigger schedule @backup "0 0 2 * * *"

  # Create a watch trigger
  ayo trigger watch ./src @build "*.go"

  # List all triggers
  ayo trigger list

  # Test a trigger
  ayo trigger test trig_123456789

  # Remove a trigger
  ayo trigger rm trig_123456789`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTriggersCmd().RunE(cmd, args)
		},
	}

	cmd.AddCommand(listTriggersCmd())
	cmd.AddCommand(showTriggerCmd())
	cmd.AddCommand(historyTriggerCmd())
	cmd.AddCommand(createTriggerCmd())
	cmd.AddCommand(addTriggerCmd())
	cmd.AddCommand(scheduleCmd())
	cmd.AddCommand(watchCmd())
	cmd.AddCommand(removeTriggerCmd())
	cmd.AddCommand(testTriggerCmd())
	cmd.AddCommand(enableTriggerCmd())
	cmd.AddCommand(disableTriggerCmd())

	return cmd
}

func listTriggersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all registered triggers",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
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
					fmt.Println("Add one with: ayo trigger add --type cron --agent @name --schedule \"0 0 * * * *\"")
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
		Use:   "show [id]",
		Short: "Show trigger details",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			id, err := resolveTriggerID(ctx, client, query, "Select trigger to show")
			if err != nil {
				return err
			}

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

func historyTriggerCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "history [id]",
		Short: "Show trigger run history",
		Long: `Show the execution history for a trigger.

Displays recent runs with start time, duration, and status.

Examples:
  ayo trigger history trig_123456789
  ayo trigger history trig_123456789 --limit 100`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			id, err := resolveTriggerID(ctx, client, query, "Select trigger to view history")
			if err != nil {
				return err
			}

			result, err := client.TriggerHistory(ctx, id, limit)
			if err != nil {
				return fmt.Errorf("get trigger history: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result.Runs)
			}

			if len(result.Runs) == 0 {
				if !globalOutput.Quiet {
					fmt.Println("No run history for this trigger")
				}
				return nil
			}

			// Styles
			purple := lipgloss.Color("#a78bfa")
			cyan := lipgloss.Color("#67e8f9")
			muted := lipgloss.Color("#6b7280")
			subtle := lipgloss.Color("#374151")
			green := lipgloss.Color("#34d399")
			red := lipgloss.Color("#f87171")

			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
			dividerStyle := lipgloss.NewStyle().Foreground(subtle)
			mutedStyle := lipgloss.NewStyle().Foreground(muted)
			successStyle := lipgloss.NewStyle().Foreground(green)
			failedStyle := lipgloss.NewStyle().Foreground(red)
			_ = cyan

			fmt.Println()
			fmt.Println(headerStyle.Render("  Run History: " + id))
			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 70)))
			fmt.Println()

			// Header row
			fmt.Printf("  %-22s %-12s %-10s %s\n",
				mutedStyle.Render("STARTED"),
				mutedStyle.Render("DURATION"),
				mutedStyle.Render("STATUS"),
				mutedStyle.Render("ERROR"))

			for _, run := range result.Runs {
				startStr := run.StartedAt.Format("2006-01-02 15:04:05")

				durationStr := "-"
				if run.Duration > 0 {
					if run.Duration < 1000 {
						durationStr = fmt.Sprintf("%dms", run.Duration)
					} else {
						durationStr = fmt.Sprintf("%.1fs", float64(run.Duration)/1000)
					}
				}

				statusStr := run.Status
				switch run.Status {
				case "success":
					statusStr = successStyle.Render("✓ success")
				case "failed":
					statusStr = failedStyle.Render("✗ failed")
				case "running":
					statusStr = mutedStyle.Render("⟳ running")
				}

				errorStr := ""
				if run.ErrorMessage != "" {
					if len(run.ErrorMessage) > 30 {
						errorStr = run.ErrorMessage[:27] + "..."
					} else {
						errorStr = run.ErrorMessage
					}
				}

				fmt.Printf("  %-22s %-12s %-10s %s\n", startStr, durationStr, statusStr, errorStr)
			}

			fmt.Println()
			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 70)))
			fmt.Println(mutedStyle.Render(fmt.Sprintf("  %d runs", len(result.Runs))))
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "number of runs to show")

	return cmd
}

func createTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new trigger",
		Long: `Create a new trigger to wake an agent.

Use one of the specialized subcommands:

  ayo trigger schedule @agent "schedule"  # Time-based trigger
  ayo trigger watch <path> @agent          # File-watching trigger

Examples:
  # Schedule trigger (every hour)
  ayo trigger schedule @backup "every hour"
  
  # Schedule trigger (cron syntax)
  ayo trigger schedule @reports "0 9 * * *"
  
  # Watch trigger (on file changes)
  ayo trigger watch ./src @build "*.go"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show help for create command
			return cmd.Help()
		},
	}

	// Add schedule and watch as subcommands for discoverability
	cmd.AddCommand(scheduleCmd())
	cmd.AddCommand(watchCmd())

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
		Use:    "add",
		Short:  "Add a new trigger (deprecated: use 'schedule' or 'watch')",
		Hidden: true,
		Long: `Add a new trigger to wake an agent.

DEPRECATED: Use 'ayo trigger schedule' or 'ayo trigger watch' instead.

Cron triggers use standard cron syntax with seconds:
  seconds minutes hours day-of-month month day-of-week

Examples:
  # Use these instead:
  ayo trigger schedule @backup "0 0 * * * *"
  ayo trigger watch ./src @build "*.go"`,
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

			client, err := connectToDaemon(ctx)
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
	)

	cmd := &cobra.Command{
		Use:     "rm [id]",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove a trigger",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			id, err := resolveTriggerID(ctx, client, query, "Select trigger to remove")
			if err != nil {
				return err
			}

			// Confirm unless forced
			if !force && !globalOutput.Quiet {
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

			if !globalOutput.Quiet {
				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("✓ Removed trigger: " + id))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")

	return cmd
}

func testTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [id]",
		Short: "Fire a trigger manually",
		Long: `Fire a trigger manually for testing purposes.

This will wake the associated agent just as if the trigger
had fired naturally.

Examples:
  ayo trigger test trig_123456789
  ayo trigger test trig_  # prefix match`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			id, err := resolveTriggerID(ctx, client, query, "Select trigger to test")
			if err != nil {
				return err
			}

			if err := client.TriggerTest(ctx, id); err != nil {
				return fmt.Errorf("test trigger: %w", err)
			}

			if !globalOutput.Quiet {
				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("✓ Trigger fired: " + id))
			}

			return nil
		},
	}

	return cmd
}

func enableTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable [id]",
		Short: "Enable a disabled trigger",
		Long: `Enable a previously disabled trigger.

The trigger will resume firing on its schedule or watch conditions.

Examples:
  ayo trigger enable trig_123456789`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			id, err := resolveTriggerID(ctx, client, query, "Select trigger to enable")
			if err != nil {
				return err
			}

			if err := client.TriggerSetEnabled(ctx, id, true); err != nil {
				return fmt.Errorf("enable trigger: %w", err)
			}

			if !globalOutput.Quiet {
				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("✓ Trigger enabled: " + id))
			}

			return nil
		},
	}

	return cmd
}

func disableTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable [id]",
		Short: "Disable a trigger",
		Long: `Disable a trigger without removing it.

The trigger will stop firing but can be re-enabled later.

Examples:
  ayo trigger disable trig_123456789`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			id, err := resolveTriggerID(ctx, client, query, "Select trigger to disable")
			if err != nil {
				return err
			}

			if err := client.TriggerSetEnabled(ctx, id, false); err != nil {
				return fmt.Errorf("disable trigger: %w", err)
			}

			if !globalOutput.Quiet {
				warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
				fmt.Println(warnStyle.Render("⏸ Trigger disabled: " + id))
			}

			return nil
		},
	}

	return cmd
}

func scheduleCmd() *cobra.Command {
	var (
		prompt string
	)

	cmd := &cobra.Command{
		Use:   "schedule <agent> <schedule>",
		Short: "Create a scheduled trigger",
		Long: `Create a cron-based trigger that wakes an agent on a schedule.

Supports natural language or cron syntax:

Natural language examples:
  ayo trigger schedule @backup "every hour"
  ayo trigger schedule @reports "every day at 9am"
  ayo trigger schedule @cleanup "every monday at 3pm"
  ayo trigger schedule @weekly "daily"

Cron syntax (with seconds):
  ayo trigger schedule @backup "0 0 * * * *"      # every hour
  ayo trigger schedule @reports "0 0 9 * * *"     # every day at 9am
  ayo trigger schedule @weekly "0 0 9 * * MON"    # every Monday at 9am

Supported patterns:
  hourly, daily, weekly, monthly, yearly
  every hour, every day, every monday
  every day at 9am, every monday at 3pm
  every 5 minutes, every 2 hours`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			agent := args[0]
			scheduleInput := args[1]
			ctx := cmd.Context()

			// Parse natural language to cron
			schedule, err := cli.ParseSchedule(scheduleInput)
			if err != nil {
				return err
			}

			client, err := connectToDaemon(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			params := daemon.TriggerRegisterParams{
				Type:     "cron",
				Agent:    agent,
				Prompt:   prompt,
				Schedule: schedule,
			}

			result, err := client.TriggerRegister(ctx, params)
			if err != nil {
				return fmt.Errorf("register trigger: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result.Trigger)
			}

			if globalOutput.Quiet {
				fmt.Println(result.Trigger.ID)
				return nil
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render("✓ Created trigger: " + result.Trigger.ID))
			fmt.Printf("  Type:     cron\n")
			fmt.Printf("  Agent:    %s\n", agent)
			if scheduleInput != schedule {
				// Show both natural language and parsed cron
				fmt.Printf("  Schedule: %s (%s)\n", scheduleInput, schedule)
			} else {
				fmt.Printf("  Schedule: %s\n", schedule)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&prompt, "prompt", "p", "", "prompt to pass to agent when triggered")

	return cmd
}

func watchCmd() *cobra.Command {
	var (
		prompt    string
		recursive bool
		events    []string
	)

	cmd := &cobra.Command{
		Use:   "watch <path> <agent> [patterns...]",
		Short: "Create a filesystem watch trigger",
		Long: `Create a trigger that wakes an agent when files change.

Examples:
  # Watch directory for any changes
  ayo trigger watch ./src @build

  # Watch for specific file patterns
  ayo trigger watch ./src @build "*.go" "*.mod"

  # Watch recursively with events filter
  ayo trigger watch ./docs @docs "*.md" --recursive --events modify,create`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			agent := args[1]
			patterns := args[2:]
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			params := daemon.TriggerRegisterParams{
				Type:      "watch",
				Agent:     agent,
				Prompt:    prompt,
				Path:      path,
				Patterns:  patterns,
				Recursive: recursive,
				Events:    events,
			}

			result, err := client.TriggerRegister(ctx, params)
			if err != nil {
				return fmt.Errorf("register trigger: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result.Trigger)
			}

			if globalOutput.Quiet {
				fmt.Println(result.Trigger.ID)
				return nil
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render("✓ Created trigger: " + result.Trigger.ID))
			fmt.Printf("  Type:  watch\n")
			fmt.Printf("  Agent: %s\n", agent)
			fmt.Printf("  Path:  %s\n", path)
			if len(patterns) > 0 {
				fmt.Printf("  Patterns: %s\n", strings.Join(patterns, ", "))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&prompt, "prompt", "p", "", "prompt to pass to agent when triggered")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "watch subdirectories")
	cmd.Flags().StringSliceVar(&events, "events", nil, "events to trigger on: create, modify, delete")

	return cmd
}

// resolveTriggerID resolves a trigger ID from a query (full ID or prefix).
// If query is empty, shows a picker or auto-selects if only one trigger exists.
func resolveTriggerID(ctx context.Context, client *daemon.Client, query string, title string) (string, error) {
	result, err := client.TriggerList(ctx)
	if err != nil {
		return "", cli.WrapWithSuggestion(err, "Check if service is running with 'ayo service status'")
	}

	triggers := result.Triggers
	if len(triggers) == 0 {
		return "", &cli.CLIError{
			Brief:      "No triggers registered",
			Suggestion: "Create one with 'ayo trigger schedule @agent \"0 * * * * *\"'",
			Code:       cli.ExitNotFound,
		}
	}

	// If query provided, do prefix matching
	if query != "" {
		var matches []daemon.TriggerInfo
		for _, t := range triggers {
			if t.ID == query || strings.HasPrefix(t.ID, query) {
				matches = append(matches, t)
			}
		}

		if len(matches) == 0 {
			return "", cli.ErrTriggerNotFound(query)
		}
		if len(matches) == 1 {
			return matches[0].ID, nil
		}
		// Multiple prefix matches - use picker
		triggers = matches
	}

	// Auto-select if only one trigger
	if len(triggers) == 1 {
		return triggers[0].ID, nil
	}

	// Show picker for multiple triggers
	if globalOutput.JSON || globalOutput.Quiet {
		return "", cli.ErrInvalidInputWithSuggestion(
			"Multiple triggers match",
			"Specify a more specific ID prefix",
		)
	}

	options := make([]huh.Option[string], len(triggers))
	for i, t := range triggers {
		config := ""
		if t.Type == "cron" {
			config = t.Schedule
		} else if t.Type == "watch" {
			config = t.Path
		}
		if len(config) > 20 {
			config = config[:17] + "..."
		}
		label := fmt.Sprintf("%-20s %-8s %-15s %s", t.ID, t.Type, t.Agent, config)
		options[i] = huh.NewOption(label, t.ID)
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
