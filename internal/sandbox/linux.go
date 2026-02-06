// Package sandbox provides container-based agent execution environments.
package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/google/uuid"
)

// LinuxProvider executes commands using systemd-nspawn on Linux.
// systemd-nspawn provides native container support with:
// - Lightweight namespace isolation
// - Built into systemd (no extra packages on most distros)
// - Good security via cgroups and namespaces
// - Simple CLI interface
type LinuxProvider struct {
	sandboxes map[string]*linuxSandbox
	available bool
	rootfsDir string
}

type linuxSandbox struct {
	id          string
	name        string
	status      providers.SandboxStatus
	createdAt   time.Time
	pool        string
	agents      []string
	image       string
	mounts      []providers.Mount
	network     providers.NetworkConfig
	rootfsPath  string
	cmd         *exec.Cmd
	cancelFunc  context.CancelFunc
}

// NewLinuxProvider creates a new Linux container sandbox provider.
func NewLinuxProvider() *LinuxProvider {
	return &LinuxProvider{
		sandboxes: make(map[string]*linuxSandbox),
		available: isLinuxContainerAvailable(),
		rootfsDir: defaultRootfsDir(),
	}
}

func (p *LinuxProvider) Name() string                  { return "systemd-nspawn" }
func (p *LinuxProvider) Type() providers.ProviderType { return providers.ProviderTypeSandbox }

func (p *LinuxProvider) Init(_ context.Context, _ map[string]any) error {
	if !p.available {
		return fmt.Errorf("systemd-nspawn is not available (requires Linux with systemd)")
	}
	return nil
}

func (p *LinuxProvider) Close() error {
	// Stop and remove all containers
	for id := range p.sandboxes {
		_ = p.Delete(context.Background(), id, true)
	}
	return nil
}

// IsAvailable returns whether systemd-nspawn is available on the system.
func (p *LinuxProvider) IsAvailable() bool {
	return p.available
}

// Create creates a new systemd-nspawn sandbox.
func (p *LinuxProvider) Create(ctx context.Context, opts providers.SandboxCreateOptions) (providers.Sandbox, error) {
	if !p.available {
		return providers.Sandbox{}, fmt.Errorf("systemd-nspawn is not available")
	}

	id := uuid.New().String()[:8]
	name := opts.Name
	if name == "" {
		name = "ayo-sandbox-" + id
	}

	debug.Log("creating linux sandbox", "id", id, "name", name)

	// Use the prepared rootfs
	rootfsPath := p.rootfsDir
	if opts.Image != "" {
		// Could support custom rootfs paths in future
		rootfsPath = p.rootfsDir
	}

	// Verify rootfs exists
	if _, err := os.Stat(rootfsPath); err != nil {
		return providers.Sandbox{}, fmt.Errorf("rootfs not found at %s: %w", rootfsPath, err)
	}

	// Build systemd-nspawn command for background container
	args := []string{
		"--directory=" + rootfsPath,
		"--machine=" + name,
		"--quiet",
	}

	// Add mounts
	for _, m := range opts.Mounts {
		if m.ReadOnly {
			args = append(args, fmt.Sprintf("--bind-ro=%s:%s", m.Source, m.Destination))
		} else {
			args = append(args, fmt.Sprintf("--bind=%s:%s", m.Source, m.Destination))
		}
	}

	// Network mode
	if !opts.Network.Enabled {
		args = append(args, "--private-network")
	}

	// Resource limits (via systemd properties)
	if opts.Resources.CPUs > 0 {
		args = append(args, fmt.Sprintf("--property=CPUQuota=%d%%", opts.Resources.CPUs*100))
	}
	if opts.Resources.MemoryMB > 0 {
		args = append(args, fmt.Sprintf("--property=MemoryMax=%dM", opts.Resources.MemoryMB))
	}

	// Run with keepalive command
	args = append(args, "/bin/sh", "-c", "sleep infinity")

	// Create a cancellable context for the container
	containerCtx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(containerCtx, "systemd-nspawn", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		cancel()
		debug.Log("systemd-nspawn start failed", "error", err, "stderr", stderr.String())
		return providers.Sandbox{}, fmt.Errorf("systemd-nspawn start failed: %w: %s", err, stderr.String())
	}

	debug.Log("linux sandbox created", "name", name, "pid", cmd.Process.Pid)

	sb := &linuxSandbox{
		id:         id,
		name:       name,
		status:     providers.SandboxStatusRunning,
		createdAt:  time.Now(),
		pool:       opts.Pool,
		agents:     make([]string, 0),
		image:      opts.Image,
		mounts:     opts.Mounts,
		network:    opts.Network,
		rootfsPath: rootfsPath,
		cmd:        cmd,
		cancelFunc: cancel,
	}
	p.sandboxes[id] = sb

	// Wait for container to exit in background
	go func() {
		_ = cmd.Wait()
		sb.status = providers.SandboxStatusStopped
	}()

	return p.sandboxToProviders(sb), nil
}

