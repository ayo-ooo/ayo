package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fantasy/schema"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/daemon"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/plugins"
	"github.com/alexcabrera/ayo/internal/squads"
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
		Long: `Manage squad sandboxes for multi-agent coordination.

Squads are isolated environments where agents collaborate on tasks.
Each squad has its own workspace, tickets, context, and SQUAD.md constitution.

Commands:
  list        List all squads
  create      Create a new squad
  show        Show squad details
  destroy     Delete a squad and its data
  start       Start a squad's sandbox
  stop        Stop a squad's sandbox
  add-agent   Add an agent to a squad
  remove-agent Remove an agent from a squad
  ticket      Manage squad tickets

Examples:
  ayo squad list                         List all squads
  ayo squad create frontend              Create frontend squad
  ayo squad add-agent frontend @react    Add agent to squad
  ayo squad show frontend                Show squad details
  ayo #frontend "build auth"             Send task to squad`,
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
	cmd.AddCommand(squadSchemaCmd())
	cmd.AddCommand(squadShellCmd())

	return cmd
}

func squadListCmd() *cobra.Command {
	var showPlugins bool

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

			// Get plugin squads if requested
			var pluginSquads []plugins.SquadInfo
			if showPlugins {
				pluginSquads = plugins.ListAllSquadInfo()
			}

			if globalOutput.JSON {
				output := map[string]any{
					"squads": result.Squads,
				}
				if showPlugins {
					output["plugin_squads"] = pluginSquads
				}
				return json.NewEncoder(os.Stdout).Encode(output)
			}

			if len(result.Squads) == 0 && len(pluginSquads) == 0 {
				fmt.Println(squadMutedStyle.Render("No squads configured"))
				return nil
			}

			// Header
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#a78bfa"))
			
			if len(result.Squads) > 0 {
				fmt.Println()
				fmt.Println(headerStyle.Render("  Squads"))
				fmt.Println(squadMutedStyle.Render("  " + strings.Repeat("─", 60)))
				fmt.Println()

				for _, squad := range result.Squads {
					status := squadStatusColor(squad.Status).Render(fmt.Sprintf("[%s]", squad.Status))
					name := squadNameStyle.Render(squad.Name)
					
					fmt.Printf("  %s %s\n", name, status)
					if squad.Description != "" {
						fmt.Printf("    %s\n", squadMutedStyle.Render(squad.Description))
					}
					if len(squad.Agents) > 0 {
						fmt.Printf("    %s %s\n", squadMutedStyle.Render("Agents:"), strings.Join(squad.Agents, ", "))
					}
					fmt.Println()
				}
			}

			// Show plugin squads if requested
			if showPlugins && len(pluginSquads) > 0 {
				fmt.Println()
				fmt.Println(headerStyle.Render("  Plugin Squads"))
				fmt.Println(squadMutedStyle.Render("  " + strings.Repeat("─", 60)))
				fmt.Println()

				pluginStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))
				for _, ps := range pluginSquads {
					name := squadNameStyle.Render(ps.Name)
					plugin := pluginStyle.Render(fmt.Sprintf("(%s)", ps.PluginName))
					
					fmt.Printf("  %s %s\n", name, plugin)
					if ps.Description != "" {
						fmt.Printf("    %s\n", squadMutedStyle.Render(ps.Description))
					}
					if len(ps.Agents) > 0 {
						fmt.Printf("    %s %s\n", squadMutedStyle.Render("Agents:"), strings.Join(ps.Agents, ", "))
					}
					fmt.Println()
				}

				fmt.Println(squadMutedStyle.Render("  Use 'ayo squad create <name> --from-plugin <squad>' to create from template"))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showPlugins, "plugins", false, "Also show available plugin squads")

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
		fromPlugin     string
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
				FromPlugin:     fromPlugin,
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
	cmd.Flags().StringVar(&fromPlugin, "from-plugin", "", "Create from a plugin squad template")

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

func squadSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Manage squad I/O schemas",
		Long: `Manage JSON schemas for squad input/output validation.

Squads can optionally define input.jsonschema and output.jsonschema files
to enable structured I/O with validation.

Examples:
  # Initialize template schemas
  ayo squad schema init dev-team

  # Show current schemas
  ayo squad schema show dev-team

  # Validate schemas are syntactically correct
  ayo squad schema validate dev-team

  # Validate input data against schema
  ayo squad schema validate-input dev-team input.json`,
	}

	cmd.AddCommand(squadSchemaInitCmd())
	cmd.AddCommand(squadSchemaShowCmd())
	cmd.AddCommand(squadSchemaValidateCmd())
	cmd.AddCommand(squadSchemaValidateInputCmd())

	return cmd
}

func squadSchemaInitCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init <squad>",
		Short: "Create template I/O schemas",
		Long:  "Creates template input.jsonschema and output.jsonschema files in the squad directory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			squadName := args[0]
			squadDir := paths.SquadDir(squadName)

			// Check squad exists
			if _, err := os.Stat(squadDir); os.IsNotExist(err) {
				return fmt.Errorf("squad %q does not exist", squadName)
			}

			// Template schemas
			inputSchema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "description": "Input schema for squad %s",
  "properties": {
    "task": {
      "type": "string",
      "description": "Description of the task to perform"
    },
    "requirements": {
      "type": "array",
      "items": { "type": "string" },
      "description": "List of requirements"
    }
  },
  "required": ["task"]
}
`
			outputSchema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "description": "Output schema for squad %s",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["success", "failure", "partial"],
      "description": "Overall result status"
    },
    "summary": {
      "type": "string",
      "description": "Summary of work completed"
    },
    "files_changed": {
      "type": "array",
      "items": { "type": "string" },
      "description": "List of files modified"
    }
  },
  "required": ["status", "summary"]
}
`
			inputPath := filepath.Join(squadDir, "input.jsonschema")
			outputPath := filepath.Join(squadDir, "output.jsonschema")

			// Check for existing files
			if !force {
				if _, err := os.Stat(inputPath); err == nil {
					return fmt.Errorf("input.jsonschema already exists (use --force to overwrite)")
				}
				if _, err := os.Stat(outputPath); err == nil {
					return fmt.Errorf("output.jsonschema already exists (use --force to overwrite)")
				}
			}

			// Write input schema
			if err := os.WriteFile(inputPath, []byte(fmt.Sprintf(inputSchema, squadName)), 0644); err != nil {
				return fmt.Errorf("write input.jsonschema: %w", err)
			}

			// Write output schema
			if err := os.WriteFile(outputPath, []byte(fmt.Sprintf(outputSchema, squadName)), 0644); err != nil {
				return fmt.Errorf("write output.jsonschema: %w", err)
			}

			if !globalOutput.JSON {
				fmt.Printf("✓ Created %s\n", inputPath)
				fmt.Printf("✓ Created %s\n", outputPath)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing schema files")
	return cmd
}

func squadSchemaShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <squad>",
		Short: "Display squad schemas",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			squadName := args[0]
			squadDir := paths.SquadDir(squadName)

			// Check squad exists
			if _, err := os.Stat(squadDir); os.IsNotExist(err) {
				return fmt.Errorf("squad %q does not exist", squadName)
			}

			schemas, err := squads.LoadSquadSchemas(squadName)
			if err != nil {
				return fmt.Errorf("load schemas: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"squad":  squadName,
					"input":  schemas.Input,
					"output": schemas.Output,
				})
			}

			// Human-readable output
			fmt.Printf("Squad: %s\n\n", squadNameStyle.Render(squadName))

			if schemas.HasInputSchema() {
				fmt.Println("Input Schema (input.jsonschema):")
				printSchemaDetails(schemas.Input)
			} else {
				fmt.Println("Input Schema: " + squadMutedStyle.Render("not defined (free-form)"))
			}

			fmt.Println()

			if schemas.HasOutputSchema() {
				fmt.Println("Output Schema (output.jsonschema):")
				printSchemaDetails(schemas.Output)
			} else {
				fmt.Println("Output Schema: " + squadMutedStyle.Render("not defined (free-form)"))
			}

			return nil
		},
	}
	return cmd
}

func squadSchemaValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <squad>",
		Short: "Validate squad schemas",
		Long:  "Check that squad schema files are valid JSON Schema.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			squadName := args[0]
			squadDir := paths.SquadDir(squadName)

			// Check squad exists
			if _, err := os.Stat(squadDir); os.IsNotExist(err) {
				return fmt.Errorf("squad %q does not exist", squadName)
			}

			inputPath := filepath.Join(squadDir, "input.jsonschema")
			outputPath := filepath.Join(squadDir, "output.jsonschema")

			var errors []string
			var valid []string

			// Validate input schema
			if _, err := os.Stat(inputPath); err == nil {
				data, err := os.ReadFile(inputPath)
				if err != nil {
					errors = append(errors, fmt.Sprintf("input.jsonschema: %v", err))
				} else {
					var s any
					if err := json.Unmarshal(data, &s); err != nil {
						errors = append(errors, fmt.Sprintf("input.jsonschema: invalid JSON: %v", err))
					} else {
						valid = append(valid, "input.jsonschema")
					}
				}
			}

			// Validate output schema
			if _, err := os.Stat(outputPath); err == nil {
				data, err := os.ReadFile(outputPath)
				if err != nil {
					errors = append(errors, fmt.Sprintf("output.jsonschema: %v", err))
				} else {
					var s any
					if err := json.Unmarshal(data, &s); err != nil {
						errors = append(errors, fmt.Sprintf("output.jsonschema: invalid JSON: %v", err))
					} else {
						valid = append(valid, "output.jsonschema")
					}
				}
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"squad":  squadName,
					"valid":  valid,
					"errors": errors,
				})
			}

			if len(valid) > 0 {
				for _, v := range valid {
					fmt.Printf("✓ %s is valid\n", v)
				}
			}
			if len(errors) > 0 {
				for _, e := range errors {
					fmt.Printf("✗ %s\n", e)
				}
				return fmt.Errorf("validation failed")
			}

			if len(valid) == 0 && len(errors) == 0 {
				fmt.Println(squadMutedStyle.Render("No schema files found"))
			}

			return nil
		},
	}
	return cmd
}

func printSchemaDetails(s *schema.Schema) {
	fmt.Printf("  Type: %s\n", s.Type)
	if s.Description != "" {
		fmt.Printf("  Description: %s\n", s.Description)
	}
	if len(s.Properties) > 0 {
		fmt.Println("  Properties:")
		for name, prop := range s.Properties {
			required := ""
			for _, r := range s.Required {
				if r == name {
					required = " (required)"
					break
				}
			}
			desc := ""
			if prop.Description != "" {
				desc = " - " + prop.Description
			}
			fmt.Printf("    %s: %s%s%s\n", name, prop.Type, required, desc)
		}
	}
}

func squadSchemaValidateInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-input <squad> <input.json>",
		Short: "Validate input data against squad schema",
		Long:  "Validates a JSON input file against the squad's input.jsonschema.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			squadName := args[0]
			inputPath := args[1]

			// Load squad schemas
			schemas, err := squads.LoadSquadSchemas(squadName)
			if err != nil {
				return fmt.Errorf("load schemas: %w", err)
			}
			if schemas == nil || schemas.Input == nil {
				return fmt.Errorf("squad %q has no input schema defined", squadName)
			}

			// Load and parse input file
			data, err := os.ReadFile(inputPath)
			if err != nil {
				return fmt.Errorf("read input file: %w", err)
			}
			var input any
			if err := json.Unmarshal(data, &input); err != nil {
				return fmt.Errorf("parse input JSON: %w", err)
			}

			// Validate against schema
			if err := schema.ValidateAgainstSchema(input, *schemas.Input); err != nil {
				if globalOutput.JSON {
					return json.NewEncoder(os.Stdout).Encode(map[string]any{
						"squad":  squadName,
						"input":  inputPath,
						"valid":  false,
						"errors": []string{err.Error()},
					})
				}
				return fmt.Errorf("validation failed: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"squad":  squadName,
					"input":  inputPath,
					"valid":  true,
					"errors": []string{},
				})
			}

			fmt.Printf("✓ %s is valid against %s input schema\n", inputPath, squadName)
			return nil
		},
	}
	return cmd
}

