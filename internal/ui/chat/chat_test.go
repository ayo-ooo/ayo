package chat

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/skills"
	"github.com/alexcabrera/ayo/internal/ui/chat/panels"
)

// mockAgent creates a minimal agent for testing.
func mockAgent(handle string) agent.Agent {
	return agent.Agent{
		Handle: handle,
		Skills: []skills.Metadata{},
	}
}

// mockSendFn creates a SendMessageFunc that returns a fixed response.
func mockSendFn(response string, err error) SendMessageFunc {
	return func(ctx context.Context, message string) (string, error) {
		return response, err
	}
}

// initModel initializes a model with a window size for testing.
func initModel(m Model, width, height int) Model {
	model, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return model.(Model)
}

func TestNew(t *testing.T) {
	ag := mockAgent("@test")
	sendFn := mockSendFn("response", nil)

	m := New(ag, "session-123", sendFn)

	if m.agentHandle != "@test" {
		t.Errorf("agentHandle = %q, want %q", m.agentHandle, "@test")
	}

	if m.sessionID != "session-123" {
		t.Errorf("sessionID = %q, want %q", m.sessionID, "session-123")
	}

	if m.state != StateInput {
		t.Errorf("initial state = %v, want StateInput", m.state)
	}

	if m.sidebar == nil {
		t.Error("sidebar should be initialized")
	}

	if len(m.messages) != 0 {
		t.Errorf("messages should be empty, got %d", len(m.messages))
	}
}

func TestInit(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))

	cmd := m.Init()

	if cmd == nil {
		t.Error("Init should return a command (batch of textarea.Blink and tickSpinner)")
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))

	if m.ready {
		t.Error("model should not be ready before WindowSizeMsg")
	}

	model, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = model.(Model)

	if !m.ready {
		t.Error("model should be ready after WindowSizeMsg")
	}

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}

	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestUpdate_TextDeltaMsg(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// Send text delta
	model, _ := m.Update(TextDeltaMsg{Delta: "Hello "})
	m = model.(Model)

	if m.state != StateStreaming {
		t.Errorf("state should be StateStreaming after TextDeltaMsg")
	}

	if m.streamBuffer.String() != "Hello " {
		t.Errorf("streamBuffer = %q, want %q", m.streamBuffer.String(), "Hello ")
	}

	// Send more text
	model, _ = m.Update(TextDeltaMsg{Delta: "World"})
	m = model.(Model)

	if m.streamBuffer.String() != "Hello World" {
		t.Errorf("streamBuffer = %q, want %q", m.streamBuffer.String(), "Hello World")
	}
}

func TestUpdate_TextEndMsg(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// Stream some text
	model, _ := m.Update(TextDeltaMsg{Delta: "Test response"})
	m = model.(Model)

	// TextEndMsg now stays in streaming (EventDone completes the flow)
	model, _ = m.Update(TextEndMsg{})
	m = model.(Model)

	// State stays in streaming until EventDone
	if m.state != StateStreaming {
		t.Errorf("state should be StateStreaming after TextEndMsg, got %v", m.state)
	}

	// But message should be accumulated
	if len(m.messages) != 1 {
		t.Fatalf("messages count = %d, want 1", len(m.messages))
	}

	if m.messages[0].Role != "assistant" {
		t.Errorf("message role = %q, want %q", m.messages[0].Role, "assistant")
	}

	if m.messages[0].Content != "Test response" {
		t.Errorf("message content = %q, want %q", m.messages[0].Content, "Test response")
	}

	if m.streamBuffer.Len() != 0 {
		t.Error("streamBuffer should be reset after TextEndMsg")
	}

	// Now send EventDone to complete the flow
	model, _ = m.Update(run.StreamEvent{Type: run.EventDone, Response: "Test response"})
	m = model.(Model)

	if m.state != StateInput {
		t.Errorf("state should be StateInput after EventDone, got %v", m.state)
	}
}

func TestUpdate_ToolCallStartMsg(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	msg := ToolCallStartMsg{
		Name:        "bash",
		Description: "Running tests",
		Command:     "go test ./...",
	}

	model, _ := m.Update(msg)
	m = model.(Model)

	if m.currentToolCall == nil {
		t.Fatal("currentToolCall should be set")
	}

	if m.currentToolCall.Name != "bash" {
		t.Errorf("currentToolCall.Name = %q, want %q", m.currentToolCall.Name, "bash")
	}

	if m.currentToolCall.Description != "Running tests" {
		t.Errorf("currentToolCall.Description = %q, want %q", m.currentToolCall.Description, "Running tests")
	}
}

