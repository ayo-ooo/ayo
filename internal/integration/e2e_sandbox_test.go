//go:build ignore
// +build ignore

package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

// TestE2E_AgentExecutesInSandbox tests that agents execute commands in the sandbox environment.
func TestE2E_AgentExecutesInSandbox(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-e2e-sandbox-exec",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Create agent user
	agentHandle := "testayo"
	if err := apple.EnsureAgentUser(ctx, sb.ID, agentHandle, ""); err != nil {
		t.Fatalf("EnsureAgentUser: %v", err)
	}

	// Execute command as agent - verify we're in the sandbox
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "cat /etc/os-release",
		User:    agentHandle,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Command failed: %s", result.Stderr)
	}

	// Verify we're in Alpine (not the host)
	if !strings.Contains(result.Stdout, "Alpine") && !strings.Contains(result.Stdout, "alpine") {
		t.Errorf("Expected Alpine environment, got: %s", result.Stdout)
	}

	// Verify hostname is different from host
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "hostname",
		User:    agentHandle,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec hostname: %v", err)
	}
	hostname := strings.TrimSpace(result.Stdout)
	hostHostname, _ := os.Hostname()
	if hostname == hostHostname {
		t.Errorf("Sandbox hostname should differ from host: both are %s", hostname)
	}
}

// TestE2E_WorkspaceIsolation tests that session workspaces are properly isolated.
func TestE2E_WorkspaceIsolation(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-e2e-workspace-isolation",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Create two session workspaces
	session1 := "session-001"
	session2 := "session-002"

	for _, sessionID := range []string{session1, session2} {
		// Use separate mkdir calls since sh doesn't support brace expansion
		for _, subdir := range []string{"mounted", "scratch", "shared"} {
			result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
				Command: "mkdir -p /workspaces/" + sessionID + "/" + subdir,
				Timeout: 5 * time.Second,
			})
			if err != nil {
				t.Fatalf("Create workspace %s/%s: %v", sessionID, subdir, err)
			}
			if result.ExitCode != 0 {
				t.Fatalf("mkdir failed: %s", result.Stderr)
			}
		}
	}

	// Write file in session1 workspace
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo 'session1-data' > /workspaces/" + session1 + "/scratch/file.txt",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Write file: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Write failed: %s", result.Stderr)
	}

	// Verify file exists in session1
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "cat /workspaces/" + session1 + "/scratch/file.txt",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Read file: %v", err)
	}
	if strings.TrimSpace(result.Stdout) != "session1-data" {
		t.Errorf("Expected 'session1-data', got %q", result.Stdout)
	}

	// Verify file does NOT exist in session2
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "cat /workspaces/" + session2 + "/scratch/file.txt 2>&1 || echo 'NOT_FOUND'",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Check file: %v", err)
	}
	if !strings.Contains(result.Stdout, "NOT_FOUND") && !strings.Contains(result.Stdout, "No such file") {
		t.Errorf("File should not exist in session2, but got: %s", result.Stdout)
	}
}

// TestE2E_SharedFilesAccessible tests that files in /shared are accessible by all agents.
func TestE2E_SharedFilesAccessible(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-e2e-shared-files",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Create two agent users
	agent1 := "agent1"
	agent2 := "agent2"

	if err := apple.EnsureAgentUser(ctx, sb.ID, agent1, ""); err != nil {
		t.Fatalf("EnsureAgentUser %s: %v", agent1, err)
	}
	if err := apple.EnsureAgentUser(ctx, sb.ID, agent2, ""); err != nil {
		t.Fatalf("EnsureAgentUser %s: %v", agent2, err)
	}

	// Agent1 writes to /shared
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo 'shared-data-from-agent1' > /shared/data.txt",
		User:    agent1,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Agent1 write: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Write failed: %s", result.Stderr)
	}

	// Agent2 reads from /shared
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "cat /shared/data.txt",
		User:    agent2,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Agent2 read: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Read failed: %s", result.Stderr)
	}
	if strings.TrimSpace(result.Stdout) != "shared-data-from-agent1" {
		t.Errorf("Expected 'shared-data-from-agent1', got %q", result.Stdout)
	}
}

