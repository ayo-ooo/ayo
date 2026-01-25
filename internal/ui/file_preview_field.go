package ui

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Shared markdown renderer with syntax highlighting for file preview
var (
	previewRendererOnce sync.Once
	previewRenderer     *glamour.TermRenderer
)

// getPreviewRenderer returns a shared glamour renderer with full syntax highlighting.
func getPreviewRenderer() *glamour.TermRenderer {
	previewRendererOnce.Do(func() {
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
		)
		if err == nil {
			previewRenderer = r
		}
	})
	return previewRenderer
}

// FilePreviewField is a custom huh field that displays file contents in a scrollable viewport.
type FilePreviewField struct {
	filePath       *string
	viewport       viewport.Model
	title          string
	terminalHeight int // Track terminal height for relative sizing
	maxHeightRatio float64

	// Caching for performance
	cachedPath  string // Path used for cached content
	cachedWidth int    // Width used for cached content

	focused bool
	width   int
	height  int
	theme   *huh.Theme
	keymap  filePreviewKeyMap
}

type filePreviewKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Next     key.Binding
	Prev     key.Binding
}

// NewFilePreviewField creates a new file preview field.
func NewFilePreviewField() *FilePreviewField {
	vp := viewport.New(80, 10)
	vp.MouseWheelEnabled = true

	return &FilePreviewField{
		viewport:       vp,
		title:          "File Preview",
		height:         15,
		maxHeightRatio: 0.6, // Default to 60% of terminal height
		keymap: filePreviewKeyMap{
			Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
			Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
			PageUp:   key.NewBinding(key.WithKeys("pgup", "ctrl+u"), key.WithHelp("pgup", "page up")),
			PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+d"), key.WithHelp("pgdn", "page down")),
			Next:     key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("enter", "continue")),
			Prev:     key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "back")),
		},
	}
}

// FilePath sets the pointer to the file path to preview.
func (f *FilePreviewField) FilePath(path *string) *FilePreviewField {
	f.filePath = path
	return f
}

// Title sets the title of the preview.
func (f *FilePreviewField) Title(title string) *FilePreviewField {
	f.title = title
	return f
}

// Height sets the height of the viewport.
func (f *FilePreviewField) Height(height int) *FilePreviewField {
	f.height = height
	// Update viewport height immediately (accounting for chrome)
	viewportHeight := height - 4 // title, border top/bottom, scroll indicator
	if viewportHeight < 3 {
		viewportHeight = 3
	}
	f.viewport.Height = viewportHeight
	return f
}

// MaxHeightRatio sets the maximum height as a ratio of terminal height (0.0-1.0).
func (f *FilePreviewField) MaxHeightRatio(ratio float64) *FilePreviewField {
	f.maxHeightRatio = ratio
	return f
}

// isMarkdownFile checks if the file is a markdown file.
func isMarkdownFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

// loadContent reads and sets the viewport content, rendering markdown if applicable.
func (f *FilePreviewField) loadContent() {
	if f.filePath == nil || *f.filePath == "" {
		f.viewport.SetContent("(no file selected)")
		f.cachedPath = ""
		return
	}

	path := *f.filePath
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = strings.Replace(path, "~", home, 1)
		}
	}

	// Skip if content already cached for this path
	if path == f.cachedPath {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		f.viewport.SetContent("Error: " + err.Error())
		f.cachedPath = ""
		return
	}

	content := string(data)

	// Render markdown files with syntax highlighting
	if isMarkdownFile(path) {
		if r := getPreviewRenderer(); r != nil {
			rendered, err := r.Render(content)
			if err == nil {
				content = strings.TrimSpace(rendered)
			}
		}
	}

	f.viewport.SetContent(content)
	f.viewport.GotoTop()
	f.cachedPath = path
}

// Init initializes the field.
func (f *FilePreviewField) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (f *FilePreviewField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case contentLoadedMsg:
		// Content was loaded, just trigger a re-render
		return f, nil
	case tea.WindowSizeMsg:
		f.terminalHeight = msg.Height
		f.width = msg.Width
		f.viewport.Width = msg.Width - 4
		f.updateViewportSize()
		// Reload content with new dimensions if focused
		if f.focused {
			f.loadContent()
		}
	}

	if !f.focused {
		return f, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keymap.Next):
			return f, huh.NextField
		case key.Matches(msg, f.keymap.Prev):
			return f, huh.PrevField
		}
	}

	var cmd tea.Cmd
	f.viewport, cmd = f.viewport.Update(msg)
	return f, cmd
}

