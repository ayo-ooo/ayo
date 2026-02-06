// Package sandbox provides container-based agent execution environments.
package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/google/uuid"
)

// AppleProvider executes commands using Apple Container on macOS 26+.
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
	user        string // User account created in the sandbox
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

func (p *AppleProvider) Name() string                  { return "apple-container" }
func (p *AppleProvider) Type() providers.ProviderType { return providers.ProviderTypeSandbox }

func (p *AppleProvider) Init(_ context.Context, _ map[string]any) error {
	if !p.available {
		return fmt.Errorf("Apple Container is not available (requires macOS 26+ on Apple Silicon)")
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
		image = "docker.io/library/busybox:stable"
	}

	debug.Log("creating apple container sandbox", "id", id, "name", name, "image", image)

	// Build container run command (detached mode with keepalive)
	// Using `container run -d` creates and starts in one command
	args := []string{"run", "-d", "--name", name}

	// Add mounts using -v (virtiofs under the hood)
	for _, m := range opts.Mounts {
		mountOpt := fmt.Sprintf("%s:%s", m.Source, m.Destination)
		if m.ReadOnly {
			mountOpt += ":ro"
		}
		args = append(args, "-v", mountOpt)
	}

	// Network mode - Apple Container uses --no-dns to disable network
	if !opts.Network.Enabled {
		args = append(args, "--no-dns")
	}

	// Resource limits
	if opts.Resources.CPUs > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%d", opts.Resources.CPUs))
	}
	if opts.Resources.MemoryMB > 0 {
		args = append(args, "--memory", fmt.Sprintf("%dM", opts.Resources.MemoryMB))
	}

	// Add image and keepalive command
	args = append(args, image, "sh", "-c", "sleep infinity")

	// Run container
	cmd := exec.CommandContext(ctx, "container", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		debug.Log("container run failed", "error", err, "stderr", stderr.String())
		return providers.Sandbox{}, fmt.Errorf("container run failed: %w: %s", err, stderr.String())
	}

	containerID := name // Apple Container uses names as IDs
	debug.Log("apple container created", "containerID", containerID)

	// Run setup commands if provided (e.g., user creation)
	for _, setupCmd := range opts.SetupCommands {
		if len(setupCmd) == 0 {
			continue
		}
		setupArgs := []string{"exec", containerID}
		setupArgs = append(setupArgs, setupCmd...)
		cmd = exec.CommandContext(ctx, "container", setupArgs...)
		stdout.Reset()
		stderr.Reset()
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			debug.Log("setup command failed", "command", setupCmd, "error", err, "stderr", stderr.String())
			// Continue with other setup commands, but log the error
		} else {
			debug.Log("setup command completed", "command", setupCmd)
		}
	}

	// If a user was specified, create the user account
	if opts.User != "" {
		debug.Log("creating user in sandbox", "user", opts.User)
		// Create user with adduser (busybox-compatible)
		// adduser -D creates without password prompt
		userCmd := []string{"exec", containerID, "adduser", "-D", "-s", "/bin/sh", opts.User}
		cmd = exec.CommandContext(ctx, "container", userCmd...)
		stdout.Reset()
		stderr.Reset()
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			debug.Log("user creation failed", "user", opts.User, "error", err, "stderr", stderr.String())
			// User might already exist, not fatal
		} else {
			debug.Log("user created", "user", opts.User)
		}
	}

	sb := &appleSandbox{
		id:          name, // Use name as ID for consistency with List()
		name:        name,
		containerID: containerID,
		status:      providers.SandboxStatusRunning,
		createdAt:   time.Now(),
		pool:        opts.Pool,
		agents:      make([]string, 0),
		user:        opts.User,
		image:       image,
		mounts:      opts.Mounts,
		network:     opts.Network,
	}
	p.sandboxes[name] = sb // Key by name, not internal ID

	return p.sandboxToProviders(sb), nil
}

// Get retrieves a sandbox by ID.
func (p *AppleProvider) Get(ctx context.Context, id string) (providers.Sandbox, error) {
	// First check in-memory cache
	if sb, ok := p.sandboxes[id]; ok {
		return p.sandboxToProviders(sb), nil
	}

	// Query the runtime for all containers
	containers, err := p.List(ctx)
	if err != nil {
		return providers.Sandbox{}, err
	}

	// Find matching container (by exact ID or prefix match)
	for _, c := range containers {
		if c.ID == id || strings.HasPrefix(c.ID, id) || strings.HasPrefix(c.Name, id) {
			return c, nil
		}
	}

	return providers.Sandbox{}, fmt.Errorf("sandbox not found: %s", id)
}

// containerListEntry represents a container from `container ls --format json`
type containerListEntry struct {
	Status        string `json:"status"`
	Configuration struct {
		ID     string `json:"id"`
		Image  struct {
			Reference string `json:"reference"`
		} `json:"image"`
		Mounts []struct {
			Source      string `json:"source"`
			Destination string `json:"destination"`
			Type        struct {
				VirtioFS *struct{} `json:"virtiofs,omitempty"`
			} `json:"type"`
		} `json:"mounts"`
	} `json:"configuration"`
	StartedDate float64 `json:"startedDate"` // macOS absolute time (seconds since 2001-01-01)
}

