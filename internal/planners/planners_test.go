package planners

import (
	"context"
	"testing"

	"charm.land/fantasy"
)

func TestPlannerType_String(t *testing.T) {
	tests := []struct {
		pt   PlannerType
		want string
	}{
		{NearTerm, "near"},
		{LongTerm, "long"},
		{PlannerType("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.pt.String(); got != tt.want {
				t.Errorf("PlannerType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlannerType_IsValid(t *testing.T) {
	tests := []struct {
		pt   PlannerType
		want bool
	}{
		{NearTerm, true},
		{LongTerm, true},
		{PlannerType(""), false},
		{PlannerType("invalid"), false},
		{PlannerType("near-term"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.pt), func(t *testing.T) {
			if got := tt.pt.IsValid(); got != tt.want {
				t.Errorf("PlannerType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlannersConfig_WithDefaults(t *testing.T) {
	tests := []struct {
		name string
		cfg  PlannersConfig
		want PlannersConfig
	}{
		{
			name: "empty config gets defaults",
			cfg:  PlannersConfig{},
			want: PlannersConfig{NearTerm: "ayo-todos", LongTerm: "ayo-tickets"},
		},
		{
			name: "partial config gets partial defaults",
			cfg:  PlannersConfig{NearTerm: "custom-todos"},
			want: PlannersConfig{NearTerm: "custom-todos", LongTerm: "ayo-tickets"},
		},
		{
			name: "full config unchanged",
			cfg:  PlannersConfig{NearTerm: "custom-todos", LongTerm: "custom-tickets"},
			want: PlannersConfig{NearTerm: "custom-todos", LongTerm: "custom-tickets"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.WithDefaults()
			if got != tt.want {
				t.Errorf("PlannersConfig.WithDefaults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlannersConfig_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		cfg  PlannersConfig
		want bool
	}{
		{
			name: "empty config",
			cfg:  PlannersConfig{},
			want: true,
		},
		{
			name: "near term only",
			cfg:  PlannersConfig{NearTerm: "ayo-todos"},
			want: false,
		},
		{
			name: "long term only",
			cfg:  PlannersConfig{LongTerm: "ayo-tickets"},
			want: false,
		},
		{
			name: "both set",
			cfg:  PlannersConfig{NearTerm: "a", LongTerm: "b"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsEmpty(); got != tt.want {
				t.Errorf("PlannersConfig.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlannerContext_Fields(t *testing.T) {
	ctx := PlannerContext{
		SandboxName: "test-squad",
		SandboxDir:  "/path/to/sandbox",
		StateDir:    "/path/to/sandbox/.planner.near",
		Config:      map[string]any{"key": "value"},
	}

	if ctx.SandboxName != "test-squad" {
		t.Errorf("SandboxName = %v, want test-squad", ctx.SandboxName)
	}
	if ctx.SandboxDir != "/path/to/sandbox" {
		t.Errorf("SandboxDir = %v, want /path/to/sandbox", ctx.SandboxDir)
	}
	if ctx.StateDir != "/path/to/sandbox/.planner.near" {
		t.Errorf("StateDir = %v, want /path/to/sandbox/.planner.near", ctx.StateDir)
	}
	if ctx.Config["key"] != "value" {
		t.Errorf("Config[key] = %v, want value", ctx.Config["key"])
	}
}

// mockPlanner is a test implementation of PlannerPlugin.
type mockPlanner struct {
	name      string
	planType  PlannerType
	stateDir  string
	initErr   error
	closeErr  error
	tools     []fantasy.AgentTool
	instrText string
}

func (m *mockPlanner) Name() string                   { return m.name }
func (m *mockPlanner) Type() PlannerType              { return m.planType }
func (m *mockPlanner) Init(ctx context.Context) error { return m.initErr }
func (m *mockPlanner) Close() error                   { return m.closeErr }
func (m *mockPlanner) Tools() []fantasy.AgentTool     { return m.tools }
func (m *mockPlanner) Instructions() string           { return m.instrText }
func (m *mockPlanner) StateDir() string               { return m.stateDir }

func TestPlannerPlugin_Interface(t *testing.T) {
	// Verify that mockPlanner implements PlannerPlugin
	var _ PlannerPlugin = (*mockPlanner)(nil)

	p := &mockPlanner{
		name:      "test-planner",
		planType:  NearTerm,
		stateDir:  "/tmp/state",
		instrText: "Use this planner wisely.",
	}

	if p.Name() != "test-planner" {
		t.Errorf("Name() = %v, want test-planner", p.Name())
	}
	if p.Type() != NearTerm {
		t.Errorf("Type() = %v, want NearTerm", p.Type())
	}
	if p.StateDir() != "/tmp/state" {
		t.Errorf("StateDir() = %v, want /tmp/state", p.StateDir())
	}
	if p.Instructions() != "Use this planner wisely." {
		t.Errorf("Instructions() = %v, want 'Use this planner wisely.'", p.Instructions())
	}
	if err := p.Init(context.Background()); err != nil {
		t.Errorf("Init() error = %v, want nil", err)
	}
	if err := p.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestPlannerFactory(t *testing.T) {
	factory := func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{
			name:     "factory-planner",
			planType: LongTerm,
			stateDir: ctx.StateDir,
		}, nil
	}

	ctx := PlannerContext{
		SandboxName: "test",
		StateDir:    "/tmp/test-state",
	}

	planner, err := factory(ctx)
	if err != nil {
		t.Fatalf("factory() error = %v", err)
	}

	if planner.Name() != "factory-planner" {
		t.Errorf("Name() = %v, want factory-planner", planner.Name())
	}
	if planner.StateDir() != "/tmp/test-state" {
		t.Errorf("StateDir() = %v, want /tmp/test-state", planner.StateDir())
	}
}
