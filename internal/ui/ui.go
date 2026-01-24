package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"

	"github.com/alexcabrera/ayo/internal/pipe"
	"github.com/alexcabrera/ayo/internal/session"
)

type SelectAgentResult struct {
	Handle string
}

type UI struct {
	debug    bool
	depth    int       // 0 = top-level, 1+ = sub-agent calls
	styles   Styles
	renderer *markdownRenderer
	out      io.Writer // Where to write UI output (stdout or stderr)
	piped    bool      // Whether output is being piped
}

// markdownRenderer wraps glamour rendering with fallback.
type markdownRenderer struct{}

func (m *markdownRenderer) render(text string) string {
	clean := cleanText(text)
	r, err := NewMarkdownRenderer()
	if err != nil {
		return clean
	}
	rendered, err := r.Render(clean)
	if err != nil {
		return clean
	}
	return strings.TrimSpace(rendered)
}

func New(debug bool) *UI {
	// When stdout is piped, write UI to stderr
	out := io.Writer(os.Stdout)
	piped := pipe.IsStdoutPiped()
	if piped {
		out = os.Stderr
	}

	return &UI{
		debug:    debug,
		depth:    0,
		styles:   DefaultStyles(),
		renderer: &markdownRenderer{},
		out:      out,
		piped:    piped,
	}
}

// NewWithDepth creates a UI at a specific nesting depth (for sub-agent calls).
func NewWithDepth(debug bool, depth int) *UI {
	out := io.Writer(os.Stdout)
	piped := pipe.IsStdoutPiped()
	if piped {
		out = os.Stderr
	}

	return &UI{
		debug:    debug,
		depth:    depth,
		styles:   DefaultStyles(),
		renderer: &markdownRenderer{},
		out:      out,
		piped:    piped,
	}
}

// NewWithWriter creates a UI with a custom output writer.
func NewWithWriter(debug bool, out io.Writer) *UI {
	return &UI{
		debug:    debug,
		depth:    0,
		styles:   DefaultStyles(),
		renderer: &markdownRenderer{},
		out:      out,
		piped:    false,
	}
}

// IsPiped returns true if output is being piped.
func (u *UI) IsPiped() bool {
	return u.piped
}

// Depth returns the nesting depth (0 = top-level).
func (u *UI) Depth() int {
	return u.depth
}

// IsNested returns true if this is a sub-agent call.
func (u *UI) IsNested() bool {
	return u.depth > 0
}

// indent returns the indentation prefix for the current depth.
func (u *UI) indent() string {
	if u.depth == 0 {
		return ""
	}
	// Use a vertical line indicator for nested output
	return strings.Repeat("  ", u.depth) + "│ "
}

// indentLines indents each line of text for nested display.
func (u *UI) indentLines(text string) string {
	if u.depth == 0 {
		return text
	}
	prefix := u.indent()
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

// print writes to the UI output.
func (u *UI) print(a ...any) {
	fmt.Fprint(u.out, a...)
}

// println writes to the UI output with a newline.
func (u *UI) println(a ...any) {
	fmt.Fprintln(u.out, a...)
}

// printf writes formatted output to the UI.
func (u *UI) printf(format string, a ...any) {
	fmt.Fprintf(u.out, format, a...)
}

func renderPlainFence(content string) string {
	var buf bytes.Buffer
	buf.WriteString("```")
	buf.WriteByte('\n')
	buf.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		buf.WriteByte('\n')
	}
	buf.WriteString("```")
	return buf.String()
}

func renderJSONFence(content string) string {
	var buf bytes.Buffer
	buf.WriteString("```json")
	buf.WriteByte('\n')
	buf.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		buf.WriteByte('\n')
	}
	buf.WriteString("```")
	return buf.String()
}

