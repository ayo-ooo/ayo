package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/build"
	"github.com/alexcabrera/ayo/internal/build/types"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/paths"
)

func newCheckitCmd() *cobra.Command {
	var verbose bool
	var evalsThreshold float64
	var evalsOnly bool

	cmd := &cobra.Command{
		Use:   "checkit [directory]",
		Short: "Validate agent or team configuration",
		Long: `Validate the configuration file (config.toml or team.toml) in a directory.

This checks:
- TOML syntax validity
- Required fields presence
- JSON Schema syntax for input/output schemas
- CLI flag configuration
- Data type consistency

Exit codes:
  0 - Validation passed
  1 - Validation failed
  2 - Error reading configuration

Examples:
  ayo checkit ./myagent
  ayo checkit --verbose ./myagent`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheckit(args[0], verbose, evalsThreshold, evalsOnly)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed validation output")
	cmd.Flags().Float64Var(&evalsThreshold, "evals-threshold", 7.0, "Score threshold for passing evals (default: 7.0)")
	cmd.Flags().BoolVar(&evalsOnly, "evals-only", false, "Only run evals, skip other validation")

	return cmd
}

func runCheckit(dir string, verbose bool, evalsThreshold float64, evalsOnly bool) error {
	// Resolve to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve directory: %w", err)
	}

	// Load build configuration
	buildConfig, configPath, err := build.LoadConfigFromDir(absDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(2)
	}

	// Load main ayo config for evals (needs LLM provider)
	mainConfig, err := config.Load(paths.ConfigFile())
	if err != nil {
		return fmt.Errorf("load main config: %w", err)
	}

	// Check if evals are enabled
	if buildConfig.Evals.Enabled && !evalsOnly {
		fmt.Printf("✓ Configuration loaded: %s\n", configPath)
		fmt.Printf("✓ Evals enabled in configuration\n")
	}

	// Run evals if enabled or requested
	if buildConfig.Evals.Enabled || evalsOnly {
		if err := runEvals(absDir, buildConfig, mainConfig, evalsThreshold, verbose); err != nil {
			fmt.Fprintf(os.Stderr, "Error running evals: %v\n", err)
			os.Exit(1)
		}
		if evalsOnly {
			return nil
		}
	}

	fmt.Printf("✓ Configuration loaded: %s\n", configPath)

	// Display configuration summary
	if verbose {
		fmt.Printf("\nAgent Information:\n")
		fmt.Printf("  Name: %s\n", buildConfig.Agent.Name)
		fmt.Printf("  Description: %s\n", buildConfig.Agent.Description)
		fmt.Printf("  Model: %s\n", buildConfig.Agent.Model)

		fmt.Printf("\nCLI Configuration:\n")
		fmt.Printf("  Mode: %s\n", buildConfig.CLI.Mode)
		fmt.Printf("  Description: %s\n", buildConfig.CLI.Description)
		fmt.Printf("  Flags: %d defined\n", len(buildConfig.CLI.Flags))

		if len(buildConfig.CLI.Flags) > 0 {
			fmt.Printf("\n  Defined Flags:\n")
			for name, flag := range buildConfig.CLI.Flags {
				posInfo := ""
				if flag.Position >= 0 {
					posInfo = fmt.Sprintf(" (position %d)", flag.Position)
				}
				fmt.Printf("    - %s: %s%s\n", name, flag.Type, posInfo)
			}
		}

		fmt.Printf("\nTools Configuration:\n")
		if len(buildConfig.Agent.Tools.Allowed) > 0 {
			for _, tool := range buildConfig.Agent.Tools.Allowed {
				fmt.Printf("  - %s\n", tool)
			}
		} else {
			fmt.Printf("  No tools configured\n")
		}

		fmt.Printf("\nMemory Configuration:\n")
		fmt.Printf("  Enabled: %v\n", buildConfig.Agent.Memory.Enabled)
		fmt.Printf("  Scope: %s\n", buildConfig.Agent.Memory.Scope)

		fmt.Printf("\nSandbox Configuration:\n")
		fmt.Printf("  Network: %v\n", buildConfig.Agent.Sandbox.Network)
		fmt.Printf("  Host Path: %s\n", buildConfig.Agent.Sandbox.HostPath)

		if buildConfig.Input.Schema != nil {
			fmt.Printf("\nInput Schema: ✓ Present\n")
		} else {
			fmt.Printf("\nInput Schema: (not defined)\n")
		}

		if buildConfig.Output.Schema != nil {
			fmt.Printf("Output Schema: ✓ Present\n")
		} else {
			fmt.Printf("Output Schema: (not defined)\n")
		}

		if buildConfig.Triggers.Watch != nil || buildConfig.Triggers.Schedule != "" || buildConfig.Triggers.Events != nil {
			fmt.Printf("\nTriggers Configuration:\n")
			if len(buildConfig.Triggers.Watch) > 0 {
				fmt.Printf("  Watch paths: %d\n", len(buildConfig.Triggers.Watch))
			}
			if buildConfig.Triggers.Schedule != "" {
				fmt.Printf("  Schedule: %s\n", buildConfig.Triggers.Schedule)
			}
			if len(buildConfig.Triggers.Events) > 0 {
				fmt.Printf("  Events: %d\n", len(buildConfig.Triggers.Events))
			}
		}
	}

	// Validate directory structure
	fmt.Printf("\n✓ Checking directory structure...\n")

	dirsToCheck := []string{
		"skills",
		"tools",
		"prompts",
	}

	for _, dirName := range dirsToCheck {
		dirPath := filepath.Join(absDir, dirName)
		if info, err := os.Stat(dirPath); err == nil {
			if info.IsDir() {
				if verbose {
					fmt.Printf("  ✓ %s/ exists\n", dirName)
				}
			}
		}
	}

	// Check for system.md in prompts
	systemPath := filepath.Join(absDir, "prompts", "system.md")
	if _, err := os.Stat(systemPath); err == nil {
		if verbose {
			fmt.Printf("  ✓ prompts/system.md exists\n")
		}
	}

	// Success message
	fmt.Printf("\n✓ Configuration is valid!\n")
	fmt.Printf("\nYou can now build this agent:\n")
	fmt.Printf("  ayo build %s\n", dir)

	return nil
}

