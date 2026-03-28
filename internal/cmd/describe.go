package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"charm.land/lipgloss/v2"
	"github.com/ayo-ooo/ayo/internal/registry"
	"github.com/spf13/cobra"
)

var describeJSON bool

var describeCmd = &cobra.Command{
	Use:   "describe <agent>",
	Short: "Show detailed information about a registered agent",
	Long: `Show detailed information about a registered agent, including its
configuration, input/output schemas, skills, and hooks.

Examples:
  ayo describe my-agent          Human-readable output
  ayo describe my-agent --json   Machine-readable JSON output`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := describeAgent(args[0]); err != nil {
			exitError(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)
	describeCmd.Flags().BoolVar(&describeJSON, "json", false, "Output as JSON")
}

func describeAgent(name string) error {
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	entry := reg.Get(name)
	if entry == nil {
		return fmt.Errorf("agent '%s' not found in registry", name)
	}

	// Try to get live metadata from the binary
	var liveMetadata map[string]any
	if entry.BinaryPath != "" {
		if _, err := os.Stat(entry.BinaryPath); err == nil {
			cmd := exec.Command(entry.BinaryPath, "--ayo-describe")
			output, err := cmd.Output()
			if err == nil {
				json.Unmarshal(output, &liveMetadata)
			}
		}
	}

	if describeJSON {
		// Prefer live metadata if available, fall back to registry
		if liveMetadata != nil {
			// Augment with registry paths
			liveMetadata["source_path"] = entry.SourcePath
			liveMetadata["binary_path"] = entry.BinaryPath
			liveMetadata["registered_at"] = entry.RegisteredAt

			// Add invocation examples
			invocation := map[string]string{
				"direct":   entry.BinaryPath,
				"via_ayo":  fmt.Sprintf("ayo run %s", entry.Name),
				"piped":    fmt.Sprintf("echo '{...}' | ayo run %s", entry.Name),
			}
			liveMetadata["invocation"] = invocation

			data, err := json.MarshalIndent(liveMetadata, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling JSON: %w", err)
			}
			fmt.Println(string(data))
		} else {
			data, err := json.MarshalIndent(entry, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling JSON: %w", err)
			}
			fmt.Println(string(data))
		}
		return nil
	}

	// Human-readable output
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("8"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

	fmt.Println()
	fmt.Printf("  %s\n", nameStyle.Render(entry.Name))
	fmt.Println()

	if entry.Description != "" {
		fmt.Printf("  %s\n\n", valueStyle.Render(entry.Description))
	}

	fmt.Printf("  %s %s\n", labelStyle.Render("Version:"), valueStyle.Render(entry.Version))
	fmt.Printf("  %s %s\n", labelStyle.Render("Type:   "), typeStyle.Render(entry.Type))

	if entry.SourcePath != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Source: "), pathStyle.Render(entry.SourcePath))
	}
	if entry.BinaryPath != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Binary: "), pathStyle.Render(entry.BinaryPath))
	}

	// Show live metadata details
	if liveMetadata != nil {
		if skills, ok := liveMetadata["skills"].([]any); ok && len(skills) > 0 {
			fmt.Printf("\n  %s\n", labelStyle.Render("Skills:"))
			for _, s := range skills {
				fmt.Printf("    - %s\n", valueStyle.Render(fmt.Sprint(s)))
			}
		}

		if hooks, ok := liveMetadata["hooks"].([]any); ok && len(hooks) > 0 {
			fmt.Printf("\n  %s\n", labelStyle.Render("Hooks:"))
			for _, h := range hooks {
				fmt.Printf("    - %s\n", valueStyle.Render(fmt.Sprint(h)))
			}
		}

		if schema, ok := liveMetadata["input_schema"]; ok && schema != nil {
			data, _ := json.MarshalIndent(schema, "    ", "  ")
			fmt.Printf("\n  %s\n", labelStyle.Render("Input Schema:"))
			fmt.Printf("    %s\n", string(data))
		}

		if schema, ok := liveMetadata["output_schema"]; ok && schema != nil {
			data, _ := json.MarshalIndent(schema, "    ", "  ")
			fmt.Printf("\n  %s\n", labelStyle.Render("Output Schema:"))
			fmt.Printf("    %s\n", string(data))
		}
	}

	// Invocation examples
	fmt.Printf("\n  %s\n", labelStyle.Render("Invocation:"))
	if entry.BinaryPath != "" {
		fmt.Printf("    %s\n", valueStyle.Render(entry.BinaryPath))
	}
	fmt.Printf("    %s\n", valueStyle.Render(fmt.Sprintf("ayo run %s", entry.Name)))
	fmt.Println()

	return nil
}
