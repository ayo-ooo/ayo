package cmd

import (
	"fmt"
	"os"

	"github.com/ayo-ooo/ayo/internal/project"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

var (
	checkitWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))

	checkitFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9B9B9B"))

	checkitLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).Bold(true)

	checkitValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCDCAA"))
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
  - input_order references exist in schema properties
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

	fmt.Printf("Validating %s...\n", checkitFileStyle.Render(proj.Config.Name))

	errors := project.ValidateProject(proj)

	// Separate errors and warnings
	var errs, warns []*project.ValidationError
	for _, e := range errors {
		if isWarning(e.Message) {
			warns = append(warns, e)
		} else {
			errs = append(errs, e)
		}
	}

	// Print validation results
	for _, e := range errs {
		fmt.Printf("  %s %s: %s\n",
			errorStyle.Render("[X]"),
			checkitFileStyle.Render(e.File),
			e.Message)
	}

	for _, e := range warns {
		fmt.Printf("  %s %s: %s\n",
			checkitWarnStyle.Render("[!]"),
			checkitFileStyle.Render(e.File),
			e.Message)
	}

	if len(errs) > 0 {
		fmt.Printf("\n%s\n", errorStyle.Render("[X] Validation failed"))
		os.Exit(1)
	}

	fmt.Printf("\n%s\n\n", successStyle.Render("[+] Valid"))

	// Print project details
	fmt.Printf("Project details:\n")
	fmt.Printf("  %s  %s\n", checkitLabelStyle.Render("Name:"), checkitValueStyle.Render(proj.Config.Name))
	fmt.Printf("  %s   %s\n", checkitLabelStyle.Render("Version:"), checkitValueStyle.Render(proj.Config.Version))
	fmt.Printf("  %s  %s\n", checkitLabelStyle.Render("Description:"), proj.Config.Description)

	if proj.Prompt != nil {
		fmt.Printf("  %s     %s\n", checkitLabelStyle.Render("Prompt:"), checkitFileStyle.Render(fmt.Sprintf("prompt.tmpl (%d bytes)", len(*proj.Prompt))))
	}
	if proj.Input != nil {
		fmt.Printf("  %s       %s\n", checkitLabelStyle.Render("Input:"), checkitFileStyle.Render("input.jsonschema"))
	}
	if proj.Output != nil {
		fmt.Printf("  %s      %s\n", checkitLabelStyle.Render("Output:"), checkitFileStyle.Render("output.jsonschema"))
	}
	if len(proj.Skills) > 0 {
		fmt.Printf("  %s      %s\n", checkitLabelStyle.Render("Skills:"), checkitValueStyle.Render(fmt.Sprintf("%d", len(proj.Skills))))
	}
	if len(proj.Hooks) > 0 {
		fmt.Printf("  %s       %s\n", checkitLabelStyle.Render("Hooks:"), checkitValueStyle.Render(fmt.Sprintf("%d", len(proj.Hooks))))
	}

	return nil
}

func isWarning(msg string) bool {
	// Messages containing "not in input_order" are warnings
	return contains(msg, "not in input_order") ||
		contains(msg, "will appear last")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
