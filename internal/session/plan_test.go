package session

import (
	"testing"
)

func TestMarshalUnmarshalPlan_FlatTasks(t *testing.T) {
	plan := Plan{
		Tasks: []Task{
			{Content: "Run tests", ActiveForm: "Running tests", Status: PlanStatusPending},
			{Content: "Fix bugs", ActiveForm: "Fixing bugs", Status: PlanStatusCompleted},
		},
	}

	data, err := marshalPlan(plan)
	if err != nil {
		t.Fatalf("marshalPlan() error = %v", err)
	}

	got, err := unmarshalPlan(data)
	if err != nil {
		t.Fatalf("unmarshalPlan() error = %v", err)
	}

	if len(got.Tasks) != len(plan.Tasks) {
		t.Fatalf("tasks length mismatch: got %d, want %d", len(got.Tasks), len(plan.Tasks))
	}

	for i, task := range got.Tasks {
		if task.Content != plan.Tasks[i].Content {
			t.Errorf("task[%d].Content = %q, want %q", i, task.Content, plan.Tasks[i].Content)
		}
		if task.Status != plan.Tasks[i].Status {
			t.Errorf("task[%d].Status = %q, want %q", i, task.Status, plan.Tasks[i].Status)
		}
	}
}

func TestMarshalUnmarshalPlan_WithPhases(t *testing.T) {
	plan := Plan{
		Phases: []Phase{
			{
				Name:   "Phase 1: Setup",
				Status: PlanStatusCompleted,
				Tasks: []Task{
					{Content: "Initialize project", ActiveForm: "Initializing", Status: PlanStatusCompleted},
				},
			},
			{
				Name:   "Phase 2: Implementation",
				Status: PlanStatusInProgress,
				Tasks: []Task{
					{Content: "Build features", ActiveForm: "Building", Status: PlanStatusInProgress},
					{Content: "Write tests", ActiveForm: "Writing tests", Status: PlanStatusPending},
				},
			},
		},
	}

	data, err := marshalPlan(plan)
	if err != nil {
		t.Fatalf("marshalPlan() error = %v", err)
	}

	got, err := unmarshalPlan(data)
	if err != nil {
		t.Fatalf("unmarshalPlan() error = %v", err)
	}

	if len(got.Phases) != len(plan.Phases) {
		t.Fatalf("phases length mismatch: got %d, want %d", len(got.Phases), len(plan.Phases))
	}

	for i, phase := range got.Phases {
		if phase.Name != plan.Phases[i].Name {
			t.Errorf("phase[%d].Name = %q, want %q", i, phase.Name, plan.Phases[i].Name)
		}
		if len(phase.Tasks) != len(plan.Phases[i].Tasks) {
			t.Errorf("phase[%d].Tasks length = %d, want %d", i, len(phase.Tasks), len(plan.Phases[i].Tasks))
		}
	}
}

func TestMarshalUnmarshalPlan_TasksWithTodos(t *testing.T) {
	plan := Plan{
		Tasks: []Task{
			{
				Content:    "Implement auth",
				ActiveForm: "Implementing auth",
				Status:     PlanStatusInProgress,
				Todos: []Todo{
					{Content: "Create model", ActiveForm: "Creating model", Status: PlanStatusCompleted},
					{Content: "Add endpoints", ActiveForm: "Adding endpoints", Status: PlanStatusInProgress},
				},
			},
		},
	}

	data, err := marshalPlan(plan)
	if err != nil {
		t.Fatalf("marshalPlan() error = %v", err)
	}

	got, err := unmarshalPlan(data)
	if err != nil {
		t.Fatalf("unmarshalPlan() error = %v", err)
	}

	if len(got.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(got.Tasks))
	}

	if len(got.Tasks[0].Todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(got.Tasks[0].Todos))
	}

	if got.Tasks[0].Todos[0].Content != "Create model" {
		t.Errorf("todo[0].Content = %q, want 'Create model'", got.Tasks[0].Todos[0].Content)
	}
}

func TestUnmarshalPlan_BackwardCompatibility(t *testing.T) {
	// Old format: array of tasks
	oldFormat := `[{"content":"Task 1","active_form":"Doing task 1","status":"pending"},{"content":"Task 2","active_form":"Doing task 2","status":"completed"}]`

	plan, err := unmarshalPlan(oldFormat)
	if err != nil {
		t.Fatalf("unmarshalPlan() error = %v", err)
	}

	if len(plan.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(plan.Tasks))
	}

	if plan.Tasks[0].Content != "Task 1" {
		t.Errorf("task[0].Content = %q, want 'Task 1'", plan.Tasks[0].Content)
	}
}

func TestUnmarshalPlanEmptyString(t *testing.T) {
	plan, err := unmarshalPlan("")
	if err != nil {
		t.Fatalf("unmarshalPlan() error = %v", err)
	}
	if !plan.IsEmpty() {
		t.Error("unmarshalPlan(\"\") should return empty plan")
	}
}

