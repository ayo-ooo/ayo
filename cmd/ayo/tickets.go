package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/daemon"
)

// Styles for ticket output
var (
	ticketIDStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#67e8f9")) // cyan
	ticketTitleStyle  = lipgloss.NewStyle().Bold(true)
	ticketMutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	ticketGreenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#34d399"))
	ticketYellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24"))
	ticketRedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))
)

func statusStyle(status string) lipgloss.Style {
	switch status {
	case "closed":
		return ticketGreenStyle
	case "in_progress":
		return ticketYellowStyle
	case "blocked":
		return ticketRedStyle
	default:
		return ticketMutedStyle
	}
}

func newTicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ticket",
		Aliases: []string{"tickets"},
		Short:   "Manage task tickets",
		Long: `Manage tickets for agent coordination.

Tickets are tasks that agents work on. They have status, dependencies,
and can be assigned to specific agents.

Examples:
  # List all tickets in the current session
  ayo ticket list

  # Create a new ticket
  ayo ticket create "Implement login" -a @coder

  # Start working on a ticket
  ayo ticket start <id>

  # Close a completed ticket
  ayo ticket close <id>

  # Show tickets ready to work on
  ayo ticket ready`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to list
			return ticketListCmd().RunE(cmd, args)
		},
	}

	cmd.AddCommand(ticketListCmd())
	cmd.AddCommand(ticketCreateCmd())
	cmd.AddCommand(ticketShowCmd())
	cmd.AddCommand(ticketStartCmd())
	cmd.AddCommand(ticketCloseCmd())
	cmd.AddCommand(ticketReopenCmd())
	cmd.AddCommand(ticketBlockCmd())
	cmd.AddCommand(ticketAssignCmd())
	cmd.AddCommand(ticketNoteCmd())
	cmd.AddCommand(ticketReadyCmd())
	cmd.AddCommand(ticketBlockedCmd())
	cmd.AddCommand(ticketDeleteCmd())
	cmd.AddCommand(ticketDepCmd())

	return cmd
}

