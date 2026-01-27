// Package chat provides a full-screen TUI chat interface using Bubble Tea.
// It features a scrollable message viewport and pinned input textarea at the bottom.
package chat

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/editor"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/ui/chat/messages"
	"github.com/alexcabrera/ayo/internal/ui/chat/panels"
)

// Result indicates the outcome of the chat session.
type Result int

const (
	// ResultQuit means the user exited the chat.
	ResultQuit Result = iota
	// ResultError means an error occurred.
	ResultError
)

// State represents the current state of the chat.
type State int

const (
	// StateInput means the user is typing.
	StateInput State = iota
	// StateWaiting means we're waiting for agent response.
	StateWaiting
	// StateStreaming means we're receiving a streaming response.
	StateStreaming
)

// SendMessageFunc is a callback to send a message to the agent.
// Returns the agent's response or an error.
type SendMessageFunc func(ctx context.Context, message string) (string, error)

// Model is the Bubble Tea model for the chat interface.
type Model struct {
	// Configuration
	agentHandle string
	skillCount  int
	sessionID   string
	sendFn      SendMessageFunc
	ctx         context.Context
	cancelFn    context.CancelFunc

	// Components
	viewport  viewport.Model
	textarea  textarea.Model
	sidebar   *panels.Sidebar
	statusBar *StatusBar
	keyMap    KeyMap

	// State
	state        State
	messages     []message
	streamBuffer strings.Builder
	ready        bool
	width        int
	height       int
	err          error

	// Tool/reasoning state
	currentToolCall   *ToolCallStartMsg
	toolCallTree      *messages.ToolCallTree // B.07: Tree-based tool rendering
	reasoningBuffer   strings.Builder
	thinkingStartTime time.Time

	// Spinner animation
	spinnerFrame int
	spinnerTick  bool

	// Scrollback dump content (for exit)
	scrollbackContent string
}

// message represents a single message in the conversation.
type message struct {
	Role    string // "user" or "assistant"
	Content string
}

// KeyMap defines the keybindings for the chat.
type KeyMap struct {
	Send       key.Binding
	Newline    key.Binding
	Editor     key.Binding
	History    key.Binding
	Quit       key.Binding
	Interrupt  key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send"),
		),
		Newline: key.NewBinding(
			key.WithKeys("shift+enter", "alt+enter"),
			key.WithHelp("shift+enter", "newline"),
		),
		Editor: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "editor"),
		),
		History: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "history"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Interrupt: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "interrupt"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
	}
}

// New creates a new chat model.
func New(ag agent.Agent, sessionID string, sendFn SendMessageFunc) Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.Prompt = "> "
	ta.CharLimit = 0 // No limit
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.KeyMap.InsertNewline.SetEnabled(false) // We handle newlines ourselves

	return Model{
		agentHandle:  ag.Handle,
		skillCount:   len(ag.Skills),
		sessionID:    sessionID,
		sendFn:       sendFn,
		textarea:     ta,
		sidebar:      panels.NewSidebar(),
		statusBar:    NewStatusBar(),
		toolCallTree: messages.NewToolCallTree(),
		keyMap:       DefaultKeyMap(),
		state:        StateInput,
		messages:     []message{},
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.tickSpinner())
}

// tickMsg triggers spinner animation.
type tickMsg struct{}

