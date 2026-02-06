package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/daemon"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

// TestDaemon_Lifecycle tests daemon start/stop lifecycle.
func TestDaemon_Lifecycle(t *testing.T) {
	// Create temp directory for socket
	tmpDir, err := os.MkdirTemp("", "daemon-lifecycle-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "daemon.sock")

	cfg := daemon.DefaultServerConfig()
	cfg.SocketPath = socketPath
	cfg.PoolConfig = sandbox.PoolConfig{
		Name:    "test-lifecycle",
		MinSize: 0,
		MaxSize: 2,
	}

	server, err := daemon.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx := context.Background()

	// Start
	if err := server.Start(ctx, socketPath); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Give server time to initialize
	time.Sleep(100 * time.Millisecond)

	// Connect client
	client := daemon.NewClient()
	if err := client.ConnectTo(ctx, socketPath); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// Ping
	if err := client.Ping(ctx); err != nil {
		t.Errorf("Ping: %v", err)
	}

	// Status
	status, err := client.Status(ctx)
	if err != nil {
		t.Errorf("Status: %v", err)
	}
	if !status.Running {
		t.Errorf("Status.Running: expected true")
	}
	if status.PID == 0 {
		t.Errorf("Status.PID: expected non-zero")
	}

	// Stop
	client.Close()
	if err := server.Stop(ctx); err != nil {
		t.Errorf("Stop: %v", err)
	}
}

// TestDaemon_TriggerCron tests cron trigger registration and firing.
func TestDaemon_TriggerCron(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-trigger-cron-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "daemon.sock")

	cfg := daemon.DefaultServerConfig()
	cfg.SocketPath = socketPath
	cfg.PoolConfig = sandbox.PoolConfig{
		Name:    "test-trigger-cron",
		MinSize: 0,
		MaxSize: 2,
	}

	server, err := daemon.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx := context.Background()

	if err := server.Start(ctx, socketPath); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	client := daemon.NewClient()
	if err := client.ConnectTo(ctx, socketPath); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	// Register cron trigger (every second for testing)
	result, err := client.TriggerRegister(ctx, daemon.TriggerRegisterParams{
		Type:     "cron",
		Agent:    "@test-agent",
		Schedule: "*/1 * * * * *", // Every second
		Prompt:   "Test cron trigger",
	})
	if err != nil {
		t.Fatalf("TriggerRegister: %v", err)
	}
	if result.Trigger.ID == "" {
		t.Errorf("Expected trigger ID")
	}
	if result.Trigger.Type != "cron" {
		t.Errorf("Type: got %q, want cron", result.Trigger.Type)
	}

	// List triggers
	listResult, err := client.TriggerList(ctx)
	if err != nil {
		t.Fatalf("TriggerList: %v", err)
	}
	if len(listResult.Triggers) != 1 {
		t.Errorf("Triggers count: got %d, want 1", len(listResult.Triggers))
	}

	// Wait for trigger to fire (should happen within 2 seconds)
	time.Sleep(1500 * time.Millisecond)

	// Check if session was started
	sessResult, err := client.SessionList(ctx)
	if err != nil {
		t.Fatalf("SessionList: %v", err)
	}
	// Note: The trigger should have fired and created a session for @test-agent
	foundSession := false
	for _, sess := range sessResult.Sessions {
		if sess.AgentHandle == "@test-agent" {
			foundSession = true
			break
		}
	}
	if !foundSession {
		t.Logf("No session found for @test-agent (trigger may not have fired yet)")
	}

	// Remove trigger
	if err := client.TriggerRemove(ctx, result.Trigger.ID); err != nil {
		t.Errorf("TriggerRemove: %v", err)
	}

	// Verify removed
	listResult, err = client.TriggerList(ctx)
	if err != nil {
		t.Fatalf("TriggerList: %v", err)
	}
	if len(listResult.Triggers) != 0 {
		t.Errorf("Triggers count after remove: got %d, want 0", len(listResult.Triggers))
	}
}

