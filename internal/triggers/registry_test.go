package triggers

import (
	"context"
	"testing"
)

func TestTriggerRegistry(t *testing.T) {
	reg := NewRegistry()

	// Create a mock trigger factory
	mockFactory := func() TriggerPlugin {
		return &mockTriggerPlugin{
			name:        "mock",
			category:    TriggerCategoryPoll,
			description: "A mock trigger",
		}
	}

	// Test Register
	err := reg.Register("mock", mockFactory, "")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Test Has
	if !reg.Has("mock") {
		t.Error("Has() returned false for registered trigger")
	}
	if reg.Has("nonexistent") {
		t.Error("Has() returned true for nonexistent trigger")
	}

	// Test Create
	plugin, err := reg.Create("mock")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if plugin.Name() != "mock" {
		t.Errorf("Create() name = %v, want mock", plugin.Name())
	}

	// Test duplicate registration
	err = reg.Register("mock", mockFactory, "")
	if err == nil {
		t.Error("expected error for duplicate registration")
	}

	// Test List
	infos := reg.List()
	if len(infos) != 1 {
		t.Errorf("List() len = %v, want 1", len(infos))
	}
	if infos[0].Name != "mock" {
		t.Errorf("List()[0].Name = %v, want mock", infos[0].Name)
	}

	// Test ListNames
	names := reg.ListNames()
	if len(names) != 1 {
		t.Errorf("ListNames() len = %v, want 1", len(names))
	}
	if names[0] != "mock" {
		t.Errorf("ListNames()[0] = %v, want mock", names[0])
	}

	// Test IsBuiltin
	if !reg.IsBuiltin("mock") {
		t.Error("IsBuiltin() returned false for builtin trigger")
	}

	// Test GetPluginName
	if reg.GetPluginName("mock") != "" {
		t.Error("GetPluginName() should return empty for builtin trigger")
	}

	// Test Unregister
	err = reg.Unregister("mock")
	if err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}
	if reg.Has("mock") {
		t.Error("Has() returned true after Unregister()")
	}

	// Test Clear
	err = reg.Register("test1", mockFactory, "plugin1")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	err = reg.Register("test2", mockFactory, "plugin2")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	reg.Clear()
	if len(reg.ListNames()) != 0 {
		t.Error("Clear() didn't remove all triggers")
	}
}

func TestTriggerRegistryPluginTriggers(t *testing.T) {
	reg := NewRegistry()

	mockFactory := func() TriggerPlugin {
		return &mockTriggerPlugin{
			name:        "imap",
			category:    TriggerCategoryPush,
			description: "IMAP email trigger",
		}
	}

	// Register as plugin trigger
	err := reg.Register("imap", mockFactory, "email-plugin")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Should not be builtin
	if reg.IsBuiltin("imap") {
		t.Error("IsBuiltin() returned true for plugin trigger")
	}

	// GetPluginName should return plugin name
	if reg.GetPluginName("imap") != "email-plugin" {
		t.Errorf("GetPluginName() = %v, want email-plugin", reg.GetPluginName("imap"))
	}

	// List should include plugin name
	infos := reg.List()
	if len(infos) != 1 {
		t.Fatalf("List() len = %v, want 1", len(infos))
	}
	if infos[0].PluginName != "email-plugin" {
		t.Errorf("List()[0].PluginName = %v, want email-plugin", infos[0].PluginName)
	}
}

// mockTriggerPlugin implements TriggerPlugin for testing
type mockTriggerPlugin struct {
	name        string
	category    TriggerCategory
	description string
}

func (m *mockTriggerPlugin) Name() string                        { return m.name }
func (m *mockTriggerPlugin) Category() TriggerCategory           { return m.category }
func (m *mockTriggerPlugin) Description() string                 { return m.description }
func (m *mockTriggerPlugin) ConfigSchema() map[string]any        { return nil }
func (m *mockTriggerPlugin) Init(_ context.Context, _ map[string]any) error { return nil }
func (m *mockTriggerPlugin) Start(_ context.Context, _ EventCallback) error { return nil }
func (m *mockTriggerPlugin) Stop() error                         { return nil }
func (m *mockTriggerPlugin) Status() TriggerStatus               { return TriggerStatus{Running: true} }
