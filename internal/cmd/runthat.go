package cmd

import (
	"fmt"
	"strings"

	"github.com/ayo-ooo/ayo/internal/build"
	"github.com/ayo-ooo/ayo/internal/project"
	"github.com/spf13/cobra"
)

var (
	buildOutputPath string
	buildPlatform   string
	buildRegister   bool
)

var runthatCmd = &cobra.Command{
	Use:   "runthat [path]",
	Short: "Compile an agent project into a standalone executable",
	Long: `Run that — compile an agent project into a standalone executable binary.

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
	runthatCmd.Flags().StringVar(&buildPlatform, "platform", "", "Target platform (e.g. linux/amd64, darwin/arm64, or 'all')")
	runthatCmd.Flags().BoolVar(&buildRegister, "register", false, "Register agent in the ayo registry after building")

	// Hidden aliases
	aliases := []string{"drop", "build", "compile", "generate", "dunn"}
	for _, name := range aliases {
		alias := &cobra.Command{
			Use:    name + " [path]",
			Hidden: true,
			Run:    runthatCmd.Run,
		}
		alias.Flags().StringVarP(&buildOutputPath, "output", "o", "", "Output binary path")
		alias.Flags().StringVar(&buildPlatform, "platform", "", "Target platform (e.g. linux/amd64, darwin/arm64, or 'all')")
		alias.Flags().BoolVar(&buildRegister, "register", false, "Register agent after building")
		rootCmd.AddCommand(alias)
	}
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

	if buildPlatform != "" {
		platforms, err := parsePlatformFlag(buildPlatform)
		if err != nil {
			return err
		}

		outputBase := buildOutputPath
		if outputBase == "" {
			outputBase = proj.Config.Name
		}

		if err := manager.BuildCross(proj, outputBase, platforms); err != nil {
			return fmt.Errorf("building: %w", err)
		}

		for _, p := range platforms {
			printSuccess(fmt.Sprintf("Built executable: %s%s", outputBase, p.Suffix()))
		}
		return nil
	}

	if err := manager.Build(proj, buildOutputPath); err != nil {
		return fmt.Errorf("building: %w", err)
	}

	output := buildOutputPath
	if output == "" {
		output = proj.Config.Name
	}

	printSuccess(fmt.Sprintf("Built executable: %s", output))

	if buildRegister {
		if err := RegisterFromBuild(proj, output); err != nil {
			return fmt.Errorf("registering: %w", err)
		}
		printSuccess(fmt.Sprintf("Registered '%s' in ayo registry", proj.Config.Name))
	}

	return nil
}

func parsePlatformFlag(value string) ([]build.Platform, error) {
	if strings.EqualFold(value, "all") {
		return build.AllPlatforms(), nil
	}

	// Support comma-separated platforms, e.g. "linux/amd64,darwin/arm64"
	parts := strings.Split(value, ",")
	platforms := make([]build.Platform, 0, len(parts))
	for _, part := range parts {
		p, err := build.ParsePlatform(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		platforms = append(platforms, p)
	}
	return platforms, nil
}
