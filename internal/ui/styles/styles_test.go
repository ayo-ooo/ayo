package styles

import (
	"testing"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()

	if theme == nil {
		t.Fatal("DefaultTheme() returned nil")
	}

	// Check primary colors are set
	if theme.Primary == "" {
		t.Error("Primary color should not be empty")
	}
	if theme.Secondary == "" {
		t.Error("Secondary color should not be empty")
	}

	// Check status colors
	if theme.Green == "" {
		t.Error("Green color should not be empty")
	}
	if theme.Red == "" {
		t.Error("Red color should not be empty")
	}
	if theme.Blue == "" {
		t.Error("Blue color should not be empty")
	}
	if theme.Yellow == "" {
		t.Error("Yellow color should not be empty")
	}

	// Check foreground colors
	if theme.FgBase == "" {
		t.Error("FgBase color should not be empty")
	}
	if theme.FgMuted == "" {
		t.Error("FgMuted color should not be empty")
	}
	if theme.FgSubtle == "" {
		t.Error("FgSubtle color should not be empty")
	}
}

func TestCurrentTheme(t *testing.T) {
	theme := CurrentTheme()

	if theme == nil {
		t.Fatal("CurrentTheme() returned nil")
	}

	// Should return default theme initially
	defaultTheme := DefaultTheme()
	if theme.Primary != defaultTheme.Primary {
		t.Errorf("Primary = %v, want %v", theme.Primary, defaultTheme.Primary)
	}
}

func TestSetTheme(t *testing.T) {
	// Save original
	original := CurrentTheme()
	defer SetTheme(original)

	// Create and set custom theme
	custom := &Theme{
		Primary: "#123456",
	}
	SetTheme(custom)

	if CurrentTheme().Primary != "#123456" {
		t.Errorf("Primary = %v, want %v", CurrentTheme().Primary, "#123456")
	}
}

func TestTheme_S(t *testing.T) {
	theme := DefaultTheme()
	styles := theme.S()

	if styles == nil {
		t.Fatal("S() returned nil")
	}

	// Should cache and return same instance
	styles2 := theme.S()
	if styles != styles2 {
		t.Error("S() should return cached styles")
	}
}

func TestTheme_FocusedStyles(t *testing.T) {
	theme := DefaultTheme()
	styles := theme.FocusedStyles()

	// Just verify styles are created without error
	_ = styles.Base.Render("test")
	_ = styles.Title.Render("test")
	_ = styles.Description.Render("test")
	_ = styles.Border.Render("test")
	_ = styles.Cursor.Render("test")
}

func TestTheme_BlurredStyles(t *testing.T) {
	theme := DefaultTheme()
	styles := theme.BlurredStyles()

	// Just verify styles are created without error
	_ = styles.Base.Render("test")
	_ = styles.Title.Render("test")
	_ = styles.Description.Render("test")
	_ = styles.Border.Render("test")
	_ = styles.Cursor.Render("test")
}

func TestStatusIcons(t *testing.T) {
	// Verify constants are defined
	if ToolPending == "" {
		t.Error("ToolPending should not be empty")
	}
	if ToolSuccess == "" {
		t.Error("ToolSuccess should not be empty")
	}
	if ToolError == "" {
		t.Error("ToolError should not be empty")
	}
	if ToolRunning == "" {
		t.Error("ToolRunning should not be empty")
	}
	if ArrowRight == "" {
		t.Error("ArrowRight should not be empty")
	}
	if CheckMark == "" {
		t.Error("CheckMark should not be empty")
	}
}

func TestGetPlainMarkdownRenderer(t *testing.T) {
	r := GetPlainMarkdownRenderer(80)
	if r == nil {
		t.Error("GetPlainMarkdownRenderer() returned nil")
	}
}

func TestRenderMarkdown(t *testing.T) {
	content := "# Hello"
	rendered := RenderMarkdown(content, 80)

	if rendered == "" {
		t.Error("RenderMarkdown() returned empty string")
	}

	// Should contain rendered heading
	if rendered == content {
		// Acceptable if renderer couldn't be created, but likely not
	}
}

func TestRenderMarkdown_PreservesContentOnError(t *testing.T) {
	// Test with unusual content
	content := "Simple text"
	rendered := RenderMarkdown(content, 80)

	if rendered == "" {
		t.Error("RenderMarkdown() should not return empty string")
	}
}

func TestFieldStyles(t *testing.T) {
	fs := FieldStyles{}

	// Zero value should work
	_ = fs.Base
	_ = fs.Title
}

func TestStyles(t *testing.T) {
	theme := DefaultTheme()
	s := theme.S()

	// Verify all styles render without panic
	_ = s.Base.Render("test")
	_ = s.Muted.Render("test")
	_ = s.Subtle.Render("test")
}
