package run

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/tools"
)

//go:embed todo.md
var todoDescription string

// TodoToolName is the name of the todo tool.
const TodoToolName = "todo"

// TodoParams defines the parameters for the todo tool.
type TodoParams struct {
	Todos []TodoItem `json:"todos" description:"The updated todo list"`
}

// TodoItem represents a single todo item.
type TodoItem struct {
	Content    string `json:"content" description:"What needs to be done (imperative form)"`
	Status     string `json:"status" description:"Todo status: pending, in_progress, or completed"`
	ActiveForm string `json:"active_form" description:"Present continuous form (e.g., 'Running tests')"`
}

// TodoStatus represents the status of a todo item.
type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"
)

// Todo represents a todo item as stored in the database.
type Todo struct {
	Content    string     `json:"content"`
	Status     TodoStatus `json:"status"`
	ActiveForm string     `json:"active_form"`
}

// TodoResponseMetadata contains metadata about todo changes for UI rendering.
type TodoResponseMetadata struct {
	IsNew         bool     `json:"is_new"`
	Todos         []Todo   `json:"todos"`
	JustCompleted []string `json:"just_completed,omitempty"`
	JustStarted   string   `json:"just_started,omitempty"`
	Completed     int      `json:"completed"`
	Total         int      `json:"total"`
}

// todoTool is a stateful tool that manages todos with its own database.
type todoTool struct {
	tools.StatefulToolBase
}

// NewTodoTool creates a new todo tool instance.
func NewTodoTool() *todoTool {
	return &todoTool{
		StatefulToolBase: tools.NewStatefulToolBase(TodoToolName),
	}
}

// todoSchema is the SQL schema for the todo tool's database.
const todoSchema = `
CREATE TABLE IF NOT EXISTS todos (
	session_id TEXT PRIMARY KEY,
	data TEXT NOT NULL DEFAULT '[]',
	created_at INTEGER NOT NULL DEFAULT (unixepoch()),
	updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_todos_updated_at ON todos(updated_at);
`

// Init initializes the todo tool's database.
func (t *todoTool) Init(ctx context.Context) error {
	_, err := t.OpenDatabase(ctx)
	if err != nil {
		return err
	}
	return t.RunMigration(ctx, todoSchema)
}

// Info returns the tool info for Fantasy.
func (t *todoTool) Info() fantasy.ToolInfo {
	return fantasy.ToolInfo{
		Name:        TodoToolName,
		Description: todoDescription,
		Parameters: map[string]any{
			"todos": map[string]any{
				"type":        "array",
				"description": "The updated todo list",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"content": map[string]any{
							"type":        "string",
							"description": "What needs to be done (imperative form)",
						},
						"status": map[string]any{
							"type":        "string",
							"description": "Todo status: pending, in_progress, or completed",
							"enum":        []string{"pending", "in_progress", "completed"},
						},
						"active_form": map[string]any{
							"type":        "string",
							"description": "Present continuous form (e.g., 'Running tests')",
						},
					},
					"required": []string{"content", "status", "active_form"},
				},
			},
		},
		Required: []string{"todos"},
	}
}

