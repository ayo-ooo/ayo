package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/plugins"
)

// Plugin CLI color palette - adaptive for light/dark terminals
var (
	pluginPurple = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#a78bfa"}
	pluginCyan   = lipgloss.AdaptiveColor{Light: "#0891b2", Dark: "#67e8f9"}
	pluginGreen  = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#34d399"}
	pluginYellow = lipgloss.AdaptiveColor{Light: "#d97706", Dark: "#fbbf24"}
	pluginRed    = lipgloss.AdaptiveColor{Light: "#dc2626", Dark: "#f87171"}
	pluginMuted  = lipgloss.AdaptiveColor{Light: "#6b7280", Dark: "#9ca3af"}
	pluginText   = lipgloss.AdaptiveColor{Light: "#1f2937", Dark: "#f3f4f6"}
)

// Shared styles
var (
	pluginTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(pluginPurple)
	pluginNameStyle    = lipgloss.NewStyle().Bold(true).Foreground(pluginCyan)
	pluginVersionStyle = lipgloss.NewStyle().Foreground(pluginGreen)
	pluginMutedStyle   = lipgloss.NewStyle().Foreground(pluginMuted)
	pluginSuccessStyle = lipgloss.NewStyle().Foreground(pluginGreen)
	pluginWarnStyle    = lipgloss.NewStyle().Foreground(pluginYellow)
	pluginErrorStyle   = lipgloss.NewStyle().Foreground(pluginRed)
	pluginTextStyle    = lipgloss.NewStyle().Foreground(pluginText)

	pluginCheckmark = pluginSuccessStyle.Render("✓")
	pluginCross     = pluginErrorStyle.Render("✗")
	pluginArrow     = pluginMutedStyle.Render("→")
)

func newPluginsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugins",
		Short:   "Manage plugins",
		Aliases: []string{"plugin"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPluginsCmd(cfgPath).RunE(cmd, args)
		},
	}

	cmd.AddCommand(listPluginsCmd(cfgPath))
	cmd.AddCommand(installPluginCmd(cfgPath))
	cmd.AddCommand(showPluginCmd(cfgPath))
	cmd.AddCommand(updatePluginCmd(cfgPath))
	cmd.AddCommand(removePluginCmd(cfgPath))

	return cmd
}

func listPluginsCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry, err := plugins.LoadRegistry()
			if err != nil {
				return fmt.Errorf("load registry: %w", err)
			}

			installed := registry.List()
			if len(installed) == 0 {
				fmt.Println(pluginMutedStyle.Render("No plugins installed."))
				fmt.Println()
				fmt.Printf("%s Install a plugin:\n", pluginArrow)
				fmt.Printf("  %s\n", pluginTextStyle.Render("ayo plugins install <git-url>"))
				return nil
			}

			// Build table
			t := table.New().
				Border(lipgloss.RoundedBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(pluginMuted)).
				Headers("PLUGIN", "VERSION", "PROVIDES").
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == table.HeaderRow {
						return lipgloss.NewStyle().
							Foreground(pluginPurple).
							Bold(true).
							Padding(0, 1)
					}
					return lipgloss.NewStyle().
						Foreground(pluginText).
						Padding(0, 1)
				})

			for _, p := range installed {
				name := p.Name
				if p.Disabled {
					name += pluginErrorStyle.Render(" (disabled)")
				}

				version := "v" + p.Version

				// Build provides column
				var provides []string
				if len(p.Agents) > 0 {
					provides = append(provides, fmt.Sprintf("%d agent(s)", len(p.Agents)))
				}
				if len(p.Skills) > 0 {
					provides = append(provides, fmt.Sprintf("%d skill(s)", len(p.Skills)))
				}
				if len(p.Tools) > 0 {
					provides = append(provides, fmt.Sprintf("%d tool(s)", len(p.Tools)))
				}
				providesStr := strings.Join(provides, ", ")
				if providesStr == "" {
					providesStr = pluginMutedStyle.Render("-")
				}

				t.Row(name, version, providesStr)
			}

			fmt.Println(t.Render())
			return nil
		},
	}
}

