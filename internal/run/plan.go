package run

import (
	"context"
	_ "embed"
	"fmt"

	"charm.land/fantasy"
	"github.com/alexcabrera/ayo/internal/session"
)

//go:embed plan.md
var planDescription string

// PlanToolName is the name of the plan tool.
const PlanToolName = "plan"

// PlanParams defines the parameters for the plan tool.
// Plans can be structured as:
// - Just tasks (flat list) - use "tasks" field
// - Phases containing tasks (must have 2+ phases) - use "phases" field
// - Tasks containing todos (atomic sub-items) - use "todos" within tasks
type PlanParams struct {
	Phases []PhaseParam `json:"phases,omitempty" description:"Optional high-level phases (requires 2+ if used)"`
	Tasks  []TaskParam  `json:"tasks,omitempty" description:"Top-level tasks (when not using phases)"`
}

// PhaseParam represents a phase from the LLM.
type PhaseParam struct {
	Name   string      `json:"name" description:"Phase name (e.g., 'Phase 1: Setup')"`
	Status string      `json:"status" description:"Phase status: pending, in_progress, or completed"`
	Tasks  []TaskParam `json:"tasks" description:"Tasks within this phase (at least 1 required)"`
}

// TaskParam represents a task from the LLM.
type TaskParam struct {
	Content    string      `json:"content" description:"What needs to be done (imperative form)"`
	ActiveForm string      `json:"active_form" description:"Present continuous form (e.g., 'Running tests')"`
	Status     string      `json:"status" description:"Task status: pending, in_progress, or completed"`
	Todos      []TodoParam `json:"todos,omitempty" description:"Optional atomic sub-items within this task"`
}

// TodoParam represents a todo item from the LLM.
type TodoParam struct {
	Content    string `json:"content" description:"What needs to be done (imperative form)"`
	ActiveForm string `json:"active_form" description:"Present continuous form"`
	Status     string `json:"status" description:"Todo status: pending, in_progress, or completed"`
}

// PlanResponseMetadata contains metadata about plan changes for UI rendering.
type PlanResponseMetadata struct {
	IsNew         bool         `json:"is_new"`
	Plan          session.Plan `json:"plan"`
	JustCompleted []string     `json:"just_completed,omitempty"`
	JustStarted   string       `json:"just_started,omitempty"`
	Completed     int          `json:"completed"`
	Total         int          `json:"total"`
}

// NewPlanTool creates the plan tool for Fantasy.
func NewPlanTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		PlanToolName,
		planDescription,
		func(ctx context.Context, params PlanParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			sessionID := GetSessionIDFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("plan tool requires a session; session ID not found in context")
			}

			services := GetServicesFromContext(ctx)
			if services == nil {
				return fantasy.ToolResponse{}, fmt.Errorf("plan tool requires session services; not found in context")
			}

			// Get current session to compare old vs new plan
			currentSession, err := services.Sessions.Get(ctx, sessionID)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to get session: %w", err)
			}

			isNew := currentSession.Plan.IsEmpty()

			// Validate the plan structure
			if err := validatePlanParams(params); err != nil {
				return fantasy.ToolResponse{}, err
			}

			// Convert params to session.Plan
			plan := convertParamsToPlan(params)

			// Build old status map for change detection
			oldStatusMap := buildStatusMap(currentSession.Plan)

			// Detect changes
			justCompleted, justStarted := detectChanges(plan, oldStatusMap)

			// Save updated plan
			_, err = services.Sessions.UpdatePlan(ctx, sessionID, plan)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to save plan: %w", err)
			}

			// Get stats
			pending, inProgress, completed := plan.Stats()
			total := pending + inProgress + completed

			response := "Plan updated successfully.\n\n"
			response += fmt.Sprintf("Status: %d pending, %d in progress, %d completed\n",
				pending, inProgress, completed)
			response += "Plan has been modified successfully. Continue to use the plan tool to track your progress. Proceed with the current tasks if applicable."

			metadata := PlanResponseMetadata{
				IsNew:         isNew,
				Plan:          plan,
				JustCompleted: justCompleted,
				JustStarted:   justStarted,
				Completed:     completed,
				Total:         total,
			}

			return fantasy.WithResponseMetadata(fantasy.NewTextResponse(response), metadata), nil
		},
	)
}

// validatePlanParams validates the plan structure.
func validatePlanParams(params PlanParams) error {
	// Check for mutual exclusivity (can't have both phases and top-level tasks)
	if len(params.Phases) > 0 && len(params.Tasks) > 0 {
		return fmt.Errorf("plan cannot have both phases and top-level tasks; use phases to group tasks or use tasks directly")
	}

	// If phases are used, must have at least 2
	if len(params.Phases) == 1 {
		return fmt.Errorf("if using phases, must have at least 2 phases; got 1")
	}

	// Validate phases
	for i, phase := range params.Phases {
		if phase.Name == "" {
			return fmt.Errorf("phase %d: name is required", i+1)
		}
		if err := validateStatus(phase.Status, fmt.Sprintf("phase %q", phase.Name)); err != nil {
			return err
		}
		if len(phase.Tasks) == 0 {
			return fmt.Errorf("phase %q: must have at least 1 task", phase.Name)
		}
		for j, task := range phase.Tasks {
			if err := validateTask(task, fmt.Sprintf("phase %q task %d", phase.Name, j+1)); err != nil {
				return err
			}
		}
	}

	// Validate top-level tasks
	for i, task := range params.Tasks {
		if err := validateTask(task, fmt.Sprintf("task %d", i+1)); err != nil {
			return err
		}
	}

	return nil
}

