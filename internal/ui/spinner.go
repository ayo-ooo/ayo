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

// SpinnerType identifies different kinds of async operations
type SpinnerType int

const (
	// SpinnerAgent is for LLM thinking/generation (default)
	SpinnerAgent SpinnerType = iota
	// SpinnerTool is for tool execution
	SpinnerTool
	// SpinnerMemory is for memory operations (embed, store, search)
	SpinnerMemory
	// SpinnerReasoning is for extended thinking/reasoning
	SpinnerReasoning
	// SpinnerSystem is for system operations (setup, install)
	SpinnerSystem
)

// Spinner animation frames by type
var (
	// Default braille spinner for agent thinking
	framesAgent = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	// Dots for tool execution
	framesTool = []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●"}
	// Pulse for memory operations
	framesMemory = []string{"◇", "◈", "◆", "◈"}
	// Meter for reasoning
	framesReasoning = []string{"▱▱▱", "▰▱▱", "▰▰▱", "▰▰▰", "▰▰▱", "▰▱▱"}
	// Line for system ops
	framesSystem = []string{"|", "/", "-", "\\"}
)

// SpinnerFrames defines the animation frames for the spinner (default/agent)
var SpinnerFrames = framesAgent

// Spinner displays an animated spinner with a message
type Spinner struct {
	message   string
	indent    string // Indentation prefix for nested spinners
	startTime time.Time
	done      chan struct{}
	closeOnce sync.Once
	mu        sync.Mutex
	frame     int
	frames    []string // Animation frames for this spinner
	style     lipgloss.Style
	msgStyle  lipgloss.Style
	isTTY     bool
	quiet     bool
}

// NewSpinner creates a new spinner with the given message (agent type)
func NewSpinner(message string) *Spinner {
	return NewSpinnerWithType(message, SpinnerAgent)
}

// NewSpinnerWithType creates a spinner with a specific type for different operations
func NewSpinnerWithType(message string, spinnerType SpinnerType) *Spinner {
	frames, style, msgStyle := spinnerConfig(spinnerType)
	return &Spinner{
		message:   message,
		indent:    "",
		startTime: time.Now(),
		done:      make(chan struct{}),
		frames:    frames,
		style:     style,
		msgStyle:  msgStyle,
		isTTY:     term.IsTerminal(int(os.Stderr.Fd())),
	}
}

// spinnerConfig returns frames and styles for a spinner type
func spinnerConfig(t SpinnerType) ([]string, lipgloss.Style, lipgloss.Style) {
	switch t {
	case SpinnerTool:
		return framesTool,
			lipgloss.NewStyle().Foreground(colorTertiary),
			lipgloss.NewStyle().Foreground(colorTextDim)
	case SpinnerMemory:
		return framesMemory,
			lipgloss.NewStyle().Foreground(colorSecondary),
			lipgloss.NewStyle().Foreground(colorTextDim)
	case SpinnerReasoning:
		return framesReasoning,
			lipgloss.NewStyle().Foreground(colorMuted),
			lipgloss.NewStyle().Foreground(colorTextDim)
	case SpinnerSystem:
		return framesSystem,
			lipgloss.NewStyle().Foreground(colorSuccess),
			lipgloss.NewStyle().Foreground(colorTextDim)
	default: // SpinnerAgent
		return framesAgent,
			lipgloss.NewStyle().Foreground(colorPrimary),
			lipgloss.NewStyle().Foreground(colorTextDim)
	}
}

// NewSpinnerWithDepth creates a spinner at a specific nesting depth.
// Depth 0 is top-level (no indent), depth 1+ adds visual indentation.
func NewSpinnerWithDepth(message string, depth int) *Spinner {
	return NewSpinnerWithTypeAndDepth(message, SpinnerAgent, depth)
}

// NewSpinnerWithTypeAndDepth creates a typed spinner with nesting depth.
func NewSpinnerWithTypeAndDepth(message string, spinnerType SpinnerType, depth int) *Spinner {
	frames, style, msgStyle := spinnerConfig(spinnerType)
	indent := ""

	if depth > 0 {
		// Use vertical line indicator and muted colors for nested spinners
		indent = strings.Repeat("  ", depth) + "│ "
		style = lipgloss.NewStyle().Foreground(colorSecondary)
		msgStyle = lipgloss.NewStyle().Foreground(colorMuted)
	}

	return &Spinner{
		message:   message,
		indent:    indent,
		startTime: time.Now(),
		done:      make(chan struct{}),
		frames:    frames,
		style:     style,
		msgStyle:  msgStyle,
		isTTY:     term.IsTerminal(int(os.Stderr.Fd())),
	}
}

// NewQuietSpinner creates a spinner that produces no output
func NewQuietSpinner() *Spinner {
	return &Spinner{
		startTime: time.Now(),
		done:      make(chan struct{}),
		frames:    framesAgent,
		quiet:     true,
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	if s.quiet {
		return
	}
	if !s.isTTY {
		fmt.Fprintf(os.Stderr, "%s%s %s\n", s.indent, s.frames[0], s.message)
		return
	}

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.mu.Lock()
				s.frame = (s.frame + 1) % len(s.frames)
				s.render()
				s.mu.Unlock()
			}
		}
	}()

	// Initial render
	s.mu.Lock()
	s.render()
	s.mu.Unlock()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.closeOnce.Do(func() { close(s.done) })

	if s.quiet {
		return
	}

	if s.isTTY {
		fmt.Fprint(os.Stderr, "\r\033[K") // Clear line
	}
}

// StopWithMessage stops the spinner and shows a final message with elapsed time
func (s *Spinner) StopWithMessage(message string) {
	s.closeOnce.Do(func() { close(s.done) })

	if s.quiet {
		return
	}

	elapsed := time.Since(s.startTime)
	elapsedStr := formatDuration(elapsed)

	if s.isTTY {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}

	// Show "Thought for Xs" style message
	msgStyle := lipgloss.NewStyle().Foreground(colorTextDim)
	fmt.Fprintf(os.Stderr, "%s%s %s\n", s.indent, msgStyle.Render(message), msgStyle.Render("("+elapsedStr+")"))
}

// StopWithError stops the spinner and shows an error indicator
func (s *Spinner) StopWithError(message string) {
	s.closeOnce.Do(func() { close(s.done) })

	if s.quiet {
		return
	}

	if s.isTTY {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}

	errorStyle := lipgloss.NewStyle().Foreground(colorError)
	fmt.Fprintf(os.Stderr, "%s%s %s\n", s.indent, errorStyle.Render(IconError), message)
}

// Elapsed returns the time elapsed since the spinner started
func (s *Spinner) Elapsed() time.Duration {
	return time.Since(s.startTime)
}

// render draws the current spinner state (must be called with lock held)
func (s *Spinner) render() {
	if s.quiet {
		return
	}
	frame := s.style.Render(s.frames[s.frame%len(s.frames)])
	msg := s.msgStyle.Render(s.message)
	fmt.Fprintf(os.Stderr, "\r\033[K%s%s %s", s.indent, frame, msg)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
