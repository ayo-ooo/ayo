package planners

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages the registration and lookup of planner plugins.
// It is thread-safe for concurrent access.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]PlannerFactory
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]PlannerFactory),
	}
}

// Register adds a planner factory to the registry.
// If a planner with the same name already exists, it will be overwritten.
// This is typically called from init() functions in planner packages.
func (r *Registry) Register(name string, factory PlannerFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins[name] = factory
}

// Unregister removes a planner factory from the registry.
// Returns true if the planner was found and removed, false otherwise.
func (r *Registry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.plugins[name]; exists {
		delete(r.plugins, name)
		return true
	}
	return false
}

// Get returns the factory for the given planner name.
// Returns nil and false if no planner with that name is registered.
func (r *Registry) Get(name string) (PlannerFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, ok := r.plugins[name]
	return factory, ok
}

// Has returns true if a planner with the given name is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.plugins[name]
	return ok
}

// List returns the names of all registered planners, sorted alphabetically.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of registered planners.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}

// Instantiate creates a new planner instance using the registered factory.
// Returns an error if no planner with the given name is registered,
// or if the factory fails to create the planner.
func (r *Registry) Instantiate(name string, ctx PlannerContext) (PlannerPlugin, error) {
	factory, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("planner not found: %s", name)
	}

	planner, err := factory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create planner %s: %w", name, err)
	}

	return planner, nil
}

// MustInstantiate is like Instantiate but panics on error.
// Use this only in initialization code where failure should halt the program.
func (r *Registry) MustInstantiate(name string, ctx PlannerContext) PlannerPlugin {
	planner, err := r.Instantiate(name, ctx)
	if err != nil {
		panic(err)
	}
	return planner
}

// DefaultRegistry is the global registry used by ayo.
// Built-in planners register themselves here via init() functions.
// External plugins are loaded into this registry at startup.
var DefaultRegistry = NewRegistry()

// Register is a convenience function that registers with the DefaultRegistry.
func Register(name string, factory PlannerFactory) {
	DefaultRegistry.Register(name, factory)
}

// Get is a convenience function that looks up from the DefaultRegistry.
func Get(name string) (PlannerFactory, bool) {
	return DefaultRegistry.Get(name)
}

// List is a convenience function that lists from the DefaultRegistry.
func List() []string {
	return DefaultRegistry.List()
}

// Instantiate is a convenience function that instantiates from the DefaultRegistry.
func Instantiate(name string, ctx PlannerContext) (PlannerPlugin, error) {
	return DefaultRegistry.Instantiate(name, ctx)
}
