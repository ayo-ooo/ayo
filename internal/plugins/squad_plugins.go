package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// PluginSquad represents a squad definition loaded from a plugin.
// This contains the full squad configuration ready for instantiation.
type PluginSquad struct {
	// Name is the squad identifier.
	Name string `json:"name"`

	// Description describes the squad purpose.
	Description string `json:"description,omitempty"`

	// Agents lists the agent handles that are part of this squad.
	Agents []string `json:"agents,omitempty"`

	// Planners configures the planners for this squad.
	Planners *SquadPlannerConfig `json:"planners,omitempty"`

	// PluginName is set when loading - indicates which plugin provides this squad.
	PluginName string `json:"-"`

	// SquadPath is set when loading - absolute path to the squad directory.
	// Contains SQUAD.md and optionally ayo.json.
	SquadPath string `json:"-"`

	// ConstitutionPath is the path to the SQUAD.md file.
	ConstitutionPath string `json:"-"`

	// ConfigPath is the path to the ayo.json file (optional).
	ConfigPath string `json:"-"`
}

// HasConstitution returns true if this squad has a SQUAD.md file.
func (ps *PluginSquad) HasConstitution() bool {
	if ps.ConstitutionPath == "" {
		return false
	}
	_, err := os.Stat(ps.ConstitutionPath)
	return err == nil
}

// HasConfig returns true if this squad has an ayo.json file.
func (ps *PluginSquad) HasConfig() bool {
	if ps.ConfigPath == "" {
		return false
	}
	_, err := os.Stat(ps.ConfigPath)
	return err == nil
}

// SquadRegistry holds all available squad definitions from plugins.
type SquadRegistry struct {
	mu     sync.RWMutex
	squads map[string]*PluginSquad // name -> squad
}

// DefaultSquadRegistry is the global squad registry.
var DefaultSquadRegistry = &SquadRegistry{
	squads: make(map[string]*PluginSquad),
}

// Register adds a squad to the registry.
func (r *SquadRegistry) Register(squad *PluginSquad) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.squads[squad.Name]; exists {
		return fmt.Errorf("squad %q already registered", squad.Name)
	}

	r.squads[squad.Name] = squad
	return nil
}

// Get retrieves a squad by name.
func (r *SquadRegistry) Get(name string) (*PluginSquad, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	squad, ok := r.squads[name]
	return squad, ok
}

// List returns all registered squads.
func (r *SquadRegistry) List() []*PluginSquad {
	r.mu.RLock()
	defer r.mu.RUnlock()

	squads := make([]*PluginSquad, 0, len(r.squads))
	for _, s := range r.squads {
		squads = append(squads, s)
	}
	return squads
}

// Has returns true if a squad with the given name exists.
func (r *SquadRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.squads[name]
	return ok
}

// Clear removes all squads from the registry (for testing).
func (r *SquadRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.squads = make(map[string]*PluginSquad)
}

// SquadLoadError represents an error loading a squad from a plugin.
type SquadLoadError struct {
	PluginName string
	SquadName  string
	Err        error
}

func (e *SquadLoadError) Error() string {
	return fmt.Sprintf("failed to load squad %s from plugin %s: %v", e.SquadName, e.PluginName, e.Err)
}

func (e *SquadLoadError) Unwrap() error {
	return e.Err
}

// LoadedSquad contains information about a successfully loaded squad.
type LoadedSquad struct {
	Name       string
	PluginName string
}

// LoadPluginSquads loads all squad definitions from installed plugins and registers
// them with the default registry. This should be called during application startup.
//
// Returns a list of successfully loaded squads and any errors encountered.
// Loading continues even if some plugins fail to load.
func LoadPluginSquads() ([]LoadedSquad, []error) {
	reg, err := LoadRegistry()
	if err != nil {
		return nil, []error{fmt.Errorf("load plugin registry: %w", err)}
	}

	var loaded []LoadedSquad
	var errs []error

	for _, plugin := range reg.ListEnabled() {
		manifest, err := LoadManifest(plugin.Path)
		if err != nil {
			errs = append(errs, &SquadLoadError{
				PluginName: plugin.Name,
				SquadName:  "",
				Err:        fmt.Errorf("load manifest: %w", err),
			})
			continue
		}

		for _, squadDef := range manifest.Squads {
			squad, err := loadPluginSquad(plugin, squadDef)
			if err != nil {
				errs = append(errs, &SquadLoadError{
					PluginName: plugin.Name,
					SquadName:  squadDef.Name,
					Err:        err,
				})
				continue
			}

			if err := DefaultSquadRegistry.Register(squad); err != nil {
				errs = append(errs, &SquadLoadError{
					PluginName: plugin.Name,
					SquadName:  squadDef.Name,
					Err:        err,
				})
				continue
			}

			loaded = append(loaded, LoadedSquad{
				Name:       squadDef.Name,
				PluginName: plugin.Name,
			})
		}
	}

	return loaded, errs
}

