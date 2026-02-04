package providers

import (
	"context"
	"fmt"
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	mp := NewMockMemoryProvider("test-memory")

	// Test Register
	if err := r.Register(mp); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Test Get
	got, err := r.Get(ProviderTypeMemory, "test-memory")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name() != "test-memory" {
		t.Errorf("Get() name = %v, want test-memory", got.Name())
	}

	// Test Get non-existent
	_, err = r.Get(ProviderTypeMemory, "non-existent")
	if err == nil {
		t.Error("Get() should return error for non-existent provider")
	}
}

func TestRegistry_RegisterNil(t *testing.T) {
	r := NewRegistry()
	err := r.Register(nil)
	if err == nil {
		t.Error("Register(nil) should return error")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()
	mp := NewMockMemoryProvider("test-memory")
	r.Register(mp)

	// Test Unregister
	if err := r.Unregister(ProviderTypeMemory, "test-memory"); err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}

	// Verify it's gone
	_, err := r.Get(ProviderTypeMemory, "test-memory")
	if err == nil {
		t.Error("Get() should return error after Unregister")
	}

	// Test Unregister non-existent
	err = r.Unregister(ProviderTypeMemory, "non-existent")
	if err == nil {
		t.Error("Unregister() should return error for non-existent provider")
	}
}

func TestRegistry_Active(t *testing.T) {
	r := NewRegistry()
	mp1 := NewMockMemoryProvider("memory-1")
	mp2 := NewMockMemoryProvider("memory-2")

	r.Register(mp1)
	r.Register(mp2)

	// No active set, should return nil (defaults not registered)
	if got := r.GetActive(ProviderTypeMemory); got != nil {
		t.Errorf("GetActive() without set = %v, want nil", got)
	}

	// Set active
	if err := r.SetActive(ProviderTypeMemory, "memory-1"); err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}

	got := r.GetActive(ProviderTypeMemory)
	if got == nil {
		t.Fatal("GetActive() = nil after SetActive")
	}
	if got.Name() != "memory-1" {
		t.Errorf("GetActive() name = %v, want memory-1", got.Name())
	}

	// Test ActiveName
	if name := r.ActiveName(ProviderTypeMemory); name != "memory-1" {
		t.Errorf("ActiveName() = %v, want memory-1", name)
	}

	// Set non-existent active
	err := r.SetActive(ProviderTypeMemory, "non-existent")
	if err == nil {
		t.Error("SetActive() should return error for non-existent provider")
	}
}

func TestRegistry_Defaults(t *testing.T) {
	r := NewRegistry()

	// Check built-in defaults
	defaults := map[ProviderType]string{
		ProviderTypeMemory:    "zettelkasten",
		ProviderTypeSandbox:   "none",
		ProviderTypeEmbedding: "ollama",
		ProviderTypeObserver:  "memory-extractor",
	}

	for pt, want := range defaults {
		if got := r.DefaultName(pt); got != want {
			t.Errorf("DefaultName(%s) = %v, want %v", pt, got, want)
		}
	}

	// SetDefault and verify
	r.SetDefault(ProviderTypeMemory, "custom-memory")
	if got := r.DefaultName(ProviderTypeMemory); got != "custom-memory" {
		t.Errorf("DefaultName() after SetDefault = %v, want custom-memory", got)
	}
}

