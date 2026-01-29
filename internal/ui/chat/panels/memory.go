package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// MemoryItem represents a single memory in the memory panel.
type MemoryItem struct {
	ID       string
	Content  string
	Category string // "preference", "fact", "correction", "pattern"
	Scope    string // "global", agent handle, or path
}

// MemoryPanel displays relevant memories.
type MemoryPanel struct {
	viewport viewport.Model
	memories []MemoryItem
	width    int
	height   int
	visible  bool
	focused  bool
}

// NewMemoryPanel creates a new memory panel.
func NewMemoryPanel() *MemoryPanel {
	return &MemoryPanel{
		memories: make([]MemoryItem, 0),
		visible:  false,
	}
}

// SetSize updates the panel dimensions.
func (p *MemoryPanel) SetSize(width, height int) {
	p.width = width
	p.height = height

	if p.viewport.Width == 0 {
		p.viewport = viewport.New(width-4, height-4)
		p.viewport.MouseWheelEnabled = true
	} else {
		p.viewport.Width = width - 4
		p.viewport.Height = height - 4
	}

	p.updateContent()
}

// SetMemories updates the memory list.
func (p *MemoryPanel) SetMemories(memories []MemoryItem) {
	p.memories = memories
	p.updateContent()
}

// Toggle shows or hides the panel.
func (p *MemoryPanel) Toggle() {
	p.visible = !p.visible
}

// Show makes the panel visible.
func (p *MemoryPanel) Show() {
	p.visible = true
}

// Hide hides the panel.
func (p *MemoryPanel) Hide() {
	p.visible = false
}

// IsVisible returns whether the panel is visible.
func (p *MemoryPanel) IsVisible() bool {
	return p.visible
}

// Focus gives focus to the panel.
func (p *MemoryPanel) Focus() {
	p.focused = true
}

// Blur removes focus from the panel.
func (p *MemoryPanel) Blur() {
	p.focused = false
}

// IsFocused returns whether the panel is focused.
func (p *MemoryPanel) IsFocused() bool {
	return p.focused
}

// View renders the panel.
func (p *MemoryPanel) View() string {
	if !p.visible || p.width <= 0 || p.height <= 0 {
		return ""
	}

	// Styles
	borderColor := lipgloss.Color("#6b7280")
	if p.focused {
		borderColor = lipgloss.Color("#a78bfa")
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#a78bfa"))

	containerStyle := lipgloss.NewStyle().
		Width(p.width).
		Height(p.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	// Header
	title := titleStyle.Render("Memory")
	count := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		Render(fmt.Sprintf("%d", len(p.memories)))
	header := lipgloss.JoinHorizontal(lipgloss.Left, title, "  ", count)

	// Content
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		p.viewport.View(),
	)

	return containerStyle.Render(content)
}

// updateContent rebuilds the viewport content from memories.
func (p *MemoryPanel) updateContent() {
	if len(p.memories) == 0 {
		p.viewport.SetContent(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")).
			Render("No relevant memories"))
		return
	}

	var lines []string
	for _, mem := range p.memories {
		line := p.renderMemoryItem(mem)
		lines = append(lines, line)
		lines = append(lines, "") // Add spacing between items
	}

	p.viewport.SetContent(strings.Join(lines, "\n"))
}

// renderMemoryItem renders a single memory item.
func (p *MemoryPanel) renderMemoryItem(mem MemoryItem) string {
	// Category icon and color
	var icon string
	var iconColor lipgloss.Color

	switch mem.Category {
	case "preference":
		icon = "★"
		iconColor = lipgloss.Color("#f59e0b") // amber
	case "fact":
		icon = "◆"
		iconColor = lipgloss.Color("#3b82f6") // blue
	case "correction":
		icon = "!"
		iconColor = lipgloss.Color("#ef4444") // red
	case "pattern":
		icon = "~"
		iconColor = lipgloss.Color("#22c55e") // green
	default:
		icon = "·"
		iconColor = lipgloss.Color("#6b7280") // gray
	}

	iconStyled := lipgloss.NewStyle().Foreground(iconColor).Render(icon)

	// Content
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e5e7eb"))

	// Truncate content if needed
	content := mem.Content
	maxWidth := p.width - 8
	if len(content) > maxWidth && maxWidth > 3 {
		content = content[:maxWidth-3] + "..."
	}

	// Scope badge (if not global)
	scopeBadge := ""
	if mem.Scope != "" && mem.Scope != "global" {
		scopeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")).
			Italic(true)
		scopeBadge = scopeStyle.Render(" [" + mem.Scope + "]")
	}

	return fmt.Sprintf("%s %s%s", iconStyled, textStyle.Render(content), scopeBadge)
}

// ScrollUp scrolls the viewport up.
func (p *MemoryPanel) ScrollUp(n int) {
	p.viewport.LineUp(n)
}

// ScrollDown scrolls the viewport down.
func (p *MemoryPanel) ScrollDown(n int) {
	p.viewport.LineDown(n)
}

// MemoriesUpdateMsg is sent when the memory list changes.
type MemoriesUpdateMsg struct {
	Memories []MemoryItem
}
