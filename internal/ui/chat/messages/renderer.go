package messages

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alexcabrera/ayo/internal/ui/shared"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/charmbracelet/x/ansi"
)

// renderer defines the interface for tool-specific rendering implementations.
type renderer interface {
	// Render returns the styled tool call view.
	Render(t *toolCallCmp) string
}

// rendererFactory creates new renderer instances.
type rendererFactory func() renderer

// renderRegistry manages the mapping of tool names to their renderers.
type renderRegistry map[string]rendererFactory

// register adds a new renderer factory to the registry.
func (rr renderRegistry) register(name string, f rendererFactory) {
	rr[name] = f
}

// lookup retrieves a renderer for the given tool name, falling back to generic.
func (rr renderRegistry) lookup(name string) renderer {
	if f, ok := rr[name]; ok {
		return f()
	}
	return genericRenderer{}
}

// registry holds all registered tool renderers.
var registry = renderRegistry{}

// roundedEnumerator creates a tree enumerator with rounded corners.
// lPadding is the left padding before the branch character.
// width is the number of dash characters in the branch line.
func roundedEnumerator(lPadding, width int) tree.Enumerator {
	if width == 0 {
		width = 2
	}
	if lPadding == 0 {
		lPadding = 1
	}
	return func(children tree.Children, index int) string {
		line := strings.Repeat("\u2500", width) // ─
		padding := strings.Repeat(" ", lPadding)
		if children.Length()-1 == index {
			return padding + "\u2570" + line // ╰ for last child
		}
		return padding + "\u251c" + line // ├
	}
}

// Status icons for tool states - re-exported from shared package.
const (
	ToolPending = shared.ToolPending
	ToolSuccess = shared.ToolSuccess
	ToolError   = shared.ToolError
	ToolRunning = shared.ToolRunning
)

// baseRenderer provides common functionality for all tool renderers.
type baseRenderer struct{}

// unmarshalParams safely unmarshals JSON parameters.
func (br baseRenderer) unmarshalParams(input string, target any) error {
	return json.Unmarshal([]byte(input), target)
}

// makeHeader builds the tool call header with status icon and parameters.
func (br baseRenderer) makeHeader(t *toolCallCmp, toolName string, width int, params ...string) string {
	if t.isNested {
		return br.makeNestedHeader(t, toolName, width, params...)
	}

	icon := lipgloss.NewStyle().Foreground(shared.ColorToolPending).Render(ToolPending)
	if t.result.ToolCallID != "" {
		if t.result.IsError {
			icon = lipgloss.NewStyle().Foreground(shared.ColorError).Render(ToolError)
		} else {
			icon = lipgloss.NewStyle().Foreground(shared.ColorSuccess).Render(ToolSuccess)
		}
	} else if t.cancelled {
		icon = lipgloss.NewStyle().Foreground(shared.ColorToolPending).Render(ToolPending)
	} else if t.spinning {
		icon = lipgloss.NewStyle().Foreground(shared.ColorToolRunning).Render(ToolRunning)
	}

	toolStyle := lipgloss.NewStyle().Foreground(shared.ColorToolName).Bold(true)
	prefix := fmt.Sprintf("%s %s ", icon, toolStyle.Render(toolName))

	return prefix + renderParamList(false, width-lipgloss.Width(prefix), params...)
}

// makeNestedHeader builds header for nested tool calls.
func (br baseRenderer) makeNestedHeader(t *toolCallCmp, toolName string, width int, params ...string) string {
	icon := lipgloss.NewStyle().Foreground(shared.ColorToolPending).Render(ToolPending)
	if t.result.ToolCallID != "" {
		if t.result.IsError {
			icon = lipgloss.NewStyle().Foreground(shared.ColorError).Render(ToolError)
		} else {
			icon = lipgloss.NewStyle().Foreground(shared.ColorSuccess).Render(ToolSuccess)
		}
	} else if t.cancelled {
		icon = lipgloss.NewStyle().Foreground(shared.ColorToolPending).Render(ToolPending)
	}

	toolStyle := lipgloss.NewStyle().Foreground(shared.ColorTextDim)
	prefix := fmt.Sprintf("%s %s ", icon, toolStyle.Render(toolName))

	return prefix + renderParamList(true, width-lipgloss.Width(prefix), params...)
}

// renderWithParams provides a common rendering pattern.
func (br baseRenderer) renderWithParams(t *toolCallCmp, toolName string, args []string, contentRenderer func() string) string {
	width := t.textWidth()
	if t.isNested {
		width -= 4
	}

	header := br.makeHeader(t, toolName, width, args...)

	if t.isNested {
		return header
	}

	if res, done := earlyState(header, t); done {
		return res
	}

	body := contentRenderer()
	return joinHeaderBody(header, body)
}

// renderError provides consistent error rendering.
func (br baseRenderer) renderError(t *toolCallCmp, message string) string {
	header := br.makeHeader(t, prettifyToolName(t.call.Name), t.textWidth())
	errorStyle := lipgloss.NewStyle().
		Background(shared.ColorError).
		Foreground(shared.ColorTextBright).
		Padding(0, 1)
	errorTag := errorStyle.Render("ERROR")
	msgStyle := lipgloss.NewStyle().Foreground(shared.ColorTextDim)
	return joinHeaderBody(header, errorTag+" "+msgStyle.Render(message))
}

// paramBuilder helps construct parameter lists for tool headers.
type paramBuilder struct {
	args []string
}

// newParamBuilder creates a new parameter builder.
func newParamBuilder() *paramBuilder {
	return &paramBuilder{args: make([]string, 0)}
}

