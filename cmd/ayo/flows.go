package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	// TODO: Re-implement db for build system
	// "github.com/alexcabrera/ayo/internal/db" - Removed as part of framework cleanup
	"github.com/alexcabrera/ayo/internal/flows"
	"github.com/alexcabrera/ayo/internal/paths"
)

func newFlowsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "flow",
		Short:   "Manage flows",
		Aliases: []string{"flows"},
		Long: `Manage flows - composable agent pipelines.

Flows come in two types:
  - Shell flows (.sh): Bash scripts with JSON I/O frontmatter
  - YAML flows (.yaml): Declarative multi-step workflows with dependencies
                        and parallel execution

Discovery priority (first found wins):
  1. Project flows (.ayo/flows/)
  2. User flows (~/.config/ayo/flows/)
  3. Built-in flows (~/.local/share/ayo/flows/)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listFlowsCmd().RunE(cmd, args)
		},
	}

	cmd.AddCommand(listFlowsCmd())
	cmd.AddCommand(showFlowCmd())
	cmd.AddCommand(validateFlowCmd())
	cmd.AddCommand(newFlowCmd())
	cmd.AddCommand(rmFlowCmd())
	cmd.AddCommand(historyFlowsCmd(cfgPath))
	cmd.AddCommand(statsFlowsCmd())

	return cmd
}

func listFlowsCmd() *cobra.Command {
	var sourceFilter string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available flows",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Discover flows from all directories
			dirs := paths.FlowsDirs()
			discovered, err := flows.Discover(dirs)
			if err != nil {
				return fmt.Errorf("discover flows: %w", err)
			}

			// Filter by source if specified
			if sourceFilter != "" {
				var filtered []flows.Flow
				for _, f := range discovered {
					if string(f.Source) == sourceFilter {
						filtered = append(filtered, f)
					}
				}
				discovered = filtered
			}

			if jsonOutput {
				return outputFlowsJSON(discovered)
			}

			if len(discovered) == 0 {
				fmt.Println("No flows found.")
				fmt.Println("\nCreate a flow with: ayo flows new <name>")
				return nil
			}

			return outputFlowsTable(discovered)
		},
	}

	cmd.Flags().StringVar(&sourceFilter, "source", "", "Filter by source (built-in, user, project)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func outputFlowsJSON(flowList []flows.Flow) error {
	type flowJSON struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Source      string `json:"source"`
		Path        string `json:"path"`
		HasInput    bool   `json:"has_input_schema"`
		HasOutput   bool   `json:"has_output_schema"`
	}

	var output []flowJSON
	for _, f := range flowList {
		output = append(output, flowJSON{
			Name:        f.Name,
			Description: f.Description,
			Source:      string(f.Source),
			Path:        f.Path,
			HasInput:    f.HasInputSchema(),
			HasOutput:   f.HasOutputSchema(),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func outputFlowsTable(flowList []flows.Flow) error {
	// Color palette
	purple := lipgloss.Color("#a78bfa")
	cyan := lipgloss.Color("#67e8f9")
	muted := lipgloss.Color("#6b7280")
	green := lipgloss.Color("#34d399")

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	nameStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
	sourceStyle := lipgloss.NewStyle().Foreground(muted)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e5e7eb"))
	yesStyle := lipgloss.NewStyle().Foreground(green)
	noStyle := lipgloss.NewStyle().Foreground(muted)

	// Calculate column widths
	nameWidth := 20
	sourceWidth := 10
	inputWidth := 6
	outputWidth := 6
	for _, f := range flowList {
		if len(f.Name) > nameWidth {
			nameWidth = len(f.Name)
		}
	}
	if nameWidth > 30 {
		nameWidth = 30
	}

	// Print header
	fmt.Printf("%s  %s  %s  %s  %s\n",
		headerStyle.Render(padRight("NAME", nameWidth)),
		headerStyle.Render(padRight("SOURCE", sourceWidth)),
		headerStyle.Render(padRight("INPUT", inputWidth)),
		headerStyle.Render(padRight("OUTPUT", outputWidth)),
		headerStyle.Render("DESCRIPTION"),
	)

	// Print flows
	for _, f := range flowList {
		name := f.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		source := string(f.Source)
		desc := f.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		inputStr := noStyle.Render("no")
		if f.HasInputSchema() {
			inputStr = yesStyle.Render("yes")
		}

		outputStr := noStyle.Render("no")
		if f.HasOutputSchema() {
			outputStr = yesStyle.Render("yes")
		}

		fmt.Printf("%s  %s  %s  %s  %s\n",
			nameStyle.Render(padRight(name, nameWidth)),
			sourceStyle.Render(padRight(source, sourceWidth)),
			padRight(inputStr, inputWidth),
			padRight(outputStr, outputWidth),
			descStyle.Render(desc),
		)
	}

	return nil
}

func showFlowCmd() *cobra.Command {
	var jsonOutput bool
	var showScript bool

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show flow details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Discover flows
			dirs := paths.FlowsDirs()
			discovered, err := flows.Discover(dirs)
			if err != nil {
				return fmt.Errorf("discover flows: %w", err)
			}

			// Find the flow
			var flow *flows.Flow
			for _, f := range discovered {
				if f.Name == name {
					flow = &f
					break
				}
			}

			if flow == nil {
				return fmt.Errorf("flow not found: %s", name)
			}

			if jsonOutput {
				return outputFlowJSON(flow)
			}

			return outputFlowDetails(flow, showScript)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&showScript, "script", false, "Show full script content")

	return cmd
}

func outputFlowJSON(f *flows.Flow) error {
	output := map[string]any{
		"name":              f.Name,
		"description":       f.Description,
		"source":            string(f.Source),
		"path":              f.Path,
		"dir":               f.Dir,
		"has_input_schema":  f.HasInputSchema(),
		"has_output_schema": f.HasOutputSchema(),
		"metadata": map[string]string{
			"version": f.Metadata.Version,
			"author":  f.Metadata.Author,
		},
		"script": f.Raw.Script,
	}

	if f.HasInputSchema() {
		output["input_schema_path"] = f.InputSchemaPath
	}
	if f.HasOutputSchema() {
		output["output_schema_path"] = f.OutputSchemaPath
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func outputFlowDetails(f *flows.Flow, showScript bool) error {
	// Styles
	purple := lipgloss.Color("#a78bfa")
	cyan := lipgloss.Color("#67e8f9")
	muted := lipgloss.Color("#6b7280")
	green := lipgloss.Color("#34d399")

	labelStyle := lipgloss.NewStyle().Foreground(muted).Width(14)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e5e7eb"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	pathStyle := lipgloss.NewStyle().Foreground(cyan)
	schemaYes := lipgloss.NewStyle().Foreground(green).Render("yes")
	schemaNo := lipgloss.NewStyle().Foreground(muted).Render("no")

	fmt.Println(headerStyle.Render(f.Name))
	fmt.Println()

	fmt.Printf("%s %s\n", labelStyle.Render("Description:"), valueStyle.Render(f.Description))
	fmt.Printf("%s %s\n", labelStyle.Render("Source:"), valueStyle.Render(string(f.Source)))
	fmt.Printf("%s %s\n", labelStyle.Render("Path:"), pathStyle.Render(f.Path))

	if f.Metadata.Version != "" {
		fmt.Printf("%s %s\n", labelStyle.Render("Version:"), valueStyle.Render(f.Metadata.Version))
	}
	if f.Metadata.Author != "" {
		fmt.Printf("%s %s\n", labelStyle.Render("Author:"), valueStyle.Render(f.Metadata.Author))
	}

	fmt.Println()

	// Schemas
	inputSchemaStr := schemaNo
	if f.HasInputSchema() {
		inputSchemaStr = schemaYes + " " + pathStyle.Render(filepath.Base(f.InputSchemaPath))
	}
	fmt.Printf("%s %s\n", labelStyle.Render("Input Schema:"), inputSchemaStr)

	outputSchemaStr := schemaNo
	if f.HasOutputSchema() {
		outputSchemaStr = schemaYes + " " + pathStyle.Render(filepath.Base(f.OutputSchemaPath))
	}
	fmt.Printf("%s %s\n", labelStyle.Render("Output Schema:"), outputSchemaStr)

	// Script preview or full
	if showScript {
		fmt.Println()
		fmt.Println(headerStyle.Render("Script:"))
		fmt.Println()

		// Print full script
		lines := strings.Split(f.Raw.Script, "\n")
		for _, line := range lines {
			fmt.Println("  " + line)
		}
	} else if f.Raw.Script != "" {
		fmt.Println()
		fmt.Println(headerStyle.Render("Script Preview:"))
		fmt.Println()

		// Show first 10 lines
		lines := strings.Split(f.Raw.Script, "\n")
		maxLines := min(10, len(lines))
		for i := range maxLines {
			fmt.Println("  " + lines[i])
		}
		if len(lines) > 10 {
			fmt.Printf("\n  ... (%d more lines, use --script to see all)\n", len(lines)-10)
		}
	}

	return nil
}

// padRight pads a string to the specified width
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func validateFlowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <name-or-path>",
		Short: "Validate a flow by name or path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrPath := args[0]

			// Try to find flow by name or path
			flow, err := flows.FindByName(nameOrPath, paths.FlowsDirs())
			if err != nil {
				// Styled error output
				errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))
				fmt.Println(errorStyle.Render("x Flow validation failed"))
				fmt.Printf("  %v\n", err)
				os.Exit(1)
			}

			// Success output
			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#34d399"))
			muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))

			fmt.Println(successStyle.Render("v Flow is valid"))
			fmt.Printf("  Name: %s\n", flow.Name)
			fmt.Printf("  Description: %s\n", flow.Description)

			inputStr := muted.Render("not defined")
			if flow.HasInputSchema() {
				inputStr = "defined"
			}
			fmt.Printf("  Input schema: %s\n", inputStr)

			outputStr := muted.Render("not defined")
			if flow.HasOutputSchema() {
				outputStr = "defined"
			}
			fmt.Printf("  Output schema: %s\n", outputStr)

			return nil
		},
	}

	return cmd
}

func newFlowCmd() *cobra.Command {
	var project bool
	var withSchemas bool
	var force bool

	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new flow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Determine target directory
			var targetDir string
			if project {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				targetDir = filepath.Join(wd, ".ayo", "flows")
			} else {
				targetDir = paths.UserFlowsDir()
			}

			// Create directory if needed
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}

			var flowPath string
			if withSchemas {
				// Create package directory
				pkgDir := filepath.Join(targetDir, name)
				if !force {
					if _, err := os.Stat(pkgDir); err == nil {
						return fmt.Errorf("flow already exists: %s (use --force to overwrite)", pkgDir)
					}
				}
				if err := os.MkdirAll(pkgDir, 0755); err != nil {
					return fmt.Errorf("create package directory: %w", err)
				}

				flowPath = filepath.Join(pkgDir, "flow.sh")

				// Create input schema
				inputSchema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "input": {
      "type": "string",
      "description": "Input value"
    }
  },
  "required": ["input"]
}
`
				if err := os.WriteFile(filepath.Join(pkgDir, "input.jsonschema"), []byte(inputSchema), 0644); err != nil {
					return fmt.Errorf("write input schema: %w", err)
				}

				// Create output schema
				outputSchema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "result": {
      "type": "string",
      "description": "Result value"
    }
  },
  "required": ["result"]
}
`
				if err := os.WriteFile(filepath.Join(pkgDir, "output.jsonschema"), []byte(outputSchema), 0644); err != nil {
					return fmt.Errorf("write output schema: %w", err)
				}

				fmt.Printf("Created: %s/\n", pkgDir)
				fmt.Println("  - flow.sh")
				fmt.Println("  - input.jsonschema")
				fmt.Println("  - output.jsonschema")
			} else {
				flowPath = filepath.Join(targetDir, name+".sh")
				if !force {
					if _, err := os.Stat(flowPath); err == nil {
						return fmt.Errorf("flow already exists: %s (use --force to overwrite)", flowPath)
					}
				}
				fmt.Printf("Created: %s\n", flowPath)
			}

			// Create flow script
			flowContent := fmt.Sprintf(`#!/usr/bin/env bash
