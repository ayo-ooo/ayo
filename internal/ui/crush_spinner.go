package ui

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// CrushSpinner displays an animated spinner inspired by Crush's loading indicator.
// Features scrambling hex characters with a gradient color effect.
type CrushSpinner struct {
	label     string
	indent    string
	startTime time.Time
	done      chan struct{}
	closeOnce sync.Once
	mu        sync.Mutex
	step      int
	isTTY     bool

	// Animation settings
	numChars int      // Number of scrambling characters
	chars    []rune   // Current character for each position
	colors   []string // Pre-computed gradient colors
}

// Scrambling character set (similar to Crush)
var scrambleChars = []rune("0123456789abcdefABCDEF~!@#$%^&*()+=_")

// NewCrushSpinner creates a new Crush-style spinner with the given label.
func NewCrushSpinner(label string) *CrushSpinner {
	return NewCrushSpinnerWithDepth(label, 0)
}

// NewCrushSpinnerWithDepth creates a Crush-style spinner at a specific nesting depth.
func NewCrushSpinnerWithDepth(label string, depth int) *CrushSpinner {
	indent := ""
	if depth > 0 {
		indent = strings.Repeat("  ", depth) + "| "
	}

	numChars := 10
	colors := makeGradient(numChars)

	// Initialize random characters
	chars := make([]rune, numChars)
	for i := range chars {
		chars[i] = scrambleChars[rand.Intn(len(scrambleChars))]
	}

	return &CrushSpinner{
		label:     label,
		indent:    indent,
		startTime: time.Now(),
		done:      make(chan struct{}),
		isTTY:     term.IsTerminal(int(os.Stderr.Fd())),
		numChars:  numChars,
		chars:     chars,
		colors:    colors,
	}
}

// makeGradient creates a gradient from magenta to cyan (Crush-like colors)
func makeGradient(size int) []string {
	colors := make([]string, size)
	for i := range colors {
		// Gradient from magenta (#FF00FF) to cyan (#00FFFF)
		t := float64(i) / float64(size-1)
		r := int(255 * (1 - t))
		g := int(255 * t)
		b := 255
		colors[i] = fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}
	return colors
}

// Start begins the spinner animation
func (s *CrushSpinner) Start() {
	if !s.isTTY {
		// Non-TTY: just print label and return
		fmt.Fprintf(os.Stderr, "%s%s...\n", s.indent, s.label)
		return
	}

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond) // 20 FPS like Crush
		defer ticker.Stop()

		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.mu.Lock()
				s.step++
				// Scramble characters
				for i := range s.chars {
					s.chars[i] = scrambleChars[rand.Intn(len(scrambleChars))]
				}
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
func (s *CrushSpinner) Stop() {
	s.closeOnce.Do(func() { close(s.done) })

	if s.isTTY {
		fmt.Fprint(os.Stderr, "\r\033[K") // Clear line
	}
}

// StopWithMessage stops the spinner and shows a final message
func (s *CrushSpinner) StopWithMessage(message string) {
	s.closeOnce.Do(func() { close(s.done) })

	elapsed := time.Since(s.startTime)
	elapsedStr := formatDuration(elapsed)

	if s.isTTY {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}

	msgStyle := lipgloss.NewStyle().Foreground(colorTextDim)
	fmt.Fprintf(os.Stderr, "%s%s %s\n", s.indent, msgStyle.Render(message), msgStyle.Render("("+elapsedStr+")"))
}

// StopWithError stops the spinner and shows an error indicator
func (s *CrushSpinner) StopWithError(message string) {
	s.closeOnce.Do(func() { close(s.done) })

	if s.isTTY {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}

	errorStyle := lipgloss.NewStyle().Foreground(colorError)
	fmt.Fprintf(os.Stderr, "%s%s %s\n", s.indent, errorStyle.Render(IconError), message)
}

// render draws the current spinner state (must be called with lock held)
func (s *CrushSpinner) render() {
	var b strings.Builder

	// Render scrambling characters with gradient colors
	for i, ch := range s.chars {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(s.colors[i]))
		b.WriteString(style.Render(string(ch)))
	}

	// Add label with muted color
	if s.label != "" {
		labelStyle := lipgloss.NewStyle().Foreground(colorMuted)
		b.WriteString(" ")
		b.WriteString(labelStyle.Render(s.label))

		// Animated ellipsis
		dots := (s.step / 8) % 4 // Change every 8 frames
		b.WriteString(labelStyle.Render(strings.Repeat(".", dots)))
	}

	fmt.Fprintf(os.Stderr, "\r\033[K%s%s", s.indent, b.String())
}
