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

var runthatCmd = &cobra.Command{
	Use:   "runthat [path]",
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
	rootCmd.AddCommand(runthatCmd)
	runthatCmd.Flags().StringVarP(&buildOutputPath, "output", "o", "", "Output binary path (default: <agent-name>)")

	// Hidden aliases
	buildAlias := &cobra.Command{
		Use:    "build [path]",
		Hidden: true,
		Run:    runthatCmd.Run,
	}
	buildAlias.Flags().StringVarP(&buildOutputPath, "output", "o", "", "Output binary path")
	rootCmd.AddCommand(buildAlias)

	compileAlias := &cobra.Command{
		Use:    "compile [path]",
		Hidden: true,
		Run:    runthatCmd.Run,
	}
	compileAlias.Flags().StringVarP(&buildOutputPath, "output", "o", "", "Output binary path")
	rootCmd.AddCommand(compileAlias)

	generateAlias := &cobra.Command{
		Use:    "generate [path]",
		Hidden: true,
		Run:    runthatCmd.Run,
	}
	generateAlias.Flags().StringVarP(&buildOutputPath, "output", "o", "", "Output binary path")
	rootCmd.AddCommand(generateAlias)

	dunnAlias := &cobra.Command{
		Use:    "dunn [path]",
		Hidden: true,
		Run:    runthatCmd.Run,
	}
	dunnAlias.Flags().StringVarP(&buildOutputPath, "output", "o", "", "Output binary path")
	rootCmd.AddCommand(dunnAlias)
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
