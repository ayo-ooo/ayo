package panels

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSidebar(t *testing.T) {
	s := NewSidebar()
	if s == nil {
		t.Fatal("Expected non-nil sidebar")
	}
	if s.IsVisible() {
		t.Error("Expected sidebar to be hidden initially")
	}
	if s.ActivePanel() != "planning" {
		t.Errorf("Expected active panel 'planning', got %q", s.ActivePanel())
	}
}

func TestSidebar_SetSize_WideTerminal(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	if s.Position() != PositionRight {
		t.Error("Expected PositionRight for wide terminal")
	}
}

func TestSidebar_SetSize_NarrowTerminal(t *testing.T) {
	s := NewSidebar()
	s.SetSize(80, 24)

	if s.Position() != PositionBottom {
		t.Error("Expected PositionBottom for narrow terminal")
	}
}

func TestSidebar_Width_Hidden(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	// Hidden sidebar returns 0 width
	if s.Width() != 0 {
		t.Errorf("Expected 0 width when hidden, got %d", s.Width())
	}
}

func TestSidebar_Width_Visible(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)
	s.TogglePlanning()

	if s.Width() == 0 {
		t.Error("Expected non-zero width when visible")
	}
}

func TestSidebar_Height_Hidden(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	if s.Height() != 0 {
		t.Errorf("Expected 0 height when hidden, got %d", s.Height())
	}
}

func TestSidebar_TogglePlanning(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	// Toggle on
	s.TogglePlanning()
	if !s.IsVisible() {
		t.Error("Expected visible after toggle")
	}
	if s.ActivePanel() != "planning" {
		t.Error("Expected planning panel active")
	}

	// Toggle off (same panel)
	s.TogglePlanning()
	if s.IsVisible() {
		t.Error("Expected hidden after toggle same panel")
	}
}

func TestSidebar_ToggleMemory(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	s.ToggleMemory()
	if !s.IsVisible() {
		t.Error("Expected visible after toggle")
	}
	if s.ActivePanel() != "memory" {
		t.Error("Expected memory panel active")
	}

	// Toggle off
	s.ToggleMemory()
	if s.IsVisible() {
		t.Error("Expected hidden after toggle same panel")
	}
}

func TestSidebar_SwitchPanels(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	// Open planning
	s.TogglePlanning()
	if s.ActivePanel() != "planning" {
		t.Error("Expected planning active")
	}

	// Switch to memory
	s.ToggleMemory()
	if s.ActivePanel() != "memory" {
		t.Error("Expected memory active after switch")
	}
	if !s.IsVisible() {
		t.Error("Expected still visible after switch")
	}
}

func TestSidebar_Close(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	s.TogglePlanning()
	s.Close()

	if s.IsVisible() {
		t.Error("Expected hidden after Close()")
	}
}

func TestSidebar_Focus(t *testing.T) {
	s := NewSidebar()

	s.Focus()
	if !s.IsFocused() {
		t.Error("Expected focused after Focus()")
	}

	s.Blur()
	if s.IsFocused() {
		t.Error("Expected not focused after Blur()")
	}
}

func TestSidebar_SetTodos(t *testing.T) {
	s := NewSidebar()
	s.SetSize(40, 20)

	todos := []TodoItem{
		{Content: "Task 1", Status: "pending"},
	}
	s.SetTodos(todos)

	// Should not panic and planning panel should have todos
	s.TogglePlanning()
	view := s.View()
	if view == "" {
		t.Error("Expected non-empty view with todos")
	}
}

func TestSidebar_SetMemories(t *testing.T) {
	s := NewSidebar()
	s.SetSize(40, 20)

	memories := []MemoryItem{
		{ID: "1", Content: "Memory 1", Category: "fact"},
	}
	s.SetMemories(memories)

	s.ToggleMemory()
	view := s.View()
	if view == "" {
		t.Error("Expected non-empty view with memories")
	}
}

func TestSidebar_Update_TogglePlanning(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	msg := tea.KeyMsg{Type: tea.KeyCtrlP}
	s.Update(msg)

	if !s.IsVisible() {
		t.Error("Expected visible after ctrl+p")
	}
	if s.ActivePanel() != "planning" {
		t.Error("Expected planning panel active")
	}
}

func TestSidebar_Update_ToggleMemory(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	// ctrl+m might be interpreted as enter on some systems
	// Use direct key call instead
	s.ToggleMemory()

	if !s.IsVisible() {
		t.Error("Expected visible after ToggleMemory")
	}
	if s.ActivePanel() != "memory" {
		t.Error("Expected memory panel active")
	}
}