func TestUpdate_ToolCallResultMsg(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// Start a tool call
	model, _ := m.Update(ToolCallStartMsg{Name: "bash", Description: "Running tests"})
	m = model.(Model)

	// Complete the tool call
	model, _ = m.Update(ToolCallResultMsg{
		Name:     "bash",
		Output:   "All tests passed",
		Duration: "1.5s",
	})
	m = model.(Model)

	// Tool message should be added
	if len(m.messages) == 0 {
		t.Error("tool result should add a message")
	}
}

func TestUpdate_ReasoningMessages(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// Start reasoning
	model, _ := m.Update(ReasoningStartMsg{})
	m = model.(Model)

	if m.thinkingStartTime.IsZero() {
		t.Error("thinkingStartTime should be set")
	}

	// Add reasoning content
	model, _ = m.Update(ReasoningDeltaMsg{Delta: "Thinking about"})
	m = model.(Model)

	if m.reasoningBuffer.String() != "Thinking about" {
		t.Errorf("reasoningBuffer = %q, want %q", m.reasoningBuffer.String(), "Thinking about")
	}

	// More reasoning
	model, _ = m.Update(ReasoningDeltaMsg{Delta: " the problem..."})
	m = model.(Model)

	if m.reasoningBuffer.String() != "Thinking about the problem..." {
		t.Errorf("reasoningBuffer = %q", m.reasoningBuffer.String())
	}

	// End reasoning
	model, _ = m.Update(ReasoningEndMsg{Duration: "2.5s"})
	m = model.(Model)

	if m.reasoningBuffer.Len() != 0 {
		t.Error("reasoningBuffer should be reset after ReasoningEndMsg")
	}
}

func TestUpdate_TodosUpdateMsg(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	todos := []panels.TodoItem{
		{Content: "Task 1", ActiveForm: "Working on task 1", Status: "in_progress"},
		{Content: "Task 2", ActiveForm: "Working on task 2", Status: "pending"},
	}

	model, _ := m.Update(panels.TodosUpdateMsg{Todos: todos})
	m = model.(Model)

	// Sidebar should have updated (we verify by checking it doesn't crash)
	// More detailed verification would require exposing sidebar state
	_ = m.View()
}

func TestUpdate_MemoriesUpdateMsg(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	memories := []panels.MemoryItem{
		{ID: "1", Content: "User prefers dark mode", Category: "preference", Scope: "global"},
		{ID: "2", Content: "Project uses Go", Category: "fact", Scope: "@test"},
	}

	model, _ := m.Update(panels.MemoriesUpdateMsg{Memories: memories})
	m = model.(Model)

	// Sidebar should have updated
	_ = m.View()
}

func TestView_NotReady(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))

	view := m.View()

	if !strings.Contains(view, "Loading") {
		t.Errorf("view before ready should show loading, got: %q", view)
	}
}

func TestView_Ready(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	view := m.View()

	// Should contain agent handle in header
	if !strings.Contains(view, "@test") {
		t.Error("view should contain agent handle")
	}

	// Should contain textarea (has prompt)
	if !strings.Contains(view, ">") {
		t.Error("view should contain textarea prompt")
	}
}

func TestView_WithStreamingContent(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// Stream some content
	model, _ := m.Update(TextDeltaMsg{Delta: "Streaming response here"})
	m = model.(Model)

	view := m.View()

	// View should include the streaming content
	if !strings.Contains(view, "Streaming") {
		t.Error("view should contain streaming content")
	}
}

func TestView_WithToolCall(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// Start a tool call
	model, _ := m.Update(ToolCallStartMsg{
		Name:        "bash",
		Description: "Installing dependencies",
	})
	m = model.(Model)

	view := m.View()

	// View should show tool call info
	if !strings.Contains(view, "Installing") || !strings.Contains(view, "bash") {
		t.Errorf("view should contain tool call info, got: %s", view)
	}
}

func TestKeyQuit_InputState(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// Press Ctrl+C in input state
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = model.(Model)

	// Should return quit command
	if cmd == nil {
		t.Error("Ctrl+C should return a command")
	}
}

