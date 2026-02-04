// Package sandbox provides container-based agent execution environments.
// It implements the SandboxProvider interface with multiple backends.
package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/google/uuid"
)

// NoneProvider executes commands directly on the host system (no isolation).
// This is the default provider when no container runtime is available.
type NoneProvider struct {
	sandboxes map[string]*noneSandbox
}

type noneSandbox struct {
	id        string
	name      string
	status    providers.SandboxStatus
	createdAt time.Time
	pool      string
	agents    []string
}

// NewNoneProvider creates a new none sandbox provider.
func NewNoneProvider() *NoneProvider {
	return &NoneProvider{
		sandboxes: make(map[string]*noneSandbox),
	}
}

func (p *NoneProvider) Name() string                   { return "none" }
func (p *NoneProvider) Type() providers.ProviderType  { return providers.ProviderTypeSandbox }
func (p *NoneProvider) Init(_ context.Context, _ map[string]any) error { return nil }
func (p *NoneProvider) Close() error                  { return nil }

// Create creates a virtual sandbox (just tracks state, no actual container).
func (p *NoneProvider) Create(ctx context.Context, opts providers.SandboxCreateOptions) (providers.Sandbox, error) {
	id := uuid.New().String()[:8]
	name := opts.Name
	if name == "" {
		name = "sandbox-" + id
	}

	sb := &noneSandbox{
		id:        id,
		name:      name,
		status:    providers.SandboxStatusRunning,
		createdAt: time.Now(),
		pool:      opts.Pool,
		agents:    make([]string, 0),
	}
	p.sandboxes[id] = sb

	return p.sandboxToProviders(sb), nil
}

// Get retrieves a sandbox by ID.
func (p *NoneProvider) Get(ctx context.Context, id string) (providers.Sandbox, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return providers.Sandbox{}, fmt.Errorf("sandbox not found: %s", id)
	}
	return p.sandboxToProviders(sb), nil
}

// List returns all sandboxes.
func (p *NoneProvider) List(ctx context.Context) ([]providers.Sandbox, error) {
	result := make([]providers.Sandbox, 0, len(p.sandboxes))
	for _, sb := range p.sandboxes {
		result = append(result, p.sandboxToProviders(sb))
	}
	return result, nil
}

// Start is a no-op for none provider (already "running").
func (p *NoneProvider) Start(ctx context.Context, id string) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}
	sb.status = providers.SandboxStatusRunning
	return nil
}

// Stop marks the sandbox as stopped.
func (p *NoneProvider) Stop(ctx context.Context, id string, opts providers.SandboxStopOptions) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}
	sb.status = providers.SandboxStatusStopped
	return nil
}

// Delete removes a sandbox.
func (p *NoneProvider) Delete(ctx context.Context, id string, force bool) error {
	if _, ok := p.sandboxes[id]; !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}
	delete(p.sandboxes, id)
	return nil
}

// Exec executes a command directly on the host (no isolation).
func (p *NoneProvider) Exec(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
	_, ok := p.sandboxes[id]
	if !ok {
		return providers.ExecResult{}, fmt.Errorf("sandbox not found: %s", id)
	}

	start := time.Now()

	// Build command
	var cmd *exec.Cmd
	if len(opts.Args) > 0 {
		cmd = exec.CommandContext(ctx, opts.Command, opts.Args...)
	} else {
		// Use shell to execute command string
		cmd = exec.CommandContext(ctx, "sh", "-c", opts.Command)
	}

	// Set working directory
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	}

	// Set environment
	if len(opts.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range opts.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	// Set stdin if provided
	if len(opts.Stdin) > 0 {
		cmd.Stdin = strings.NewReader(string(opts.Stdin))
	}

	// Run with timeout
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Create a context with timeout if specified
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
		cmd.Dir = opts.WorkingDir
		cmd.Env = cmd.Env
		cmd.Stdin = cmd.Stdin
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	duration := time.Since(start)

	result := providers.ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = -1
		return result, nil
	}

	// Get exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
			result.Stderr = err.Error()
		}
	}

	return result, nil
}

// Status returns the current status of a sandbox.
func (p *NoneProvider) Status(ctx context.Context, id string) (providers.SandboxStatus, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return "", fmt.Errorf("sandbox not found: %s", id)
	}
	return sb.status, nil
}

func (p *NoneProvider) sandboxToProviders(sb *noneSandbox) providers.Sandbox {
	return providers.Sandbox{
		ID:        sb.id,
		Name:      sb.name,
		Image:     "none",
		Status:    sb.status,
		Pool:      sb.pool,
		Agents:    sb.agents,
		CreatedAt: sb.createdAt,
	}
}

// AssignAgent assigns an agent to a sandbox.
func (p *NoneProvider) AssignAgent(id, agentHandle string) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}
	sb.agents = append(sb.agents, agentHandle)
	return nil
}

// Verify NoneProvider implements SandboxProvider.
var _ providers.SandboxProvider = (*NoneProvider)(nil)
