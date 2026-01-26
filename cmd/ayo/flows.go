package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/flows"
	"github.com/alexcabrera/ayo/internal/paths"
)

func newFlowsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "flows",
		Short:   "Manage flows",
		Aliases: []string{"flow"},
		Long: `Manage flows - composable agent pipelines.

Flows are shell scripts with structured frontmatter that compose agents
into pipelines. They are the unit of work that external orchestrators invoke.

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
	cmd.AddCommand(runFlowCmd(cfgPath))
	cmd.AddCommand(validateFlowCmd())
	cmd.AddCommand(newFlowCmd())
	cmd.AddCommand(historyFlowsCmd(cfgPath))
	cmd.AddCommand(replayFlowCmd(cfgPath))

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
	output := map[string]interface{}{
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
		maxLines := 10
		if len(lines) < maxLines {
			maxLines = len(lines)
		}
		for i := 0; i < maxLines; i++ {
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

func runFlowCmd(cfgPath *string) *cobra.Command {
	var inputFile string
	var timeout int
	var validate bool
	var noHistory bool

	cmd := &cobra.Command{
		Use:   "run <name> [input]",
		Short: "Execute a flow",
		Long: `Execute a flow and return JSON output.

Input can be provided as:
  - Second argument: ayo flows run myflow '{"key": "value"}'
  - Stdin: echo '{"key": "value"}' | ayo flows run myflow
  - File: ayo flows run myflow --input data.json

Output:
  - Stdout: JSON result from the flow
  - Stderr: Logs and progress (streamed in real-time)

Exit codes:
  0 - Success
  1 - General error
  2 - Input validation failed
  3 - Flow execution failed
  124 - Timeout`,
		Args: cobra.RangeArgs(1, 2),
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

			// Build run options
			opts := flows.RunOptions{
				Timeout:  time.Duration(timeout) * time.Second,
				Validate: validate,
			}

			// Input from argument
			if len(args) > 1 {
				opts.Input = args[1]
			}

			// Input from file
			if inputFile != "" {
				opts.InputFile = inputFile
			}

			// Check if stdin has data (only if no other input provided)
			if opts.Input == "" && opts.InputFile == "" && !isTerminal(os.Stdin) {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				opts.Input = string(data)
			}

			// Setup history recording if not disabled
			if !noHistory && !validate {
				cfg, err := config.Load(*cfgPath)
				if err == nil {
					_, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
					if err == nil {
						opts.History = flows.NewHistoryService(queries)
						opts.AutoPrune = true
						opts.RetentionDays = cfg.Flows.HistoryRetentionDays
						opts.MaxRuns = int64(cfg.Flows.HistoryMaxRuns)
					}
				}
			}

			// Run the flow with stderr streaming
			result, err := flows.RunStreaming(cmd.Context(), flow, opts, os.Stderr)
			if err != nil {
				return err
			}

			// Handle result
			if result.Error != nil {
				fmt.Fprintln(os.Stderr, result.Error)
			}

			// Output stdout (JSON)
			if result.Stdout != "" {
				fmt.Print(result.Stdout)
			}

			// Exit with appropriate code
			switch result.Status {
			case flows.RunStatusSuccess:
				return nil
			case flows.RunStatusValidationFailed:
				os.Exit(2)
			case flows.RunStatusTimeout:
				os.Exit(124)
			default:
				if result.ExitCode != 0 {
					os.Exit(result.ExitCode)
				}
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file path")
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 300, "Timeout in seconds (default 5 minutes)")
	cmd.Flags().BoolVar(&validate, "validate", false, "Validate input only, don't run")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "Don't record run in history")

	return cmd
}

func validateFlowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <path>",
		Short: "Validate a flow file or directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Try to load the flow
			flow, err := flows.DiscoverOne(path)
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

// isTerminal checks if a file descriptor is a terminal
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
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

func historyShowCmd(cfgPath *string) *cobra.Command {
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
	var obj interface{}
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return "  " + s
	}
	formatted, err := json.MarshalIndent(obj, "  ", "  ")
	if err != nil {
		return "  " + s
	}
	return "  " + string(formatted)
}

func replayFlowCmd(cfgPath *string) *cobra.Command {
	var timeout int
	var noHistory bool

	cmd := &cobra.Command{
		Use:   "replay <run-id>",
		Short: "Replay a flow run with its original input",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := args[0]

			_, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			history := flows.NewHistoryService(queries)

			// Get the original run
			run, err := history.GetRun(cmd.Context(), runID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("run not found: %s", runID)
				}
				return fmt.Errorf("get run: %w", err)
			}

			// Discover flows to find the current version
			dirs := paths.FlowsDirs()
			discovered, err := flows.Discover(dirs)
			if err != nil {
				return fmt.Errorf("discover flows: %w", err)
			}

			// Find the flow
			var flow *flows.Flow
			for _, f := range discovered {
				if f.Name == run.FlowName {
					flow = &f
					break
				}
			}

			if flow == nil {
				return fmt.Errorf("flow no longer exists: %s", run.FlowName)
			}

			// Build run options
			opts := flows.RunOptions{
				Input:   run.InputJSON,
				Timeout: time.Duration(timeout) * time.Second,
			}

			// Setup history recording if not disabled
			if !noHistory {
				cfg, err := config.Load(*cfgPath)
				if err == nil {
					opts.History = history
					opts.AutoPrune = true
					opts.RetentionDays = cfg.Flows.HistoryRetentionDays
					opts.MaxRuns = int64(cfg.Flows.HistoryMaxRuns)
					opts.ParentRunID = run.ID // Link to original run
				}
			}

			// Run the flow with stderr streaming
			result, err := flows.RunStreaming(cmd.Context(), flow, opts, os.Stderr)
			if err != nil {
				return err
			}

			// Handle result
			if result.Error != nil {
				fmt.Fprintln(os.Stderr, result.Error)
			}

			// Output stdout (JSON)
			if result.Stdout != "" {
				fmt.Print(result.Stdout)
			}

			// Exit with appropriate code
			switch result.Status {
			case flows.RunStatusSuccess:
				return nil
			case flows.RunStatusValidationFailed:
				os.Exit(2)
			case flows.RunStatusTimeout:
				os.Exit(124)
			default:
				if result.ExitCode != 0 {
					os.Exit(result.ExitCode)
				}
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&timeout, "timeout", "t", 300, "Timeout in seconds (default 5 minutes)")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "Don't record replay in history")

	return cmd
}