// tickSpinner returns a command that ticks the spinner.
func (m Model) tickSpinner() tea.Cmd {
	return tea.Tick(time.Millisecond*80, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// AgentResponseMsg is sent when the agent responds.
type AgentResponseMsg struct {
	Response string
	Err      error
}

// StreamDeltaMsg is sent for streaming content.
type StreamDeltaMsg struct {
	Delta string
}

// StreamEndMsg signals the end of streaming.
type StreamEndMsg struct{}

// OpenEditorMsg is sent when returning from external editor.
type OpenEditorMsg struct {
	Text string
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Track sidebar visibility before update
	sidebarWasVisible := m.sidebar.IsVisible()

	// Handle sidebar messages first
	if cmd := m.sidebar.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Update hints if sidebar visibility changed
	if m.sidebar.IsVisible() != sidebarWasVisible {
		m.updateStatusBarHints()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.sidebar.SetSize(msg.Width, msg.Height)
		return m.handleResize(), nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tickMsg:
		if m.state == StateWaiting || m.state == StateStreaming {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
			m.updateViewportContent()
		}
		return m, m.tickSpinner()

	case AgentResponseMsg:
		return m.handleAgentResponse(msg)

	case StreamDeltaMsg:
		return m.handleStreamDelta(msg)

	case StreamEndMsg:
		return m.handleStreamEnd()

	case TextDeltaMsg:
		return m.handleTextDelta(msg)

	case TextEndMsg:
		return m.handleTextEnd()

	case ToolCallStartMsg:
		return m.handleToolCallStart(msg)

	case ToolCallResultMsg:
		return m.handleToolCallResult(msg)

	case SubAgentStartMsg:
		return m.handleSubAgentStart(msg)

	case SubAgentEndMsg:
		return m.handleSubAgentEnd(msg)

	case ReasoningStartMsg:
		m.thinkingStartTime = time.Now()
		m.reasoningBuffer.Reset()
		m.updateViewportContent()
		return m, nil

	case ReasoningDeltaMsg:
		m.reasoningBuffer.WriteString(msg.Delta)
		m.updateViewportContent()
		return m, nil

	case ReasoningEndMsg:
		m.reasoningBuffer.Reset()
		m.updateViewportContent()
		return m, nil

	case OpenEditorMsg:
		m.textarea.SetValue(msg.Text)
		m.textarea.CursorEnd()
		return m, nil

	case panels.TodosUpdateMsg:
		m.sidebar.SetTodos(msg.Todos)
		// Update status bar with task progress
		var completed, total int
		var current string
		for _, todo := range msg.Todos {
			total++
			if todo.Status == "completed" {
				completed++
			} else if todo.Status == "in_progress" {
				current = todo.ActiveForm
			}
		}
		m.statusBar.SetTaskProgress(current, completed, total)
		return m, nil

	case panels.MemoriesUpdateMsg:
		m.sidebar.SetMemories(msg.Memories)
		m.statusBar.SetMemoryCount(len(msg.Memories))
		return m, nil
	}

	// Update textarea if in input state and sidebar not focused
	if m.state == StateInput && !m.sidebar.IsFocused() {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)

		// Update textarea height dynamically based on content
		m.updateTextareaHeight()
	}

	// Update viewport for scroll events (if sidebar not focused)
	if !m.sidebar.IsFocused() {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// Spinner frames
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// setState updates the state and refreshes status bar hints.
func (m *Model) setState(state State) {
	m.state = state
	m.updateStatusBarHints()
}

// updateStatusBarHints refreshes the status bar hints based on current state.
func (m *Model) updateStatusBarHints() {
	var hints string
	switch m.state {
	case StateInput:
		// Show line indicator for multiline input
		content := m.textarea.Value()
		lineCount := strings.Count(content, "\n") + 1
		if lineCount > 1 {
			row := m.textarea.Line()
			hints = fmt.Sprintf("line %d/%d · ", row+1, lineCount)
		}
		hints += "enter send · shift+enter newline · ctrl+e editor · ctrl+c quit"
		if m.sidebar.IsVisible() {
			hints += " · ctrl+p plan · ctrl+m memory"
		}
	case StateWaiting, StateStreaming:
		hints = "ctrl+c interrupt"
	}
	m.statusBar.SetHints(hints)
}

// updateTextareaHeight adjusts textarea height based on content.
// Min height: 3, Max height: 10.
func (m *Model) updateTextareaHeight() {
	const minHeight = 3
	const maxHeight = 10

	// Count lines in textarea content
	content := m.textarea.Value()
	lineCount := strings.Count(content, "\n") + 1

	// Add 1 for the prompt line
	desiredHeight := lineCount + 1
	if desiredHeight < minHeight {
		desiredHeight = minHeight
	}
	if desiredHeight > maxHeight {
		desiredHeight = maxHeight
	}

	currentHeight := m.textarea.Height()
	if currentHeight != desiredHeight {
		m.textarea.SetHeight(desiredHeight)
		// Recalculate viewport height
		m.handleResize()
	}

	// Update hints to show line indicator for multiline
	m.updateStatusBarHints()
}

// handleTextDelta handles streaming text.
func (m Model) handleTextDelta(msg TextDeltaMsg) (tea.Model, tea.Cmd) {
	m.setState(StateStreaming)
	m.streamBuffer.WriteString(msg.Delta)
	m.updateViewportContent()
	m.viewport.GotoBottom()
	return m, nil
}

// handleTextEnd handles end of text streaming.
func (m Model) handleTextEnd() (tea.Model, tea.Cmd) {
	if m.streamBuffer.Len() > 0 {
		m.messages = append(m.messages, message{
			Role:    "assistant",
			Content: m.streamBuffer.String(),
		})
		m.streamBuffer.Reset()
	}
	m.setState(StateInput)
	m.updateViewportContent()
	return m, nil
}

// handleToolCallStart handles the start of a tool call.
func (m Model) handleToolCallStart(msg ToolCallStartMsg) (tea.Model, tea.Cmd) {
	m.currentToolCall = &msg

	// Create ToolCallCmp and add to tree (B.07)
	tc := messages.ToolCall{
		ID:    msg.ID,
		Name:  msg.Name,
		Input: msg.Input,
	}

	if msg.ParentID != "" {
		// B.08: Nested tool call - find parent and add as nested
		if parent := m.toolCallTree.Get(msg.ParentID); parent != nil {
			nestedCmp := messages.NewToolCallCmp("", tc, messages.WithToolCallNested(true))
			nestedCmp.SetSize(m.viewport.Width-4, 0)
			parent.SetNestedToolCalls(append(parent.GetNestedToolCalls(), nestedCmp))
		}
	} else {
		// Top-level tool call
		cmp := messages.NewToolCallCmp("", tc)
		cmp.SetSize(m.viewport.Width, 0)
		m.toolCallTree.Add(cmp)
	}

	m.updateViewportContent()
	m.viewport.GotoBottom()
	return m, nil
}

// handleToolCallResult handles the completion of a tool call.
func (m Model) handleToolCallResult(msg ToolCallResultMsg) (tea.Model, tea.Cmd) {
	// Update ToolCallCmp in tree (B.07)
	if cmp := m.toolCallTree.Get(msg.ID); cmp != nil {
		result := messages.ToolResult{
			ToolCallID: msg.ID,
			Name:       msg.Name,
			Content:    msg.Output,
			IsError:    msg.Error != "",
			Metadata:   msg.Metadata,
		}
		if msg.Error != "" {
			result.Content = msg.Error
		}
		cmp.SetToolResult(result)
	}

	// Add tool result to messages for display (legacy format)
	toolContent := fmt.Sprintf("**%s** %s\n```\n%s\n```",
		msg.Name,
		msg.Duration,
		truncateOutput(msg.Output, 20))

	if msg.Error != "" {
		toolContent = fmt.Sprintf("**%s** (failed) %s\n```\n%s\n```",
			msg.Name,
			msg.Duration,
			msg.Error)
	}

	m.messages = append(m.messages, message{
		Role:    "tool",
		Content: toolContent,
	})
	m.currentToolCall = nil

	// B.10: Auto-collapse completed tool calls with nested children
	m.toolCallTree.AutoCollapse()

	m.updateViewportContent()
	m.viewport.GotoBottom()
	return m, nil
}

// handleSubAgentStart handles the start of a sub-agent invocation (B.08).
func (m Model) handleSubAgentStart(msg SubAgentStartMsg) (tea.Model, tea.Cmd) {
	// Sub-agent calls are treated as agent tool calls
	// They will be added via ToolCallStartMsg with the appropriate ParentID
	m.updateViewportContent()
	m.viewport.GotoBottom()
	return m, nil
}

// handleSubAgentEnd handles the completion of a sub-agent (B.08).
func (m Model) handleSubAgentEnd(msg SubAgentEndMsg) (tea.Model, tea.Cmd) {
	// Sub-agent completion is handled via ToolCallResultMsg
	m.updateViewportContent()
	return m, nil
}

// truncateOutput limits output to maxLines.
func truncateOutput(output string, maxLines int) string {
	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return output
	}
	head := lines[:maxLines/2]
	tail := lines[len(lines)-maxLines/2:]
	return strings.Join(head, "\n") + fmt.Sprintf("\n... (%d lines omitted) ...\n", len(lines)-maxLines) + strings.Join(tail, "\n")
}

// handleResize adjusts layout when window size changes.
func (m Model) handleResize() Model {
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	inputHeight := m.textarea.Height() + 2 // +2 for borders/padding

	// Calculate content area accounting for sidebar
	contentWidth := m.sidebar.ContentWidth(m.width)
	contentHeight := m.sidebar.ContentHeight(m.height)

	viewportHeight := contentHeight - headerHeight - footerHeight - inputHeight

	if !m.ready {
		m.viewport = viewport.New(contentWidth, viewportHeight)
		m.viewport.MouseWheelEnabled = true
		m.viewport.MouseWheelDelta = 3
		m.ready = true
	} else {
		m.viewport.Width = contentWidth
		m.viewport.Height = viewportHeight
	}

	m.textarea.SetWidth(contentWidth - 4) // -4 for padding
	m.statusBar.SetWidth(contentWidth)    // Set status bar width
	m.toolCallTree.SetWidth(contentWidth) // B.07: Set tree width
	m.updateViewportContent()

	return m
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		if m.state == StateInput {
			m.scrollbackContent = m.renderScrollback()
			return m, tea.Quit
		}
		// If waiting/streaming, cancel the request
		if m.cancelFn != nil {
			m.cancelFn()
		}
		m.setState(StateInput)
		return m, nil

	case key.Matches(msg, m.keyMap.Send) && m.state == StateInput:
		return m.sendMessage()

	case key.Matches(msg, m.keyMap.Newline) && m.state == StateInput:
		m.textarea.InsertRune('\n')
		return m, nil

	case key.Matches(msg, m.keyMap.Editor) && m.state == StateInput:
		return m.openEditor()

	case key.Matches(msg, m.keyMap.History):
		// TODO: Open history viewer dialog
		return m, nil
	}

	// Pass key to textarea if in input state
	if m.state == StateInput {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	return m, nil
}

// sendMessage sends the current input to the agent.
func (m Model) sendMessage() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.textarea.Value())
	if text == "" {
		return m, nil
	}

	// Add user message
	m.messages = append(m.messages, message{Role: "user", Content: text})
	m.textarea.Reset()
	m.updateViewportContent()
	m.viewport.GotoBottom()

	// Create cancellable context for this request
	ctx, cancel := context.WithCancel(m.ctx)
	m.cancelFn = cancel
	m.setState(StateWaiting)

	// Send message asynchronously
	return m, func() tea.Msg {
		response, err := m.sendFn(ctx, text)
		return AgentResponseMsg{Response: response, Err: err}
	}
}

