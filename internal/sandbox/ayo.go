package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/user"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/planners"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/share"
	ayosync "github.com/alexcabrera/ayo/internal/sync"
)

// AyoSandboxName is the fixed name for @ayo's dedicated sandbox.
const AyoSandboxName = "ayo-orchestrator"

// EnsureAyoSandbox ensures the dedicated @ayo orchestrator sandbox exists and is running.
// It creates the sandbox if it doesn't exist, or starts it if stopped.
// Returns the sandbox info.
func EnsureAyoSandbox(ctx context.Context, provider *AppleProvider) (providers.Sandbox, error) {
	// Check if sandbox already exists
	sb, err := provider.Get(ctx, AyoSandboxName)
	if err == nil {
		// Sandbox exists
		if sb.Status == providers.SandboxStatusRunning {
			debug.Log("ayo sandbox already running", "id", sb.ID)
			return sb, nil
		}
		// Start if stopped
		debug.Log("starting existing ayo sandbox", "id", sb.ID)
		if err := provider.Start(ctx, sb.ID); err != nil {
			return providers.Sandbox{}, fmt.Errorf("start ayo sandbox: %w", err)
		}
		sb.Status = providers.SandboxStatusRunning
		return sb, nil
	}

	// Sandbox doesn't exist, create it
	debug.Log("creating ayo sandbox")

	// Ensure host directories exist
	if err := paths.EnsureAyoSandboxDirs(); err != nil {
		return providers.Sandbox{}, fmt.Errorf("ensure ayo sandbox dirs: %w", err)
	}

	// Load config
	cfg, err := config.LoadAyoSandboxConfig()
	if err != nil {
		return providers.Sandbox{}, fmt.Errorf("load ayo sandbox config: %w", err)
	}

	// Build mounts
	mounts := []providers.Mount{
		// @ayo's persistent home directory
		{
			Source:      paths.AyoSandboxHomeDir(),
			Destination: "/home/ayo",
			Mode:        providers.MountModeVirtioFS,
		},
		// Output staging directory
		{
			Source:      paths.AyoSandboxOutputDir(),
			Destination: "/output",
			Mode:        providers.MountModeVirtioFS,
		},
		// Squads directory (will be populated with squad mounts)
		{
			Source:      paths.SquadsDir(),
			Destination: "/squads",
			Mode:        providers.MountModeVirtioFS,
		},
		// Workspace directory for user-shared files
		{
			Source:      ayosync.WorkspaceDir(),
			Destination: "/workspace",
			Mode:        providers.MountModeVirtioFS,
		},
		// Near-term planner state directory
		{
			Source:      paths.AyoSandboxPlannerNearDir(),
			Destination: "/.planner.near",
			Mode:        providers.MountModeVirtioFS,
		},
		// Long-term planner state directory
		{
			Source:      paths.AyoSandboxPlannerLongDir(),
			Destination: "/.planner.long",
			Mode:        providers.MountModeVirtioFS,
		},
	}

	// Add custom mounts from config
	for _, m := range cfg.Mounts {
		mount, err := parseMountSpec(m)
		if err != nil {
			debug.Log("invalid mount spec", "spec", m, "error", err)
			continue
		}
		mounts = append(mounts, mount)
	}

	// Add shares as direct VirtioFS mounts (instead of relying on symlinks in workspace)
	shareService := share.NewService()
	shares := shareService.List()
	for _, s := range shares {
		mounts = append(mounts, providers.Mount{
			Source:      s.Path,
			Destination: "/workspace/" + s.Name,
			Mode:        providers.MountModeVirtioFS,
		})
		debug.Log("adding share mount", "name", s.Name, "path", s.Path)
	}

	// Add host home directory as read-only mount at /mnt/{username}
	if homeDir, err := os.UserHomeDir(); err == nil {
		if u, err := user.Current(); err == nil {
			mounts = append(mounts, providers.Mount{
				Source:      homeDir,
				Destination: "/mnt/" + u.Username,
				Mode:        providers.MountModeVirtioFS,
				ReadOnly:    true,
			})
			debug.Log("adding host home mount", "user", u.Username, "path", homeDir)
		}
	}

	// Network config
	networkEnabled := true
	if cfg.Network != nil {
		networkEnabled = *cfg.Network
	}

	// Resource limits
	resources := providers.Resources{
		CPUs:     cfg.Resources.CPUs,
		MemoryMB: cfg.Resources.MemoryMB,
		DiskMB:   cfg.Resources.DiskMB,
	}
	if resources.CPUs == 0 {
		resources.CPUs = 2
	}
	if resources.MemoryMB == 0 {
		resources.MemoryMB = 2048
	}
	if resources.DiskMB == 0 {
		resources.DiskMB = 10240
	}

	// Build setup commands for package installation
	var setupCommands [][]string
	if len(cfg.Packages) > 0 {
		// Install packages using apk (Alpine)
		installCmd := append([]string{"apk", "add", "--no-cache"}, cfg.Packages...)
		setupCommands = append(setupCommands, installCmd)
	}

	// Create the sandbox
	opts := providers.SandboxCreateOptions{
		Name:          AyoSandboxName,
		Image:         cfg.Image,
		Mounts:        mounts,
		Resources:     resources,
		Network:       providers.NetworkConfig{Enabled: networkEnabled},
		User:          "ayo",
		SetupCommands: setupCommands,
		Labels: map[string]string{
			"ayo.type": "orchestrator",
		},
	}

	if opts.Image == "" {
		opts.Image = "docker.io/library/alpine:3.21"
	}

	sb, err = provider.Create(ctx, opts)
	if err != nil {
		return providers.Sandbox{}, fmt.Errorf("create ayo sandbox: %w", err)
	}

	debug.Log("ayo sandbox created", "id", sb.ID)
	return sb, nil
}