// renderPlainOutput renders plain text output with line and width truncation.
// It normalizes line endings, replaces tabs with spaces, and truncates both
// vertically (by line count) and horizontally (by terminal width).
func renderPlainOutput(out string, maxLines int) string {
	// Normalize line endings and tabs
	out = strings.ReplaceAll(out, "\r\n", "\n")
	out = strings.ReplaceAll(out, "\t", "    ")
	out = strings.TrimSpace(out)

	lines := strings.Split(out, "\n")
	totalLines := len(lines)

	// Get max width for line truncation
	maxWidth := getTerminalWidth() - 4 // Account for box padding
	if maxWidth > 116 {
		maxWidth = 116
	}

	lineStyle := lipgloss.NewStyle().Foreground(colorTextDim)
	truncStyle := lipgloss.NewStyle().Foreground(colorMuted).Italic(true)

	var result []string
	displayLines := min(len(lines), maxLines)

	for i := 0; i < displayLines; i++ {
		line := lines[i]
		// Truncate long lines with ellipsis
		if lipgloss.Width(line) > maxWidth {
			line = ansi.Truncate(line, maxWidth, "…")
		}
		result = append(result, lineStyle.Render(line))
	}

	// Add truncation indicator if needed
	if totalLines > maxLines {
		result = append(result, truncStyle.Render(fmt.Sprintf("… (%d more lines)", totalLines-maxLines)))
	}

	return strings.Join(result, "\n")
}

func (u *UI) RenderFinal(text string) string {
	return u.renderer.render(text)
}

// RenderJSON renders JSON with syntax highlighting using glamour/chroma.
func (u *UI) RenderJSON(jsonStr string) string {
	return u.renderer.render(renderJSONFence(jsonStr))
}

// RenderJSONString renders JSON with syntax highlighting (standalone function).
func RenderJSONString(jsonStr string) string {
	r := &markdownRenderer{}
	return r.render(renderJSONFence(jsonStr))
}

func (u *UI) renderReasoning(text string) {
	label := FormatReasoningLabel()
	u.println(label)

	body := u.renderer.render(text)
	box := u.styles.ReasoningBox.Render(body)
	u.println(box)
}

func (u *UI) PrintToolOutputs(order []string, outputs []string) {
	for i, out := range outputs {
		var obj map[string]any
		if err := json.Unmarshal([]byte(out), &obj); err == nil {
			toolName, _ := obj["tool"].(string)
			typeStr, _ := obj["type"].(string)
			content, _ := obj["content"].(string)
			status, _ := obj["status"].(string)

			if toolName == "" {
				toolName = fmt.Sprintf("tool #%d", i+1)
			}

			if toolName == "bash" {
				u.renderBashResult(toolName, extractBashCommand(obj), content, status)
				continue
			}
			u.renderToolResult(toolName, typeStr, content, status)
			continue
		}
		clean := cleanText(out)
		u.renderToolResult(fmt.Sprintf("tool #%d", i+1), "text", clean, "")
	}
}

func extractBashCommand(obj map[string]any) string {
	if parsed, ok := obj["parsed_args"].(map[string]any); ok {
		if cmd, ok := parsed["command"].(string); ok {
			return cmd
		}
	}
	rawArgs, ok := obj["raw_args"].(string)
	if !ok {
		return ""
	}
	var rawObj map[string]any
	if err := json.Unmarshal([]byte(rawArgs), &rawObj); err == nil {
		if cmd, ok := rawObj["command"].(string); ok {
			return cmd
		}
	}
	return ""
}

func cleanText(text string) string {
	if unquoted, err := strconv.Unquote("\"" + text + "\""); err == nil {
		text = unquoted
	} else {
		text = strings.ReplaceAll(text, "\\n", "\n")
		text = strings.ReplaceAll(text, "\\t", "\t")
	}
	return ansi.Strip(text)
}

func (u *UI) SelectAgent(ctx context.Context, handles []string) (SelectAgentResult, error) {
	var selected string
	options := make([]huh.Option[string], 0, len(handles))
	for _, h := range handles {
		options = append(options, huh.NewOption(h, h))
	}

	theme := huh.ThemeCharm()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose an agent").
				Options(options...).
				Value(&selected),
		),
	).WithTheme(theme)

	if err := form.RunWithContext(ctx); err != nil {
		return SelectAgentResult{}, err
	}

	return SelectAgentResult{Handle: selected}, nil
}

func (u *UI) PrintResult(text string) {
	rendered := u.renderer.render(text)
	u.println(rendered)
}

func (u *UI) PrintReasoning(text string) {
	u.renderReasoning(text)
}

