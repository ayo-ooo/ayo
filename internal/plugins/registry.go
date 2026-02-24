package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

	// Squads lists the installed squad names.
	Squads []string `json:"squads,omitempty"`

	// Triggers lists the installed trigger type names.
	Triggers []string `json:"triggers,omitempty"`

	// SandboxConfigs lists the installed sandbox config names.
	SandboxConfigs []string `json:"sandbox_configs,omitempty"`

	// Planners lists the installed planner names.
	Planners []string `json:"planners,omitempty"`

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

// testDataDir is used to override the data directory in tests.
// When set, it takes precedence over paths.DataDir().
var testDataDir string

// SetTestDataDir sets a custom data directory for testing.
// Call with empty string to reset to default behavior.
func SetTestDataDir(dir string) {
	testDataDir = dir
}

func getDataDir() string {
	if testDataDir != "" {
		return testDataDir
	}
	return paths.DataDir()
}

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
	return filepath.Join(getDataDir(), RegistryFile)
}

// PluginsDir returns the directory where plugins are installed.
func PluginsDir() string {
	return filepath.Join(getDataDir(), "plugins")
}

// PluginDir returns the directory for a specific plugin.
func PluginDir(name string) string {
	return filepath.Join(PluginsDir(), name)
}

// IsPluginAgent checks if an agent handle belongs to any installed plugin.
func IsPluginAgent(handle string) bool {
	reg, err := LoadRegistry()
	if err != nil {
		return false
	}
	for _, plugin := range reg.Plugins {
		for _, agentHandle := range plugin.Agents {
			if agentHandle == handle {
				return true
			}
		}
	}
	return false
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

// PluginAgentInfo contains information about a plugin agent for display.
type PluginAgentInfo struct {
	Handle      string
	Description string
	PluginName  string
}

// ListPluginAgents returns info about all agents from enabled plugins.
func ListPluginAgents() []PluginAgentInfo {
	reg, err := LoadRegistry()
	if err != nil {
		return nil
	}

	var agents []PluginAgentInfo
	for _, plugin := range reg.ListEnabled() {
		for _, handle := range plugin.Agents {
			resolved := plugin.GetResolvedAgentHandle(handle)
			agents = append(agents, PluginAgentInfo{
				Handle:     resolved,
				PluginName: plugin.Name,
			})
		}
	}
	return agents
}

// PluginTriggerInfo contains information about a plugin trigger for display.
type PluginTriggerInfo struct {
	Name       string
	PluginName string
}

// ListPluginTriggers returns info about all triggers from enabled plugins.
func ListPluginTriggers() []PluginTriggerInfo {
	reg, err := LoadRegistry()
	if err != nil {
		return nil
	}

	var triggers []PluginTriggerInfo
	for _, plugin := range reg.ListEnabled() {
		for _, name := range plugin.Triggers {
			triggers = append(triggers, PluginTriggerInfo{
				Name:       name,
				PluginName: plugin.Name,
			})
		}
	}
	return triggers
}

// ComponentInfo is a generic component reference for search results.
type ComponentInfo struct {
	Type       string // "agent", "skill", "tool", "squad", "trigger", "sandbox_config", "planner"
	Name       string
	PluginName string
}

// Search returns all components matching the query across all types.
func (r *Registry) Search(query string) []ComponentInfo {
	var results []ComponentInfo
	query = strings.ToLower(query)

	for _, plugin := range r.ListEnabled() {
		// Search agents
		for _, name := range plugin.Agents {
			if strings.Contains(strings.ToLower(name), query) {
				results = append(results, ComponentInfo{Type: "agent", Name: name, PluginName: plugin.Name})
			}
		}
		// Search skills
		for _, name := range plugin.Skills {
			if strings.Contains(strings.ToLower(name), query) {
				results = append(results, ComponentInfo{Type: "skill", Name: name, PluginName: plugin.Name})
			}
		}
		// Search tools
		for _, name := range plugin.Tools {
			if strings.Contains(strings.ToLower(name), query) {
				results = append(results, ComponentInfo{Type: "tool", Name: name, PluginName: plugin.Name})
			}
		}
		// Search squads
		for _, name := range plugin.Squads {
			if strings.Contains(strings.ToLower(name), query) {
				results = append(results, ComponentInfo{Type: "squad", Name: name, PluginName: plugin.Name})
			}
		}
		// Search triggers
		for _, name := range plugin.Triggers {
			if strings.Contains(strings.ToLower(name), query) {
				results = append(results, ComponentInfo{Type: "trigger", Name: name, PluginName: plugin.Name})
			}
		}
		// Search sandbox configs
		for _, name := range plugin.SandboxConfigs {
			if strings.Contains(strings.ToLower(name), query) {
				results = append(results, ComponentInfo{Type: "sandbox_config", Name: name, PluginName: plugin.Name})
			}
		}
		// Search planners
		for _, name := range plugin.Planners {
			if strings.Contains(strings.ToLower(name), query) {
				results = append(results, ComponentInfo{Type: "planner", Name: name, PluginName: plugin.Name})
			}
		}
	}

	return results
}

// GetPluginStats returns statistics about a plugin's components.
type PluginStats struct {
	Agents         int
	Skills         int
	Tools          int
	Squads         int
	Triggers       int
	SandboxConfigs int
	Planners       int
}

// Stats returns component counts for the plugin.
func (p *InstalledPlugin) Stats() PluginStats {
	return PluginStats{
		Agents:         len(p.Agents),
		Skills:         len(p.Skills),
		Tools:          len(p.Tools),
		Squads:         len(p.Squads),
		Triggers:       len(p.Triggers),
		SandboxConfigs: len(p.SandboxConfigs),
		Planners:       len(p.Planners),
	}
}
