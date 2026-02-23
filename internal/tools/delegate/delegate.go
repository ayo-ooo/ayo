// Package delegate provides the delegate tool for agent-to-agent delegation within squads.
package delegate

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/squads"
)

// DelegateParams defines parameters for the delegate tool.
type DelegateParams struct {
	Agent  string `json:"agent" jsonschema:"required,description=Agent handle to delegate to (e.g. @reviewer)"`
	Prompt string `json:"prompt" jsonschema:"required,description=Prompt/task to send to the agent"`
}

// ToolConfig configures the delegate tool.
type ToolConfig struct {
	SquadName    string
	SquadAgents  []string
	Constitution *squads.Constitution
	Invoker      squads.AgentInvoker
}

// NewDelegateTool creates a new delegate tool for agent-to-agent delegation.
func NewDelegateTool(cfg ToolConfig) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"delegate",
		"Delegate a task to another agent in the squad. The agent will run in the same sandbox and receive the squad constitution.",
		func(ctx context.Context, params DelegateParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			agentHandle := params.Agent
			if !strings.HasPrefix(agentHandle, "@") {
				agentHandle = "@" + agentHandle
			}

			// Check if agent is in squad
			if len(cfg.SquadAgents) > 0 && !slices.Contains(cfg.SquadAgents, agentHandle) {
				return fantasy.NewTextErrorResponse(fmt.Sprintf(
					"agent %s is not a member of squad %s; available: %v",
					agentHandle, cfg.SquadName, cfg.SquadAgents,
				)), nil
			}

			// Invoke the agent
			result, err := cfg.Invoker.Invoke(ctx, squads.InvokeParams{
				SquadName:    cfg.SquadName,
				AgentHandle:  agentHandle,
				Prompt:       params.Prompt,
				Constitution: cfg.Constitution,
			})
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("delegation failed: %v", err)), nil
			}

			if result.Error != "" {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("agent %s returned error: %s", agentHandle, result.Error)), nil
			}

			return fantasy.NewTextResponse(fmt.Sprintf(
				"Response from %s:\n\n%s", agentHandle, result.Response,
			)), nil
		},
	)
}
