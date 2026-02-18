package todos

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNewTodosTool(t *testing.T) {
	p := &Plugin{}
	tool := p.newTodosTool()

	info := tool.Info()
	if info.Name != ToolName {
		t.Errorf("tool name = %q, want %q", info.Name, ToolName)
	}
	if info.Description != ToolDescription {
		t.Errorf("tool description = %q, want %q", info.Description, ToolDescription)
	}
}

func TestHandleTodos_NotInitialized(t *testing.T) {
	p := &Plugin{}

	params := TodosParams{
		Todos: []TodoParam{
			{Content: "Task 1", Status: "pending", ActiveForm: "Working on task 1"},
		},
	}

	resp, err := p.handleTodos(context.Background(), params)
	if err != nil {
		t.Fatalf("handleTodos() failed: %v", err)
	}

	// Should return error response since state is nil
	if !resp.IsError {
		t.Error("expected error response for uninitialized plugin")
	}
}

func TestHandleTodos_InvalidStatus(t *testing.T) {
	tmpDir := t.TempDir()
	p := &Plugin{stateDir: tmpDir}
	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	params := TodosParams{
		Todos: []TodoParam{
			{Content: "Task 1", Status: "invalid_status", ActiveForm: "Working"},
		},
	}

	resp, err := p.handleTodos(context.Background(), params)
	if err != nil {
		t.Fatalf("handleTodos() failed: %v", err)
	}

	if !resp.IsError {
		t.Error("expected error response for invalid status")
	}
}

func TestHandleTodos_Success(t *testing.T) {
	tmpDir := t.TempDir()
	p := &Plugin{stateDir: tmpDir}
	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	params := TodosParams{
		Todos: []TodoParam{
			{Content: "First task", Status: "pending", ActiveForm: "Starting first task"},
			{Content: "Second task", Status: "in_progress", ActiveForm: "Working on second"},
			{Content: "Third task", Status: "completed", ActiveForm: "Completed third"},
		},
	}

	resp, err := p.handleTodos(context.Background(), params)
	if err != nil {
		t.Fatalf("handleTodos() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	// Parse response
	var result TodosResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Pending != 1 {
		t.Errorf("Pending = %d, want 1", result.Pending)
	}
	if result.InProgress != 1 {
		t.Errorf("InProgress = %d, want 1", result.InProgress)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1", result.Completed)
	}

	// Verify state was updated
	todos := p.state.List()
	if len(todos) != 3 {
		t.Fatalf("state has %d todos, want 3", len(todos))
	}
	if todos[0].Content != "First task" {
		t.Errorf("todos[0].Content = %q, want %q", todos[0].Content, "First task")
	}
}

func TestHandleTodos_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()
	p := &Plugin{stateDir: tmpDir}
	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// First add some todos
	params := TodosParams{
		Todos: []TodoParam{
			{Content: "Task 1", Status: "pending", ActiveForm: "Working"},
		},
	}
	_, _ = p.handleTodos(context.Background(), params)

	// Then clear them
	params = TodosParams{Todos: []TodoParam{}}
	resp, err := p.handleTodos(context.Background(), params)
	if err != nil {
		t.Fatalf("handleTodos() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	// Verify state was cleared
	if !p.state.IsEmpty() {
		t.Errorf("state should be empty, got %d todos", p.state.Count())
	}
}
