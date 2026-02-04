// Package sandbox provides container-based agent execution environments.
package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/google/uuid"
)

// AppleProvider executes commands using Apple Container on macOS 15+.
// Apple Container provides native Linux container support with:
// - Native virtualization (faster startup than Docker)
// - Lower resource usage
// - Optimized for Apple Silicon
// - virtiofs for fast file sharing
type AppleProvider struct {
	sandboxes map[string]*appleSandbox
	available bool
}

type appleSandbox struct {
	id          string
	name        string
	containerID string
	status      providers.SandboxStatus
	createdAt   time.Time
	pool        string
	agents      []string
	image       string
	mounts      []providers.Mount
	network     providers.NetworkConfig
}

// NewAppleProvider creates a new Apple Container sandbox provider.
func NewAppleProvider() *AppleProvider {
	return &AppleProvider{
		sandboxes: make(map[string]*appleSandbox),
		available: isAppleContainerAvailable(),
	}
}

func (p *AppleProvider) Name() string                  { return "apple" }
func (p *AppleProvider) Type() providers.ProviderType { return providers.ProviderTypeSandbox }

func (p *AppleProvider) Init(_ context.Context, _ map[string]any) error {
	if !p.available {
		return fmt.Errorf("Apple Container is not available (requires macOS 15+ on Apple Silicon)")
	}
	return nil
}

func (p *AppleProvider) Close() error {
	// Stop and remove all containers
	for id := range p.sandboxes {
		_ = p.Delete(context.Background(), id, true)
	}
	return nil
}

// IsAvailable returns whether Apple Container is available on the system.
func (p *AppleProvider) IsAvailable() bool {
	return p.available
}

// Create creates a new Apple Container sandbox.
func (p *AppleProvider) Create(ctx context.Context, opts providers.SandboxCreateOptions) (providers.Sandbox, error) {
	if !p.available {
		return providers.Sandbox{}, fmt.Errorf("Apple Container is not available")
	}

	id := uuid.New().String()[:8]
	name := opts.Name
	if name == "" {
		name = "ayo-sandbox-" + id
	}

	image := opts.Image
	if image == "" {
		image = "docker.io/library/busybox:latest"
	}

	debug.Log("creating apple container sandbox", "id", id, "name", name, "image", image)

	// Build container create command
	args := []string{"create", "--name", name}

	// Add mounts (using virtiofs for better performance)
	for _, m := range opts.Mounts {
		mountOpt := fmt.Sprintf("%s:%s", m.Source, m.Destination)
		if m.ReadOnly {
			mountOpt += ":ro"
		}
		args = append(args, "--volume", mountOpt)
	}

	// Network mode - Apple Container uses --network flag
	if !opts.Network.Enabled {
		args = append(args, "--network", "none")
	}

	// Resource limits
	if opts.Resources.CPUs > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%d", opts.Resources.CPUs))
	}
	if opts.Resources.MemoryMB > 0 {
		args = append(args, "--memory", fmt.Sprintf("%dM", opts.Resources.MemoryMB))
	}

	// Add image
	args = append(args, image)

	// Create container
	cmd := exec.CommandContext(ctx, "container", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		debug.Log("container create failed", "error", err, "stderr", stderr.String())
		return providers.Sandbox{}, fmt.Errorf("container create failed: %w: %s", err, stderr.String())
	}

	// Start the container
	startCmd := exec.CommandContext(ctx, "container", "start", name)
	startCmd.Stdout = &stdout
	startCmd.Stderr = &stderr

	if err := startCmd.Run(); err != nil {
		debug.Log("container start failed", "error", err, "stderr", stderr.String())
		// Clean up the created container
		_ = exec.CommandContext(ctx, "container", "rm", "-f", name).Run()
		return providers.Sandbox{}, fmt.Errorf("container start failed: %w: %s", err, stderr.String())
	}

	containerID := name // Apple Container uses names as IDs
	debug.Log("apple container created", "containerID", containerID)

	sb := &appleSandbox{
		id:          id,
		name:        name,
		containerID: containerID,
		status:      providers.SandboxStatusRunning,
		createdAt:   time.Now(),
		pool:        opts.Pool,
		agents:      make([]string, 0),
		image:       image,
		mounts:      opts.Mounts,
		network:     opts.Network,
	}
	p.sandboxes[id] = sb

	return p.sandboxToProviders(sb), nil
}

// Get retrieves a sandbox by ID.
func (p *AppleProvider) Get(ctx context.Context, id string) (providers.Sandbox, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return providers.Sandbox{}, fmt.Errorf("sandbox not found: %s", id)
	}
	return p.sandboxToProviders(sb), nil
}

