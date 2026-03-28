package cmd

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
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
	Short: "Agents You Orchestrate — build, register, and run AI agents",
	Long: headerStyle.Render("Ayo — Agents You Orchestrate") + `

Build AI agents from plain files, compile them into standalone binaries,
and orchestrate them from your terminal.

Build:
  ayo fresh my-agent              Scaffold a new agent project
  ayo checkit ./my-agent          Validate an agent project
  ayo runthat ./my-agent          Compile to standalone executable

Orchestrate:
  ayo runthat . --register        Build and register in one step
  ayo register ./my-agent         Register an agent or binary
  ayo list                        List all registered agents
  ayo describe my-agent           Show agent details and schemas
  ayo run my-agent [args]         Run a registered agent by name
  ayo remove my-agent             Remove from registry`,
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
