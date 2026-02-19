package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/cli"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/daemon"
)

// Ensure cli package is used (globalOutput is defined in root.go)
var _ = cli.Output{}

func newAgentsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agents",
		Short:   "Manage agents",
		Aliases: []string{"agent"},
		Long: `Manage AI agents with custom prompts and tool access.

Agents are stored as directories containing:
  config.json      Agent configuration
  system.md        System prompt
  input.jsonschema Optional input validation (for pipelines)
  output.jsonschema Optional output format (for pipelines)
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
	cmd.AddCommand(rmAgentCmd(cfgPath))
	cmd.AddCommand(statusAgentsCmd())
	cmd.AddCommand(wakeAgentCmd())
	cmd.AddCommand(sleepAgentCmd())
	cmd.AddCommand(capabilitiesAgentsCmd(cfgPath))
	cmd.AddCommand(promoteAgentCmd(cfgPath))
	cmd.AddCommand(archiveAgentCmd(cfgPath))
	cmd.AddCommand(unarchiveAgentCmd(cfgPath))
	cmd.AddCommand(refineAgentCmd(cfgPath))

	return cmd
}

func listAgentsCmd(cfgPath *string) *cobra.Command {
	var (
		trustFilter string
		typeFilter  string
	)

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
				yellow := lipgloss.Color("#fbbf24")
				red := lipgloss.Color("#ef4444")
				green := lipgloss.Color("#34d399")

				// Styles
				headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
				handleStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
				descStyle := lipgloss.NewStyle().Foreground(text)
				countStyle := lipgloss.NewStyle().Foreground(muted)
				dividerStyle := lipgloss.NewStyle().Foreground(subtle)
				emptyStyle := lipgloss.NewStyle().Foreground(muted).Italic(true)
				warnStyle := lipgloss.NewStyle().Foreground(yellow)
				dangerStyle := lipgloss.NewStyle().Foreground(red)
				okStyle := lipgloss.NewStyle().Foreground(green)

				// Collect agent info with metadata
				type agentInfo struct {
					handle    string
					desc      string
					trust     agent.TrustLevel
					agentType string // "builtin", "user"
				}
				var agents []agentInfo

				for _, h := range handles {
					// Get agent config
					ag, err := agent.Load(cfg, h)
					if err != nil {
						continue
					}

					info := agentInfo{
						handle:    h,
						desc:      ag.Config.Description,
						trust:     ag.Config.TrustLevel,
						agentType: "user",
					}

					// Determine source
					isBuiltin := builtin.HasAgent(h)
					if isBuiltin {
						userDir := filepath.Join(cfg.AgentsDir, h)
						if _, err := os.Stat(userDir); err != nil {
							info.agentType = "builtin"
						}
					}

					// Apply filters
					if trustFilter != "" {
						agentTrust := string(info.trust)
						if agentTrust == "" {
							agentTrust = "sandboxed"
						}
						if agentTrust != trustFilter {
							continue
						}
					}
					if typeFilter != "" && info.agentType != typeFilter {
						continue
					}

					agents = append(agents, info)
				}

				// JSON output
				if globalOutput.JSON {
					type agentJSON struct {
						Handle      string `json:"handle"`
						Description string `json:"description,omitempty"`
						TrustLevel  string `json:"trust_level"`
						Type        string `json:"type"`
					}
					var jsonAgents []agentJSON
					for _, a := range agents {
						trust := string(a.trust)
						if trust == "" {
							trust = "sandboxed"
						}
						jsonAgents = append(jsonAgents, agentJSON{
							Handle:      a.handle,
							Description: a.desc,
							TrustLevel:  trust,
							Type:        a.agentType,
						})
					}
					globalOutput.PrintData(jsonAgents, "")
					return nil
				}

				// Quiet mode: just list handles
				if globalOutput.Quiet {
					for _, a := range agents {
						fmt.Println(a.handle)
					}
					return nil
				}

				// Format trust level with color
				formatTrust := func(t agent.TrustLevel) string {
					switch t {
					case agent.TrustLevelUnrestricted:
						return dangerStyle.Render("⚠ unrestricted")
					case agent.TrustLevelPrivileged:
						return warnStyle.Render("privileged")
					default:
						return okStyle.Render("sandboxed")
					}
				}

				// Header
				fmt.Println()
				fmt.Println(headerStyle.Render("  Agents"))
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 80)))
				fmt.Println()

				// Column headers
				fmt.Printf("  %-12s %-14s %-10s %s\n",
					countStyle.Render("NAME"),
					countStyle.Render("TRUST"),
					countStyle.Render("TYPE"),
					countStyle.Render("DESCRIPTION"))

				if len(agents) == 0 {
					fmt.Println()
					fmt.Printf("    %s\n", emptyStyle.Render("No agents match the filters"))
				} else {
					for _, a := range agents {
						// Truncate description if too long
						desc := a.desc
						if len(desc) > 45 {
							desc = desc[:42] + "..."
						}
						if desc == "" {
							desc = emptyStyle.Render("(no description)")
						}
						fmt.Printf("  %-12s %-14s %-10s %s\n",
							handleStyle.Render(a.handle),
							formatTrust(a.trust),
							descStyle.Render(a.agentType),
							desc)
					}
				}

				fmt.Println()
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 80)))
				fmt.Println(countStyle.Render(fmt.Sprintf("  %d agents", len(agents))))
				fmt.Println()

				return nil
			})
		},
	}

	cmd.Flags().StringVar(&trustFilter, "trust", "", "filter by trust level (sandboxed, privileged, unrestricted)")
	cmd.Flags().StringVar(&typeFilter, "type", "", "filter by type (builtin, user)")

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

		// I/O Schemas
		inputSchema  string
		outputSchema string

		// Guardrails
		noGuardrails bool
	)

	cmd := &cobra.Command{
		Use:   "create @handle",
		Short: "Create a new agent",
		Long: `Create a new agent with the specified configuration.

