package goals

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"charm.land/fantasy"
)

// ToolName is the name of the goals tool.
const ToolName = "goals"

// ToolDescription describes what the goals tool does.
const ToolDescription = "Sets and tracks session goals for outcome-focused work management."

// GoalsParams are the parameters for the goals tool.
type GoalsParams struct {
	// Goals is the complete list of goals to set.
	Goals []GoalParam `json:"goals" jsonschema:"required,description=The list of session goals"`
}

// GoalParam represents a single goal in the input parameters.
type GoalParam struct {
	// Goal describes the desired outcome.
	Goal string `json:"goal" jsonschema:"required,description=What you're trying to achieve (outcome-focused)"`

	// Status is the current status: active, achieved, or abandoned.
	Status string `json:"status" jsonschema:"required,enum=active,enum=achieved,enum=abandoned,description=Goal status"`

	// Progress is a percentage estimate (0-100).
	Progress int `json:"progress" jsonschema:"description=Progress percentage 0-100"`

	// Milestones are checkpoints toward the goal.
	Milestones []MilestoneParam `json:"milestones,omitempty" jsonschema:"description=Checkpoints toward the goal"`

	// Notes contain context, learnings, or blockers.
	Notes []string `json:"notes,omitempty" jsonschema:"description=Context\\, learnings\\, or blockers"`
}

// MilestoneParam represents a milestone in the input parameters.
type MilestoneParam struct {
	// Description of this milestone.
	Description string `json:"description" jsonschema:"required,description=Milestone description"`

	// Completed indicates if this milestone has been reached.
	Completed bool `json:"completed" jsonschema:"description=Whether this milestone is complete"`
}

// GoalsResult contains the result of a goals operation.
type GoalsResult struct {
	// Message describes what happened.
	Message string `json:"message"`

	// Active is the count of active goals.
	Active int `json:"active"`

	// Achieved is the count of achieved goals.
	Achieved int `json:"achieved"`

	// Abandoned is the count of abandoned goals.
	Abandoned int `json:"abandoned"`

	// CurrentGoal is the description of the active goal, if any.
	CurrentGoal string `json:"current_goal,omitempty"`

	// CurrentProgress is the progress of the active goal.
	CurrentProgress int `json:"current_progress,omitempty"`
}

// newGoalsTool creates the goals tool for this plugin instance.
func (p *Plugin) newGoalsTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ToolName,
		ToolDescription,
		func(ctx context.Context, params GoalsParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return p.handleGoals(ctx, params)
		},
	)
}

// handleGoals processes a goals tool invocation.
func (p *Plugin) handleGoals(ctx context.Context, params GoalsParams) (fantasy.ToolResponse, error) {
	if p.state == nil {
		return fantasy.NewTextErrorResponse("plugin not initialized"), nil
	}

	now := time.Now()

	// Build map of existing goals by description for update detection
	existingByGoal := make(map[string]*Goal)
	for i := range p.state.Goals {
		existingByGoal[p.state.Goals[i].Goal] = &p.state.Goals[i]
	}

	// Convert params to internal Goal format
	goals := make([]Goal, len(params.Goals))
	for i, param := range params.Goals {
		status := GoalStatus(param.Status)
		if !status.IsValid() {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid status %q for goal %d; must be active, achieved, or abandoned", param.Status, i)), nil
		}

		// Validate progress
		if param.Progress < 0 || param.Progress > 100 {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid progress %d for goal %d; must be 0-100", param.Progress, i)), nil
		}

		// Convert milestones
		milestones := make([]Milestone, len(param.Milestones))
		for j, m := range param.Milestones {
			milestones[j] = Milestone{
				Description: m.Description,
				Completed:   m.Completed,
			}
		}

		// Check if this is an existing goal being updated
		createdAt := now
		if existing, ok := existingByGoal[param.Goal]; ok {
			createdAt = existing.CreatedAt
		}

		goals[i] = Goal{
			ID:         fmt.Sprintf("goal-%d", i+1),
			Goal:       param.Goal,
			Status:     status,
			Progress:   param.Progress,
			Milestones: milestones,
			Notes:      param.Notes,
			CreatedAt:  createdAt,
			UpdatedAt:  now,
		}
	}

	// Update state
	p.state.Set(goals)

	// Save state
	if err := p.state.Save(p.statePath()); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("save state: %v", err)), nil
	}

	// Build result
	counts := p.state.CountByStatus()
	result := GoalsResult{
		Message:   "Goals updated successfully.",
		Active:    counts[StatusActive],
		Achieved:  counts[StatusAchieved],
		Abandoned: counts[StatusAbandoned],
	}

	// Find the active goal
	if active := p.state.ActiveGoal(); active != nil {
		result.CurrentGoal = active.Goal
		result.CurrentProgress = active.Progress
	}

	// Format response
	var summary string
	if result.CurrentGoal != "" {
		summary = fmt.Sprintf("\nCurrent goal: %s (%d%% complete)", result.CurrentGoal, result.CurrentProgress)
	} else {
		summary = "\nNo active goal set."
	}
	summary += fmt.Sprintf("\nGoals: %d active, %d achieved, %d abandoned", result.Active, result.Achieved, result.Abandoned)
	result.Message += summary

	// Return as JSON
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("marshal result: %v", err)), nil
	}

	return fantasy.NewTextResponse(string(jsonResult)), nil
}