func (u *UI) PrintToolOutput(label, body string) {
	styledLabel := FormatToolLabel(label, 0)
	u.println(styledLabel)
	rendered := u.renderer.render(body)
	box := u.styles.ToolBox.Render(rendered)
	u.println(box)
}

func (u *UI) renderBashResult(label, command, content, status string) {
	// Build the header with tool name and status indicator
	headerStyle := u.styles.ToolLabel
	statusIcon := ""
	if status == "error" || status == "failed" {
		statusIcon = " " + u.styles.StatusError.String()
	} else if status == "success" || status == "" {
		statusIcon = ""
	}

	header := headerStyle.Render(IconBash + " " + label + statusIcon)
	u.println(header)

	cmd := cleanText(command)
	out := cleanText(content)

	var parts []string

	// Render command with special styling
	if strings.TrimSpace(cmd) != "" {
		cmdStyle := lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)
		cmdLine := cmdStyle.Render("$ " + cmd)
		parts = append(parts, cmdLine)
	}

	// Render output
	if strings.TrimSpace(out) != "" {
		// Check if output looks like JSON and pretty-print it
		trimmed := strings.TrimSpace(out)
		if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
			(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, []byte(trimmed), "", "  "); err == nil {
				// Successfully parsed as JSON, render with syntax highlighting
				rendered := u.renderer.render(renderJSONFence(prettyJSON.String()))
				parts = append(parts, rendered)
			} else {
				// Not valid JSON, render as plain text
				parts = append(parts, renderPlainOutput(out, 50))
			}
		} else {
			parts = append(parts, renderPlainOutput(out, 50))
		}
	}

	if len(parts) > 0 {
		content := strings.Join(parts, "\n\n")
		box := u.styles.ToolBox.Render(content)
		u.println(box)
	}
}

func (u *UI) renderToolResult(label, typ, content, status string) {
	// Build the header
	header := FormatToolLabel(label, 0)
	u.println(header)

	plain := cleanText(content)

	// Choose rendering based on type
	var rendered string
	switch typ {
	case "json", "object":
		// Pretty format JSON if possible
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, []byte(plain), "", "  "); err == nil {
			plain = prettyJSON.String()
		}
		rendered = u.renderer.render(renderPlainFence(plain))
	case "markdown", "md":
		rendered = u.renderer.render(plain)
	default:
		// Limit output display for very long outputs
		lines := strings.Split(plain, "\n")
		if len(lines) > 50 {
			truncated := strings.Join(lines[:25], "\n")
			truncated += fmt.Sprintf("\n\n... (%d lines omitted) ...\n\n", len(lines)-50)
			truncated += strings.Join(lines[len(lines)-25:], "\n")
			plain = truncated
		}
		rendered = u.renderer.render(renderPlainFence(plain))
	}

	box := u.styles.ToolBox.Render(rendered)
	u.println(box)
}

// PrintError prints an error message with appropriate styling.
func (u *UI) PrintError(msg string) {
	label := FormatErrorLabel("Error")
	u.println(label)
	box := u.styles.ErrorBox.Render(msg)
	u.println(box)
}

// PrintSuccess prints a success message with appropriate styling.
func (u *UI) PrintSuccess(msg string) {
	label := FormatSuccessLabel("Success")
	u.println(label)
	u.println(u.styles.Muted.Render(msg))
}

// PrintInfo prints an info message.
func (u *UI) PrintInfo(msg string) {
	label := u.styles.InfoLabel.Render(IconInfo + " Info")
	u.println(label)
	u.println(u.styles.Muted.Render(msg))
}

// PrintDivider prints a horizontal divider.
func (u *UI) PrintDivider() {
	width := u.styles.MaxWidth
	if width > 60 {
		width = 60
	}
	divider := lipgloss.NewStyle().
		Foreground(colorSubtle).
		Render(strings.Repeat("─", width))
	u.println(divider)
}

// PrintChatHeader prints the header for an interactive chat session.
func (u *UI) PrintChatHeader(agentHandle string) {
	u.PrintChatHeaderWithSkills(agentHandle, 0)
}