// handleAgentResponse processes the agent's response.
func (m Model) handleAgentResponse(msg AgentResponseMsg) (tea.Model, tea.Cmd) {
	m.setState(StateInput)
	m.cancelFn = nil

	if msg.Err != nil {
		if msg.Err != context.Canceled {
			m.err = msg.Err
		}
		return m, nil
	}

	m.messages = append(m.messages, message{Role: "assistant", Content: msg.Response})
	m.updateViewportContent()
	m.viewport.GotoBottom()

	return m, nil
}

// handleStreamDelta appends streaming content.
func (m Model) handleStreamDelta(msg StreamDeltaMsg) (tea.Model, tea.Cmd) {
	m.setState(StateStreaming)
	m.streamBuffer.WriteString(msg.Delta)
	m.updateViewportContent()
	m.viewport.GotoBottom()
	return m, nil
}

// handleStreamEnd finalizes streaming.
func (m Model) handleStreamEnd() (tea.Model, tea.Cmd) {
	if m.streamBuffer.Len() > 0 {
		m.messages = append(m.messages, message{
			Role:    "assistant",
			Content: m.streamBuffer.String(),
		})
		m.streamBuffer.Reset()
	}
	m.setState(StateInput)
	m.updateViewportContent()
	return m, nil
}

// openEditor opens the external editor.
func (m Model) openEditor() (tea.Model, tea.Cmd) {
	value := m.textarea.Value()

	tmpfile, err := os.CreateTemp("", "ayo_msg_*.md")
	if err != nil {
		m.err = err
		return m, nil
	}
	defer tmpfile.Close()

	if _, err := tmpfile.WriteString(value); err != nil {
		m.err = err
		return m, nil
	}

	cmd, err := editor.Command("ayo", tmpfile.Name())
	if err != nil {
		m.err = err
		return m, nil
	}

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return nil
		}
		content, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			return nil
		}
		os.Remove(tmpfile.Name())
		return OpenEditorMsg{Text: strings.TrimSpace(string(content))}
	})
}