# ayo:flow
# name: %s
# description: TODO: Describe what this flow does

set -euo pipefail

INPUT="${1:-$(cat)}"

# TODO: Implement your flow
# Example: pipe input through an agent
# echo "$INPUT" | ayo @ayo "Process this input and return JSON"

# For now, just echo the input
echo "$INPUT"
`, name)

			if err := os.WriteFile(flowPath, []byte(flowContent), 0755); err != nil {
				return fmt.Errorf("write flow: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&project, "project", false, "Create in project directory (.ayo/flows/)")
	cmd.Flags().BoolVar(&withSchemas, "with-schemas", false, "Create with input/output schemas")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite if exists")

	return cmd
}

func historyFlowsCmd(cfgPath *string) *cobra.Command {
	var flowName string
	var status string
	var limit int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show flow run history",
		Long: `Show history of flow executions.

Filter by:
  --flow <name>    Show runs for a specific flow
  --status <status> Filter by status (success, failed, timeout, running)
  --limit <n>      Limit number of results (default 50)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			history := flows.NewHistoryService(queries)

			filter := flows.RunFilter{
				FlowName: flowName,
				Status:   flows.RunStatus(status),
				Limit:    int64(limit),
			}

			runs, err := history.ListRuns(cmd.Context(), filter)
			if err != nil {
				return fmt.Errorf("list runs: %w", err)
			}

			if jsonOutput {
				return outputHistoryJSON(runs)
			}

			if len(runs) == 0 {
				fmt.Println("No flow runs found.")
				return nil
			}

			return outputHistoryTable(runs)
		},
	}

	cmd.Flags().StringVar(&flowName, "flow", "", "Filter by flow name")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (success, failed, timeout, running)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of runs to show")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	// Add show subcommand
	cmd.AddCommand(historyShowCmd(cfgPath))

	return cmd
}

