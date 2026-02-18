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
// Duplicated here to avoid import cycle (flows depends on daemon for execution).
type AgentInvoker interface {
	Invoke(ctx context.Context, agent, prompt string) (string, error)
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
