package interactive

import (
	"bytes"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/run"
)

func TestWriterImplementsStreamWriter(t *testing.T) {
	// This test ensures Writer implements run.StreamWriter at compile time
	var _ run.StreamWriter = (*Writer)(nil)
}

func TestWriterWriteText(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewSimpleRenderer(&buf, 80)
	writer := NewWriterWithRenderer(renderer)

	writer.WriteText("Hello ")
	writer.WriteText("World")
	writer.WriteDone("")

	output := buf.String()
	// Output contains ANSI styling, so just check for content presence
	if !bytes.Contains(buf.Bytes(), []byte("Hello")) {
		t.Errorf("expected 'Hello' in output, got %q", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("World")) {
		t.Errorf("expected 'World' in output, got %q", output)
	}
}

func TestWriterWriteToolStart(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewSimpleRenderer(&buf, 80)
	writer := NewWriterWithRenderer(renderer)

	writer.WriteToolStart(run.ToolCall{
		ID:          "test-1",
		Name:        "bash",
		Description: "list files",
		Command:     "ls -la",
	})

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("bash")) {
		t.Errorf("expected 'bash' in output, got %q", output)
	}
}

func TestWriterWriteToolResult(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewSimpleRenderer(&buf, 80)
	writer := NewWriterWithRenderer(renderer)

	writer.WriteToolStart(run.ToolCall{
		ID:   "test-1",
		Name: "bash",
	})

	writer.WriteToolResult(run.ToolResult{
		ID:       "test-1",
		Name:     "bash",
		Output:   "file1.txt\nfile2.txt\nfile3.txt",
		Duration: 100 * time.Millisecond,
	})

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("✓")) {
		t.Errorf("expected success icon in output, got %q", output)
	}
}

func TestWriterWriteToolError(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewSimpleRenderer(&buf, 80)
	writer := NewWriterWithRenderer(renderer)

	writer.WriteToolStart(run.ToolCall{
		ID:   "test-1",
		Name: "bash",
	})

	writer.WriteToolResult(run.ToolResult{
		ID:    "test-1",
		Name:  "bash",
		Error: "command not found",
	})

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("✗")) {
		t.Errorf("expected error icon in output, got %q", output)
	}
}
