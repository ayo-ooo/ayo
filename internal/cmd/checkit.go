package cmd

import (
	"fmt"
	"os"

	"github.com/ayo-ooo/ayo/internal/project"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

var checkitCmd = &cobra.Command{
	Use:   "checkit [path]",
	Short: "Validate an agent project",
	Long: `Validate the structure and contents of an agent project.

Checks:
  - config.toml exists and is valid
  - system.md exists and is non-empty
  - input.jsonschema is valid JSON Schema (if present)
  - output.jsonschema is valid JSON Schema (if present)
  - skills/ contains valid skill packages (if present)
  - hooks/ contains valid hook executables (if present)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		if err := validateProject(path); err != nil {
			exitError(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(checkitCmd)

	// Hidden aliases
	checkAlias := &cobra.Command{
		Use:    "check [path]",
		Hidden: true,
		Run:    checkitCmd.Run,
	}
	rootCmd.AddCommand(checkAlias)

	validateAlias := &cobra.Command{
		Use:    "validate [path]",
		Hidden: true,
		Run:    checkitCmd.Run,
	}
	rootCmd.AddCommand(validateAlias)
}

func validateProject(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access path '%s': %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path '%s' is not a directory", path)
	}

	proj, err := project.ParseProject(path)
	if err != nil {
		return fmt.Errorf("parsing project: %w", err)
	}

	errors := project.ValidateProject(proj)
	if len(errors) > 0 {
		fmt.Fprintln(os.Stderr, warnStyle.Render("Validation failed:\n"))
		for _, e := range errors {
			fmt.Fprintf(os.Stderr, "  %s %s\n", fileStyle.Render(e.File), e.Message)
		}
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Project '%s' is valid", proj.Config.Name))
	
	fmt.Printf("\nProject details:\n")
	fmt.Printf("  Name:        %s\n", proj.Config.Name)
	fmt.Printf("  Version:     %s\n", proj.Config.Version)
	fmt.Printf("  Description: %s\n", proj.Config.Description)
	
	if proj.Prompt != nil {
		fmt.Printf("  Prompt:      prompt.tmpl (%d bytes)\n", len(*proj.Prompt))
	}
	if proj.Input != nil {
		fmt.Printf("  Input:       input.jsonschema\n")
	}
	if proj.Output != nil {
		fmt.Printf("  Output:      output.jsonschema\n")
	}
	if len(proj.Skills) > 0 {
		fmt.Printf("  Skills:      %d\n", len(proj.Skills))
	}
	if len(proj.Hooks) > 0 {
		fmt.Printf("  Hooks:       %d\n", len(proj.Hooks))
	}

	return nil
}
