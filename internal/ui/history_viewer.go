// Package ui provides terminal user interface components.
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexcabrera/ayo/internal/session"
)

// HistoryViewerResult indicates the outcome of the history viewer.
type HistoryViewerResult int

const (
	// HistoryViewerQuit means the user chose to quit without continuing.
	HistoryViewerQuit HistoryViewerResult = iota
	// HistoryViewerContinue means the user wants to continue the session.
	HistoryViewerContinue
)

// historyKeyMap defines keybindings for the history viewer.
type historyKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Continue key.Binding
	Quit     key.Binding
}

func defaultHistoryKeyMap() historyKeyMap {
	return historyKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "f", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
		Continue: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "continue"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// HistoryViewer is a bubbletea model for viewing session history.
type HistoryViewer struct {
	viewport     viewport.Model
	keyMap       historyKeyMap
	agentHandle  string
	sessionTitle string
	messageCount int
	ready        bool
	result       HistoryViewerResult
	width        int
	height       int
	content      string
}

// NewHistoryViewer creates a new history viewer for the given session messages.
func NewHistoryViewer(messages []session.Message, agentHandle, sessionTitle string) HistoryViewer {
	content := RenderHistory(messages, agentHandle)

	return HistoryViewer{
		keyMap:       defaultHistoryKeyMap(),
		agentHandle:  agentHandle,
		sessionTitle: sessionTitle,
		messageCount: len(messages),
		content:      content,
		result:       HistoryViewerQuit, // Default to quit if closed unexpectedly
	}
}

// Result returns the outcome of the history viewer.
func (m HistoryViewer) Result() HistoryViewerResult {
	return m.result
}

// Init implements tea.Model.
func (m HistoryViewer) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m HistoryViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			m.result = HistoryViewerQuit
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Continue):
			m.result = HistoryViewerContinue
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Top):
			m.viewport.GotoTop()

		case key.Matches(msg, m.keyMap.Bottom):
			m.viewport.GotoBottom()
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMargin := headerHeight + footerHeight

		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMargin)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.content)
			m.viewport.MouseWheelEnabled = true
			m.viewport.MouseWheelDelta = 3
			// Start at bottom to show most recent messages
			m.viewport.GotoBottom()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMargin
		}
	}

	// Handle viewport scrolling
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m HistoryViewer) View() string {
	if !m.ready {
		return "\n  Loading history..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

// headerView renders the header with session info.
func (m HistoryViewer) headerView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	infoStyle := lipgloss.NewStyle().
		Foreground(colorMuted)

	lineStyle := lipgloss.NewStyle().
		Foreground(colorSubtle)

	title := titleStyle.Render(fmt.Sprintf("Session History: %s", m.agentHandle))
	info := infoStyle.Render(fmt.Sprintf(" (%d messages)", m.messageCount))

	// Fill remaining width with line
	contentWidth := lipgloss.Width(title) + lipgloss.Width(info)
	lineWidth := m.viewport.Width - contentWidth
	if lineWidth < 0 {
		lineWidth = 0
	}
	line := lineStyle.Render(strings.Repeat("─", lineWidth))

	return lipgloss.JoinHorizontal(lipgloss.Center, title, info, line)
}

// footerView renders the footer with scroll position and help.
func (m HistoryViewer) footerView() string {
	lineStyle := lipgloss.NewStyle().
		Foreground(colorSubtle)

	percentStyle := lipgloss.NewStyle().
		Foreground(colorMuted)

	helpStyle := lipgloss.NewStyle().
		Foreground(colorMuted)

	// Scroll percentage
	percent := percentStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))

	// Help text
	help := helpStyle.Render("↑/↓ scroll · Enter continue · q quit")

	// Calculate line width
	contentWidth := lipgloss.Width(percent) + lipgloss.Width(help) + 4
	lineWidth := m.viewport.Width - contentWidth
	if lineWidth < 0 {
		lineWidth = 0
	}
	line := lineStyle.Render(strings.Repeat("─", lineWidth))

	return lipgloss.JoinHorizontal(lipgloss.Center, percent, " ", line, " ", help)
}

// RunHistoryViewer displays the history viewer and returns the result.
func RunHistoryViewer(messages []session.Message, agentHandle, sessionTitle string) (HistoryViewerResult, error) {
	viewer := NewHistoryViewer(messages, agentHandle, sessionTitle)

	p := tea.NewProgram(
		viewer,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return HistoryViewerQuit, err
	}

	return finalModel.(HistoryViewer).Result(), nil
}