// updateViewportContent renders all messages to the viewport.
func (m *Model) updateViewportContent() {
	var content strings.Builder

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			content.WriteString(m.renderUserMessage(msg.Content))
		case "assistant":
			content.WriteString(m.renderAssistantMessage(msg.Content))
		case "tool":
			content.WriteString(m.renderToolMessage(msg.Content))
		}
		content.WriteString("\n\n")
	}

	// Add reasoning content if any
	if m.reasoningBuffer.Len() > 0 {
		content.WriteString(m.renderReasoning(m.reasoningBuffer.String()))
		content.WriteString("\n")
	}

	// Add current tool call if any
	if m.currentToolCall != nil {
		content.WriteString(m.renderToolInProgress(*m.currentToolCall))
		content.WriteString("\n")
	}

	// Add streaming content if any
	if m.streamBuffer.Len() > 0 {
		content.WriteString(m.renderAssistantMessage(m.streamBuffer.String()))
	}

	// Add waiting indicator
	if m.state == StateWaiting && m.currentToolCall == nil && m.reasoningBuffer.Len() == 0 {
		content.WriteString(m.renderWaiting())
	}

	m.viewport.SetContent(content.String())
}

// renderUserMessage styles a user message.
func (m Model) renderUserMessage(content string) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#67e8f9")).
		Bold(true)

	return labelStyle.Render("> ") + content
}

