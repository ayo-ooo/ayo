package plugins

import (
	"fmt"
	"os"

	"github.com/alexcabrera/ayo/internal/paths"
)

// Remove uninstalls a plugin.
func Remove(name string) error {
	// Load registry
	registry, err := LoadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	// Check if installed
	if !registry.Has(name) {
		return fmt.Errorf("%w: %s", ErrPluginNotInstalled, name)
	}

	// Remove plugin directory
	pluginDir := paths.PluginDir(name)
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("remove plugin directory: %w", err)
	}

	// Remove from registry
	if err := registry.Remove(name); err != nil {
		return fmt.Errorf("update registry: %w", err)
	}

	return nil
}

// RemoveResult contains information about a removal operation.
type RemoveResult struct {
	Name    string
	Path    string
	Agents  []string
	Skills  []string
	Tools   []string
}

// RemoveWithInfo uninstalls a plugin and returns info about what was removed.
func RemoveWithInfo(name string) (*RemoveResult, error) {
	// Load registry first to get info
	registry, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}

	plugin, err := registry.Get(name)
	if err != nil {
		return nil, err
	}

	result := &RemoveResult{
		Name:   name,
		Path:   plugin.Path,
		Agents: plugin.Agents,
		Skills: plugin.Skills,
		Tools:  plugin.Tools,
	}

	// Remove plugin
	if err := Remove(name); err != nil {
		return nil, err
	}

	return result, nil
}

// Purge removes all installed plugins.
func Purge() error {
	registry, err := LoadRegistry()
	if err != nil {
		return err
	}

	plugins := registry.List()
	for _, plugin := range plugins {
		if err := Remove(plugin.Name); err != nil {
			return fmt.Errorf("remove %s: %w", plugin.Name, err)
		}
	}

	// Remove plugins directory
	pluginsDir := paths.PluginsDir()
	if err := os.RemoveAll(pluginsDir); err != nil {
		return fmt.Errorf("remove plugins directory: %w", err)
	}

	return nil
}