func historyShowCmd(_ *string) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <run-id>",
		Short: "Show details of a specific run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := args[0]

			_, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			history := flows.NewHistoryService(queries)

			run, err := history.GetRun(cmd.Context(), runID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("run not found: %s", runID)
				}
				return fmt.Errorf("get run: %w", err)
			}

			if jsonOutput {
				return outputRunJSON(run)
			}

			return outputRunDetails(run)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func outputHistoryJSON(runs []*flows.FlowRun) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(runs)
}

func outputHistoryTable(runs []*flows.FlowRun) error {
	// Color palette
	purple := lipgloss.Color("#a78bfa")
	cyan := lipgloss.Color("#67e8f9")
	muted := lipgloss.Color("#6b7280")
	green := lipgloss.Color("#34d399")
	red := lipgloss.Color("#ef4444")
	yellow := lipgloss.Color("#fbbf24")

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	idStyle := lipgloss.NewStyle().Foreground(muted)
	nameStyle := lipgloss.NewStyle().Foreground(cyan)
	successStyle := lipgloss.NewStyle().Foreground(green)
	failedStyle := lipgloss.NewStyle().Foreground(red)
	runningStyle := lipgloss.NewStyle().Foreground(yellow)
	timeoutStyle := lipgloss.NewStyle().Foreground(yellow)

	// Column widths
	idWidth := 8
	nameWidth := 20
	statusWidth := 10
	durationWidth := 10

	// Print header
	fmt.Printf("%s  %s  %s  %s  %s\n",
		headerStyle.Render(padRight("ID", idWidth)),
		headerStyle.Render(padRight("FLOW", nameWidth)),
		headerStyle.Render(padRight("STATUS", statusWidth)),
		headerStyle.Render(padRight("DURATION", durationWidth)),
		headerStyle.Render("STARTED"),
	)

	// Print runs
	for _, run := range runs {
		id := run.ID
		if len(id) > idWidth {
			id = id[:idWidth]
		}

		name := run.FlowName
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		var statusStyled string
		switch run.Status {
		case flows.RunStatusSuccess:
			statusStyled = successStyle.Render(padRight("success", statusWidth))
		case flows.RunStatusFailed:
			statusStyled = failedStyle.Render(padRight("failed", statusWidth))
		case flows.RunStatusTimeout:
			statusStyled = timeoutStyle.Render(padRight("timeout", statusWidth))
		case flows.RunStatusRunning:
			statusStyled = runningStyle.Render(padRight("running", statusWidth))
		default:
			statusStyled = padRight(string(run.Status), statusWidth)
		}

		duration := "-"
		if run.DurationMs > 0 {
			duration = formatDuration(time.Duration(run.DurationMs) * time.Millisecond)
		}

		started := run.StartedAt.Format("2006-01-02 15:04:05")

		fmt.Printf("%s  %s  %s  %s  %s\n",
			idStyle.Render(padRight(id, idWidth)),
			nameStyle.Render(padRight(name, nameWidth)),
			statusStyled,
			padRight(duration, durationWidth),
			started,
		)
	}

	return nil
}