// renderAssistantMessage styles an assistant message.
func (m Model) renderAssistantMessage(content string) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a78bfa")).
		Bold(true)

	// Use glamour for markdown rendering
	rendered := renderMarkdown(content, m.width-4)

	return labelStyle.Render(m.agentHandle) + "\n" + rendered
}

// renderMarkdown renders markdown content using glamour.
func renderMarkdown(content string, width int) string {
	if width < 40 {
		width = 40
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}
	rendered, err := r.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimSpace(rendered)
}

// renderToolMessage renders a completed tool call.
func (m Model) renderToolMessage(content string) string {
	toolStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#fbbf24"))

	return toolStyle.Render("  ") + renderMarkdown(content, m.width-6)
}

// renderToolInProgress renders a tool that is currently executing.
func (m Model) renderToolInProgress(tc ToolCallStartMsg) string {
	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#fbbf24")).
		Bold(true)
	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#fbbf24")).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ca3af"))
	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a78bfa"))

	spinner := spinnerStyle.Render(spinnerFrames[m.spinnerFrame])

	var line string
	if tc.Name == "bash" {
		line = fmt.Sprintf("  %s %s %s · %s",
			iconStyle.Render("❯"),
			nameStyle.Render("bash"),
			descStyle.Render(tc.Description),
			spinner)
		if tc.Command != "" {
			cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
			line += "\n    " + cmdStyle.Render("$ "+tc.Command)
		}
	} else {
		line = fmt.Sprintf("  %s %s · %s",
			iconStyle.Render("▶"),
			nameStyle.Render(tc.Name),
			spinner)
	}

	return line
}