// PrintChatHeaderWithSkills prints the header with skill count.
func (u *UI) PrintChatHeaderWithSkills(agentHandle string, skillCount int) {
	width := getTerminalWidth()
	if width > 100 {
		width = 100
	}

	// Create styled components
	titleStyle := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	hintStyle := lipgloss.NewStyle().
		Foreground(colorMuted)

	skillStyle := lipgloss.NewStyle().
		Foreground(colorTertiary)

	lineStyle := lipgloss.NewStyle().
		Foreground(colorSubtle)

	// Build the title part
	title := titleStyle.Render(fmt.Sprintf("Chat with %s", agentHandle))

	var skillsInfo string
	if skillCount > 0 {
		skillsInfo = skillStyle.Render(fmt.Sprintf(" %s %d skills", IconBullet, skillCount))
	}

	hint := hintStyle.Render("^H history · ^C exit")

	// Calculate line length to fill remaining space
	// Format: "  title skills ─── hint  "
	contentWidth := lipgloss.Width(title) + lipgloss.Width(skillsInfo) + lipgloss.Width(hint) + 6 // 6 = padding + min gap
	lineWidth := width - contentWidth
	if lineWidth < 3 {
		lineWidth = 3
	}

	line := lineStyle.Render(strings.Repeat("─", lineWidth))

	// Compose the header line
	header := "  " + title + skillsInfo + " " + line + " " + hint

	u.println()
	u.println(header)
	u.println()
}

// PrintUserPrompt prints the styled user input prompt and returns the styled prefix.
func (u *UI) PrintUserPrompt() {
	prompt := lipgloss.NewStyle().
		Foreground(colorSecondary).
		Bold(true).
		Render("> ")
	u.print(prompt)
}

// PrintAssistantLabel prints a label before assistant responses.
func (u *UI) PrintAssistantLabel() {
	label := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Render(IconArrowRight)
	u.print(label + " ")
}

// PrintAgentResponseHeader prints a header for the agent's response.
func (u *UI) PrintAgentResponseHeader(agentHandle string) {
	indent := u.indent()
	iconStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	handleStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)

	u.printf("%s%s %s\n", indent, iconStyle.Render(IconArrowRight), handleStyle.Render(agentHandle))
}

// ToolCallInfo represents a tool call for display.
type ToolCallInfo struct {
	Name        string
	Description string // For bash, the description param
	Command     string // For bash, the actual command
	Input       string // JSON input
	Output      string // Result output
	Error       string // Error message if failed
	Duration    string // How long the call took
	Metadata    string // Tool-specific metadata (JSON)
}

// PrintToolCallStart prints the tool call header with the command.
func (u *UI) PrintToolCallStart(tc ToolCallInfo) {
	indent := u.indent()

	// Print header: ❯ bash · description
	iconStyle := lipgloss.NewStyle().Foreground(colorTertiary).Bold(true)
	toolStyle := lipgloss.NewStyle().Foreground(colorTertiary).Bold(true)
	sepStyle := lipgloss.NewStyle().Foreground(colorMuted)
	labelStyle := lipgloss.NewStyle().Foreground(colorText)

	label := tc.Description
	if label == "" {
		label = tc.Name
	}

	switch tc.Name {
	case "bash":
		u.printf("%s%s %s %s %s\n",
			indent,
			iconStyle.Render(IconBash),
			toolStyle.Render("bash"),
			sepStyle.Render("·"),
			labelStyle.Render(label))
	case "plan":
		u.printf("%s%s %s %s %s\n",
			indent,
			iconStyle.Render(IconPlan),
			toolStyle.Render("plan"),
			sepStyle.Render("·"),
			labelStyle.Render("updating plan"))
	default:
		u.printf("%s%s %s\n",
			indent,
			iconStyle.Render(IconTool),
			labelStyle.Render(tc.Name))
	}

	// Print the actual command indented
	if tc.Command != "" {
		cmdStyle := lipgloss.NewStyle().Foreground(colorMuted)
		u.printf("%s  %s\n", indent, cmdStyle.Render("$ "+tc.Command))
	}
}

