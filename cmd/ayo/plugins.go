package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/plugins"
)

// newPluginsCmd creates the plugins command group
func newPluginsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage ayo plugins",
		Long: `Manage ayo plugins - install, remove, and list extensions.

Plugins extend ayo with additional agents, tools, skills, and providers.
They are distributed as git repositories with the naming convention:
ayo-plugins-<name>

Examples:
  ayo plugin install https://github.com/acme/ayo-plugins-devtools
  ayo plugin install ./my-local-plugin
  ayo plugin list
  ayo plugin show @acme/devtools
  ayo plugin remove @acme/devtools`,
	}

	cmd.AddCommand(newPluginInstallCmd())
	cmd.AddCommand(newPluginRemoveCmd())
	cmd.AddCommand(newPluginListCmd())
	cmd.AddCommand(newPluginShowCmd())

	return cmd
}

// newPluginInstallCmd creates the plugin install command
func newPluginInstallCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "install [plugin-ref]",
		Short: "Install a plugin",
		Long: `Install a plugin from a git repository or local directory.

Plugin references can be:
- Git repository URL: https://github.com/user/ayo-plugins-name
- Git SSH URL: git@github.com:user/ayo-plugins-name.git
- Local directory path: ./my-plugin

Examples:
  ayo plugin install https://github.com/acme/ayo-plugins-devtools
  ayo plugin install ./my-local-plugin
  ayo plugin install git@github.com:acme/ayo-plugins-tools.git`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginRef := args[0]
			return runPluginInstall(pluginRef, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinstall if plugin already exists")
	return cmd
}

// newPluginRemoveCmd creates the plugin remove command
func newPluginRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [plugin-name]",
		Short: "Remove an installed plugin",
		Long: `Remove an installed plugin and all its components.

Examples:
  ayo plugin remove @acme/devtools
  ayo plugin remove my-plugin`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]
			return runPluginRemove(pluginName)
		},
	}
	return cmd
}

