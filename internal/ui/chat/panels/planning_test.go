package panels

import (
	"strings"
	"testing"
)

func TestNewPlanningPanel(t *testing.T) {
	p := NewPlanningPanel()
	if p == nil {
		t.Fatal("Expected non-nil panel")
	}
	if p.IsVisible() {
		t.Error("Expected panel to be hidden initially")
	}
}

func TestPlanningPanel_Toggle(t *testing.T) {
	p := NewPlanningPanel()

	p.Toggle()
	if !p.IsVisible() {
		t.Error("Expected visible after first toggle")
	}

	p.Toggle()
	if p.IsVisible() {
		t.Error("Expected hidden after second toggle")
	}
}

func TestPlanningPanel_ShowHide(t *testing.T) {
	p := NewPlanningPanel()

	p.Show()
	if !p.IsVisible() {
		t.Error("Expected visible after Show()")
	}

	p.Hide()
	if p.IsVisible() {
		t.Error("Expected hidden after Hide()")
	}
}

func TestPlanningPanel_Focus(t *testing.T) {
	p := NewPlanningPanel()

	p.Focus()
	if !p.IsFocused() {
		t.Error("Expected focused after Focus()")
	}

	p.Blur()
	if p.IsFocused() {
		t.Error("Expected not focused after Blur()")
	}
}

func TestPlanningPanel_SetTodos(t *testing.T) {
	p := NewPlanningPanel()
	p.SetSize(40, 20)

	todos := []TodoItem{
		{Content: "Task 1", Status: "pending"},
		{Content: "Task 2", Status: "in_progress", ActiveForm: "Working on Task 2"},
		{Content: "Task 3", Status: "completed"},
	}

	p.SetTodos(todos)

	// Panel should have content
	p.Show()
	view := p.View()
	if view == "" {
		t.Error("Expected non-empty view with todos")
	}
}

func TestPlanningPanel_SetSize(t *testing.T) {
	p := NewPlanningPanel()
	p.SetSize(80, 24)

	// Should initialize viewport
	p.Show()
	view := p.View()
	if view == "" {
		t.Error("Expected non-empty view after SetSize")
	}
}

func TestPlanningPanel_View_Hidden(t *testing.T) {
	p := NewPlanningPanel()
	p.SetSize(80, 24)

	// Hidden by default
	if p.View() != "" {
		t.Error("Expected empty view when hidden")
	}
}

func TestPlanningPanel_View_NoSize(t *testing.T) {
	p := NewPlanningPanel()
	p.Show()

	// No size set
	if p.View() != "" {
		t.Error("Expected empty view with no size")
	}
}

func TestPlanningPanel_View_EmptyTodos(t *testing.T) {
	p := NewPlanningPanel()
	p.SetSize(40, 20)
	p.Show()

	view := p.View()
	if !strings.Contains(view, "No tasks") {
		t.Error("Expected 'No tasks' message in empty panel")
	}
}

func TestPlanningPanel_View_WithTodos(t *testing.T) {
	p := NewPlanningPanel()
	p.SetSize(60, 20)
	p.SetTodos([]TodoItem{
		{Content: "First task", Status: "completed"},
		{Content: "Second task", Status: "in_progress", ActiveForm: "Doing second"},
		{Content: "Third task", Status: "pending"},
	})
	p.Show()

	view := p.View()

	// Should contain Planning title
	if !strings.Contains(view, "Planning") {
		t.Error("Expected 'Planning' title in view")
	}

	// Should contain status indicator (completed/total)
	if !strings.Contains(view, "1/3") {
		t.Error("Expected '1/3' progress indicator")
	}
}

func TestPlanningPanel_renderTodoItem(t *testing.T) {
	p := NewPlanningPanel()
	p.SetSize(80, 20)

	tests := []struct {
		name     string
		todo     TodoItem
		contains string
	}{
		{
			name:     "completed task",
			todo:     TodoItem{Content: "Done task", Status: "completed"},
			contains: "Done task",
		},
		{
			name:     "in_progress uses ActiveForm",
			todo:     TodoItem{Content: "Task", Status: "in_progress", ActiveForm: "Working"},
			contains: "Working",
		},
		{
			name:     "pending task",
			todo:     TodoItem{Content: "Pending task", Status: "pending"},
			contains: "Pending task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.renderTodoItem(tt.todo)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestPlanningPanel_renderStats(t *testing.T) {
	p := NewPlanningPanel()

	// Empty todos
	if p.renderStats() != "" {
		t.Error("Expected empty stats for no todos")
	}

	// With todos
	p.SetTodos([]TodoItem{
		{Content: "A", Status: "completed"},
		{Content: "B", Status: "completed"},
		{Content: "C", Status: "pending"},
	})

	stats := p.renderStats()
	if !strings.Contains(stats, "2/3") {
		t.Errorf("Expected '2/3' in stats, got %q", stats)
	}
}

func TestPlanningPanel_Scroll(t *testing.T) {
	p := NewPlanningPanel()
	p.SetSize(40, 10)

	// Add many todos to enable scrolling
	todos := make([]TodoItem, 20)
	for i := range todos {
		todos[i] = TodoItem{Content: "Task " + string(rune('A'+i)), Status: "pending"}
	}
	p.SetTodos(todos)
	p.Show()

	// These should not panic
	p.ScrollDown(5)
	p.ScrollUp(2)
}