func TestUnmarshalPlanInvalidJSON(t *testing.T) {
	_, err := unmarshalPlan("not valid json")
	if err == nil {
		t.Error("unmarshalPlan() should return error for invalid JSON")
	}
}

func TestPlanStatusConstants(t *testing.T) {
	if PlanStatusPending != "pending" {
		t.Errorf("PlanStatusPending = %q, want %q", PlanStatusPending, "pending")
	}
	if PlanStatusInProgress != "in_progress" {
		t.Errorf("PlanStatusInProgress = %q, want %q", PlanStatusInProgress, "in_progress")
	}
	if PlanStatusCompleted != "completed" {
		t.Errorf("PlanStatusCompleted = %q, want %q", PlanStatusCompleted, "completed")
	}
}

func TestPlan_IsEmpty(t *testing.T) {
	empty := Plan{}
	if !empty.IsEmpty() {
		t.Error("empty plan should be empty")
	}

	withTasks := Plan{Tasks: []Task{{Content: "test"}}}
	if withTasks.IsEmpty() {
		t.Error("plan with tasks should not be empty")
	}

	withPhases := Plan{Phases: []Phase{{Name: "test"}}}
	if withPhases.IsEmpty() {
		t.Error("plan with phases should not be empty")
	}
}

func TestPlan_IsFlat(t *testing.T) {
	flat := Plan{Tasks: []Task{{Content: "test"}}}
	if !flat.IsFlat() {
		t.Error("plan with only tasks should be flat")
	}

	withPhases := Plan{Phases: []Phase{{Name: "test"}}}
	if withPhases.IsFlat() {
		t.Error("plan with phases should not be flat")
	}
}

func TestPlan_AllTasks(t *testing.T) {
	// Flat plan
	flat := Plan{
		Tasks: []Task{
			{Content: "Task 1"},
			{Content: "Task 2"},
		},
	}
	allTasks := flat.AllTasks()
	if len(allTasks) != 2 {
		t.Errorf("flat plan AllTasks() = %d, want 2", len(allTasks))
	}

	// Plan with phases
	phased := Plan{
		Phases: []Phase{
			{
				Name:  "Phase 1",
				Tasks: []Task{{Content: "Task A"}, {Content: "Task B"}},
			},
			{
				Name:  "Phase 2",
				Tasks: []Task{{Content: "Task C"}},
			},
		},
	}
	allTasks = phased.AllTasks()
	if len(allTasks) != 3 {
		t.Errorf("phased plan AllTasks() = %d, want 3", len(allTasks))
	}
}

func TestPlan_Stats(t *testing.T) {
	plan := Plan{
		Tasks: []Task{
			{Content: "Completed", Status: PlanStatusCompleted},
			{Content: "In Progress", Status: PlanStatusInProgress},
			{Content: "Pending 1", Status: PlanStatusPending},
			{Content: "Pending 2", Status: PlanStatusPending},
		},
	}

	pending, inProgress, completed := plan.Stats()
	if pending != 2 {
		t.Errorf("pending = %d, want 2", pending)
	}
	if inProgress != 1 {
		t.Errorf("inProgress = %d, want 1", inProgress)
	}
	if completed != 1 {
		t.Errorf("completed = %d, want 1", completed)
	}
}

func TestPlan_Stats_WithTodos(t *testing.T) {
	plan := Plan{
		Tasks: []Task{
			{
				Content: "Task with todos",
				Status:  PlanStatusInProgress,
				Todos: []Todo{
					{Content: "Todo 1", Status: PlanStatusCompleted},
					{Content: "Todo 2", Status: PlanStatusInProgress},
					{Content: "Todo 3", Status: PlanStatusPending},
				},
			},
		},
	}

	pending, inProgress, completed := plan.Stats()
	// When task has todos, we count todos not the task
	if pending != 1 {
		t.Errorf("pending = %d, want 1", pending)
	}
	if inProgress != 1 {
		t.Errorf("inProgress = %d, want 1", inProgress)
	}
	if completed != 1 {
		t.Errorf("completed = %d, want 1", completed)
	}
}

func TestPlan_CurrentActivity(t *testing.T) {
	plan := Plan{
		Tasks: []Task{
			{Content: "Completed", Status: PlanStatusCompleted},
			{Content: "Active Task", ActiveForm: "Doing active task", Status: PlanStatusInProgress},
		},
	}

	activity := plan.CurrentActivity()
	if activity != "Doing active task" {
		t.Errorf("CurrentActivity() = %q, want 'Doing active task'", activity)
	}
}

func TestPlan_CurrentActivity_InTodo(t *testing.T) {
	plan := Plan{
		Tasks: []Task{
			{
				Content: "Task",
				Status:  PlanStatusInProgress,
				Todos: []Todo{
					{Content: "Done todo", Status: PlanStatusCompleted},
					{Content: "Active todo", ActiveForm: "Doing todo", Status: PlanStatusInProgress},
				},
			},
		},
	}

	activity := plan.CurrentActivity()
	if activity != "Doing todo" {
		t.Errorf("CurrentActivity() = %q, want 'Doing todo'", activity)
	}
}
