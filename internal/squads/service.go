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

	// Constitution is the squad's SQUAD.md constitution.
	// Loaded during squad initialization.
	Constitution *Constitution

	// LeadReady indicates the squad lead is ready to accept input.
	// Set to true when the squad has been fully initialized.
	LeadReady bool

	// Invoker is used to invoke agents within the squad context.
	// If nil, dispatch returns routing info only without actual invocation.
	Invoker AgentInvoker
}

// CanAcceptInput returns true if the squad is ready to accept input.
// The squad must be running and have its lead ready.
func (sq *Squad) CanAcceptInput() bool {
	return sq.Status == SquadStatusRunning && sq.LeadReady
}

// GetAllAgents returns all agents in this squad.
// This includes agents from config, constitution, and the lead.
func (sq *Squad) GetAllAgents() []string {
	agents := make(map[string]bool)

	// Add agents from config
	for _, a := range sq.Config.Agents {
		agent := a
		if len(agent) > 0 && agent[0] != '@' {
			agent = "@" + agent
		}
		agents[agent] = true
	}

	// Add lead
	lead := sq.Config.Lead
	if lead == "" && sq.Constitution != nil {
		lead = sq.Constitution.Frontmatter.Lead
	}
	if lead == "" {
		lead = "@ayo"
	}
	if len(lead) > 0 && lead[0] != '@' {
		lead = "@" + lead
	}
	agents[lead] = true

	// Add agents from constitution
	if sq.Constitution != nil {
		for _, a := range sq.Constitution.GetAgents() {
			agent := a
			if len(agent) > 0 && agent[0] != '@' {
				agent = "@" + agent
			}
			agents[agent] = true
		}
	}

	// Convert to slice
	result := make([]string, 0, len(agents))
	for a := range agents {
		result = append(result, a)
	}
	return result
}

// HasAgent returns true if the given agent is part of this squad.
func (sq *Squad) HasAgent(agentHandle string) bool {
	agent := agentHandle
	if len(agent) > 0 && agent[0] != '@' {
		agent = "@" + agent
	}

	for _, a := range sq.GetAllAgents() {
		if a == agent {
			return true
		}
	}
	return false
}

// IsRunning returns true if the squad is currently running.
func (sq *Squad) IsRunning() bool {
	return sq.Status == SquadStatusRunning
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

	// Load constitution for squad lead (may have been created above or already existed)
	constitution, err := LoadConstitution(cfg.Name)
	if err != nil {
		debug.Log("failed to load constitution", "squad", cfg.Name, "error", err)
		// Continue without constitution - squad can still function
	}

	squad := &Squad{
		Name:         cfg.Name,
		Config:       cfg,
		Sandbox:      &sb,
		Status:       SquadStatusRunning,
		Schemas:      schemas,
		Constitution: constitution,
		LeadReady:    constitution != nil, // Lead is ready if we have a constitution
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

	// Load constitution for squad lead
	constitution, _ := LoadConstitution(name)

	// Check for deprecation warning
	if warning := DeprecationWarning(name); warning != "" {
		debug.Log("squad needs migration", "squad", name, "warning", warning)
	}

	// Check if sandbox exists
	sb, sandboxErr := sandbox.GetSquadSandbox(ctx, s.provider, name)

	squad := &Squad{
		Name:         name,
		Config:       cfg,
		Status:       SquadStatusStopped,
		Schemas:      schemas,
		Constitution: constitution,
		LeadReady:    false, // Not ready until squad is running
	}

	if sandboxErr == nil {
		squad.Sandbox = &sb
		if sb.Status == providers.SandboxStatusRunning {
			squad.Status = SquadStatusRunning
			squad.LeadReady = squad.Constitution != nil
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

	// Load constitution if not already loaded
	if squad.Constitution == nil {
		squad.Constitution, _ = LoadConstitution(name)
	}
	// Mark lead as ready if we have a constitution
	squad.LeadReady = squad.Constitution != nil

	debug.Log("squad started", "name", name, "lead_ready", squad.LeadReady)
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