// loadPluginSquad loads a single squad from a plugin.
func loadPluginSquad(installedPlugin *InstalledPlugin, def SquadDef) (*PluginSquad, error) {
	if def.Name == "" {
		return nil, errors.New("squad name is required")
	}

	// Determine squad directory path
	squadPath := def.Path
	if squadPath == "" {
		squadPath = filepath.Join("squads", def.Name)
	}
	fullSquadPath := filepath.Join(installedPlugin.Path, squadPath)

	// Check squad directory exists
	if _, err := os.Stat(fullSquadPath); err != nil {
		return nil, fmt.Errorf("squad directory not found: %s", squadPath)
	}

	// Build paths for constitution and config
	constitutionPath := filepath.Join(fullSquadPath, "SQUAD.md")
	configPath := filepath.Join(fullSquadPath, "ayo.json")

	// Verify SQUAD.md exists (required)
	if _, err := os.Stat(constitutionPath); err != nil {
		return nil, fmt.Errorf("SQUAD.md not found in squad directory: %s", squadPath)
	}

	squad := &PluginSquad{
		Name:             def.Name,
		Description:      def.Description,
		Agents:           def.Agents,
		Planners:         def.Planners,
		PluginName:       installedPlugin.Name,
		SquadPath:        fullSquadPath,
		ConstitutionPath: constitutionPath,
	}

	// Check if ayo.json exists (optional)
	if _, err := os.Stat(configPath); err == nil {
		squad.ConfigPath = configPath
	}

	return squad, nil
}

// ListPluginSquads returns all squad definitions from enabled plugins.
// This is useful for displaying what squads are available from plugins
// without necessarily loading them.
func ListPluginSquads() ([]SquadDef, error) {
	reg, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load plugin registry: %w", err)
	}

	var allSquads []SquadDef

	for _, plugin := range reg.ListEnabled() {
		manifest, err := LoadManifest(plugin.Path)
		if err != nil {
			continue // Skip plugins with invalid manifests
		}

		allSquads = append(allSquads, manifest.Squads...)
	}

	return allSquads, nil
}

// GetSquadPlugin returns the plugin that provides a given squad.
// Returns nil if the squad is not from a plugin.
func GetSquadPlugin(squadName string) (*InstalledPlugin, error) {
	reg, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load plugin registry: %w", err)
	}

	for _, plugin := range reg.ListEnabled() {
		manifest, err := LoadManifest(plugin.Path)
		if err != nil {
			continue
		}

		for _, squadDef := range manifest.Squads {
			if squadDef.Name == squadName {
				return plugin, nil
			}
		}
	}

	return nil, nil // Not found in any plugin
}

// ReadSquadConstitution reads the SQUAD.md content from a plugin squad.
func (ps *PluginSquad) ReadConstitution() ([]byte, error) {
	if ps.ConstitutionPath == "" {
		return nil, errors.New("no constitution path set")
	}
	return os.ReadFile(ps.ConstitutionPath)
}

// ReadConfig reads the ayo.json content from a plugin squad.
// Returns nil, nil if no config file exists.
func (ps *PluginSquad) ReadConfig() (map[string]any, error) {
	if ps.ConfigPath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(ps.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read ayo.json: %w", err)
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse ayo.json: %w", err)
	}

	return config, nil
}

// SquadInfo contains information about a squad for display.
type SquadInfo struct {
	Name        string
	Description string
	PluginName  string
	Agents      []string
	IsPlugin    bool
}

// ListAllSquadInfo returns info about all squads from enabled plugins.
func ListAllSquadInfo() []SquadInfo {
	reg, err := LoadRegistry()
	if err != nil {
		return nil
	}

	var squads []SquadInfo
	for _, plugin := range reg.ListEnabled() {
		for _, name := range plugin.Squads {
			manifest, err := LoadManifest(plugin.Path)
			if err != nil {
				continue
			}

			// Find the squad definition
			for _, def := range manifest.Squads {
				if def.Name == name {
					squads = append(squads, SquadInfo{
						Name:        name,
						Description: def.Description,
						PluginName:  plugin.Name,
						Agents:      def.Agents,
						IsPlugin:    true,
					})
					break
				}
			}
		}
	}
	return squads
}