// Run executes the todo tool.
func (t *todoTool) Run(ctx context.Context, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	// Parse parameters
	var params TodoParams
	if err := json.Unmarshal([]byte(call.Input), &params); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid parameters: %v", err)), nil
	}

	// Get session ID from context
	sessionID := GetSessionIDFromContext(ctx)
	if sessionID == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("todo tool requires a session; session ID not found in context")
	}

	// Initialize database if not already done
	if t.DB() == nil {
		if err := t.Init(ctx); err != nil {
			return fantasy.ToolResponse{}, fmt.Errorf("failed to initialize todo database: %w", err)
		}
	}

	// Get current todos for this session
	currentTodos, err := t.getTodos(ctx, sessionID)
	if err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to get todos: %w", err)
	}

	isNew := len(currentTodos) == 0

	// Build map of old statuses for change detection
	oldStatusByContent := make(map[string]TodoStatus)
	for _, todo := range currentTodos {
		oldStatusByContent[todo.Content] = todo.Status
	}

	// Validate and convert params
	for _, item := range params.Todos {
		switch item.Status {
		case "pending", "in_progress", "completed":
		default:
			return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid status %q for todo %q", item.Status, item.Content)), nil
		}
	}

	// Convert to Todo slice and detect changes
	todos := make([]Todo, len(params.Todos))
	var justCompleted []string
	var justStarted string
	completedCount := 0
	pendingCount := 0
	inProgressCount := 0

	for i, item := range params.Todos {
		todos[i] = Todo{
			Content:    item.Content,
			Status:     TodoStatus(item.Status),
			ActiveForm: item.ActiveForm,
		}

		newStatus := TodoStatus(item.Status)
		oldStatus, existed := oldStatusByContent[item.Content]

		switch newStatus {
		case TodoStatusCompleted:
			completedCount++
			if existed && oldStatus != TodoStatusCompleted {
				justCompleted = append(justCompleted, item.Content)
			}
		case TodoStatusInProgress:
			inProgressCount++
			if !existed || oldStatus != TodoStatusInProgress {
				if item.ActiveForm != "" {
					justStarted = item.ActiveForm
				} else {
					justStarted = item.Content
				}
			}
		case TodoStatusPending:
			pendingCount++
		}
	}

	// Save todos
	if err := t.saveTodos(ctx, sessionID, todos); err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to save todos: %w", err)
	}

	// Build response
	response := "Todo list updated successfully.\n\n"
	response += fmt.Sprintf("Status: %d pending, %d in progress, %d completed\n",
		pendingCount, inProgressCount, completedCount)
	response += "Todos have been modified successfully. Ensure that you continue to use the todo list to track your progress. Please proceed with the current todos if applicable."

	metadata := TodoResponseMetadata{
		IsNew:         isNew,
		Todos:         todos,
		JustCompleted: justCompleted,
		JustStarted:   justStarted,
		Completed:     completedCount,
		Total:         len(todos),
	}

	return fantasy.WithResponseMetadata(fantasy.NewTextResponse(response), metadata), nil
}

// ProviderOptions returns empty provider options.
func (t *todoTool) ProviderOptions() fantasy.ProviderOptions {
	return fantasy.ProviderOptions{}
}

// SetProviderOptions is a no-op for this tool.
func (t *todoTool) SetProviderOptions(opts fantasy.ProviderOptions) {}

// getTodos retrieves todos for a session from the database.
func (t *todoTool) getTodos(ctx context.Context, sessionID string) ([]Todo, error) {
	db := t.DB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var data string
	err := db.QueryRowContext(ctx,
		"SELECT data FROM todos WHERE session_id = ?",
		sessionID,
	).Scan(&data)

	if err != nil {
		// Not found is fine - return empty list
		return []Todo{}, nil
	}

	var todos []Todo
	if err := json.Unmarshal([]byte(data), &todos); err != nil {
		return nil, fmt.Errorf("unmarshal todos: %w", err)
	}

	return todos, nil
}

// saveTodos saves todos for a session to the database.
func (t *todoTool) saveTodos(ctx context.Context, sessionID string, todos []Todo) error {
	db := t.DB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	data, err := json.Marshal(todos)
	if err != nil {
		return fmt.Errorf("marshal todos: %w", err)
	}

	_, err = db.ExecContext(ctx,
		`INSERT INTO todos (session_id, data, updated_at) 
		 VALUES (?, ?, unixepoch())
		 ON CONFLICT(session_id) DO UPDATE SET 
		   data = excluded.data,
		   updated_at = unixepoch()`,
		sessionID, string(data),
	)
	if err != nil {
		return fmt.Errorf("upsert todos: %w", err)
	}

	return nil
}

// GetTodosForSession retrieves todos for a session (used by UI).
func GetTodosForSession(ctx context.Context, sessionID string) ([]Todo, error) {
	tool := NewTodoTool()
	if err := tool.Init(ctx); err != nil {
		return nil, err
	}
	defer tool.Close()

	return tool.getTodos(ctx, sessionID)
}

// CurrentTodoActivity returns the active_form of the current in-progress todo.
func CurrentTodoActivity(todos []Todo) string {
	for _, todo := range todos {
		if todo.Status == TodoStatusInProgress {
			if todo.ActiveForm != "" {
				return todo.ActiveForm
			}
			return todo.Content
		}
	}
	return ""
}

// TodoStats returns counts of pending, in-progress, and completed todos.
func TodoStats(todos []Todo) (pending, inProgress, completed int) {
	for _, todo := range todos {
		switch todo.Status {
		case TodoStatusPending:
			pending++
		case TodoStatusInProgress:
			inProgress++
		case TodoStatusCompleted:
			completed++
		}
	}
	return
}
