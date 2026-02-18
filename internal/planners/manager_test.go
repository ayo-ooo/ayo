package planners

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"charm.land/fantasy"
	"github.com/alexcabrera/ayo/internal/config"
)

// managerMockPlanner implements PlannerPlugin for testing the manager.
type managerMockPlanner struct {
	name        string
	plannerType PlannerType
	stateDir    string
	initCalled  bool
	closeCalled bool
	initErr     error
	closeErr    error
}

func (m *managerMockPlanner) Name() string                   { return m.name }
func (m *managerMockPlanner) Type() PlannerType              { return m.plannerType }
func (m *managerMockPlanner) StateDir() string               { return m.stateDir }
func (m *managerMockPlanner) Tools() []fantasy.AgentTool     { return nil }
func (m *managerMockPlanner) Instructions() string           { return "mock instructions" }
func (m *managerMockPlanner) Init(ctx context.Context) error { m.initCalled = true; return m.initErr }
func (m *managerMockPlanner) Close() error                   { m.closeCalled = true; return m.closeErr }

// managerMockPlannerFactory creates a factory that returns the given planner.
func managerMockPlannerFactory(p *managerMockPlanner) PlannerFactory {
	return func(ctx PlannerContext) (PlannerPlugin, error) {
		p.stateDir = ctx.StateDir
		return p, nil
	}
}

// failingFactory creates a factory that always returns an error.
func failingFactory(errMsg string) PlannerFactory {
	return func(ctx PlannerContext) (PlannerPlugin, error) {
		return nil, errors.New(errMsg)
	}
}

func TestNewSandboxPlannerManager(t *testing.T) {
	// With explicit registry
	registry := NewRegistry()
	cfg := config.Config{}
	manager := NewSandboxPlannerManager(registry, cfg)

	if manager.registry != registry {
		t.Error("expected registry to be set")
	}
	if manager.instances == nil {
		t.Error("expected instances map to be initialized")
	}

	// With nil registry (should use default)
	manager2 := NewSandboxPlannerManager(nil, cfg)
	if manager2.registry != DefaultRegistry {
		t.Error("expected DefaultRegistry to be used when nil")
	}
}

func TestSandboxPlannerManager_GetPlanners(t *testing.T) {
	tmpDir := t.TempDir()
	sandboxDir := filepath.Join(tmpDir, "sandbox")
	if err := os.MkdirAll(sandboxDir, 0755); err != nil {
		t.Fatalf("create sandbox dir: %v", err)
	}

	nearTermPlanner := &managerMockPlanner{name: "test-todos", plannerType: NearTerm}
	longTermPlanner := &managerMockPlanner{name: "test-tickets", plannerType: LongTerm}

	registry := NewRegistry()
	registry.Register("test-todos", managerMockPlannerFactory(nearTermPlanner))
	registry.Register("test-tickets", managerMockPlannerFactory(longTermPlanner))

	cfg := config.Config{
		Planners: config.PlannersConfig{
			NearTerm: "test-todos",
			LongTerm: "test-tickets",
		},
	}

	manager := NewSandboxPlannerManager(registry, cfg)

	// Get planners
	planners, err := manager.GetPlanners("test-sandbox", sandboxDir, nil)
	if err != nil {
		t.Fatalf("GetPlanners() error = %v", err)
	}

	if planners.NearTerm == nil {
		t.Error("expected NearTerm planner")
	}
	if planners.LongTerm == nil {
		t.Error("expected LongTerm planner")
	}
	if !nearTermPlanner.initCalled {
		t.Error("expected NearTerm Init() to be called")
	}
	if !longTermPlanner.initCalled {
		t.Error("expected LongTerm Init() to be called")
	}

	// Verify state directories created
	nearStateDir := filepath.Join(sandboxDir, ".planner.near")
	longStateDir := filepath.Join(sandboxDir, ".planner.long")

	if _, err := os.Stat(nearStateDir); os.IsNotExist(err) {
		t.Error("near-term state dir should exist")
	}
	if _, err := os.Stat(longStateDir); os.IsNotExist(err) {
		t.Error("long-term state dir should exist")
	}

	// Verify caching - second call should return same instance
	planners2, err := manager.GetPlanners("test-sandbox", sandboxDir, nil)
	if err != nil {
		t.Fatalf("GetPlanners() second call error = %v", err)
	}
	if planners2 != planners {
		t.Error("expected cached planners to be returned")
	}
}

