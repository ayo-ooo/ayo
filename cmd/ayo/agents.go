package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/skills"
)

func newAgentsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agents",
		Short:   "Manage agents",
		Aliases: []string{"agent"},
		Long: `Manage AI agents with custom prompts and tool access.

Agents are stored as directories containing:
  config.json      Agent configuration
  system.md        System prompt
  input.jsonschema Optional input validation (for chaining)
  output.jsonschema Optional output format (for chaining)
  skills/          Optional agent-specific skills

Locations:
  User agents: ~/.config/ayo/agents/
  Built-in:    ~/.local/share/ayo/agents/

For help designing agents, chat with @ayo:
  ayo "help me create an agent for code review"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to list
			return listAgentsCmd(cfgPath).RunE(cmd, args)
		},
	}

	cmd.AddCommand(listAgentsCmd(cfgPath))
	cmd.AddCommand(createAgentCmd(cfgPath))
	cmd.AddCommand(showAgentCmd(cfgPath))
	cmd.AddCommand(updateAgentsCmd(cfgPath))

	return cmd
}

func listAgentsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				// Ensure builtins are installed
				if err := builtin.Install(); err != nil {
					return fmt.Errorf("install builtins: %w", err)
				}

				handles, err := agent.ListHandles(cfg)
				if err != nil {
					return err
				}

				// Color palette
				purple := lipgloss.Color("#a78bfa")
				cyan := lipgloss.Color("#67e8f9")
				muted := lipgloss.Color("#6b7280")
				text := lipgloss.Color("#e5e7eb")
				subtle := lipgloss.Color("#374151")

				// Styles
				headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
				sectionStyle := lipgloss.NewStyle().Foreground(muted).Bold(true)
				iconStyle := lipgloss.NewStyle().Foreground(cyan)
				handleStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
				descStyle := lipgloss.NewStyle().Foreground(text)
				countStyle := lipgloss.NewStyle().Foreground(muted)
				dividerStyle := lipgloss.NewStyle().Foreground(subtle)
				emptyStyle := lipgloss.NewStyle().Foreground(muted).Italic(true)

				// Categorize agents
				type agentInfo struct {
					handle string
					desc   string
				}
				var userAgents, builtinAgents []agentInfo

				for _, h := range handles {
					// Get description
					ag, err := agent.Load(cfg, h)
					desc := ""
					if err == nil {
						desc = ag.Config.Description
					}

					// Determine source
					isBuiltin := builtin.HasAgent(h)
					if isBuiltin {
						// Check if user has overridden
						userDir := filepath.Join(cfg.AgentsDir, h)
						if _, err := os.Stat(userDir); err == nil {
							userAgents = append(userAgents, agentInfo{h, desc})
						} else {
							builtinAgents = append(builtinAgents, agentInfo{h, desc})
						}
					} else {
						userAgents = append(userAgents, agentInfo{h, desc})
					}
				}

				// Render function for an agent
				renderAgent := func(a agentInfo) {
					icon := iconStyle.Render("◆")
					handle := handleStyle.Render(a.handle)
					fmt.Printf("  %s %s\n", icon, handle)

					// Description (truncated, indented)
					if a.desc != "" {
						desc := a.desc
						if len(desc) > 52 {
							desc = desc[:49] + "..."
						}
						fmt.Printf("    %s\n", descStyle.Render(desc))
					}
				}

				// Header
				fmt.Println()
				fmt.Println(headerStyle.Render("  Agents"))
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))

				// User-defined agents section
				fmt.Println()
				fmt.Printf("  %s\n", sectionStyle.Render("User-defined"))
				if len(userAgents) == 0 {
					fmt.Printf("    %s\n", emptyStyle.Render("No user-defined agents"))
					fmt.Printf("    %s\n", emptyStyle.Render("Create one with: ayo agents create @name"))
				} else {
					for _, a := range userAgents {
						renderAgent(a)
					}
				}

				// Built-in agents section
				fmt.Println()
				fmt.Printf("  %s\n", sectionStyle.Render("Built-in"))
				if len(builtinAgents) == 0 {
					fmt.Printf("    %s\n", emptyStyle.Render("No built-in agents installed"))
					fmt.Printf("    %s\n", emptyStyle.Render("Run: ayo setup"))
				} else {
					for _, a := range builtinAgents {
						renderAgent(a)
					}
				}

				fmt.Println()
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))
				fmt.Println(countStyle.Render(fmt.Sprintf("  %d agents", len(handles))))
				fmt.Println()

				return nil
			})
		},
	}

	return cmd
}

func createAgentCmd(cfgPath *string) *cobra.Command {
	var (
		// Core
		model       string
		description string
		system      string
		systemFile  string

		// Tools
		tools []string

		// Skills
		skills_             []string
		excludeSkills       []string
		ignoreBuiltinSkills bool
		ignoreSharedSkills  bool

		// Chaining
		inputSchema  string
		outputSchema string

		// Guardrails
		noGuardrails bool
	)

	cmd := &cobra.Command{
		Use:   "create @handle",
		Short: "Create a new agent",
		Long: `Create a new agent with the specified configuration.

