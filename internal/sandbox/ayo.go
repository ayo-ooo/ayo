package sandbox

import (
	"context"
	"fmt"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/providers"
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
