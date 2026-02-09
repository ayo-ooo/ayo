// Package findagent provides a tool for @ayo to find agents capable of performing tasks.
// This enables dynamic agent discovery based on inferred capabilities.
package findagent

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/capabilities"
)

// FindAgentParams are the parameters for the find_agent tool.
type FindAgentParams struct {
	// Task is a description of the task to find an agent for.
	Task string `json:"task" jsonschema:"required,description=Description of the task to delegate or find an agent for"`

	// Count is the number of agent candidates to return.
	// Defaults to 3.
	Count int `json:"count,omitempty" jsonschema:"description=Number of agent candidates to return (default: 3)"`
}

// AgentMatch represents a matched agent for a task.
type AgentMatch struct {
	// Name is the agent handle (e.g., "@code-reviewer").
	Name string `json:"name"`

	// Similarity is the semantic similarity score (0 to 1).
	Similarity float64 `json:"similarity"`

	// Capability is the name of the matching capability.
	Capability string `json:"matching_capability"`

	// Description describes the matched capability.
	Description string `json:"description"`
}

// FindAgentResult contains the result of finding agents.
type FindAgentResult struct {
	// Agents is a list of matching agents, sorted by similarity (highest first).
	Agents []AgentMatch `json:"agents"`

	// NoMatch is true if no agents matched the task.
	NoMatch bool `json:"no_match,omitempty"`
}

func (r FindAgentResult) String() string {
	if r.NoMatch || len(r.Agents) == 0 {
		return "No matching agents found. Consider creating a new agent for this task."
	}

	var sb strings.Builder
	sb.WriteString("Matching agents:\n")
	for _, agent := range r.Agents {
		sb.WriteString(fmt.Sprintf("  %s (similarity: %.2f)\n", agent.Name, agent.Similarity))
		sb.WriteString(fmt.Sprintf("    %s: %s\n", agent.Capability, agent.Description))
	}
	return sb.String()
}

// ToolConfig configures the find_agent tool.
type ToolConfig struct {
	// Searcher is the capability searcher to use.
	Searcher *capabilities.CapabilitySearcher
}

// NewFindAgentTool creates a find_agent tool for the given configuration.
func NewFindAgentTool(cfg ToolConfig) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"find_agent",
		"Find agents capable of performing a task based on their inferred capabilities",
		func(ctx context.Context, params FindAgentParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Task == "" {
				return fantasy.NewTextErrorResponse("task is required; provide a description of what you need done"), nil
			}

			count := params.Count
			if count <= 0 {
				count = 3
			}

			// Search for matching capabilities
			results, err := cfg.Searcher.Search(ctx, params.Task, count)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("search failed: %v", err)), nil
			}

			if len(results) == 0 {
				result := FindAgentResult{NoMatch: true}
				return fantasy.NewTextResponse(result.String()), nil
			}

			// Convert to AgentMatch
			matches := make([]AgentMatch, len(results))
			for i, r := range results {
				// Ensure agent name has @ prefix
				name := r.AgentID
				if !strings.HasPrefix(name, "@") {
					name = "@" + name
				}

				matches[i] = AgentMatch{
					Name:        name,
					Similarity:  float64(r.Similarity),
					Capability:  r.Capability.Name,
					Description: r.Capability.Description,
				}
			}

			result := FindAgentResult{Agents: matches}
			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}
