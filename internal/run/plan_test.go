package run

import (
	"context"
	"testing"

	"charm.land/fantasy"
	"github.com/alexcabrera/ayo/internal/session"
)

func TestPlanTool_NoSessionID(t *testing.T) {
	tool := NewPlanTool()

	ctx := context.Background()

	_, err := tool.Run(ctx, fantasy.ToolCall{
		ID:    "test-call",
		Name:  "plan",
		Input: `{"tasks":[{"content":"Test task","status":"pending","active_form":"Testing task"}]}`,
	})

	if err == nil {
		t.Error("expected error when session ID is missing")
	}
}

func TestPlanTool_NoServices(t *testing.T) {
	tool := NewPlanTool()

	ctx := WithSessionID(context.Background(), "test-session")

	_, err := tool.Run(ctx, fantasy.ToolCall{
		ID:    "test-call",
		Name:  "plan",
		Input: `{"tasks":[{"content":"Test task","status":"pending","active_form":"Testing task"}]}`,
	})

	if err == nil {
		t.Error("expected error when services are missing")
	}
}

func TestValidatePlanParams_ValidFlatTasks(t *testing.T) {
	params := PlanParams{
		Tasks: []TaskParam{
			{Content: "Task 1", ActiveForm: "Doing task 1", Status: "pending"},
			{Content: "Task 2", ActiveForm: "Doing task 2", Status: "in_progress"},
		},
	}

	if err := validatePlanParams(params); err != nil {
		t.Errorf("validatePlanParams() error = %v", err)
	}
}

func TestValidatePlanParams_ValidPhases(t *testing.T) {
	params := PlanParams{
		Phases: []PhaseParam{
			{
				Name:   "Phase 1: Setup",
				Status: "completed",
				Tasks:  []TaskParam{{Content: "Init", ActiveForm: "Initializing", Status: "completed"}},
			},
			{
				Name:   "Phase 2: Build",
				Status: "in_progress",
				Tasks:  []TaskParam{{Content: "Build", ActiveForm: "Building", Status: "in_progress"}},
			},
		},
	}

	if err := validatePlanParams(params); err != nil {
		t.Errorf("validatePlanParams() error = %v", err)
	}
}

func TestValidatePlanParams_SinglePhaseError(t *testing.T) {
	params := PlanParams{
		Phases: []PhaseParam{
			{
				Name:   "Only Phase",
				Status: "pending",
				Tasks:  []TaskParam{{Content: "Task", ActiveForm: "Doing", Status: "pending"}},
			},
		},
	}

	err := validatePlanParams(params)
	if err == nil {
		t.Error("expected error for single phase")
	}
}

func TestValidatePlanParams_BothPhasesAndTasksError(t *testing.T) {
	params := PlanParams{
		Phases: []PhaseParam{
			{Name: "Phase 1", Status: "pending", Tasks: []TaskParam{{Content: "T1", ActiveForm: "D1", Status: "pending"}}},
			{Name: "Phase 2", Status: "pending", Tasks: []TaskParam{{Content: "T2", ActiveForm: "D2", Status: "pending"}}},
		},
		Tasks: []TaskParam{
			{Content: "Top level task", ActiveForm: "Doing", Status: "pending"},
		},
	}

	err := validatePlanParams(params)
	if err == nil {
		t.Error("expected error when both phases and tasks are provided")
	}
}

