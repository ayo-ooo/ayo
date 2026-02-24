package triggers

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages registered trigger types.
type Registry struct {
	mu       sync.RWMutex
	triggers map[string]*registeredTrigger
}

type registeredTrigger struct {
	factory    TriggerFactory
	pluginName string
}

// DefaultRegistry is the global trigger registry.
var DefaultRegistry = NewRegistry()

// NewRegistry creates a new trigger registry.
func NewRegistry() *Registry {
	return &Registry{
		triggers: make(map[string]*registeredTrigger),
	}
}

// Register adds a trigger type to the registry.
// pluginName should be empty for built-in triggers.
func (r *Registry) Register(name string, factory TriggerFactory, pluginName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.triggers[name]; exists {
		return fmt.Errorf("trigger type %q already registered", name)
	}

	r.triggers[name] = &registeredTrigger{
		factory:    factory,
		pluginName: pluginName,
	}
	return nil
}

// Unregister removes a trigger type from the registry.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.triggers[name]; !exists {
		return fmt.Errorf("trigger type %q not registered", name)
	}

	delete(r.triggers, name)
	return nil
}

// Create creates a new instance of a trigger plugin by name.
func (r *Registry) Create(name string) (TriggerPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rt, exists := r.triggers[name]
	if !exists {
		return nil, fmt.Errorf("trigger type %q not found", name)
	}

	return rt.factory(), nil
}

// Has returns true if a trigger type is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.triggers[name]
	return exists
}

// List returns information about all registered trigger types.
func (r *Registry) List() []TriggerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var infos []TriggerInfo
	for name, rt := range r.triggers {
		plugin := rt.factory()
		infos = append(infos, TriggerInfo{
			Name:         name,
			Category:     plugin.Category(),
			Description:  plugin.Description(),
			PluginName:   rt.pluginName,
			ConfigSchema: plugin.ConfigSchema(),
		})
	}

	// Sort by name for consistent output
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos
}

// ListNames returns the names of all registered trigger types.
func (r *Registry) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.triggers))
	for name := range r.triggers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetPluginName returns the plugin name for a trigger type.
// Returns empty string for built-in triggers.
func (r *Registry) GetPluginName(name string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if rt, exists := r.triggers[name]; exists {
		return rt.pluginName
	}
	return ""
}

// Clear removes all registered trigger types (for testing).
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.triggers = make(map[string]*registeredTrigger)
}

// IsBuiltin returns true if the trigger type is built-in (not from a plugin).
func (r *Registry) IsBuiltin(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if rt, exists := r.triggers[name]; exists {
		return rt.pluginName == ""
	}
	return false
}
