package todos

import (
	"context"
	"encoding/json"
	"fmt"

	"charm.land/fantasy"
)

// ToolName is the name of the todos tool.
const ToolName = "todos"

// ToolDescription describes what the todos tool does.
const ToolDescription = "Creates and manages a structured task list for tracking progress on complex, multi-step coding tasks."

// TodosParams are the parameters for the todos tool.
type TodosParams struct {
	// Todos is the complete list of todos to set.
	// The tool replaces all existing todos with this list.
	Todos []TodoParam `json:"todos" jsonschema:"required,description=The updated todo list"`
}

// TodoParam represents a single todo in the input parameters.
type TodoParam struct {
	// Content describes what needs to be done (imperative form).
	Content string `json:"content" jsonschema:"required,description=What needs to be done (imperative form)"`

	// Status is the current status: pending, in_progress, or completed.
	Status string `json:"status" jsonschema:"required,enum=pending,enum=in_progress,enum=completed,description=Task status: pending\\, in_progress\\, or completed"`

	// ActiveForm is the present continuous form shown during execution.
	ActiveForm string `json:"active_form" jsonschema:"required,description=Present continuous form (e.g.\\, 'Running tests')"`
}

// TodosResult contains the result of a todos operation.
type TodosResult struct {
	// Message describes what happened.
	Message string `json:"message"`

	// Pending is the count of pending todos.
	Pending int `json:"pending"`

	// InProgress is the count of in-progress todos.
	InProgress int `json:"in_progress"`

	// Completed is the count of completed todos.
	Completed int `json:"completed"`
}

// newTodosTool creates the todos tool for this plugin instance.
func (p *Plugin) newTodosTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ToolName,
		ToolDescription,
		func(ctx context.Context, params TodosParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return p.handleTodos(ctx, params)
		},
	)
}

// handleTodos processes a todos tool invocation.
func (p *Plugin) handleTodos(ctx context.Context, params TodosParams) (fantasy.ToolResponse, error) {
	if p.state == nil {
		return fantasy.NewTextErrorResponse("plugin not initialized"), nil
	}

	// Convert params to internal Todo format
	todos := make([]Todo, len(params.Todos))
	for i, param := range params.Todos {
		status := TodoStatus(param.Status)
		if !status.IsValid() {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid status %q for todo %d; must be pending, in_progress, or completed", param.Status, i)), nil
		}

		todos[i] = Todo{
			ID:         fmt.Sprintf("%d", i+1),
			Content:    param.Content,
			ActiveForm: param.ActiveForm,
			Status:     status,
		}
	}

	// Update state
	p.state.Set(todos)

	// Save state
	if err := p.state.Save(p.statePath()); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("save state: %v", err)), nil
	}

	// Build result
	counts := p.state.CountByStatus()
	result := TodosResult{
		Message:    "Todo list updated successfully.",
		Pending:    counts[StatusPending],
		InProgress: counts[StatusInProgress],
		Completed:  counts[StatusCompleted],
	}

	// Format response
	summary := fmt.Sprintf("\nStatus: %d pending, %d in progress, %d completed",
		result.Pending, result.InProgress, result.Completed)
	result.Message += summary

	// Return as JSON
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("marshal result: %v", err)), nil
	}

	return fantasy.NewTextResponse(string(jsonResult)), nil
}