// renderReasoning renders thinking/reasoning content.
func (m Model) renderReasoning(content string) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		Italic(true)
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ca3af")).
		Italic(true)

	// Truncate reasoning to last few lines
	lines := strings.Split(content, "\n")
	if len(lines) > 5 {
		lines = lines[len(lines)-5:]
	}
	truncated := strings.Join(lines, "\n")

	return labelStyle.Render("  Thinking: ") + contentStyle.Render(truncated)
}

// renderWaiting shows a waiting indicator.
func (m Model) renderWaiting() string {
	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a78bfa"))
	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		Italic(true)

	return spinnerStyle.Render(spinnerFrames[m.spinnerFrame]) + " " + textStyle.Render("Thinking...")
}

// View renders the model.
func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	// Render main content
	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		m.headerView(),
		m.viewport.View(),
		m.inputView(),
		m.footerView(),
	)

	// Combine with sidebar if visible
	if m.sidebar.IsVisible() {
		return m.sidebar.RenderWithContent(mainContent, m.width, m.height)
	}

	return mainContent
}

// headerView renders the header bar.
func (m Model) headerView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a78bfa")).
		Bold(true)

	skillStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#fbbf24"))

	lineStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3f3f46"))

	title := titleStyle.Render(fmt.Sprintf("Chat with %s", m.agentHandle))

	var skillsInfo string
	if m.skillCount > 0 {
		skillsInfo = skillStyle.Render(fmt.Sprintf(" (%d skills)", m.skillCount))
	}

	contentWidth := lipgloss.Width(title) + lipgloss.Width(skillsInfo) + 4
	lineWidth := m.width - contentWidth
	if lineWidth < 0 {
		lineWidth = 0
	}

	line := lineStyle.Render(strings.Repeat("─", lineWidth))

	return "  " + title + skillsInfo + " " + line
}

// inputView renders the input textarea.
func (m Model) inputView() string {
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#a78bfa")).
		Padding(0, 1)

	if m.state != StateInput {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("#3f3f46"))
	}

	return inputStyle.Width(m.width - 4).Render(m.textarea.View())
}

// footerView renders the footer with status bar.
func (m Model) footerView() string {
	return m.statusBar.Render()
}

// renderScrollback generates content for terminal scrollback after exit.
func (m Model) renderScrollback() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("─── Session with %s ", m.agentHandle))
	sb.WriteString(strings.Repeat("─", 50))
	sb.WriteString("\n\n")

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			sb.WriteString(fmt.Sprintf("> %s\n\n", msg.Content))
		case "assistant":
			sb.WriteString(fmt.Sprintf("%s:\n%s\n\n", m.agentHandle, msg.Content))
		}
	}

	sb.WriteString(strings.Repeat("─", 60))
	sb.WriteString("\n")

	if m.sessionID != "" {
		sb.WriteString(fmt.Sprintf("Session: %s\n", m.sessionID))
		sb.WriteString(fmt.Sprintf("To review: ayo sessions show %s\n", m.sessionID))
	}

	return sb.String()
}

// ScrollbackContent returns the content to dump to scrollback on exit.
func (m Model) ScrollbackContent() string {
	return m.scrollbackContent
}

// Run starts the chat TUI.
func Run(ctx context.Context, ag agent.Agent, sessionID string, sendFn SendMessageFunc) (Result, string, error) {
	model := New(ag, sessionID, sendFn)
	model.ctx = ctx

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return ResultError, "", err
	}

	m := finalModel.(Model)
	return ResultQuit, m.ScrollbackContent(), nil
}

// RunWithProgram starts the chat TUI and returns the program for external stream handlers.
// This allows setting up a TUIStreamHandler before running the program.
func RunWithProgram(ctx context.Context, ag agent.Agent, sessionID string, sendFn SendMessageFunc) (*tea.Program, Model) {
	model := New(ag, sessionID, sendFn)
	model.ctx = ctx

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	return p, model
}
