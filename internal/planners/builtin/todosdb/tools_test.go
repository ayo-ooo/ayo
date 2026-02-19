package todosdb

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNewTodosTool(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

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

	// Should return error response since db is nil
	if !resp.IsError {
		t.Error("expected error response for uninitialized plugin")
	}
}

func TestHandleTodos_InvalidStatus(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	params := TodosParams{
		Todos: []TodoParam{
			{Content: "Task 1", Status: "invalid_status", ActiveForm: "Working"},
		},
	}

	resp, err := p.handleTodos(ctx, params)
	if err != nil {
		t.Fatalf("handleTodos() failed: %v", err)
	}

	if !resp.IsError {
		t.Error("expected error response for invalid status")
	}
}

func TestHandleTodos_Success(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	params := TodosParams{
		Todos: []TodoParam{
			{Content: "First task", Status: "pending", ActiveForm: "Starting first task"},
			{Content: "Second task", Status: "in_progress", ActiveForm: "Working on second"},
			{Content: "Third task", Status: "completed", ActiveForm: "Completed third"},
		},
	}

	resp, err := p.handleTodos(ctx, params)
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

	// Verify database state
	todos, err := p.ListTodos(ctx)
	if err != nil {
		t.Fatalf("ListTodos() failed: %v", err)
	}
	if len(todos) != 3 {
		t.Fatalf("database has %d todos, want 3", len(todos))
	}
	if todos[0].Content != "First task" {
		t.Errorf("todos[0].Content = %q, want %q", todos[0].Content, "First task")
	}
}

func TestHandleTodos_EmptyList(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	// First add some todos
	params := TodosParams{
		Todos: []TodoParam{
			{Content: "Task 1", Status: "pending", ActiveForm: "Working"},
		},
	}
	_, _ = p.handleTodos(ctx, params)

	// Then clear them
	params = TodosParams{Todos: []TodoParam{}}
	resp, err := p.handleTodos(ctx, params)
	if err != nil {
		t.Fatalf("handleTodos() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	// Verify database was cleared
	todos, err := p.ListTodos(ctx)
	if err != nil {
		t.Fatalf("ListTodos() failed: %v", err)
	}
	if len(todos) != 0 {
		t.Errorf("expected 0 todos, got %d", len(todos))
	}
}

func TestHandleTodos_ReplacesAllTodos(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	// Add initial todos
	params := TodosParams{
		Todos: []TodoParam{
			{Content: "Task 1", Status: "pending", ActiveForm: "Working on 1"},
			{Content: "Task 2", Status: "pending", ActiveForm: "Working on 2"},
			{Content: "Task 3", Status: "pending", ActiveForm: "Working on 3"},
		},
	}
	_, _ = p.handleTodos(ctx, params)

	// Replace with different todos
	params = TodosParams{
		Todos: []TodoParam{
			{Content: "New task A", Status: "in_progress", ActiveForm: "Working on A"},
		},
	}
	_, _ = p.handleTodos(ctx, params)

	// Verify only the new todo exists
	todos, err := p.ListTodos(ctx)
	if err != nil {
		t.Fatalf("ListTodos() failed: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	if todos[0].Content != "New task A" {
		t.Errorf("todo content = %q, want %q", todos[0].Content, "New task A")
	}
	if todos[0].Status != StatusInProgress {
		t.Errorf("todo status = %q, want %q", todos[0].Status, StatusInProgress)
	}
}
