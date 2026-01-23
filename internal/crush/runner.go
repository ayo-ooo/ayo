package crush

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// DefaultTimeout is the default timeout for Crush execution.
const DefaultTimeout = 10 * time.Minute

// Runner executes Crush commands.
type Runner struct {
	binaryPath string
}

// NewRunner creates a new Crush runner.
// Returns an error if Crush is not installed or version is incompatible.
func NewRunner(ctx context.Context) (*Runner, error) {
	info, err := FindBinary(ctx)
	if err != nil {
		return nil, err
	}
	return &Runner{binaryPath: info.Path}, nil
}

// NewRunnerWithPath creates a runner with an explicit binary path.
// Useful for testing or when the path is already known.
func NewRunnerWithPath(path string) *Runner {
	return &Runner{binaryPath: path}
}

// RunOptions configures a Crush execution.
type RunOptions struct {
	// Model specifies the model to use (e.g., "claude-3.5-sonnet" or "openrouter/claude-3.5-sonnet")
	Model string

	// SmallModel specifies the small model for auxiliary tasks
	SmallModel string

	// WorkingDir is the directory to run Crush in
	WorkingDir string

	// Quiet hides the spinner output
	Quiet bool

	// Stdin is optional content to pipe to Crush's stdin
	Stdin string

	// OnStdout is called for each line of stdout (streaming)
	OnStdout func(line string)

	// OnStderr is called for each line of stderr (progress/spinner)
	OnStderr func(line string)

	// Env is additional environment variables to set
	Env map[string]string

	// Timeout overrides the default timeout
	Timeout time.Duration
}

// RunResult contains the output from a Crush execution.
type RunResult struct {
	// Stdout is the complete stdout output
	Stdout string

	// Stderr is the complete stderr output
	Stderr string

	// ExitCode is the process exit code
	ExitCode int

	// Duration is how long the execution took
	Duration time.Duration

	// TimedOut indicates if the execution timed out
	TimedOut bool

	// Cancelled indicates if the execution was cancelled
	Cancelled bool
}

// Run executes Crush with the given prompt and options.
func (r *Runner) Run(ctx context.Context, prompt string, opts RunOptions) (*RunResult, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	// Apply timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build command
	args := r.buildArgs(prompt, opts)
	cmd := exec.CommandContext(ctx, r.binaryPath, args...)

	// Set working directory
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	}

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range opts.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set up stdin if provided
	if opts.Stdin != "" {
		cmd.Stdin = strings.NewReader(opts.Stdin)
	}

	// Set up stdout/stderr capture with optional streaming
	var stdoutBuf, stderrBuf bytes.Buffer
	var stdoutWriter, stderrWriter io.Writer = &stdoutBuf, &stderrBuf

	if opts.OnStdout != nil {
		stdoutWriter = io.MultiWriter(&stdoutBuf, &lineWriter{callback: opts.OnStdout})
	}
	if opts.OnStderr != nil {
		stderrWriter = io.MultiWriter(&stderrBuf, &lineWriter{callback: opts.OnStderr})
	}

	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	// Execute
	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	result := &RunResult{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: duration,
	}

	// Handle errors
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.TimedOut = true
			result.ExitCode = -1
			return result, nil
		}
		if ctx.Err() == context.Canceled {
			result.Cancelled = true
			result.ExitCode = -1
			return result, nil
		}

		var exitErr *exec.ExitError
		if ok := isExitError(err, &exitErr); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}

		return nil, fmt.Errorf("failed to execute crush: %w", err)
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	return result, nil
}

// RunStream executes Crush and streams output line by line.
// This is a convenience method that sets up streaming callbacks.
func (r *Runner) RunStream(ctx context.Context, prompt string, opts RunOptions, onOutput func(line string, isStderr bool)) (*RunResult, error) {
	opts.OnStdout = func(line string) {
		onOutput(line, false)
	}
	opts.OnStderr = func(line string) {
		onOutput(line, true)
	}
	return r.Run(ctx, prompt, opts)
}

// buildArgs constructs the command line arguments for 'crush run'.
func (r *Runner) buildArgs(prompt string, opts RunOptions) []string {
	args := []string{"run"}

	if opts.Quiet {
		args = append(args, "--quiet")
	}

	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	if opts.SmallModel != "" {
		args = append(args, "--small-model", opts.SmallModel)
	}

	// Prompt goes last
	args = append(args, prompt)

	return args
}

// isExitError checks if an error is an exec.ExitError and sets the target.
func isExitError(err error, target **exec.ExitError) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		*target = exitErr
		return true
	}
	return false
}

// lineWriter writes to a callback line by line.
type lineWriter struct {
	callback func(string)
	mu       sync.Mutex
	buf      bytes.Buffer
}

func (w *lineWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err = w.buf.Write(p)
	if err != nil {
		return n, err
	}

	// Process complete lines
	for {
		line, readErr := w.buf.ReadString('\n')
		if readErr != nil {
			// No complete line yet, put partial back
			if line != "" {
				// Create a new buffer with the partial line
				var newBuf bytes.Buffer
				newBuf.WriteString(line)
				w.buf = newBuf
			}
			break
		}
		// Call callback with complete line (trim newline)
		w.callback(strings.TrimSuffix(line, "\n"))
	}

	return n, nil
}

// Flush sends any remaining buffered content to the callback.
func (w *lineWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buf.Len() > 0 {
		w.callback(w.buf.String())
		w.buf.Reset()
	}
}

// StreamingRunner provides a higher-level interface for streaming Crush output.
type StreamingRunner struct {
	runner *Runner
}

// NewStreamingRunner creates a streaming runner.
func NewStreamingRunner(ctx context.Context) (*StreamingRunner, error) {
	runner, err := NewRunner(ctx)
	if err != nil {
		return nil, err
	}
	return &StreamingRunner{runner: runner}, nil
}

// RunWithCallbacks executes Crush with separate callbacks for different output types.
type OutputCallbacks struct {
	OnText     func(text string)   // Called for text content
	OnProgress func(status string) // Called for progress updates (stderr)
	OnError    func(err string)    // Called for error output
}

// Run executes Crush with streaming callbacks.
func (sr *StreamingRunner) Run(ctx context.Context, prompt string, opts RunOptions, callbacks OutputCallbacks) (*RunResult, error) {
	// Set up streaming
	opts.OnStdout = func(line string) {
		if callbacks.OnText != nil {
			callbacks.OnText(line)
		}
	}
	opts.OnStderr = func(line string) {
		if callbacks.OnProgress != nil {
			callbacks.OnProgress(line)
		}
	}

	result, err := sr.runner.Run(ctx, prompt, opts)
	if err != nil {
		if callbacks.OnError != nil {
			callbacks.OnError(err.Error())
		}
		return nil, err
	}

	// Check for non-zero exit
	if result.ExitCode != 0 && callbacks.OnError != nil {
		callbacks.OnError(fmt.Sprintf("crush exited with code %d", result.ExitCode))
	}

	return result, nil
}

// lineScanner wraps a reader and provides line-by-line streaming.
type lineScanner struct {
	reader  io.Reader
	scanner *bufio.Scanner
}

func newLineScanner(r io.Reader) *lineScanner {
	return &lineScanner{
		reader:  r,
		scanner: bufio.NewScanner(r),
	}
}

func (ls *lineScanner) Scan() bool {
	return ls.scanner.Scan()
}

func (ls *lineScanner) Text() string {
	return ls.scanner.Text()
}

func (ls *lineScanner) Err() error {
	return ls.scanner.Err()
}
