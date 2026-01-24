package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/plugins"
)

func newPluginsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugins",
		Short:   "Manage plugins",
		Aliases: []string{"plugin"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to list
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
				fmt.Println("No plugins installed.")
				fmt.Println("")
				fmt.Println("Install a plugin with:")
				fmt.Println("  ayo plugins install <owner/name>")
				return nil
			}

			// Styles
			purple := lipgloss.Color("#a78bfa")
			cyan := lipgloss.Color("#67e8f9")
			muted := lipgloss.Color("#6b7280")
			text := lipgloss.Color("#e5e7eb")
			green := lipgloss.Color("#34d399")
			red := lipgloss.Color("#f87171")

			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
			nameStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
			versionStyle := lipgloss.NewStyle().Foreground(green)
			descStyle := lipgloss.NewStyle().Foreground(text)
			mutedStyle := lipgloss.NewStyle().Foreground(muted)
			disabledStyle := lipgloss.NewStyle().Foreground(red)

			fmt.Println(headerStyle.Render("Installed Plugins"))
			fmt.Println()

			for _, p := range installed {
				status := ""
				if p.Disabled {
					status = disabledStyle.Render(" (disabled)")
				}

				fmt.Printf("%s %s%s\n",
					nameStyle.Render(p.Name),
					versionStyle.Render("v"+p.Version),
					status,
				)

				// Show agents
				if len(p.Agents) > 0 {
					fmt.Printf("  %s %s\n",
						mutedStyle.Render("Agents:"),
						descStyle.Render(strings.Join(p.Agents, ", ")),
					)
				}

				// Show skills
				if len(p.Skills) > 0 {
					fmt.Printf("  %s %s\n",
						mutedStyle.Render("Skills:"),
						descStyle.Render(strings.Join(p.Skills, ", ")),
					)
				}

				// Show tools
				if len(p.Tools) > 0 {
					fmt.Printf("  %s %s\n",
						mutedStyle.Render("Tools:"),
						descStyle.Render(strings.Join(p.Tools, ", ")),
					)
				}

				fmt.Println()
			}

			return nil
		},
	}
}

