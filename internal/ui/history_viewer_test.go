package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexcabrera/ayo/internal/session"
)

func TestNewHistoryViewer(t *testing.T) {
	messages := []session.Message{
		{
			Role: session.RoleUser,
			Parts: []session.ContentPart{
				session.TextContent{Text: "Hello"},
			},
		},
	}

	viewer := NewHistoryViewer(messages, "@ayo", "Test Session")

	if viewer.agentHandle != "@ayo" {
		t.Errorf("agentHandle = %q, want %q", viewer.agentHandle, "@ayo")
	}

	if viewer.sessionTitle != "Test Session" {
		t.Errorf("sessionTitle = %q, want %q", viewer.sessionTitle, "Test Session")
	}

	if viewer.messageCount != 1 {
		t.Errorf("messageCount = %d, want 1", viewer.messageCount)
	}

	// Content should be pre-rendered
	if viewer.content == "" {
		t.Error("content should be pre-rendered")
	}

	// Default result should be quit
	if viewer.result != HistoryViewerQuit {
		t.Errorf("default result should be HistoryViewerQuit")
	}
}

func TestHistoryViewer_Init(t *testing.T) {
	viewer := NewHistoryViewer(nil, "@ayo", "Test")
	cmd := viewer.Init()

	if cmd != nil {
		t.Error("Init should return nil cmd")
	}
}

func TestHistoryViewer_Update_Quit(t *testing.T) {
	viewer := NewHistoryViewer(nil, "@ayo", "Test")

	// Simulate window size first to initialize viewport
	model, _ := viewer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	viewer = model.(HistoryViewer)

	// Press 'q' to quit
	model, cmd := viewer.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	viewer = model.(HistoryViewer)

	if viewer.Result() != HistoryViewerQuit {
		t.Error("pressing 'q' should set result to HistoryViewerQuit")
	}

	// Should return tea.Quit
	if cmd == nil {
		t.Error("pressing 'q' should return a command")
	}
}

func TestHistoryViewer_Update_Continue(t *testing.T) {
	viewer := NewHistoryViewer(nil, "@ayo", "Test")

	// Simulate window size first
	model, _ := viewer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	viewer = model.(HistoryViewer)

	// Press Enter to continue
	model, cmd := viewer.Update(tea.KeyMsg{Type: tea.KeyEnter})
	viewer = model.(HistoryViewer)

	if viewer.Result() != HistoryViewerContinue {
		t.Error("pressing Enter should set result to HistoryViewerContinue")
	}

	// Should return tea.Quit
	if cmd == nil {
		t.Error("pressing Enter should return a command")
	}
}

func TestHistoryViewer_Update_WindowSize(t *testing.T) {
	viewer := NewHistoryViewer(nil, "@ayo", "Test")

	if viewer.ready {
		t.Error("viewer should not be ready before WindowSizeMsg")
	}

	model, _ := viewer.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	viewer = model.(HistoryViewer)

	if !viewer.ready {
		t.Error("viewer should be ready after WindowSizeMsg")
	}

	if viewer.width != 100 {
		t.Errorf("width = %d, want 100", viewer.width)
	}

	if viewer.height != 40 {
		t.Errorf("height = %d, want 40", viewer.height)
	}
}

func TestHistoryViewer_View_NotReady(t *testing.T) {
	viewer := NewHistoryViewer(nil, "@ayo", "Test")

	view := viewer.View()

	if !strings.Contains(view, "Loading") {
		t.Errorf("view before ready should show loading, got: %q", view)
	}
}

func TestHistoryViewer_View_Ready(t *testing.T) {
	messages := []session.Message{
		{
			Role: session.RoleUser,
			Parts: []session.ContentPart{
				session.TextContent{Text: "Hello world"},
			},
		},
	}

	viewer := NewHistoryViewer(messages, "@ayo", "Test Session")

	// Initialize with window size
	model, _ := viewer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	viewer = model.(HistoryViewer)

	view := viewer.View()

	// Should contain header with agent handle
	if !strings.Contains(view, "@ayo") {
		t.Error("view should contain agent handle in header")
	}

	// Should contain message count
	if !strings.Contains(view, "1 messages") {
		t.Error("view should contain message count")
	}

	// Should contain help text
	if !strings.Contains(view, "continue") {
		t.Error("view should contain help text with 'continue'")
	}

	if !strings.Contains(view, "quit") {
		t.Error("view should contain help text with 'quit'")
	}
}

func TestHistoryViewer_Update_Navigation(t *testing.T) {
	// Create viewer with enough content to scroll
	longContent := strings.Repeat("Line of text\n", 100)
	messages := []session.Message{
		{
			Role: session.RoleUser,
			Parts: []session.ContentPart{
				session.TextContent{Text: longContent},
			},
		},
	}

	viewer := NewHistoryViewer(messages, "@ayo", "Test")

	// Initialize with small window to ensure scrolling is needed
	model, _ := viewer.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	viewer = model.(HistoryViewer)

	// Test 'g' goes to top
	model, _ = viewer.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	viewer = model.(HistoryViewer)

	if !viewer.viewport.AtTop() {
		t.Error("'g' should go to top")
	}

	// Test 'G' goes to bottom
	model, _ = viewer.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	viewer = model.(HistoryViewer)

	if !viewer.viewport.AtBottom() {
		t.Error("'G' should go to bottom")
	}
}

func TestHistoryViewer_EscQuits(t *testing.T) {
	viewer := NewHistoryViewer(nil, "@ayo", "Test")

	// Initialize
	model, _ := viewer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	viewer = model.(HistoryViewer)

	// Press Esc
	model, cmd := viewer.Update(tea.KeyMsg{Type: tea.KeyEsc})
	viewer = model.(HistoryViewer)

	if viewer.Result() != HistoryViewerQuit {
		t.Error("Esc should quit")
	}

	if cmd == nil {
		t.Error("Esc should return quit command")
	}
}

func TestHistoryViewer_CtrlCQuits(t *testing.T) {
	viewer := NewHistoryViewer(nil, "@ayo", "Test")

	// Initialize
	model, _ := viewer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	viewer = model.(HistoryViewer)

	// Press Ctrl+C
	model, cmd := viewer.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	viewer = model.(HistoryViewer)

	if viewer.Result() != HistoryViewerQuit {
		t.Error("Ctrl+C should quit")
	}

	if cmd == nil {
		t.Error("Ctrl+C should return quit command")
	}
}

func TestHistoryViewerResult_Constants(t *testing.T) {
	// Verify result constants are distinct
	if HistoryViewerQuit == HistoryViewerContinue {
		t.Error("HistoryViewerQuit and HistoryViewerContinue should be different")
	}
}
