package interactive

import (
	"strings"
	"testing"
)

func TestToolDisplay_Running(t *testing.T) {
	display := NewToolDisplay("bash", "ls -la")

	view := display.View()
	if !strings.Contains(view, "bash") {
		t.Error("expected tool name in view")
	}
	if !strings.Contains(view, "ls -la") {
		t.Error("expected input in view")
	}
	// Running state should show spinner indicator
	if !strings.Contains(view, "⋯") {
		t.Error("expected running indicator in view")
	}
}

func TestToolDisplay_Success(t *testing.T) {
	display := NewToolDisplay("bash", "ls")
	display.SetResult("file1.txt\nfile2.txt", true)

	view := display.View()
	if !strings.Contains(view, "▸") {
		t.Error("expected chevron in view")
	}
	if !strings.Contains(view, "└─") {
		t.Error("expected result line in view")
	}
}

func TestToolDisplay_Error(t *testing.T) {
	display := NewToolDisplay("bash", "cat missing.txt")
	display.SetResult("file not found", false)

	view := display.View()
	if !strings.Contains(view, "✗") {
		t.Error("expected error marker in view")
	}
	if !strings.Contains(view, "file not found") {
		t.Error("expected error message in view")
	}
}

func TestSummarizeResult_Bash(t *testing.T) {
	tests := []struct {
		output   string
		expected string
	}{
		{"file1.txt", "file1.txt"},
		{"line1\nline2\nline3\nline4\nline5\nline6\nline7", "6 lines of output"},
		{"", "completed"},
		{strings.Repeat("x", 100), strings.Repeat("x", 57) + "..."},
	}

	for _, tt := range tests {
		result := summarizeResult("bash", tt.output)
		if result != tt.expected {
			t.Errorf("summarizeResult('bash', %q) = %q, want %q", tt.output, result, tt.expected)
		}
	}
}

func TestSummarizeResult_View(t *testing.T) {
	result := summarizeResult("view", "line1\nline2\nline3")
	if result != "Read 2 lines" {
		t.Errorf("expected 'Read 2 lines', got %q", result)
	}
}

func TestSummarizeResult_Edit(t *testing.T) {
	result := summarizeResult("edit", "anything")
	if result != "File updated" {
		t.Errorf("expected 'File updated', got %q", result)
	}
}

func TestSummarizeResult_Grep(t *testing.T) {
	tests := []struct {
		output   string
		expected string
	}{
		{"", "No matches"},
		{"match1\nmatch2\nmatch3", "2 matches"},
	}

	for _, tt := range tests {
		result := summarizeResult("grep", tt.output)
		if result != tt.expected {
			t.Errorf("summarizeResult('grep', %q) = %q, want %q", tt.output, result, tt.expected)
		}
	}
}

func TestSummarizeResult_Unknown(t *testing.T) {
	// Unknown tool should use default summarizer
	result := summarizeResult("unknown_tool", "some output")
	if result != "some output" {
		t.Errorf("expected 'some output', got %q", result)
	}
}

func TestSummarizeInput(t *testing.T) {
	// Short input should pass through
	result := summarizeInput("bash", "ls -la")
	if result != "ls -la" {
		t.Errorf("expected 'ls -la', got %q", result)
	}

	// Long input should be truncated
	longInput := strings.Repeat("x", 100)
	result = summarizeInput("bash", longInput)
	if len(result) > 53 {
		t.Errorf("expected truncated input, got length %d", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("expected ... suffix for truncated input")
	}
}

func TestFormatToolCall(t *testing.T) {
	result := FormatToolCall("view", "src/main.go", "100 lines of code", true)
	if !strings.Contains(result, "view") {
		t.Error("expected tool name in output")
	}
	if !strings.Contains(result, "src/main.go") {
		t.Error("expected input in output")
	}
}

func TestFormatToolCallRunning(t *testing.T) {
	result := FormatToolCallRunning("edit", "src/main.go")
	if !strings.Contains(result, "edit") {
		t.Error("expected tool name in output")
	}
	if !strings.Contains(result, "⋯") {
		t.Error("expected running indicator in output")
	}
}

func TestToolDisplay_Depth(t *testing.T) {
	display := NewToolDisplay("delegate", "@reviewer")
	display.depth = 1

	view := display.View()
	// Should have more indentation
	if !strings.HasPrefix(view, "    ") {
		t.Error("expected increased indentation for nested tool")
	}
}
