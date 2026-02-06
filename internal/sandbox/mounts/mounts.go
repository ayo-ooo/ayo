// Package mounts provides filesystem mount handling for sandbox environments.
package mounts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alexcabrera/ayo/internal/providers"
)

// MountValidator validates mount configurations.
type MountValidator struct {
	allowedPrefixes []string
	blockedPaths    []string
}

// NewMountValidator creates a new mount validator with default security rules.
func NewMountValidator() *MountValidator {
	return &MountValidator{
		allowedPrefixes: []string{
			"/tmp",
			"/var/tmp",
			os.Getenv("HOME"),
		},
		blockedPaths: []string{
			"/",
			"/etc",
			"/var",
			"/usr",
			"/bin",
			"/sbin",
			"/lib",
			"/lib64",
			"/boot",
			"/dev",
			"/proc",
			"/sys",
			"/run",
		},
	}
}

// Validate checks if a mount configuration is safe.
func (v *MountValidator) Validate(mount providers.Mount) error {
	// Validate source path exists
	source := mount.Source
	if source == "" {
		return fmt.Errorf("mount source path is empty")
	}

	// Expand home directory
	if strings.HasPrefix(source, "~") {
		source = filepath.Join(os.Getenv("HOME"), source[1:])
	}

	// Check if path exists
	info, err := os.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("mount source does not exist: %s", source)
		}
		return fmt.Errorf("cannot access mount source: %w", err)
	}

	// For virtiofs, source must be a directory
	if mount.Mode == providers.MountModeVirtioFS && !info.IsDir() {
		return fmt.Errorf("virtiofs mount source must be a directory: %s", source)
	}

	// Check blocked paths
	absSource, _ := filepath.Abs(source)
	for _, blocked := range v.blockedPaths {
		if absSource == blocked {
			return fmt.Errorf("mount source is a blocked system path: %s", source)
		}
	}

	// Validate destination path
	if mount.Destination == "" {
		return fmt.Errorf("mount destination path is empty")
	}
	if !filepath.IsAbs(mount.Destination) {
		return fmt.Errorf("mount destination must be absolute: %s", mount.Destination)
	}

	return nil
}

// PrepareMount prepares a mount for use, handling path resolution and defaults.
func PrepareMount(mount providers.Mount) (providers.Mount, error) {
	result := mount

	// Expand home directory in source
	if strings.HasPrefix(result.Source, "~") {
		result.Source = filepath.Join(os.Getenv("HOME"), result.Source[1:])
	}

	// Convert to absolute path
	absSource, err := filepath.Abs(result.Source)
	if err != nil {
		return result, fmt.Errorf("cannot resolve source path: %w", err)
	}
	result.Source = absSource

	// Set default mode based on provider availability
	if result.Mode == "" {
		result.Mode = DefaultMountMode()
	}

	return result, nil
}

// DefaultMountMode returns the best available mount mode for the current platform.
func DefaultMountMode() providers.MountMode {
	// Use virtiofs on macOS with Apple Silicon (Apple Container)
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return providers.MountModeVirtioFS
	}
	// Default to bind mounts elsewhere
	return providers.MountModeBind
}

// IsVirtioFSAvailable checks if virtiofs is available on the current system.
func IsVirtioFSAvailable() bool {
	// virtiofs is available on:
	// - macOS with Apple Container (Virtualization.framework)
	// - Linux with virtiofsd
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		// Check if Apple Container is available
		_, err := os.Stat("/usr/local/bin/container")
		return err == nil
	}
	if runtime.GOOS == "linux" {
		// Check for virtiofsd
		_, err := os.Stat("/usr/lib/virtiofsd")
		if err == nil {
			return true
		}
		_, err = os.Stat("/usr/libexec/virtiofsd")
		return err == nil
	}
	return false
}

// AppleContainerMountArgs converts a mount to Apple Container CLI arguments.
func AppleContainerMountArgs(mount providers.Mount) []string {
	var args []string

	switch mount.Mode {
	case providers.MountModeVirtioFS, providers.MountModeBind:
		// Apple Container uses --volume for virtiofs mounts
		mountOpt := fmt.Sprintf("%s:%s", mount.Source, mount.Destination)
		if mount.ReadOnly {
			mountOpt += ":ro"
		}
		args = append(args, "--volume", mountOpt)

	case providers.MountModeTmpfs:
		// Apple Container supports tmpfs via --tmpfs
		args = append(args, "--tmpfs", mount.Destination)

	case providers.MountModeOverlay:
		// Fallback to virtiofs for overlay
		mountOpt := fmt.Sprintf("%s:%s", mount.Source, mount.Destination)
		if mount.ReadOnly {
			mountOpt += ":ro"
		}
		args = append(args, "--volume", mountOpt)
	}

	return args
}

// LinuxMountArgs converts a mount to systemd-nspawn CLI arguments.
func LinuxMountArgs(mount providers.Mount) []string {
	var args []string

	switch mount.Mode {
	case providers.MountModeBind, providers.MountModeVirtioFS:
		// systemd-nspawn uses --bind for read-write, --bind-ro for read-only
		if mount.ReadOnly {
			args = append(args, fmt.Sprintf("--bind-ro=%s:%s", mount.Source, mount.Destination))
		} else {
			args = append(args, fmt.Sprintf("--bind=%s:%s", mount.Source, mount.Destination))
		}

	case providers.MountModeTmpfs:
		// systemd-nspawn supports tmpfs via --tmpfs
		args = append(args, fmt.Sprintf("--tmpfs=%s", mount.Destination))

	case providers.MountModeOverlay:
		// Fallback to bind for overlay
		if mount.ReadOnly {
			args = append(args, fmt.Sprintf("--bind-ro=%s:%s", mount.Source, mount.Destination))
		} else {
			args = append(args, fmt.Sprintf("--bind=%s:%s", mount.Source, mount.Destination))
		}
	}

	return args
}

// ProjectMount creates a mount for the project directory.
func ProjectMount(projectPath string) (providers.Mount, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return providers.Mount{}, fmt.Errorf("cannot resolve project path: %w", err)
	}

	return providers.Mount{
		Source:      absPath,
		Destination: "/workspace",
		Mode:        DefaultMountMode(),
		ReadOnly:    false,
	}, nil
}

// ReadOnlyProjectMount creates a read-only mount for the project directory.
func ReadOnlyProjectMount(projectPath string) (providers.Mount, error) {
	mount, err := ProjectMount(projectPath)
	if err != nil {
		return mount, err
	}
	mount.ReadOnly = true
	return mount, nil
}

// HomeMount creates a mount for a subdirectory of the home directory.
func HomeMount(subdir, destination string) (providers.Mount, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return providers.Mount{}, fmt.Errorf("HOME environment variable not set")
	}

	source := filepath.Join(home, subdir)
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return providers.Mount{}, fmt.Errorf("home subdirectory does not exist: %s", source)
	}

	return providers.Mount{
		Source:      source,
		Destination: destination,
		Mode:        DefaultMountMode(),
		ReadOnly:    false,
	}, nil
}

// TmpfsMount creates a tmpfs (memory-backed) mount.
func TmpfsMount(destination string) providers.Mount {
	return providers.Mount{
		Destination: destination,
		Mode:        providers.MountModeTmpfs,
	}
}