// List returns all sandboxes by querying the container runtime.
func (p *AppleProvider) List(ctx context.Context) ([]providers.Sandbox, error) {
	cmd := exec.CommandContext(ctx, "container", "ls", "--all", "--format", "json")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("container ls failed: %w", err)
	}

	var entries []containerListEntry
	if err := json.Unmarshal(out.Bytes(), &entries); err != nil {
		return nil, fmt.Errorf("parse container ls output: %w", err)
	}

	// Filter to ayo containers (those with "ayo-" prefix)
	var result []providers.Sandbox
	for _, e := range entries {
		name := e.Configuration.ID
		if !strings.HasPrefix(name, "ayo-") {
			continue
		}

		// Convert macOS absolute time to Go time
		// macOS absolute time is seconds since 2001-01-01
		macEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		createdAt := macEpoch.Add(time.Duration(e.StartedDate * float64(time.Second)))

		status := providers.SandboxStatusStopped
		if e.Status == "running" {
			status = providers.SandboxStatusRunning
		}

		var mounts []providers.Mount
		for _, m := range e.Configuration.Mounts {
			mounts = append(mounts, providers.Mount{
				Source:      m.Source,
				Destination: m.Destination,
				Mode:        providers.MountModeVirtioFS,
			})
		}

		result = append(result, providers.Sandbox{
			ID:        name, // Use container name as ID for CLI operations
			Name:      name,
			Image:     e.Configuration.Image.Reference,
			Status:    status,
			CreatedAt: createdAt,
			Mounts:    mounts,
		})
	}

	return result, nil
}

// Start starts a stopped container.
func (p *AppleProvider) Start(ctx context.Context, id string) error {
	// Resolve container ID
	containerID, err := p.resolveContainerID(ctx, id)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "container", "start", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("container start failed: %w", err)
	}

	if sb, ok := p.sandboxes[id]; ok {
		sb.status = providers.SandboxStatusRunning
	}
	return nil
}

// Stop stops a running container.
func (p *AppleProvider) Stop(ctx context.Context, id string, opts providers.SandboxStopOptions) error {
	// Resolve container ID
	containerID, err := p.resolveContainerID(ctx, id)
	if err != nil {
		return err
	}

	args := []string{"stop"}
	if opts.Timeout > 0 {
		args = append(args, "--time", fmt.Sprintf("%d", int(opts.Timeout.Seconds())))
	}
	args = append(args, containerID)

	cmd := exec.CommandContext(ctx, "container", args...)
	if err := cmd.Run(); err != nil {
		// Try force kill as fallback
		_ = exec.CommandContext(ctx, "container", "kill", containerID).Run()
	}

	if sb, ok := p.sandboxes[id]; ok {
		sb.status = providers.SandboxStatusStopped
	}
	return nil
}

// Delete removes a container.
func (p *AppleProvider) Delete(ctx context.Context, id string, force bool) error {
	// Resolve container ID
	containerID, err := p.resolveContainerID(ctx, id)
	if err != nil {
		return err
	}

	args := []string{"delete"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, containerID)

	cmd := exec.CommandContext(ctx, "container", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("container delete failed: %w", err)
	}

	delete(p.sandboxes, id)
	return nil
}

// Exec executes a command in the container.
func (p *AppleProvider) Exec(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
	// Resolve container ID
	containerID, err := p.resolveContainerID(ctx, id)
	if err != nil {
		return providers.ExecResult{}, err
	}

	start := time.Now()

	// Build container exec command
	args := []string{"exec"}

	// Working directory (Apple Container uses --workdir)
	if opts.WorkingDir != "" {
		args = append(args, "--workdir", opts.WorkingDir)
	}

	// User (Apple Container uses --user)
	if opts.User != "" {
		args = append(args, "--user", opts.User)
	}

	// Environment variables (Apple Container uses --env)
	for k, v := range opts.Env {
		args = append(args, "--env", k+"="+v)
	}

	args = append(args, containerID)

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
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
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
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
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
	// First try in-memory cache
	if sb, ok := p.sandboxes[id]; ok {
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
		}
		return sb.status, nil
	}

	// Query the runtime
	sb, err := p.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return sb.Status, nil
}

func (p *AppleProvider) sandboxToProviders(sb *appleSandbox) providers.Sandbox {
	return providers.Sandbox{
		ID:        sb.id,
		Name:      sb.name,
		Image:     sb.image,
		Status:    sb.status,
		Pool:      sb.pool,
		Agents:    sb.agents,
		User:      sb.user,
		Mounts:    sb.mounts,
		CreatedAt: sb.createdAt,
	}
}

// resolveContainerID resolves an ID or prefix to a full container name.
// It first checks the in-memory cache, then queries the runtime.
func (p *AppleProvider) resolveContainerID(ctx context.Context, id string) (string, error) {
	// First check in-memory cache
	if sb, ok := p.sandboxes[id]; ok {
		return sb.containerID, nil
	}

	// Query the runtime
	sb, err := p.Get(ctx, id)
	if err != nil {
		return "", err
	}

	return sb.ID, nil // Container name is used as both ID and container name
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
// Requires macOS 26+ on Apple Silicon.
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
