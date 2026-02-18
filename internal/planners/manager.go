package planners

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/alexcabrera/ayo/internal/config"
)

// SandboxPlanners holds the instantiated planners for a sandbox.
type SandboxPlanners struct {
	// NearTerm is the near-term planner (e.g., todos).
	NearTerm PlannerPlugin

	// LongTerm is the long-term planner (e.g., tickets).
	LongTerm PlannerPlugin
}

// Close releases resources for both planners.
func (sp *SandboxPlanners) Close() error {
	var errs []error
	if sp.NearTerm != nil {
		if err := sp.NearTerm.Close(); err != nil {
			errs = append(errs, fmt.Errorf("near-term planner: %w", err))
		}
	}
	if sp.LongTerm != nil {
		if err := sp.LongTerm.Close(); err != nil {
			errs = append(errs, fmt.Errorf("long-term planner: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("closing planners: %v", errs)
	}
	return nil
}

// SandboxPlannerManager manages planner instances for each sandbox.
// It handles instantiation, caching, and cleanup of planners.
type SandboxPlannerManager struct {
	registry  *Registry
	config    config.Config
	instances map[string]*SandboxPlanners
	mu        sync.RWMutex
}

// NewSandboxPlannerManager creates a new manager.
func NewSandboxPlannerManager(registry *Registry, cfg config.Config) *SandboxPlannerManager {
	if registry == nil {
		registry = DefaultRegistry
	}
	return &SandboxPlannerManager{
		registry:  registry,
		config:    cfg,
		instances: make(map[string]*SandboxPlanners),
	}
}

// GetPlanners returns the planners for a sandbox, creating them if needed.
// If override is provided, it takes precedence over global config.
// The sandboxDir should be the root of the sandbox filesystem.
func (m *SandboxPlannerManager) GetPlanners(sandboxName, sandboxDir string, override *config.PlannersConfig) (*SandboxPlanners, error) {
	// Check cache first
	m.mu.RLock()
	if planners, ok := m.instances[sandboxName]; ok {
		m.mu.RUnlock()
		return planners, nil
	}
	m.mu.RUnlock()

	// Create new planners
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if planners, ok := m.instances[sandboxName]; ok {
		return planners, nil
	}

	planners, err := m.createPlanners(sandboxName, sandboxDir, override)
	if err != nil {
		return nil, err
	}

	m.instances[sandboxName] = planners
	return planners, nil
}

// ClosePlanners closes and removes the planners for a sandbox.
func (m *SandboxPlannerManager) ClosePlanners(sandboxName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	planners, ok := m.instances[sandboxName]
	if !ok {
		return nil // Nothing to close
	}

	delete(m.instances, sandboxName)
	return planners.Close()
}

// CloseAll closes all cached planners.
func (m *SandboxPlannerManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, planners := range m.instances {
		if err := planners.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}
	m.instances = make(map[string]*SandboxPlanners)

	if len(errs) > 0 {
		return fmt.Errorf("closing all planners: %v", errs)
	}
	return nil
}

// HasPlanners returns true if planners are cached for the given sandbox.
func (m *SandboxPlannerManager) HasPlanners(sandboxName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.instances[sandboxName]
	return ok
}

// createPlanners instantiates planners for a sandbox.
func (m *SandboxPlannerManager) createPlanners(sandboxName, sandboxDir string, override *config.PlannersConfig) (*SandboxPlanners, error) {
	// Resolve which planners to use
	cfg := m.resolvePlannersConfig(override)

	// Create state directories
	nearTermStateDir := filepath.Join(sandboxDir, ".planner.near")
	longTermStateDir := filepath.Join(sandboxDir, ".planner.long")

	if err := os.MkdirAll(nearTermStateDir, 0755); err != nil {
		return nil, fmt.Errorf("create near-term state dir: %w", err)
	}
	if err := os.MkdirAll(longTermStateDir, 0755); err != nil {
		return nil, fmt.Errorf("create long-term state dir: %w", err)
	}

	// Instantiate near-term planner
	nearTermCtx := PlannerContext{
		SandboxName: sandboxName,
		SandboxDir:  sandboxDir,
		StateDir:    nearTermStateDir,
		Config:      nil, // TODO: planner-specific config
	}
	nearTerm, err := m.registry.Instantiate(cfg.NearTerm, nearTermCtx)
	if err != nil {
		return nil, fmt.Errorf("instantiate near-term planner %q: %w", cfg.NearTerm, err)
	}

	// Initialize near-term planner
	if err := nearTerm.Init(context.Background()); err != nil {
		return nil, fmt.Errorf("init near-term planner %q: %w", cfg.NearTerm, err)
	}

	// Instantiate long-term planner
	longTermCtx := PlannerContext{
		SandboxName: sandboxName,
		SandboxDir:  sandboxDir,
		StateDir:    longTermStateDir,
		Config:      nil, // TODO: planner-specific config
	}
	longTerm, err := m.registry.Instantiate(cfg.LongTerm, longTermCtx)
	if err != nil {
		// Clean up near-term planner on failure
		_ = nearTerm.Close()
		return nil, fmt.Errorf("instantiate long-term planner %q: %w", cfg.LongTerm, err)
	}

	// Initialize long-term planner
	if err := longTerm.Init(context.Background()); err != nil {
		// Clean up both planners on failure
		_ = nearTerm.Close()
		_ = longTerm.Close()
		return nil, fmt.Errorf("init long-term planner %q: %w", cfg.LongTerm, err)
	}

	return &SandboxPlanners{
		NearTerm: nearTerm,
		LongTerm: longTerm,
	}, nil
}

// resolvePlannersConfig merges override with global config.
func (m *SandboxPlannerManager) resolvePlannersConfig(override *config.PlannersConfig) config.PlannersConfig {
	// Start with global config defaults
	result := m.config.Planners.WithDefaults()

	// Apply override if provided
	if override != nil {
		if override.NearTerm != "" {
			result.NearTerm = override.NearTerm
		}
		if override.LongTerm != "" {
			result.LongTerm = override.LongTerm
		}
	}

	return result
}

// DefaultManager is a global manager using the DefaultRegistry.
// Initialize this in main() or via NewSandboxPlannerManager.
var DefaultManager *SandboxPlannerManager