func outputRunJSON(run *flows.FlowRun) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(run)
}

func outputRunDetails(run *flows.FlowRun) error {
	// Styles
	purple := lipgloss.Color("#a78bfa")
	cyan := lipgloss.Color("#67e8f9")
	muted := lipgloss.Color("#6b7280")
	green := lipgloss.Color("#34d399")
	red := lipgloss.Color("#ef4444")

	labelStyle := lipgloss.NewStyle().Foreground(muted).Width(14)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e5e7eb"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	pathStyle := lipgloss.NewStyle().Foreground(cyan)
	successStyle := lipgloss.NewStyle().Foreground(green)
	failedStyle := lipgloss.NewStyle().Foreground(red)

	fmt.Println(headerStyle.Render("Flow Run: " + run.ID))
	fmt.Println()

	fmt.Printf("%s %s\n", labelStyle.Render("Flow:"), valueStyle.Render(run.FlowName))
	fmt.Printf("%s %s\n", labelStyle.Render("Path:"), pathStyle.Render(run.FlowPath))
	fmt.Printf("%s %s\n", labelStyle.Render("Source:"), valueStyle.Render(string(run.FlowSource)))

	var statusStyled string
	switch run.Status {
	case flows.RunStatusSuccess:
		statusStyled = successStyle.Render(string(run.Status))
	case flows.RunStatusFailed, flows.RunStatusError:
		statusStyled = failedStyle.Render(string(run.Status))
	default:
		statusStyled = valueStyle.Render(string(run.Status))
	}
	fmt.Printf("%s %s\n", labelStyle.Render("Status:"), statusStyled)

	if run.ExitCode != nil {
		fmt.Printf("%s %d\n", labelStyle.Render("Exit Code:"), *run.ExitCode)
	}

	fmt.Printf("%s %s\n", labelStyle.Render("Started:"), valueStyle.Render(run.StartedAt.Format(time.RFC3339)))
	if run.FinishedAt != nil {
		fmt.Printf("%s %s\n", labelStyle.Render("Finished:"), valueStyle.Render(run.FinishedAt.Format(time.RFC3339)))
	}
	if run.DurationMs > 0 {
		fmt.Printf("%s %s\n", labelStyle.Render("Duration:"), valueStyle.Render(formatDuration(time.Duration(run.DurationMs)*time.Millisecond)))
	}

	if run.ErrorMessage != "" {
		fmt.Println()
		fmt.Println(headerStyle.Render("Error:"))
		fmt.Println(failedStyle.Render("  " + run.ErrorMessage))
	}

	if run.InputJSON != "" {
		fmt.Println()
		fmt.Println(headerStyle.Render("Input:"))
		fmt.Println(formatJSON(run.InputJSON))
	}

	if run.OutputJSON != "" {
		fmt.Println()
		fmt.Println(headerStyle.Render("Output:"))
		fmt.Println(formatJSON(run.OutputJSON))
	}

	if run.StderrLog != "" {
		fmt.Println()
		fmt.Println(headerStyle.Render("Stderr:"))
		lines := strings.Split(run.StderrLog, "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Println("  " + line)
			}
		}
	}

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

