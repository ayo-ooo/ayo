package goals

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StateFile is the name of the state file within the planner's state directory.
const StateFile = "goals.json"

// GoalStatus represents the status of a goal.
type GoalStatus string

const (
	// StatusActive indicates the goal is being worked on.
	StatusActive GoalStatus = "active"
	// StatusAchieved indicates the goal was accomplished.
	StatusAchieved GoalStatus = "achieved"
	// StatusAbandoned indicates the goal was dropped.
	StatusAbandoned GoalStatus = "abandoned"
)

// IsValid returns true if the status is a known value.
func (s GoalStatus) IsValid() bool {
	return s == StatusActive || s == StatusAchieved || s == StatusAbandoned
}

// Milestone represents a checkpoint toward a goal.
type Milestone struct {
	// Description of this milestone.
	Description string `json:"description"`

	// Completed indicates if this milestone has been reached.
	Completed bool `json:"completed"`
}

// Goal represents a session goal.
type Goal struct {
	// ID is a unique identifier for this goal.
	ID string `json:"id"`

	// Goal describes the desired outcome.
	Goal string `json:"goal"`

	// Status is the current status of this goal.
	Status GoalStatus `json:"status"`

	// Progress is a percentage estimate (0-100).
	Progress int `json:"progress"`

	// Milestones are checkpoints toward the goal.
	Milestones []Milestone `json:"milestones,omitempty"`

	// Notes contain context, learnings, or blockers.
	Notes []string `json:"notes,omitempty"`

	// CreatedAt is when the goal was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the goal was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// State holds the persistent state for the ayo-goals plugin.
type State struct {
	mu    sync.RWMutex
	Goals []Goal `json:"goals"`
}

// NewState creates a new empty State.
func NewState() *State {
	return &State{
		Goals: make([]Goal, 0),
	}
}

// LoadState reads state from a JSON file.
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

	if state.Goals == nil {
		state.Goals = make([]Goal, 0)
	}

	return &state, nil
}

// Save writes the state to a JSON file.
func (s *State) Save(statePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0o644)
}

// List returns all goals.
func (s *State) List() []Goal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Goal, len(s.Goals))
	copy(result, s.Goals)
	return result
}

// Set replaces all goals with the given list.
func (s *State) Set(goals []Goal) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Goals = make([]Goal, len(goals))
	copy(s.Goals, goals)
}

// ActiveGoal returns the currently active goal, if any.
func (s *State) ActiveGoal() *Goal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.Goals {
		if s.Goals[i].Status == StatusActive {
			g := s.Goals[i]
			return &g
		}
	}
	return nil
}

// CountByStatus returns counts of goals by status.
func (s *State) CountByStatus() map[GoalStatus]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := make(map[GoalStatus]int)
	for _, g := range s.Goals {
		counts[g.Status]++
	}
	return counts
}

// IsEmpty returns true if there are no goals.
func (s *State) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Goals) == 0
}

// Count returns the total number of goals.
func (s *State) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Goals)
}
