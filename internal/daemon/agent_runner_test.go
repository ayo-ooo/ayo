package daemon

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/tickets"
)

func TestBuildTicketPrompt(t *testing.T) {
	tests := []struct {
		name     string
		ticket   *tickets.Ticket
		contains []string
	}{
		{
			name: "basic ticket",
			ticket: &tickets.Ticket{
				ID:       "test-001",
				Title:    "Fix the bug",
				Priority: 1,
			},
			contains: []string{
				"# Ticket Assignment",
				"**Ticket ID:** test-001",
				"**Title:** Fix the bug",
				"**Priority:** P1",
				"`ayo ticket close test-001`",
			},
		},
		{
			name: "ticket with description and tags",
			ticket: &tickets.Ticket{
				ID:          "test-002",
				Title:       "Add feature",
				Priority:    2,
				Description: "Implement the new feature as described.",
				Tags:        []string{"feature", "backend"},
			},
			contains: []string{
				"**Tags:** feature, backend",
				"## Description",
				"Implement the new feature as described.",
			},
		},
		{
			name: "ticket with dependencies",
			ticket: &tickets.Ticket{
				ID:       "test-003",
				Title:    "Integration test",
				Priority: 2,
				Deps:     []string{"test-001", "test-002"},
			},
			contains: []string{
				"## Dependencies",
				"- test-001",
				"- test-002",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildTicketPrompt(tt.ticket, "session-123")

			for _, want := range tt.contains {
				if !strings.Contains(prompt, want) {
					t.Errorf("prompt missing expected content %q\n\nGot:\n%s", want, prompt)
				}
			}
		})
	}
}

func TestBuildTicketPrompt_Instructions(t *testing.T) {
	ticket := &tickets.Ticket{
		ID:       "test-inst",
		Title:    "Test instructions",
		Priority: 2,
	}

	prompt := BuildTicketPrompt(ticket, "session-456")

	// Should contain instructions
	if !strings.Contains(prompt, "## Instructions") {
		t.Error("missing instructions section")
	}
	if !strings.Contains(prompt, "Work autonomously") {
		t.Error("missing autonomy instruction")
	}
	if !strings.Contains(prompt, "ayo ticket close test-inst") {
		t.Error("missing close instruction")
	}
	if !strings.Contains(prompt, "ayo ticket update test-inst --status blocked") {
		t.Error("missing blocked instruction")
	}
}

// MockAgentRunner implements AgentRunner for testing
type MockAgentRunner struct {
	RunCalls  atomic.Int32
	LastAgent string
	LastTicket string
	ShouldFail bool
}

func (m *MockAgentRunner) RunWithTicket(ctx context.Context, agentHandle, sessionID, ticketID string) error {
	m.RunCalls.Add(1)
	m.LastAgent = agentHandle
	m.LastTicket = ticketID
	
	if m.ShouldFail {
		return context.Canceled
	}
	return nil
}

func TestDaemonAgentRunner_Concurrency(t *testing.T) {
	runner := NewDaemonAgentRunner(DaemonAgentRunnerConfig{
		MaxConcurrent: 2,
	})

	if runner.maxConcurrent != 2 {
		t.Errorf("expected maxConcurrent=2, got %d", runner.maxConcurrent)
	}

	if cap(runner.semaphore) != 2 {
		t.Errorf("expected semaphore capacity=2, got %d", cap(runner.semaphore))
	}
}

func TestDaemonAgentRunner_DefaultConcurrency(t *testing.T) {
	runner := NewDaemonAgentRunner(DaemonAgentRunnerConfig{})

	if runner.maxConcurrent != 3 {
		t.Errorf("expected default maxConcurrent=3, got %d", runner.maxConcurrent)
	}
}

func TestDaemonAgentRunner_RunningCount(t *testing.T) {
	runner := NewDaemonAgentRunner(DaemonAgentRunnerConfig{
		MaxConcurrent: 5,
	})

	if runner.RunningCount() != 0 {
		t.Errorf("expected 0 running, got %d", runner.RunningCount())
	}
}

func TestDaemonAgentRunner_Cancel(t *testing.T) {
	runner := NewDaemonAgentRunner(DaemonAgentRunnerConfig{})

	// Add a mock cancel func
	called := false
	runner.mu.Lock()
	runner.running["test-ticket"] = func() { called = true }
	runner.mu.Unlock()

	// Cancel it
	runner.Cancel("test-ticket")

	if !called {
		t.Error("expected cancel func to be called")
	}
}

func TestDaemonAgentRunner_CancelNonexistent(t *testing.T) {
	runner := NewDaemonAgentRunner(DaemonAgentRunnerConfig{})

	// Should not panic
	runner.Cancel("nonexistent-ticket")
}

func TestTicketWatcher_Creation(t *testing.T) {
	mock := &MockAgentRunner{}
	
	watcher, err := NewTicketWatcher(TicketWatcherConfig{
		Runner: mock,
	})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Stop(context.Background())

	if watcher.runner != mock {
		t.Error("runner not set correctly")
	}
}

func TestTicketWatcher_Counts(t *testing.T) {
	watcher, err := NewTicketWatcher(TicketWatcherConfig{})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Stop(context.Background())

	if watcher.RunningAgents() != 0 {
		t.Errorf("expected 0 running agents, got %d", watcher.RunningAgents())
	}

	if watcher.WatchedSessions() != 0 {
		t.Errorf("expected 0 watched sessions, got %d", watcher.WatchedSessions())
	}

	if watcher.WatchedSquads() != 0 {
		t.Errorf("expected 0 watched squads, got %d", watcher.WatchedSquads())
	}
}

func TestTicketWatcher_StartStop(t *testing.T) {
	watcher, err := NewTicketWatcher(TicketWatcherConfig{})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start should not error
	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// Stop should not error
	if err := watcher.Stop(ctx); err != nil {
		t.Fatalf("failed to stop watcher: %v", err)
	}
}
