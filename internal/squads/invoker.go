package squads

import (
	"context"
)

// AgentInvoker invokes agents within a squad context.
// This interface allows the squads package to invoke agents without
// depending directly on the run package.
type AgentInvoker interface {
	// Invoke runs an agent with a prompt in the squad context.
	// The agent should receive the squad's SQUAD.md constitution in its system prompt.
	// Returns the agent's text response or an error.
	Invoke(ctx context.Context, params InvokeParams) (InvokeResult, error)
}

// InvokeParams specifies parameters for invoking an agent.
type InvokeParams struct {
	// SquadName is the name of the squad for context injection.
	SquadName string

	// AgentHandle is the agent to invoke (e.g., "@ayo").
	AgentHandle string

	// Prompt is the text prompt to send to the agent.
	Prompt string

	// Constitution is the squad's SQUAD.md constitution to inject.
	Constitution *Constitution
}

// InvokeResult contains the result of an agent invocation.
type InvokeResult struct {
	// Response is the agent's text response.
	Response string

	// Error is any error message from the agent.
	Error string
}

// NoOpInvoker is an AgentInvoker that returns routing information only.
// Used when actual agent invocation is not configured.
type NoOpInvoker struct{}

// Invoke returns a placeholder result indicating routing only.
func (n *NoOpInvoker) Invoke(ctx context.Context, params InvokeParams) (InvokeResult, error) {
	return InvokeResult{
		Response: "dispatch routed to " + params.AgentHandle + " (no invoker configured)",
	}, nil
}
