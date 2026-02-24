package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/alexcabrera/ayo/internal/planners"
)

// PlannerLoadError represents an error loading a planner plugin.
type PlannerLoadError struct {
	PluginName  string
	PlannerName string
	Err         error
}

func (e *PlannerLoadError) Error() string {
	return fmt.Sprintf("failed to load planner %s from plugin %s: %v", e.PlannerName, e.PluginName, e.Err)
}

func (e *PlannerLoadError) Unwrap() error {
	return e.Err
}

// LoadedPlanner contains information about a successfully loaded planner.
type LoadedPlanner struct {
	Name       string
	PluginName string
	Type       PlannerType
}

// LoadPlanners loads all planner plugins from installed plugins and registers
// them with the planner registry. This should be called during application startup.
//
// Returns a list of successfully loaded planners and any errors encountered.
// Loading continues even if some plugins fail to load.
func LoadPlanners() ([]LoadedPlanner, []error) {
	reg, err := LoadRegistry()
	if err != nil {
		return nil, []error{fmt.Errorf("load plugin registry: %w", err)}
	}

	var loaded []LoadedPlanner
	var errors []error

	for _, plugin := range reg.ListEnabled() {
		manifest, err := LoadManifest(plugin.Path)
		if err != nil {
			errors = append(errors, &PlannerLoadError{
				PluginName:  plugin.Name,
				PlannerName: "",
				Err:         fmt.Errorf("load manifest: %w", err),
			})
			continue
		}

		for _, plannerDef := range manifest.Planners {
			if err := loadPlanner(plugin, plannerDef); err != nil {
				errors = append(errors, &PlannerLoadError{
					PluginName:  plugin.Name,
					PlannerName: plannerDef.Name,
					Err:         err,
				})
				continue
			}

			loaded = append(loaded, LoadedPlanner{
				Name:       plannerDef.Name,
				PluginName: plugin.Name,
				Type:       plannerDef.Type,
			})
		}
	}

	return loaded, errors
}

// loadPlanner loads a single planner from a plugin and registers it.
func loadPlanner(installedPlugin *InstalledPlugin, def PlannerDef) error {
	// Validate planner type
	if !ValidPlannerType(def.Type) {
		return fmt.Errorf("invalid planner type: %s", def.Type)
	}

	// Check if already registered (built-in takes precedence)
	if planners.DefaultRegistry.Has(def.Name) {
		return fmt.Errorf("planner already registered (built-in or another plugin)")
	}

	// For plugins with an entry point, load the Go plugin.
	if def.EntryPoint != "" {
		entryPath := filepath.Join(installedPlugin.Path, def.EntryPoint)
		if _, err := os.Stat(entryPath); err != nil {
			return fmt.Errorf("entry point not found: %s", def.EntryPoint)
		}

		// Load the Go plugin (.so file)
		factory, err := loadExternalPlanner(entryPath)
		if err != nil {
			return fmt.Errorf("load external planner: %w", err)
		}

		// Register the planner
		planners.DefaultRegistry.Register(def.Name, factory)
		return nil
	}

	// If no entry point, this is a built-in planner that should already be registered.
	// We skip registration as built-in planners register themselves via init().
	return nil
}

// loadExternalPlanner loads a planner plugin from a .so file and returns the factory.
func loadExternalPlanner(path string) (planners.PlannerFactory, error) {
	// Validate file extension
	if !strings.HasSuffix(path, ".so") {
		return nil, fmt.Errorf("external planner must be a .so file, got: %s", filepath.Base(path))
	}

	// Open the Go plugin
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open plugin: %w", err)
	}

	// Look for the NewPlanner symbol
	sym, err := p.Lookup("NewPlanner")
	if err != nil {
		return nil, fmt.Errorf("lookup NewPlanner symbol: %w (plugin must export 'NewPlanner' as planners.PlannerFactory)", err)
	}

	// Assert it's the right type
	factory, ok := sym.(*planners.PlannerFactory)
	if ok {
		return *factory, nil
	}

	// Try as function directly
	factoryFn, ok := sym.(func(planners.PlannerContext) (planners.PlannerPlugin, error))
	if ok {
		return factoryFn, nil
	}

	return nil, fmt.Errorf("NewPlanner has wrong signature: expected planners.PlannerFactory")
}

// ListPluginPlanners returns all planner definitions from enabled plugins.
// This is useful for displaying what planners are available from plugins
// without necessarily loading them.
func ListPluginPlanners() ([]PlannerDef, error) {
	reg, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load plugin registry: %w", err)
	}

	var allPlanners []PlannerDef

	for _, plugin := range reg.ListEnabled() {
		manifest, err := LoadManifest(plugin.Path)
		if err != nil {
			continue // Skip plugins with invalid manifests
		}

		allPlanners = append(allPlanners, manifest.Planners...)
	}

	return allPlanners, nil
}

// GetPlannerPlugin returns the plugin that provides a given planner name.
// Returns nil if the planner is not from a plugin (e.g., built-in).
func GetPlannerPlugin(plannerName string) (*InstalledPlugin, error) {
	reg, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load plugin registry: %w", err)
	}

	for _, plugin := range reg.ListEnabled() {
		manifest, err := LoadManifest(plugin.Path)
		if err != nil {
			continue
		}

		for _, plannerDef := range manifest.Planners {
			if plannerDef.Name == plannerName {
				return plugin, nil
			}
		}
	}

	return nil, nil // Not found in any plugin
}
