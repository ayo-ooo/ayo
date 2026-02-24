package interactive

import (
	"bytes"
	"strings"
	"testing"
)

func TestStreamRenderer_WriteToken_BuffersPartialLines(t *testing.T) {
	var out bytes.Buffer
	r := NewStreamRenderer(&out, 80)

	// Write partial line - should be buffered, not output
	if err := r.WriteToken("hello"); err != nil {
		t.Fatalf("WriteToken failed: %v", err)
	}

	if out.Len() != 0 {
		t.Errorf("expected no output for partial line, got %q", out.String())
	}

	if r.Buffered() != "hello" {
		t.Errorf("expected buffer to contain 'hello', got %q", r.Buffered())
	}
}

func TestStreamRenderer_WriteToken_RendersCompleteLines(t *testing.T) {
	var out bytes.Buffer
	r := NewStreamRenderer(&out, 80)

	// Write complete line
	if err := r.WriteToken("hello\n"); err != nil {
		t.Fatalf("WriteToken failed: %v", err)
	}

	// Should have output (with glamour rendering)
	if out.Len() == 0 {
		t.Error("expected output for complete line")
	}

	// Buffer should be empty after newline
	if r.Buffered() != "" {
		t.Errorf("expected empty buffer after newline, got %q", r.Buffered())
	}
}

func TestStreamRenderer_WriteToken_MultipleLines(t *testing.T) {
	var out bytes.Buffer
	r := NewStreamRenderer(&out, 80)

	// Write multiple lines at once
	if err := r.WriteToken("line1\nline2\n"); err != nil {
		t.Fatalf("WriteToken failed: %v", err)
	}

	// Should have output
	if out.Len() == 0 {
		t.Error("expected output for complete lines")
	}

	// Buffer should be empty
	if r.Buffered() != "" {
		t.Errorf("expected empty buffer, got %q", r.Buffered())
	}
}

func TestStreamRenderer_WriteToken_PartialAfterComplete(t *testing.T) {
	var out bytes.Buffer
	r := NewStreamRenderer(&out, 80)

	// Write line with trailing partial
	if err := r.WriteToken("line1\npartial"); err != nil {
		t.Fatalf("WriteToken failed: %v", err)
	}

	// Should have output for line1
	if out.Len() == 0 {
		t.Error("expected output for complete line")
	}

	// Buffer should contain "partial"
	if r.Buffered() != "partial" {
		t.Errorf("expected buffer to contain 'partial', got %q", r.Buffered())
	}
}

func TestStreamRenderer_Flush_RendersRemaining(t *testing.T) {
	var out bytes.Buffer
	r := NewStreamRenderer(&out, 80)

	// Write partial line
	if err := r.WriteToken("remaining content"); err != nil {
		t.Fatalf("WriteToken failed: %v", err)
	}

	// Flush
	if err := r.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Should have output
	if out.Len() == 0 {
		t.Error("expected output after flush")
	}

	// Buffer should be empty
	if r.Buffered() != "" {
		t.Errorf("expected empty buffer after flush, got %q", r.Buffered())
	}
}

func TestStreamRenderer_Flush_EmptyBuffer(t *testing.T) {
	var out bytes.Buffer
	r := NewStreamRenderer(&out, 80)

	// Flush empty buffer
	if err := r.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Should have no output
	if out.Len() != 0 {
		t.Errorf("expected no output for empty flush, got %q", out.String())
	}
}

func TestStreamRenderer_Reset(t *testing.T) {
	var out bytes.Buffer
	r := NewStreamRenderer(&out, 80)

	// Write partial line
	if err := r.WriteToken("hello"); err != nil {
		t.Fatalf("WriteToken failed: %v", err)
	}

	// Reset
	r.Reset()

	// Buffer should be empty
	if r.Buffered() != "" {
		t.Errorf("expected empty buffer after reset, got %q", r.Buffered())
	}

	// Should have no output (reset discards buffer)
	if out.Len() != 0 {
		t.Errorf("expected no output after reset, got %q", out.String())
	}
}

func TestStreamRenderer_CodeBlock(t *testing.T) {
	var out bytes.Buffer
	r := NewStreamRenderer(&out, 80)

	// Write code block
	code := "```go\nfunc main() {}\n```\n"
	if err := r.WriteToken(code); err != nil {
		t.Fatalf("WriteToken failed: %v", err)
	}

	// Should have rendered output
	if out.Len() == 0 {
		t.Error("expected output for code block")
	}

	// Output should contain the code (glamour adds ANSI codes around each character)
	// Check for "func" and "main" separately as they're wrapped in ANSI codes
	output := out.String()
	if !strings.Contains(output, "func") {
		t.Errorf("expected output to contain 'func', got %q", output)
	}
	if !strings.Contains(output, "main") {
		t.Errorf("expected output to contain 'main', got %q", output)
	}
}

func TestStreamRenderer_IncrementalTokens(t *testing.T) {
	var out bytes.Buffer
	r := NewStreamRenderer(&out, 80)

	// Simulate token-by-token streaming
	tokens := []string{"He", "llo", " ", "wo", "rld", "!\n"}
	for _, token := range tokens {
		if err := r.WriteToken(token); err != nil {
			t.Fatalf("WriteToken failed: %v", err)
		}
	}

	// Should have output after the newline token
	if out.Len() == 0 {
		t.Error("expected output after complete line")
	}

	// Buffer should be empty
	if r.Buffered() != "" {
		t.Errorf("expected empty buffer, got %q", r.Buffered())
	}
}

func TestStreamRenderer_WidthClamping(t *testing.T) {
	var out bytes.Buffer

	// Very small width should be clamped to 40
	r := NewStreamRenderer(&out, 10)
	if r.width != 40 {
		t.Errorf("expected width clamped to 40, got %d", r.width)
	}

	// Very large width should be clamped to 200
	r = NewStreamRenderer(&out, 500)
	if r.width != 200 {
		t.Errorf("expected width clamped to 200, got %d", r.width)
	}

	// Normal width should be preserved
	r = NewStreamRenderer(&out, 100)
	if r.width != 100 {
		t.Errorf("expected width 100, got %d", r.width)
	}
}