// runEvals executes the evaluation suite
func runEvals(dir string, buildConfig *types.Config, mainConfig config.Config, threshold float64, verbose bool) error {
	// Check if evals file exists
	evalsPath := filepath.Join(dir, buildConfig.Evals.File)
	if _, err := os.Stat(evalsPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: evals file not found: %s\n", evalsPath)
		return fmt.Errorf("evals file not found")
	}

	fmt.Printf("\n⚙ Running evaluations...\n\n")

	// Create judge
	judge, err := build.NewJudge(mainConfig.Provider, buildConfig.Evals.JudgeModel, buildConfig.Evals.Criteria)
	if err != nil {
		return fmt.Errorf("create judge: %w", err)
	}

	// Create eval runner
	runner := build.NewEvalRunner(buildConfig)

	// Run all evals
	results, err := runner.RunAllEvals(evalsPath)
	if err != nil {
		return fmt.Errorf("run evals: %w", err)
	}

	// Judge each result
	ctx := context.Background()
	passedCount := 0
	totalScore := 0.0

	for i, result := range results {
		if result.Error != nil {
			fmt.Printf("Test %d: Error executing eval\n", i+1)
			fmt.Printf("  Error: %v\n", result.Error)
			continue
		}

		// Judge the output
		judgeResult, err := judge.Compare(ctx, result.Case.Input, result.Case.Expected, result.Actual, result.Case.Criteria)
		if err != nil {
			fmt.Printf("Test %d: Error judging output\n", i+1)
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		// Update result with judge info
		result.Score = judgeResult.Score
		result.Reasoning = judgeResult.Reasoning
		result.Passed = judgeResult.Score >= threshold
		if result.Passed {
			passedCount++
		}
		totalScore += judgeResult.Score

		// Display result
		description := result.Case.Description
		if description == "" {
			description = fmt.Sprintf("Test %d", i+1)
		}

		status := "✗"
		if result.Passed {
			status = "✓"
		}
		fmt.Printf("%s %s: Score %.1f/10\n", status, description, judgeResult.Score)

		if verbose {
			fmt.Printf("  Input: %s\n", result.ActualJSON)
			fmt.Printf("  Reasoning: %s\n", judgeResult.Reasoning)
		}
	}

	// Print summary
	numResults := len(results)
	if numResults == 0 {
		fmt.Printf("\nNo test cases found in evals.csv\n")
		return nil
	}

	avgScore := totalScore / float64(numResults)
	percentage := float64(passedCount) / float64(numResults) * 100


	fmt.Printf("\nSummary: %d/%d passed (%.0f%%)\n", passedCount, numResults, percentage)
	fmt.Printf("Average score: %.1f/10\n", avgScore)

	if passedCount != numResults {
		return fmt.Errorf("%d/%d tests failed", numResults-passedCount, numResults)
	}

	return nil
}