// updateViewportSize recalculates viewport size based on available height.
func (f *FilePreviewField) updateViewportSize() {
	// Start with configured height
	maxHeight := f.height

	// Clamp to terminal height ratio if terminal height is known
	if f.terminalHeight > 0 && f.maxHeightRatio > 0 {
		calculatedMax := int(float64(f.terminalHeight) * f.maxHeightRatio)
		if calculatedMax < maxHeight {
			maxHeight = calculatedMax
		}
	}

	// Account for title (1 line), border (2 lines), scroll indicator (1 line)
	// Border: top + bottom = 2 lines
	viewportHeight := maxHeight - 4
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	f.viewport.Height = viewportHeight
}

// View renders the field.
func (f *FilePreviewField) View() string {
	// Reload content when viewed while focused to ensure it's up to date
	if f.focused && f.viewport.TotalLineCount() == 0 {
		f.loadContent()
	}

	var sb strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	if f.theme != nil {
		titleStyle = f.theme.Focused.Title
	}
	sb.WriteString(titleStyle.Render(f.title))
	sb.WriteString("\n")

	// Viewport with border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	if f.focused {
		borderStyle = borderStyle.BorderForeground(lipgloss.Color("212"))
	}

	// Clamp viewport content to its height to prevent overflow
	viewportContent := f.viewport.View()
	sb.WriteString(borderStyle.Render(viewportContent))

	// Scroll indicator - always show for consistency
	scrollInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
		Render(strings.Repeat(" ", 2) + "↑/↓ scroll • enter to continue")
	sb.WriteString("\n")
	sb.WriteString(scrollInfo)

	return sb.String()
}

// contentLoadedMsg signals that content has been loaded.
type contentLoadedMsg struct{}

// Focus focuses the field.
func (f *FilePreviewField) Focus() tea.Cmd {
	f.focused = true
	f.updateViewportSize()
	f.loadContent()
	// Return a command to trigger immediate re-render
	return func() tea.Msg {
		return contentLoadedMsg{}
	}
}

// Blur blurs the field.
func (f *FilePreviewField) Blur() tea.Cmd {
	f.focused = false
	return nil
}

// Error returns nil (no validation).
func (f *FilePreviewField) Error() error {
	return nil
}

// Run runs the field standalone.
func (f *FilePreviewField) Run() error {
	return huh.Run(f)
}

// RunAccessible runs in accessible mode.
func (f *FilePreviewField) RunAccessible(w io.Writer, r io.Reader) error {
	// Just print the content in accessible mode
	if f.filePath != nil && *f.filePath != "" {
		data, _ := os.ReadFile(*f.filePath)
		_, _ = w.Write(data)
	}
	return nil
}

// Skip returns false - this field should not be skipped.
func (f *FilePreviewField) Skip() bool {
	return false
}

// Zoom returns false - let the group manage height distribution.
// Returning true causes layout conflicts with the form's viewport.
func (f *FilePreviewField) Zoom() bool {
	return false
}

// KeyBinds returns the keybindings.
func (f *FilePreviewField) KeyBinds() []key.Binding {
	return []key.Binding{f.keymap.Up, f.keymap.Down, f.keymap.Next, f.keymap.Prev}
}

// WithTheme sets the theme.
func (f *FilePreviewField) WithTheme(theme *huh.Theme) huh.Field {
	f.theme = theme
	return f
}

// WithAccessible is deprecated but required.
func (f *FilePreviewField) WithAccessible(accessible bool) huh.Field {
	return f
}

// WithKeyMap sets the keymap.
func (f *FilePreviewField) WithKeyMap(k *huh.KeyMap) huh.Field {
	// Use huh's keybindings for navigation
	if k != nil {
		f.keymap.Next = k.Note.Next
		f.keymap.Prev = k.Note.Prev
	}
	return f
}

// WithWidth sets the width.
func (f *FilePreviewField) WithWidth(width int) huh.Field {
	f.width = width
	// Account for border (2 chars) and padding (2 chars)
	viewportWidth := width - 4
	if viewportWidth < 20 {
		viewportWidth = 20
	}
	f.viewport.Width = viewportWidth
	return f
}

// WithHeight sets the height allocated by the form.
func (f *FilePreviewField) WithHeight(height int) huh.Field {
	f.height = height
	// Recalculate viewport based on new height
	viewportHeight := height - 4 // title, border top/bottom, scroll indicator
	if viewportHeight < 3 {
		viewportHeight = 3
	}
	f.viewport.Height = viewportHeight
	return f
}

// WithPosition sets position info.
func (f *FilePreviewField) WithPosition(p huh.FieldPosition) huh.Field {
	f.keymap.Prev.SetEnabled(!p.IsFirst())
	f.keymap.Next.SetEnabled(!p.IsLast())
	return f
}

// GetKey returns empty string (no key).
func (f *FilePreviewField) GetKey() string {
	return ""
}

// GetValue returns nil (no value).
func (f *FilePreviewField) GetValue() any {
	return nil
}