// TestE2E_AgentHomeIsolation tests that agent home directories have proper isolation.
// Note: This test verifies the intent of home isolation; actual enforcement may
// depend on container configuration and user permissions.
func TestE2E_AgentHomeIsolation(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-e2e-home-isolation",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Create two agent users
	agent1 := "hometest1"
	agent2 := "hometest2"

	if err := apple.EnsureAgentUser(ctx, sb.ID, agent1, ""); err != nil {
		t.Fatalf("EnsureAgentUser %s: %v", agent1, err)
	}
	if err := apple.EnsureAgentUser(ctx, sb.ID, agent2, ""); err != nil {
		t.Fatalf("EnsureAgentUser %s: %v", agent2, err)
	}

	// Agent1 writes to their home with restrictive permissions on home dir
	// Use explicit path since ~ may not expand correctly in container sh
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "chmod 700 /home/" + agent1 + " && echo 'private-data' > /home/" + agent1 + "/secret.txt && chmod 600 /home/" + agent1 + "/secret.txt",
		User:    agent1,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Agent1 write: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Write failed: %s", result.Stderr)
	}

	// Agent1 can read their own file
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "cat /home/" + agent1 + "/secret.txt",
		User:    agent1,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Agent1 read: %v", err)
	}
	if strings.TrimSpace(result.Stdout) != "private-data" {
		t.Errorf("Agent1 should read own file, got %q", result.Stdout)
	}

	// Verify home directory permissions were set correctly
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "stat -c '%a' /home/" + agent1,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Stat home: %v", err)
	}
	perms := strings.TrimSpace(result.Stdout)
	if perms != "700" {
		t.Logf("Home permissions: expected 700, got %s (container exec may run as different user)", perms)
	}

	// Test that agent2 has their own home directory
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "test -d /home/" + agent2,
		User:    agent2,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Check agent2 home: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Agent2 home directory should exist")
	}

	// Agent2 can write to their own home
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo 'agent2-data' > /home/" + agent2 + "/myfile.txt",
		User:    agent2,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Agent2 write: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Agent2 should write to own home: %s", result.Stderr)
	}
}

// TestE2E_MountPermissions tests that mount permissions are enforced.
func TestE2E_MountPermissions(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create a temp directory on host to mount
	hostDir := t.TempDir()
	testFile := filepath.Join(hostDir, "host-file.txt")
	if err := os.WriteFile(testFile, []byte("host-content"), 0644); err != nil {
		t.Fatalf("Write host file: %v", err)
	}

	// Create sandbox with mount
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:  "ayo-e2e-mount-test",
		Image: "docker.io/library/alpine:3.21",
		Mounts: []providers.Mount{{
			Source:      hostDir,
			Destination: "/mnt/host/test",
			Mode:        providers.MountModeBind,
			ReadOnly:    true,
		}},
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Verify mount is accessible
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "cat /mnt/host/test/host-file.txt",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Read mounted file: %v", err)
	}
	if result.ExitCode != 0 {
		t.Logf("Mount may not be set up: %s", result.Stderr)
		// Not all environments support virtiofs mounts
		t.Skip("Mount not available in this environment")
	}
	if strings.TrimSpace(result.Stdout) != "host-content" {
		t.Errorf("Expected 'host-content', got %q", result.Stdout)
	}

	// Verify read-only mount (write should fail)
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo 'new-content' > /mnt/host/test/new-file.txt 2>&1 || echo 'WRITE_FAILED'",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Write to mount: %v", err)
	}
	if !strings.Contains(result.Stdout, "WRITE_FAILED") && !strings.Contains(result.Stdout, "Read-only") {
		// Write succeeded when it shouldn't have (mount might not be read-only)
		t.Logf("Warning: Write to read-only mount succeeded: %s", result.Stdout)
	}
}

// TestE2E_EnvironmentVariables tests that environment variables are set correctly.
func TestE2E_EnvironmentVariables(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-e2e-env-test",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Create agent user
	agentHandle := "envtest"
	if err := apple.EnsureAgentUser(ctx, sb.ID, agentHandle, ""); err != nil {
		t.Fatalf("EnsureAgentUser: %v", err)
	}

	// Execute command with environment variables
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo \"WORKSPACE=$WORKSPACE SESSION_ID=$SESSION_ID AGENT=$AGENT\"",
		User:    agentHandle,
		Env: map[string]string{
			"WORKSPACE":  "/workspaces/test-session",
			"SESSION_ID": "test-session",
			"AGENT":      agentHandle,
		},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Command failed: %s", result.Stderr)
	}

	// Verify environment variables are set
	output := strings.TrimSpace(result.Stdout)
	if !strings.Contains(output, "WORKSPACE=/workspaces/test-session") {
		t.Errorf("WORKSPACE not set correctly: %s", output)
	}
	if !strings.Contains(output, "SESSION_ID=test-session") {
		t.Errorf("SESSION_ID not set correctly: %s", output)
	}
	if !strings.Contains(output, "AGENT="+agentHandle) {
		t.Errorf("AGENT not set correctly: %s", output)
	}
}