REQUIRED:
  @handle              Agent handle (e.g., @myagent)
  --model, -m          Model to use (or configure default_model in ayo.json)

OPTIONAL:
  --description, -d    Brief description of the agent
  --system, -s         System prompt text (inline)
  --system-file, -f    Path to system prompt file (.md or .txt)
  --tools, -t          Allowed tools (default: bash)
  --skills             Skills to include
  --input-schema       JSON schema for structured input
  --output-schema      JSON schema for structured output

For help designing agents, chat with @ayo:
  ayo "help me create an agent for code review"

Examples:
  # Minimal agent
  ayo agents create @helper -m gpt-4.1

  # With description and custom system prompt
  ayo agents create @reviewer \
    -m gpt-4.1 \
    -d "Reviews code for best practices" \
    -f ~/prompts/reviewer.md

  # With tools and skills
  ayo agents create @debugger \
    -m gpt-4.1 \
    -t bash,agent_call \
    --skills debugging

  # Chainable agent with schemas
  ayo agents create @analyzer \
    -m gpt-4.1 \
    -f system.md \
    --input-schema input.jsonschema \
    --output-schema output.jsonschema`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// No handle provided - show help
				return cmd.Help()
			}
			handle := agent.NormalizeHandle(args[0])

			return withConfig(cfgPath, func(cfg config.Config) error {
				providerModels := config.ConfiguredModels(cfg)
				modelSet := make(map[string]struct{}, len(providerModels))
				for _, m := range providerModels {
					modelSet[m.ID] = struct{}{}
				}

				// Check reserved namespace
				if agent.IsReservedNamespace(handle) {
					return fmt.Errorf("cannot use reserved handle %s", handle)
				}

				// Check if already exists
				agentDir := filepath.Join(cfg.AgentsDir, handle)
				if _, err := os.Stat(agentDir); err == nil {
					return fmt.Errorf("agent already exists: %s", handle)
				}

				// Load system from file if specified
				if system == "" && systemFile != "" {
					expanded := expandPath(systemFile)
					data, err := os.ReadFile(expanded)
					if err != nil {
						return fmt.Errorf("read system file: %w", err)
					}
					system = string(data)
				}

				// Default system message
				if system == "" {
					system = "You are a helpful assistant."
				}

				// Default model
				if model == "" {
					model = cfg.DefaultModel
				}

				// Validate model is set
				if model == "" {
					return fmt.Errorf("model is required (use -m or configure default_model in ayo.json)")
				}

				// Validate model if we have a configured set
				if len(modelSet) > 0 {
					if _, ok := modelSet[model]; !ok {
						return fmt.Errorf("model %s is not configured", model)
					}
				}

				// Default tools
				if len(tools) == 0 {
					tools = []string{"bash"}
				}

				// Merge required skills based on selected tools
				requiredSkills := skills.GetRequiredSkillsForTools(tools)
				if len(requiredSkills) > 0 {
					skillSet := make(map[string]struct{}, len(skills_))
					for _, s := range skills_ {
						skillSet[s] = struct{}{}
					}
					for _, s := range requiredSkills {
						if _, exists := skillSet[s]; !exists {
							skills_ = append(skills_, s)
						}
					}
				}

				agCfg := agent.Config{
					Model:               model,
					Description:         description,
					AllowedTools:        tools,
					Guardrails:          boolPtr(!noGuardrails),
					Skills:              skills_,
					ExcludeSkills:       excludeSkills,
					IgnoreBuiltinSkills: ignoreBuiltinSkills,
					IgnoreSharedSkills:  ignoreSharedSkills,
				}

				ag, err := agent.SaveWithSchemas(cfg, handle, agCfg, system, inputSchema, outputSchema)
				if err != nil {
					return err
				}

				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("Created agent: " + ag.Handle))
				fmt.Printf("  Location: %s\n", ag.Dir)

				// Show what was configured (from config, not resolved)
				if len(ag.Config.AllowedTools) > 0 {
					fmt.Printf("  Tools: %s\n", strings.Join(ag.Config.AllowedTools, ", "))
				}
				if len(ag.Config.Skills) > 0 {
					fmt.Printf("  Skills: %s\n", strings.Join(ag.Config.Skills, ", "))
				}
				if ag.HasInputSchema() || ag.HasOutputSchema() {
					fmt.Println("  Chaining: enabled")
				}

				return nil
			})
		},
	}

	// Core flags
	cmd.Flags().StringVarP(&model, "model", "m", "", "model to use (see 'ayo config show' for available models)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "brief description of what this agent does")
	cmd.Flags().StringVarP(&system, "system", "s", "", "system prompt text (mutually exclusive with --system-file)")
	cmd.Flags().StringVarP(&systemFile, "system-file", "f", "", "path to system prompt file (.md or .txt)")

	// Tool flags
	cmd.Flags().StringSliceVarP(&tools, "tools", "t", nil, "allowed tools: bash, agent_call, todo (comma-separated)")

	// Skill flags
	cmd.Flags().StringSliceVar(&skills_, "skills", nil, "skills to include (comma-separated, see 'ayo skills list')")
	cmd.Flags().StringSliceVar(&excludeSkills, "exclude-skills", nil, "skills to exclude from this agent")
	cmd.Flags().BoolVar(&ignoreBuiltinSkills, "ignore-builtin-skills", false, "don't load built-in skills")
	cmd.Flags().BoolVar(&ignoreSharedSkills, "ignore-shared-skills", false, "don't load user shared skills")

	// Schema flags for chaining
	cmd.Flags().StringVar(&inputSchema, "input-schema", "", "JSON schema file for validating stdin input")
	cmd.Flags().StringVar(&outputSchema, "output-schema", "", "JSON schema file for structuring stdout output")

	// Guardrails
	cmd.Flags().BoolVar(&noGuardrails, "no-guardrails", false, "disable safety guardrails (dangerous - use with caution)")

	return cmd
}

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool {
	return &b
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return strings.Replace(path, "~", home, 1)
	}
	return path
}

func showAgentCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <handle>",
		Short: "Show agent details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Ensure builtins are installed
				if err := builtin.Install(); err != nil {
					return fmt.Errorf("install builtins: %w", err)
				}

				ag, err := agent.Load(cfg, handle)
				if err != nil {
					return fmt.Errorf("agent not found: %s", handle)
				}

				// Color palette
				purple := lipgloss.Color("#a78bfa")
				cyan := lipgloss.Color("#67e8f9")
				muted := lipgloss.Color("#6b7280")
				text := lipgloss.Color("#e5e7eb")
				subtle := lipgloss.Color("#374151")

				// Styles
				headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
				iconStyle := lipgloss.NewStyle().Foreground(cyan)
				labelStyle := lipgloss.NewStyle().Foreground(muted)
				valueStyle := lipgloss.NewStyle().Foreground(text)
				dividerStyle := lipgloss.NewStyle().Foreground(subtle)

				fmt.Println()
				fmt.Println("  " + iconStyle.Render("◆") + " " + headerStyle.Render(ag.Handle))
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))

				source := "user"
				if ag.BuiltIn {
					source = "built-in"
				}
				fmt.Printf("  %s %s\n", labelStyle.Render("Source:"), valueStyle.Render(source))
				fmt.Printf("  %s  %s\n", labelStyle.Render("Model:"), valueStyle.Render(ag.Model))

				if ag.Config.Description != "" {
					fmt.Printf("  %s   %s\n", labelStyle.Render("Desc:"), valueStyle.Render(ag.Config.Description))
				}

				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))
				fmt.Printf("  %s   %s\n", labelStyle.Render("Path:"), valueStyle.Render(ag.Dir))

				if len(ag.Skills) > 0 {
					skillNames := make([]string, len(ag.Skills))
					for i, s := range ag.Skills {
						skillNames[i] = s.Name
					}
					sort.Strings(skillNames)
					fmt.Printf("  %s %s\n", labelStyle.Render("Skills:"), valueStyle.Render(strings.Join(skillNames, ", ")))
				}

				if len(ag.Config.AllowedTools) > 0 {
					fmt.Printf("  %s  %s\n", labelStyle.Render("Tools:"), valueStyle.Render(strings.Join(ag.Config.AllowedTools, ", ")))
				}

				fmt.Println()

				return nil
			})
		},
	}

	return cmd
}

func updateAgentsCmd(cfgPath *string) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update built-in agents to latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				sui := newSetupUI(cmd.OutOrStdout())

				if !force {
					// Check for modified agents
					modified, err := builtin.CheckModifiedAgents()
					if err != nil {
						return fmt.Errorf("check modified agents: %w", err)
					}

					if len(modified) > 0 {
						sui.Warning("The following agents have local modifications:")
						for _, m := range modified {
							sui.Info(fmt.Sprintf("  %s: %v", m.Handle, m.ModifiedFiles))
						}
						sui.Blank()
						sui.Info("Use --force to overwrite, or copy modifications to user directory first:")
						sui.Info(fmt.Sprintf("  %s", cfg.AgentsDir))
						return fmt.Errorf("agents have local modifications")
					}
				}

				sui.Step("Updating built-in agents...")
				installDir, err := builtin.ForceInstall()
				if err != nil {
					return err
				}
				sui.SuccessPath("Updated agents at", installDir)
				return nil
			})
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "overwrite without checking for modifications")

	return cmd
}