// newPluginListCmd creates the plugin list command
func newPluginListCmd() *cobra.Command {
	var showDisabled bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Long: `List all installed plugins.

By default, only enabled plugins are shown. Use --all to include disabled plugins.

Examples:
  ayo plugin list
  ayo plugin list --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginList(showDisabled)
		},
	}

	cmd.Flags().BoolVarP(&showDisabled, "all", "a", false, "Show all plugins including disabled ones")
	return cmd
}

// newPluginShowCmd creates the plugin show command
func newPluginShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [plugin-name]",
		Short: "Show plugin details",
		Long: `Show detailed information about an installed plugin.

Examples:
  ayo plugin show @acme/devtools
  ayo plugin show my-plugin`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]
			return runPluginShow(pluginName)
		},
	}
	return cmd
}

// runPluginInstall handles plugin installation
func runPluginInstall(pluginRef string, force bool) error {
	// Set up installation options
	opts := &plugins.InstallOptions{
		Force: force,
	}

	var result *plugins.InstallResult
	var err error

	if strings.HasPrefix(pluginRef, "https://") || strings.HasPrefix(pluginRef, "git@") {
		// Git repository installation
		fmt.Printf("Installing plugin from git repository: %s\n", pluginRef)
		result, err = plugins.Install(pluginRef, opts)
	} else {
		// Local directory installation
		fmt.Printf("Installing plugin from local directory: %s\n", pluginRef)
		result, err = plugins.InstallFromLocal(pluginRef, opts)
	}

	if err != nil {
		return fmt.Errorf("plugin installation failed: %w", err)
	}

	// Display installation results
	fmt.Printf("✓ Plugin %s (%s) installed successfully\n", result.Plugin.Name, result.Plugin.Version)

	if len(result.MissingDeps) > 0 {
		fmt.Printf("⚠️  Missing dependencies (%d):\n", len(result.MissingDeps))
		for _, dep := range result.MissingDeps {
			fmt.Printf("  - %s", dep.Name)
			if dep.InstallHint != "" {
				fmt.Printf(" (%s)", dep.InstallHint)
			}
			fmt.Printf("\n")
		}
	}

	if result.Manifest.Description != "" {
		fmt.Printf("\nDescription: %s\n", result.Manifest.Description)
	}

	// Show what was installed
	if len(result.Manifest.Agents) > 0 {
		fmt.Printf("\nAgents (%d):\n", len(result.Manifest.Agents))
		for _, agent := range result.Manifest.Agents {
			fmt.Printf("  - %s\n", agent)
		}
	}

	if len(result.Manifest.Tools) > 0 {
		fmt.Printf("\nTools (%d):\n", len(result.Manifest.Tools))
		for _, tool := range result.Manifest.Tools {
			fmt.Printf("  - %s\n", tool)
		}
	}

	if len(result.Manifest.Skills) > 0 {
		fmt.Printf("\nSkills (%d):\n", len(result.Manifest.Skills))
		for _, skill := range result.Manifest.Skills {
			fmt.Printf("  - %s\n", skill)
		}
	}

	return nil
}

// runPluginRemove handles plugin removal
func runPluginRemove(pluginName string) error {
	// Load registry
	reg, err := plugins.LoadRegistry()
	if err != nil {
		return fmt.Errorf("load plugin registry: %w", err)
	}

	// Check if plugin exists
	if !reg.Has(pluginName) {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	// Get plugin info
	plugin, err := reg.Get(pluginName)
	if err != nil {
		return fmt.Errorf("get plugin info: %w", err)
	}

	// Remove from registry
	if err := reg.Remove(pluginName); err != nil {
		return fmt.Errorf("remove plugin from registry: %w", err)
	}

	// Remove plugin files
	pluginDir := plugins.PluginDir(pluginName)
	if err := os.RemoveAll(pluginDir); err != nil && !os.IsNotExist(err) {
		fmt.Printf("⚠️  Warning: Could not remove plugin files: %v\n", err)
	} else {
		fmt.Printf("✓ Plugin %s removed from registry and files\n", pluginName)
	}

	return nil
}

// runPluginList handles plugin listing
func runPluginList(showDisabled bool) error {
	// Load registry
	reg, err := plugins.LoadRegistry()
	if err != nil {
		return fmt.Errorf("load plugin registry: %w", err)
	}

	var pluginList []*plugins.InstalledPlugin
	if showDisabled {
		pluginList = reg.List()
	} else {
		pluginList = reg.ListEnabled()
	}

	if len(pluginList) == 0 {
		fmt.Println("No plugins installed")
		return nil
	}

	fmt.Printf("Installed plugins (%d):\n", len(pluginList))
	for _, plugin := range pluginList {
		status := "enabled"
		if plugin.Disabled {
			status = "disabled"
		}
		fmt.Printf("  %s (%s) - %s\n", plugin.Name, plugin.Version, status)
	}

	return nil
}

// runPluginShow handles plugin details display
func runPluginShow(pluginName string) error {
	// Load registry
	reg, err := plugins.LoadRegistry()
	if err != nil {
		return fmt.Errorf("load plugin registry: %w", err)
	}

	// Get plugin info
	plugin, err := reg.Get(pluginName)
	if err != nil {
		return fmt.Errorf("get plugin info: %w", err)
	}

	fmt.Printf("Plugin: %s\n", plugin.Name)
	fmt.Printf("Version: %s\n", plugin.Version)
	fmt.Printf("Status: ")
	if plugin.Disabled {
		fmt.Printf("disabled\n")
	} else {
		fmt.Printf("enabled\n")
	}
	fmt.Printf("Installed: %s\n", plugin.InstalledAt.Format("2006-01-02 15:04:05"))

	if plugin.GitURL != "" {
		fmt.Printf("Source: %s\n", plugin.GitURL)
	}

	if len(plugin.Agents) > 0 {
		fmt.Printf("\nAgents (%d):\n", len(plugin.Agents))
		for _, agent := range plugin.Agents {
			fmt.Printf("  - %s\n", agent)
		}
	}

	if len(plugin.Tools) > 0 {
		fmt.Printf("\nTools (%d):\n", len(plugin.Tools))
		for _, tool := range plugin.Tools {
			fmt.Printf("  - %s\n", tool)
		}
	}

	if len(plugin.Skills) > 0 {
		fmt.Printf("\nSkills (%d):\n", len(plugin.Skills))
		for _, skill := range plugin.Skills {
			fmt.Printf("  - %s\n", skill)
		}
	}

	return nil
}