// PrintToolCallResult prints the result of a tool call.
func (u *UI) PrintToolCallResult(tc ToolCallInfo) {
	indent := u.indent()

	// Handle plan tool specially
	if tc.Name == "plan" {
		u.printPlanResult(tc)
		return
	}

	// Status line: ✓ completed (1.2s) or ✕ failed (1.2s)
	var statusIcon string
	var statusColor lipgloss.Color
	var statusText string

	if tc.Error != "" {
		statusIcon = IconError
		statusColor = colorError
		statusText = "failed"
	} else {
		statusIcon = IconSuccess
		statusColor = colorSuccess
		statusText = "completed"
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	durationStyle := lipgloss.NewStyle().Foreground(colorMuted)

	u.printf("%s  %s %s %s\n",
		indent,
		statusStyle.Render(statusIcon),
		statusStyle.Render(statusText),
		durationStyle.Render("("+tc.Duration+")"))

	// Parse bash tool JSON output to extract stdout/stderr
	output := tc.Output
	isError := tc.Error != ""
	if output != "" {
		var bashResult struct {
			Stdout   string `json:"stdout"`
			Stderr   string `json:"stderr"`
			ExitCode int    `json:"exit_code"`
			Error    string `json:"error"`
		}
		if err := json.Unmarshal([]byte(output), &bashResult); err == nil {
			// Successfully parsed bash JSON - use stdout/stderr
			if bashResult.Error != "" {
				output = bashResult.Error
				isError = true
			} else if bashResult.Stderr != "" {
				output = bashResult.Stderr
				isError = bashResult.ExitCode != 0
			} else {
				output = bashResult.Stdout
			}
		}
		u.printCommandOutput(output, isError)
	}

	u.println() // Blank line after each tool call
}

// planResponseMetadata mirrors run.PlanResponseMetadata to avoid circular imports.
type planResponseMetadata struct {
	IsNew         bool         `json:"is_new"`
	Plan          session.Plan `json:"plan"`
	JustCompleted []string     `json:"just_completed,omitempty"`
	JustStarted   string       `json:"just_started,omitempty"`
	Completed     int          `json:"completed"`
	Total         int          `json:"total"`
}

// printPlanResult handles plan tool output display.
func (u *UI) printPlanResult(tc ToolCallInfo) {
	indent := u.indent()

	if tc.Error != "" {
		statusStyle := lipgloss.NewStyle().Foreground(colorError)
		durationStyle := lipgloss.NewStyle().Foreground(colorMuted)
		u.printf("%s  %s %s %s\n",
			indent,
			statusStyle.Render(IconError),
			statusStyle.Render("failed"),
			durationStyle.Render("("+tc.Duration+")"))
		u.println(lipgloss.NewStyle().Foreground(colorError).Render(indent + "  " + tc.Error))
		u.println()
		return
	}

	statusStyle := lipgloss.NewStyle().Foreground(colorSuccess)
	durationStyle := lipgloss.NewStyle().Foreground(colorMuted)

	// Try to parse metadata for rich display
	var meta planResponseMetadata
	if tc.Metadata != "" {
		if err := json.Unmarshal([]byte(tc.Metadata), &meta); err == nil {
			u.printPlanWithMetadata(tc.Duration, meta)
			return
		}
	}

	// Fallback: parse counts from output text
	summaryStyle := lipgloss.NewStyle().Foreground(colorTextDim)
	var summary string
	lines := strings.Split(tc.Output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Status:") {
			summary = strings.TrimPrefix(line, "Status: ")
			break
		}
	}

	if summary != "" {
		u.printf("%s  %s %s %s %s\n",
			indent,
			statusStyle.Render(IconSuccess),
			statusStyle.Render("plan updated"),
			durationStyle.Render("("+tc.Duration+")"),
			summaryStyle.Render("· "+summary))
	} else {
		u.printf("%s  %s %s %s\n",
			indent,
			statusStyle.Render(IconSuccess),
			statusStyle.Render("plan updated"),
			durationStyle.Render("("+tc.Duration+")"))
	}

	u.println()
}