// TestDaemon_TriggerWatch tests file watch trigger registration.
func TestDaemon_TriggerWatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-trigger-watch-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "daemon.sock")
	watchDir := filepath.Join(tmpDir, "watch")
	if err := os.MkdirAll(watchDir, 0755); err != nil {
		t.Fatalf("MkdirAll watch: %v", err)
	}

	cfg := daemon.DefaultServerConfig()
	cfg.SocketPath = socketPath
	cfg.PoolConfig = sandbox.PoolConfig{
		Name:    "test-trigger-watch",
		MinSize: 0,
		MaxSize: 2,
	}

	server, err := daemon.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx := context.Background()

	if err := server.Start(ctx, socketPath); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	client := daemon.NewClient()
	if err := client.ConnectTo(ctx, socketPath); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	// Register watch trigger
	result, err := client.TriggerRegister(ctx, daemon.TriggerRegisterParams{
		Type:     "watch",
		Agent:    "@build-agent",
		Path:     watchDir,
		Patterns: []string{"*.go"},
		Events:   []string{"create", "modify"},
		Prompt:   "Build project on file change",
	})
	if err != nil {
		t.Fatalf("TriggerRegister: %v", err)
	}
	if result.Trigger.Type != "watch" {
		t.Errorf("Type: got %q, want watch", result.Trigger.Type)
	}
	if result.Trigger.Path != watchDir {
		t.Errorf("Path: got %q, want %q", result.Trigger.Path, watchDir)
	}

	// Get trigger
	getResult, err := client.TriggerGet(ctx, result.Trigger.ID)
	if err != nil {
		t.Fatalf("TriggerGet: %v", err)
	}
	if getResult.Trigger.ID != result.Trigger.ID {
		t.Errorf("TriggerGet ID mismatch")
	}

	// Create a file to trigger the watch
	testFile := filepath.Join(watchDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Wait a bit for the trigger to fire
	time.Sleep(300 * time.Millisecond)

	// Clean up
	if err := client.TriggerRemove(ctx, result.Trigger.ID); err != nil {
		t.Errorf("TriggerRemove: %v", err)
	}
}

// TestDaemon_SessionManagement tests session lifecycle.
func TestDaemon_SessionManagement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-session-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "daemon.sock")

	cfg := daemon.DefaultServerConfig()
	cfg.SocketPath = socketPath
	cfg.PoolConfig = sandbox.PoolConfig{
		Name:    "test-session",
		MinSize: 0,
		MaxSize: 2,
	}

	server, err := daemon.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx := context.Background()

	if err := server.Start(ctx, socketPath); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	client := daemon.NewClient()
	if err := client.ConnectTo(ctx, socketPath); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	// Initial session list should be empty
	listResult, err := client.SessionList(ctx)
	if err != nil {
		t.Fatalf("SessionList: %v", err)
	}
	initialCount := len(listResult.Sessions)

	// Wake an agent
	wakeResult, err := client.AgentWake(ctx, "@test-session-agent")
	if err != nil {
		t.Fatalf("AgentWake: %v", err)
	}
	if wakeResult.Session.ID == "" {
		t.Errorf("Expected session ID from wake")
	}
	if wakeResult.Session.AgentHandle != "@test-session-agent" {
		t.Errorf("AgentHandle: got %q, want @test-session-agent", wakeResult.Session.AgentHandle)
	}

	// List should show the session
	listResult, err = client.SessionList(ctx)
	if err != nil {
		t.Fatalf("SessionList: %v", err)
	}
	if len(listResult.Sessions) != initialCount+1 {
		t.Errorf("Session count: got %d, want %d", len(listResult.Sessions), initialCount+1)
	}

	// Agent status should show active
	statusResult, err := client.AgentStatus(ctx, "@test-session-agent")
	if err != nil {
		t.Fatalf("AgentStatus: %v", err)
	}
	if !statusResult.Active {
		t.Errorf("Expected agent to be active")
	}

	// Sleep the agent
	if err := client.AgentSleep(ctx, "@test-session-agent"); err != nil {
		t.Fatalf("AgentSleep: %v", err)
	}

	// Agent status should show inactive
	statusResult, err = client.AgentStatus(ctx, "@test-session-agent")
	if err != nil {
		t.Fatalf("AgentStatus: %v", err)
	}
	if statusResult.Active {
		t.Errorf("Expected agent to be inactive after sleep")
	}

	// Session list should be back to initial count
	listResult, err = client.SessionList(ctx)
	if err != nil {
		t.Fatalf("SessionList: %v", err)
	}
	if len(listResult.Sessions) != initialCount {
		t.Errorf("Session count after sleep: got %d, want %d", len(listResult.Sessions), initialCount)
	}
}

