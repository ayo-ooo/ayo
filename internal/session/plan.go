package session

import "encoding/json"

// PlanStatus represents the status of a plan item.
type PlanStatus string

const (
	PlanStatusPending    PlanStatus = "pending"
	PlanStatusInProgress PlanStatus = "in_progress"
	PlanStatusCompleted  PlanStatus = "completed"
)

// Plan represents a hierarchical task plan.
// Plans can be structured as:
// - Just tasks (flat list)
// - Phases containing tasks (must have 2+ phases)
// - Tasks containing todos (atomic sub-items)
type Plan struct {
	Phases []Phase `json:"phases,omitempty"` // Optional: high-level phases (requires 2+ if used)
	Tasks  []Task  `json:"tasks,omitempty"`  // Top-level tasks (when no phases)
}

// Phase represents a high-level grouping of tasks.
// If phases are used, there must be at least 2.
type Phase struct {
	Name   string     `json:"name"`   // Phase name (e.g., "Phase 1: Setup")
	Status PlanStatus `json:"status"` // Computed from child tasks
	Tasks  []Task     `json:"tasks"`  // Must have at least 1 task
}

// Task represents a unit of work within a plan or phase.
type Task struct {
	Content    string     `json:"content"`           // What needs to be done (imperative form)
	ActiveForm string     `json:"active_form"`       // Present continuous form (e.g., "Running tests")
	Status     PlanStatus `json:"status"`            // pending/in_progress/completed
	Todos      []Todo     `json:"todos,omitempty"`   // Optional atomic sub-items
}

// Todo represents an atomic item within a task.
type Todo struct {
	Content    string     `json:"content"`     // What needs to be done (imperative form)
	ActiveForm string     `json:"active_form"` // Present continuous form
	Status     PlanStatus `json:"status"`      // pending/in_progress/completed
}

// PlanItem is kept for backward compatibility with existing code.
// It represents a flat task (equivalent to Task without todos).
type PlanItem = Task

// IsEmpty returns true if the plan has no content.
func (p Plan) IsEmpty() bool {
	return len(p.Phases) == 0 && len(p.Tasks) == 0
}

// IsFlat returns true if the plan has no phases (just tasks).
func (p Plan) IsFlat() bool {
	return len(p.Phases) == 0
}

// AllTasks returns all tasks in the plan, flattened from phases if needed.
func (p Plan) AllTasks() []Task {
	if p.IsFlat() {
		return p.Tasks
	}
	var tasks []Task
	for _, phase := range p.Phases {
		tasks = append(tasks, phase.Tasks...)
	}
	return tasks
}

// Stats returns counts of pending, in-progress, and completed items.
func (p Plan) Stats() (pending, inProgress, completed int) {
	for _, task := range p.AllTasks() {
		if len(task.Todos) > 0 {
			// Count todos instead of task
			for _, todo := range task.Todos {
				switch todo.Status {
				case PlanStatusPending:
					pending++
				case PlanStatusInProgress:
					inProgress++
				case PlanStatusCompleted:
					completed++
				}
			}
		} else {
			// Count the task itself
			switch task.Status {
			case PlanStatusPending:
				pending++
			case PlanStatusInProgress:
				inProgress++
			case PlanStatusCompleted:
				completed++
			}
		}
	}
	return
}

// CurrentActivity returns the active_form of the current in-progress item.
func (p Plan) CurrentActivity() string {
	for _, task := range p.AllTasks() {
		if len(task.Todos) > 0 {
			for _, todo := range task.Todos {
				if todo.Status == PlanStatusInProgress {
					if todo.ActiveForm != "" {
						return todo.ActiveForm
					}
					return todo.Content
				}
			}
		}
		if task.Status == PlanStatusInProgress {
			if task.ActiveForm != "" {
				return task.ActiveForm
			}
			return task.Content
		}
	}
	return ""
}

// marshalPlan converts a plan to JSON string for storage.
func marshalPlan(plan Plan) (string, error) {
	if plan.IsEmpty() {
		return "", nil
	}
	data, err := json.Marshal(plan)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// unmarshalPlan converts a JSON string to a plan.
func unmarshalPlan(data string) (Plan, error) {
	if data == "" {
		return Plan{}, nil
	}

	// First try to unmarshal as the new Plan struct
	var plan Plan
	if err := json.Unmarshal([]byte(data), &plan); err == nil {
		// Check if it's actually a new-style plan
		if len(plan.Phases) > 0 || len(plan.Tasks) > 0 {
			return plan, nil
		}
	}

	// Fall back to old format: array of PlanItem/Task
	var oldPlan []Task
	if err := json.Unmarshal([]byte(data), &oldPlan); err != nil {
		return Plan{}, err
	}
	return Plan{Tasks: oldPlan}, nil
}
