package panels

import (
	"strings"
	"testing"
)

func TestNewMemoryPanel(t *testing.T) {
	p := NewMemoryPanel()
	if p == nil {
		t.Fatal("Expected non-nil panel")
	}
	if p.IsVisible() {
		t.Error("Expected panel to be hidden initially")
	}
}

func TestMemoryPanel_Toggle(t *testing.T) {
	p := NewMemoryPanel()

	p.Toggle()
	if !p.IsVisible() {
		t.Error("Expected visible after first toggle")
	}

	p.Toggle()
	if p.IsVisible() {
		t.Error("Expected hidden after second toggle")
	}
}

func TestMemoryPanel_ShowHide(t *testing.T) {
	p := NewMemoryPanel()

	p.Show()
	if !p.IsVisible() {
		t.Error("Expected visible after Show()")
	}

	p.Hide()
	if p.IsVisible() {
		t.Error("Expected hidden after Hide()")
	}
}

func TestMemoryPanel_Focus(t *testing.T) {
	p := NewMemoryPanel()

	p.Focus()
	if !p.IsFocused() {
		t.Error("Expected focused after Focus()")
	}

	p.Blur()
	if p.IsFocused() {
		t.Error("Expected not focused after Blur()")
	}
}

func TestMemoryPanel_SetMemories(t *testing.T) {
	p := NewMemoryPanel()
	p.SetSize(40, 20)

	memories := []MemoryItem{
		{ID: "1", Content: "User prefers dark mode", Category: "preference"},
		{ID: "2", Content: "Project uses Go", Category: "fact"},
		{ID: "3", Content: "Don't use tabs", Category: "correction"},
	}

	p.SetMemories(memories)

	p.Show()
	view := p.View()
	if view == "" {
		t.Error("Expected non-empty view with memories")
	}
}

func TestMemoryPanel_SetSize(t *testing.T) {
	p := NewMemoryPanel()
	p.SetSize(80, 24)

	p.Show()
	view := p.View()
	if view == "" {
		t.Error("Expected non-empty view after SetSize")
	}
}

func TestMemoryPanel_View_Hidden(t *testing.T) {
	p := NewMemoryPanel()
	p.SetSize(80, 24)

	// Hidden by default
	if p.View() != "" {
		t.Error("Expected empty view when hidden")
	}
}

func TestMemoryPanel_View_NoSize(t *testing.T) {
	p := NewMemoryPanel()
	p.Show()

	// No size set
	if p.View() != "" {
		t.Error("Expected empty view with no size")
	}
}

func TestMemoryPanel_View_EmptyMemories(t *testing.T) {
	p := NewMemoryPanel()
	p.SetSize(40, 20)
	p.Show()

	view := p.View()
	if !strings.Contains(view, "No relevant memories") {
		t.Error("Expected 'No relevant memories' message in empty panel")
	}
}

func TestMemoryPanel_View_WithMemories(t *testing.T) {
	p := NewMemoryPanel()
	p.SetSize(60, 20)
	p.SetMemories([]MemoryItem{
		{ID: "1", Content: "First memory", Category: "preference"},
		{ID: "2", Content: "Second memory", Category: "fact"},
	})
	p.Show()

	view := p.View()

	// Should contain Memory title
	if !strings.Contains(view, "Memory") {
		t.Error("Expected 'Memory' title in view")
	}

	// Should contain count
	if !strings.Contains(view, "2") {
		t.Error("Expected '2' count indicator")
	}
}

func TestMemoryPanel_renderMemoryItem(t *testing.T) {
	p := NewMemoryPanel()
	p.SetSize(80, 20)

	tests := []struct {
		name     string
		memory   MemoryItem
		contains string
	}{
		{
			name:     "preference",
			memory:   MemoryItem{Content: "Dark mode", Category: "preference"},
			contains: "Dark mode",
		},
		{
			name:     "fact",
			memory:   MemoryItem{Content: "Uses Go", Category: "fact"},
			contains: "Uses Go",
		},
		{
			name:     "correction",
			memory:   MemoryItem{Content: "No tabs", Category: "correction"},
			contains: "No tabs",
		},
		{
			name:     "pattern",
			memory:   MemoryItem{Content: "Pattern found", Category: "pattern"},
			contains: "Pattern found",
		},
		{
			name:     "with scope",
			memory:   MemoryItem{Content: "Scoped memory", Category: "fact", Scope: "@ayo"},
			contains: "@ayo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.renderMemoryItem(tt.memory)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestMemoryPanel_renderMemoryItem_GlobalScope(t *testing.T) {
	p := NewMemoryPanel()
	p.SetSize(80, 20)

	// Global scope should not show scope badge
	mem := MemoryItem{Content: "Global memory", Category: "fact", Scope: "global"}
	result := p.renderMemoryItem(mem)

	if strings.Contains(result, "[global]") {
		t.Error("Expected global scope to not show badge")
	}
}

func TestMemoryPanel_Scroll(t *testing.T) {
	p := NewMemoryPanel()
	p.SetSize(40, 10)

	// Add many memories to enable scrolling
	memories := make([]MemoryItem, 20)
	for i := range memories {
		memories[i] = MemoryItem{
			ID:       string(rune('a' + i)),
			Content:  "Memory " + string(rune('A'+i)),
			Category: "fact",
		}
	}
	p.SetMemories(memories)
	p.Show()

	// These should not panic
	p.ScrollDown(5)
	p.ScrollUp(2)
}
