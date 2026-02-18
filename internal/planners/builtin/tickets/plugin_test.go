package tickets

import (
	"context"
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/planners"
)

func TestPlugin_Name(t *testing.T) {
	p := &Plugin{}
	if got := p.Name(); got != PluginName {
		t.Errorf("Name() = %q, want %q", got, PluginName)
	}
}

func TestPlugin_Type(t *testing.T) {
	p := &Plugin{}
	if got := p.Type(); got != planners.LongTerm {
		t.Errorf("Type() = %q, want %q", got, planners.LongTerm)
	}
}

func TestPlugin_StateDir(t *testing.T) {
	p := &Plugin{stateDir: "/test/state/dir"}
	if got := p.StateDir(); got != "/test/state/dir" {
		t.Errorf("StateDir() = %q, want %q", got, "/test/state/dir")
	}
}

func TestNew(t *testing.T) {
	factory := New()
	if factory == nil {
		t.Fatal("New() returned nil")
	}

	// Use temp dir to allow directory creation
	tempDir := t.TempDir()
	ctx := planners.PlannerContext{
		SandboxName: "test-sandbox",
		SandboxDir:  "/sandbox",
		StateDir:    tempDir,
	}

	plugin, err := factory(ctx)
	if err != nil {
		t.Fatalf("factory() failed: %v", err)
	}
	if plugin == nil {
		t.Fatal("factory() returned nil plugin")
	}

	if plugin.Name() != PluginName {
		t.Errorf("plugin.Name() = %q, want %q", plugin.Name(), PluginName)
	}
	if plugin.StateDir() != ctx.StateDir {
		t.Errorf("plugin.StateDir() = %q, want %q", plugin.StateDir(), ctx.StateDir)
	}
}

func TestNew_CreatesService(t *testing.T) {
	factory := New()
	tempDir := t.TempDir()
	ctx := planners.PlannerContext{
		SandboxName: "test-sandbox",
		SandboxDir:  "/sandbox",
		StateDir:    tempDir,
	}

	plugin, err := factory(ctx)
	if err != nil {
		t.Fatalf("factory() failed: %v", err)
	}

	// Cast to Plugin to access Service()
	ticketsPlugin, ok := plugin.(*Plugin)
	if !ok {
		t.Fatal("plugin is not *Plugin")
	}

	svc := ticketsPlugin.Service()
	if svc == nil {
		t.Error("Service() returned nil, want non-nil tickets.Service")
	}
}

func TestRegistration(t *testing.T) {
	// The plugin should be registered via init()
	if !planners.DefaultRegistry.Has(PluginName) {
		t.Errorf("plugin %q not found in DefaultRegistry", PluginName)
	}

	factory, ok := planners.DefaultRegistry.Get(PluginName)
	if !ok {
		t.Fatalf("Get(%q) returned false", PluginName)
	}
	if factory == nil {
		t.Errorf("Get(%q) returned nil factory", PluginName)
	}
}

func TestPlugin_Init(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	if err := p.Init(context.Background()); err != nil {
		t.Errorf("Init() failed: %v", err)
	}
}

func TestPlugin_Close(t *testing.T) {
	p := &Plugin{}
	if err := p.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestPlugin_Tools(t *testing.T) {
	p := &Plugin{}
	tools := p.Tools()
	// Should return 6 tools: create, list, start, close, block, note
	if len(tools) != 6 {
		t.Errorf("Tools() returned %d tools, want 6", len(tools))
	}

	// Verify tool names
	expectedNames := []string{
		ToolTicketCreate,
		ToolTicketList,
		ToolTicketStart,
		ToolTicketClose,
		ToolTicketBlock,
		ToolTicketNote,
	}
	for i, tool := range tools {
		if tool.Info().Name != expectedNames[i] {
			t.Errorf("Tools()[%d].Info().Name = %q, want %q", i, tool.Info().Name, expectedNames[i])
		}
	}
}

func TestPlugin_Instructions(t *testing.T) {
	p := &Plugin{}
	instructions := p.Instructions()
	// Instructions should return TicketsInstructions constant
	if instructions == "" {
		t.Error("Instructions() should not be empty")
	}
	if instructions != TicketsInstructions {
		t.Errorf("Instructions() should return TicketsInstructions constant")
	}
	// Verify key content is present
	if !strings.Contains(instructions, "Long-Term Work Planning") {
		t.Error("Instructions should contain 'Long-Term Work Planning'")
	}
	if !strings.Contains(instructions, "ticket_create") {
		t.Error("Instructions should document ticket_create tool")
	}
	if !strings.Contains(instructions, "ticket_close") {
		t.Error("Instructions should document ticket_close tool")
	}
}

func TestPlugin_Service(t *testing.T) {
	// Service returns nil when not initialized via factory
	p := &Plugin{}
	if svc := p.Service(); svc != nil {
		t.Error("Service() on uninitialized plugin should return nil")
	}
}
