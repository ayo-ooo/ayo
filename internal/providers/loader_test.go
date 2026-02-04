package providers

import (
	"context"
	"testing"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/plugins"
)

// Ensure plugins import is used
var _ = plugins.PluginTypeMemory

// simpleProvider is a minimal Provider implementation for testing.
type simpleProvider struct {
	name         string
	providerType ProviderType
}

func (p *simpleProvider) Name() string                                     { return p.name }
func (p *simpleProvider) Type() ProviderType                               { return p.providerType }
func (p *simpleProvider) Init(_ context.Context, _ map[string]any) error   { return nil }
func (p *simpleProvider) Close() error                                     { return nil }

func TestLoaderLoadBuiltins(t *testing.T) {
	// Register a test built-in provider
	testFactory := func() Provider {
		return &simpleProvider{
			name:         "test-memory",
			providerType: ProviderTypeMemory,
		}
	}
	RegisterBuiltin(ProviderTypeMemory, "test-memory", testFactory)
	defer func() {
		// Cleanup
		delete(builtinProviders, "memory.test-memory")
	}()

	registry := NewRegistry()
	loader := NewLoader(registry, config.DefaultProvidersConfig(), "")

	// Load just built-ins
	if err := loader.loadBuiltins(context.Background()); err != nil {
		t.Fatalf("loadBuiltins failed: %v", err)
	}

	// Check the provider was registered
	p, err := registry.Get(ProviderTypeMemory, "test-memory")
	if err != nil {
		t.Fatalf("provider not registered: %v", err)
	}
	if p.Name() != "test-memory" {
		t.Errorf("expected name 'test-memory', got %q", p.Name())
	}
}

func TestLoaderSetActiveProviders(t *testing.T) {
	registry := NewRegistry()

	// Register some providers
	registry.Register(&simpleProvider{name: "mem1", providerType: ProviderTypeMemory})
	registry.Register(&simpleProvider{name: "mem2", providerType: ProviderTypeMemory})
	registry.Register(&simpleProvider{name: "sand1", providerType: ProviderTypeSandbox})

	cfg := config.ProvidersConfig{
		Active: map[string]string{
			"memory":  "mem2",
			"sandbox": "sand1",
		},
	}

	loader := NewLoader(registry, cfg, "")
	if err := loader.setActiveProviders(); err != nil {
		t.Fatalf("setActiveProviders failed: %v", err)
	}

	// Check active providers
	if registry.ActiveName(ProviderTypeMemory) != "mem2" {
		t.Errorf("expected active memory 'mem2', got %q", registry.ActiveName(ProviderTypeMemory))
	}
	if registry.ActiveName(ProviderTypeSandbox) != "sand1" {
		t.Errorf("expected active sandbox 'sand1', got %q", registry.ActiveName(ProviderTypeSandbox))
	}
}

func TestLoaderBuildProviderConfigs(t *testing.T) {
	enabled := true
	cfg := config.ProvidersConfig{
		Memory: config.MemoryProviderConfig{
			Directory: "/custom/memory",
			AutoMerge: true,
		},
		Sandbox: config.SandboxProviderConfig{
			Image: "custom:latest",
			Pool: config.SandboxPoolConfig{
				MinSize:     3,
				MaxSize:     6,
				IdleTimeout: "1h",
			},
			Resources: config.SandboxResourcesConfig{
				CPUs:     4,
				MemoryMB: 4096,
			},
			Network: config.SandboxNetworkConfig{
				Enabled: &enabled,
			},
		},
		Embedding: config.EmbeddingProviderConfig{
			Model: "custom-embed",
		},
		Observer: config.ObserverProviderConfig{
			Enabled:    &enabled,
			BatchSize:  20,
			DebounceMs: 500,
		},
	}

	loader := NewLoader(nil, cfg, "")
	configs := loader.buildProviderConfigs()

	// Check memory config
	memCfg := configs["memory.zettelkasten"]
	if memCfg == nil {
		t.Fatal("missing memory.zettelkasten config")
	}
	if memCfg["directory"] != "/custom/memory" {
		t.Errorf("expected directory '/custom/memory', got %v", memCfg["directory"])
	}
	if memCfg["auto_merge"] != true {
		t.Errorf("expected auto_merge true, got %v", memCfg["auto_merge"])
	}

	// Check sandbox config
	sandCfg := configs["sandbox.apple-container"]
	if sandCfg == nil {
		t.Fatal("missing sandbox.apple-container config")
	}
	if sandCfg["image"] != "custom:latest" {
		t.Errorf("expected image 'custom:latest', got %v", sandCfg["image"])
	}
	if sandCfg["min_size"] != 3 {
		t.Errorf("expected min_size 3, got %v", sandCfg["min_size"])
	}
	if sandCfg["max_size"] != 6 {
		t.Errorf("expected max_size 6, got %v", sandCfg["max_size"])
	}

	// Check embedding config
	embedCfg := configs["embedding.ollama"]
	if embedCfg == nil {
		t.Fatal("missing embedding.ollama config")
	}
	if embedCfg["model"] != "custom-embed" {
		t.Errorf("expected model 'custom-embed', got %v", embedCfg["model"])
	}

	// Check observer config
	obsCfg := configs["observer.memory-extractor"]
	if obsCfg == nil {
		t.Fatal("missing observer.memory-extractor config")
	}
	if obsCfg["enabled"] != true {
		t.Errorf("expected enabled true, got %v", obsCfg["enabled"])
	}
	if obsCfg["batch_size"] != 20 {
		t.Errorf("expected batch_size 20, got %v", obsCfg["batch_size"])
	}
}