The @handle argument is required. A model is also required - either via --model
flag or by setting default_model in ayo.json.

For help designing agents, chat with @ayo:
  ayo "help me create an agent for code review"

Examples:
  # Minimal agent
  ayo agents create @helper -m gpt-5.2

  # With description and custom system prompt
  ayo agents create @reviewer \
    -m gpt-5.2 \
    -d "Reviews code for best practices" \
    -f ~/prompts/reviewer.md

  # With tools and skills
  ayo agents create @debugger \
    -m gpt-5.2 \
    -t bash,memory \
    --skills debugging

  # Agent with I/O schemas for pipelines
  ayo agents create @analyzer \
    -m gpt-5.2 \
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

				// Tools and skills are empty by default - user must explicitly add them
				// No automatic defaults

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
					fmt.Println("  I/O Schemas: enabled")
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

	// Schema flags for pipelines
	cmd.Flags().StringVar(&inputSchema, "input-schema", "", "JSON schema file for validating stdin input")
	cmd.Flags().StringVar(&outputSchema, "output-schema", "", "JSON schema file for structuring stdout output")

	// Guardrails
	cmd.Flags().BoolVar(&noGuardrails, "no-guardrails", false, "disable safety guardrails (dangerous - use with caution)")

	// Internal flags for @ayo agent creation (hidden)
	// These are used when @ayo creates agents programmatically
	cmd.Flags().String("created-by", "", "internal: agent that created this agent")
	cmd.Flags().String("creation-reason", "", "internal: why this agent was created")
	_ = cmd.Flags().MarkHidden("created-by")
	_ = cmd.Flags().MarkHidden("creation-reason")

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

				// JSON output
				if globalOutput.JSON {
					type agentJSON struct {
						Handle      string   `json:"handle"`
						Description string   `json:"description,omitempty"`
						Model       string   `json:"model"`
						TrustLevel  string   `json:"trust_level"`
						Type        string   `json:"type"`
						Path        string   `json:"path"`
						Skills      []string `json:"skills,omitempty"`
						Tools       []string `json:"allowed_tools,omitempty"`
					}
					trust := string(ag.Config.TrustLevel)
					if trust == "" {
						trust = "sandboxed"
					}
					agentType := "user"
					if ag.BuiltIn {
						agentType = "builtin"
					}
					skillNames := make([]string, len(ag.Skills))
					for i, s := range ag.Skills {
						skillNames[i] = s.Name
					}
					globalOutput.PrintData(agentJSON{
						Handle:      ag.Handle,
						Description: ag.Config.Description,
						Model:       ag.Model,
						TrustLevel:  trust,
						Type:        agentType,
						Path:        ag.Dir,
						Skills:      skillNames,
						Tools:       ag.Config.AllowedTools,
					}, "")
					return nil
				}

				// Color palette
				purple := lipgloss.Color("#a78bfa")
				cyan := lipgloss.Color("#67e8f9")
				muted := lipgloss.Color("#6b7280")
				text := lipgloss.Color("#e5e7eb")
				subtle := lipgloss.Color("#374151")
				yellow := lipgloss.Color("#fbbf24")
				red := lipgloss.Color("#ef4444")
				green := lipgloss.Color("#34d399")

				// Styles
				headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
				iconStyle := lipgloss.NewStyle().Foreground(cyan)
				labelStyle := lipgloss.NewStyle().Foreground(muted)
				valueStyle := lipgloss.NewStyle().Foreground(text)
				dividerStyle := lipgloss.NewStyle().Foreground(subtle)
				warnStyle := lipgloss.NewStyle().Foreground(yellow)
				dangerStyle := lipgloss.NewStyle().Foreground(red)
				okStyle := lipgloss.NewStyle().Foreground(green)

				fmt.Println()
				fmt.Println("  " + iconStyle.Render("◆") + " " + headerStyle.Render(ag.Handle))
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))

				source := "user"
				if ag.BuiltIn {
					source = "built-in"
				}
				fmt.Printf("  %s %s\n", labelStyle.Render("Source:"), valueStyle.Render(source))
				fmt.Printf("  %s  %s\n", labelStyle.Render("Model:"), valueStyle.Render(ag.Model))

				// Trust level with color
				trust := ag.Config.TrustLevel
				var trustDisplay string
				switch trust {
				case agent.TrustLevelUnrestricted:
					trustDisplay = dangerStyle.Render("⚠ unrestricted")
				case agent.TrustLevelPrivileged:
					trustDisplay = warnStyle.Render("privileged")
				default:
					trustDisplay = okStyle.Render("sandboxed")
				}
				fmt.Printf("  %s  %s\n", labelStyle.Render("Trust:"), trustDisplay)

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

