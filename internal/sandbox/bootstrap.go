// Package sandbox provides container-based agent execution environments.
// This file implements the bootstrap process for ayod-enabled sandboxes.
package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/alexcabrera/ayo/internal/ayod"
	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/providers"
)

// BootstrapOptions configures sandbox bootstrap.
type BootstrapOptions struct {
	// Provider is the sandbox provider to use.
	Provider providers.SandboxProvider

	// SandboxID is the ID of the sandbox to bootstrap.
	SandboxID string

	// ContainerName is the name of the container.
	ContainerName string

	// InitialUser is the username to create during bootstrap (e.g., "ayo").
	InitialUser string

	// SocketMountPath is the host path where the socket should be accessible.
	// This is mounted into the container so the host can connect to ayod.
	SocketMountPath string
}

// BootstrapResult contains the result of bootstrapping a sandbox.
type BootstrapResult struct {
	// AyodClient is the connected ayod client.
	AyodClient *ayod.Client

	// SocketPath is the path to the ayod socket (host-side).
	SocketPath string

	// InitialUserUID is the UID of the initial user.
	InitialUserUID int
}

// Bootstrap initializes a sandbox with ayod and standard directory structure.
// This should be called after the sandbox container is created but before
// running agent commands.
func Bootstrap(ctx context.Context, opts BootstrapOptions) (*BootstrapResult, error) {
	debug.Log("bootstrapping sandbox", "id", opts.SandboxID, "user", opts.InitialUser)

	// 1. Copy ayod binary into sandbox
	ayodPath, err := getAyodBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("locate ayod binary: %w", err)
	}

	debug.Log("copying ayod binary", "source", ayodPath)
	if err := copyFileToContainer(ctx, opts.ContainerName, ayodPath, "/usr/local/bin/ayod"); err != nil {
		return nil, fmt.Errorf("copy ayod to sandbox: %w", err)
	}

	// 2. Start ayod inside the container
	// We execute ayod in the background - it will listen on /run/ayod.sock
	if err := startAyodInContainer(ctx, opts.ContainerName); err != nil {
		return nil, fmt.Errorf("start ayod: %w", err)
	}

	// 3. Wait for ayod to be ready
	socketPath := opts.SocketMountPath
	if socketPath == "" {
		socketPath = filepath.Join(paths.DataDir(), "sandboxes", opts.SandboxID, "ayod.sock")
	}

	client, err := waitForAyod(ctx, socketPath, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect to ayod: %w", err)
	}

	// 4. Create initial user if specified
	var uid int
	if opts.InitialUser != "" {
		resp, err := client.UserAdd(ayod.UserAddRequest{
			Username: opts.InitialUser,
			Shell:    "/bin/sh",
		})
		if err != nil {
			client.Close()
			return nil, fmt.Errorf("create initial user: %w", err)
		}
		uid = resp.UID
		debug.Log("created initial user", "username", opts.InitialUser, "uid", uid)
	}

	// 5. Create standard directories via ayod
	if err := setupStandardDirectories(client); err != nil {
		debug.Log("warning: failed to set up directories", "error", err)
		// Non-fatal - continue anyway
	}

	return &BootstrapResult{
		AyodClient:     client,
		SocketPath:     socketPath,
		InitialUserUID: uid,
	}, nil
}

// getAyodBinaryPath returns the path to the ayod binary.
// It looks in several locations:
// 1. Build directory (for development)
// 2. Same directory as the ayo binary
// 3. Standard installation paths
func getAyodBinaryPath() (string, error) {
	// Check for platform-specific binary
	binaryName := "ayod-linux-amd64"
	if runtime.GOARCH == "arm64" {
		binaryName = "ayod-linux-arm64"
	}

	// 1. Build directory (development)
	buildPath := filepath.Join("build", binaryName)
	if _, err := os.Stat(buildPath); err == nil {
		return filepath.Abs(buildPath)
	}

	// 2. Same directory as ayo executable
	exe, err := os.Executable()
	if err == nil {
		siblingPath := filepath.Join(filepath.Dir(exe), binaryName)
		if _, err := os.Stat(siblingPath); err == nil {
			return siblingPath, nil
		}
	}

	// 3. Data directory
	dataPath := filepath.Join(paths.DataDir(), "bin", binaryName)
	if _, err := os.Stat(dataPath); err == nil {
		return dataPath, nil
	}

	// 4. Check for generic "ayod" in PATH (for local builds)
	if path, err := exec.LookPath("ayod"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("ayod binary not found (tried build/, alongside executable, %s)", dataPath)
}

// copyFileToContainer copies a file from the host into a container.
func copyFileToContainer(ctx context.Context, containerName, hostPath, containerPath string) error {
	// Use container cp command (Apple Container supports this)
	// The exact command depends on the container runtime
	cmd := exec.CommandContext(ctx, "container", "cp", hostPath, containerName+":"+containerPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("container cp: %w: %s", err, string(output))
	}

	// Make executable
	chmodCmd := exec.CommandContext(ctx, "container", "exec", containerName, "chmod", "+x", containerPath)
	if output, err := chmodCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("chmod: %w: %s", err, string(output))
	}

	return nil
}

// startAyodInContainer starts the ayod daemon inside the container.
func startAyodInContainer(ctx context.Context, containerName string) error {
	// Start ayod in background - it will daemonize itself
	// We use nohup to ensure it survives after exec returns
	cmd := exec.CommandContext(ctx, "container", "exec", containerName,
		"sh", "-c", "nohup /usr/local/bin/ayod > /var/log/ayod.log 2>&1 &")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("start ayod: %w: %s", err, string(output))
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	return nil
}

// waitForAyod waits for ayod to become available on the socket.
func waitForAyod(ctx context.Context, socketPath string, timeout time.Duration) (*ayod.Client, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		client, err := ayod.Connect(socketPath)
		if err == nil {
			// Verify it's responsive
			if _, err := client.Health(); err == nil {
				return client, nil
			}
			client.Close()
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("timeout waiting for ayod on %s", socketPath)
}

// setupStandardDirectories creates the standard sandbox directory structure via ayod.
func setupStandardDirectories(client *ayod.Client) error {
	// Create directories using ayod's WriteFile (which creates parent dirs)
	// We create a placeholder file in each directory
	dirs := []struct {
		path string
		mode os.FileMode
	}{
		{"/workspace/.keep", 0644},
		{"/output/.keep", 0644},
		{"/mnt/.keep", 0644},
	}

	for _, d := range dirs {
		err := client.WriteFile(ayod.WriteFileRequest{
			Path:    d.path,
			Content: []byte("# Created by ayo bootstrap\n"),
			Mode:    d.mode,
		})
		if err != nil {
			return fmt.Errorf("create %s: %w", d.path, err)
		}
	}

	return nil
}
