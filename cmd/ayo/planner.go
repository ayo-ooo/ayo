package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/planners"

	// Import planner plugins to register them
	_ "github.com/alexcabrera/ayo/internal/planners/builtin/tickets"
	_ "github.com/alexcabrera/ayo/internal/planners/builtin/todos"
)

func newPlannerCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "planner",
		Short:   "Manage planner plugins",
		Aliases: []string{"planners"},
		Long: `Manage planner plugins for work coordination.

Planners provide work tracking tools to agents:
  - Near-term planners (default: ayo-todos) handle session-scoped tasks
  - Long-term planners (default: ayo-tickets) handle persistent work tracking

Each sandbox can have its own planner configuration via SQUAD.md frontmatter.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to list
			return listPlannersCmd(cfgPath).RunE(cmd, args)
		},
	}

	cmd.AddCommand(listPlannersCmd(cfgPath))
	cmd.AddCommand(showPlannerCmd(cfgPath))
	cmd.AddCommand(setPlannerCmd(cfgPath))

	return cmd
}

func listPlannersCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available planner plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				names := planners.DefaultRegistry.List()

				if len(names) == 0 {
					fmt.Println("No planners registered.")
					return nil
				}

				// Get configured defaults
				defaults := cfg.Planners.WithDefaults()

				// Styles
				nameStyle := lipgloss.NewStyle().Bold(true)
				typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
				activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

				fmt.Println(lipgloss.NewStyle().Bold(true).Render("Available Planners"))
				fmt.Println()

				for _, name := range names {
					// Create a temporary instance to get the type
					factory, _ := planners.DefaultRegistry.Get(name)
					tempDir, _ := os.MkdirTemp("", "planner-info")
					defer os.RemoveAll(tempDir)

					ctx := planners.PlannerContext{
						SandboxName: "temp",
						SandboxDir:  tempDir,
						StateDir:    tempDir,
					}
					plugin, err := factory(ctx)
					if err != nil {
						continue
					}
					defer plugin.Close()

					plannerType := plugin.Type()
					typeLabel := fmt.Sprintf("[%s]", plannerType)

					// Check if this is the active planner
					isActive := false
					var activeLabel string
					if plannerType == planners.NearTerm && defaults.NearTerm == name {
						isActive = true
						activeLabel = " (default near-term)"
					} else if plannerType == planners.LongTerm && defaults.LongTerm == name {
						isActive = true
						activeLabel = " (default long-term)"
					}

					line := fmt.Sprintf("  %s %s",
						nameStyle.Render(name),
						typeStyle.Render(typeLabel),
					)
					if isActive {
						line += activeStyle.Render(activeLabel)
					}
					fmt.Println(line)
				}

				return nil
			})
		},
	}
}

func showPlannerCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show details about a planner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			return withConfig(cfgPath, func(cfg config.Config) error {
				if !planners.DefaultRegistry.Has(name) {
					return fmt.Errorf("planner not found: %s", name)
				}

				factory, _ := planners.DefaultRegistry.Get(name)
				tempDir, err := os.MkdirTemp("", "planner-info")
				if err != nil {
					return err
				}
				defer os.RemoveAll(tempDir)

				ctx := planners.PlannerContext{
					SandboxName: "temp",
					SandboxDir:  tempDir,
					StateDir:    tempDir,
				}
				plugin, err := factory(ctx)
				if err != nil {
					return fmt.Errorf("failed to instantiate planner: %w", err)
				}
				defer plugin.Close()

				defaults := cfg.Planners.WithDefaults()

				// Header
				headerStyle := lipgloss.NewStyle().Bold(true)
				labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

				fmt.Println(headerStyle.Render(name))
				fmt.Println()

				// Type
				fmt.Printf("%s %s\n", labelStyle.Render("Type:"), plugin.Type())

				// Default status
				isNearDefault := defaults.NearTerm == name
				isLongDefault := defaults.LongTerm == name
				if isNearDefault || isLongDefault {
					status := "yes"
					if isNearDefault {
						status += " (near-term)"
					}
					if isLongDefault {
						status += " (long-term)"
					}
					fmt.Printf("%s %s\n", labelStyle.Render("Default:"), status)
				} else {
					fmt.Printf("%s no\n", labelStyle.Render("Default:"))
				}

				// Tools
				tools := plugin.Tools()
				if len(tools) > 0 {
					var toolNames []string
					for _, t := range tools {
						toolNames = append(toolNames, t.Info().Name)
					}
					fmt.Printf("%s %s\n", labelStyle.Render("Tools:"), strings.Join(toolNames, ", "))
				} else {
					fmt.Printf("%s none\n", labelStyle.Render("Tools:"))
				}

				// Instructions - show full content
				instructions := plugin.Instructions()
				if instructions != "" {
					fmt.Printf("%s\n", labelStyle.Render("Instructions:"))
					for _, line := range strings.Split(instructions, "\n") {
						fmt.Printf("  %s\n", line)
					}
				}

				return nil
			})
		},
	}
}

func setPlannerCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <near|long> <name>",
		Short: "Set the default planner for a slot",
		Long: `Set the default near-term or long-term planner.

Examples:
  ayo planner set near ayo-todos
  ayo planner set long ayo-tickets`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			slot := strings.ToLower(args[0])
			name := args[1]

			if slot != "near" && slot != "long" {
				return fmt.Errorf("slot must be 'near' or 'long', got %q", slot)
			}

			if !planners.DefaultRegistry.Has(name) {
				return fmt.Errorf("planner not found: %s", name)
			}

			// Verify the planner type matches the slot
			factory, _ := planners.DefaultRegistry.Get(name)
			tempDir, err := os.MkdirTemp("", "planner-verify")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tempDir)

			ctx := planners.PlannerContext{
				SandboxName: "temp",
				SandboxDir:  tempDir,
				StateDir:    tempDir,
			}
			plugin, err := factory(ctx)
			if err != nil {
				return fmt.Errorf("failed to instantiate planner: %w", err)
			}
			defer plugin.Close()

			plannerType := plugin.Type()
			if slot == "near" && plannerType != planners.NearTerm {
				return fmt.Errorf("planner %s is type %q, expected %q for near-term slot", name, plannerType, planners.NearTerm)
			}
			if slot == "long" && plannerType != planners.LongTerm {
				return fmt.Errorf("planner %s is type %q, expected %q for long-term slot", name, plannerType, planners.LongTerm)
			}

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Update the config
				if slot == "near" {
					cfg.Planners.NearTerm = name
				} else {
					cfg.Planners.LongTerm = name
				}

				// Save config
				configPath := config.DefaultPath()
				if cfgPath != nil && *cfgPath != "" {
					configPath = *cfgPath
				}
				if err := config.Save(configPath, cfg); err != nil {
					return fmt.Errorf("save config: %w", err)
				}

				fmt.Printf("Set default %s-term planner to %s\n", slot, name)
				return nil
			})
		},
	}

	return cmd
}
