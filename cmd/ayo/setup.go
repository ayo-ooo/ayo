package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/ollama"
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

				// Phase 4: Provider and model detection
				sui.Blank()
				sui.Header("Checking provider credentials...")
				sui.Blank()

				providers := config.DetectProviders()
				hasProvider := false
				for _, p := range providers {
					if p.HasKey {
						sui.SuccessPath(p.Name, "configured")
						hasProvider = true
					}
				}

				// Check Ollama status (always, not just when no cloud provider)
				ctx := cmd.Context()
				ollamaStatus := ollama.GetStatus(ctx)

				if ollamaStatus == ollama.StatusRunning {
					sui.SuccessPath("Ollama", "running locally")
				}

				// Skip interactive prompts when --force is set
				if forceOverwrite {
					if !hasProvider {
						sui.Warning("No cloud provider credentials detected.")
						sui.Info("Run 'ayo setup' without --force to configure providers interactively.")
					}
				} else if !hasProvider {
					sui.Warning("No cloud provider credentials detected.")
					sui.Blank()

					switch ollamaStatus {
					case ollama.StatusRunning:
						// Check for capable models
						ollamaClient := ollama.NewClient()
						capable, err := ollamaClient.ListCapableModels(ctx)
						if err != nil {
							sui.Error(fmt.Sprintf("Failed to list Ollama models: %v", err))
						} else if len(capable) > 0 {
							sui.Info("Available local models:")
							for _, m := range capable {
								sui.Info(fmt.Sprintf("  - %s", m.Name))
							}
							sui.Blank()

							// Offer to set default model to Ollama
							if cfg.DefaultModel == "gpt-5.2" {
								if err := offerOllamaDefault(sui, &cfg, *cfgPath, capable); err != nil {
									return err
								}
							}
						} else {
							sui.Warning("No chat-capable models installed.")
							if err := offerOllamaInstall(ctx, sui); err != nil {
								return err
							}
						}
					case ollama.StatusInstalled:
						sui.Warning("Ollama is installed but not running.")
						sui.Info("Start it with: ollama serve")
					case ollama.StatusNotInstalled:
						sui.Warning("Ollama is not installed.")
						sui.Info("Install it from: https://ollama.ai")
					}
					sui.Blank()

					// Offer credential entry
					if err := offerCredentialEntry(sui, &cfg, *cfgPath); err != nil {
						return err
					}
				} else {
					// Has cloud provider - still offer model selection
					sui.Blank()
					if err := offerModelSelection(sui, &cfg, *cfgPath, providers, ollamaStatus == ollama.StatusRunning); err != nil {
						return err
					}
				}

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

// offerOllamaDefault offers to set an Ollama model as the default.
func offerOllamaDefault(sui *setupUI, cfg *config.Config, cfgPath string, capable []ollama.Model) error {
	var useOllama bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use Ollama for default model?").
				Description("No cloud API key detected. Use local Ollama model?").
				Value(&useOllama),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return err
	}

	if useOllama && len(capable) > 0 {
		// Build options from capable models
		var options []huh.Option[string]
		for _, m := range capable {
			options = append(options, huh.NewOption(m.Name, "ollama/"+m.Name))
		}

		var selectedModel string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select model").
					Description("Choose an Ollama model to use as default").
					Options(options...).
					Value(&selectedModel),
			),
		).WithTheme(huh.ThemeCharm())

		if err := form.Run(); err != nil {
			return err
		}

		if selectedModel != "" {
			cfg.DefaultModel = selectedModel
			if err := config.Save(cfgPath, *cfg); err != nil {
				return err
			}
			sui.SuccessPath("Default model set to", selectedModel)
		}
	}
	return nil
}

// offerOllamaInstall offers to install a recommended Ollama model.
func offerOllamaInstall(ctx context.Context, sui *setupUI) error {
	suggested := ollama.GetSuggestedModels()
	if len(suggested) == 0 {
		return nil
	}

	sui.Blank()
	sui.Info("Suggested models for first-time users:")
	for _, m := range suggested {
		sui.Info(fmt.Sprintf("  - %s: %s", m.Name, m.Description))
	}
	sui.Blank()

	var install bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Install a model now?").
				Description(fmt.Sprintf("Pull %s (~%.1f GB)?", suggested[0].Name, suggested[0].SizeGB)).
				Value(&install),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return err
	}

	if install {
		modelName := suggested[0].Name
		sui.Step(fmt.Sprintf("Pulling %s...", modelName))

		var pullErr error
		_ = spinner.New().
			Title(fmt.Sprintf("Downloading %s...", modelName)).
			Action(func() {
				ollamaClient := ollama.NewClient()
				pullErr = ollamaClient.PullModel(ctx, modelName, func(p ollama.PullProgress) {
					// Progress shown by spinner
				})
			}).
			Run()

		if pullErr != nil {
			sui.Error(fmt.Sprintf("Failed to pull model: %v", pullErr))
		} else {
			sui.SuccessPath("Installed", modelName)
		}
	}

	return nil
}