// printPlanWithMetadata renders a rich plan display using metadata.
func (u *UI) printPlanWithMetadata(duration string, meta planResponseMetadata) {
	indent := u.indent()
	statusStyle := lipgloss.NewStyle().Foreground(colorSuccess)
	durationStyle := lipgloss.NewStyle().Foreground(colorMuted)
	summaryStyle := lipgloss.NewStyle().Foreground(colorTextDim)

	// Get terminal width for formatting
	width := 80
	if w, _, err := term.GetSize(os.Stdout.Fd()); err == nil && w > 0 {
		width = w
	}

	if meta.IsNew {
		// New plan: show full plan structure
		headerText := fmt.Sprintf("created %d items", meta.Total)
		u.printf("%s  %s %s %s %s\n",
			indent,
			statusStyle.Render(IconSuccess),
			statusStyle.Render("plan"),
			durationStyle.Render("("+duration+")"),
			summaryStyle.Render("· "+headerText))

		// Show the full plan indented
		planStr := FormatPlan(meta.Plan, width-4)
		if planStr != "" {
			for _, line := range strings.Split(planStr, "\n") {
				u.printf("%s  %s\n", indent, line)
			}
		}
	} else {
		// Update: show compact progress line with current action
		progressText := fmt.Sprintf("%d/%d", meta.Completed, meta.Total)

		// Show what's in progress
		if meta.JustStarted != "" {
			// Truncate if needed
			maxLen := width - 40
			actionText := meta.JustStarted
			if maxLen > 0 && len(actionText) > maxLen {
				actionText = actionText[:maxLen-1] + "…"
			}

			actionStyle := lipgloss.NewStyle().Foreground(colorText)
			u.printf("%s  %s %s %s %s %s\n",
				indent,
				statusStyle.Render(IconSuccess),
				statusStyle.Render("plan"),
				durationStyle.Render("("+duration+")"),
				summaryStyle.Render("· "+progressText),
				actionStyle.Render("▸ "+actionText))
		} else if len(meta.JustCompleted) > 0 {
			// Show what was just completed
			completedText := meta.JustCompleted[len(meta.JustCompleted)-1]
			maxLen := width - 40
			if maxLen > 0 && len(completedText) > maxLen {
				completedText = completedText[:maxLen-1] + "…"
			}

			completedStyle := lipgloss.NewStyle().Foreground(colorTextDim)
			u.printf("%s  %s %s %s %s %s\n",
				indent,
				statusStyle.Render(IconSuccess),
				statusStyle.Render("plan"),
				durationStyle.Render("("+duration+")"),
				summaryStyle.Render("· "+progressText),
				completedStyle.Render("✓ "+completedText))
		} else {
			u.printf("%s  %s %s %s %s\n",
				indent,
				statusStyle.Render(IconSuccess),
				statusStyle.Render("plan"),
				durationStyle.Render("("+duration+")"),
				summaryStyle.Render("· "+progressText))
		}
	}

	u.println()
}

func (u *UI) printCommandOutput(output string, isError bool) {
	indent := u.indent()
	clean := cleanText(output)
	if strings.TrimSpace(clean) == "" {
		return
	}

	// Check if output is JSON and render with lipgloss components
	trimmed := strings.TrimSpace(clean)
	if (strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")) {
		if rendered, ok := JSONToRenderedOutput(trimmed); ok {
			for _, line := range strings.Split(rendered, "\n") {
				fmt.Println("  " + line)
			}
			return
		}
	}

	lines := strings.Split(clean, "\n")

	// Truncate long output
	maxLines := 20
	truncated := false
	if len(lines) > maxLines {
		truncated = true
		// Show first 10 and last 5
		head := lines[:10]
		tail := lines[len(lines)-5:]
		lines = append(head, fmt.Sprintf("  ... (%d lines omitted) ...", len(lines)-15))
		lines = append(lines, tail...)
	}

	outputStyle := lipgloss.NewStyle().Foreground(colorTextDim)
	if isError {
		outputStyle = lipgloss.NewStyle().Foreground(colorError)
	}

	for _, line := range lines {
		u.println(outputStyle.Render(indent + "  " + line))
	}

	if truncated {
		hintStyle := lipgloss.NewStyle().Foreground(colorMuted).Italic(true)
		u.println(hintStyle.Render(indent + "  (output truncated)"))
	}
}

// PrintReasoningStart prints the start of reasoning.
func (u *UI) PrintReasoningStart() {
	indent := u.indent()
	style := lipgloss.NewStyle().Foreground(colorMuted).Italic(true)
	u.print(style.Render(indent + "Thinking: "))
}

