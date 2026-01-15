package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/pipe"
)

func newChainCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain",
		Short: "Explore and validate agent chaining",
		Long:  "Commands for discovering compatible agents and validating chain connections.",
	}

	cmd.AddCommand(newChainLsCmd(cfgPath))
	cmd.AddCommand(newChainInspectCmd(cfgPath))
	cmd.AddCommand(newChainFromCmd(cfgPath))
	cmd.AddCommand(newChainToCmd(cfgPath))
	cmd.AddCommand(newChainValidateCmd(cfgPath))
	cmd.AddCommand(newChainExampleCmd(cfgPath))

	return cmd
}

// newChainLsCmd lists all chainable agents.
func newChainLsCmd(cfgPath *string) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List all chainable agents",
		Long:  "Lists all agents that have input or output schemas defined.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				agents, err := agent.ListChainableAgents(cfg)
				if err != nil {
					return err
				}

				if jsonOutput || pipe.IsStdoutPiped() {
					return outputJSON(agents)
				}

				return displayChainableAgents(agents)
			})
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	return cmd
}

// newChainInspectCmd shows an agent's input and output schemas.
func newChainInspectCmd(cfgPath *string) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "inspect <agent>",
		Short: "Show agent's input and output schemas",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				handle := agent.NormalizeHandle(args[0])
				ag, err := agent.Load(cfg, handle)
				if err != nil {
					return err
				}

				if jsonOutput || pipe.IsStdoutPiped() {
					return outputSchemaJSON(ag)
				}

				return displayAgentSchemas(ag)
			})
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	return cmd
}

// newChainFromCmd lists agents that can receive output from the given agent.
func newChainFromCmd(cfgPath *string) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "from <agent>",
		Short: "List agents that can receive output from this agent",
		Long:  "Shows all agents whose input schema is compatible with this agent's output schema.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				handle := agent.NormalizeHandle(args[0])
				source, err := agent.Load(cfg, handle)
				if err != nil {
					return err
				}

				if source.OutputSchema == nil {
					return fmt.Errorf("%s has no output schema and cannot chain to other agents", handle)
				}

				compatible, err := agent.FindDownstreamAgents(cfg, source)
				if err != nil {
					return err
				}

				if jsonOutput || pipe.IsStdoutPiped() {
					return outputCompatibleJSON(compatible)
				}

				return displayCompatibleAgents("Agents that can receive output from "+handle, compatible)
			})
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	return cmd
}

// newChainToCmd lists agents whose output this agent can receive.
func newChainToCmd(cfgPath *string) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "to <agent>",
		Short: "List agents whose output this agent can receive",
		Long:  "Shows all agents whose output schema is compatible with this agent's input schema.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				handle := agent.NormalizeHandle(args[0])
				target, err := agent.Load(cfg, handle)
				if err != nil {
					return err
				}

				compatible, err := agent.FindUpstreamAgents(cfg, target)
				if err != nil {
					return err
				}

				if jsonOutput || pipe.IsStdoutPiped() {
					return outputCompatibleJSON(compatible)
				}

				if target.InputSchema == nil {
					fmt.Printf("%s accepts freeform input and can receive output from any agent with an output schema.\n\n", handle)
				}

				return displayCompatibleAgents("Agents that can send output to "+handle, compatible)
			})
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	return cmd
}

// newChainValidateCmd validates JSON against an agent's input schema.
func newChainValidateCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <agent> [json]",
		Short: "Validate JSON against an agent's input schema",
		Long:  "Validates JSON input against the specified agent's input schema. JSON can be provided as an argument or piped via stdin.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				handle := agent.NormalizeHandle(args[0])
				ag, err := agent.Load(cfg, handle)
				if err != nil {
					return err
				}

				if ag.InputSchema == nil {
					fmt.Printf("%s has no input schema - it accepts any input.\n", handle)
					return nil
				}

				// Get JSON from args or stdin
				var jsonInput string
				if len(args) > 1 {
					jsonInput = args[1]
				} else if pipe.IsStdinPiped() {
					data, err := pipe.ReadStdin()
					if err != nil {
						return fmt.Errorf("read stdin: %w", err)
					}
					jsonInput = strings.TrimSpace(data)
				} else {
					return fmt.Errorf("provide JSON as an argument or pipe via stdin")
				}

				if err := ag.ValidateInput(jsonInput); err != nil {
					return fmt.Errorf("validation failed: %w", err)
				}

				fmt.Printf("✓ Valid input for %s\n", handle)
				return nil
			})
		},
	}

	return cmd
}

// newChainExampleCmd generates example input JSON for an agent.
func newChainExampleCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "example <agent>",
		Short: "Generate example input JSON for an agent",
		Long:  "Generates example JSON that matches the agent's input schema.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				handle := agent.NormalizeHandle(args[0])
				ag, err := agent.Load(cfg, handle)
				if err != nil {
					return err
				}

				if ag.InputSchema == nil {
					return fmt.Errorf("%s has no input schema", handle)
				}

				example := generateSchemaExample(ag.InputSchema)
				jsonBytes, err := json.MarshalIndent(example, "", "  ")
				if err != nil {
					return err
				}

				fmt.Println(string(jsonBytes))
				return nil
			})
		},
	}

	return cmd
}

// Helper functions for display