// List returns all sandboxes.
func (p *AppleProvider) List(ctx context.Context) ([]providers.Sandbox, error) {
	result := make([]providers.Sandbox, 0, len(p.sandboxes))
	for _, sb := range p.sandboxes {
		result = append(result, p.sandboxToProviders(sb))
	}
	return result, nil
}

// Start starts a stopped container.
func (p *AppleProvider) Start(ctx context.Context, id string) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	cmd := exec.CommandContext(ctx, "container", "start", sb.containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("container start failed: %w", err)
	}

	sb.status = providers.SandboxStatusRunning
	return nil
}

// Stop stops a running container.
func (p *AppleProvider) Stop(ctx context.Context, id string, opts providers.SandboxStopOptions) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	args := []string{"stop"}
	if opts.Timeout > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", int(opts.Timeout.Seconds())))
	}
	args = append(args, sb.containerID)

	cmd := exec.CommandContext(ctx, "container", args...)
	if err := cmd.Run(); err != nil {
		// Try force kill as fallback
		_ = exec.CommandContext(ctx, "container", "kill", sb.containerID).Run()
	}

	sb.status = providers.SandboxStatusStopped
	return nil
}

// Delete removes a container.
func (p *AppleProvider) Delete(ctx context.Context, id string, force bool) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, sb.containerID)

	cmd := exec.CommandContext(ctx, "container", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("container rm failed: %w", err)
	}

	delete(p.sandboxes, id)
	return nil
}

// Exec executes a command in the container.
func (p *AppleProvider) Exec(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return providers.ExecResult{}, fmt.Errorf("sandbox not found: %s", id)
	}

	start := time.Now()

	// Build container exec command
	args := []string{"exec"}

	// Working directory
	if opts.WorkingDir != "" {
		args = append(args, "-w", opts.WorkingDir)
	}

	// Environment variables
	for k, v := range opts.Env {
		args = append(args, "-e", k+"="+v)
	}

	args = append(args, sb.containerID)

	// Add the command
	if len(opts.Args) > 0 {
		args = append(args, opts.Command)
		args = append(args, opts.Args...)
	} else {
		// Use shell to execute command string
		args = append(args, "sh", "-c", opts.Command)
	}

	// Create command with timeout
	execCtx := ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(execCtx, "container", args...)

	// Set stdin if provided
	if len(opts.Stdin) > 0 {
		cmd.Stdin = strings.NewReader(string(opts.Stdin))
		args = append([]string{"exec", "-i"}, args[1:]...)
		cmd = exec.CommandContext(execCtx, "container", args...)
		cmd.Stdin = strings.NewReader(string(opts.Stdin))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := providers.ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	// Check for timeout
	if execCtx.Err() == context.DeadlineExceeded {
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
func (p *AppleProvider) Status(ctx context.Context, id string) (providers.SandboxStatus, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return "", fmt.Errorf("sandbox not found: %s", id)
	}

	// Check actual container status using container inspect
	cmd := exec.CommandContext(ctx, "container", "inspect", sb.containerID)
	output, err := cmd.Output()
	if err != nil {
		return sb.status, nil // Fall back to cached status
	}

	// Parse status from output - Apple Container outputs JSON
	outputStr := string(output)
	switch {
	case strings.Contains(outputStr, `"running"`):
		sb.status = providers.SandboxStatusRunning
	case strings.Contains(outputStr, `"exited"`), strings.Contains(outputStr, `"stopped"`):
		sb.status = providers.SandboxStatusStopped
	case strings.Contains(outputStr, `"created"`):
		sb.status = providers.SandboxStatusCreating
	default:
		// Keep existing status if we can't parse
	}

	return sb.status, nil
}

func (p *AppleProvider) sandboxToProviders(sb *appleSandbox) providers.Sandbox {
	return providers.Sandbox{
		ID:        sb.id,
		Name:      sb.name,
		Image:     sb.image,
		Status:    sb.status,
		Pool:      sb.pool,
		Agents:    sb.agents,
		Mounts:    sb.mounts,
		CreatedAt: sb.createdAt,
	}
}

// AssignAgent assigns an agent to a sandbox.
func (p *AppleProvider) AssignAgent(id, agentHandle string) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}
	sb.agents = append(sb.agents, agentHandle)
	return nil
}

// isAppleContainerAvailable checks if Apple Container is available on the system.
// Requires macOS 15+ on Apple Silicon.
func isAppleContainerAvailable() bool {
	// Check if we're on macOS
	if runtime.GOOS != "darwin" {
		return false
	}

	// Check if we're on Apple Silicon (ARM64)
	if runtime.GOARCH != "arm64" {
		return false
	}

	// Check if container command exists
	cmd := exec.Command("container", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}

	// Check if container service is running
	cmd = exec.Command("container", "system", "status")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

// Verify AppleProvider implements SandboxProvider.
var _ providers.SandboxProvider = (*AppleProvider)(nil)
