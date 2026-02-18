package todos

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// StateFile is the name of the state file within the planner's state directory.
const StateFile = "todos.json"

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

// Todo represents a single todo item.
type Todo struct {
	// ID is a unique identifier for this todo.
	ID string `json:"id"`

	// Content describes what needs to be done (imperative form).
	// Example: "Run tests", "Fix the bug", "Update documentation"
	Content string `json:"content"`

	// ActiveForm is the present continuous form shown during execution.
	// Example: "Running tests", "Fixing the bug", "Updating documentation"
	ActiveForm string `json:"active_form,omitempty"`

	// Status is the current status of this todo.
	Status TodoStatus `json:"status"`
}

// State holds the persistent state for the ayo-todos plugin.
type State struct {
	mu    sync.RWMutex
	Todos []Todo `json:"todos"`
}

// NewState creates a new empty State.
func NewState() *State {
	return &State{
		Todos: make([]Todo, 0),
	}
}

// Load reads state from a JSON file.
// If the file doesn't exist, returns an empty state.
func LoadState(statePath string) (*State, error) {
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewState(), nil
		}
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	// Initialize if nil
	if state.Todos == nil {
		state.Todos = make([]Todo, 0)
	}

	return &state, nil
}

// Save writes the state to a JSON file.
func (s *State) Save(statePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0o644)
}

// List returns all todos.
func (s *State) List() []Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Todo, len(s.Todos))
	copy(result, s.Todos)
	return result
}

// Set replaces all todos with the given list.
// This is used by the todos tool which provides the complete list.
func (s *State) Set(todos []Todo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Todos = make([]Todo, len(todos))
	copy(s.Todos, todos)
}

// Get returns a todo by ID.
// Returns nil if not found.
func (s *State) Get(id string) *Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.Todos {
		if s.Todos[i].ID == id {
			// Return a copy
			t := s.Todos[i]
			return &t
		}
	}
	return nil
}

// CountByStatus returns counts of todos by status.
func (s *State) CountByStatus() map[TodoStatus]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := make(map[TodoStatus]int)
	for _, t := range s.Todos {
		counts[t.Status]++
	}
	return counts
}

// IsEmpty returns true if there are no todos.
func (s *State) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Todos) == 0
}

// Count returns the total number of todos.
func (s *State) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Todos)
}
