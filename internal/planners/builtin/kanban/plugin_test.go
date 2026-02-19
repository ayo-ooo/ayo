package kanban

import (
	"context"
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

	ctx := planners.PlannerContext{
		SandboxName: "test-sandbox",
		SandboxDir:  "/sandbox",
		StateDir:    "/sandbox/.planner.long",
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

func TestRegistration(t *testing.T) {
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
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Errorf("Init() failed: %v", err)
	}
	defer p.Close()

	// Database should be open
	if p.DB() == nil {
		t.Error("expected database to be open after Init")
	}
}

func TestPlugin_Close(t *testing.T) {
	p := &Plugin{}
	if err := p.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestPlugin_Tools(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	tools := p.Tools()
	if len(tools) != 1 {
		t.Errorf("Tools() returned %d tools, want 1", len(tools))
	}
	if len(tools) > 0 && tools[0].Info().Name != ToolName {
		t.Errorf("Tools()[0].Info().Name = %q, want %q", tools[0].Info().Name, ToolName)
	}
}

func TestPlugin_Instructions(t *testing.T) {
	p := &Plugin{}
	instructions := p.Instructions()
	if instructions == "" {
		t.Error("Instructions() should not be empty")
	}
	if instructions != KanbanInstructions {
		t.Errorf("Instructions() = %q, want KanbanInstructions", instructions)
	}
}

func TestPlugin_InitCreatesDefaultColumns(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	// Check that default columns were created
	rows, err := p.db.QueryContext(ctx, "SELECT name FROM columns ORDER BY position")
	if err != nil {
		t.Fatalf("query columns: %v", err)
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		cols = append(cols, name)
	}

	if len(cols) != len(DefaultColumns) {
		t.Errorf("got %d columns, want %d", len(cols), len(DefaultColumns))
	}
	for i, col := range DefaultColumns {
		if cols[i] != col {
			t.Errorf("cols[%d] = %q, want %q", i, cols[i], col)
		}
	}
}
