package providers

import (
	"context"
	"fmt"

	"github.com/alexcabrera/ayo/internal/config"
)

// Loader discovers and loads providers from built-ins.
type Loader struct {
	registry *Registry
	cfg      config.ProvidersConfig
}

// NewLoader creates a new provider loader.
func NewLoader(registry *Registry, cfg config.ProvidersConfig) *Loader {
	return &Loader{
		registry: registry,
		cfg:      cfg,
	}
}

// BuiltinProvider is a function that creates a built-in provider.
type BuiltinProvider func() Provider

// builtinProviders maps provider type and name to factory functions.
// These are populated by init() functions in provider implementation packages.
var builtinProviders = make(map[string]BuiltinProvider)

// RegisterBuiltin registers a built-in provider factory.
// Key format is "type.name" (e.g., "memory.zettelkasten").
func RegisterBuiltin(providerType ProviderType, name string, factory BuiltinProvider) {
	key := string(providerType) + "." + name
	builtinProviders[key] = factory
}

// LoadAll discovers and loads all providers, then sets active providers
// based on configuration.
func (l *Loader) LoadAll(ctx context.Context) error {
	// 1. Load built-in providers
	if err := l.loadBuiltins(ctx); err != nil {
		return fmt.Errorf("load builtins: %w", err)
	}

	// 2. Set active providers based on config
	if err := l.setActiveProviders(); err != nil {
		return fmt.Errorf("set active providers: %w", err)
	}

	// 3. Initialize all registered providers
	configs := l.buildProviderConfigs()
	if err := l.registry.InitAll(ctx, configs); err != nil {
		return fmt.Errorf("init providers: %w", err)
	}

	return nil
}

// loadBuiltins registers all built-in providers.
func (l *Loader) loadBuiltins(_ context.Context) error {
	for key, factory := range builtinProviders {
		p := factory()
		if p == nil {
			return fmt.Errorf("built-in provider factory returned nil: %s", key)
		}
		if err := l.registry.Register(p); err != nil {
			return fmt.Errorf("register built-in %s: %w", key, err)
		}
	}
	return nil
}

// setActiveProviders sets the active provider for each type based on config.
func (l *Loader) setActiveProviders() error {
	if l.cfg.Active == nil {
		return nil
	}

	for typeStr, name := range l.cfg.Active {
		pt := ProviderType(typeStr)

		// Validate the type is known
		switch pt {
		case ProviderTypeMemory, ProviderTypeSandbox, ProviderTypeEmbedding, ProviderTypeObserver:
			// Valid type
		default:
			continue // Skip unknown types
		}

		// Try to set the active provider (may not exist yet if plugin not loaded)
		if err := l.registry.SetActive(pt, name); err != nil {
			// Provider not found - this is not an error if we're using defaults
			continue
		}
	}

	return nil
}

// buildProviderConfigs builds the config map for provider initialization.
func (l *Loader) buildProviderConfigs() map[string]map[string]any {
	configs := make(map[string]map[string]any)

	// Memory provider config
	if l.cfg.Memory.Directory != "" || l.cfg.Memory.AutoMerge {
		configs["memory.zettelkasten"] = map[string]any{
			"directory":  l.cfg.Memory.Directory,
			"index_path": l.cfg.Memory.IndexPath,
			"auto_merge": l.cfg.Memory.AutoMerge,
		}
	}

	// Sandbox provider config
	if l.cfg.Sandbox.Image != "" {
		sandboxConfig := map[string]any{
			"image": l.cfg.Sandbox.Image,
		}
		if l.cfg.Sandbox.Pool.MinSize > 0 || l.cfg.Sandbox.Pool.MaxSize > 0 {
			sandboxConfig["min_size"] = l.cfg.Sandbox.Pool.MinSize
			sandboxConfig["max_size"] = l.cfg.Sandbox.Pool.MaxSize
			sandboxConfig["idle_timeout"] = l.cfg.Sandbox.Pool.IdleTimeout
		}
		if l.cfg.Sandbox.Resources.CPUs > 0 {
			sandboxConfig["cpus"] = l.cfg.Sandbox.Resources.CPUs
			sandboxConfig["memory_mb"] = l.cfg.Sandbox.Resources.MemoryMB
			sandboxConfig["disk_mb"] = l.cfg.Sandbox.Resources.DiskMB
		}
		if l.cfg.Sandbox.Network.Enabled != nil {
			sandboxConfig["network_enabled"] = *l.cfg.Sandbox.Network.Enabled
			sandboxConfig["dns"] = l.cfg.Sandbox.Network.DNS
		}
		configs["sandbox.apple-container"] = sandboxConfig
		configs["sandbox.none"] = sandboxConfig
	}

	// Embedding provider config
	if l.cfg.Embedding.Model != "" {
		configs["embedding.ollama"] = map[string]any{
			"model":      l.cfg.Embedding.Model,
			"endpoint":   l.cfg.Embedding.Endpoint,
			"dimensions": l.cfg.Embedding.Dimensions,
		}
	}

	// Observer provider config
	if l.cfg.Observer.Enabled != nil {
		configs["observer.memory-extractor"] = map[string]any{
			"enabled":     *l.cfg.Observer.Enabled,
			"batch_size":  l.cfg.Observer.BatchSize,
			"debounce_ms": l.cfg.Observer.DebounceMs,
		}
	}

	return configs
}
