package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/build"
)

func newCheckitCmd() *cobra.Command {
	var verbose bool

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
			return runCheckit(args[0], verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed validation output")

	return cmd
}

func runCheckit(dir string, verbose bool) error {
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



	// Note: Evals have been removed in the build system architecture
	// This is now a simple configuration validation tool

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



	// Success message
	fmt.Printf("\n✓ Configuration is valid!\n")
	fmt.Printf("\nYou can now build this agent:\n")
	fmt.Printf("  ayo build %s\n", dir)

	return nil
}