func validateTask(task TaskParam, context string) error {
	if task.Content == "" {
		return fmt.Errorf("%s: content is required", context)
	}
	if err := validateStatus(task.Status, context); err != nil {
		return err
	}
	for i, todo := range task.Todos {
		if todo.Content == "" {
			return fmt.Errorf("%s todo %d: content is required", context, i+1)
		}
		if err := validateStatus(todo.Status, fmt.Sprintf("%s todo %d", context, i+1)); err != nil {
			return err
		}
	}
	return nil
}

func validateStatus(status, context string) error {
	switch status {
	case "pending", "in_progress", "completed":
		return nil
	default:
		return fmt.Errorf("%s: invalid status %q; must be pending, in_progress, or completed", context, status)
	}
}

// convertParamsToPlan converts LLM params to session.Plan.
func convertParamsToPlan(params PlanParams) session.Plan {
	var plan session.Plan

	// Convert phases
	for _, p := range params.Phases {
		phase := session.Phase{
			Name:   p.Name,
			Status: session.PlanStatus(p.Status),
		}
		for _, t := range p.Tasks {
			phase.Tasks = append(phase.Tasks, convertTask(t))
		}
		plan.Phases = append(plan.Phases, phase)
	}

	// Convert top-level tasks
	for _, t := range params.Tasks {
		plan.Tasks = append(plan.Tasks, convertTask(t))
	}

	return plan
}

func convertTask(t TaskParam) session.Task {
	task := session.Task{
		Content:    t.Content,
		ActiveForm: t.ActiveForm,
		Status:     session.PlanStatus(t.Status),
	}
	for _, td := range t.Todos {
		task.Todos = append(task.Todos, session.Todo{
			Content:    td.Content,
			ActiveForm: td.ActiveForm,
			Status:     session.PlanStatus(td.Status),
		})
	}
	return task
}

// buildStatusMap creates a map of content -> status for change detection.
func buildStatusMap(plan session.Plan) map[string]session.PlanStatus {
	statusMap := make(map[string]session.PlanStatus)

	for _, phase := range plan.Phases {
		statusMap["phase:"+phase.Name] = phase.Status
		for _, task := range phase.Tasks {
			statusMap["task:"+task.Content] = task.Status
			for _, todo := range task.Todos {
				statusMap["todo:"+todo.Content] = todo.Status
			}
		}
	}

	for _, task := range plan.Tasks {
		statusMap["task:"+task.Content] = task.Status
		for _, todo := range task.Todos {
			statusMap["todo:"+todo.Content] = todo.Status
		}
	}

	return statusMap
}

// detectChanges finds what was just completed and what just started.
func detectChanges(plan session.Plan, oldStatusMap map[string]session.PlanStatus) (justCompleted []string, justStarted string) {
	// Check phases
	for _, phase := range plan.Phases {
		oldStatus := oldStatusMap["phase:"+phase.Name]
		if phase.Status == session.PlanStatusCompleted && oldStatus != session.PlanStatusCompleted {
			justCompleted = append(justCompleted, phase.Name)
		}
		if phase.Status == session.PlanStatusInProgress && oldStatus != session.PlanStatusInProgress {
			justStarted = phase.Name
		}

		for _, task := range phase.Tasks {
			jc, js := detectTaskChanges(task, oldStatusMap)
			justCompleted = append(justCompleted, jc...)
			if js != "" {
				justStarted = js
			}
		}
	}

	// Check top-level tasks
	for _, task := range plan.Tasks {
		jc, js := detectTaskChanges(task, oldStatusMap)
		justCompleted = append(justCompleted, jc...)
		if js != "" {
			justStarted = js
		}
	}

	return
}

func detectTaskChanges(task session.Task, oldStatusMap map[string]session.PlanStatus) (justCompleted []string, justStarted string) {
	oldStatus := oldStatusMap["task:"+task.Content]
	if task.Status == session.PlanStatusCompleted && oldStatus != session.PlanStatusCompleted {
		justCompleted = append(justCompleted, task.Content)
	}
	if task.Status == session.PlanStatusInProgress && oldStatus != session.PlanStatusInProgress {
		if task.ActiveForm != "" {
			justStarted = task.ActiveForm
		} else {
			justStarted = task.Content
		}
	}

	for _, todo := range task.Todos {
		oldTodoStatus := oldStatusMap["todo:"+todo.Content]
		if todo.Status == session.PlanStatusCompleted && oldTodoStatus != session.PlanStatusCompleted {
			justCompleted = append(justCompleted, todo.Content)
		}
		if todo.Status == session.PlanStatusInProgress && oldTodoStatus != session.PlanStatusInProgress {
			if todo.ActiveForm != "" {
				justStarted = todo.ActiveForm
			} else {
				justStarted = todo.Content
			}
		}
	}

	return
}