func TestValidatePlanParams_InvalidStatus(t *testing.T) {
	params := PlanParams{
		Tasks: []TaskParam{
			{Content: "Task", ActiveForm: "Doing", Status: "invalid_status"},
		},
	}

	err := validatePlanParams(params)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestValidatePlanParams_EmptyPhaseTasksError(t *testing.T) {
	params := PlanParams{
		Phases: []PhaseParam{
			{Name: "Phase 1", Status: "pending", Tasks: []TaskParam{{Content: "T", ActiveForm: "D", Status: "pending"}}},
			{Name: "Phase 2", Status: "pending", Tasks: []TaskParam{}}, // Empty tasks
		},
	}

	err := validatePlanParams(params)
	if err == nil {
		t.Error("expected error for phase with no tasks")
	}
}

func TestValidatePlanParams_TasksWithTodos(t *testing.T) {
	params := PlanParams{
		Tasks: []TaskParam{
			{
				Content:    "Main task",
				ActiveForm: "Doing main task",
				Status:     "in_progress",
				Todos: []TodoParam{
					{Content: "Sub item 1", ActiveForm: "Doing sub 1", Status: "completed"},
					{Content: "Sub item 2", ActiveForm: "Doing sub 2", Status: "in_progress"},
				},
			},
		},
	}

	if err := validatePlanParams(params); err != nil {
		t.Errorf("validatePlanParams() error = %v", err)
	}
}

func TestConvertParamsToPlan_FlatTasks(t *testing.T) {
	params := PlanParams{
		Tasks: []TaskParam{
			{Content: "Task 1", ActiveForm: "Doing 1", Status: "pending"},
			{Content: "Task 2", ActiveForm: "Doing 2", Status: "completed"},
		},
	}

	plan := convertParamsToPlan(params)

	if len(plan.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(plan.Tasks))
	}
	if plan.Tasks[0].Content != "Task 1" {
		t.Errorf("task[0].Content = %q, want 'Task 1'", plan.Tasks[0].Content)
	}
	if plan.Tasks[1].Status != session.PlanStatusCompleted {
		t.Errorf("task[1].Status = %q, want 'completed'", plan.Tasks[1].Status)
	}
}

func TestConvertParamsToPlan_WithPhases(t *testing.T) {
	params := PlanParams{
		Phases: []PhaseParam{
			{
				Name:   "Phase 1",
				Status: "completed",
				Tasks: []TaskParam{
					{Content: "Task A", ActiveForm: "Doing A", Status: "completed"},
				},
			},
			{
				Name:   "Phase 2",
				Status: "in_progress",
				Tasks: []TaskParam{
					{Content: "Task B", ActiveForm: "Doing B", Status: "in_progress"},
				},
			},
		},
	}

	plan := convertParamsToPlan(params)

	if len(plan.Phases) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(plan.Phases))
	}
	if plan.Phases[0].Name != "Phase 1" {
		t.Errorf("phase[0].Name = %q, want 'Phase 1'", plan.Phases[0].Name)
	}
	if len(plan.Phases[1].Tasks) != 1 {
		t.Errorf("phase[1] should have 1 task")
	}
}

func TestConvertParamsToPlan_TasksWithTodos(t *testing.T) {
	params := PlanParams{
		Tasks: []TaskParam{
			{
				Content:    "Main task",
				ActiveForm: "Doing main",
				Status:     "in_progress",
				Todos: []TodoParam{
					{Content: "Sub 1", ActiveForm: "Doing sub 1", Status: "completed"},
					{Content: "Sub 2", ActiveForm: "Doing sub 2", Status: "pending"},
				},
			},
		},
	}

	plan := convertParamsToPlan(params)

	if len(plan.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(plan.Tasks))
	}
	if len(plan.Tasks[0].Todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(plan.Tasks[0].Todos))
	}
	if plan.Tasks[0].Todos[0].Status != session.PlanStatusCompleted {
		t.Errorf("todo[0].Status = %q, want 'completed'", plan.Tasks[0].Todos[0].Status)
	}
}

func TestPlanResponseMetadata(t *testing.T) {
	metadata := PlanResponseMetadata{
		IsNew: true,
		Plan: session.Plan{
			Tasks: []session.Task{
				{Content: "Task 1", ActiveForm: "Doing task 1", Status: session.PlanStatusInProgress},
				{Content: "Task 2", ActiveForm: "Doing task 2", Status: session.PlanStatusPending},
			},
		},
		JustCompleted: []string{},
		JustStarted:   "Doing task 1",
		Completed:     0,
		Total:         2,
	}

	if !metadata.IsNew {
		t.Error("expected IsNew to be true")
	}
	if len(metadata.Plan.Tasks) != 2 {
		t.Errorf("expected 2 plan tasks, got %d", len(metadata.Plan.Tasks))
	}
	if metadata.JustStarted != "Doing task 1" {
		t.Errorf("expected JustStarted to be 'Doing task 1', got %q", metadata.JustStarted)
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()

	if id := GetSessionIDFromContext(ctx); id != "" {
		t.Errorf("expected empty session ID, got %q", id)
	}

	ctx = WithSessionID(ctx, "test-session-123")
	if id := GetSessionIDFromContext(ctx); id != "test-session-123" {
		t.Errorf("expected 'test-session-123', got %q", id)
	}

	if svc := GetServicesFromContext(ctx); svc != nil {
		t.Error("expected nil services")
	}
}