func displayChainableAgents(agents []agent.Agent) error {
	if len(agents) == 0 {
		fmt.Println("No chainable agents found.")
		return nil
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	handleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	fmt.Println(headerStyle.Render("Chainable Agents"))
	fmt.Println()

	for _, ag := range agents {
		var badges []string
		if ag.InputSchema != nil {
			badges = append(badges, "input")
		}
		if ag.OutputSchema != nil {
			badges = append(badges, "output")
		}

		badgeStr := mutedStyle.Render("[" + strings.Join(badges, ", ") + "]")
		desc := ""
		if ag.Config.Description != "" {
			desc = " - " + ag.Config.Description
		}

		fmt.Printf("  %s %s%s\n", handleStyle.Render(ag.Handle), badgeStr, mutedStyle.Render(desc))
	}

	return nil
}

func displayAgentSchemas(ag agent.Agent) error {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))

	fmt.Println(headerStyle.Render(ag.Handle))
	if ag.Config.Description != "" {
		fmt.Println(ag.Config.Description)
	}
	fmt.Println()

	if ag.InputSchema != nil {
		fmt.Println(labelStyle.Render("Input Schema:"))
		printSchema(ag.InputSchema, "  ")
		fmt.Println()
	} else {
		fmt.Println(labelStyle.Render("Input Schema:"), "none (accepts freeform)")
		fmt.Println()
	}

	if ag.OutputSchema != nil {
		fmt.Println(labelStyle.Render("Output Schema:"))
		printSchema(ag.OutputSchema, "  ")
	} else {
		fmt.Println(labelStyle.Render("Output Schema:"), "none (freeform output)")
	}

	return nil
}

func displayCompatibleAgents(title string, agents []agent.ChainableAgent) error {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	handleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	tierStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	fmt.Println(headerStyle.Render(title))
	fmt.Println()

	if len(agents) == 0 {
		fmt.Println("  No compatible agents found.")
		return nil
	}

	for _, ca := range agents {
		tier := tierStyle.Render("[" + ca.Compatibility.String() + "]")
		desc := ""
		if ca.Agent.Config.Description != "" {
			desc = " - " + ca.Agent.Config.Description
		}
		fmt.Printf("  %s %s%s\n", handleStyle.Render(ca.Agent.Handle), tier, tierStyle.Render(desc))
	}

	return nil
}

func printSchema(s *agent.Schema, indent string) {
	if s == nil {
		return
	}

	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229"))

	fmt.Printf("%stype: %s\n", indent, s.Type)

	if len(s.Required) > 0 {
		fmt.Printf("%srequired: %s\n", indent, strings.Join(s.Required, ", "))
	}

	if len(s.Properties) > 0 {
		fmt.Printf("%sproperties:\n", indent)
		for name, prop := range s.Properties {
			req := ""
			for _, r := range s.Required {
				if r == name {
					req = " *"
					break
				}
			}
			desc := ""
			if prop.Description != "" {
				desc = mutedStyle.Render(" - " + prop.Description)
			}
			fmt.Printf("%s  %s: %s%s%s\n", indent, keyStyle.Render(name), prop.Type, req, desc)
		}
	}
}

// JSON output helpers

type chainableAgentJSON struct {
	Handle      string `json:"handle"`
	Description string `json:"description,omitempty"`
	HasInput    bool   `json:"has_input_schema"`
	HasOutput   bool   `json:"has_output_schema"`
}

func outputJSON(agents []agent.Agent) error {
	result := make([]chainableAgentJSON, len(agents))
	for i, ag := range agents {
		result[i] = chainableAgentJSON{
			Handle:      ag.Handle,
			Description: ag.Config.Description,
			HasInput:    ag.InputSchema != nil,
			HasOutput:   ag.OutputSchema != nil,
		}
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}

type schemaJSON struct {
	Handle      string         `json:"handle"`
	Description string         `json:"description,omitempty"`
	Input       *agent.Schema  `json:"input_schema,omitempty"`
	Output      *agent.Schema  `json:"output_schema,omitempty"`
}

func outputSchemaJSON(ag agent.Agent) error {
	result := schemaJSON{
		Handle:      ag.Handle,
		Description: ag.Config.Description,
		Input:       ag.InputSchema,
		Output:      ag.OutputSchema,
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}

type compatibleAgentJSON struct {
	Handle        string `json:"handle"`
	Description   string `json:"description,omitempty"`
	Compatibility string `json:"compatibility"`
}

func outputCompatibleJSON(agents []agent.ChainableAgent) error {
	result := make([]compatibleAgentJSON, len(agents))
	for i, ca := range agents {
		result[i] = compatibleAgentJSON{
			Handle:        ca.Agent.Handle,
			Description:   ca.Agent.Config.Description,
			Compatibility: ca.Compatibility.String(),
		}
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}

// generateSchemaExample generates example data for a schema.
func generateSchemaExample(s *agent.Schema) any {
	if s == nil {
		return nil
	}

	switch s.Type {
	case "object":
		obj := make(map[string]any)
		for name, prop := range s.Properties {
			obj[name] = generateSchemaExample(prop)
		}
		return obj
	case "array":
		if s.Items != nil {
			return []any{generateSchemaExample(s.Items)}
		}
		return []any{}
	case "string":
		if len(s.Enum) > 0 {
			return s.Enum[0]
		}
		if s.Description != "" {
			return "<" + s.Description + ">"
		}
		return "example"
	case "integer":
		return 0
	case "number":
		return 0.0
	case "boolean":
		return false
	default:
		return nil
	}
}