func rmAgentCmd(cfgPath *string) *cobra.Command {
	var (
		force   bool
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:     "rm @handle",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove an agent",
		Long: `Remove a user-defined agent.

Built-in agents cannot be removed. Use with caution - this permanently
deletes the agent directory including config.json, system.md, and any
agent-specific skills.

Examples:
  # Remove with confirmation prompt
  ayo agents rm @myagent

  # Skip confirmation (dangerous)
  ayo agents rm @myagent --force

  # Preview what would be deleted
  ayo agents rm @myagent --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Check if agent exists
				ag, err := agent.Load(cfg, handle)
				if err != nil {
					return fmt.Errorf("agent not found: %s", handle)
				}

				// Prevent removing built-in agents
				if ag.BuiltIn {
					return fmt.Errorf("cannot remove built-in agent %s", handle)
				}

				// Get agent directory
				agentDir := ag.Dir

				// Dry run mode
				if dryRun {
					fmt.Println("Would remove:")
					fmt.Printf("  %s\n", agentDir)
					
					// List contents
					entries, _ := os.ReadDir(agentDir)
					for _, e := range entries {
						fmt.Printf("    - %s\n", e.Name())
					}
					return nil
				}

				// Confirmation prompt (unless --force)
				if !force {
					fmt.Printf("Remove agent %s?\n", handle)
					fmt.Printf("  Location: %s\n", agentDir)
					
					// List what will be deleted
					entries, _ := os.ReadDir(agentDir)
					if len(entries) > 0 {
						fmt.Println("  Contents:")
						for _, e := range entries {
							fmt.Printf("    - %s\n", e.Name())
						}
					}
					
					fmt.Print("\nType the agent handle to confirm: ")
					var confirm string
					fmt.Scanln(&confirm)
					
					if confirm != handle {
						return fmt.Errorf("confirmation failed: expected %s, got %s", handle, confirm)
					}
				}

				// Remove the agent directory
				if err := os.RemoveAll(agentDir); err != nil {
					return fmt.Errorf("remove agent: %w", err)
				}

				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("✓ Removed agent: " + handle))
				return nil
			})
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be deleted without removing")

	return cmd
}

func statusAgentsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show active agent sessions",
		Long: `List all active agent sessions managed by the daemon.

Shows running agents with their status, start time, and last activity.

Examples:
  ayo agents status
  ayo agents status --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := daemon.ConnectOrStart(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			result, err := client.SessionList(ctx)
			if err != nil {
				return fmt.Errorf("list sessions: %w", err)
			}

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(result.Sessions)
			}

			if len(result.Sessions) == 0 {
				fmt.Println("No active agent sessions")
				return nil
			}

			// Styles
			purple := lipgloss.Color("#a78bfa")
			cyan := lipgloss.Color("#67e8f9")
			muted := lipgloss.Color("#6b7280")
			subtle := lipgloss.Color("#374151")
			green := lipgloss.Color("#34d399")
			yellow := lipgloss.Color("#fbbf24")

			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
			dividerStyle := lipgloss.NewStyle().Foreground(subtle)
			handleStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
			runningStyle := lipgloss.NewStyle().Foreground(green)
			idleStyle := lipgloss.NewStyle().Foreground(yellow)
			mutedStyle := lipgloss.NewStyle().Foreground(muted)

			fmt.Println()
			fmt.Println(headerStyle.Render("  Active Agent Sessions"))
			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))
			fmt.Println()

			// Header row
			fmt.Printf("  %-15s %-10s %-15s %-15s\n",
				mutedStyle.Render("Agent"),
				mutedStyle.Render("Status"),
				mutedStyle.Render("Started"),
				mutedStyle.Render("Last Active"))

			for _, sess := range result.Sessions {
				statusStr := sess.Status
				var styledStatus string
				switch sess.Status {
				case "running":
					styledStatus = runningStyle.Render(statusStr)
				case "idle":
					styledStatus = idleStyle.Render(statusStr)
				default:
					styledStatus = mutedStyle.Render(statusStr)
				}

				started := formatTimeAgo(sess.StartedAt)
				lastActive := formatTimeAgo(sess.LastActive)

				fmt.Printf("  %-15s %-10s %-15s %-15s\n",
					handleStyle.Render(sess.AgentHandle),
					styledStatus,
					started,
					lastActive)
			}

			fmt.Println()
			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))
			fmt.Println(mutedStyle.Render(fmt.Sprintf("  %d active sessions", len(result.Sessions))))
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}

func wakeAgentCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "wake @handle",
		Short: "Start an agent session",
		Long: `Wake up an agent by starting a new session.

If the agent already has an active session, returns the existing session.
If the daemon is not running, it will be started automatically.

Examples:
  ayo agents wake @ayo
  ayo agents wake @research --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])
			ctx := cmd.Context()

			client, err := daemon.ConnectOrStart(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			result, err := client.AgentWake(ctx, handle)
			if err != nil {
				return fmt.Errorf("wake agent: %w", err)
			}

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(result.Session)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Agent %s is awake", handle)))
			fmt.Printf("  Session: %s\n", result.Session.ID)
			fmt.Printf("  Status:  %s\n", result.Session.Status)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}

func sleepAgentCmd() *cobra.Command {
	var quiet bool

	cmd := &cobra.Command{
		Use:   "sleep @handle",
		Short: "Stop an agent session",
		Long: `Put an agent to sleep by stopping its session.

The agent can be woken again with 'ayo agents wake'.

Examples:
  ayo agents sleep @crush
  ayo agents sleep @research --quiet`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])
			ctx := cmd.Context()

			client, err := daemon.ConnectOrStart(ctx)
			if err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.AgentSleep(ctx, handle); err != nil {
				return fmt.Errorf("sleep agent: %w", err)
			}

			if !quiet {
				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ Agent %s is asleep", handle)))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")

	return cmd
}