// Get retrieves a sandbox by ID.
func (p *LinuxProvider) Get(ctx context.Context, id string) (providers.Sandbox, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return providers.Sandbox{}, fmt.Errorf("sandbox not found: %s", id)
	}
	return p.sandboxToProviders(sb), nil
}

// List returns all sandboxes.
func (p *LinuxProvider) List(ctx context.Context) ([]providers.Sandbox, error) {
	result := make([]providers.Sandbox, 0, len(p.sandboxes))
	for _, sb := range p.sandboxes {
		result = append(result, p.sandboxToProviders(sb))
	}
	return result, nil
}

// Start starts a stopped container.
func (p *LinuxProvider) Start(ctx context.Context, id string) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	if sb.status == providers.SandboxStatusRunning {
		return nil
	}

	// Recreate the container
	opts := providers.SandboxCreateOptions{
		Name:    sb.name,
		Pool:    sb.pool,
		Image:   sb.image,
		Mounts:  sb.mounts,
		Network: sb.network,
	}

	newSb, err := p.Create(ctx, opts)
	if err != nil {
		return err
	}

	// Update the sandbox reference
	if newSandbox, ok := p.sandboxes[newSb.ID]; ok {
		newSandbox.id = id
		delete(p.sandboxes, newSb.ID)
		p.sandboxes[id] = newSandbox
	}

	return nil
}

// Stop stops a running container.
func (p *LinuxProvider) Stop(ctx context.Context, id string, opts providers.SandboxStopOptions) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	if sb.cancelFunc != nil {
		sb.cancelFunc()
	}

	if sb.cmd != nil && sb.cmd.Process != nil {
		// Give it time to stop gracefully
		done := make(chan error, 1)
		go func() {
			done <- sb.cmd.Wait()
		}()

		timeout := opts.Timeout
		if timeout == 0 {
			timeout = 10 * time.Second
		}

		select {
		case <-done:
			// Stopped gracefully
		case <-time.After(timeout):
			// Force kill
			_ = sb.cmd.Process.Kill()
		}
	}

	sb.status = providers.SandboxStatusStopped
	return nil
}

// Delete removes a container.
func (p *LinuxProvider) Delete(ctx context.Context, id string, force bool) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	// Stop if running
	if sb.status == providers.SandboxStatusRunning {
		_ = p.Stop(ctx, id, providers.SandboxStopOptions{Timeout: 5 * time.Second})
	}

	delete(p.sandboxes, id)
	return nil
}