// addMain adds the main parameter.
func (pb *paramBuilder) addMain(value string) *paramBuilder {
	if value != "" {
		pb.args = append(pb.args, value)
	}
	return pb
}

// addKeyValue adds a key-value pair.
func (pb *paramBuilder) addKeyValue(key, value string) *paramBuilder {
	if value != "" {
		pb.args = append(pb.args, key, value)
	}
	return pb
}

// addFlag adds a boolean flag.
func (pb *paramBuilder) addFlag(key string, value bool) *paramBuilder {
	if value {
		pb.args = append(pb.args, key, "true")
	}
	return pb
}

// build returns the parameter list.
func (pb *paramBuilder) build() []string {
	return pb.args
}

// earlyState returns immediately-rendered error/cancelled/pending states.
func earlyState(header string, t *toolCallCmp) (string, bool) {
	style := lipgloss.NewStyle().Foreground(shared.ColorTextDim)
	var message string

	switch {
	case t.result.IsError:
		errorStyle := lipgloss.NewStyle().
			Background(shared.ColorError).
			Foreground(shared.ColorTextBright).
			Padding(0, 1)
		message = errorStyle.Render("ERROR") + " " + style.Render(truncateText(t.result.Content, t.textWidth()-10))
	case t.cancelled:
		message = style.Render("Cancelled.")
	case t.result.ToolCallID == "":
		if t.permissionRequested && !t.permissionGranted {
			message = style.Render("Requesting permission...")
		} else {
			return header, false // Still waiting, just show header
		}
	default:
		return "", false
	}

	indented := lipgloss.NewStyle().PaddingLeft(2).Render(message)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", indented), true
}

// joinHeaderBody joins header and body with proper spacing.
func joinHeaderBody(header, body string) string {
	if body == "" {
		return header
	}
	indented := lipgloss.NewStyle().PaddingLeft(2).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", indented)
}

// renderParamList renders parameter list.
func renderParamList(nested bool, maxWidth int, params ...string) string {
	if len(params) == 0 {
		return ""
	}

	style := lipgloss.NewStyle().Foreground(shared.ColorTextDim)
	mainParam := params[0]

	if maxWidth > 0 && lipgloss.Width(mainParam) > maxWidth {
		mainParam = ansi.Truncate(mainParam, maxWidth, "...")
	}

	if len(params) == 1 {
		return style.Render(mainParam)
	}

	// Build key=value pairs
	otherParams := params[1:]
	if len(otherParams)%2 != 0 {
		otherParams = append(otherParams, "")
	}

	parts := make([]string, 0, len(otherParams)/2)
	for i := 0; i < len(otherParams); i += 2 {
		key := otherParams[i]
		value := otherParams[i+1]
		if value == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	if len(parts) > 0 {
		mainParam = fmt.Sprintf("%s (%s)", mainParam, strings.Join(parts, ", "))
	}

	if maxWidth > 0 && lipgloss.Width(mainParam) > maxWidth {
		mainParam = ansi.Truncate(mainParam, maxWidth, "...")
	}

	return style.Render(mainParam)
}

// renderPlainContent renders plain text with truncation.
func renderPlainContent(t *toolCallCmp, content string, maxLines int) string {
	if maxLines <= 0 {
		maxLines = 10
	}

	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\t", "    ")
	content = strings.TrimSpace(content)

	lines := strings.Split(content, "\n")
	width := t.textWidth() - 2

	style := lipgloss.NewStyle().
		Foreground(shared.ColorTextDim).
		Background(shared.ColorBgDark)

	var out []string
	for i, ln := range lines {
		if i >= maxLines {
			break
		}
		if lipgloss.Width(ln) > width {
			ln = ansi.Truncate(ln, width, "...")
		}
		out = append(out, style.Width(width).Render(" "+ln))
	}

	if len(lines) > maxLines {
		out = append(out, style.Width(width).Render(
			fmt.Sprintf(" ... (%d lines)", len(lines)-maxLines)))
	}

	return strings.Join(out, "\n")
}

// renderMarkdownContent renders markdown with glamour.
func renderMarkdownContent(t *toolCallCmp, content string, maxLines int) string {
	if maxLines <= 0 {
		maxLines = 10
	}

	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.TrimSpace(content)

	width := t.textWidth() - 2
	if width > 120 {
		width = 120
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return renderPlainContent(t, content, maxLines)
	}

	rendered, err := r.Render(content)
	if err != nil {
		return renderPlainContent(t, content, maxLines)
	}

	lines := strings.Split(strings.TrimSpace(rendered), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, fmt.Sprintf("... (%d lines)", len(strings.Split(rendered, "\n"))-maxLines))
	}

	return strings.Join(lines, "\n")
}

// truncateHeight truncates content to max lines.
func truncateHeight(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines {
		return strings.Join(lines[:maxLines], "\n")
	}
	return s
}

// truncateText truncates text to max width.
func truncateText(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	return ansi.Truncate(s, maxWidth, "...")
}

// prettifyToolName returns a human-readable tool name.
func prettifyToolName(name string) string {
	switch name {
	case "bash":
		return "Bash"
	case "todo", "todos":
		return "Todo"
	case "agent":
		return "Agent"
	case "view":
		return "View"
	case "edit":
		return "Edit"
	case "write":
		return "Write"
	case "glob":
		return "Glob"
	case "grep":
		return "Grep"
	case "ls":
		return "List"
	case "fetch":
		return "Fetch"
	case "download":
		return "Download"
	case "sourcegraph":
		return "Sourcegraph"
	case "memory":
		return "Memory"
	default:
		return name
	}
}
