package providers

import (
	"context"
	"fmt"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/plugins"
)

// Loader discovers and loads providers from built-ins and plugins.
type Loader struct {
	registry   *Registry
	cfg        config.ProvidersConfig
	pluginsDir string
}

// NewLoader creates a new provider loader.
func NewLoader(registry *Registry, cfg config.ProvidersConfig, pluginsDir string) *Loader {
	return &Loader{
		registry:   registry,
		cfg:        cfg,
		pluginsDir: pluginsDir,
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

	// 2. Load plugin providers
	if err := l.loadPlugins(ctx); err != nil {
		return fmt.Errorf("load plugins: %w", err)
	}

	// 3. Set active providers based on config
	if err := l.setActiveProviders(); err != nil {
		return fmt.Errorf("set active providers: %w", err)
	}

	// 4. Initialize all registered providers
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

// loadPlugins discovers and loads providers from installed plugins.
func (l *Loader) loadPlugins(ctx context.Context) error {
	if l.pluginsDir == "" {
		return nil
	}

	// Get installed plugins from the registry
	pkgRegistry, err := plugins.LoadRegistry()
	if err != nil {
		// No registry file means no plugins installed
		return nil
	}

	for _, pkg := range pkgRegistry.Plugins {
		if pkg.Disabled {
			continue
		}
		if err := l.loadPluginProviders(ctx, pkg); err != nil {
			// Log warning but continue - don't fail entire load for one plugin
			continue
		}
	}

	return nil
}

// loadPluginProviders loads providers from a single plugin.
func (l *Loader) loadPluginProviders(_ context.Context, pkg *plugins.InstalledPlugin) error {
	// Load the plugin's manifest to check for providers
	manifest, err := plugins.LoadManifest(pkg.Path)
	if err != nil {
		return err
	}

	if !manifest.HasProviders() {
		return nil
	}

	for _, providerDef := range manifest.Providers {
		// Convert plugin type to provider type
		pt, err := pluginTypeToProviderType(providerDef.Type)
		if err != nil {
			continue
		}

		// Create a placeholder provider that can be replaced with actual implementation
		// For now, we create stub providers that indicate they came from a plugin
		stub := &PluginStubProvider{
			name:         providerDef.Name,
			providerType: pt,
			pluginName:   pkg.Name,
			entryPoint:   providerDef.EntryPoint,
			config:       providerDef.Config,
		}

		if err := l.registry.Register(stub); err != nil {
			continue
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

// pluginTypeToProviderType converts a plugin type to a provider type.
func pluginTypeToProviderType(pt plugins.PluginType) (ProviderType, error) {
	switch pt {
	case plugins.PluginTypeMemory:
		return ProviderTypeMemory, nil
	case plugins.PluginTypeSandbox:
		return ProviderTypeSandbox, nil
	case plugins.PluginTypeEmbedding:
		return ProviderTypeEmbedding, nil
	case plugins.PluginTypeObserver:
		return ProviderTypeObserver, nil
	default:
		return "", fmt.Errorf("not a provider type: %s", pt)
	}
}

// PluginStubProvider is a placeholder for providers loaded from plugins.
// It implements the Provider interface but delegates actual work to an external process.
type PluginStubProvider struct {
	name         string
	providerType ProviderType
	pluginName   string
	entryPoint   string
	config       map[string]any
	initialized  bool
}

func (p *PluginStubProvider) Name() string       { return p.name }
func (p *PluginStubProvider) Type() ProviderType { return p.providerType }

func (p *PluginStubProvider) Init(_ context.Context, cfg map[string]any) error {
	// Merge configs
	for k, v := range cfg {
		p.config[k] = v
	}
	p.initialized = true
	return nil
}

func (p *PluginStubProvider) Close() error {
	return nil
}

// PluginName returns the name of the plugin this provider came from.
func (p *PluginStubProvider) PluginName() string { return p.pluginName }

// EntryPoint returns the entry point for this provider.
func (p *PluginStubProvider) EntryPoint() string { return p.entryPoint }

// Config returns the provider configuration.
func (p *PluginStubProvider) Config() map[string]any { return p.config }
