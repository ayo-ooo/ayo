package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// SandboxConfigDef describes a pre-configured sandbox environment from a plugin.
// These are "harnesses" that set up specialized container configurations.
type SandboxConfigDef struct {
	// Name is the unique identifier for this config (e.g., "gpu-enabled", "python-ml").
	Name string `json:"name"`

	// Description briefly describes what this sandbox config provides.
	Description string `json:"description,omitempty"`

	// Path is the relative path to the config directory within the plugin.
	// If empty, defaults to "sandboxes/{name}".
	Path string `json:"path,omitempty"`

	// Requirements lists what the host must provide (e.g., "nvidia-gpu", "8gb-ram").
	Requirements []string `json:"requirements,omitempty"`
}

// SandboxConfig represents the full sandbox configuration loaded from a plugin.
// This is stored in sandboxes/{name}/sandbox.json within the plugin.
type SandboxConfig struct {
	// Name is the config identifier.
	Name string `json:"name"`

	// Description describes the sandbox purpose.
	Description string `json:"description,omitempty"`

	// BaseImage is the container image to use (if applicable).
	BaseImage string `json:"base_image,omitempty"`

	// ProviderRequirements specifies what the sandbox provider must support.
	ProviderRequirements *ProviderRequirements `json:"provider_requirements,omitempty"`

	// Mounts specifies additional mount points.
	Mounts []Mount `json:"mounts,omitempty"`

	// Env specifies environment variables to set.
	Env map[string]string `json:"env,omitempty"`

	// Packages lists packages to install.
	Packages []string `json:"packages,omitempty"`

	// PostCreate is a script to run after sandbox creation.
	// Path is relative to the config directory.
	PostCreate string `json:"post_create,omitempty"`

	// PluginName is set when loading - indicates which plugin provides this config.
	PluginName string `json:"-"`

	// ConfigPath is set when loading - absolute path to the config directory.
	ConfigPath string `json:"-"`
}

// ProviderRequirements specifies what the sandbox provider must support.
type ProviderRequirements struct {
	// GPU indicates GPU passthrough is required.
	GPU bool `json:"gpu,omitempty"`

	// MinMemory is the minimum memory required (e.g., "8G").
	MinMemory string `json:"min_memory,omitempty"`

	// Network specifies network requirements.
	Network string `json:"network,omitempty"`
}

// Mount represents a filesystem mount for the sandbox.
type Mount struct {
	// Src is the source path on the host.
	Src string `json:"src"`

	// Dst is the destination path in the sandbox.
	Dst string `json:"dst"`

	// Type is the mount type (e.g., "bind", "device").
	Type string `json:"type,omitempty"`

	// ReadOnly indicates the mount should be read-only.
	ReadOnly bool `json:"readonly,omitempty"`
}

// SandboxConfigRegistry holds all available sandbox configs from plugins.
type SandboxConfigRegistry struct {
	mu      sync.RWMutex
	configs map[string]*SandboxConfig // name -> config
}

// DefaultSandboxConfigRegistry is the global sandbox config registry.
var DefaultSandboxConfigRegistry = &SandboxConfigRegistry{
	configs: make(map[string]*SandboxConfig),
}

// Register adds a sandbox config to the registry.
func (r *SandboxConfigRegistry) Register(config *SandboxConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[config.Name]; exists {
		return fmt.Errorf("sandbox config %q already registered", config.Name)
	}

	r.configs[config.Name] = config
	return nil
}

// Get retrieves a sandbox config by name.
func (r *SandboxConfigRegistry) Get(name string) (*SandboxConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, ok := r.configs[name]
	return config, ok
}

// List returns all registered sandbox configs.
func (r *SandboxConfigRegistry) List() []*SandboxConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	configs := make([]*SandboxConfig, 0, len(r.configs))
	for _, c := range r.configs {
		configs = append(configs, c)
	}
	return configs
}

// Has returns true if a config with the given name exists.
func (r *SandboxConfigRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.configs[name]
	return ok
}

// Clear removes all configs from the registry (for testing).
func (r *SandboxConfigRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.configs = make(map[string]*SandboxConfig)
}

// SandboxConfigLoadError represents an error loading a sandbox config.
type SandboxConfigLoadError struct {
	PluginName string
	ConfigName string
	Err        error
}

func (e *SandboxConfigLoadError) Error() string {
	return fmt.Sprintf("failed to load sandbox config %s from plugin %s: %v", e.ConfigName, e.PluginName, e.Err)
}

