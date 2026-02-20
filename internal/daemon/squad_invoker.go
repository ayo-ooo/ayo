package daemon

import (
	"context"
	"fmt"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/planners"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/sandbox"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/squads"
)

// SquadAgentInvoker implements squads.AgentInvoker for invoking agents
// within a squad context. It uses run.Runner with SquadName set so that
// the SQUAD.md constitution is injected into the agent's system prompt.
type SquadAgentInvoker struct {
	config          config.Config
	services        *session.Services
	sandboxProvider providers.SandboxProvider
	plannerManager  *planners.SandboxPlannerManager
}

// SquadAgentInvokerConfig configures the squad agent invoker.
type SquadAgentInvokerConfig struct {
	Config          config.Config
	Services        *session.Services
	SandboxProvider providers.SandboxProvider
	PlannerManager  *planners.SandboxPlannerManager
}

// NewSquadAgentInvoker creates a new squad agent invoker.
func NewSquadAgentInvoker(cfg SquadAgentInvokerConfig) *SquadAgentInvoker {
	return &SquadAgentInvoker{
		config:          cfg.Config,
		services:        cfg.Services,
		sandboxProvider: cfg.SandboxProvider,
		plannerManager:  cfg.PlannerManager,
	}
}

// Invoke runs an agent with a prompt in the squad context.
// The agent receives the squad's SQUAD.md constitution in its system prompt
// and has access to squad-specific planners (configured via SQUAD.md frontmatter).
func (i *SquadAgentInvoker) Invoke(ctx context.Context, params squads.InvokeParams) (squads.InvokeResult, error) {
	// Normalize agent handle (ensure @ prefix)
	agentHandle := params.AgentHandle
	if len(agentHandle) > 0 && agentHandle[0] != '@' {
		agentHandle = "@" + agentHandle
	}

	// Load the agent
	ag, err := agent.Load(i.config, agentHandle)
	if err != nil {
		return squads.InvokeResult{
			Error: fmt.Sprintf("load agent %s: %v", agentHandle, err),
		}, nil
	}

	// Initialize squad-specific planners if we have a planner manager and constitution
	var squadPlanners *planners.SandboxPlanners
	if i.plannerManager != nil && params.Constitution != nil {
		// Get planner config from constitution frontmatter
		var override *config.PlannersConfig
		if params.Constitution.Frontmatter.Planners.NearTerm != "" ||
			params.Constitution.Frontmatter.Planners.LongTerm != "" {
			override = &params.Constitution.Frontmatter.Planners
		}

		// Initialize squad planners
		squadPlanners, err = sandbox.InitSquadPlanners(i.plannerManager, params.SquadName, override)
		if err != nil {
			// Log but don't fail - planners are optional
			// The agent can still work without planners
		}
	}

	// Create runner with squad context
	// The SquadName field causes constitution injection in Chat() and buildMessagesWithAttachments()
	runnerOpts := run.RunnerOptions{
		Services:        i.services,
		SandboxProvider: i.sandboxProvider,
		SquadName:       params.SquadName,
	}

	// Pass planner manager if we have squad planners initialized
	if squadPlanners != nil {
		runnerOpts.PlannerManager = i.plannerManager
	}

	runner, err := run.NewRunner(i.config, false, runnerOpts)
	if err != nil {
		return squads.InvokeResult{
			Error: fmt.Sprintf("create runner: %v", err),
		}, nil
	}

	// Execute the agent
	result, err := runner.TextWithSession(ctx, ag, params.Prompt, nil)
	if err != nil {
		return squads.InvokeResult{
			Error: fmt.Sprintf("invoke agent: %v", err),
		}, nil
	}

	return squads.InvokeResult{
		Response: result.Response,
	}, nil
}

// Ensure SquadAgentInvoker implements squads.AgentInvoker
var _ squads.AgentInvoker = (*SquadAgentInvoker)(nil)
