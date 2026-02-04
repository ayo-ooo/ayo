// Package sandbox provides container-based agent execution environments.
package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/google/uuid"
)

// DockerProvider executes commands in Docker containers.
type DockerProvider struct {
	sandboxes map[string]*dockerSandbox
	available bool
}

type dockerSandbox struct {
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

// NewDockerProvider creates a new Docker sandbox provider.
func NewDockerProvider() *DockerProvider {
	return &DockerProvider{
		sandboxes: make(map[string]*dockerSandbox),
		available: isDockerAvailable(),
	}
}

func (p *DockerProvider) Name() string                  { return "docker" }
func (p *DockerProvider) Type() providers.ProviderType { return providers.ProviderTypeSandbox }

func (p *DockerProvider) Init(_ context.Context, _ map[string]any) error {
	if !p.available {
		return fmt.Errorf("docker is not available")
	}
	return nil
}

func (p *DockerProvider) Close() error {
	// Stop and remove all containers
	for id := range p.sandboxes {
		_ = p.Delete(context.Background(), id, true)
	}
	return nil
}

// IsAvailable returns whether Docker is available on the system.
func (p *DockerProvider) IsAvailable() bool {
	return p.available
}

// Create creates a new Docker container sandbox.
func (p *DockerProvider) Create(ctx context.Context, opts providers.SandboxCreateOptions) (providers.Sandbox, error) {
	if !p.available {
		return providers.Sandbox{}, fmt.Errorf("docker is not available")
	}

	id := uuid.New().String()[:8]
	name := opts.Name
	if name == "" {
		name = "ayo-sandbox-" + id
	}

	image := opts.Image
	if image == "" {
		image = "busybox:latest"
	}

	debug.Log("creating docker sandbox", "id", id, "name", name, "image", image)

	// Build docker run command
	args := []string{"run", "-d", "--name", name}

	// Add mounts
	for _, m := range opts.Mounts {
		mountOpt := m.Source + ":" + m.Destination
		if m.ReadOnly {
			mountOpt += ":ro"
		}
		args = append(args, "-v", mountOpt)
	}

	// Network mode
	if opts.Network.Enabled {
		// Default bridge network
	} else {
		args = append(args, "--network", "none")
	}

	// Resource limits
	if opts.Resources.CPUs > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%.1f", float64(opts.Resources.CPUs)))
	}
	if opts.Resources.MemoryMB > 0 {
		args = append(args, "--memory", fmt.Sprintf("%dm", opts.Resources.MemoryMB))
	}

	// Add image and command (keep container running)
	args = append(args, image, "sh", "-c", "while true; do sleep 3600; done")

	// Run container
	cmd := exec.CommandContext(ctx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		debug.Log("docker run failed", "error", err, "stderr", stderr.String())
		return providers.Sandbox{}, fmt.Errorf("docker run failed: %w: %s", err, stderr.String())
	}

	containerID := strings.TrimSpace(stdout.String())
	debug.Log("docker container created", "containerID", containerID[:12])

	sb := &dockerSandbox{
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
func (p *DockerProvider) Get(ctx context.Context, id string) (providers.Sandbox, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return providers.Sandbox{}, fmt.Errorf("sandbox not found: %s", id)
	}
	return p.sandboxToProviders(sb), nil
}

// List returns all sandboxes.
func (p *DockerProvider) List(ctx context.Context) ([]providers.Sandbox, error) {
	result := make([]providers.Sandbox, 0, len(p.sandboxes))
	for _, sb := range p.sandboxes {
		result = append(result, p.sandboxToProviders(sb))
	}
	return result, nil
}

// Start starts a stopped container.
func (p *DockerProvider) Start(ctx context.Context, id string) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	cmd := exec.CommandContext(ctx, "docker", "start", sb.containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker start failed: %w", err)
	}

	sb.status = providers.SandboxStatusRunning
	return nil
}

// Stop stops a running container.
func (p *DockerProvider) Stop(ctx context.Context, id string, opts providers.SandboxStopOptions) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	args := []string{"stop"}
	if opts.Timeout > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", int(opts.Timeout.Seconds())))
	}
	args = append(args, sb.containerID)

	cmd := exec.CommandContext(ctx, "docker", args...)
	if err := cmd.Run(); err != nil {
		// Try force kill as fallback
		_ = exec.CommandContext(ctx, "docker", "kill", sb.containerID).Run()
	}

	sb.status = providers.SandboxStatusStopped
	return nil
}

// Delete removes a container.
func (p *DockerProvider) Delete(ctx context.Context, id string, force bool) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, sb.containerID)

	cmd := exec.CommandContext(ctx, "docker", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker rm failed: %w", err)
	}

	delete(p.sandboxes, id)
	return nil
}

// Exec executes a command in the container.
func (p *DockerProvider) Exec(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return providers.ExecResult{}, fmt.Errorf("sandbox not found: %s", id)
	}

	start := time.Now()

	// Build docker exec command
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

	cmd := exec.CommandContext(execCtx, "docker", args...)

	// Set stdin if provided
	if len(opts.Stdin) > 0 {
		cmd.Stdin = strings.NewReader(string(opts.Stdin))
		args = append([]string{"exec", "-i"}, args[1:]...)
		cmd = exec.CommandContext(execCtx, "docker", args...)
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
func (p *DockerProvider) Status(ctx context.Context, id string) (providers.SandboxStatus, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return "", fmt.Errorf("sandbox not found: %s", id)
	}

	// Check actual container status
	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Status}}", sb.containerID)
	output, err := cmd.Output()
	if err != nil {
		return sb.status, nil // Fall back to cached status
	}

	status := strings.TrimSpace(string(output))
	switch status {
	case "running":
		sb.status = providers.SandboxStatusRunning
	case "exited", "dead":
		sb.status = providers.SandboxStatusStopped
	case "created":
		sb.status = providers.SandboxStatusCreating
	default:
		sb.status = providers.SandboxStatusFailed
	}

	return sb.status, nil
}

func (p *DockerProvider) sandboxToProviders(sb *dockerSandbox) providers.Sandbox {
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
func (p *DockerProvider) AssignAgent(id, agentHandle string) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}
	sb.agents = append(sb.agents, agentHandle)
	return nil
}

// isDockerAvailable checks if Docker is available on the system.
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// Verify DockerProvider implements SandboxProvider.
var _ providers.SandboxProvider = (*DockerProvider)(nil)