func TestState_Transitions(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// Initial state is Input
	if m.state != StateInput {
		t.Errorf("initial state = %v, want StateInput", m.state)
	}

	// TextDeltaMsg transitions to Streaming
	model, _ := m.Update(TextDeltaMsg{Delta: "Hi"})
	m = model.(Model)

	if m.state != StateStreaming {
		t.Errorf("state after TextDelta = %v, want StateStreaming", m.state)
	}

	// TextEndMsg now stays in StateStreaming (EventDone transitions to Input)
	model, _ = m.Update(TextEndMsg{})
	m = model.(Model)

	if m.state != StateStreaming {
		t.Errorf("state after TextEnd = %v, want StateStreaming", m.state)
	}

	// EventDone transitions to Input
	model, _ = m.Update(run.StreamEvent{Type: run.EventDone, Response: "Hello"})
	m = model.(Model)

	if m.state != StateInput {
		t.Errorf("state after EventDone = %v, want StateInput", m.state)
	}
}

// TestMultipleMessageFlow simulates a complete conversation flow.
func TestMultipleMessageFlow(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// First response
	model, _ := m.Update(TextDeltaMsg{Delta: "First response"})
	m = model.(Model)
	model, _ = m.Update(TextEndMsg{})
	m = model.(Model)

	if len(m.messages) != 1 {
		t.Errorf("messages count = %d, want 1", len(m.messages))
	}

	// Second response
	model, _ = m.Update(TextDeltaMsg{Delta: "Second response"})
	m = model.(Model)
	model, _ = m.Update(TextEndMsg{})
	m = model.(Model)

	if len(m.messages) != 2 {
		t.Errorf("messages count = %d, want 2", len(m.messages))
	}
}

// TestToolCallWithNestedContent simulates tool calls with responses.
func TestToolCallWithNestedContent(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// Stream some initial text
	model, _ := m.Update(TextDeltaMsg{Delta: "Let me check that..."})
	m = model.(Model)

	// Tool call starts (interrupts streaming)
	model, _ = m.Update(ToolCallStartMsg{Name: "bash", Description: "Checking"})
	m = model.(Model)

	// Tool call completes
	model, _ = m.Update(ToolCallResultMsg{Name: "bash", Output: "OK", Duration: "0.5s"})
	m = model.(Model)

	// More streaming
	model, _ = m.Update(TextDeltaMsg{Delta: " Done!"})
	m = model.(Model)

	// End streaming
	model, _ = m.Update(TextEndMsg{})
	m = model.(Model)

	// Should have messages recorded
	if len(m.messages) == 0 {
		t.Error("should have messages after flow")
	}
}

// Group C: Input Field Improvements Tests

func TestTextarea_InitialHeight(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))

	// Initial height should be 3
	if m.textarea.Height() != 3 {
		t.Errorf("initial textarea height = %d, want 3", m.textarea.Height())
	}
}

func TestTextarea_CharLimitUnlimited(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))

	// CharLimit should be 0 (unlimited)
	// Note: We can verify this works by checking the textarea accepts content
	m = initModel(m, 100, 40)

	// The textarea should accept any amount of text
	// This is a basic sanity check
	_ = m.View()
}

func TestTextarea_WidthSetOnResize(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	initialWidth := m.textarea.Width()
	if initialWidth <= 0 {
		t.Error("textarea width should be set after resize")
	}

	// Resize to different width
	model, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	m = model.(Model)

	newWidth := m.textarea.Width()
	if newWidth <= initialWidth {
		t.Errorf("textarea width should increase with wider terminal, got %d from %d", newWidth, initialWidth)
	}
}

func TestNewlineKeybinding(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// Newline keybinding should be defined
	if len(m.keyMap.Newline.Keys()) == 0 {
		t.Error("Newline keybinding should have keys defined")
	}

	// Should include shift+enter
	keys := m.keyMap.Newline.Keys()
	hasShiftEnter := false
	for _, k := range keys {
		if k == "shift+enter" {
			hasShiftEnter = true
			break
		}
	}
	if !hasShiftEnter {
		t.Errorf("Newline keybinding should include shift+enter, got: %v", keys)
	}
}

func TestStatusBar_HasHints(t *testing.T) {
	ag := mockAgent("@test")
	m := New(ag, "session-123", mockSendFn("", nil))
	m = initModel(m, 100, 40)

	// StatusBar should have hints
	view := m.View()

	// Should contain some hints
	if !strings.Contains(view, "enter") {
		t.Error("view should contain hint about enter key")
	}
}
