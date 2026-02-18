// Package squads provides squad management for agent team coordination.
package squads

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

// Squad represents a running squad sandbox with its configuration.
type Squad struct {
	// Name is the squad identifier.
	Name string

	// Config is the squad configuration.
	Config config.SquadConfig

	// Sandbox is the sandbox info if running.
	Sandbox *providers.Sandbox

	// Status is the current squad status.
	Status SquadStatus

	// Schemas contains input/output JSON schemas for validation.
	// Nil if no schemas are defined (free-form mode).
	Schemas *SquadSchemas
}

// SquadStatus represents the status of a squad.
type SquadStatus string

const (
	SquadStatusUnknown  SquadStatus = ""
	SquadStatusStopped  SquadStatus = "stopped"
	SquadStatusRunning  SquadStatus = "running"
	SquadStatusCreating SquadStatus = "creating"
	SquadStatusFailed   SquadStatus = "failed"
)

// Service manages squad sandboxes.
type Service struct {
	provider *sandbox.AppleProvider
	mu       sync.RWMutex
	squads   map[string]*Squad
}

// NewService creates a new squad service.
func NewService(provider *sandbox.AppleProvider) *Service {
	return &Service{
		provider: provider,
		squads:   make(map[string]*Squad),
	}
}

// Create creates a new squad with the given configuration.
// If the squad already exists, returns the existing squad.
func (s *Service) Create(ctx context.Context, cfg config.SquadConfig) (*Squad, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("squad name is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already exists
	if existing, ok := s.squads[cfg.Name]; ok {
		return existing, nil
	}

	debug.Log("creating squad", "name", cfg.Name)

	// Save config
	if err := config.SaveSquadConfig(cfg); err != nil {
		return nil, fmt.Errorf("save squad config: %w", err)
	}

	// Create sandbox (this also creates squad directories)
	sb, err := sandbox.EnsureSquadSandbox(ctx, s.provider, cfg.Name)
	if err != nil {
		return nil, fmt.Errorf("create squad sandbox: %w", err)
	}

	// Create default SQUAD.md constitution (must be after sandbox/dirs are created)
	if constitution, _ := LoadConstitution(cfg.Name); constitution == nil {
		// File doesn't exist, create default
		if err := CreateDefaultConstitution(cfg.Name, cfg.Agents); err != nil {
			debug.Log("failed to create default constitution", "squad", cfg.Name, "error", err)
		}
	}

	// Load schemas (optional)
	schemas, err := LoadSquadSchemas(cfg.Name)
	if err != nil {
		return nil, fmt.Errorf("load squad schemas: %w", err)
	}

	squad := &Squad{
		Name:    cfg.Name,
		Config:  cfg,
		Sandbox: &sb,
		Status:  SquadStatusRunning,
		Schemas: schemas,
	}

	s.squads[cfg.Name] = squad

	debug.Log("squad created", "name", cfg.Name, "sandbox", sb.ID)
	return squad, nil
}

// Get returns a squad by name.
func (s *Service) Get(ctx context.Context, name string) (*Squad, error) {
	s.mu.RLock()
	if squad, ok := s.squads[name]; ok {
		s.mu.RUnlock()
		return squad, nil
	}
	s.mu.RUnlock()

	// Try to load from config and sandbox
	cfg, err := config.LoadSquadConfig(name)
	if err != nil {
		return nil, fmt.Errorf("load squad config: %w", err)
	}

	// Load schemas (optional)
	schemas, err := LoadSquadSchemas(name)
	if err != nil {
		return nil, fmt.Errorf("load squad schemas: %w", err)
	}

	// Check if sandbox exists
	sb, sandboxErr := sandbox.GetSquadSandbox(ctx, s.provider, name)

	squad := &Squad{
		Name:    name,
		Config:  cfg,
		Status:  SquadStatusStopped,
		Schemas: schemas,
	}

	if sandboxErr == nil {
		squad.Sandbox = &sb
		if sb.Status == providers.SandboxStatusRunning {
			squad.Status = SquadStatusRunning
		}
	}

	s.mu.Lock()
	s.squads[name] = squad
	s.mu.Unlock()

	return squad, nil
}

// List returns all configured squads.
func (s *Service) List(ctx context.Context) ([]*Squad, error) {
	// Get squads from config
	names, err := config.ListSquadConfigs()
	if err != nil {
		return nil, fmt.Errorf("list squad configs: %w", err)
	}

	var squads []*Squad
	for _, name := range names {
		squad, err := s.Get(ctx, name)
		if err != nil {
			debug.Log("error loading squad", "name", name, "error", err)
			continue
		}
		squads = append(squads, squad)
	}

	return squads, nil
}

// Start starts a squad sandbox.
func (s *Service) Start(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	squad, ok := s.squads[name]
	if !ok {
		// Load it first
		s.mu.Unlock()
		var err error
		squad, err = s.Get(ctx, name)
		s.mu.Lock()
		if err != nil {
			return fmt.Errorf("get squad: %w", err)
		}
	}

	sb, err := sandbox.EnsureSquadSandbox(ctx, s.provider, name)
	if err != nil {
		return fmt.Errorf("start squad sandbox: %w", err)
	}

	squad.Sandbox = &sb
	squad.Status = SquadStatusRunning

	debug.Log("squad started", "name", name)
	return nil
}

// Stop stops a squad sandbox.
func (s *Service) Stop(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := sandbox.StopSquadSandbox(ctx, s.provider, name); err != nil {
		return fmt.Errorf("stop squad sandbox: %w", err)
	}

	if squad, ok := s.squads[name]; ok {
		squad.Status = SquadStatusStopped
	}

	debug.Log("squad stopped", "name", name)
	return nil
}

// Destroy destroys a squad sandbox and optionally deletes its data.
func (s *Service) Destroy(ctx context.Context, name string, deleteData bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete sandbox
	if err := sandbox.DeleteSquadSandbox(ctx, s.provider, name, deleteData); err != nil {
		debug.Log("error deleting squad sandbox", "name", name, "error", err)
	}

	// Delete config if deleting data
	if deleteData {
		if err := config.DeleteSquadConfig(name); err != nil {
			debug.Log("error deleting squad config", "name", name, "error", err)
		}
	}

	delete(s.squads, name)

	debug.Log("squad destroyed", "name", name, "deleteData", deleteData)
	return nil
}

// EnsureAgentUser ensures an agent user exists in a squad sandbox.
func (s *Service) EnsureAgentUser(ctx context.Context, squadName, agentHandle string) error {
	return sandbox.EnsureSquadAgentUser(ctx, s.provider, squadName, agentHandle)
}

// GetTicketsDir returns the tickets directory for a squad.
func (s *Service) GetTicketsDir(name string) string {
	return paths.SquadTicketsDir(name)
}

// GetContextDir returns the context directory for a squad.
func (s *Service) GetContextDir(name string) string {
	return paths.SquadContextDir(name)
}

// GetWorkspaceDir returns the workspace directory for a squad.
func (s *Service) GetWorkspaceDir(name string) string {
	return paths.SquadWorkspaceDir(name)
}
