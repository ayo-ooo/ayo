package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// MemoryFormationDisplay shows memory formation progress.
type MemoryFormationDisplay struct {
	mu      sync.Mutex
	pending map[string]string // id -> content preview
	isTTY   bool
}

// NewMemoryFormationDisplay creates a new memory formation display.
func NewMemoryFormationDisplay() *MemoryFormationDisplay {
	return &MemoryFormationDisplay{
		pending: make(map[string]string),
		isTTY:   term.IsTerminal(int(os.Stderr.Fd())),
	}
}

// StartFormation shows that a memory is being formed.
func (d *MemoryFormationDisplay) StartFormation(id, content string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.pending[id] = content

	if !d.isTTY {
		return
	}

	// Style
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(0, 1)

	spinner := lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render("◐")
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Memory")
	text := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Italic(true).Render(truncateMemory(content, 50))

	line := fmt.Sprintf("%s %s %s", spinner, label, text)
	box := style.Render(line)

	fmt.Fprintln(os.Stderr, box)
}

// CompleteFormation shows that a memory was successfully formed.
func (d *MemoryFormationDisplay) CompleteFormation(id, content string, elapsed time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.pending, id)

	if !d.isTTY {
		return
	}

	// Style
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("76")).
		Padding(0, 1)

	check := lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Render("✓")
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Remembered:")
	text := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Italic(true).Render(truncateMemory(content, 45))
	duration := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("(%s)", elapsed.Round(time.Millisecond)))

	line := fmt.Sprintf("%s %s %s %s", check, label, text, duration)
	box := style.Render(line)

	fmt.Fprintln(os.Stderr, box)
}

// FailFormation shows that memory formation failed.
func (d *MemoryFormationDisplay) FailFormation(id string, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.pending, id)

	if !d.isTTY {
		return
	}

	// Style
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(0, 1)

	x := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("×")
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Memory failed:")
	errText := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(truncateMemory(err.Error(), 40))

	line := fmt.Sprintf("%s %s %s", x, label, errText)
	box := style.Render(line)

	fmt.Fprintln(os.Stderr, box)
}

// ShowRetrievedMemories displays retrieved memories context.
func ShowRetrievedMemories(count int) {
	if count == 0 {
		return
	}

	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
	text := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	var label string
	if count == 1 {
		label = "1 memory recalled"
	} else {
		label = fmt.Sprintf("%d memories recalled", count)
	}

	line := fmt.Sprintf("%s %s", style.Render("◆"), text.Render(label))
	fmt.Fprintln(os.Stderr, line)
}

func truncateMemory(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
