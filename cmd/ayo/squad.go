package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/daemon"
)

// Squad status styles
var (
	squadNameStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#67e8f9")).Bold(true)
	squadStatusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#34d399"))
	squadMutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
)

func squadStatusColor(status string) lipgloss.Style {
	switch status {
	case "running":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#34d399"))
	case "stopped":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24"))
	case "failed":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))
	default:
		return squadMutedStyle
	}
}

func newSquadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "squad",
		Aliases: []string{"squads"},
		Short:   "Manage agent squads",
		Long: `Manage squad sandboxes for agent team coordination.

Squads are isolated sandboxes where multiple agents can collaborate on tasks.
Each squad has its own workspace, tickets directory, and context.

Examples:
  # List all squads
  ayo squad list

  # Create a new squad
  ayo squad create frontend --description "Frontend development team"

  # Add an agent to a squad
  ayo squad add-agent frontend @react-dev

  # Show squad details
  ayo squad show frontend

  # Destroy a squad
  ayo squad destroy frontend`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return squadListCmd().RunE(cmd, args)
		},
	}

	cmd.AddCommand(squadListCmd())
	cmd.AddCommand(squadCreateCmd())
	cmd.AddCommand(squadShowCmd())
	cmd.AddCommand(squadDestroyCmd())
	cmd.AddCommand(squadStartCmd())
	cmd.AddCommand(squadStopCmd())
	cmd.AddCommand(squadAddAgentCmd())
	cmd.AddCommand(squadRemoveAgentCmd())
	cmd.AddCommand(squadTicketCmd())

	return cmd
}

func squadListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all squads",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			result, err := client.SquadList(ctx)
			if err != nil {
				return fmt.Errorf("list squads: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			if len(result.Squads) == 0 {
				fmt.Println(squadMutedStyle.Render("No squads configured"))
				return nil
			}

			for _, squad := range result.Squads {
				status := squadStatusColor(squad.Status).Render(fmt.Sprintf("[%s]", squad.Status))
				name := squadNameStyle.Render(squad.Name)
				
				line := fmt.Sprintf("%s %s", name, status)
				if squad.Description != "" {
					line += " - " + squadMutedStyle.Render(squad.Description)
				}
				fmt.Println(line)

				if len(squad.Agents) > 0 {
					agents := squadMutedStyle.Render("  agents: " + strings.Join(squad.Agents, ", "))
					fmt.Println(agents)
				}
			}

			return nil
		},
	}

	return cmd
}

func squadCreateCmd() *cobra.Command {
	var (
		description    string
		image          string
		ephemeral      bool
		agents         []string
		workspaceMount string
		packages       []string
		outputPath     string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new squad",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			params := daemon.SquadCreateParams{
				Name:           name,
				Description:    description,
				Image:          image,
				Ephemeral:      ephemeral,
				Agents:         agents,
				WorkspaceMount: workspaceMount,
				Packages:       packages,
				OutputPath:     outputPath,
			}

			result, err := client.SquadCreate(ctx, params)
			if err != nil {
				return fmt.Errorf("create squad: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			fmt.Printf("Created squad %s\n", squadNameStyle.Render(name))
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Squad description")
	cmd.Flags().StringVar(&image, "image", "", "Container image")
	cmd.Flags().BoolVar(&ephemeral, "ephemeral", false, "Create ephemeral squad")
	cmd.Flags().StringSliceVarP(&agents, "agents", "a", nil, "Initial agents")
	cmd.Flags().StringVar(&workspaceMount, "workspace", "", "Host directory to mount as workspace")
	cmd.Flags().StringSliceVarP(&packages, "packages", "p", nil, "Packages to install")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Host directory for syncing work products")

	return cmd
}

func squadShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show squad details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			result, err := client.SquadGet(ctx, name)
			if err != nil {
				return fmt.Errorf("get squad: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			squad := result.Squad
			fmt.Printf("%s %s\n", squadNameStyle.Render(squad.Name), squadStatusColor(squad.Status).Render(fmt.Sprintf("[%s]", squad.Status)))
			
			if squad.Description != "" {
				fmt.Printf("Description: %s\n", squad.Description)
			}
			if squad.Ephemeral {
				fmt.Println("Ephemeral: yes")
			}
			if len(squad.Agents) > 0 {
				fmt.Printf("Agents: %s\n", strings.Join(squad.Agents, ", "))
			}
			fmt.Printf("Tickets: %s\n", squadMutedStyle.Render(squad.TicketsDir))
			fmt.Printf("Context: %s\n", squadMutedStyle.Render(squad.ContextDir))
			fmt.Printf("Workspace: %s\n", squadMutedStyle.Render(squad.WorkspaceDir))

			return nil
		},
	}

	return cmd
}

func squadDestroyCmd() *cobra.Command {
	var deleteData bool

	cmd := &cobra.Command{
		Use:   "destroy <name>",
		Short: "Destroy a squad",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.SquadDestroy(ctx, name, deleteData); err != nil {
				return fmt.Errorf("destroy squad: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]bool{"success": true})
			}

			fmt.Printf("Destroyed squad %s\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&deleteData, "delete-data", false, "Also delete squad data directories")

	return cmd
}

func squadStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <name>",
		Short: "Start a squad sandbox",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.SquadStart(ctx, name); err != nil {
				return fmt.Errorf("start squad: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]bool{"success": true})
			}

			fmt.Printf("Started squad %s\n", name)
			return nil
		},
	}

	return cmd
}

func squadStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <name>",
		Short: "Stop a squad sandbox",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.SquadStop(ctx, name); err != nil {
				return fmt.Errorf("stop squad: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]bool{"success": true})
			}

			fmt.Printf("Stopped squad %s\n", name)
			return nil
		},
	}

	return cmd
}

func squadAddAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-agent <squad> <agent>",
		Short: "Add an agent to a squad",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			squadName := args[0]
			agentHandle := args[1]

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.SquadAddAgent(ctx, squadName, agentHandle); err != nil {
				return fmt.Errorf("add agent: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]bool{"success": true})
			}

			fmt.Printf("Added %s to squad %s\n", agentHandle, squadName)
			return nil
		},
	}

	return cmd
}

func squadRemoveAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-agent <squad> <agent>",
		Short: "Remove an agent from a squad",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			squadName := args[0]
			agentHandle := args[1]

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.SquadRemoveAgent(ctx, squadName, agentHandle); err != nil {
				return fmt.Errorf("remove agent: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]bool{"success": true})
			}

			fmt.Printf("Removed %s from squad %s\n", agentHandle, squadName)
			return nil
		},
	}

	return cmd
}

func squadTicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ticket <squad> <command>",
		Short: "Manage tickets within a squad",
		Long: `Manage tickets within a squad's .tickets directory.

Examples:
  ayo squad ticket myteam create "Implement login" -a @backend
  ayo squad ticket myteam list
  ayo squad ticket myteam show abc-1234
  ayo squad ticket myteam start abc-1234
  ayo squad ticket myteam close abc-1234`,
	}

	cmd.AddCommand(squadTicketCreateCmd())
	cmd.AddCommand(squadTicketListCmd())
	cmd.AddCommand(squadTicketStartCmd())
	cmd.AddCommand(squadTicketCloseCmd())

	return cmd
}

func squadTicketCreateCmd() *cobra.Command {
	var (
		assignee    string
		description string
		deps        []string
		priority    int
		ticketType  string
	)

	cmd := &cobra.Command{
		Use:   "create <squad> <title>",
		Short: "Create a ticket in a squad",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			squadName := args[0]
			title := args[1]
			ctx := cmd.Context()

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			result, err := client.TicketCreate(ctx, daemon.TicketCreateParams{
				SquadName:   squadName,
				Title:       title,
				Description: description,
				Type:        ticketType,
				Priority:    priority,
				Assignee:    assignee,
				Deps:        deps,
			})
			if err != nil {
				return fmt.Errorf("create ticket: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			fmt.Printf("✓ Created: %s\n", result.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Assign to agent")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Ticket description")
	cmd.Flags().StringSliceVar(&deps, "deps", nil, "Dependency ticket IDs")
	cmd.Flags().IntVarP(&priority, "priority", "p", 2, "Priority 0-4 (0=highest)")
	cmd.Flags().StringVarP(&ticketType, "type", "t", "task", "Type (epic, feature, task, bug, chore)")

	return cmd
}

func squadTicketListCmd() *cobra.Command {
	var (
		status   string
		assignee string
	)

	cmd := &cobra.Command{
		Use:   "list <squad>",
		Short: "List tickets in a squad",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			squadName := args[0]
			ctx := cmd.Context()

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			result, err := client.TicketList(ctx, daemon.TicketListParams{
				SquadName: squadName,
				Status:    status,
				Assignee:  assignee,
			})
			if err != nil {
				return fmt.Errorf("list tickets: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			if len(result.Tickets) == 0 {
				fmt.Println("No tickets")
				return nil
			}

			for _, t := range result.Tickets {
				assigneeStr := ""
				if t.Assignee != "" {
					assigneeStr = t.Assignee + "  "
				}
				fmt.Printf("%s  %s  %s%s\n",
					ticketIDStyle.Render(t.ID),
					statusStyle(t.Status).Render(t.Status),
					assigneeStr,
					t.Title)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Filter by assignee")

	return cmd
}

func squadTicketStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <squad> <ticket-id>",
		Short: "Start working on a ticket",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			squadName := args[0]
			ticketID := args[1]
			ctx := cmd.Context()

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.TicketStartSquad(ctx, squadName, ticketID); err != nil {
				return fmt.Errorf("start ticket: %w", err)
			}

			if !globalOutput.JSON {
				fmt.Printf("▶ Started: %s\n", ticketID)
			}
			return nil
		},
	}
	return cmd
}

func squadTicketCloseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <squad> <ticket-id>",
		Short: "Close a ticket",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			squadName := args[0]
			ticketID := args[1]
			ctx := cmd.Context()

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.TicketCloseSquad(ctx, squadName, ticketID); err != nil {
				return fmt.Errorf("close ticket: %w", err)
			}

			if !globalOutput.JSON {
				fmt.Printf("✓ Closed: %s\n", ticketID)
			}
			return nil
		},
	}
	return cmd
}