func TestSandboxPlannerManager_GetPlanners_WithOverride(t *testing.T) {
	tmpDir := t.TempDir()
	sandboxDir := filepath.Join(tmpDir, "sandbox")
	if err := os.MkdirAll(sandboxDir, 0755); err != nil {
		t.Fatalf("create sandbox dir: %v", err)
	}

	defaultNear := &managerMockPlanner{name: "default-todos", plannerType: NearTerm}
	overrideNear := &managerMockPlanner{name: "custom-todos", plannerType: NearTerm}
	defaultLong := &managerMockPlanner{name: "default-tickets", plannerType: LongTerm}

	registry := NewRegistry()
	registry.Register("default-todos", managerMockPlannerFactory(defaultNear))
	registry.Register("custom-todos", managerMockPlannerFactory(overrideNear))
	registry.Register("default-tickets", managerMockPlannerFactory(defaultLong))

	cfg := config.Config{
		Planners: config.PlannersConfig{
			NearTerm: "default-todos",
			LongTerm: "default-tickets",
		},
	}

	manager := NewSandboxPlannerManager(registry, cfg)

	// Override near-term only
	override := &config.PlannersConfig{
		NearTerm: "custom-todos",
	}

	planners, err := manager.GetPlanners("test-sandbox", sandboxDir, override)
	if err != nil {
		t.Fatalf("GetPlanners() error = %v", err)
	}

	// Should use override for near-term
	if planners.NearTerm.Name() != "custom-todos" {
		t.Errorf("expected custom-todos, got %s", planners.NearTerm.Name())
	}
	// Should use default for long-term
	if planners.LongTerm.Name() != "default-tickets" {
		t.Errorf("expected default-tickets, got %s", planners.LongTerm.Name())
	}
}

func TestSandboxPlannerManager_ClosePlanners(t *testing.T) {
	tmpDir := t.TempDir()
	sandboxDir := filepath.Join(tmpDir, "sandbox")
	if err := os.MkdirAll(sandboxDir, 0755); err != nil {
		t.Fatalf("create sandbox dir: %v", err)
	}

	nearTermPlanner := &managerMockPlanner{name: "test-todos", plannerType: NearTerm}
	longTermPlanner := &managerMockPlanner{name: "test-tickets", plannerType: LongTerm}

	registry := NewRegistry()
	registry.Register("test-todos", managerMockPlannerFactory(nearTermPlanner))
	registry.Register("test-tickets", managerMockPlannerFactory(longTermPlanner))

	cfg := config.Config{
		Planners: config.PlannersConfig{
			NearTerm: "test-todos",
			LongTerm: "test-tickets",
		},
	}

	manager := NewSandboxPlannerManager(registry, cfg)

	// Get planners first
	_, err := manager.GetPlanners("test-sandbox", sandboxDir, nil)
	if err != nil {
		t.Fatalf("GetPlanners() error = %v", err)
	}

	if !manager.HasPlanners("test-sandbox") {
		t.Error("expected HasPlanners to return true")
	}

	// Close planners
	if err := manager.ClosePlanners("test-sandbox"); err != nil {
		t.Fatalf("ClosePlanners() error = %v", err)
	}

	if !nearTermPlanner.closeCalled {
		t.Error("expected NearTerm Close() to be called")
	}
	if !longTermPlanner.closeCalled {
		t.Error("expected LongTerm Close() to be called")
	}
	if manager.HasPlanners("test-sandbox") {
		t.Error("expected HasPlanners to return false after close")
	}

	// Closing non-existent sandbox should be no-op
	if err := manager.ClosePlanners("nonexistent"); err != nil {
		t.Errorf("ClosePlanners(nonexistent) should not error: %v", err)
	}
}