func formatJSON(s string) string {
	var obj any
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return "  " + s
	}
	formatted, err := json.MarshalIndent(obj, "  ", "  ")
	if err != nil {
		return "  " + s
	}
	return "  " + string(formatted)
}

func rmFlowCmd() *cobra.Command {
	var (
		force  bool
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:     "rm <name>",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove a flow",
		Long: `Remove a user-defined flow.

Built-in flows cannot be removed. Use with caution - this permanently
deletes the flow file (and associated schemas for packaged flows).

Examples:
  # Remove with confirmation prompt
  ayo flows rm my-flow

  # Skip confirmation (dangerous)
  ayo flows rm my-flow --force

  # Preview what would be deleted
  ayo flows rm my-flow --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Discover all flows
			dirs := paths.FlowsDirs()
			discovered, err := flows.Discover(dirs)
			if err != nil {
				return fmt.Errorf("discover flows: %w", err)
			}

			// Find the flow
			var flow *flows.Flow
			for _, f := range discovered {
				if f.Name == name {
					flow = &f
					break
				}
			}

			if flow == nil {
				return fmt.Errorf("flow not found: %s", name)
			}

			// Prevent removing built-in flows
			if flow.Source == flows.FlowSourceBuiltin {
				return fmt.Errorf("cannot remove built-in flow %s", name)
			}

			// Determine what to delete
			var toDelete string
			var isDir bool
			if flow.Dir != "" && flow.Dir != filepath.Dir(flow.Path) {
				// Packaged flow - delete the directory
				toDelete = flow.Dir
				isDir = true
			} else {
				// Simple flow - delete just the file
				toDelete = flow.Path
				isDir = false
			}

			// Dry run mode
			if dryRun {
				fmt.Println("Would remove:")
				fmt.Printf("  %s\n", toDelete)

				if isDir {
					entries, _ := os.ReadDir(toDelete)
					for _, e := range entries {
						fmt.Printf("    - %s\n", e.Name())
					}
				}
				return nil
			}

			// Confirmation prompt (unless --force)
			if !force {
				fmt.Printf("Remove flow %s?\n", name)
				fmt.Printf("  Source: %s\n", flow.Source)
				fmt.Printf("  Location: %s\n", toDelete)

				if isDir {
					entries, _ := os.ReadDir(toDelete)
					if len(entries) > 0 {
						fmt.Println("  Contents:")
						for _, e := range entries {
							fmt.Printf("    - %s\n", e.Name())
						}
					}
				}

				fmt.Print("\nType the flow name to confirm: ")
				var confirm string
				fmt.Scanln(&confirm)

				if confirm != name {
					return fmt.Errorf("confirmation failed: expected %s, got %s", name, confirm)
				}
			}

			// Remove the flow
			var removeErr error
			if isDir {
				removeErr = os.RemoveAll(toDelete)
			} else {
				removeErr = os.Remove(toDelete)
			}
			if removeErr != nil {
				return fmt.Errorf("remove flow: %w", removeErr)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render("✓ Removed flow: " + name))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be deleted without removing")

	return cmd
}

func statsFlowsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "stats [flow-name]",
		Short: "Show flow execution statistics",
		Long: `Show statistics for flow executions.

Without arguments, shows statistics for all flows.
With a flow name, shows detailed statistics for that flow.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			history := flows.NewHistoryService(queries)

			if len(args) == 1 {
				// Stats for specific flow
				flowName := args[0]
				stats, err := history.GetStats(cmd.Context(), flowName)
				if err != nil {
					return fmt.Errorf("get stats: %w", err)
				}

				if jsonOutput {
					return outputStatsJSON(stats)
				}

				return outputFlowStats(stats)
			}

			// Stats for all flows
			allStats, err := history.GetAllFlowStats(cmd.Context())
			if err != nil {
				return fmt.Errorf("get all stats: %w", err)
			}

			if len(allStats) == 0 {
				fmt.Println("No flow runs found.")
				return nil
			}

			if jsonOutput {
				return outputAllStatsJSON(allStats)
			}

			return outputAllFlowStats(allStats)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func outputStatsJSON(stats *flows.FlowStats) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(stats)
}

func outputAllStatsJSON(allStats []*flows.FlowStats) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(allStats)
}

