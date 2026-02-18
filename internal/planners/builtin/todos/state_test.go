package todos

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTodoStatus_IsValid(t *testing.T) {
	tests := []struct {
		status TodoStatus
		valid  bool
	}{
		{StatusPending, true},
		{StatusInProgress, true},
		{StatusCompleted, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("TodoStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestNewState(t *testing.T) {
	state := NewState()
	if state == nil {
		t.Fatal("NewState() returned nil")
	}
	if state.Todos == nil {
		t.Error("Todos slice should not be nil")
	}
	if len(state.Todos) != 0 {
		t.Errorf("expected 0 todos, got %d", len(state.Todos))
	}
}

func TestLoadState_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "nonexistent.json")

	state, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState() failed: %v", err)
	}
	if state == nil {
		t.Fatal("LoadState() returned nil state")
	}
	if len(state.Todos) != 0 {
		t.Errorf("expected empty state, got %d todos", len(state.Todos))
	}
}

func TestState_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	// Create state with todos
	state := NewState()
	state.Set([]Todo{
		{ID: "1", Content: "First task", Status: StatusPending},
		{ID: "2", Content: "Second task", ActiveForm: "Doing second task", Status: StatusInProgress},
		{ID: "3", Content: "Done task", Status: StatusCompleted},
	})

	// Save
	if err := state.Save(statePath); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("state file not created: %v", err)
	}

	// Load
	loaded, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState() failed: %v", err)
	}

	// Verify
	if len(loaded.Todos) != 3 {
		t.Fatalf("expected 3 todos, got %d", len(loaded.Todos))
	}

	if loaded.Todos[0].ID != "1" || loaded.Todos[0].Content != "First task" {
		t.Errorf("todo 0 mismatch: %+v", loaded.Todos[0])
	}
	if loaded.Todos[1].ActiveForm != "Doing second task" {
		t.Errorf("todo 1 active_form mismatch: %+v", loaded.Todos[1])
	}
	if loaded.Todos[2].Status != StatusCompleted {
		t.Errorf("todo 2 status mismatch: %+v", loaded.Todos[2])
	}
}

func TestState_List(t *testing.T) {
	state := NewState()
	state.Set([]Todo{
		{ID: "1", Content: "Task 1", Status: StatusPending},
		{ID: "2", Content: "Task 2", Status: StatusPending},
	})

	todos := state.List()
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}

	// Modify returned slice shouldn't affect state
	todos[0].Content = "Modified"
	if state.Todos[0].Content == "Modified" {
		t.Error("List() should return a copy")
	}
}

func TestState_Get(t *testing.T) {
	state := NewState()
	state.Set([]Todo{
		{ID: "abc", Content: "Target task", Status: StatusPending},
		{ID: "def", Content: "Other task", Status: StatusPending},
	})

	// Find existing
	todo := state.Get("abc")
	if todo == nil {
		t.Fatal("Get() returned nil for existing todo")
	}
	if todo.Content != "Target task" {
		t.Errorf("Content = %q, want %q", todo.Content, "Target task")
	}

	// Modify returned shouldn't affect state
	todo.Content = "Modified"
	if state.Todos[0].Content == "Modified" {
		t.Error("Get() should return a copy")
	}

	// Find non-existing
	todo = state.Get("nonexistent")
	if todo != nil {
		t.Error("Get() should return nil for non-existing todo")
	}
}

func TestState_CountByStatus(t *testing.T) {
	state := NewState()
	state.Set([]Todo{
		{ID: "1", Status: StatusPending},
		{ID: "2", Status: StatusPending},
		{ID: "3", Status: StatusInProgress},
		{ID: "4", Status: StatusCompleted},
		{ID: "5", Status: StatusCompleted},
		{ID: "6", Status: StatusCompleted},
	})

	counts := state.CountByStatus()
	if counts[StatusPending] != 2 {
		t.Errorf("pending count = %d, want 2", counts[StatusPending])
	}
	if counts[StatusInProgress] != 1 {
		t.Errorf("in_progress count = %d, want 1", counts[StatusInProgress])
	}
	if counts[StatusCompleted] != 3 {
		t.Errorf("completed count = %d, want 3", counts[StatusCompleted])
	}
}

func TestState_IsEmpty(t *testing.T) {
	state := NewState()
	if !state.IsEmpty() {
		t.Error("new state should be empty")
	}

	state.Set([]Todo{{ID: "1", Content: "Task", Status: StatusPending}})
	if state.IsEmpty() {
		t.Error("state with todos should not be empty")
	}
}

func TestState_Count(t *testing.T) {
	state := NewState()
	if state.Count() != 0 {
		t.Errorf("Count() = %d, want 0", state.Count())
	}

	state.Set([]Todo{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	})
	if state.Count() != 3 {
		t.Errorf("Count() = %d, want 3", state.Count())
	}
}

func TestLoadState_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(statePath, []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := LoadState(statePath)
	if err == nil {
		t.Error("LoadState() should fail for invalid JSON")
	}
}

func TestState_Save_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "nested", "deep", "state.json")

	state := NewState()
	if err := state.Save(statePath); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(statePath); err != nil {
		t.Errorf("state file not created: %v", err)
	}
}