// offerCredentialEntry offers to enter a cloud provider API key.
func offerCredentialEntry(sui *setupUI, cfg *config.Config, cfgPath string) error {
	sui.Info("To use cloud providers, set environment variables or enter a key now.")
	sui.Blank()

	var enterKey bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enter a cloud provider API key?").
				Description("Store a key for Anthropic, OpenAI, or another provider").
				Value(&enterKey),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return err
	}

	if !enterKey {
		return nil
	}

	// Show provider selection
	providers := config.DetectProviders()
	var options []huh.Option[string]
	for _, p := range providers {
		if !p.HasKey {
			options = append(options, huh.NewOption(p.Name, p.ID))
		}
	}

	if len(options) == 0 {
		sui.Info("All known providers already have credentials.")
		return nil
	}

	var selectedProvider string
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select provider").
				Options(options...).
				Value(&selectedProvider),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return err
	}

	if selectedProvider == "" {
		return nil
	}

	// Get the provider info
	var providerInfo config.DetectedProvider
	for _, p := range providers {
		if p.ID == selectedProvider {
			providerInfo = p
			break
		}
	}

	var apiKey string
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("%s API Key", providerInfo.Name)).
				Description(fmt.Sprintf("Will be stored in ~/.config/ayo/credentials.json")).
				EchoMode(huh.EchoModePassword).
				Value(&apiKey),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return err
	}

	if strings.TrimSpace(apiKey) == "" {
		sui.Cancelled("No key entered.")
		return nil
	}

	// Store the credential
	if err := config.StoreCredential(selectedProvider, apiKey); err != nil {
		return fmt.Errorf("store credential: %w", err)
	}

	// Inject immediately so it's available
	if err := config.InjectCredentials(); err != nil {
		return fmt.Errorf("inject credentials: %w", err)
	}

	sui.SuccessPath(providerInfo.Name+" API key", "stored")

	// Also update config default model if using a cloud provider for the first time
	if cfg.DefaultModel == "gpt-5.2" && selectedProvider != "openai" {
		// Suggest changing model based on provider
		var suggestedModel string
		switch selectedProvider {
		case "anthropic":
			suggestedModel = "claude-sonnet-4-20250514"
		case "google":
			suggestedModel = "gemini-2.0-flash"
		case "openrouter":
			suggestedModel = "anthropic/claude-sonnet-4"
		}

		if suggestedModel != "" {
			var changeModel bool
			form = huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Update default model?").
						Description(fmt.Sprintf("Change from gpt-5.2 to %s?", suggestedModel)).
						Value(&changeModel),
				),
			).WithTheme(huh.ThemeCharm())

			if err := form.Run(); err != nil {
				return err
			}

			if changeModel {
				cfg.DefaultModel = suggestedModel
				if err := config.Save(cfgPath, *cfg); err != nil {
					return err
				}
				sui.SuccessPath("Default model set to", suggestedModel)
			}
		}
	}

	return nil
}

// offerModelSelection offers to select a default model when cloud providers are available.
func offerModelSelection(sui *setupUI, cfg *config.Config, cfgPath string, providers []config.DetectedProvider, ollamaRunning bool) error {
	sui.Info(fmt.Sprintf("Current default model: %s", cfg.DefaultModel))
	sui.Blank()

	var changeModel bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Change default model?").
				Description("Select a different model for ayo to use").
				Value(&changeModel),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return err
	}

	if !changeModel {
		return nil
	}

	// Build model options from all configured providers using catwalk
	var options []huh.Option[string]

	allModels := config.AllConfiguredModels()
	for _, m := range allModels {
		label := m.Name
		if label == "" {
			label = m.ID
		}
		// Add provider suffix for clarity
		label = fmt.Sprintf("%s (%s)", label, m.Provider)
		options = append(options, huh.NewOption(label, m.ID))
	}

	// Add Ollama models if running
	if ollamaRunning {
		ollamaClient := ollama.NewClient()
		capable, err := ollamaClient.ListCapableModels(context.Background())
		if err == nil && len(capable) > 0 {
			for _, m := range capable {
				options = append(options, huh.NewOption(fmt.Sprintf("%s (Ollama)", m.Name), "ollama/"+m.Name))
			}
		}
	}

	if len(options) == 0 {
		sui.Warning("No models available to select.")
		return nil
	}

	var selectedModel string
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select default model").
				Description("Choose the model ayo will use by default").
				Options(options...).
				Height(15). // Show more options
				Value(&selectedModel),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return err
	}

	if selectedModel != "" && selectedModel != cfg.DefaultModel {
		cfg.DefaultModel = selectedModel
		if err := config.Save(cfgPath, *cfg); err != nil {
			return err
		}
		sui.SuccessPath("Default model set to", selectedModel)
	}

	return nil
}
