package crush

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunOptions_Defaults(t *testing.T) {
	opts := RunOptions{}

	// Verify defaults
	if opts.Timeout != 0 {
		t.Errorf("expected zero timeout (will use default), got %v", opts.Timeout)
	}
	if opts.Quiet {
		t.Error("expected Quiet to be false by default")
	}
}

func TestRunner_BuildArgs(t *testing.T) {
	runner := &Runner{binaryPath: "/usr/bin/crush"}

	tests := []struct {
		name   string
		prompt string
		opts   RunOptions
		want   []string
	}{
		{
			name:   "basic prompt",
			prompt: "hello world",
			opts:   RunOptions{},
			want:   []string{"run", "hello world"},
		},
		{
			name:   "with quiet",
			prompt: "fix bug",
			opts:   RunOptions{Quiet: true},
			want:   []string{"run", "--quiet", "fix bug"},
		},
		{
			name:   "with model",
			prompt: "refactor code",
			opts:   RunOptions{Model: "claude-3.5-sonnet"},
			want:   []string{"run", "--model", "claude-3.5-sonnet", "refactor code"},
		},
		{
			name:   "with provider/model",
			prompt: "add tests",
			opts:   RunOptions{Model: "openrouter/claude-3.5-sonnet"},
			want:   []string{"run", "--model", "openrouter/claude-3.5-sonnet", "add tests"},
		},
		{
			name:   "with small model",
			prompt: "review code",
			opts:   RunOptions{SmallModel: "gpt-4o-mini"},
			want:   []string{"run", "--small-model", "gpt-4o-mini", "review code"},
		},
		{
			name:   "all options",
			prompt: "complex task",
			opts: RunOptions{
				Quiet:      true,
				Model:      "claude-3.5-sonnet",
				SmallModel: "gpt-4o-mini",
			},
			want: []string{"run", "--quiet", "--model", "claude-3.5-sonnet", "--small-model", "gpt-4o-mini", "complex task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runner.buildArgs(tt.prompt, tt.opts)
			if len(got) != len(tt.want) {
				t.Errorf("buildArgs() length = %d, want %d\ngot:  %v\nwant: %v", len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("buildArgs()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRunner_Run_EmptyPrompt(t *testing.T) {
	runner := &Runner{binaryPath: "/usr/bin/echo"}

	_, err := runner.Run(context.Background(), "", RunOptions{})
	if err == nil {
		t.Error("expected error for empty prompt")
	}
	if !strings.Contains(err.Error(), "prompt is required") {
		t.Errorf("expected 'prompt is required' error, got: %v", err)
	}
}

func TestRunner_Run_WithMockBinary(t *testing.T) {
	// Create a mock script that echoes the input
	tmpDir := t.TempDir()
	mockBinary := filepath.Join(tmpDir, "mock-crush")

	script := `#!/bin/sh
echo "Mock output for: $@"
`
	if err := os.WriteFile(mockBinary, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock binary: %v", err)
	}

	runner := NewRunnerWithPath(mockBinary)
	result, err := runner.Run(context.Background(), "test prompt", RunOptions{})

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "Mock output") {
		t.Errorf("expected stdout to contain 'Mock output', got: %q", result.Stdout)
	}
}

func TestRunner_Run_WithWorkingDir(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinary := filepath.Join(tmpDir, "mock-crush")

	// Script that prints current directory
	script := `#!/bin/sh
pwd
`
	if err := os.WriteFile(mockBinary, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock binary: %v", err)
	}

	runner := NewRunnerWithPath(mockBinary)
	result, err := runner.Run(context.Background(), "test", RunOptions{
		WorkingDir: tmpDir,
	})

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// The output should contain the temp directory path
	if !strings.Contains(result.Stdout, tmpDir) {
		t.Errorf("expected working dir in output, got: %q", result.Stdout)
	}
}

func TestRunner_Run_WithStdin(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinary := filepath.Join(tmpDir, "mock-crush")

	// Script that reads and echoes stdin
	script := `#!/bin/sh
cat
`
	if err := os.WriteFile(mockBinary, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock binary: %v", err)
	}

	runner := NewRunnerWithPath(mockBinary)
	result, err := runner.Run(context.Background(), "ignored", RunOptions{
		Stdin: "piped content here",
	})

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(result.Stdout, "piped content here") {
		t.Errorf("expected stdin content in output, got: %q", result.Stdout)
	}
}

func TestRunner_Run_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinary := filepath.Join(tmpDir, "mock-crush")

	// Script that sleeps longer than timeout
	script := `#!/bin/sh
sleep 10
`
	if err := os.WriteFile(mockBinary, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock binary: %v", err)
	}

	runner := NewRunnerWithPath(mockBinary)
	result, err := runner.Run(context.Background(), "test", RunOptions{
		Timeout: 100 * time.Millisecond,
	})

	if err != nil {
		t.Fatalf("Run() should not return error on timeout, got: %v", err)
	}

	if !result.TimedOut {
		t.Error("expected TimedOut to be true")
	}

	if result.ExitCode != -1 {
		t.Errorf("expected exit code -1 for timeout, got %d", result.ExitCode)
	}
}

func TestRunner_Run_Cancellation(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinary := filepath.Join(tmpDir, "mock-crush")

	// Script that sleeps
	script := `#!/bin/sh
sleep 10
`
	if err := os.WriteFile(mockBinary, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock binary: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	runner := NewRunnerWithPath(mockBinary)

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	result, err := runner.Run(ctx, "test", RunOptions{})

	if err != nil {
		t.Fatalf("Run() should not return error on cancel, got: %v", err)
	}

	if !result.Cancelled {
		t.Error("expected Cancelled to be true")
	}
}

func TestRunner_Run_Streaming(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinary := filepath.Join(tmpDir, "mock-crush")

	// Script that outputs multiple lines
	script := `#!/bin/sh
echo "line 1"
echo "line 2"
echo "line 3"
`
	if err := os.WriteFile(mockBinary, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock binary: %v", err)
	}

	runner := NewRunnerWithPath(mockBinary)

	var lines []string
	result, err := runner.Run(context.Background(), "test", RunOptions{
		OnStdout: func(line string) {
			lines = append(lines, line)
		},
	})

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}
}

func TestRunner_Run_NonZeroExit(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinary := filepath.Join(tmpDir, "mock-crush")

	// Script that exits with error
	script := `#!/bin/sh
echo "error output" >&2
exit 42
`
	if err := os.WriteFile(mockBinary, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock binary: %v", err)
	}

	runner := NewRunnerWithPath(mockBinary)
	result, err := runner.Run(context.Background(), "test", RunOptions{})

	if err != nil {
		t.Fatalf("Run() should not return error on non-zero exit, got: %v", err)
	}

	if result.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stderr, "error output") {
		t.Errorf("expected stderr to contain error output, got: %q", result.Stderr)
	}
}

func TestRunner_Run_BinaryNotFound(t *testing.T) {
	runner := NewRunnerWithPath("/nonexistent/path/to/crush")
	_, err := runner.Run(context.Background(), "test", RunOptions{})

	if err == nil {
		t.Error("expected error for nonexistent binary")
	}
}

func TestLineWriter(t *testing.T) {
	var lines []string
	w := &lineWriter{callback: func(line string) {
		lines = append(lines, line)
	}}

	// Write partial line
	w.Write([]byte("hello"))
	if len(lines) != 0 {
		t.Errorf("expected no lines yet, got %d", len(lines))
	}

	// Complete the line
	w.Write([]byte(" world\n"))
	if len(lines) != 1 || lines[0] != "hello world" {
		t.Errorf("expected ['hello world'], got %v", lines)
	}

	// Write multiple lines at once
	w.Write([]byte("line2\nline3\npartial"))
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}

	// Flush remaining
	w.Flush()
	if len(lines) != 4 || lines[3] != "partial" {
		t.Errorf("expected 4 lines ending with 'partial', got %v", lines)
	}
}

func TestNewRunner_WithRealCrush(t *testing.T) {
	// Skip if crush is not installed
	if _, err := exec.LookPath("crush"); err != nil {
		t.Skip("crush not installed, skipping real binary test")
	}

	ctx := context.Background()
	runner, err := NewRunner(ctx)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	if runner.binaryPath == "" {
		t.Error("expected non-empty binary path")
	}
}
