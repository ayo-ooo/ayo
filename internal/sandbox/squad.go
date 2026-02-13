package sandbox

import (
	"context"
	"fmt"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/providers"
)

// SquadSandboxPrefix is the prefix for squad sandbox names.
const SquadSandboxPrefix = "ayo-squad-"

// SquadSandboxName returns the container name for a squad.
func SquadSandboxName(name string) string {
	return SquadSandboxPrefix + name
}

// EnsureSquadSandbox ensures a squad sandbox exists and is running.
// It creates the sandbox if it doesn't exist, or starts it if stopped.
// Returns the sandbox info.
func EnsureSquadSandbox(ctx context.Context, provider *AppleProvider, name string) (providers.Sandbox, error) {
	containerName := SquadSandboxName(name)

	// Check if sandbox already exists
	sb, err := provider.Get(ctx, containerName)
	if err == nil {
		// Sandbox exists
		if sb.Status == providers.SandboxStatusRunning {
			debug.Log("squad sandbox already running", "squad", name, "id", sb.ID)
			return sb, nil
		}
		// Start if stopped
		debug.Log("starting existing squad sandbox", "squad", name, "id", sb.ID)
		if err := provider.Start(ctx, sb.ID); err != nil {
			return providers.Sandbox{}, fmt.Errorf("start squad sandbox: %w", err)
		}
		sb.Status = providers.SandboxStatusRunning
		return sb, nil
	}

	// Sandbox doesn't exist, create it
	debug.Log("creating squad sandbox", "squad", name)

	// Ensure host directories exist
	if err := paths.EnsureSquadDirs(name); err != nil {
		return providers.Sandbox{}, fmt.Errorf("ensure squad dirs: %w", err)
	}

	// Load config
	cfg, err := config.LoadSquadConfig(name)
	if err != nil {
		return providers.Sandbox{}, fmt.Errorf("load squad config: %w", err)
	}

	// Build mounts
	mounts := []providers.Mount{
		// Tickets directory
		{
			Source:      paths.SquadTicketsDir(name),
			Destination: "/.tickets",
			Mode:        providers.MountModeVirtioFS,
		},
		// Context directory
		{
			Source:      paths.SquadContextDir(name),
			Destination: "/.context",
			Mode:        providers.MountModeVirtioFS,
		},
		// Workspace directory
		{
			Source:      paths.SquadWorkspaceDir(name),
			Destination: "/workspace",
			Mode:        providers.MountModeVirtioFS,
		},
	}

	// Add workspace mount if configured
	if cfg.WorkspaceMount != "" {
		mounts = append(mounts, providers.Mount{
			Source:      cfg.WorkspaceMount,
			Destination: "/workspace",
			Mode:        providers.MountModeVirtioFS,
		})
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
		Name:          containerName,
		Image:         cfg.Image,
		Mounts:        mounts,
		Resources:     resources,
		Network:       providers.NetworkConfig{Enabled: networkEnabled},
		SetupCommands: setupCommands,
		Labels: map[string]string{
			"ayo.type":  "squad",
			"ayo.squad": name,
		},
	}

	if opts.Image == "" {
		opts.Image = "docker.io/library/alpine:3.21"
	}

	sb, err = provider.Create(ctx, opts)
	if err != nil {
		return providers.Sandbox{}, fmt.Errorf("create squad sandbox: %w", err)
	}

	debug.Log("squad sandbox created", "squad", name, "id", sb.ID)
	return sb, nil
}

// GetSquadSandbox returns a squad sandbox if it exists.
func GetSquadSandbox(ctx context.Context, provider *AppleProvider, name string) (providers.Sandbox, error) {
	return provider.Get(ctx, SquadSandboxName(name))
}

// StopSquadSandbox stops a squad sandbox.
func StopSquadSandbox(ctx context.Context, provider *AppleProvider, name string) error {
	return provider.Stop(ctx, SquadSandboxName(name), providers.SandboxStopOptions{})
}

// DeleteSquadSandbox deletes a squad sandbox and optionally its data.
func DeleteSquadSandbox(ctx context.Context, provider *AppleProvider, name string, deleteData bool) error {
	containerName := SquadSandboxName(name)

	// Delete the container
	if err := provider.Delete(ctx, containerName, true); err != nil {
		// Ignore not found errors
		debug.Log("delete squad sandbox", "name", name, "error", err)
	}

	// Delete data directories if requested
	if deleteData {
		if err := paths.RemoveSquadDir(name); err != nil {
			return fmt.Errorf("remove squad data: %w", err)
		}
	}

	return nil
}

// ListSquadSandboxes returns all squad sandboxes.
func ListSquadSandboxes(ctx context.Context, provider *AppleProvider) ([]providers.Sandbox, error) {
	all, err := provider.List(ctx)
	if err != nil {
		return nil, err
	}

	var squads []providers.Sandbox
	for _, sb := range all {
		if len(sb.Name) > len(SquadSandboxPrefix) && sb.Name[:len(SquadSandboxPrefix)] == SquadSandboxPrefix {
			squads = append(squads, sb)
		}
	}
	return squads, nil
}

// EnsureSquadAgentUser ensures an agent user exists in a squad sandbox.
func EnsureSquadAgentUser(ctx context.Context, provider *AppleProvider, squadName, agentHandle string) error {
	containerName := SquadSandboxName(squadName)

	// Get the agent's dotfiles path
	dotfilesPath := paths.AgentHomeDir(agentHandle)

	return provider.EnsureAgentUser(ctx, containerName, agentHandle, dotfilesPath)
}