func TestSandboxPlannerManager_CloseAll(t *testing.T) {
	tmpDir := t.TempDir()
	sandbox1Dir := filepath.Join(tmpDir, "sandbox1")
	sandbox2Dir := filepath.Join(tmpDir, "sandbox2")
	os.MkdirAll(sandbox1Dir, 0755)
	os.MkdirAll(sandbox2Dir, 0755)

	near1 := &managerMockPlanner{name: "test-todos", plannerType: NearTerm}
	long1 := &managerMockPlanner{name: "test-tickets", plannerType: LongTerm}
	near2 := &managerMockPlanner{name: "test-todos", plannerType: NearTerm}
	long2 := &managerMockPlanner{name: "test-tickets", plannerType: LongTerm}

	// Track which mock to return
	nearCount := 0
	longCount := 0

	registry := NewRegistry()
	registry.Register("test-todos", func(ctx PlannerContext) (PlannerPlugin, error) {
		nearCount++
		if nearCount == 1 {
			near1.stateDir = ctx.StateDir
			return near1, nil
		}
		near2.stateDir = ctx.StateDir
		return near2, nil
	})
	registry.Register("test-tickets", func(ctx PlannerContext) (PlannerPlugin, error) {
		longCount++
		if longCount == 1 {
			long1.stateDir = ctx.StateDir
			return long1, nil
		}
		long2.stateDir = ctx.StateDir
		return long2, nil
	})

	cfg := config.Config{
		Planners: config.PlannersConfig{
			NearTerm: "test-todos",
			LongTerm: "test-tickets",
		},
	}

	manager := NewSandboxPlannerManager(registry, cfg)

	// Create multiple sandboxes
	manager.GetPlanners("sandbox1", sandbox1Dir, nil)
	manager.GetPlanners("sandbox2", sandbox2Dir, nil)

	// Close all
	if err := manager.CloseAll(); err != nil {
		t.Fatalf("CloseAll() error = %v", err)
	}

	if near1.closeCalled != true || long1.closeCalled != true {
		t.Error("sandbox1 planners should be closed")
	}
	if near2.closeCalled != true || long2.closeCalled != true {
		t.Error("sandbox2 planners should be closed")
	}
	if manager.HasPlanners("sandbox1") || manager.HasPlanners("sandbox2") {
		t.Error("no planners should remain after CloseAll")
	}
}

func TestSandboxPlannerManager_GetPlanners_PlannerNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	sandboxDir := filepath.Join(tmpDir, "sandbox")
	os.MkdirAll(sandboxDir, 0755)

	registry := NewRegistry()
	// Don't register any planners

	cfg := config.Config{
		Planners: config.PlannersConfig{
			NearTerm: "nonexistent-todos",
			LongTerm: "nonexistent-tickets",
		},
	}

	manager := NewSandboxPlannerManager(registry, cfg)

	_, err := manager.GetPlanners("test-sandbox", sandboxDir, nil)
	if err == nil {
		t.Error("expected error for nonexistent planner")
	}
}

func TestSandboxPlannerManager_GetPlanners_InitError(t *testing.T) {
	tmpDir := t.TempDir()
	sandboxDir := filepath.Join(tmpDir, "sandbox")
	os.MkdirAll(sandboxDir, 0755)

	failingPlanner := &managerMockPlanner{
		name:        "failing-todos",
		plannerType: NearTerm,
		initErr:     errors.New("init failed"),
	}

	registry := NewRegistry()
	registry.Register("failing-todos", managerMockPlannerFactory(failingPlanner))
	registry.Register("test-tickets", managerMockPlannerFactory(&managerMockPlanner{name: "test-tickets", plannerType: LongTerm}))

	cfg := config.Config{
		Planners: config.PlannersConfig{
			NearTerm: "failing-todos",
			LongTerm: "test-tickets",
		},
	}

	manager := NewSandboxPlannerManager(registry, cfg)

	_, err := manager.GetPlanners("test-sandbox", sandboxDir, nil)
	if err == nil {
		t.Error("expected error when planner Init() fails")
	}
}