func TestPluginTypeToProviderType(t *testing.T) {
	tests := []struct {
		input   string
		want    ProviderType
		wantErr bool
	}{
		{"memory", ProviderTypeMemory, false},
		{"sandbox", ProviderTypeSandbox, false},
		{"embedding", ProviderTypeEmbedding, false},
		{"observer", ProviderTypeObserver, false},
		{"agent", "", true},
		{"skill", "", true},
		{"tool", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Use the internal function via exported test helper
			pt := pluginTypeFromString(tt.input)
			got, err := pluginTypeToProviderType(pt)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("got %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestPluginStubProvider(t *testing.T) {
	stub := &PluginStubProvider{
		name:         "test-stub",
		providerType: ProviderTypeMemory,
		pluginName:   "test-plugin",
		entryPoint:   "bin/provider",
		config:       map[string]any{"key": "value"},
	}

	if stub.Name() != "test-stub" {
		t.Errorf("Name() = %q, want 'test-stub'", stub.Name())
	}
	if stub.Type() != ProviderTypeMemory {
		t.Errorf("Type() = %q, want 'memory'", stub.Type())
	}
	if stub.PluginName() != "test-plugin" {
		t.Errorf("PluginName() = %q, want 'test-plugin'", stub.PluginName())
	}
	if stub.EntryPoint() != "bin/provider" {
		t.Errorf("EntryPoint() = %q, want 'bin/provider'", stub.EntryPoint())
	}

	// Test Init
	if err := stub.Init(context.Background(), map[string]any{"new": "config"}); err != nil {
		t.Errorf("Init failed: %v", err)
	}
	if !stub.initialized {
		t.Error("expected initialized to be true")
	}
	if stub.config["new"] != "config" {
		t.Errorf("expected merged config, got %v", stub.config)
	}

	// Test Close
	if err := stub.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// pluginTypeFromString converts a string to plugins.PluginType for testing.
func pluginTypeFromString(s string) plugins.PluginType {
	switch s {
	case "memory":
		return plugins.PluginTypeMemory
	case "sandbox":
		return plugins.PluginTypeSandbox
	case "embedding":
		return plugins.PluginTypeEmbedding
	case "observer":
		return plugins.PluginTypeObserver
	default:
		return plugins.PluginType(s)
	}
}

func TestRegisterBuiltin(t *testing.T) {
	// Test that RegisterBuiltin properly stores factories
	testFactory := func() Provider {
		return &simpleProvider{name: "builtin-test", providerType: ProviderTypeSandbox}
	}

	// Register
	RegisterBuiltin(ProviderTypeSandbox, "builtin-test", testFactory)
	defer delete(builtinProviders, "sandbox.builtin-test")

	// Check it's in the map
	key := "sandbox.builtin-test"
	if _, ok := builtinProviders[key]; !ok {
		t.Errorf("expected %q to be registered", key)
	}

	// Call the factory
	p := builtinProviders[key]()
	if p.Name() != "builtin-test" {
		t.Errorf("factory returned wrong name: %q", p.Name())
	}
}