func TestRegistry_DefaultFallback(t *testing.T) {
	r := NewRegistry()

	// Register provider with default name
	mp := NewMockMemoryProvider("zettelkasten")
	r.Register(mp)

	// Without explicit SetActive, should get default
	got := r.GetActive(ProviderTypeMemory)
	if got == nil {
		t.Fatal("GetActive() = nil, want default provider")
	}
	if got.Name() != "zettelkasten" {
		t.Errorf("GetActive() name = %v, want zettelkasten", got.Name())
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	mp1 := NewMockMemoryProvider("memory-1")
	mp2 := NewMockMemoryProvider("memory-2")
	sp := NewMockSandboxProvider("sandbox-1")

	r.Register(mp1)
	r.Register(mp2)
	r.Register(sp)

	// Test List by type
	memProviders := r.List(ProviderTypeMemory)
	if len(memProviders) != 2 {
		t.Errorf("List(memory) len = %d, want 2", len(memProviders))
	}

	sandboxProviders := r.List(ProviderTypeSandbox)
	if len(sandboxProviders) != 1 {
		t.Errorf("List(sandbox) len = %d, want 1", len(sandboxProviders))
	}

	// Test Names
	names := r.Names(ProviderTypeMemory)
	if len(names) != 2 {
		t.Errorf("Names() len = %d, want 2", len(names))
	}

	// Test ListAll
	all := r.ListAll()
	if len(all[ProviderTypeMemory]) != 2 {
		t.Errorf("ListAll()[memory] len = %d, want 2", len(all[ProviderTypeMemory]))
	}
	if len(all[ProviderTypeSandbox]) != 1 {
		t.Errorf("ListAll()[sandbox] len = %d, want 1", len(all[ProviderTypeSandbox]))
	}
}

func TestRegistry_TypedAccessors(t *testing.T) {
	r := NewRegistry()
	mp := NewMockMemoryProvider("zettelkasten")
	sp := NewMockSandboxProvider("none")
	ep := NewMockEmbeddingProvider("ollama", 384)

	r.Register(mp)
	r.Register(sp)
	r.Register(ep)

	// Test Memory()
	if got := r.Memory(); got == nil {
		t.Error("Memory() = nil, want provider")
	}

	// Test Sandbox()
	if got := r.Sandbox(); got == nil {
		t.Error("Sandbox() = nil, want provider")
	}

	// Test Embedding()
	if got := r.Embedding(); got == nil {
		t.Error("Embedding() = nil, want provider")
	}

	// Test Observer() - none registered
	if got := r.Observer(); got != nil {
		t.Errorf("Observer() = %v, want nil", got)
	}
}

func TestRegistry_InitAll(t *testing.T) {
	r := NewRegistry()
	mp := NewMockMemoryProvider("test-memory")
	sp := NewMockSandboxProvider("test-sandbox")

	r.Register(mp)
	r.Register(sp)

	ctx := context.Background()
	configs := map[string]map[string]any{
		"memory.test-memory": {"setting": "value"},
		"sandbox.test-sandbox": {"limit": 100},
	}

	if err := r.InitAll(ctx, configs); err != nil {
		t.Fatalf("InitAll() error = %v", err)
	}
}

func TestRegistry_CloseAll(t *testing.T) {
	r := NewRegistry()
	mp := NewMockMemoryProvider("test-memory")
	sp := NewMockSandboxProvider("test-sandbox")

	r.Register(mp)
	r.Register(sp)

	if err := r.CloseAll(); err != nil {
		t.Fatalf("CloseAll() error = %v", err)
	}
}

func TestRegistry_Concurrency(t *testing.T) {
	r := NewRegistry()

	// Concurrent registration
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			mp := NewMockMemoryProvider(fmt.Sprintf("memory-%d", n))
			r.Register(mp)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all registered
	if len(r.List(ProviderTypeMemory)) != 10 {
		t.Errorf("Concurrent registration: got %d providers, want 10", len(r.List(ProviderTypeMemory)))
	}
}

func TestRegistry_UnregisterClearsActive(t *testing.T) {
	r := NewRegistry()
	mp := NewMockMemoryProvider("test-memory")

	r.Register(mp)
	r.SetActive(ProviderTypeMemory, "test-memory")

	// Verify active is set
	if r.ActiveName(ProviderTypeMemory) != "test-memory" {
		t.Fatal("ActiveName() should be test-memory")
	}

	// Unregister
	r.Unregister(ProviderTypeMemory, "test-memory")

	// Active should be cleared (fall back to default)
	if r.ActiveName(ProviderTypeMemory) != "zettelkasten" {
		t.Errorf("ActiveName() after unregister = %v, want zettelkasten (default)", r.ActiveName(ProviderTypeMemory))
	}
}

// MockObserverProvider is a test implementation of ObserverProvider.
type MockObserverProvider struct {
	name      string
	started   bool
	messages  []MessageEvent
}

func NewMockObserverProvider(name string) *MockObserverProvider {
	return &MockObserverProvider{name: name}
}

func (m *MockObserverProvider) Name() string            { return m.name }
func (m *MockObserverProvider) Type() ProviderType     { return ProviderTypeObserver }
func (m *MockObserverProvider) Init(ctx context.Context, config map[string]any) error { return nil }
func (m *MockObserverProvider) Close() error           { return nil }

func (m *MockObserverProvider) Start(ctx context.Context) error {
	m.started = true
	return nil
}

func (m *MockObserverProvider) Stop() error {
	m.started = false
	return nil
}

func (m *MockObserverProvider) OnMessage(ctx context.Context, event MessageEvent) error {
	m.messages = append(m.messages, event)
	return nil
}

func TestRegistry_ObserverAccessor(t *testing.T) {
	r := NewRegistry()

	// No provider registered yet
	if ob := r.Observer(); ob != nil {
		t.Error("Observer() should return nil when no observer registered")
	}

	// Register observer
	op := NewMockObserverProvider("test-observer")
	r.Register(op)
	r.SetActive(ProviderTypeObserver, "test-observer")

	// Now should return the observer
	got := r.Observer()
	if got == nil {
		t.Fatal("Observer() returned nil after registration")
	}
	if got.Name() != "test-observer" {
		t.Errorf("Observer().Name() = %q, want 'test-observer'", got.Name())
	}
}

func TestRegistry_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	r := NewRegistry()

	// Register multiple provider types
	mem := NewMockMemoryProvider("test-mem")
	sand := NewMockSandboxProvider("test-sand")
	embed := NewMockEmbeddingProvider("test-embed", 256)
	obs := NewMockObserverProvider("test-obs")

	r.Register(mem)
	r.Register(sand)
	r.Register(embed)
	r.Register(obs)

	// Set all active
	r.SetActive(ProviderTypeMemory, "test-mem")
	r.SetActive(ProviderTypeSandbox, "test-sand")
	r.SetActive(ProviderTypeEmbedding, "test-embed")
	r.SetActive(ProviderTypeObserver, "test-obs")

	// Initialize all
	configs := map[string]map[string]any{
		"memory.test-mem": {"key": "value"},
	}
	if err := r.InitAll(ctx, configs); err != nil {
		t.Fatalf("InitAll() error = %v", err)
	}

	// Verify all accessible
	if r.Memory() == nil {
		t.Error("Memory() returned nil")
	}
	if r.Sandbox() == nil {
		t.Error("Sandbox() returned nil")
	}
	if r.Embedding() == nil {
		t.Error("Embedding() returned nil")
	}
	if r.Observer() == nil {
		t.Error("Observer() returned nil")
	}

	// Close all
	if err := r.CloseAll(); err != nil {
		t.Fatalf("CloseAll() error = %v", err)
	}
}

func TestRegistry_ListAll(t *testing.T) {
	r := NewRegistry()

	mem1 := NewMockMemoryProvider("mem1")
	mem2 := NewMockMemoryProvider("mem2")
	sand := NewMockSandboxProvider("sand1")

	r.Register(mem1)
	r.Register(mem2)
	r.Register(sand)

	all := r.ListAll()

	if len(all[ProviderTypeMemory]) != 2 {
		t.Errorf("ListAll() memory count = %d, want 2", len(all[ProviderTypeMemory]))
	}
	if len(all[ProviderTypeSandbox]) != 1 {
		t.Errorf("ListAll() sandbox count = %d, want 1", len(all[ProviderTypeSandbox]))
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()

	// No providers - should return nil
	names := r.Names(ProviderTypeMemory)
	if names != nil {
		t.Errorf("Names() with no providers should return nil, got %v", names)
	}

	// Register providers
	r.Register(NewMockMemoryProvider("mem1"))
	r.Register(NewMockMemoryProvider("mem2"))

	names = r.Names(ProviderTypeMemory)
	if len(names) != 2 {
		t.Errorf("Names() = %d, want 2", len(names))
	}
}
