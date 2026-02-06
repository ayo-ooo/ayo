// Package sandbox provides container-based agent execution environments.
package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
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
	id           string
	name         string
	containerID  string
	status       providers.SandboxStatus
	createdAt    time.Time
	pool         string
	agents       []string
	user         string // User account created in the sandbox
	image        string
	mounts       []providers.Mount
	network      providers.NetworkConfig
	createdUsers map[string]bool // Tracks which agent users have been created
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
		image = "docker.io/library/alpine:3.21"
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
	} else {
		// Use Cloudflare DNS for reliable resolution
		// The default gateway DNS (192.168.64.1) often doesn't work
		args = append(args, "--dns", "1.1.1.1")
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

	// Create standard sandbox directory structure
	if err := p.setupDirectories(ctx, containerID); err != nil {
		debug.Log("directory setup failed", "error", err)
		// Not fatal - directories can be created later
	}

	// Install and configure ngircd for inter-agent communication
	if err := p.setupNgircd(ctx, containerID); err != nil {
		debug.Log("ngircd setup failed", "error", err)
		// Not fatal - sandbox can still function without IRC
	}

	// Install IRC helper scripts (msg, irc-log, irc-join, irc-nick)
	if err := p.setupIRCHelpers(ctx, containerID); err != nil {
		debug.Log("IRC helpers setup failed", "error", err)
		// Not fatal - scripts are convenience utilities
	}

	sb := &appleSandbox{
		id:           name, // Use name as ID for consistency with List()
		name:         name,
		containerID:  containerID,
		status:       providers.SandboxStatusRunning,
		createdAt:    time.Now(),
		pool:         opts.Pool,
		agents:       make([]string, 0),
		user:         opts.User,
		image:        image,
		mounts:       opts.Mounts,
		network:      opts.Network,
		createdUsers: make(map[string]bool),
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

// Stats returns resource usage statistics for a sandbox.
func (p *AppleProvider) Stats(ctx context.Context, id string) (providers.SandboxStats, error) {
	// Resolve container ID
	containerID, err := p.resolveContainerID(ctx, id)
	if err != nil {
		return providers.SandboxStats{}, err
	}

	// Get uptime from in-memory cache if available
	var uptime time.Duration
	if sb, ok := p.sandboxes[id]; ok {
		uptime = time.Since(sb.createdAt)
	}

	// Try to get stats from container runtime
	// Apple Container supports `container stats <name>` for resource info
	cmd := exec.CommandContext(ctx, "container", "stats", containerID, "--format", "json")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		debug.Log("container stats failed, returning basic stats", "id", id, "error", err)
		// Return basic stats on error
		return providers.SandboxStats{
			Uptime: uptime,
		}, nil
	}

	// Parse stats output (Apple Container JSON format)
	var statsOutput struct {
		CPUPercent  float64 `json:"cpuPercent"`
		MemoryBytes int64   `json:"memoryBytes"`
		MemoryLimit int64   `json:"memoryLimit"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &statsOutput); err != nil {
		debug.Log("failed to parse stats JSON", "error", err)
		return providers.SandboxStats{Uptime: uptime}, nil
	}

	return providers.SandboxStats{
		CPUPercent:       statsOutput.CPUPercent,
		MemoryUsageBytes: statsOutput.MemoryBytes,
		MemoryLimitBytes: statsOutput.MemoryLimit,
		Uptime:           uptime,
	}, nil
}

// ngircdConfig is a minimal ngircd configuration for inter-agent communication.
const ngircdConfig = `[Global]
Name = ayo.sandbox
Info = Ayo Sandbox IRC Server
Ports = 6667
Listen = 127.0.0.1

[Limits]
MaxConnections = 100
MaxChannels = 50
MaxNickLength = 32

[Options]
AllowRemoteOper = no
ChrootDir =
OperCanUseMode = yes
OperChanPAutoOp = yes
PredefChannelsOnly = no
PAM = no

[Channel]
Name = #general
Topic = General agent communication channel
Modes = tn
`

// ircHelperScripts contains shell scripts for IRC communication.
// These are installed in /usr/local/bin/ for easy agent access.
var ircHelperScripts = map[string]string{
	// msg - Send an IRC message to a channel or user
	"msg": `#!/bin/sh
# Usage: msg <target> <message>
# target: #channel or @agent
if [ $# -lt 2 ]; then
  echo "Usage: msg <target> <message>" >&2
  exit 1
fi
target="$1"; shift
# Convert @user to plain user for PRIVMSG
case "$target" in
  @*) target="${target#@}" ;;
esac
printf 'PRIVMSG %s :%s\r\n' "$target" "$*" | nc -q0 localhost 6667
`,

	// irc-log - Read IRC logs for a channel
	"irc-log": `#!/bin/sh
# Usage: irc-log [channel] [lines]
channel="${1:-general}"
lines="${2:-20}"
logfile="/var/log/irc/${channel}.log"
if [ -f "$logfile" ]; then
  tail -n "$lines" "$logfile"
else
  echo "No logs for channel: $channel" >&2
  exit 1
fi
`,

	// irc-join - Join an IRC channel
	"irc-join": `#!/bin/sh
# Usage: irc-join <channel>
if [ $# -lt 1 ]; then
  echo "Usage: irc-join <channel>" >&2
  exit 1
fi
channel="$1"
# Ensure channel starts with #
case "$channel" in
  \#*) ;;
  *) channel="#$channel" ;;
esac
printf 'JOIN %s\r\n' "$channel" | nc -q0 localhost 6667
`,

	// irc-nick - Set IRC nickname
	"irc-nick": `#!/bin/sh
# Usage: irc-nick <nickname>
if [ $# -lt 1 ]; then
  echo "Usage: irc-nick <nickname>" >&2
  exit 1
fi
nick="$1"
printf 'NICK %s\r\nUSER %s 0 * :%s\r\n' "$nick" "$nick" "$nick" | nc -q0 localhost 6667
`,
}

// setupDirectories creates the standard sandbox directory structure.
func (p *AppleProvider) setupDirectories(ctx context.Context, containerID string) error {
	debug.Log("setting up sandbox directories", "container", containerID)

	// Create standard directories with appropriate permissions
	dirs := []struct {
		path string
		mode string
	}{
		{"/shared", "1777"},           // Sticky bit, world-writable for cross-agent sharing
		{"/workspaces", "755"},        // Session workspaces, root-owned
		{"/var/log/irc", "755"},       // IRC logs
		{"/mnt/host", "755"},          // Host filesystem mount point
	}

	for _, dir := range dirs {
		if err := p.execSimple(ctx, containerID, "mkdir", "-p", dir.path); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir.path, err)
		}
		if err := p.execSimple(ctx, containerID, "chmod", dir.mode, dir.path); err != nil {
			debug.Log("failed to chmod directory", "path", dir.path, "error", err)
		}
	}

	debug.Log("sandbox directories setup complete", "container", containerID)
	return nil
}

// setupNgircd installs and configures ngircd in the sandbox container.
func (p *AppleProvider) setupNgircd(ctx context.Context, containerID string) error {
	debug.Log("setting up ngircd", "container", containerID)

	// Create IRC log directory
	if err := p.execSimple(ctx, containerID, "mkdir", "-p", "/var/log/irc"); err != nil {
		debug.Log("failed to create IRC log directory", "error", err)
	}

	// Update apk index and install ngircd and netcat (for IRC connectivity)
	if err := p.execSimple(ctx, containerID, "apk", "update"); err != nil {
		return fmt.Errorf("failed to update apk index: %w", err)
	}
	if err := p.execSimple(ctx, containerID, "apk", "add", "--no-cache", "ngircd", "netcat-openbsd"); err != nil {
		return fmt.Errorf("failed to install ngircd: %w", err)
	}

	// Write ngircd config
	// Using sh -c with heredoc to write the config file
	configCmd := fmt.Sprintf("cat > /etc/ngircd/ngircd.conf << 'EOF'\n%sEOF", ngircdConfig)
	if err := p.execShell(ctx, containerID, configCmd); err != nil {
		return fmt.Errorf("failed to write ngircd config: %w", err)
	}

	// Start ngircd in the background
	// ngircd -n runs in foreground, so we use nohup + & to background it
	if err := p.execShell(ctx, containerID, "nohup ngircd -n > /var/log/irc/ngircd.log 2>&1 &"); err != nil {
		return fmt.Errorf("failed to start ngircd: %w", err)
	}

	// Verify ngircd is running
	if err := p.execSimple(ctx, containerID, "pgrep", "ngircd"); err != nil {
		return fmt.Errorf("ngircd failed to start: %w", err)
	}

	debug.Log("ngircd setup complete", "container", containerID)
	return nil
}

// setupIRCHelpers installs IRC helper scripts in /usr/local/bin/.
func (p *AppleProvider) setupIRCHelpers(ctx context.Context, containerID string) error {
	debug.Log("setting up IRC helper scripts", "container", containerID)

	// Create /usr/local/bin if it doesn't exist
	if err := p.execSimple(ctx, containerID, "mkdir", "-p", "/usr/local/bin"); err != nil {
		return fmt.Errorf("failed to create /usr/local/bin: %w", err)
	}

	// Install each helper script
	for name, content := range ircHelperScripts {
		scriptPath := fmt.Sprintf("/usr/local/bin/%s", name)
		// Write script content using heredoc
		writeCmd := fmt.Sprintf("cat > %s << 'SCRIPT_EOF'\n%sSCRIPT_EOF", scriptPath, content)
		if err := p.execShell(ctx, containerID, writeCmd); err != nil {
			return fmt.Errorf("failed to write %s script: %w", name, err)
		}
		// Make executable
		if err := p.execSimple(ctx, containerID, "chmod", "+x", scriptPath); err != nil {
			return fmt.Errorf("failed to chmod %s: %w", name, err)
		}
		debug.Log("installed IRC helper script", "name", name, "path", scriptPath)
	}

	debug.Log("IRC helper scripts setup complete", "container", containerID)
	return nil
}

// execSimple executes a simple command in the container.
func (p *AppleProvider) execSimple(ctx context.Context, containerID string, cmdArgs ...string) error {
	args := []string{"exec", containerID}
	args = append(args, cmdArgs...)
	cmd := exec.CommandContext(ctx, "container", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err, stderr.String())
	}
	return nil
}

// execShell executes a shell command string in the container.
func (p *AppleProvider) execShell(ctx context.Context, containerID string, command string) error {
	args := []string{"exec", containerID, "sh", "-c", command}
	cmd := exec.CommandContext(ctx, "container", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err, stderr.String())
	}
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

// EnsureAgentUser ensures a Unix user exists for the agent in the sandbox.
// If dotfilesPath is non-empty, copies dotfiles from that host directory to the user's home.
func (p *AppleProvider) EnsureAgentUser(ctx context.Context, id string, agentHandle string, dotfilesPath string) error {
	// Resolve container ID
	containerID, err := p.resolveContainerID(ctx, id)
	if err != nil {
		return err
	}

	// Check if we've already created this user in this sandbox instance
	sb, ok := p.sandboxes[id]
	if ok && sb.createdUsers != nil && sb.createdUsers[agentHandle] {
		debug.Log("agent user already tracked as created", "agent", agentHandle)
		return nil
	}

	// Check if user already exists in container
	if err := p.execSimple(ctx, containerID, "id", agentHandle); err == nil {
		debug.Log("agent user already exists", "agent", agentHandle)
		// Track that user exists
		if ok && sb.createdUsers != nil {
			sb.createdUsers[agentHandle] = true
		}
		return nil // User exists
	}

	debug.Log("creating agent user", "agent", agentHandle, "container", containerID)

	// Create user with adduser (Alpine/BusyBox-compatible)
	// -D: don't assign a password
	// -s /bin/sh: set shell
	// -h /home/{agent}: set home directory
	if err := p.execSimple(ctx, containerID, "adduser", "-D", "-s", "/bin/sh", agentHandle); err != nil {
		return fmt.Errorf("failed to create user %s: %w", agentHandle, err)
	}

	// Track that we created this user
	if ok && sb.createdUsers != nil {
		sb.createdUsers[agentHandle] = true
	}

	debug.Log("agent user created", "agent", agentHandle)

	// Copy agent dotfiles if provided
	if dotfilesPath != "" {
		if err := p.copyDotfiles(ctx, containerID, agentHandle, dotfilesPath); err != nil {
			debug.Log("failed to copy dotfiles", "agent", agentHandle, "error", err)
			// Not fatal - user is still created
		}
	}

	return nil
}

// copyDotfiles copies dotfiles from a host directory to the agent's home directory.
func (p *AppleProvider) copyDotfiles(ctx context.Context, containerID, agentHandle, dotfilesPath string) error {
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

		// Write file content to container using heredoc
		writeCmd := fmt.Sprintf("cat > %s << 'DOTFILE_EOF'\n%sDOTFILE_EOF", dstPath, string(content))
		if err := p.execShell(ctx, containerID, writeCmd); err != nil {
			debug.Log("failed to write dotfile", "file", dstPath, "error", err)
			continue
		}

		// Set ownership to the agent user
		if err := p.execSimple(ctx, containerID, "chown", agentHandle+":"+agentHandle, dstPath); err != nil {
			debug.Log("failed to chown dotfile", "file", dstPath, "error", err)
		}

		debug.Log("copied dotfile", "src", srcPath, "dst", dstPath)
	}

	return nil
}

// Verify AppleProvider implements SandboxProvider.
var _ providers.SandboxProvider = (*AppleProvider)(nil)
