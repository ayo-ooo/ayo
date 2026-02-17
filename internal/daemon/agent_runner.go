package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/tickets"
)

// DaemonAgentRunner implements AgentRunner for the daemon.
// It spawns agents to work on tickets using run.Runner.
type DaemonAgentRunner struct {
	config          config.Config
	services        *session.Services
	sandboxProvider providers.SandboxProvider
	ticketService   *tickets.Service
	logger          *slog.Logger

	// Concurrency control
	maxConcurrent int
	semaphore     chan struct{}

	// Track running executions
	mu      sync.RWMutex
	running map[string]context.CancelFunc // ticketID -> cancel
}

// DaemonAgentRunnerConfig configures the runner.
type DaemonAgentRunnerConfig struct {
	Config          config.Config
	Services        *session.Services
	SandboxProvider providers.SandboxProvider
	TicketService   *tickets.Service
	Logger          *slog.Logger
	MaxConcurrent   int // Default: 3
}

// NewDaemonAgentRunner creates a new daemon agent runner.
func NewDaemonAgentRunner(cfg DaemonAgentRunnerConfig) *DaemonAgentRunner {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 3
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &DaemonAgentRunner{
		config:          cfg.Config,
		services:        cfg.Services,
		sandboxProvider: cfg.SandboxProvider,
		ticketService:   cfg.TicketService,
		logger:          cfg.Logger,
		maxConcurrent:   cfg.MaxConcurrent,
		semaphore:       make(chan struct{}, cfg.MaxConcurrent),
		running:         make(map[string]context.CancelFunc),
	}
}

// RunWithTicket starts an agent to work on a specific ticket.
// The agent receives the ticket context and should work autonomously.
func (r *DaemonAgentRunner) RunWithTicket(
	ctx context.Context,
	agentHandle, sessionID, ticketID string,
) error {
	// 1. Acquire concurrency slot
	select {
	case r.semaphore <- struct{}{}:
		defer func() { <-r.semaphore }()
	case <-ctx.Done():
		return ctx.Err()
	}

	// 2. Track this execution
	execCtx, cancel := context.WithCancel(ctx)
	r.mu.Lock()
	r.running[ticketID] = cancel
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		delete(r.running, ticketID)
		r.mu.Unlock()
	}()

	// 3. Load ticket
	ticket, err := r.ticketService.Get(sessionID, ticketID)
	if err != nil {
		return fmt.Errorf("load ticket: %w", err)
	}

	r.logger.Info("starting agent for ticket",
		"ticket", ticketID,
		"agent", agentHandle,
		"title", ticket.Title,
	)

	// 4. Update ticket to in-progress
	ticket.Status = tickets.StatusInProgress
	if err := r.ticketService.Update(ticket); err != nil {
		r.logger.Warn("failed to update ticket status", "ticket", ticketID, "err", err)
	}

	// 5. Load agent
	ag, err := agent.Load(r.config, agentHandle)
	if err != nil {
		return fmt.Errorf("load agent: %w", err)
	}

	// 6. Create runner
	runner, err := run.NewRunner(r.config, false, run.RunnerOptions{
		Services:        r.services,
		SandboxProvider: r.sandboxProvider,
	})
	if err != nil {
		return fmt.Errorf("create runner: %w", err)
	}

	// 7. Build ticket prompt
	prompt := BuildTicketPrompt(ticket, sessionID)

	// 8. Execute
	result, err := runner.TextWithSession(execCtx, ag, prompt, nil)
	if err != nil {
		// Check if context was cancelled (ticket closed/reassigned)
		if execCtx.Err() != nil {
			r.logger.Info("agent execution cancelled", "ticket", ticketID)
			return execCtx.Err()
		}
		r.logger.Error("agent execution failed", "ticket", ticketID, "err", err)
		return err
	}

	r.logger.Info("agent completed ticket",
		"ticket", ticketID,
		"session", result.SessionID,
		"response_len", len(result.Response),
	)

	return nil
}

// Cancel cancels a running agent execution for a ticket.
func (r *DaemonAgentRunner) Cancel(ticketID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cancel, ok := r.running[ticketID]; ok {
		cancel()
	}
}

// RunningCount returns the number of currently running agents.
func (r *DaemonAgentRunner) RunningCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.running)
}
