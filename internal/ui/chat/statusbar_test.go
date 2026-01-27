package chat

import (
	"strings"
	"testing"
)

func TestNewStatusBar(t *testing.T) {
	sb := NewStatusBar()

	if sb == nil {
		t.Fatal("NewStatusBar returned nil")
	}

	if sb.hints == "" {
		t.Error("default hints should not be empty")
	}
}

func TestStatusBar_SetWidth(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(100)

	// Width is used in Render, so we just verify it doesn't panic
	_ = sb.Render()
}

func TestStatusBar_SetMemoryCount(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(100)
	sb.SetMemoryCount(5)

	rendered := sb.Render()
	if !strings.Contains(rendered, "5 memories") {
		t.Errorf("render should contain memory count, got: %s", rendered)
	}
}

func TestStatusBar_SetMemoryCount_Zero(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(100)
	sb.SetMemoryCount(0)

	rendered := sb.Render()
	if strings.Contains(rendered, "memories") {
		t.Error("render should not show memory count when zero")
	}
}

func TestStatusBar_SetTaskProgress(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(100)
	sb.SetTaskProgress("Running tests", 2, 5)

	rendered := sb.Render()
	if !strings.Contains(rendered, "2/5") {
		t.Errorf("render should contain task progress, got: %s", rendered)
	}
	if !strings.Contains(rendered, "Running tests") {
		t.Errorf("render should contain current task, got: %s", rendered)
	}
}

func TestStatusBar_SetTaskProgress_NoCurrentTask(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(100)
	sb.SetTaskProgress("", 3, 5)

	rendered := sb.Render()
	if !strings.Contains(rendered, "3/5") {
		t.Errorf("render should contain task progress, got: %s", rendered)
	}
}

func TestStatusBar_SetTaskProgress_ZeroTotal(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(100)
	sb.SetTaskProgress("", 0, 0)

	rendered := sb.Render()
	if strings.Contains(rendered, "/") {
		t.Error("render should not show progress when total is zero")
	}
}

func TestStatusBar_SetHints(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(100)
	sb.SetHints("custom hints")

	rendered := sb.Render()
	if !strings.Contains(rendered, "custom hints") {
		t.Errorf("render should contain custom hints, got: %s", rendered)
	}
}

func TestStatusBar_Render_NarrowWidth(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(30) // Very narrow
	sb.SetMemoryCount(10)
	sb.SetTaskProgress("Very long task name here", 5, 10)

	// Should not panic and should return something
	rendered := sb.Render()
	if rendered == "" {
		t.Error("render should return something even at narrow width")
	}
}

func TestStatusBar_Render_DefaultWidth(t *testing.T) {
	sb := NewStatusBar()
	// Don't set width - should use default of 80

	rendered := sb.Render()
	if rendered == "" {
		t.Error("render should return something with default width")
	}
}

func TestStatusBar_Update_MemoryCountMsg(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(100)

	sb.Update(MemoryCountMsg{Count: 7})

	if sb.memoryCount != 7 {
		t.Errorf("memoryCount = %d, want 7", sb.memoryCount)
	}
}

func TestStatusBar_Update_TaskProgressMsg(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(100)

	sb.Update(TaskProgressMsg{
		Current:   "Building",
		Completed: 3,
		Total:     8,
	})

	if sb.currentTask != "Building" {
		t.Errorf("currentTask = %q, want %q", sb.currentTask, "Building")
	}
	if sb.completedTasks != 3 {
		t.Errorf("completedTasks = %d, want 3", sb.completedTasks)
	}
	if sb.totalTasks != 8 {
		t.Errorf("totalTasks = %d, want 8", sb.totalTasks)
	}
}

func TestStatusBar_Update_HintsMsg(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(100)

	sb.Update(HintsMsg{Hints: "new hints"})

	if sb.hints != "new hints" {
		t.Errorf("hints = %q, want %q", sb.hints, "new hints")
	}
}

func TestStatusBar_truncateTask(t *testing.T) {
	sb := NewStatusBar()

	tests := []struct {
		name     string
		task     string
		maxLen   int
		expected string
	}{
		{
			name:     "short task",
			task:     "Build",
			maxLen:   10,
			expected: "Build",
		},
		{
			name:     "exact length",
			task:     "1234567890",
			maxLen:   10,
			expected: "1234567890",
		},
		{
			name:     "long task",
			task:     "Very long task name that needs truncation",
			maxLen:   20,
			expected: "Very long task na...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sb.truncateTask(tt.task, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateTask(%q, %d) = %q, want %q", tt.task, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestStatusBar_CombinedState(t *testing.T) {
	sb := NewStatusBar()
	sb.SetWidth(120)
	sb.SetMemoryCount(3)
	sb.SetTaskProgress("Testing", 2, 5)
	sb.SetHints("ctrl+c quit")

	rendered := sb.Render()

	// Should contain all elements
	if !strings.Contains(rendered, "3 memories") {
		t.Error("missing memory count")
	}
	if !strings.Contains(rendered, "2/5") {
		t.Error("missing task progress")
	}
	if !strings.Contains(rendered, "Testing") {
		t.Error("missing current task")
	}
	if !strings.Contains(rendered, "ctrl+c quit") {
		t.Error("missing hints")
	}
}