func outputFlowStats(stats *flows.FlowStats) error {
	cyan := lipgloss.Color("#67e8f9")
	green := lipgloss.Color("#34d399")
	red := lipgloss.Color("#ef4444")
	muted := lipgloss.Color("#6b7280")

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(cyan)
	labelStyle := lipgloss.NewStyle().Foreground(muted)
	successStyle := lipgloss.NewStyle().Foreground(green)
	failStyle := lipgloss.NewStyle().Foreground(red)

	fmt.Println(titleStyle.Render("Flow: " + stats.FlowName))
	fmt.Println()

	fmt.Printf("%s %d\n", labelStyle.Render("Total Runs:"), stats.TotalRuns)

	if stats.TotalRuns > 0 {
		rateStr := fmt.Sprintf("%.1f%% (%d/%d)", stats.SuccessRate, stats.SuccessCount, stats.TotalRuns)
		if stats.SuccessRate >= 90 {
			fmt.Printf("%s %s\n", labelStyle.Render("Success Rate:"), successStyle.Render(rateStr))
		} else if stats.SuccessRate >= 70 {
			fmt.Printf("%s %s\n", labelStyle.Render("Success Rate:"), rateStr)
		} else {
			fmt.Printf("%s %s\n", labelStyle.Render("Success Rate:"), failStyle.Render(rateStr))
		}

		if stats.AvgDuration > 0 {
			fmt.Printf("%s %s\n", labelStyle.Render("Avg Duration:"), formatDuration(stats.AvgDuration))
		}

		if stats.LastRun != nil {
			statusStr := string(stats.LastRun.Status)
			if stats.LastRun.Status == flows.RunStatusSuccess {
				statusStr = successStyle.Render("success")
			} else if stats.LastRun.Status == flows.RunStatusFailed {
				statusStr = failStyle.Render("failed")
			}
			fmt.Printf("%s %s (%s)\n", labelStyle.Render("Last Run:"), stats.LastRun.StartedAt.Format("2006-01-02 15:04"), statusStr)
		}
	}

	return nil
}

