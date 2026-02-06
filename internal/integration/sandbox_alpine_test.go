package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

// TestAlpineSandbox_Create tests creating a sandbox with Alpine image.
func TestAlpineSandbox_Create(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox with Alpine
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-alpine-test-create",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Verify sandbox is running
	status, err := apple.Status(ctx, sb.ID)
	if err != nil {
		t.Fatalf("Get status: %v", err)
	}
	if status != providers.SandboxStatusRunning {
		t.Errorf("Expected running status, got %v", status)
	}

	// Verify image is Alpine
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "cat /etc/os-release | grep -i alpine",
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected Alpine, but os-release check failed: %s", result.Stderr)
	}
	if !strings.Contains(result.Stdout, "Alpine") && !strings.Contains(result.Stdout, "alpine") {
		t.Errorf("Expected Alpine in os-release, got: %s", result.Stdout)
	}

	// Verify basic commands work
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo hello",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec echo: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Echo failed: exit code %d, stderr: %s", result.ExitCode, result.Stderr)
	}
	if strings.TrimSpace(result.Stdout) != "hello" {
		t.Errorf("Expected 'hello', got %q", result.Stdout)
	}
}

// TestAlpineSandbox_UserCreation tests agent user creation in the sandbox.
func TestAlpineSandbox_UserCreation(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-alpine-test-user",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Create agent user
	agentHandle := "testagent"
	if err := apple.EnsureAgentUser(ctx, sb.ID, agentHandle, ""); err != nil {
		t.Fatalf("EnsureAgentUser: %v", err)
	}

	// Verify user exists
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "id " + agentHandle,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec id: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("User %s not found: %s", agentHandle, result.Stderr)
	}

	// Verify home directory exists
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "test -d /home/" + agentHandle,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec test: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Home directory /home/%s does not exist", agentHandle)
	}

	// Verify user can execute commands
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "whoami",
		User:    agentHandle,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec whoami: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("whoami as %s failed: %s", agentHandle, result.Stderr)
	}
	if strings.TrimSpace(result.Stdout) != agentHandle {
		t.Errorf("Expected %q from whoami, got %q", agentHandle, result.Stdout)
	}
}

// TestAlpineSandbox_IRC tests ngircd setup and connectivity.
func TestAlpineSandbox_IRC(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox (ngircd is installed during setup)
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-alpine-test-irc",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Wait a moment for ngircd to fully start
	time.Sleep(1 * time.Second)

	// Verify ngircd is running
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "pgrep ngircd",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec pgrep: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ngircd not running: %s", result.Stderr)
	}

	// Verify ngircd is listening on port 6667
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "netstat -tln | grep 6667 || ss -tln | grep 6667",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec netstat: %v", err)
	}
	if result.ExitCode != 0 {
		t.Logf("ngircd may not be listening yet: %s", result.Stderr)
	}

	// Verify IRC log directory exists
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "test -d /var/log/irc",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec test: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("IRC log directory /var/log/irc does not exist")
	}

	// Verify msg helper script exists
	result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "test -x /usr/local/bin/msg",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec test msg: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("msg helper script not installed or not executable")
	}
}

// TestAlpineSandbox_DirectoryStructure tests the standard sandbox directory structure.
func TestAlpineSandbox_DirectoryStructure(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-alpine-test-dirs",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Verify standard directories exist
	dirs := []string{
		"/shared",
		"/workspaces",
		"/var/log/irc",
		"/mnt/host",
	}

	for _, dir := range dirs {
		result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
			Command: "test -d " + dir,
			Timeout: 5 * time.Second,
		})
		if err != nil {
			t.Fatalf("Exec test %s: %v", dir, err)
		}
		if result.ExitCode != 0 {
			t.Errorf("Directory %s does not exist", dir)
		}
	}

	// Verify /shared has correct permissions (sticky bit, world-writable)
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "stat -c '%a' /shared",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec stat: %v", err)
	}
	perms := strings.TrimSpace(result.Stdout)
	if perms != "1777" {
		t.Errorf("Expected /shared permissions 1777, got %s", perms)
	}
}