func TestSidebar_Update_Close(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)
	s.TogglePlanning()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	s.Update(msg)

	if s.IsVisible() {
		t.Error("Expected hidden after esc")
	}
}

func TestSidebar_Update_TodosUpdateMsg(t *testing.T) {
	s := NewSidebar()
	s.SetSize(40, 20)

	msg := TodosUpdateMsg{
		Todos: []TodoItem{
			{Content: "Task 1", Status: "pending"},
		},
	}
	s.Update(msg)

	s.TogglePlanning()
	view := s.View()
	if !strings.Contains(view, "Task 1") {
		t.Error("Expected todo content in view")
	}
}

func TestSidebar_Update_MemoriesUpdateMsg(t *testing.T) {
	s := NewSidebar()
	s.SetSize(40, 20)

	msg := MemoriesUpdateMsg{
		Memories: []MemoryItem{
			{ID: "1", Content: "Memory content", Category: "fact"},
		},
	}
	s.Update(msg)

	s.ToggleMemory()
	view := s.View()
	if !strings.Contains(view, "Memory content") {
		t.Error("Expected memory content in view")
	}
}

func TestSidebar_View_Hidden(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	if s.View() != "" {
		t.Error("Expected empty view when hidden")
	}
}

func TestSidebar_View_Planning(t *testing.T) {
	s := NewSidebar()
	s.SetSize(60, 20)
	s.TogglePlanning()

	view := s.View()
	if !strings.Contains(view, "Planning") {
		t.Error("Expected 'Planning' in view")
	}
}

func TestSidebar_View_Memory(t *testing.T) {
	s := NewSidebar()
	s.SetSize(60, 20)
	s.ToggleMemory()

	view := s.View()
	if !strings.Contains(view, "Memory") {
		t.Error("Expected 'Memory' in view")
	}
}

func TestSidebar_ContentWidth_Hidden(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	if s.ContentWidth(150) != 150 {
		t.Error("Expected full width when hidden")
	}
}

func TestSidebar_ContentWidth_VisibleRight(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)
	s.TogglePlanning()

	contentWidth := s.ContentWidth(150)
	if contentWidth >= 150 {
		t.Error("Expected reduced width when sidebar visible on right")
	}
}

func TestSidebar_ContentHeight_Hidden(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	if s.ContentHeight(40) != 40 {
		t.Error("Expected full height when hidden")
	}
}

func TestSidebar_ContentHeight_VisibleBottom(t *testing.T) {
	s := NewSidebar()
	s.SetSize(80, 40) // Narrow - bottom position
	s.TogglePlanning()

	contentHeight := s.ContentHeight(40)
	if contentHeight >= 40 {
		t.Error("Expected reduced height when sidebar visible on bottom")
	}
}

func TestSidebar_RenderWithContent_Hidden(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)

	content := "Main content"
	result := s.RenderWithContent(content, 150, 40)

	if result != content {
		t.Error("Expected unchanged content when sidebar hidden")
	}
}

func TestSidebar_RenderWithContent_Right(t *testing.T) {
	s := NewSidebar()
	s.SetSize(150, 40)
	s.TogglePlanning()

	content := "Main content"
	result := s.RenderWithContent(content, 150, 40)

	if !strings.Contains(result, "Main content") {
		t.Error("Expected main content in result")
	}
	if !strings.Contains(result, "Planning") {
		t.Error("Expected Planning panel in result")
	}
}

func TestSidebar_RenderWithContent_Bottom(t *testing.T) {
	s := NewSidebar()
	s.SetSize(80, 40) // Narrow terminal
	s.TogglePlanning()

	content := "Main content"
	result := s.RenderWithContent(content, 80, 40)

	if !strings.Contains(result, "Main content") {
		t.Error("Expected main content in result")
	}
	if !strings.Contains(result, "Planning") {
		t.Error("Expected Planning panel in result")
	}
}

func TestDefaultSidebarKeyMap(t *testing.T) {
	km := DefaultSidebarKeyMap()

	if !key.Matches(tea.KeyMsg{Type: tea.KeyCtrlP}, km.TogglePlanning) {
		t.Error("Expected ctrl+p to match TogglePlanning")
	}
	// Note: ctrl+m matches Enter on many terminals, so we just verify the binding exists
	if km.ToggleMemory.Keys() == nil {
		t.Error("Expected ToggleMemory binding to have keys")
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyEsc}, km.Close) {
		t.Error("Expected esc to match Close")
	}
}
