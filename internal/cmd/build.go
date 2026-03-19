package cmd

import (
	"fmt"

	"github.com/charmbracelet/ayo/internal/build"
	"github.com/charmbracelet/ayo/internal/project"
	"github.com/spf13/cobra"
)

var (
	buildOutputPath string
)

var buildCmd = &cobra.Command{
	Use:   "build [path]",
	Short: "Compile an agent project into a standalone executable",
	Long: `Compile an agent project into a standalone executable binary.

The build process:
  1. Validates the project structure
  2. Generates Go source code from schemas and templates
  3. Embeds system.md, prompt.tmpl, skills, and hooks
  4. Compiles with static linking for portability`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		if err := buildProject(path); err != nil {
			exitError(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&buildOutputPath, "output", "o", "", "Output binary path (default: <agent-name>)")
}

func buildProject(path string) error {
	proj, err := project.ParseProject(path)
	if err != nil {
		return fmt.Errorf("parsing project: %w", err)
	}

	errors := project.ValidateProject(proj)
	if len(errors) > 0 {
		return fmt.Errorf("project validation failed: %s", errors[0].Message)
	}

	printSuccess(fmt.Sprintf("Validated project '%s'", proj.Config.Name))

	manager := build.NewManager()
	if err := manager.Build(proj, buildOutputPath); err != nil {
		return fmt.Errorf("building: %w", err)
	}

	output := buildOutputPath
	if output == "" {
		output = proj.Config.Name
	}

	printSuccess(fmt.Sprintf("Built executable: %s", output))
	return nil
}