func installPluginCmd(cfgPath *string) *cobra.Command {
	var force bool
	var local string
	var skipDeps bool

	cmd := &cobra.Command{
		Use:   "install <git-url>",
		Short: "Install a plugin from a git repository",
		Long: `Install a plugin from a git repository.

The plugin reference must be a full git URL (https:// or git@).

Examples:
  ayo plugins install https://github.com/alexcabrera/ayo-plugins-crush
  ayo plugins install git@gitlab.com:org/ayo-plugins-tools.git
  ayo plugins install --local ./my-plugin`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := &plugins.InstallOptions{
				Force:               force,
				SkipDependencyCheck: skipDeps,
			}

			var result *plugins.InstallResult
			var installErr error

			if local != "" {
				// Install from local directory - no spinner needed
				result, installErr = plugins.InstallFromLocal(local, opts)
			} else if len(args) == 0 {
				return fmt.Errorf("plugin reference required (or use --local)")
			} else {
				gitURL, name, parseErr := plugins.ParsePluginURL(args[0])
				if parseErr != nil {
					return fmt.Errorf("invalid plugin reference: %w", parseErr)
				}

				// Use spinner for git install
				spinnerErr := spinner.New().
					Title(fmt.Sprintf("Installing %s...", pluginNameStyle.Render(name))).
					Type(spinner.Dots).
					Style(lipgloss.NewStyle().Foreground(pluginPurple)).
					ActionWithErr(func(ctx context.Context) error {
						_ = gitURL // Used in message above
						result, installErr = plugins.Install(args[0], opts)
						return installErr
					}).
					Run()

				if spinnerErr != nil {
					return spinnerErr
				}
			}

			if installErr != nil {
				return installErr
			}

			// Success output
			fmt.Printf("%s Installed %s %s\n",
				pluginCheckmark,
				pluginNameStyle.Render(result.Plugin.Name),
				pluginVersionStyle.Render("v"+result.Manifest.Version),
			)

			// Show what was installed
			if len(result.Plugin.Agents) > 0 {
				fmt.Printf("  %s Agents: %s\n", pluginArrow, strings.Join(result.Plugin.Agents, ", "))
			}
			if len(result.Plugin.Skills) > 0 {
				fmt.Printf("  %s Skills: %s\n", pluginArrow, strings.Join(result.Plugin.Skills, ", "))
			}
			if len(result.Plugin.Tools) > 0 {
				fmt.Printf("  %s Tools: %s\n", pluginArrow, strings.Join(result.Plugin.Tools, ", "))
			}

			// Show warnings
			for _, warn := range result.Warnings {
				fmt.Printf("  %s %s\n", pluginWarnStyle.Render("!"), warn)
			}

			// Handle delegation setup if plugin declares delegates
			if len(result.Manifest.Delegates) > 0 {
				fmt.Println()
				if err := handleDelegateSetup(result.Manifest.Delegates); err != nil {
					// Don't fail install, just warn
					fmt.Printf("  %s Could not configure delegates: %v\n", pluginWarnStyle.Render("!"), err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing installation")
	cmd.Flags().StringVar(&local, "local", "", "Install from local directory")
	cmd.Flags().BoolVar(&skipDeps, "skip-deps", false, "Skip dependency checks")

	return cmd
}

func showPluginCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show details about an installed plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry, err := plugins.LoadRegistry()
			if err != nil {
				return fmt.Errorf("load registry: %w", err)
			}

			plugin, err := registry.Get(args[0])
			if err != nil {
				return err
			}

			labelStyle := pluginMutedStyle.Width(12)

			// Header
			fmt.Printf("%s %s\n",
				pluginNameStyle.Render(plugin.Name),
				pluginVersionStyle.Render("v"+plugin.Version),
			)
			fmt.Println()

			// Details
			fmt.Printf("%s %s\n", labelStyle.Render("Source"), pluginTextStyle.Render(plugin.GitURL))
			fmt.Printf("%s %s\n", labelStyle.Render("Path"), pluginTextStyle.Render(plugin.Path))
			fmt.Printf("%s %s\n", labelStyle.Render("Installed"), pluginTextStyle.Render(formatTimeAgo(plugin.InstalledAt.Unix())))

			if !plugin.UpdatedAt.IsZero() {
				fmt.Printf("%s %s\n", labelStyle.Render("Updated"), pluginTextStyle.Render(formatTimeAgo(plugin.UpdatedAt.Unix())))
			}

			// Agents
			if len(plugin.Agents) > 0 {
				fmt.Println()
				fmt.Println(pluginTitleStyle.Render("Agents"))
				for _, a := range plugin.Agents {
					fmt.Printf("  %s %s\n", pluginArrow, a)
				}
			}

			// Skills
			if len(plugin.Skills) > 0 {
				fmt.Println()
				fmt.Println(pluginTitleStyle.Render("Skills"))
				for _, s := range plugin.Skills {
					fmt.Printf("  %s %s\n", pluginArrow, s)
				}
			}

			// Tools
			if len(plugin.Tools) > 0 {
				fmt.Println()
				fmt.Println(pluginTitleStyle.Render("Tools"))
				for _, t := range plugin.Tools {
					fmt.Printf("  %s %s\n", pluginArrow, t)
				}
			}

			return nil
		},
	}
}

func updatePluginCmd(cfgPath *string) *cobra.Command {
	var force bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update installed plugins",
		Long: `Update one or all installed plugins.

If no name is provided, updates all plugins.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var results []*plugins.UpdateResult
			var updateErr error

			opts := &plugins.UpdateOptions{
				Force:  force,
				DryRun: dryRun,
			}

			title := "Checking for updates..."
			if len(args) == 1 {
				title = fmt.Sprintf("Updating %s...", pluginNameStyle.Render(args[0]))
			}

			spinnerErr := spinner.New().
				Title(title).
				Type(spinner.Dots).
				Style(lipgloss.NewStyle().Foreground(pluginPurple)).
				ActionWithErr(func(ctx context.Context) error {
					if len(args) == 1 {
						result, err := plugins.Update(args[0], opts)
						if err != nil {
							updateErr = err
							return err
						}
						results = []*plugins.UpdateResult{result}
					} else {
						var err error
						results, err = plugins.UpdateAll(opts)
						if err != nil {
							updateErr = err
							return err
						}
					}
					return nil
				}).
				Run()

			if spinnerErr != nil {
				return spinnerErr
			}
			if updateErr != nil {
				return updateErr
			}

			if dryRun {
				fmt.Println(pluginMutedStyle.Render("Dry run - no changes made"))
				fmt.Println()
			}

			anyUpdated := false
			anyPending := false

			for _, r := range results {
				if r.WasUpdated {
					anyUpdated = true
					fmt.Printf("%s %s: %s %s %s\n",
						pluginCheckmark,
						pluginNameStyle.Render(r.Plugin.Name),
						pluginMutedStyle.Render(r.OldVersion),
						pluginArrow,
						pluginVersionStyle.Render(r.NewVersion),
					)
				} else if r.SkipReason != "" {
					fmt.Printf("%s %s: %s\n",
						pluginMutedStyle.Render("-"),
						pluginNameStyle.Render(r.Plugin.Name),
						pluginMutedStyle.Render(r.SkipReason),
					)
				} else if dryRun && r.NewCommit != r.OldCommit {
					anyPending = true
					fmt.Printf("%s %s: %s\n",
						pluginWarnStyle.Render("!"),
						pluginNameStyle.Render(r.Plugin.Name),
						"update available",
					)
				}
			}

			if !anyUpdated && !anyPending && !dryRun {
				fmt.Printf("%s All plugins up to date\n", pluginCheckmark)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force update even if at same version")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be updated without making changes")

	return cmd
}

func removePluginCmd(cfgPath *string) *cobra.Command {
	var noConfirm bool

	cmd := &cobra.Command{
		Use:     "remove <name>",
		Short:   "Remove an installed plugin",
		Aliases: []string{"uninstall", "rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Confirm unless --yes flag
			if !noConfirm {
				var confirm bool
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title(fmt.Sprintf("Remove plugin %s?", pluginNameStyle.Render(name))).
							Description("This will remove the plugin and all its agents, skills, and tools.").
							Affirmative("Remove").
							Negative("Cancel").
							Value(&confirm),
					),
				).WithTheme(huh.ThemeCharm())

				if err := form.Run(); err != nil {
					return err
				}

				if !confirm {
					fmt.Println(pluginMutedStyle.Render("Cancelled"))
					return nil
				}
			}

			result, err := plugins.RemoveWithInfo(name)
			if err != nil {
				return err
			}

			fmt.Printf("%s Removed %s\n",
				pluginCheckmark,
				pluginNameStyle.Render(result.Name),
			)

			// Show what was removed
			var removed []string
			if len(result.Agents) > 0 {
				removed = append(removed, fmt.Sprintf("%d agent(s)", len(result.Agents)))
			}
			if len(result.Skills) > 0 {
				removed = append(removed, fmt.Sprintf("%d skill(s)", len(result.Skills)))
			}
			if len(result.Tools) > 0 {
				removed = append(removed, fmt.Sprintf("%d tool(s)", len(result.Tools)))
			}
			if len(removed) > 0 {
				fmt.Printf("  %s %s\n", pluginArrow, strings.Join(removed, ", "))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&noConfirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// handleDelegateSetup prompts the user to configure delegates declared by a plugin.
func handleDelegateSetup(delegates map[string]string) error {
	for taskType, agentHandle := range delegates {
		// Check if there's already a delegate configured for this task type
		currentDelegate, err := config.GetDelegate(taskType)
		if err != nil {
			return err
		}

		if currentDelegate != "" && currentDelegate == agentHandle {
			// Already configured correctly
			fmt.Printf("%s %s delegate already set to %s\n",
				pluginCheckmark,
				taskType,
				pluginNameStyle.Render(agentHandle),
			)
			continue
		}

		// Build the prompt
		var title, description string
		if currentDelegate != "" {
			title = fmt.Sprintf("Set %s as the default %s agent?",
				pluginNameStyle.Render(agentHandle),
				taskType,
			)
			description = fmt.Sprintf("Current %s delegate: %s", taskType, currentDelegate)
		} else {
			title = fmt.Sprintf("Set %s as the default %s agent?",
				pluginNameStyle.Render(agentHandle),
				taskType,
			)
			description = fmt.Sprintf("This will handle all %s tasks automatically.", taskType)
		}

		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(title).
					Description(description).
					Affirmative("Yes").
					Negative("No").
					Value(&confirm),
			),
		).WithTheme(huh.ThemeCharm())

		if err := form.Run(); err != nil {
			return err
		}

		if confirm {
			previous, err := config.SetDelegate(taskType, agentHandle)
			if err != nil {
				return err
			}

			if previous != "" {
				fmt.Printf("%s %s delegate: %s %s %s\n",
					pluginCheckmark,
					taskType,
					pluginMutedStyle.Render(previous),
					pluginArrow,
					pluginNameStyle.Render(agentHandle),
				)
			} else {
				fmt.Printf("%s %s delegate set to %s\n",
					pluginCheckmark,
					taskType,
					pluginNameStyle.Render(agentHandle),
				)
			}
		} else {
			fmt.Printf("%s Skipped %s delegate configuration\n",
				pluginMutedStyle.Render("-"),
				taskType,
			)
		}
	}

	return nil
}