// PrintReasoningDelta prints streaming reasoning content.
func (u *UI) PrintReasoningDelta(text string) {
	style := lipgloss.NewStyle().Foreground(colorTextDim).Italic(true)
	u.print(style.Render(text))
}

// PrintReasoningEnd prints a newline after reasoning is complete.
func (u *UI) PrintReasoningEnd() {
	u.println()
	u.println()
}

// PrintTextDelta prints streaming text content.
func (u *UI) PrintTextDelta(text string) {
	u.print(text)
}

// PrintTextEnd prints a newline after text streaming is complete.
func (u *UI) PrintTextEnd() {
	u.println()
}

// PrintThinkingDone prints the "Thought for Xs" summary.
func (u *UI) PrintThinkingDone(duration string) {
	indent := u.indent()
	style := lipgloss.NewStyle().Foreground(colorMuted)
	u.println(style.Render(indent + fmt.Sprintf("Thought for %s", duration)))
	u.println()
}

// PrintSubAgentStart prints the header for a sub-agent call.
func (u *UI) PrintSubAgentStart(agentHandle, prompt string) {
	indent := u.indent()

	// Header with agent icon and handle
	iconStyle := lipgloss.NewStyle().Foreground(colorSecondary).Bold(true)
	handleStyle := lipgloss.NewStyle().Foreground(colorSecondary).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(colorMuted)

	u.printf("%s%s %s %s\n",
		indent,
		iconStyle.Render(IconSubAgent),
		handleStyle.Render(agentHandle),
		labelStyle.Render("sub-agent"))

	// Show truncated prompt
	if prompt != "" {
		promptStyle := lipgloss.NewStyle().Foreground(colorTextDim).Italic(true)
		displayPrompt := prompt
		if len(displayPrompt) > 80 {
			displayPrompt = displayPrompt[:77] + "..."
		}
		// Replace newlines with spaces for display
		displayPrompt = strings.ReplaceAll(displayPrompt, "\n", " ")
		u.printf("%s  %s\n", indent, promptStyle.Render(displayPrompt))
	}
}

// PrintSubAgentEnd prints the completion status for a sub-agent call.
func (u *UI) PrintSubAgentEnd(agentHandle string, duration string, hasError bool) {
	indent := u.indent()

	var statusIcon string
	var statusColor lipgloss.Color
	var statusText string

	if hasError {
		statusIcon = IconError
		statusColor = colorError
		statusText = "failed"
	} else {
		statusIcon = IconSuccess
		statusColor = colorSuccess
		statusText = "completed"
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	durationStyle := lipgloss.NewStyle().Foreground(colorMuted)

	u.printf("%s%s %s %s\n\n",
		indent,
		statusStyle.Render(statusIcon),
		statusStyle.Render(statusText),
		durationStyle.Render("("+duration+")"))
}

// MemoryEventType represents the type of memory event.
type MemoryEventType string

const (
	MemoryCreated    MemoryEventType = "created"
	MemorySkipped    MemoryEventType = "skipped"
	MemorySuperseded MemoryEventType = "superseded"
	MemoryFailed     MemoryEventType = "failed"
)

// PrintMemoryEvent prints memory formation feedback.
func (u *UI) PrintMemoryEvent(event MemoryEventType) {
	if u.piped {
		return
	}

	indent := u.indent()
	var icon, msg string
	var color lipgloss.Color

	switch event {
	case MemoryCreated:
		icon = "◆"
		msg = "Remembered"
		color = lipgloss.Color("70") // Green
	case MemorySkipped:
		icon = "◇"
		msg = "Already remembered"
		color = lipgloss.Color("242") // Gray
	case MemorySuperseded:
		icon = "◆"
		msg = "Memory updated"
		color = lipgloss.Color("214") // Orange
	case MemoryFailed:
		icon = "×"
		msg = "Failed to remember"
		color = lipgloss.Color("196") // Red
	default:
		return
	}

	style := lipgloss.NewStyle().Foreground(color)
	u.printf("%s%s\n", indent, style.Render(fmt.Sprintf("  %s %s", icon, msg)))
}