func outputAllFlowStats(allStats []*flows.FlowStats) error {
	cyan := lipgloss.Color("#67e8f9")
	muted := lipgloss.Color("#6b7280")
	green := lipgloss.Color("#34d399")
	red := lipgloss.Color("#ef4444")

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(cyan)
	mutedStyle := lipgloss.NewStyle().Foreground(muted)
	successStyle := lipgloss.NewStyle().Foreground(green)
	failStyle := lipgloss.NewStyle().Foreground(red)

	// Calculate column widths
	maxNameLen := 4 // "NAME"
	for _, s := range allStats {
		if len(s.FlowName) > maxNameLen {
			maxNameLen = len(s.FlowName)
		}
	}

	// Header
	fmt.Printf("%s  %s  %s  %s\n",
		headerStyle.Render(padRight("FLOW", maxNameLen)),
		headerStyle.Render("RUNS"),
		headerStyle.Render("SUCCESS"),
		headerStyle.Render("AVG TIME"),
	)

	// Rows
	for _, s := range allStats {
		rateStr := fmt.Sprintf("%5.1f%%", s.SuccessRate)
		if s.SuccessRate >= 90 {
			rateStr = successStyle.Render(rateStr)
		} else if s.SuccessRate < 70 {
			rateStr = failStyle.Render(rateStr)
		}

		durationStr := "-"
		if s.AvgDuration > 0 {
			durationStr = formatDuration(s.AvgDuration)
		}

		fmt.Printf("%s  %s  %s  %s\n",
			padRight(s.FlowName, maxNameLen),
			mutedStyle.Render(fmt.Sprintf("%4d", s.TotalRuns)),
			rateStr,
			mutedStyle.Render(durationStr),
		)
	}

	return nil
}