func TestSandboxPlanners_Close(t *testing.T) {
	near := &managerMockPlanner{name: "near", plannerType: NearTerm}
	long := &managerMockPlanner{name: "long", plannerType: LongTerm}

	sp := &SandboxPlanners{
		NearTerm: near,
		LongTerm: long,
	}

	if err := sp.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !near.closeCalled || !long.closeCalled {
		t.Error("both planners should have Close() called")
	}
}

func TestSandboxPlanners_Close_WithErrors(t *testing.T) {
	near := &managerMockPlanner{name: "near", plannerType: NearTerm, closeErr: errors.New("near error")}
	long := &managerMockPlanner{name: "long", plannerType: LongTerm, closeErr: errors.New("long error")}

	sp := &SandboxPlanners{
		NearTerm: near,
		LongTerm: long,
	}

	err := sp.Close()
	if err == nil {
		t.Error("expected error when planners fail to close")
	}
	// Both should still be called even if one fails
	if !near.closeCalled || !long.closeCalled {
		t.Error("both planners should have Close() called even on error")
	}
}

func TestSandboxPlanners_Close_NilPlanners(t *testing.T) {
	sp := &SandboxPlanners{
		NearTerm: nil,
		LongTerm: nil,
	}

	if err := sp.Close(); err != nil {
		t.Errorf("Close() with nil planners should not error: %v", err)
	}
}

func TestResolvePlannersConfig(t *testing.T) {
	cfg := config.Config{
		Planners: config.PlannersConfig{
			NearTerm: "global-todos",
			LongTerm: "global-tickets",
		},
	}

	manager := NewSandboxPlannerManager(NewRegistry(), cfg)

	// No override - use global
	result := manager.resolvePlannersConfig(nil)
	if result.NearTerm != "global-todos" {
		t.Errorf("expected global-todos, got %s", result.NearTerm)
	}
	if result.LongTerm != "global-tickets" {
		t.Errorf("expected global-tickets, got %s", result.LongTerm)
	}

	// Partial override
	override := &config.PlannersConfig{NearTerm: "custom-todos"}
	result = manager.resolvePlannersConfig(override)
	if result.NearTerm != "custom-todos" {
		t.Errorf("expected custom-todos, got %s", result.NearTerm)
	}
	if result.LongTerm != "global-tickets" {
		t.Errorf("expected global-tickets (from global), got %s", result.LongTerm)
	}

	// Full override
	fullOverride := &config.PlannersConfig{NearTerm: "my-todos", LongTerm: "my-tickets"}
	result = manager.resolvePlannersConfig(fullOverride)
	if result.NearTerm != "my-todos" {
		t.Errorf("expected my-todos, got %s", result.NearTerm)
	}
	if result.LongTerm != "my-tickets" {
		t.Errorf("expected my-tickets, got %s", result.LongTerm)
	}
}

func TestResolvePlannersConfig_WithDefaults(t *testing.T) {
	// Empty global config should use defaults
	cfg := config.Config{}

	manager := NewSandboxPlannerManager(NewRegistry(), cfg)

	result := manager.resolvePlannersConfig(nil)
	if result.NearTerm != "ayo-todos" {
		t.Errorf("expected default ayo-todos, got %s", result.NearTerm)
	}
	if result.LongTerm != "ayo-tickets" {
		t.Errorf("expected default ayo-tickets, got %s", result.LongTerm)
	}
}