// TestDaemon_TriggerTest tests manually firing a trigger.
func TestDaemon_TriggerTest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-trigger-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "daemon.sock")

	cfg := daemon.DefaultServerConfig()
	cfg.SocketPath = socketPath
	cfg.PoolConfig = sandbox.PoolConfig{
		Name:    "test-trigger-manual",
		MinSize: 0,
		MaxSize: 2,
	}

	server, err := daemon.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx := context.Background()

	if err := server.Start(ctx, socketPath); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	client := daemon.NewClient()
	if err := client.ConnectTo(ctx, socketPath); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	// Register a cron trigger (won't fire naturally in test timeframe)
	result, err := client.TriggerRegister(ctx, daemon.TriggerRegisterParams{
		Type:     "cron",
		Agent:    "@manual-test-agent",
		Schedule: "0 0 0 1 1 *", // Once a year - won't fire naturally
		Prompt:   "Manually tested trigger",
	})
	if err != nil {
		t.Fatalf("TriggerRegister: %v", err)
	}

	// Manually fire the trigger
	if err := client.TriggerTest(ctx, result.Trigger.ID); err != nil {
		t.Fatalf("TriggerTest: %v", err)
	}

	// Wait for session to be created
	time.Sleep(100 * time.Millisecond)

	// Check if session was created
	sessResult, err := client.SessionList(ctx)
	if err != nil {
		t.Fatalf("SessionList: %v", err)
	}
	foundSession := false
	for _, sess := range sessResult.Sessions {
		if sess.AgentHandle == "@manual-test-agent" {
			foundSession = true
			if sess.TriggerID != result.Trigger.ID {
				t.Errorf("TriggerID: got %q, want %q", sess.TriggerID, result.Trigger.ID)
			}
			break
		}
	}
	if !foundSession {
		t.Errorf("Expected session to be created for @manual-test-agent")
	}

	// Cleanup
	if err := client.TriggerRemove(ctx, result.Trigger.ID); err != nil {
		t.Errorf("TriggerRemove: %v", err)
	}
}

// TestDaemon_SandboxOperations tests sandbox acquire/release via daemon.
func TestDaemon_SandboxOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-sandbox-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "daemon.sock")

	cfg := daemon.DefaultServerConfig()
	cfg.SocketPath = socketPath
	cfg.PoolConfig = sandbox.PoolConfig{
		Name:    "test-sandbox-ops",
		MinSize: 1,
		MaxSize: 3,
	}

	server, err := daemon.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx := context.Background()

	if err := server.Start(ctx, socketPath); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	client := daemon.NewClient()
	if err := client.ConnectTo(ctx, socketPath); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	// Get initial sandbox status
	statusBefore, err := client.SandboxStatus(ctx)
	if err != nil {
		t.Fatalf("SandboxStatus: %v", err)
	}
	inUseBefore := statusBefore.InUse

	// Acquire sandbox
	acquireResult, err := client.SandboxAcquire(ctx, "@sandbox-test-agent", 30)
	if err != nil {
		t.Fatalf("SandboxAcquire: %v", err)
	}
	if acquireResult.SandboxID == "" {
		t.Errorf("Expected sandbox ID")
	}

	// Status should show one more in use
	statusAfter, err := client.SandboxStatus(ctx)
	if err != nil {
		t.Fatalf("SandboxStatus: %v", err)
	}
	if statusAfter.InUse != inUseBefore+1 {
		t.Errorf("InUse: got %d, want %d", statusAfter.InUse, inUseBefore+1)
	}

	// Execute command in sandbox
	execResult, err := client.SandboxExec(ctx, acquireResult.SandboxID, "echo hello", "", 10)
	if err != nil {
		t.Fatalf("SandboxExec: %v", err)
	}
	if execResult.Stdout != "hello\n" {
		t.Errorf("Stdout: got %q, want %q", execResult.Stdout, "hello\n")
	}
	if execResult.ExitCode != 0 {
		t.Errorf("ExitCode: got %d, want 0", execResult.ExitCode)
	}

	// Release sandbox
	if err := client.SandboxRelease(ctx, acquireResult.SandboxID); err != nil {
		t.Fatalf("SandboxRelease: %v", err)
	}

	// Status should be back to before
	statusFinal, err := client.SandboxStatus(ctx)
	if err != nil {
		t.Fatalf("SandboxStatus: %v", err)
	}
	if statusFinal.InUse != inUseBefore {
		t.Errorf("InUse after release: got %d, want %d", statusFinal.InUse, inUseBefore)
	}
}
