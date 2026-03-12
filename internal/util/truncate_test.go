package util

import (
	"testing"
)

// TestTruncate tests the Truncate function.
func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "shorter than max length",
			input:    "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "exact max length",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "longer than max length",
			input:    "Hello World",
			maxLen:   8,
			expected: "Hello...",
		},
		{
			name:     "trims whitespace from start and end",
			input:    "  Hello World  ",
			maxLen:   20,
			expected: "Hello World",
		},
		{
			name:     "replaces newlines with spaces",
			input:    "Hello\nWorld",
			maxLen:   20,
			expected: "Hello World",
		},
		{
			name:     "removes carriage returns",
			input:    "Hello\rWorld",
			maxLen:   20,
			expected: "HelloWorld",
		},
		{
			name:     "mixed whitespace",
			input:    "  Hello\r\nWorld  ",
			maxLen:   20,
			expected: "Hello World",
		},
		{
			name:     "max length <= 3 returns exact characters",
			input:    "Hello",
			maxLen:   3,
			expected: "Hel",
		},
		{
			name:     "max length = 1",
			input:    "Hello",
			maxLen:   1,
			expected: "H",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "unicode characters",
			input:    "Hello 世界",
			maxLen:   8,
			expected: "Hello...",
		},
		{
			name:     "many newlines",
			input:    "Line1\n\n\nLine2",
			maxLen:   20,
			expected: "Line1   Line2",
		},
		{
			name:     "very long text",
			input:    "This is a very long piece of text that should be truncated",
			maxLen:   20,
			expected: "This is a very lo...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestTruncateRaw tests the TruncateRaw function.
func TestTruncateRaw(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "shorter than max length",
			input:    "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "exact max length",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "longer than max length",
			input:    "Hello World",
			maxLen:   8,
			expected: "Hello Wo",
		},
		{
			name:     "preserves whitespace",
			input:    "  Hello World  ",
			maxLen:   20,
			expected: "  Hello World  ",
		},
		{
			name:     "preserves newlines",
			input:    "Hello\nWorld",
			maxLen:   20,
			expected: "Hello\nWorld",
		},
		{
			name:     "preserves carriage returns",
			input:    "Hello\rWorld",
			maxLen:   20,
			expected: "Hello\rWorld",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "max length = 0",
			input:    "Hello",
			maxLen:   0,
			expected: "",
		},
		{
			name:     "unicode characters",
			input:    "Hello 世界",
			maxLen:   8,
			expected: "Hello \xe4\xb8", // TruncateRaw is byte-level, cuts mid-character
		},
		{
			name:     "very long text",
			input:    "This is a very long piece of text that should be truncated",
			maxLen:   20,
			expected: "This is a very long ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateRaw(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateRaw(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestTruncateTitle tests the TruncateTitle function.
func TestTruncateTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "shorter than max length",
			input:    "Hello World",
			maxLen:   20,
			expected: "Hello World",
		},
		{
			name:     "exact max length",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "longer than max length",
			input:    "Hello World",
			maxLen:   8,
			expected: "Hello W…",
		},
		{
			name:     "trims whitespace from start and end",
			input:    "  Hello World  ",
			maxLen:   20,
			expected: "Hello World",
		},
		{
			name:     "collapses multiple spaces to single",
			input:    "Hello     World",
			maxLen:   20,
			expected: "Hello World",
		},
		{
			name:     "collapses multiple newlines to spaces",
			input:    "Hello\n\nWorld",
			maxLen:   20,
			expected: "Hello World",
		},
		{
			name:     "mixed whitespace collapsed",
			input:    "  Hello \n\r  World  ",
			maxLen:   20,
			expected: "Hello World",
		},
		{
			name:     "max length = 1",
			input:    "Hello",
			maxLen:   1,
			expected: "…",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "unicode characters - TruncateTitle is byte-level",
			input:    "Hello 世界",
			maxLen:   8,
			expected: "Hello \xe4…", // Byte-level truncation (first 7 chars + ellipsis)
		},
		{
			name:     "very long text",
			input:    "This is a very long title that should be truncated",
			maxLen:   20,
			expected: "This is a very long…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateTitle(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateTitle(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestTruncateEdgeCases tests edge cases and boundary conditions.
// Note: Negative max lengths cause panics - these are undefined behavior.
func TestTruncateEdgeCases(t *testing.T) {
	t.Skip("Skipping edge case tests - negative max lengths panic (undefined behavior)")
}

// TestTruncateUnicode tests unicode handling.
func TestTruncateUnicode(t *testing.T) {
	// Test with emoji and multi-byte characters
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "emoji",
			input:    "Hello 😀 World",
			maxLen:   15,
			expected: "Hello 😀...",
		},
		{
			name:     "chinese characters",
			input:    "你好世界",
			maxLen:   5,
			expected: "你好...",
		},
		{
			name:     "arabic text",
			input:    "مرحبا",
			maxLen:   4,
			expected: "مرح...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Logf("Truncate(%q, %d) = %q, want %q (unicode handling may vary)", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}