func ticketListCmd() *cobra.Command {
	var (
		sessionID  string
		status     string
		assignee   string
		ticketType string
		tags       []string
		parent     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tickets",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			result, err := client.TicketList(ctx, daemon.TicketListParams{
				SessionID: sessionID,
				Status:    status,
				Assignee:  assignee,
				Type:      ticketType,
				Tags:      tags,
				Parent:    parent,
			})
			if err != nil {
				return err
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			if globalOutput.Quiet {
				for _, t := range result.Tickets {
					fmt.Println(t.ID)
				}
				return nil
			}

			if len(result.Tickets) == 0 {
				fmt.Println(ticketMutedStyle.Render("No tickets found"))
				return nil
			}

			for _, t := range result.Tickets {
				printTicketLine(t)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	cmd.Flags().StringVar(&status, "status", "", "filter by status (open, in_progress, blocked, closed)")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "filter by assignee")
	cmd.Flags().StringVarP(&ticketType, "type", "t", "", "filter by type (epic, feature, task, bug, chore)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "filter by tags")
	cmd.Flags().StringVar(&parent, "parent", "", "filter by parent ticket")

	return cmd
}

func ticketCreateCmd() *cobra.Command {
	var (
		sessionID   string
		description string
		ticketType  string
		priority    int
		assignee    string
		deps        []string
		parent      string
		tags        []string
		externalRef string
	)

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			title := args[0]

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			result, err := client.TicketCreate(ctx, daemon.TicketCreateParams{
				SessionID:   sessionID,
				Title:       title,
				Description: description,
				Type:        ticketType,
				Priority:    priority,
				Assignee:    assignee,
				Deps:        deps,
				Parent:      parent,
				Tags:        tags,
				ExternalRef: externalRef,
			})
			if err != nil {
				return err
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			if globalOutput.Quiet {
				fmt.Println(result.ID)
				return nil
			}

			fmt.Println(ticketGreenStyle.Render("✓ Created:"), ticketIDStyle.Render(result.ID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "ticket description")
	cmd.Flags().StringVarP(&ticketType, "type", "t", "task", "type (epic, feature, task, bug, chore)")
	cmd.Flags().IntVarP(&priority, "priority", "p", 2, "priority 0-4 (0=highest)")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "assign to agent")
	cmd.Flags().StringSliceVar(&deps, "deps", nil, "dependency ticket IDs")
	cmd.Flags().StringVar(&parent, "parent", "", "parent ticket ID")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "tags")
	cmd.Flags().StringVar(&externalRef, "ref", "", "external reference (e.g., GitHub issue)")

	return cmd
}

func ticketShowCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show ticket details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ticketID := args[0]

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			result, err := client.TicketGet(ctx, sessionID, ticketID)
			if err != nil {
				return err
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			printTicketFull(result.Ticket)
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	return cmd
}

func ticketStartCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "start <id>",
		Short: "Start working on a ticket (set to in_progress)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			ticketID, err := resolveTicketArg(ctx, client, sessionID, args, "Select ticket to start")
			if err != nil {
				return err
			}

			if err := client.TicketStart(ctx, sessionID, ticketID); err != nil {
				return err
			}

			if globalOutput.Quiet {
				return nil
			}

			fmt.Println(ticketYellowStyle.Render("▶ Started:"), ticketIDStyle.Render(ticketID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	return cmd
}

func ticketCloseCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "close <id>",
		Short: "Close a ticket (mark as done)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			ticketID, err := resolveTicketArg(ctx, client, sessionID, args, "Select ticket to close")
			if err != nil {
				return err
			}

			if err := client.TicketClose(ctx, sessionID, ticketID); err != nil {
				return err
			}

			if globalOutput.Quiet {
				return nil
			}

			fmt.Println(ticketGreenStyle.Render("✓ Closed:"), ticketIDStyle.Render(ticketID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	return cmd
}

func ticketReopenCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "reopen <id>",
		Short: "Reopen a closed ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ticketID := args[0]

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			if err := client.TicketReopen(ctx, sessionID, ticketID); err != nil {
				return err
			}

			if globalOutput.Quiet {
				return nil
			}

			fmt.Println(ticketMutedStyle.Render("↺ Reopened:"), ticketIDStyle.Render(ticketID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	return cmd
}

func ticketBlockCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "block <id>",
		Short: "Mark ticket as blocked",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ticketID := args[0]

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			if err := client.TicketBlock(ctx, sessionID, ticketID); err != nil {
				return err
			}

			if globalOutput.Quiet {
				return nil
			}

			fmt.Println(ticketRedStyle.Render("⊘ Blocked:"), ticketIDStyle.Render(ticketID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	return cmd
}

func ticketAssignCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "assign <id> <agent>",
		Short: "Assign ticket to an agent",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ticketID := args[0]
			assignee := args[1]

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			if err := client.TicketAssign(ctx, sessionID, ticketID, assignee); err != nil {
				return err
			}

			if globalOutput.Quiet {
				return nil
			}

			fmt.Printf("%s %s → %s\n",
				ticketGreenStyle.Render("✓ Assigned:"),
				ticketIDStyle.Render(ticketID),
				assignee)
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	return cmd
}

func ticketNoteCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "note <id> <content>",
		Short: "Add a note to a ticket",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ticketID := args[0]
			content := strings.Join(args[1:], " ")

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			if err := client.TicketAddNote(ctx, sessionID, ticketID, content); err != nil {
				return err
			}

			if globalOutput.Quiet {
				return nil
			}

			fmt.Println(ticketGreenStyle.Render("✓ Note added to"), ticketIDStyle.Render(ticketID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	return cmd
}

func ticketReadyCmd() *cobra.Command {
	var (
		sessionID string
		assignee  string
	)

	cmd := &cobra.Command{
		Use:   "ready",
		Short: "List tickets ready to work on (dependencies resolved)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			result, err := client.TicketReady(ctx, sessionID, assignee)
			if err != nil {
				return err
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			if globalOutput.Quiet {
				for _, t := range result.Tickets {
					fmt.Println(t.ID)
				}
				return nil
			}

			if len(result.Tickets) == 0 {
				fmt.Println(ticketMutedStyle.Render("No tickets ready"))
				return nil
			}

			fmt.Println(ticketTitleStyle.Render("Ready to work:"))
			for _, t := range result.Tickets {
				printTicketLine(t)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "filter by assignee")

	return cmd
}

func ticketBlockedCmd() *cobra.Command {
	var (
		sessionID string
		assignee  string
	)

	cmd := &cobra.Command{
		Use:   "blocked",
		Short: "List tickets blocked on dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			result, err := client.TicketBlocked(ctx, sessionID, assignee)
			if err != nil {
				return err
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			if globalOutput.Quiet {
				for _, t := range result.Tickets {
					fmt.Println(t.ID)
				}
				return nil
			}

			if len(result.Tickets) == 0 {
				fmt.Println(ticketMutedStyle.Render("No blocked tickets"))
				return nil
			}

			fmt.Println(ticketTitleStyle.Render("Blocked:"))
			for _, t := range result.Tickets {
				printTicketLine(t)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "filter by assignee")

	return cmd
}

func ticketDeleteCmd() *cobra.Command {
	var (
		sessionID string
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ticketID := args[0]

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			if !force && !globalOutput.Quiet {
				var confirm bool
				err := huh.NewConfirm().
					Title(fmt.Sprintf("Delete ticket %s?", ticketID)).
					Value(&confirm).
					Run()
				if err != nil {
					return err
				}
				if !confirm {
					return nil
				}
			}

			if err := client.TicketDelete(ctx, sessionID, ticketID); err != nil {
				return err
			}

			if !globalOutput.Quiet {
				fmt.Println(ticketRedStyle.Render("✗ Deleted:"), ticketIDStyle.Render(ticketID))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")

	return cmd
}

func ticketDepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "Manage ticket dependencies",
	}

	cmd.AddCommand(ticketDepAddCmd())
	cmd.AddCommand(ticketDepRemoveCmd())

	return cmd
}

func ticketDepAddCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "add <ticket-id> <dependency-id>",
		Short: "Add a dependency to a ticket",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ticketID := args[0]
			depID := args[1]

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			if err := client.TicketAddDep(ctx, sessionID, ticketID, depID); err != nil {
				return err
			}

			if !globalOutput.Quiet {
				fmt.Printf("%s %s → %s\n",
					ticketGreenStyle.Render("✓ Dependency added:"),
					ticketIDStyle.Render(ticketID),
					ticketIDStyle.Render(depID))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	return cmd
}

func ticketDepRemoveCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "remove <ticket-id> <dependency-id>",
		Short: "Remove a dependency from a ticket",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ticketID := args[0]
			depID := args[1]

			client, err := connectToDaemon(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if sessionID == "" {
				sessionID = getCurrentSessionID()
			}

			if err := client.TicketRemoveDep(ctx, sessionID, ticketID, depID); err != nil {
				return err
			}

			if !globalOutput.Quiet {
				fmt.Printf("%s %s ↛ %s\n",
					ticketYellowStyle.Render("✓ Dependency removed:"),
					ticketIDStyle.Render(ticketID),
					ticketIDStyle.Render(depID))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "session ID (default: current)")
	return cmd
}

// Helper functions

func printTicketLine(t daemon.TicketInfo) {
	// ID  STATUS  TYPE  ASSIGNEE  TITLE
	id := ticketIDStyle.Render(t.ID)
	status := statusStyle(t.Status).Render(t.Status)
	title := t.Title

	var parts []string
	parts = append(parts, id)
	parts = append(parts, status)
	if t.Assignee != "" {
		parts = append(parts, ticketMutedStyle.Render(t.Assignee))
	}
	parts = append(parts, title)

	fmt.Println(strings.Join(parts, "  "))
}

func printTicketFull(t daemon.TicketInfo) {
	fmt.Println(ticketTitleStyle.Render("# " + t.Title))
	fmt.Println()

	fmt.Printf("  %s  %s\n", ticketMutedStyle.Render("ID:"), ticketIDStyle.Render(t.ID))
	fmt.Printf("  %s  %s\n", ticketMutedStyle.Render("Status:"), statusStyle(t.Status).Render(t.Status))
	fmt.Printf("  %s  %s\n", ticketMutedStyle.Render("Type:"), t.Type)
	fmt.Printf("  %s  %d\n", ticketMutedStyle.Render("Priority:"), t.Priority)

	if t.Assignee != "" {
		fmt.Printf("  %s  %s\n", ticketMutedStyle.Render("Assignee:"), t.Assignee)
	}

	if len(t.Deps) > 0 {
		fmt.Printf("  %s  %s\n", ticketMutedStyle.Render("Deps:"), strings.Join(t.Deps, ", "))
	}

	if t.Parent != "" {
		fmt.Printf("  %s  %s\n", ticketMutedStyle.Render("Parent:"), t.Parent)
	}

	if len(t.Tags) > 0 {
		fmt.Printf("  %s  %s\n", ticketMutedStyle.Render("Tags:"), strings.Join(t.Tags, ", "))
	}

	created := time.Unix(t.Created, 0).Format(time.RFC3339)
	fmt.Printf("  %s  %s\n", ticketMutedStyle.Render("Created:"), created)

	if t.Started > 0 {
		started := time.Unix(t.Started, 0).Format(time.RFC3339)
		fmt.Printf("  %s  %s\n", ticketMutedStyle.Render("Started:"), started)
	}

	if t.Closed > 0 {
		closed := time.Unix(t.Closed, 0).Format(time.RFC3339)
		fmt.Printf("  %s  %s\n", ticketMutedStyle.Render("Closed:"), closed)
	}

	if t.Description != "" {
		fmt.Println()
		fmt.Println(t.Description)
	}
}

func resolveTicketArg(ctx context.Context, client *daemon.Client, sessionID string, args []string, prompt string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	// No ID provided - show picker
	result, err := client.TicketList(ctx, daemon.TicketListParams{
		SessionID: sessionID,
	})
	if err != nil {
		return "", err
	}

	if len(result.Tickets) == 0 {
		return "", fmt.Errorf("no tickets found")
	}

	// Build options
	options := make([]huh.Option[string], 0, len(result.Tickets))
	for _, t := range result.Tickets {
		if t.Status == "closed" {
			continue // Skip closed tickets
		}
		label := fmt.Sprintf("%s  %s  %s", t.ID, statusStyle(t.Status).Render(t.Status), t.Title)
		options = append(options, huh.NewOption(label, t.ID))
	}

	if len(options) == 0 {
		return "", fmt.Errorf("no open tickets found")
	}

	if len(options) == 1 {
		return options[0].Value, nil
	}

	var selected string
	err = huh.NewSelect[string]().
		Title(prompt).
		Options(options...).
		Value(&selected).
		Run()
	if err != nil {
		return "", err
	}

	return selected, nil
}

func getCurrentSessionID() string {
	// Check environment variable
	if id := os.Getenv("AYO_SESSION_ID"); id != "" {
		return id
	}
	// Check for .tickets directory in current working directory for sandbox context
	if id := os.Getenv("TICKETS_SESSION"); id != "" {
		return id
	}
	// Default to "default" session for development
	return "default"
}
