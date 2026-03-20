package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))
)

var rootCmd = &cobra.Command{
	Use:   "ayo",
	Short: "Compile AI agent definitions into standalone executables",
	Long: headerStyle.Render("Ayo") + `

Ayo compiles AI agent definitions into standalone, dependency-free CLI executables.

Agents are defined by a directory convention containing:
  - config.toml (required): metadata, model requirements, defaults
  - system.md (required): system message governing agent behavior
  - prompt.tmpl (optional): Go template for rendering prompts
  - input.jsonschema (optional): CLI interface definition
  - output.jsonschema (optional): structured output schema
  - skills/ (optional): Agent Skills compatible packages
  - hooks/ (optional): lifecycle hook executables

Examples:
  ayo fresh my-agent      Create a new agent project
  ayo checkit ./my-agent  Validate an agent project
  ayo runthat ./my-agent  Compile to standalone executable`,
	Version: Version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}

func printError(msg string) {
	fmt.Fprintln(os.Stderr, errorStyle.Render("Error: ")+msg)
}

func printSuccess(msg string) {
	fmt.Fprintln(os.Stdout, successStyle.Render("✓ ")+msg)
}

func exitError(msg string) {
	printError(msg)
	os.Exit(1)
}