// GetAyoSandbox returns the @ayo sandbox if it exists.
func GetAyoSandbox(ctx context.Context, provider *AppleProvider) (providers.Sandbox, error) {
	return provider.Get(ctx, AyoSandboxName)
}

// StopAyoSandbox stops the @ayo sandbox.
func StopAyoSandbox(ctx context.Context, provider *AppleProvider) error {
	return provider.Stop(ctx, AyoSandboxName, providers.SandboxStopOptions{})
}

// DeleteAyoSandbox deletes the @ayo sandbox.
func DeleteAyoSandbox(ctx context.Context, provider *AppleProvider, force bool) error {
	return provider.Delete(ctx, AyoSandboxName, force)
}

// InitAyoPlanners initializes the planners for @ayo's sandbox.
// This should be called after EnsureAyoSandbox to set up work tracking.
// Returns the initialized planners, which can be used to get tools and instructions.
func InitAyoPlanners(manager *planners.SandboxPlannerManager) (*planners.SandboxPlanners, error) {
	sandboxDir := paths.AyoSandboxDir()
	ayoPlanners, err := manager.GetPlanners("ayo", sandboxDir, nil) // Use global defaults
	if err != nil {
		return nil, fmt.Errorf("init ayo planners: %w", err)
	}
	debug.Log("initialized ayo planners",
		"near", ayoPlanners.NearTerm.Name(),
		"long", ayoPlanners.LongTerm.Name())
	return ayoPlanners, nil
}

// CloseAyoPlanners closes the planners for @ayo's sandbox.
// This should be called when shutting down @ayo to release resources.
func CloseAyoPlanners(manager *planners.SandboxPlannerManager) error {
	return manager.ClosePlanners("ayo")
}

// EnsureAgentHome creates and returns the home directory for an agent running in @ayo's sandbox.
// The directory persists across sessions at ~/.local/share/ayo/sandboxes/ayo/home/{agent}/
func EnsureAgentHome(agentHandle string) (string, error) {
	return paths.EnsureAyoAgentHomeDir(agentHandle)
}

// AgentHomeMount returns a mount configuration for an agent's home directory in @ayo's sandbox.
// The host path is created if it doesn't exist.
// Returns the mount and the container home path for the agent.
func AgentHomeMount(agentHandle string) (providers.Mount, string, error) {
	hostPath, err := paths.EnsureAyoAgentHomeDir(agentHandle)
	if err != nil {
		return providers.Mount{}, "", fmt.Errorf("create agent home: %w", err)
	}

	// Strip @ prefix for container path
	name := agentHandle
	if len(name) > 0 && name[0] == '@' {
		name = name[1:]
	}
	containerPath := "/home/" + name

	mount := providers.Mount{
		Source:      hostPath,
		Destination: containerPath,
		Mode:        providers.MountModeVirtioFS,
		ReadOnly:    false,
	}

	return mount, containerPath, nil
}

// AgentHomeContainerPath returns the container path where an agent's home would be mounted.
// This does not create the directory, only computes the path.
func AgentHomeContainerPath(agentHandle string) string {
	name := agentHandle
	if len(name) > 0 && name[0] == '@' {
		name = name[1:]
	}
	return "/home/" + name
}

// parseMountSpec parses a mount spec string into a Mount.
// Format: "host_path:container_path" or "path" (same on both).
func parseMountSpec(spec string) (providers.Mount, error) {
	var source, dest string

	// Split on first colon (to handle Windows paths)
	for i, c := range spec {
		if c == ':' {
			source = spec[:i]
			dest = spec[i+1:]
			break
		}
	}

	if source == "" {
		// No colon, use same path for both
		source = spec
		dest = spec
	}

	if source == "" || dest == "" {
		return providers.Mount{}, fmt.Errorf("invalid mount spec: %s", spec)
	}

	return providers.Mount{
		Source:      source,
		Destination: dest,
		Mode:        providers.MountModeVirtioFS,
	}, nil
}
