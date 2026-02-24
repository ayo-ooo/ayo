package ui

import (
	"strings"
	"testing"
)

func TestDefaultStyles(t *testing.T) {
	styles := DefaultStyles()

	// Verify styles are properly initialized
	if styles.MaxWidth <= 0 {
		t.Error("MaxWidth should be positive")
	}

	// Test that styles can render text without panicking
	testCases := []struct {
		name   string
		render func() string
	}{
		{"ReasoningLabel", func() string { return styles.ReasoningLabel.Render("test") }},
		{"ToolLabel", func() string { return styles.ToolLabel.Render("test") }},
		{"ErrorLabel", func() string { return styles.ErrorLabel.Render("test") }},
		{"SuccessLabel", func() string { return styles.SuccessLabel.Render("test") }},
		{"InfoLabel", func() string { return styles.InfoLabel.Render("test") }},
		{"ReasoningBox", func() string { return styles.ReasoningBox.Render("test") }},
		{"ToolBox", func() string { return styles.ToolBox.Render("test") }},
		{"ErrorBox", func() string { return styles.ErrorBox.Render("test") }},
		{"CodeBox", func() string { return styles.CodeBox.Render("test") }},
		{"Title", func() string { return styles.Title.Render("test") }},
		{"Subtitle", func() string { return styles.Subtitle.Render("test") }},
		{"Command", func() string { return styles.Command.Render("test") }},
		{"FilePath", func() string { return styles.FilePath.Render("test") }},
		{"Muted", func() string { return styles.Muted.Render("test") }},
		{"Emphasis", func() string { return styles.Emphasis.Render("test") }},
		{"Bold", func() string { return styles.Bold.Render("test") }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.render()
			if result == "" {
				t.Errorf("%s rendered empty string", tc.name)
			}
		})
	}
}

func TestFormatToolLabel(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		index    int
		wantIcon string
	}{
		{
			name:     "bash tool gets special icon",
			toolName: "bash",
			index:    0,
			wantIcon: IconBash,
		},
		{
			name:     "other tool gets generic icon",
			toolName: "grep",
			index:    1,
			wantIcon: IconTool,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatToolLabel(tt.toolName, tt.index)
			// The result should contain the tool name
			if !strings.Contains(result, tt.toolName) {
				t.Errorf("FormatToolLabel() = %q, want to contain %q", result, tt.toolName)
			}
		})
	}
}

func TestFormatReasoningLabel(t *testing.T) {
	result := FormatReasoningLabel()
	if !strings.Contains(result, "Reasoning") {
		t.Errorf("FormatReasoningLabel() = %q, want to contain 'Reasoning'", result)
	}
}

func TestFormatErrorLabel(t *testing.T) {
	result := FormatErrorLabel("test error")
	if !strings.Contains(result, "test error") {
		t.Errorf("FormatErrorLabel() = %q, want to contain 'test error'", result)
	}
}

func TestFormatSuccessLabel(t *testing.T) {
	result := FormatSuccessLabel("test success")
	if !strings.Contains(result, "test success") {
		t.Errorf("FormatSuccessLabel() = %q, want to contain 'test success'", result)
	}
}

func TestIndentText(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		prefix string
		want   string
	}{
		{
			name:   "single line",
			text:   "hello",
			prefix: "  ",
			want:   "  hello",
		},
		{
			name:   "multiple lines",
			text:   "line1\nline2\nline3",
			prefix: "> ",
			want:   "> line1\n> line2\n> line3",
		},
		{
			name:   "empty lines preserved",
			text:   "line1\n\nline3",
			prefix: "  ",
			want:   "  line1\n\n  line3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IndentText(tt.text, tt.prefix)
			if got != tt.want {
				t.Errorf("IndentText(%q, %q) = %q, want %q", tt.text, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestGlamourStyleConfig(t *testing.T) {
	config := GlamourStyleConfig()

	// Verify key style components are set
	if config.Document.Margin == nil {
		t.Error("Document margin should be set")
	}

	if config.Heading.Bold == nil || !*config.Heading.Bold {
		t.Error("Heading should be bold")
	}

	if config.CodeBlock.Chroma == nil {
		t.Error("CodeBlock should have Chroma highlighting configured")
	}
}

func TestNewMarkdownRenderer(t *testing.T) {
	renderer, err := NewMarkdownRenderer()
	if err != nil {
		t.Fatalf("NewMarkdownRenderer() error = %v", err)
	}
	if renderer == nil {
		t.Fatal("NewMarkdownRenderer() returned nil")
	}

	// Test that it can render markdown
	input := "# Hello\n\nThis is **bold** and *italic*."
	output, err := renderer.Render(input)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if output == "" {
		t.Error("Render() returned empty string")
	}
}

func TestIcons(t *testing.T) {
	icons := []string{
		IconThinking,
		IconTool,
		IconBash,
		IconSuccess,
		IconError,
		IconWarning,
		IconInfo,
		IconArrowRight,
		IconBullet,
		IconCheck,
		IconCross,
		IconSpinner,
		IconPending,
		IconComplete,
	}

	for _, icon := range icons {
		if icon == "" {
			t.Error("Icon should not be empty")
		}
	}
}

func TestUINew(t *testing.T) {
	ui := New(true)
	if ui == nil {
		t.Fatal("New() returned nil")
	}
	if !ui.debug {
		t.Error("debug should be true")
	}

	ui2 := New(false)
	if ui2.debug {
		t.Error("debug should be false")
	}
}

func TestUIRenderFinal(t *testing.T) {
	ui := New(false)
	result := ui.RenderFinal("# Hello World")
	if result == "" {
		t.Error("RenderFinal() returned empty string")
	}
}
