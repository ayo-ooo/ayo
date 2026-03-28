package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/ayo-ooo/ayo/internal/registry"
	"github.com/spf13/cobra"
)

var (
	listJSON   bool
	listType   string
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List registered agents",
	Long: `List all agents registered in the ayo registry.

Examples:
  ayo list                    List all agents
  ayo list --type tool        List only tool agents
  ayo list --json             Output as JSON (for machine consumption)`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := listAgents(); err != nil {
			exitError(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	listCmd.Flags().StringVar(&listType, "type", "", "Filter by type (tool, conversational)")
}

func listAgents() error {
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	agents := reg.List(listType)

	if listJSON {
		data, err := json.MarshalIndent(agents, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if len(agents) == 0 {
		fmt.Println("No agents registered.")
		fmt.Println()
		fmt.Println("Register an agent with:")
		fmt.Println("  ayo register ./my-agent")
		fmt.Println("  ayo runthat ./my-agent --register")
		return nil
	}

	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	versionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

	fmt.Fprintf(os.Stdout, "%s\n\n", headerStyle.Render(fmt.Sprintf("%d agent(s) registered", len(agents))))

	for _, agent := range agents {
		name := nameStyle.Render(agent.Name)
		version := versionStyle.Render("v" + agent.Version)
		agentType := typeStyle.Render(agent.Type)
		desc := dimStyle.Render(agent.Description)

		fmt.Fprintf(os.Stdout, "  %s %s  %s\n", name, version, agentType)
		if agent.Description != "" {
			fmt.Fprintf(os.Stdout, "  %s\n", desc)
		}
		if agent.BinaryPath != "" {
			fmt.Fprintf(os.Stdout, "  %s\n", dimStyle.Render(agent.BinaryPath))
		}
		fmt.Println()
	}

	return nil
}