func installPluginCmd(cfgPath *string) *cobra.Command {
	var force bool
	var local string
	var skipDeps bool

	cmd := &cobra.Command{
		Use:   "install <plugin>",
		Short: "Install a plugin from a git repository",
		Long: `Install a plugin from a git repository.

Plugin references can be:
  - Full URL: https://github.com/owner/ayo-plugins-name
  - Shorthand: owner/name or owner/ayo-plugins-name  
  - Name only: name (assumes github.com/name/ayo-plugins-name)

Examples:
  ayo plugins install alexcabrera/crush
  ayo plugins install https://github.com/alexcabrera/ayo-plugins-crush
  ayo plugins install --local ./my-plugin`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Styles
			green := lipgloss.Color("#34d399")
			yellow := lipgloss.Color("#fbbf24")
			cyan := lipgloss.Color("#67e8f9")

			successStyle := lipgloss.NewStyle().Foreground(green)
			warnStyle := lipgloss.NewStyle().Foreground(yellow)
			nameStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)

			opts := &plugins.InstallOptions{
				Force:               force,
				SkipDependencyCheck: skipDeps,
			}

			var result *plugins.InstallResult
			var err error

			if local != "" {
				// Install from local directory
				result, err = plugins.InstallFromLocal(local, opts)
			} else if len(args) == 0 {
				return fmt.Errorf("plugin reference required (or use --local)")
			} else {
				// First check for conflicts
				gitURL, name, parseErr := plugins.ParsePluginURL(args[0])
				if parseErr != nil {
					return fmt.Errorf("parse plugin reference: %w", parseErr)
				}

				fmt.Printf("Installing %s from %s...\n", nameStyle.Render(name), gitURL)

				result, err = plugins.Install(args[0], opts)
			}

			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("%s Installed %s v%s\n",
				successStyle.Render("[ok]"),
				nameStyle.Render(result.Plugin.Name),
				result.Manifest.Version,
			)

			// Show what was installed
			if len(result.Plugin.Agents) > 0 {
				fmt.Printf("  Agents: %s\n", strings.Join(result.Plugin.Agents, ", "))
			}
			if len(result.Plugin.Skills) > 0 {
				fmt.Printf("  Skills: %s\n", strings.Join(result.Plugin.Skills, ", "))
			}
			if len(result.Plugin.Tools) > 0 {
				fmt.Printf("  Tools: %s\n", strings.Join(result.Plugin.Tools, ", "))
			}

			// Show warnings
			for _, warn := range result.Warnings {
				fmt.Printf("%s %s\n", warnStyle.Render("[warn]"), warn)
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

			// Styles
			purple := lipgloss.Color("#a78bfa")
			cyan := lipgloss.Color("#67e8f9")
			muted := lipgloss.Color("#6b7280")
			text := lipgloss.Color("#e5e7eb")
			green := lipgloss.Color("#34d399")

			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
			nameStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
			labelStyle := lipgloss.NewStyle().Foreground(muted).Width(12)
			valueStyle := lipgloss.NewStyle().Foreground(text)
			versionStyle := lipgloss.NewStyle().Foreground(green)

			fmt.Println(headerStyle.Render("Plugin Details"))
			fmt.Println()

			fmt.Printf("%s %s\n", labelStyle.Render("Name:"), nameStyle.Render(plugin.Name))
			fmt.Printf("%s %s\n", labelStyle.Render("Version:"), versionStyle.Render("v"+plugin.Version))
			fmt.Printf("%s %s\n", labelStyle.Render("Source:"), valueStyle.Render(plugin.GitURL))
			fmt.Printf("%s %s\n", labelStyle.Render("Path:"), valueStyle.Render(plugin.Path))
			fmt.Printf("%s %s\n", labelStyle.Render("Installed:"), valueStyle.Render(plugin.InstalledAt.Format(time.RFC3339)))

			if !plugin.UpdatedAt.IsZero() {
				fmt.Printf("%s %s\n", labelStyle.Render("Updated:"), valueStyle.Render(plugin.UpdatedAt.Format(time.RFC3339)))
			}

			fmt.Println()

			if len(plugin.Agents) > 0 {
				fmt.Printf("%s\n", labelStyle.Render("Agents:"))
				for _, a := range plugin.Agents {
					fmt.Printf("  %s\n", valueStyle.Render(a))
				}
			}

			if len(plugin.Skills) > 0 {
				fmt.Printf("%s\n", labelStyle.Render("Skills:"))
				for _, s := range plugin.Skills {
					fmt.Printf("  %s\n", valueStyle.Render(s))
				}
			}

			if len(plugin.Tools) > 0 {
				fmt.Printf("%s\n", labelStyle.Render("Tools:"))
				for _, t := range plugin.Tools {
					fmt.Printf("  %s\n", valueStyle.Render(t))
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
			// Styles
			green := lipgloss.Color("#34d399")
			yellow := lipgloss.Color("#fbbf24")
			cyan := lipgloss.Color("#67e8f9")
			muted := lipgloss.Color("#6b7280")

			successStyle := lipgloss.NewStyle().Foreground(green)
			warnStyle := lipgloss.NewStyle().Foreground(yellow)
			nameStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
			mutedStyle := lipgloss.NewStyle().Foreground(muted)

			opts := &plugins.UpdateOptions{
				Force:  force,
				DryRun: dryRun,
			}

			var results []*plugins.UpdateResult

			if len(args) == 1 {
				result, err := plugins.Update(args[0], opts)
				if err != nil {
					return err
				}
				results = []*plugins.UpdateResult{result}
			} else {
				var err error
				results, err = plugins.UpdateAll(opts)
				if err != nil {
					return err
				}
			}

			if dryRun {
				fmt.Println("Dry run - no changes made")
				fmt.Println()
			}

			anyUpdated := false
			for _, r := range results {
				if r.WasUpdated {
					anyUpdated = true
					fmt.Printf("%s %s: %s -> %s\n",
						successStyle.Render("[updated]"),
						nameStyle.Render(r.Plugin.Name),
						r.OldVersion,
						r.NewVersion,
					)
				} else if r.SkipReason != "" {
					fmt.Printf("%s %s: %s\n",
						mutedStyle.Render("[skipped]"),
						nameStyle.Render(r.Plugin.Name),
						r.SkipReason,
					)
				} else if dryRun && r.NewCommit != r.OldCommit {
					fmt.Printf("%s %s: update available\n",
						warnStyle.Render("[pending]"),
						nameStyle.Render(r.Plugin.Name),
					)
				}
			}

			if !anyUpdated && !dryRun {
				fmt.Println("All plugins are up to date.")
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

			// Styles
			green := lipgloss.Color("#34d399")
			cyan := lipgloss.Color("#67e8f9")

			successStyle := lipgloss.NewStyle().Foreground(green)
			nameStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)

			// Confirm unless --yes flag
			if !noConfirm {
				var confirm bool
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title(fmt.Sprintf("Remove plugin %s?", name)).
							Description("This will remove the plugin and all its agents, skills, and tools.").
							Value(&confirm),
					),
				)

				if err := form.Run(); err != nil {
					return err
				}

				if !confirm {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			result, err := plugins.RemoveWithInfo(name)
			if err != nil {
				return err
			}

			fmt.Printf("%s Removed %s\n",
				successStyle.Render("[ok]"),
				nameStyle.Render(result.Name),
			)

			if len(result.Agents) > 0 {
				fmt.Printf("  Agents removed: %s\n", strings.Join(result.Agents, ", "))
			}
			if len(result.Skills) > 0 {
				fmt.Printf("  Skills removed: %s\n", strings.Join(result.Skills, ", "))
			}
			if len(result.Tools) > 0 {
				fmt.Printf("  Tools removed: %s\n", strings.Join(result.Tools, ", "))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&noConfirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