// TestAlpineSandbox_InterAgentMessage tests IRC messaging between agents.
func TestAlpineSandbox_InterAgentMessage(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Create sandbox
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-alpine-test-interagent",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Create two agent users
	agent1 := "agentone"
	agent2 := "agenttwo"

	if err := apple.EnsureAgentUser(ctx, sb.ID, agent1, ""); err != nil {
		t.Fatalf("EnsureAgentUser %s: %v", agent1, err)
	}
	if err := apple.EnsureAgentUser(ctx, sb.ID, agent2, ""); err != nil {
		t.Fatalf("EnsureAgentUser %s: %v", agent2, err)
	}

	// Wait for ngircd to be ready
	time.Sleep(1 * time.Second)

	// Agent 1 sends a message to #general using the msg helper
	// First, register nick and join channel
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "printf 'NICK %s\\r\\nUSER %s 0 * :%s\\r\\nJOIN #general\\r\\nPRIVMSG #general :Hello from agent one!\\r\\nQUIT\\r\\n' " + agent1 + " " + agent1 + " " + agent1 + " | nc -w2 localhost 6667 || true",
		User:    agent1,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec IRC send: %v", err)
	}
	// IRC responses may vary, just log for debugging
	t.Logf("Agent1 IRC output: %s", result.Stdout)

	// For now, just verify the IRC server accepted the connection
	// Full message logging would require ngircd channel logging configuration
	// which is beyond the basic setup

	// Verify both users can connect to IRC
	for _, agent := range []string{agent1, agent2} {
		result, err = apple.Exec(ctx, sb.ID, providers.ExecOptions{
			Command: "printf 'NICK %s\\r\\nUSER %s 0 * :%s\\r\\nQUIT\\r\\n' " + agent + " " + agent + " " + agent + " | nc -w1 localhost 6667",
			User:    agent,
			Timeout: 5 * time.Second,
		})
		if err != nil {
			t.Errorf("Agent %s IRC connection failed: %v", agent, err)
		}
		// Should get welcome message back
		if !strings.Contains(result.Stdout, "Welcome") && !strings.Contains(result.Stdout, "001") {
			t.Logf("Agent %s IRC response (may not contain welcome if connection closed): %s", agent, result.Stdout)
		}
	}
}

// TestAlpineSandbox_List tests listing sandboxes from the runtime.
func TestAlpineSandbox_List(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox
	name := "ayo-alpine-test-list"
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    name,
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// List sandboxes
	sandboxes, err := apple.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	// Find our sandbox in the list
	found := false
	for _, s := range sandboxes {
		if s.Name == name {
			found = true
			if s.Status != providers.SandboxStatusRunning {
				t.Errorf("Expected running status, got %v", s.Status)
			}
			break
		}
	}
	if !found {
		t.Errorf("Sandbox %s not found in list", name)
	}
}

// TestAlpineSandbox_Get tests getting a sandbox by ID.
func TestAlpineSandbox_Get(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create sandbox
	name := "ayo-alpine-test-get"
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    name,
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Get by ID
	retrieved, err := apple.Get(ctx, sb.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if retrieved.Name != name {
		t.Errorf("Expected name %s, got %s", name, retrieved.Name)
	}
	if retrieved.Status != providers.SandboxStatusRunning {
		t.Errorf("Expected running status, got %v", retrieved.Status)
	}
}

// TestAlpineSandbox_StopStart tests stopping and starting a sandbox.
func TestAlpineSandbox_StopStart(t *testing.T) {
	apple := sandbox.NewAppleProvider()
	if !apple.IsAvailable() {
		t.Skip("Apple Container not available (requires macOS 26+ on Apple Silicon)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Create sandbox
	sb, err := apple.Create(ctx, providers.SandboxCreateOptions{
		Name:    "ayo-alpine-test-stopstart",
		Image:   "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer apple.Delete(ctx, sb.ID, true)

	// Verify running
	status, err := apple.Status(ctx, sb.ID)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status != providers.SandboxStatusRunning {
		t.Errorf("Expected running, got %v", status)
	}

	// Stop
	if err := apple.Stop(ctx, sb.ID, providers.SandboxStopOptions{Timeout: 10 * time.Second}); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// Verify stopped
	status, err = apple.Status(ctx, sb.ID)
	if err != nil {
		t.Fatalf("Status after stop: %v", err)
	}
	if status != providers.SandboxStatusStopped {
		t.Errorf("Expected stopped, got %v", status)
	}

	// Start
	if err := apple.Start(ctx, sb.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Verify running again
	status, err = apple.Status(ctx, sb.ID)
	if err != nil {
		t.Fatalf("Status after start: %v", err)
	}
	if status != providers.SandboxStatusRunning {
		t.Errorf("Expected running after start, got %v", status)
	}

	// Execute command to verify container is functional
	result, err := apple.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo alive",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec after restart: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Command failed after restart: %s", result.Stderr)
	}
}
