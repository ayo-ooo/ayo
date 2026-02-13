package squads

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/tickets"
)

// AgentSpawner manages spawning agent sessions for squad tickets.
type AgentSpawner struct {
	service *Service

	mu              sync.RWMutex
	runningAgents   map[string]context.CancelFunc // key: "squad:agent" -> cancel func
	agentTickets    map[string][]string           // key: "squad:agent" -> ticket IDs
}

// NewAgentSpawner creates a new agent spawner.
func NewAgentSpawner(service *Service) *AgentSpawner {
	return &AgentSpawner{
		service:       service,
		runningAgents: make(map[string]context.CancelFunc),
		agentTickets:  make(map[string][]string),
	}
}

// SpawnAgentSession spawns an agent session in a squad sandbox to work on tickets.
// If the agent is already running for this squad, it updates the ticket list.
func (s *AgentSpawner) SpawnAgentSession(ctx context.Context, squadName, agentHandle string, ticketList []*tickets.Ticket) error {
	key := fmt.Sprintf("%s:%s", squadName, agentHandle)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Extract ticket IDs
	var ticketIDs []string
	for _, t := range ticketList {
		ticketIDs = append(ticketIDs, t.ID)
	}

	// Check if agent is already running
	if _, running := s.runningAgents[key]; running {
		// Update ticket list
		s.agentTickets[key] = ticketIDs
		debug.Log("updated agent tickets", "squad", squadName, "agent", agentHandle, "tickets", len(ticketIDs))
		return nil
	}

	// Ensure agent user exists in squad
	if err := s.service.EnsureAgentUser(ctx, squadName, agentHandle); err != nil {
		return fmt.Errorf("ensure agent user: %w", err)
	}

	// Create cancellable context for this agent session
	agentCtx, cancel := context.WithCancel(ctx)
	s.runningAgents[key] = cancel
	s.agentTickets[key] = ticketIDs

	// Start the agent session in background
	go s.runAgentSession(agentCtx, squadName, agentHandle, ticketIDs)

	debug.Log("spawned agent session", "squad", squadName, "agent", agentHandle, "tickets", len(ticketIDs))
	return nil
}

// StopAgentSession stops an agent session.
func (s *AgentSpawner) StopAgentSession(squadName, agentHandle string) {
	key := fmt.Sprintf("%s:%s", squadName, agentHandle)

	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, ok := s.runningAgents[key]; ok {
		cancel()
		delete(s.runningAgents, key)
		delete(s.agentTickets, key)
		debug.Log("stopped agent session", "squad", squadName, "agent", agentHandle)
	}
}

// StopAllForSquad stops all agent sessions for a squad.
func (s *AgentSpawner) StopAllForSquad(squadName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefix := squadName + ":"
	var toDelete []string

	for key, cancel := range s.runningAgents {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			cancel()
			toDelete = append(toDelete, key)
		}
	}

	for _, key := range toDelete {
		delete(s.runningAgents, key)
		delete(s.agentTickets, key)
	}

	if len(toDelete) > 0 {
		debug.Log("stopped all agents for squad", "squad", squadName, "count", len(toDelete))
	}
}

// RunningAgents returns the number of running agents.
func (s *AgentSpawner) RunningAgents() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.runningAgents)
}

// IsAgentRunning returns true if the agent is running for the squad.
func (s *AgentSpawner) IsAgentRunning(squadName, agentHandle string) bool {
	key := fmt.Sprintf("%s:%s", squadName, agentHandle)
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.runningAgents[key]
	return ok
}

// GetAgentTickets returns the current ticket IDs for a running agent.
func (s *AgentSpawner) GetAgentTickets(squadName, agentHandle string) []string {
	key := fmt.Sprintf("%s:%s", squadName, agentHandle)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if tickets, ok := s.agentTickets[key]; ok {
		// Return a copy
		result := make([]string, len(tickets))
		copy(result, tickets)
		return result
	}
	return nil
}

// runAgentSession runs an agent session - placeholder for actual agent execution.
// This will be integrated with the run package to actually start the agent.
func (s *AgentSpawner) runAgentSession(ctx context.Context, squadName, agentHandle string, ticketIDs []string) {
	key := fmt.Sprintf("%s:%s", squadName, agentHandle)

	defer func() {
		s.mu.Lock()
		delete(s.runningAgents, key)
		delete(s.agentTickets, key)
		s.mu.Unlock()
		debug.Log("agent session ended", "squad", squadName, "agent", agentHandle)
	}()

	// Wait for context cancellation
	// In the full implementation, this would:
	// 1. Start the agent process in the squad sandbox as the agent user
	// 2. Pass ticket context to the agent
	// 3. Monitor for completion
	<-ctx.Done()
}