func (e *SandboxConfigLoadError) Unwrap() error {
	return e.Err
}

// LoadedSandboxConfig contains information about a successfully loaded config.
type LoadedSandboxConfig struct {
	Name       string
	PluginName string
}

// LoadSandboxConfigs loads all sandbox configs from installed plugins and registers
// them with the default registry. This should be called during application startup.
//
// Returns a list of successfully loaded configs and any errors encountered.
// Loading continues even if some plugins fail to load.
func LoadSandboxConfigs() ([]LoadedSandboxConfig, []error) {
	reg, err := LoadRegistry()
	if err != nil {
		return nil, []error{fmt.Errorf("load plugin registry: %w", err)}
	}

	var loaded []LoadedSandboxConfig
	var errs []error

	for _, plugin := range reg.ListEnabled() {
		manifest, err := LoadManifest(plugin.Path)
		if err != nil {
			errs = append(errs, &SandboxConfigLoadError{
				PluginName: plugin.Name,
				ConfigName: "",
				Err:        fmt.Errorf("load manifest: %w", err),
			})
			continue
		}

		for _, configDef := range manifest.SandboxConfigs {
			config, err := loadSandboxConfig(plugin, configDef)
			if err != nil {
				errs = append(errs, &SandboxConfigLoadError{
					PluginName: plugin.Name,
					ConfigName: configDef.Name,
					Err:        err,
				})
				continue
			}

			if err := DefaultSandboxConfigRegistry.Register(config); err != nil {
				errs = append(errs, &SandboxConfigLoadError{
					PluginName: plugin.Name,
					ConfigName: configDef.Name,
					Err:        err,
				})
				continue
			}

			loaded = append(loaded, LoadedSandboxConfig{
				Name:       configDef.Name,
				PluginName: plugin.Name,
			})
		}
	}

	return loaded, errs
}

// loadSandboxConfig loads a single sandbox config from a plugin.
func loadSandboxConfig(installedPlugin *InstalledPlugin, def SandboxConfigDef) (*SandboxConfig, error) {
	if def.Name == "" {
		return nil, errors.New("config name is required")
	}

	// Determine config directory path
	configPath := def.Path
	if configPath == "" {
		configPath = filepath.Join("sandboxes", def.Name)
	}
	fullConfigPath := filepath.Join(installedPlugin.Path, configPath)

	// Check config directory exists
	if _, err := os.Stat(fullConfigPath); err != nil {
		return nil, fmt.Errorf("config directory not found: %s", configPath)
	}

	// Load sandbox.json
	sandboxJSONPath := filepath.Join(fullConfigPath, "sandbox.json")
	data, err := os.ReadFile(sandboxJSONPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If no sandbox.json, create minimal config from def
			return &SandboxConfig{
				Name:        def.Name,
				Description: def.Description,
				PluginName:  installedPlugin.Name,
				ConfigPath:  fullConfigPath,
			}, nil
		}
		return nil, fmt.Errorf("read sandbox.json: %w", err)
	}

	var config SandboxConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse sandbox.json: %w", err)
	}

	// Override name if not set in file
	if config.Name == "" {
		config.Name = def.Name
	}

	// Set plugin info
	config.PluginName = installedPlugin.Name
	config.ConfigPath = fullConfigPath

	return &config, nil
}

// ListPluginSandboxConfigs returns all sandbox config definitions from enabled plugins.
// This is useful for displaying what configs are available from plugins
// without necessarily loading them.
func ListPluginSandboxConfigs() ([]SandboxConfigDef, error) {
	reg, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load plugin registry: %w", err)
	}

	var allConfigs []SandboxConfigDef

	for _, plugin := range reg.ListEnabled() {
		manifest, err := LoadManifest(plugin.Path)
		if err != nil {
			continue // Skip plugins with invalid manifests
		}

		allConfigs = append(allConfigs, manifest.SandboxConfigs...)
	}

	return allConfigs, nil
}

// GetSandboxConfigPlugin returns the plugin that provides a given sandbox config.
// Returns nil if the config is not from a plugin.
func GetSandboxConfigPlugin(configName string) (*InstalledPlugin, error) {
	reg, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load plugin registry: %w", err)
	}

	for _, plugin := range reg.ListEnabled() {
		manifest, err := LoadManifest(plugin.Path)
		if err != nil {
			continue
		}

		for _, configDef := range manifest.SandboxConfigs {
			if configDef.Name == configName {
				return plugin, nil
			}
		}
	}

	return nil, nil // Not found in any plugin
}