// Exec executes a command in the container.
func (p *LinuxProvider) Exec(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return providers.ExecResult{}, fmt.Errorf("sandbox not found: %s", id)
	}

	start := time.Now()

	// Build systemd-nspawn exec command
	// We run a new nspawn process that attaches to the same rootfs
	args := []string{
		"--directory=" + sb.rootfsPath,
		"--quiet",
	}

	// Working directory
	if opts.WorkingDir != "" {
		args = append(args, "--chdir="+opts.WorkingDir)
	}

	// Environment variables
	for k, v := range opts.Env {
		args = append(args, fmt.Sprintf("--setenv=%s=%s", k, v))
	}

	// Add the same mounts as the original container
	for _, m := range sb.mounts {
		if m.ReadOnly {
			args = append(args, fmt.Sprintf("--bind-ro=%s:%s", m.Source, m.Destination))
		} else {
			args = append(args, fmt.Sprintf("--bind=%s:%s", m.Source, m.Destination))
		}
	}

	// Network
	if !sb.network.Enabled {
		args = append(args, "--private-network")
	}

	// Add command
	if len(opts.Args) > 0 {
		args = append(args, opts.Command)
		args = append(args, opts.Args...)
	} else {
		args = append(args, "/bin/sh", "-c", opts.Command)
	}

	// Create command with timeout
	execCtx := ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(execCtx, "systemd-nspawn", args...)

	// Set stdin if provided
	if len(opts.Stdin) > 0 {
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
func (p *LinuxProvider) Status(ctx context.Context, id string) (providers.SandboxStatus, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return "", fmt.Errorf("sandbox not found: %s", id)
	}

	// Check if process is still running
	if sb.cmd != nil && sb.cmd.Process != nil {
		// Try to send signal 0 to check if process exists
		if err := sb.cmd.Process.Signal(os.Signal(nil)); err != nil {
			sb.status = providers.SandboxStatusStopped
		}
	}

	return sb.status, nil
}

func (p *LinuxProvider) sandboxToProviders(sb *linuxSandbox) providers.Sandbox {
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
func (p *LinuxProvider) AssignAgent(id, agentHandle string) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}
	sb.agents = append(sb.agents, agentHandle)
	return nil
}

// Stats returns resource usage statistics for a sandbox.
func (p *LinuxProvider) Stats(ctx context.Context, id string) (providers.SandboxStats, error) {
	sb, ok := p.sandboxes[id]
	if !ok {
		return providers.SandboxStats{}, fmt.Errorf("sandbox not found: %s", id)
	}

	uptime := time.Since(sb.createdAt)

	// Try to read cgroup stats if the machine is running
	// systemd-nspawn creates a machine scope under /sys/fs/cgroup/machine.slice/
	cgroupPath := fmt.Sprintf("/sys/fs/cgroup/machine.slice/machine-%s.scope", sb.name)

	var stats providers.SandboxStats
	stats.Uptime = uptime

	// Read memory stats
	memCurrent, err := os.ReadFile(filepath.Join(cgroupPath, "memory.current"))
	if err == nil {
		var memBytes int64
		fmt.Sscanf(string(memCurrent), "%d", &memBytes)
		stats.MemoryUsageBytes = memBytes
	}

	memMax, err := os.ReadFile(filepath.Join(cgroupPath, "memory.max"))
	if err == nil && string(memMax) != "max\n" {
		var memLimit int64
		fmt.Sscanf(string(memMax), "%d", &memLimit)
		stats.MemoryLimitBytes = memLimit
	}

	// Read CPU stats (simplified - would need more parsing for accurate percentage)
	cpuStat, err := os.ReadFile(filepath.Join(cgroupPath, "cpu.stat"))
	if err == nil {
		debug.Log("cpu stats", "raw", string(cpuStat))
		// CPU percentage would require sampling over time, so we leave it at 0
	}

	// Count processes
	cgroupProcs, err := os.ReadFile(filepath.Join(cgroupPath, "cgroup.procs"))
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(cgroupProcs)), "\n")
		if lines[0] != "" {
			stats.ProcessCount = len(lines)
		}
	}

	return stats, nil
}

// isLinuxContainerAvailable checks if systemd-nspawn is available on the system.
func isLinuxContainerAvailable() bool {
	// Check if we're on Linux
	if runtime.GOOS != "linux" {
		return false
	}

	// Check if systemd-nspawn exists
	cmd := exec.Command("systemd-nspawn", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}

	// Check if we have a rootfs available
	rootfs := defaultRootfsDir()
	if _, err := os.Stat(rootfs); err != nil {
		return false
	}

	return true
}

