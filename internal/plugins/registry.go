package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
)

// Registry tracks installed plugins.
// Stored as packages.json in the data directory.
type Registry struct {
	// Version is the registry format version.
	Version int `json:"version"`

	// Plugins maps plugin names to their installation info.
	Plugins map[string]*InstalledPlugin `json:"plugins"`
}

// InstalledPlugin contains information about an installed plugin.
type InstalledPlugin struct {
	// Name is the plugin identifier.
	Name string `json:"name"`

	// Version is the installed version (from manifest).
	Version string `json:"version"`

	// GitURL is the repository URL used for installation.
	GitURL string `json:"git_url"`

	// GitCommit is the commit hash of the installed version.
	GitCommit string `json:"git_commit"`

	// InstalledAt is when the plugin was installed.
	InstalledAt time.Time `json:"installed_at"`

	// UpdatedAt is when the plugin was last updated.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Path is the absolute path to the plugin directory.
	Path string `json:"path"`

	// Agents lists the installed agent handles.
	Agents []string `json:"agents,omitempty"`

	// Skills lists the installed skill names.
	Skills []string `json:"skills,omitempty"`

	// Tools lists the installed tool names.
	Tools []string `json:"tools,omitempty"`

	// Disabled indicates the plugin is installed but not active.
	Disabled bool `json:"disabled,omitempty"`

	// Renames maps original names to renamed versions for conflict resolution.
	// Key is original name (e.g., "@crush"), value is renamed (e.g., "@my-crush").
	Renames map[string]string `json:"renames,omitempty"`
}

// RegistryFile is the filename for the registry.
const RegistryFile = "packages.json"

// CurrentRegistryVersion is the current registry format version.
const CurrentRegistryVersion = 1

// Registry errors
var (
	ErrPluginNotInstalled = errors.New("plugin not installed")
	ErrPluginExists       = errors.New("plugin already installed")
)

// LoadRegistry reads the plugin registry from disk.
// Returns an empty registry if the file doesn't exist.
func LoadRegistry() (*Registry, error) {
	registryPath := RegistryPath()

	data, err := os.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{
				Version: CurrentRegistryVersion,
				Plugins: make(map[string]*InstalledPlugin),
			}, nil
		}
		return nil, fmt.Errorf("read registry: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}

	// Initialize map if nil (from older version)
	if reg.Plugins == nil {
		reg.Plugins = make(map[string]*InstalledPlugin)
	}

	return &reg, nil
}

// Save writes the registry to disk.
func (r *Registry) Save() error {
	registryPath := RegistryPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(registryPath), 0o755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(registryPath, data, 0o644); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}

	return nil
}

// Add adds a plugin to the registry.
func (r *Registry) Add(plugin *InstalledPlugin) error {
	if _, exists := r.Plugins[plugin.Name]; exists {
		return fmt.Errorf("%w: %s", ErrPluginExists, plugin.Name)
	}

	r.Plugins[plugin.Name] = plugin
	return r.Save()
}

// Update updates an existing plugin in the registry.
func (r *Registry) Update(plugin *InstalledPlugin) error {
	if _, exists := r.Plugins[plugin.Name]; !exists {
		return fmt.Errorf("%w: %s", ErrPluginNotInstalled, plugin.Name)
	}

	plugin.UpdatedAt = time.Now()
	r.Plugins[plugin.Name] = plugin
	return r.Save()
}

// Remove removes a plugin from the registry.
func (r *Registry) Remove(name string) error {
	if _, exists := r.Plugins[name]; !exists {
		return fmt.Errorf("%w: %s", ErrPluginNotInstalled, name)
	}

	delete(r.Plugins, name)
	return r.Save()
}

// Get returns a plugin by name.
func (r *Registry) Get(name string) (*InstalledPlugin, error) {
	plugin, exists := r.Plugins[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrPluginNotInstalled, name)
	}
	return plugin, nil
}

// Has checks if a plugin is installed.
func (r *Registry) Has(name string) bool {
	_, exists := r.Plugins[name]
	return exists
}

// List returns all installed plugins sorted by name.
func (r *Registry) List() []*InstalledPlugin {
	plugins := make([]*InstalledPlugin, 0, len(r.Plugins))
	for _, p := range r.Plugins {
		plugins = append(plugins, p)
	}
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})
	return plugins
}

// ListEnabled returns all enabled (non-disabled) plugins.
func (r *Registry) ListEnabled() []*InstalledPlugin {
	var plugins []*InstalledPlugin
	for _, p := range r.Plugins {
		if !p.Disabled {
			plugins = append(plugins, p)
		}
	}
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})
	return plugins
}

// RegistryPath returns the path to the registry file.
func RegistryPath() string {
	return filepath.Join(paths.DataDir(), RegistryFile)
}

// PluginsDir returns the directory where plugins are installed.
func PluginsDir() string {
	return filepath.Join(paths.DataDir(), "plugins")
}

// PluginDir returns the directory for a specific plugin.
func PluginDir(name string) string {
	return filepath.Join(PluginsDir(), name)
}

// GetResolvedAgentHandle returns the effective handle for an agent,
// accounting for any renames that may have been applied.
func (p *InstalledPlugin) GetResolvedAgentHandle(originalHandle string) string {
	if p.Renames != nil {
		if renamed, ok := p.Renames[originalHandle]; ok {
			return renamed
		}
	}
	return originalHandle
}

// GetOriginalAgentHandle returns the original handle for a renamed agent.
func (p *InstalledPlugin) GetOriginalAgentHandle(renamedHandle string) string {
	if p.Renames != nil {
		for original, renamed := range p.Renames {
			if renamed == renamedHandle {
				return original
			}
		}
	}
	return renamedHandle
}
