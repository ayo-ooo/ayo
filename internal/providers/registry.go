package providers

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages provider instances by type.
// It supports built-in defaults.
type Registry struct {
	mu        sync.RWMutex
	providers map[ProviderType]map[string]Provider
	active    map[ProviderType]string // Currently active provider name per type
	defaults  map[ProviderType]string // Built-in default names per type
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[ProviderType]map[string]Provider),
		active:    make(map[ProviderType]string),
		defaults: map[ProviderType]string{
			ProviderTypeMemory:    "zettelkasten",
			ProviderTypeSandbox:   "none",
			ProviderTypeEmbedding: "ollama",
			ProviderTypeObserver:  "memory-extractor",
		},
	}
}

// Register adds a provider to the registry.
// If a provider with the same name and type exists, it will be replaced.
func (r *Registry) Register(p Provider) error {
	if p == nil {
		return fmt.Errorf("cannot register nil provider")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	pt := p.Type()
	if r.providers[pt] == nil {
		r.providers[pt] = make(map[string]Provider)
	}

	r.providers[pt][p.Name()] = p
	return nil
}

// Unregister removes a provider from the registry.
func (r *Registry) Unregister(pt ProviderType, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.providers[pt] == nil {
		return fmt.Errorf("no providers of type %s registered", pt)
	}

	if _, ok := r.providers[pt][name]; !ok {
		return fmt.Errorf("provider %s of type %s not found", name, pt)
	}

	delete(r.providers[pt], name)

	// If this was the active provider, clear it
	if r.active[pt] == name {
		delete(r.active, pt)
	}

	return nil
}

// Get retrieves a specific provider by type and name.
func (r *Registry) Get(pt ProviderType, name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.providers[pt] == nil {
		return nil, fmt.Errorf("no providers of type %s registered", pt)
	}

	p, ok := r.providers[pt][name]
	if !ok {
		return nil, fmt.Errorf("provider %s of type %s not found", name, pt)
	}

	return p, nil
}

// GetActive returns the currently active provider for a type.
// If no provider is explicitly set, returns the default.
// If no default exists, returns nil.
func (r *Registry) GetActive(pt ProviderType) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check for explicitly set active provider
	if name, ok := r.active[pt]; ok {
		if p, ok := r.providers[pt][name]; ok {
			return p
		}
	}

	// Fall back to default
	if name, ok := r.defaults[pt]; ok {
		if r.providers[pt] != nil {
			if p, ok := r.providers[pt][name]; ok {
				return p
			}
		}
	}

	return nil
}

// SetActive sets the active provider for a type.
func (r *Registry) SetActive(pt ProviderType, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify provider exists
	if r.providers[pt] == nil || r.providers[pt][name] == nil {
		return fmt.Errorf("provider %s of type %s not found", name, pt)
	}

	r.active[pt] = name
	return nil
}

// SetDefault sets the default provider for a type.
// This is typically called during built-in registration.
func (r *Registry) SetDefault(pt ProviderType, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaults[pt] = name
}

// List returns all registered providers of a given type.
func (r *Registry) List(pt ProviderType) []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.providers[pt] == nil {
		return nil
	}

	result := make([]Provider, 0, len(r.providers[pt]))
	for _, p := range r.providers[pt] {
		result = append(result, p)
	}
	return result
}

// ListAll returns all registered providers across all types.
func (r *Registry) ListAll() map[ProviderType][]Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[ProviderType][]Provider)
	for pt, providers := range r.providers {
		for _, p := range providers {
			result[pt] = append(result[pt], p)
		}
	}
	return result
}

// Names returns the names of all registered providers of a given type.
func (r *Registry) Names(pt ProviderType) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.providers[pt] == nil {
		return nil
	}

	result := make([]string, 0, len(r.providers[pt]))
	for name := range r.providers[pt] {
		result = append(result, name)
	}
	return result
}

// ActiveName returns the name of the active provider for a type.
// Returns the default name if no explicit active is set.
// Returns empty string if nothing is configured.
func (r *Registry) ActiveName(pt ProviderType) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if name, ok := r.active[pt]; ok {
		return name
	}
	if name, ok := r.defaults[pt]; ok {
		return name
	}
	return ""
}

// DefaultName returns the default provider name for a type.
func (r *Registry) DefaultName(pt ProviderType) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaults[pt]
}

// InitAll initializes all registered providers with their configs.
// The configs map is keyed by "type.name" (e.g., "memory.zettelkasten").
func (r *Registry) InitAll(ctx context.Context, configs map[string]map[string]any) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for pt, providers := range r.providers {
		for name, p := range providers {
			key := string(pt) + "." + name
			config := configs[key]
			if config == nil {
				config = make(map[string]any)
			}
			if err := p.Init(ctx, config); err != nil {
				return fmt.Errorf("failed to init provider %s.%s: %w", pt, name, err)
			}
		}
	}
	return nil
}

// CloseAll closes all registered providers.
func (r *Registry) CloseAll() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var errs []error
	for _, providers := range r.providers {
		for name, p := range providers {
			if err := p.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close %s: %w", name, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}
	return nil
}

// Memory returns the active MemoryProvider, or nil if none.
func (r *Registry) Memory() MemoryProvider {
	p := r.GetActive(ProviderTypeMemory)
	if p == nil {
		return nil
	}
	mp, ok := p.(MemoryProvider)
	if !ok {
		return nil
	}
	return mp
}

// Sandbox returns the active SandboxProvider, or nil if none.
func (r *Registry) Sandbox() SandboxProvider {
	p := r.GetActive(ProviderTypeSandbox)
	if p == nil {
		return nil
	}
	sp, ok := p.(SandboxProvider)
	if !ok {
		return nil
	}
	return sp
}

// Embedding returns the active EmbeddingProvider, or nil if none.
func (r *Registry) Embedding() EmbeddingProvider {
	p := r.GetActive(ProviderTypeEmbedding)
	if p == nil {
		return nil
	}
	ep, ok := p.(EmbeddingProvider)
	if !ok {
		return nil
	}
	return ep
}

// Observer returns the active ObserverProvider, or nil if none.
func (r *Registry) Observer() ObserverProvider {
	p := r.GetActive(ProviderTypeObserver)
	if p == nil {
		return nil
	}
	op, ok := p.(ObserverProvider)
	if !ok {
		return nil
	}
	return op
}