// defaultRootfsDir returns the default rootfs directory for ayo sandboxes.
func defaultRootfsDir() string {
	// Check for ayo-specific rootfs
	ayoRootfs := filepath.Join(os.Getenv("HOME"), ".local", "share", "ayo", "rootfs")
	if _, err := os.Stat(ayoRootfs); err == nil {
		return ayoRootfs
	}

	// Fallback to system machines directory
	return "/var/lib/machines/ayo-sandbox"
}

// EnsureAgentUser ensures a Unix user exists for the agent in the sandbox.
// If dotfilesPath is non-empty, copies dotfiles from that host directory to the user's home.
func (p *LinuxProvider) EnsureAgentUser(ctx context.Context, id string, agentHandle string, dotfilesPath string) error {
	sb, ok := p.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	// Check if user already exists using machinectl shell
	checkCmd := exec.CommandContext(ctx, "machinectl", "shell", sb.name, "/usr/bin/id", agentHandle)
	if err := checkCmd.Run(); err == nil {
		return nil // User exists
	}

	// Create user using machinectl shell
	createCmd := exec.CommandContext(ctx, "machinectl", "shell", sb.name,
		"/usr/sbin/adduser", "-D", "-s", "/bin/sh", agentHandle)
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create user %s: %w", agentHandle, err)
	}

	// Copy dotfiles if provided
	if dotfilesPath != "" {
		if err := p.copyDotfiles(ctx, sb.name, agentHandle, dotfilesPath); err != nil {
			debug.Log("failed to copy dotfiles", "agent", agentHandle, "error", err)
			// Not fatal - user is still created
		}
	}

	return nil
}

// copyDotfiles copies dotfiles from a host directory to the agent's home directory.
func (p *LinuxProvider) copyDotfiles(ctx context.Context, machineName, agentHandle, dotfilesPath string) error {
	// Check if dotfiles directory exists on host
	info, err := os.Stat(dotfilesPath)
	if err != nil {
		if os.IsNotExist(err) {
			debug.Log("no dotfiles directory", "path", dotfilesPath)
			return nil // No dotfiles to copy
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("dotfiles path is not a directory: %s", dotfilesPath)
	}

	// Read dotfiles from host directory
	entries, err := os.ReadDir(dotfilesPath)
	if err != nil {
		return fmt.Errorf("read dotfiles directory: %w", err)
	}

	homeDir := "/home/" + agentHandle

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories for now
		}

		srcPath := filepath.Join(dotfilesPath, entry.Name())
		dstPath := homeDir + "/" + entry.Name()

		// Read file content from host
		content, err := os.ReadFile(srcPath)
		if err != nil {
			debug.Log("failed to read dotfile", "file", srcPath, "error", err)
			continue
		}

		// Write file content to container using machinectl shell with heredoc
		writeCmd := fmt.Sprintf("cat > %s << 'DOTFILE_EOF'\n%sDOTFILE_EOF", dstPath, string(content))
		cmd := exec.CommandContext(ctx, "machinectl", "shell", machineName, "/bin/sh", "-c", writeCmd)
		if err := cmd.Run(); err != nil {
			debug.Log("failed to write dotfile", "file", dstPath, "error", err)
			continue
		}

		// Set ownership to the agent user
		chownCmd := exec.CommandContext(ctx, "machinectl", "shell", machineName,
			"/bin/chown", agentHandle+":"+agentHandle, dstPath)
		if err := chownCmd.Run(); err != nil {
			debug.Log("failed to chown dotfile", "file", dstPath, "error", err)
		}

		debug.Log("copied dotfile", "src", srcPath, "dst", dstPath)
	}

	return nil
}

// Verify LinuxProvider implements SandboxProvider.
var _ providers.SandboxProvider = (*LinuxProvider)(nil)
