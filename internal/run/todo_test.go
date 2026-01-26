package run

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"charm.land/fantasy"
)

// newTestSessionID generates a unique session ID for tests.
func newTestSessionID() string {
	return "test-session-" + time.Now().Format("20060102150405.000000000")
}

func TestTodoTool_Init(t *testing.T) {
	tool := NewTodoTool()
	ctx := context.Background()

	t.Cleanup(func() {
		tool.Close()
	})

	err := tool.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify database was created
	if tool.DB() == nil {
		t.Error("DB() should not be nil after Init()")
	}
}

func TestTodoTool_Info(t *testing.T) {
	tool := NewTodoTool()
	info := tool.Info()

	if info.Name != TodoToolName {
		t.Errorf("Info().Name = %q, want %q", info.Name, TodoToolName)
	}

	if info.Description == "" {
		t.Error("Info().Description should not be empty")
	}

	if len(info.Required) != 1 || info.Required[0] != "todos" {
		t.Errorf("Info().Required = %v, want [todos]", info.Required)
	}
}

func TestTodoTool_Run(t *testing.T) {
	tool := NewTodoTool()
	ctx := context.Background()

	err := tool.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer tool.Close()

	// Use unique session ID to ensure clean slate
	sessionID := newTestSessionID()
	ctx = WithSessionID(ctx, sessionID)

	// Create initial todos
	params := TodoParams{
		Todos: []TodoItem{
			{Content: "Todo 1", Status: "pending", ActiveForm: "Working on todo 1"},
			{Content: "Todo 2", Status: "in_progress", ActiveForm: "Working on todo 2"},
		},
	}

	input, _ := json.Marshal(params)
	call := fantasy.ToolCall{
		ID:    "call-1",
		Name:  TodoToolName,
		Input: string(input),
	}

	resp, err := tool.Run(ctx, call)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if resp.IsError {
		t.Errorf("Run() returned error response: %s", resp.Content)
	}

	// Verify metadata
	var metadata TodoResponseMetadata
	if err := json.Unmarshal([]byte(resp.Metadata), &metadata); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if !metadata.IsNew {
		t.Error("metadata.IsNew should be true for first call")
	}

	if metadata.Total != 2 {
		t.Errorf("metadata.Total = %d, want 2", metadata.Total)
	}

	if metadata.JustStarted != "Working on todo 2" {
		t.Errorf("metadata.JustStarted = %q, want %q", metadata.JustStarted, "Working on todo 2")
	}
}

func TestTodoTool_StatusTransitions(t *testing.T) {
	tool := NewTodoTool()
	ctx := context.Background()

	err := tool.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer tool.Close()

	// Use unique session ID
	sessionID := newTestSessionID()
	ctx = WithSessionID(ctx, sessionID)

	// First call - create todos
	params1 := TodoParams{
		Todos: []TodoItem{
			{Content: "Todo A", Status: "pending", ActiveForm: "Doing A"},
			{Content: "Todo B", Status: "pending", ActiveForm: "Doing B"},
		},
	}
	input1, _ := json.Marshal(params1)
	_, err = tool.Run(ctx, fantasy.ToolCall{ID: "call-1", Name: TodoToolName, Input: string(input1)})
	if err != nil {
		t.Fatalf("First Run() error = %v", err)
	}

	// Second call - complete Todo A, start Todo B
	params2 := TodoParams{
		Todos: []TodoItem{
			{Content: "Todo A", Status: "completed", ActiveForm: "Doing A"},
			{Content: "Todo B", Status: "in_progress", ActiveForm: "Doing B"},
		},
	}
	input2, _ := json.Marshal(params2)
	resp2, err := tool.Run(ctx, fantasy.ToolCall{ID: "call-2", Name: TodoToolName, Input: string(input2)})
	if err != nil {
		t.Fatalf("Second Run() error = %v", err)
	}

	var metadata2 TodoResponseMetadata
	if err := json.Unmarshal([]byte(resp2.Metadata), &metadata2); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if metadata2.IsNew {
		t.Error("metadata.IsNew should be false for second call")
	}

	if len(metadata2.JustCompleted) != 1 || metadata2.JustCompleted[0] != "Todo A" {
		t.Errorf("metadata.JustCompleted = %v, want [Todo A]", metadata2.JustCompleted)
	}

	if metadata2.JustStarted != "Doing B" {
		t.Errorf("metadata.JustStarted = %q, want %q", metadata2.JustStarted, "Doing B")
	}
}

func TestTodoTool_InvalidStatus(t *testing.T) {
	tool := NewTodoTool()
	ctx := context.Background()

	err := tool.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer tool.Close()

	ctx = WithSessionID(ctx, newTestSessionID())

	params := TodoParams{
		Todos: []TodoItem{
			{Content: "Todo", Status: "invalid_status", ActiveForm: "Doing"},
		},
	}
	input, _ := json.Marshal(params)

	resp, err := tool.Run(ctx, fantasy.ToolCall{ID: "call-1", Name: TodoToolName, Input: string(input)})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !resp.IsError {
		t.Error("Run() should return error response for invalid status")
	}
}

func TestTodoStats(t *testing.T) {
	todos := []Todo{
		{Content: "A", Status: TodoStatusPending},
		{Content: "B", Status: TodoStatusInProgress},
		{Content: "C", Status: TodoStatusCompleted},
		{Content: "D", Status: TodoStatusCompleted},
	}

	pending, inProgress, completed := TodoStats(todos)

	if pending != 1 {
		t.Errorf("pending = %d, want 1", pending)
	}
	if inProgress != 1 {
		t.Errorf("inProgress = %d, want 1", inProgress)
	}
	if completed != 2 {
		t.Errorf("completed = %d, want 2", completed)
	}
}

func TestCurrentTodoActivity(t *testing.T) {
	tests := []struct {
		name     string
		todos    []Todo
		expected string
	}{
		{
			name:     "no todos",
			todos:    []Todo{},
			expected: "",
		},
		{
			name: "no in_progress",
			todos: []Todo{
				{Content: "A", Status: TodoStatusPending, ActiveForm: "Doing A"},
			},
			expected: "",
		},
		{
			name: "has in_progress with active_form",
			todos: []Todo{
				{Content: "A", Status: TodoStatusInProgress, ActiveForm: "Doing A"},
			},
			expected: "Doing A",
		},
		{
			name: "has in_progress without active_form",
			todos: []Todo{
				{Content: "Todo B", Status: TodoStatusInProgress, ActiveForm: ""},
			},
			expected: "Todo B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CurrentTodoActivity(tt.todos)
			if got != tt.expected {
				t.Errorf("CurrentTodoActivity() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTodoTool_NoSession(t *testing.T) {
	tool := NewTodoTool()
	ctx := context.Background()

	err := tool.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer tool.Close()

	// Don't add session ID to context
	params := TodoParams{
		Todos: []TodoItem{
			{Content: "Todo", Status: "pending", ActiveForm: "Doing"},
		},
	}
	input, _ := json.Marshal(params)

	_, err = tool.Run(ctx, fantasy.ToolCall{ID: "call-1", Name: TodoToolName, Input: string(input)})
	if err == nil {
		t.Error("Run() should return error when no session ID in context")
	}
}

// Unused import removal
var _ = os.Getenv
