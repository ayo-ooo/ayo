package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/paths"
)

func newSetupCmd(cfgPath *string) *cobra.Command {
	var forceOverwrite bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Complete ayo setup (agents, skills)",
		Long:  "Runs complete ayo setup: installs built-in agents and skills, creates user directories.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				sui := newSetupUI(cmd.OutOrStdout())

				// Show mode
				if paths.IsDevMode() {
					sui.Header("Dev mode detected")
					sui.Info(fmt.Sprintf("  Repo: %s", paths.DevRoot()))
					sui.Blank()
				}

				// Phase 1: Detect ALL modifications upfront before any file changes
				sui.Header("Checking for local modifications...")
				sui.Blank()

				modifiedAgents, err := builtin.CheckModifiedAgents()
				if err != nil {
					return fmt.Errorf("check modified agents: %w", err)
				}

				modifiedSkills, err := builtin.CheckModifiedSkills()
				if err != nil {
					return fmt.Errorf("check modified skills: %w", err)
				}

				// Phase 2: Get user confirmation for all modifications before proceeding
				if len(modifiedAgents) > 0 || len(modifiedSkills) > 0 {
					if !forceOverwrite {
						sui.Warning("The following items have been modified locally:")
						sui.Blank()

						for _, m := range modifiedAgents {
							sui.Info(fmt.Sprintf("Agent %s:", m.Handle))
							for _, f := range m.ModifiedFiles {
								sui.Info(fmt.Sprintf("  • %s", f))
							}
						}
						for _, m := range modifiedSkills {
							sui.Info(fmt.Sprintf("Skill %s:", m.Name))
							for _, f := range m.ModifiedFiles {
								sui.Info(fmt.Sprintf("  • %s", f))
							}
						}
						sui.Blank()

						// Ask for confirmation
						var confirm bool
						form := huh.NewForm(
							huh.NewGroup(
								huh.NewConfirm().
									Title("Overwrite all modifications with fresh copies?").
									Description("Your changes will be lost.").
									Value(&confirm),
							),
						).WithTheme(huh.ThemeCharm())

						if err := form.Run(); err != nil {
							return err
						}

						if !confirm {
							sui.Cancelled("Setup cancelled.")
							sui.Blank()
							sui.Info("To keep your modifications, copy them to ~/.config/ayo/ first:")
							sui.Info(fmt.Sprintf("  Agents: %s", cfg.AgentsDir))
							sui.Info(fmt.Sprintf("  Skills: %s", cfg.SkillsDir))
							sui.Blank()
							sui.Info("User directories take priority over built-in ones.")
							return nil
						}
					}
				} else {
					sui.Info("No local modifications detected.")
				}
				sui.Blank()

				// Phase 3: Now safe to make file changes
				sui.Header("Setting up ayo...")
				sui.Blank()

				// 1. Install built-in agents and skills
				sui.Step("Installing built-in agents and skills...")
				installDir, err := builtin.ForceInstall()
				if err != nil {
					return fmt.Errorf("install builtins: %w", err)
				}
				sui.SuccessPath("Agents installed to", installDir)
				sui.SuccessPath("Skills installed to", builtin.SkillsInstallDir())
				sui.Blank()

				// 2. Create user directories
				sui.Step("Creating user directories...")
				userDirs := []struct {
					name string
					path string
				}{
					{"User agents", cfg.AgentsDir},
					{"User skills", cfg.SkillsDir},
					{"Prompts", paths.SystemPromptsDir()},
				}
				for _, d := range userDirs {
					if err := os.MkdirAll(d.path, 0o755); err != nil {
						sui.Error(fmt.Sprintf("Failed to create %s: %v", d.name, err))
					} else {
						sui.SuccessPath(d.name, d.path)
					}
				}
				sui.Blank()

				// 4. Summary
				sui.Header("Directory structure:")
				if paths.IsDevMode() {
					sui.Info(fmt.Sprintf("  Mode:            dev (%s)", paths.DevRoot()))
				}
				sui.Info(fmt.Sprintf("  User config:     %s", paths.ConfigDir()))
				sui.Info(fmt.Sprintf("  Built-in data:   %s", paths.DataDir()))
				sui.Blank()
				sui.Header("Load priority (first found wins):")
				sui.Info("  1. ./.config/ayo        (local project)")
				sui.Info("  2. ./.local/share/ayo   (local project data)")
				sui.Info("  3. ~/.config/ayo        (user config)")
				sui.Info("  4. ~/.local/share/ayo   (built-in data)")
				sui.Blank()
				sui.Header("Available commands:")
				sui.Info("  ayo                    Start chat with @ayo")
				sui.Info("  ayo agents list        List available agents")
				sui.Info("  ayo agents create      Create a new agent")
				sui.Info("  ayo skills list        List available skills")
				sui.Info("  ayo skills create      Create a new skill")
				sui.Blank()

				sui.Complete("Setup complete!")
				return nil
			})
		},
	}

	cmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "overwrite modifications without prompting")

	return cmd
}

// setupUI provides styled output for setup commands
type setupUI struct {
	out io.Writer
}

func newSetupUI(out io.Writer) *setupUI {
	return &setupUI{out: out}
}

func (s *setupUI) Header(msg string) {
	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	fmt.Fprintln(s.out, style.Render(msg))
}

func (s *setupUI) Step(msg string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	fmt.Fprintln(s.out, style.Render("→ "+msg))
}

func (s *setupUI) Info(msg string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	fmt.Fprintln(s.out, style.Render("  "+msg))
}

func (s *setupUI) SuccessPath(label, path string) {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	fmt.Fprintf(s.out, "  %s %s\n", labelStyle.Render("✓ "+label+":"), pathStyle.Render(path))
}

func (s *setupUI) Error(msg string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	fmt.Fprintln(s.out, style.Render("  ✗ "+msg))
}

func (s *setupUI) Code(code string) {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(0, 1)
	fmt.Fprintln(s.out, "  "+style.Render(code))
}

func (s *setupUI) Complete(msg string) {
	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	fmt.Fprintln(s.out, style.Render("✓ "+msg))
}

func (s *setupUI) Cancelled(msg string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	fmt.Fprintln(s.out, style.Render("⚠ "+msg))
}

func (s *setupUI) Warning(msg string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	fmt.Fprintln(s.out, style.Render("⚠ "+msg))
}

func (s *setupUI) Blank() {
	fmt.Fprintln(s.out)
}
