package todosdb

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

// TodoStatus represents the status of a todo item.
type TodoStatus string

const (
	// StatusPending indicates the todo has not been started.
	StatusPending TodoStatus = "pending"
	// StatusInProgress indicates the todo is currently being worked on.
	StatusInProgress TodoStatus = "in_progress"
	// StatusCompleted indicates the todo has been finished.
	StatusCompleted TodoStatus = "completed"
)

// IsValid returns true if the status is a known value.
func (s TodoStatus) IsValid() bool {
	return s == StatusPending || s == StatusInProgress || s == StatusCompleted
}

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
	if p.db == nil {
		return fantasy.NewTextErrorResponse("plugin not initialized"), nil
	}

	// Validate params
	for i, param := range params.Todos {
		status := TodoStatus(param.Status)
		if !status.IsValid() {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid status %q for todo %d; must be pending, in_progress, or completed", param.Status, i)), nil
		}
	}

	// Clear existing todos and insert new ones in a transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("begin transaction: %v", err)), nil
	}
	defer tx.Rollback()

	// Delete all existing todos
	if _, err := tx.ExecContext(ctx, "DELETE FROM todos"); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("clear todos: %v", err)), nil
	}

	// Insert new todos
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO todos (content, active_form, status, created_at, updated_at)
		VALUES (?, ?, ?, unixepoch(), unixepoch())
	`)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("prepare insert: %v", err)), nil
	}
	defer stmt.Close()

	var pending, inProgress, completed int
	for _, param := range params.Todos {
		if _, err := stmt.ExecContext(ctx, param.Content, param.ActiveForm, param.Status); err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("insert todo: %v", err)), nil
		}

		switch TodoStatus(param.Status) {
		case StatusPending:
			pending++
		case StatusInProgress:
			inProgress++
		case StatusCompleted:
			completed++
		}
	}

	if err := tx.Commit(); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("commit: %v", err)), nil
	}

	// Build result
	result := TodosResult{
		Message:    "Todo list updated successfully.",
		Pending:    pending,
		InProgress: inProgress,
		Completed:  completed,
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

// ListTodos retrieves all todos from the database.
func (p *Plugin) ListTodos(ctx context.Context) ([]Todo, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := p.db.QueryContext(ctx, `
		SELECT id, content, active_form, status, created_at, updated_at
		FROM todos ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query todos: %w", err)
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Content, &t.ActiveForm, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		todos = append(todos, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return todos, nil
}

// Todo represents a single todo item in the database.
type Todo struct {
	// ID is the database primary key.
	ID int64 `json:"id"`

	// Content describes what needs to be done (imperative form).
	Content string `json:"content"`

	// ActiveForm is the present continuous form shown during execution.
	ActiveForm string `json:"active_form"`

	// Status is the current status of this todo.
	Status TodoStatus `json:"status"`

	// CreatedAt is the Unix timestamp when the todo was created.
	CreatedAt int64 `json:"created_at"`

	// UpdatedAt is the Unix timestamp when the todo was last updated.
	UpdatedAt int64 `json:"updated_at"`
}
