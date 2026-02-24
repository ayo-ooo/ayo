// Package daemon provides flow invocation support.
// This file implements AgentInvoker variants for different contexts.
package daemon

import (
	"context"
	"fmt"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/run"
)

// AgentInvoker is the interface for invoking agents from flow steps.
// NOTE: This interface differs from squads.AgentInvoker intentionally.
// The squads interface uses InvokeParams/InvokeResult for constitution injection,
// while this interface provides a simpler string-based API for flow orchestration.
// The duplication also avoids an import cycle (flows depends on daemon).
type AgentInvoker interface {
	Invoke(ctx context.Context, agent, prompt string) (string, error)
	InvokeInSquad(ctx context.Context, squad, agent, prompt string) (string, error)
}

// SandboxAwareInvoker invokes agents via the daemon client.
// Use this when invoking agents from external code that connects to the daemon.
type SandboxAwareInvoker struct {
	client *Client
}

// NewSandboxAwareInvoker creates an invoker that uses the daemon for agent invocation.
func NewSandboxAwareInvoker(client *Client) AgentInvoker {
	return &SandboxAwareInvoker{client: client}
}

// Invoke sends a prompt to an agent and returns the response.
// The agent is invoked via the daemon which manages sandbox context.
func (i *SandboxAwareInvoker) Invoke(ctx context.Context, agentName, prompt string) (string, error) {
	if i.client == nil {
		return "", fmt.Errorf("daemon client not initialized")
	}

	result, err := i.client.AgentInvoke(ctx, AgentInvokeParams{
		Agent:  agentName,
		Prompt: prompt,
	})
	if err != nil {
		return "", fmt.Errorf("invoke agent %q: %w", agentName, err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("agent %q error: %s", agentName, result.Error)
	}

	return result.Response, nil
}

// InvokeInSquad invokes an agent within a specific squad's sandbox context.
func (i *SandboxAwareInvoker) InvokeInSquad(ctx context.Context, squad, agentName, prompt string) (string, error) {
	if i.client == nil {
		return "", fmt.Errorf("daemon client not initialized")
	}

	// Normalize squad name (remove # prefix if present)
	squadName := squad
	if len(squadName) > 0 && squadName[0] == '#' {
		squadName = squadName[1:]
	}

	// Use squad dispatch with agent targeting
	result, err := i.client.SquadDispatch(ctx, SquadDispatchParams{
		Name:           squadName,
		Prompt:         fmt.Sprintf("@%s %s", normalizeAgentHandle(agentName), prompt),
		StartIfStopped: true,
	})
	if err != nil {
		return "", fmt.Errorf("invoke agent %q in squad %q: %w", agentName, squadName, err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("squad %q error: %s", squadName, result.Error)
	}

	return result.Raw, nil
}

// normalizeAgentHandle removes @ prefix if present.
func normalizeAgentHandle(handle string) string {
	if len(handle) > 0 && handle[0] == '@' {
		return handle[1:]
	}
	return handle
}

// ServerAgentInvoker invokes agents directly within the daemon server context.
// Use this when invoking agents from daemon RPC handlers (e.g., flow execution).
type ServerAgentInvoker struct {
	config config.Config
}

// NewServerAgentInvoker creates an invoker for server-side agent execution.
func NewServerAgentInvoker(cfg config.Config) AgentInvoker {
	return &ServerAgentInvoker{config: cfg}
}

// Invoke sends a prompt to an agent and returns the response.
// The agent is loaded and executed directly without going through RPC.
func (i *ServerAgentInvoker) Invoke(ctx context.Context, agentName, prompt string) (string, error) {
	// Normalize agent handle
	agentHandle := agentName
	if len(agentHandle) > 0 && agentHandle[0] != '@' {
		agentHandle = "@" + agentHandle
	}

	// Load the agent
	ag, err := agent.Load(i.config, agentHandle)
	if err != nil {
		return "", fmt.Errorf("load agent %s: %w", agentHandle, err)
	}

	// Create runner
	runner, err := run.NewRunnerFromConfig(i.config, false)
	if err != nil {
		return "", fmt.Errorf("create runner: %w", err)
	}

	// Execute the agent
	result, err := runner.TextWithSession(ctx, ag, prompt, nil)
	if err != nil {
		return "", fmt.Errorf("invoke agent: %w", err)
	}

	return result.Response, nil
}

// InvokeInSquad invokes an agent within a specific squad's sandbox context.
// For server-side invocation, this requires the squad RPC system which is
// handled by dispatching to the squad. The ServerAgentInvoker needs a client
// to dispatch to squads, so it stores one.
func (i *ServerAgentInvoker) InvokeInSquad(ctx context.Context, squad, agentName, prompt string) (string, error) {
	// Server-side squad invocation is not directly supported
	// because we can't dispatch from within a dispatch.
	// Instead, return an error instructing to use the client-side invoker
	// for squad-based invocations.
	return "", fmt.Errorf("server-side InvokeInSquad not supported; use SandboxAwareInvoker for squad context")
}
