package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/capabilities"
	"github.com/alexcabrera/ayo/internal/config"
	// TODO: Re-implement db for build system
	// "github.com/alexcabrera/ayo/internal/db" - Removed as part of framework cleanup
	"github.com/alexcabrera/ayo/internal/paths"
)

func capabilitiesAgentsCmd(cfgPath *string) *cobra.Command {
	var (
		all    bool
		search string
	)

	cmd := &cobra.Command{
		Use:   "capabilities [agent]",
		Short: "Show agent capabilities",
		Long: `Show inferred capabilities for agents.

Capabilities are automatically inferred from agent system prompts, skills,
and schemas. They are used by @ayo to select the best agent for a task.

Examples:
  # Show capabilities for a specific agent
  ayo agents capabilities @code-reviewer

  # List all capabilities across all agents
  ayo agents capabilities --all

  # Search for agents with specific capabilities
  ayo agents capabilities --search "code review"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				// Capabilities system has been removed as part of framework cleanup
				// In the build system, capabilities are determined at build time
				// and embedded in the compiled executable
				return fmt.Errorf("agent capabilities are no longer supported in the build system. Capabilities are now determined at build time and embedded in executables")
			})
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "list capabilities for all agents")
	cmd.Flags().StringVar(&search, "search", "", "search for agents with matching capabilities")

	cmd.AddCommand(refreshCapabilitiesCmd(cfgPath))

	return cmd
}

func showAgentCapabilities(ctx context.Context, cfg config.Config, repo *capabilities.Repository, handle string) error {
	// Load agent to verify it exists
	ag, err := agent.Load(cfg, handle)
	if err != nil {
		return fmt.Errorf("agent not found: %s", handle)
	}

	// Get capabilities
	caps, err := repo.GetCapabilities(ctx, ag.Handle)
	if err != nil {
		return fmt.Errorf("get capabilities: %w", err)
	}

	// JSON output
	if globalOutput.JSON {
		type capJSON struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Confidence  float64 `json:"confidence"`
			Source      string  `json:"source"`
		}
		type output struct {
			Agent        string    `json:"agent"`
			Capabilities []capJSON `json:"capabilities"`
			LastUpdated  string    `json:"last_updated,omitempty"`
			InputHash    string    `json:"input_hash,omitempty"`
		}

		out := output{Agent: handle}
		for _, cap := range caps {
			out.Capabilities = append(out.Capabilities, capJSON{
				Name:        cap.Name,
				Description: cap.Description,
				Confidence:  cap.Confidence,
				Source:      cap.Source,
			})
			if out.LastUpdated == "" && cap.UpdatedAt > 0 {
				out.LastUpdated = time.Unix(cap.UpdatedAt, 0).Format(time.RFC3339)
				out.InputHash = cap.InputHash
			}
		}
		globalOutput.PrintData(out, "")
		return nil
	}

	// Quiet mode
	if globalOutput.Quiet {
		for _, cap := range caps {
			fmt.Println(cap.Name)
		}
		return nil
	}

	// Color palette
	purple := lipgloss.Color("#a78bfa")
	cyan := lipgloss.Color("#67e8f9")
	muted := lipgloss.Color("#6b7280")
	text := lipgloss.Color("#e5e7eb")
	subtle := lipgloss.Color("#374151")
	green := lipgloss.Color("#34d399")
	yellow := lipgloss.Color("#fbbf24")
	red := lipgloss.Color("#f87171")

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	subheaderStyle := lipgloss.NewStyle().Bold(true).Foreground(cyan)
	mutedStyle := lipgloss.NewStyle().Foreground(muted)
	textStyle := lipgloss.NewStyle().Foreground(text)
	subtleStyle := lipgloss.NewStyle().Foreground(subtle)
	greenStyle := lipgloss.NewStyle().Foreground(green)
	yellowStyle := lipgloss.NewStyle().Foreground(yellow)
	redStyle := lipgloss.NewStyle().Foreground(red)

	// Header
	fmt.Println(headerStyle.Render("Agent Capabilities"))
	fmt.Println(subheaderStyle.Render(handle))
	fmt.Println()

	// Capabilities list
	if len(caps) == 0 {
		fmt.Println(mutedStyle.Render("No capabilities found"))
		fmt.Println(mutedStyle.Render("Run: ayo agents capabilities refresh --all"))
		return nil
	}

	for _, cap := range caps {
		fmt.Println(greenStyle.Render("• " + cap.Name))
		if cap.Description != "" {
			fmt.Println(textStyle.Render("  " + cap.Description))
		}
		if cap.Source != "" {
			fmt.Println(subtleStyle.Render("  Source: " + cap.Source))
		}
		if cap.Confidence > 0 {
			confidence := fmt.Sprintf("%.1f%%", cap.Confidence*100)
			fmt.Println(subtleStyle.Render("  Confidence: "+confidence))
		}
		fmt.Println()
	}

	// Footer
	fmt.Println(subtleStyle.Render("Capabilities help @ayo route tasks to the right agent"))

	return nil
}

func listAllCapabilities(ctx context.Context, cfg config.Config, repo *capabilities.Repository) error {
	// Get all agents
	agents, err := agent.List(cfg)
	if err != nil {
		return fmt.Errorf("list agents: %w", err)
	}

	if len(agents) == 0 {
		fmt.Println("No agents found")
		return nil
	}

	// JSON output
	if globalOutput.JSON {
		type agentCapJSON struct {
			Agent string `json:"agent"`
			Capabilities []struct {
				Name        string  `json:"name"`
				Description string  `json:"description"`
				Confidence  float64 `json:"confidence"`
				Source      string  `json:"source"`
			} `json:"capabilities"`
		}

		var result []agentCapJSON
		for _, ag := range agents {
			caps, err := repo.GetCapabilities(ctx, ag.Handle)
			if err != nil {
				continue
			}

			agentCaps := agentCapJSON{Agent: ag.Handle}
			for _, cap := range caps {
				agentCaps.Capabilities = append(agentCaps.Capabilities, struct {
					Name        string  `json:"name"`
					Description string  `json:"description"`
					Confidence  float64 `json:"confidence"`
					Source      string  `json:"source"`
				}{
					Name:        cap.Name,
					Description: cap.Description,
					Confidence:  cap.Confidence,
					Source:      cap.Source,
				})
			}
			result = append(result, agentCaps)
		}

		globalOutput.PrintData(result, "")
		return nil
	}

	// Quiet mode
	if globalOutput.Quiet {
		for _, ag := range agents {
			caps, err := repo.GetCapabilities(ctx, ag.Handle)
			if err != nil {
				continue
			}
			for _, cap := range caps {
				fmt.Printf("%s:%s\n", ag.Handle, cap.Name)
			}
		}
		return nil
	}

	// Color palette
	purple := lipgloss.Color("#a78bfa")
	cyan := lipgloss.Color("#67e8f9")
	muted := lipgloss.Color("#6b7280")
	text := lipgloss.Color("#e5e7eb")

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	subheaderStyle := lipgloss.NewStyle().Bold(true).Foreground(cyan)
	mutedStyle := lipgloss.NewStyle().Foreground(muted)
	textStyle := lipgloss.NewStyle().Foreground(text)

	// Header
	fmt.Println(headerStyle.Render("All Agent Capabilities"))
	fmt.Println()

	// List capabilities by agent
	for _, ag := range agents {
		caps, err := repo.GetCapabilities(ctx, ag.Handle)
		if err != nil {
			continue
		}

		if len(caps) > 0 {
			fmt.Println(subheaderStyle.Render(ag.Handle))
			for _, cap := range caps {
				fmt.Println(textStyle.Render("  • " + cap.Name))
			}
			fmt.Println()
		}
	}

	// Footer
	fmt.Println(mutedStyle.Render("Use --search to find agents with specific capabilities"))

	return nil
}

func searchCapabilities(ctx context.Context, cfg config.Config, repo *capabilities.Repository, query string) error {
	// Get all agents
	agents, err := agent.List(cfg)
	if err != nil {
		return fmt.Errorf("list agents: %w", err)
	}

	// Search capabilities
	var matches []struct {
		Agent string
		Cap   capabilities.Capability
	}

	for _, ag := range agents {
		caps, err := repo.GetCapabilities(ctx, ag.Handle)
		if err != nil {
			continue
		}

		for _, cap := range caps {
			if strings.Contains(strings.ToLower(cap.Name), strings.ToLower(query)) ||
				(strings.Contains(strings.ToLower(cap.Description), strings.ToLower(query))) {
				matches = append(matches, struct {
					Agent string
					Cap   capabilities.Capability
				}{
					Agent: ag.Handle,
					Cap:   cap,
				})
			}
		}
	}

	// JSON output
	if globalOutput.JSON {
		type matchJSON struct {
			Agent        string  `json:"agent"`
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Confidence  float64 `json:"confidence"`
			Source      string  `json:"source"`
		}

		var result []matchJSON
		for _, match := range matches {
			result = append(result, matchJSON{
				Agent:       match.Agent,
				Name:        match.Cap.Name,
				Description: match.Cap.Description,
				Confidence:  match.Cap.Confidence,
				Source:      match.Cap.Source,
			})
		}

		globalOutput.PrintData(result, "")
		return nil
	}

	// Quiet mode
	if globalOutput.Quiet {
		for _, match := range matches {
			fmt.Printf("%s:%s\n", match.Agent, match.Cap.Name)
		}
		return nil
	}

	// Color palette
	purple := lipgloss.Color("#a78bfa")
	cyan := lipgloss.Color("#67e8f9")
	muted := lipgloss.Color("#6b7280")
	text := lipgloss.Color("#e5e7eb")
	green := lipgloss.Color("#34d399")

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	subheaderStyle := lipgloss.NewStyle().Bold(true).Foreground(cyan)
	mutedStyle := lipgloss.NewStyle().Foreground(muted)
	textStyle := lipgloss.NewStyle().Foreground(text)
	greenStyle := lipgloss.NewStyle().Foreground(green)

	// Header
	fmt.Println(headerStyle.Render(fmt.Sprintf("Search Results for '%s'", query)))
	fmt.Println()

	if len(matches) == 0 {
		fmt.Println(mutedStyle.Render("No matching capabilities found"))
		return nil
	}

	// Group by agent
	byAgent := make(map[string][]capabilities.Capability)
	for _, match := range matches {
		byAgent[match.Agent] = append(byAgent[match.Agent], match.Cap)
	}

	for agent, caps := range byAgent {
		fmt.Println(subheaderStyle.Render(agent))
		for _, cap := range caps {
			fmt.Println(greenStyle.Render("  • " + cap.Name))
			if cap.Description != "" {
				fmt.Println(textStyle.Render("    " + cap.Description))
			}
		}
		fmt.Println()
	}

	return nil
}

func refreshCapabilitiesCmd(cfgPath *string) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "refresh [agent]",
		Short: "Refresh agent capabilities",
		Long: `Refresh inferred capabilities for agents.

This re-analyzes agent system prompts, skills, and schemas to update
capability inferences.

Examples:
  # Refresh capabilities for a specific agent
  ayo agents capabilities refresh @code-reviewer

  # Refresh capabilities for all agents
  ayo agents capabilities refresh --all`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				// Capabilities system has been removed as part of framework cleanup
				// In the build system, capabilities are determined at build time
				// and embedded in the compiled executable
				return fmt.Errorf("agent capabilities refresh is no longer supported in the build system. Capabilities are now determined at build time and embedded in executables")
			})
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "refresh capabilities for all agents")

	return cmd
}
