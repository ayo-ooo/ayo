// Package integration provides integration and end-to-end tests for ayo.
// This file contains end-to-end tests that validate complete workflows.
package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/memory/zettelkasten"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/session/jsonl"
)

// TestE2E_SandboxWorkflow tests a complete sandbox workflow:
// Create -> Execute -> Get Results -> Cleanup
func TestE2E_SandboxWorkflow(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	ctx, cancel := env.Context()
	defer cancel()

	// 1. Create a sandbox
	sb, err := env.SandboxProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "e2e-test",
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}

	// 2. Execute multiple commands
	commands := []struct {
		cmd    string
		expect string
	}{
		{"echo hello", "hello\n"},
		{"echo world", "world\n"},
		{"pwd", ""}, // Just verify it runs
	}

	for _, tc := range commands {
		result, err := env.SandboxProvider.Exec(ctx, sb.ID, providers.ExecOptions{
			Command: tc.cmd,
			Timeout: 10 * time.Second,
		})
		if err != nil {
			t.Errorf("Exec %q: %v", tc.cmd, err)
			continue
		}
		if result.ExitCode != 0 {
			t.Errorf("Exec %q: exit code %d", tc.cmd, result.ExitCode)
		}
		if tc.expect != "" && result.Stdout != tc.expect {
			t.Errorf("Exec %q: got %q, want %q", tc.cmd, result.Stdout, tc.expect)
		}
	}

	// 3. Check status
	status, err := env.SandboxProvider.Status(ctx, sb.ID)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status != providers.SandboxStatusRunning {
		t.Errorf("Status: got %v, want running", status)
	}

	// 4. Cleanup
	if err := env.SandboxProvider.Delete(ctx, sb.ID, true); err != nil {
		t.Errorf("Delete: %v", err)
	}
}

// TestE2E_SessionPersistence tests session write and read workflow.
func TestE2E_SessionPersistence(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Setup session structure
	structure := jsonl.NewStructure(env.SessionDir)
	if err := structure.Initialize(); err != nil {
		t.Fatalf("Initialize session structure: %v", err)
	}

	now := time.Now()
	sess := session.Session{
		ID:          "e2e-sess-001",
		AgentHandle: "@ayo",
		Title:       "E2E Test Session",
		Source:      "ayo",
		CreatedAt:   now.Unix(),
		UpdatedAt:   now.Unix(),
	}

	// 1. Write session
	writer, err := jsonl.NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	// 2. Write messages
	messages := []session.Message{
		{
			ID:        "msg-001",
			SessionID: sess.ID,
			Role:      session.RoleUser,
			CreatedAt: now.Unix(),
			UpdatedAt: now.Unix(),
			Parts:     []session.ContentPart{session.TextContent{Text: "Hello"}},
		},
		{
			ID:        "msg-002",
			SessionID: sess.ID,
			Role:      session.RoleAssistant,
			Model:     "test-model",
			CreatedAt: now.Unix(),
			UpdatedAt: now.Unix(),
			Parts:     []session.ContentPart{session.TextContent{Text: "Hello! How can I help?"}},
		},
	}

	for _, msg := range messages {
		if err := writer.WriteMessage(msg); err != nil {
			t.Fatalf("WriteMessage: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close writer: %v", err)
	}

	// 3. Read session back
	readSess, readMsgs, err := jsonl.ReadSession(writer.Path())
	if err != nil {
		t.Fatalf("ReadSession: %v", err)
	}

	// 4. Verify
	if readSess.ID != sess.ID {
		t.Errorf("Session ID: got %q, want %q", readSess.ID, sess.ID)
	}
	if readSess.Title != sess.Title {
		t.Errorf("Session Title: got %q, want %q", readSess.Title, sess.Title)
	}
	if len(readMsgs) != len(messages) {
		t.Fatalf("Message count: got %d, want %d", len(readMsgs), len(messages))
	}
	if readMsgs[0].ID != messages[0].ID {
		t.Errorf("Message[0].ID: got %q, want %q", readMsgs[0].ID, messages[0].ID)
	}
}

// TestE2E_MemoryPersistence tests memory write and read workflow.
func TestE2E_MemoryPersistence(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Setup memory structure
	structure := zettelkasten.NewStructure(env.MemoryDir)
	if err := structure.Initialize(); err != nil {
		t.Fatalf("Initialize memory structure: %v", err)
	}

	now := time.Now()

	// 1. Create memory file
	mem := &zettelkasten.MemoryFile{
		Frontmatter: zettelkasten.Frontmatter{
			ID:       "mem-e2e-001",
			Created:  now,
			Category: "preference",
			Topics:   []string{"testing", "e2e"},
		},
		Content: "User prefers detailed error messages in tests.",
	}

	// 2. Write memory
	memPath := filepath.Join(structure.CategoryDir("preference"), "e2e-test.md")
	if err := mem.WriteFile(memPath); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// 3. Read memory back
	readMem, err := zettelkasten.ParseFile(memPath)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	// 4. Verify
	if readMem.Frontmatter.ID != mem.Frontmatter.ID {
		t.Errorf("Memory ID: got %q, want %q", readMem.Frontmatter.ID, mem.Frontmatter.ID)
	}
	if readMem.Frontmatter.Category != mem.Frontmatter.Category {
		t.Errorf("Memory Category: got %q, want %q", readMem.Frontmatter.Category, mem.Frontmatter.Category)
	}
	if readMem.Content != mem.Content {
		t.Errorf("Memory Content: got %q, want %q", readMem.Content, mem.Content)
	}
}

// TestE2E_AgentSetup tests creating an agent with skills.
func TestE2E_AgentSetup(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	// 1. Create agent
	agentDir := env.CreateAgent("@e2e-agent", "End-to-end test agent", "You are a test agent for E2E testing.")

	// 2. Create agent-specific skill
	skillDir := filepath.Join(agentDir, "skills", "e2e-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Create skill dir: %v", err)
	}

	skillContent := `---
name: e2e-skill
description: A skill for end-to-end testing.
---

# E2E Testing Skill

This skill provides testing capabilities.
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatalf("Write SKILL.md: %v", err)
	}

	// 3. Verify files exist
	expectedFiles := []string{
		filepath.Join(agentDir, "config.json"),
		filepath.Join(agentDir, "system.md"),
		filepath.Join(skillDir, "SKILL.md"),
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("Expected file %s not found: %v", f, err)
		}
	}
}

// TestE2E_PooledExecution tests sandbox pool acquire/release workflow.
func TestE2E_PooledExecution(t *testing.T) {
	provider := sandbox.NewNoneProvider()
	pool := sandbox.NewPool(sandbox.PoolConfig{
		Name:    "e2e-pool",
		MinSize: 1,
		MaxSize: 3,
	}, provider)

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		t.Fatalf("Start pool: %v", err)
	}
	defer pool.Stop(ctx)

	// 1. Acquire sandbox for agent
	sb1, err := pool.Acquire(ctx, "@ayo")
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	// 2. Execute command
	result, err := pool.Exec(ctx, sb1.ID, providers.ExecOptions{
		Command: "echo pooled",
	})
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if result.Stdout != "pooled\n" {
		t.Errorf("Stdout: got %q, want %q", result.Stdout, "pooled\n")
	}

	// 3. Same agent gets same sandbox
	sb2, _ := pool.Acquire(ctx, "@ayo")
	if sb1.ID != sb2.ID {
		t.Errorf("Same agent should get same sandbox")
	}

	// 4. Different agent gets different sandbox
	sb3, err := pool.Acquire(ctx, "@other")
	if err != nil {
		t.Fatalf("Acquire other: %v", err)
	}
	if sb1.ID == sb3.ID {
		t.Errorf("Different agent should get different sandbox")
	}

	// 5. Release and verify pool status
	if err := pool.Release(ctx, sb1.ID); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if err := pool.Release(ctx, sb3.ID); err != nil {
		t.Fatalf("Release other: %v", err)
	}

	status := pool.Status()
	if status.InUse != 0 {
		t.Errorf("InUse after release: got %d, want 0", status.InUse)
	}
}

// TestE2E_MockProvider tests using mock provider for CI scenarios.
func TestE2E_MockProvider(t *testing.T) {
	mock := sandbox.NewMockProvider()

	// Configure custom behavior
	execCount := 0
	mock.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		execCount++
		return providers.ExecResult{
			Stdout:   "mock response " + string(rune('0'+execCount)),
			ExitCode: 0,
		}, nil
	}

	ctx := context.Background()

	// Create and exec
	sb, _ := mock.Create(ctx, providers.SandboxCreateOptions{Name: "mock-e2e"})

	for i := 0; i < 3; i++ {
		_, err := mock.Exec(ctx, sb.ID, providers.ExecOptions{Command: "test"})
		if err != nil {
			t.Errorf("Exec %d: %v", i, err)
		}
	}

	// Verify call count
	if len(mock.ExecCalls) != 3 {
		t.Errorf("ExecCalls: got %d, want 3", len(mock.ExecCalls))
	}
	if execCount != 3 {
		t.Errorf("execCount: got %d, want 3", execCount)
	}
}

// TestE2E_SandboxWithUser tests sandbox creation with a dedicated user.
func TestE2E_SandboxWithUser(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	ctx, cancel := env.Context()
	defer cancel()

	// Create sandbox with a user
	sb, err := env.SandboxProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "e2e-user-test",
		User: "testuser",
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer env.SandboxProvider.Delete(ctx, sb.ID, true)

	// Verify user was stored
	if sb.User != "testuser" {
		t.Errorf("User: got %q, want %q", sb.User, "testuser")
	}

	// Execute whoami as the user
	result, err := env.SandboxProvider.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "whoami",
		User:    "testuser",
	})
	if err != nil {
		t.Fatalf("Exec whoami: %v", err)
	}
	// NoneProvider runs on host so whoami returns actual user
	// This test validates the flow works, not the actual user isolation
	if result.ExitCode != 0 {
		t.Errorf("whoami exit code: %d, stderr: %s", result.ExitCode, result.Stderr)
	}
}

// TestE2E_SandboxWithPersistentHome tests sandbox with persistent home mount.
func TestE2E_SandboxWithPersistentHome(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	ctx, cancel := env.Context()
	defer cancel()

	// Create a persistent home directory
	homeDir := filepath.Join(env.DataDir, "agent-homes", "test-agent")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Create home dir: %v", err)
	}

	// Write a test file to the home
	testFile := filepath.Join(homeDir, "config.txt")
	if err := os.WriteFile(testFile, []byte("test-config"), 0644); err != nil {
		t.Fatalf("Write test file: %v", err)
	}

	// Create sandbox with home mount
	sb, err := env.SandboxProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "e2e-home-test",
		User: "ayo",
		Mounts: []providers.Mount{{
			Source:      homeDir,
			Destination: "/home/ayo",
			Mode:        providers.MountModeBind,
			ReadOnly:    false,
		}},
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}
	defer env.SandboxProvider.Delete(ctx, sb.ID, true)

	// Verify mount was recorded
	foundMount := false
	for _, m := range sb.Mounts {
		if m.Destination == "/home/ayo" {
			foundMount = true
			break
		}
	}
	if !foundMount {
		t.Errorf("Expected mount at /home/ayo not found")
	}

	// NoneProvider uses host filesystem directly, so test file should exist
	if _, err := os.Stat(testFile); err != nil {
		t.Errorf("Persistent home file not accessible: %v", err)
	}
}

// TestE2E_FileTransferMock tests file transfer via mock provider.
func TestE2E_FileTransferMock(t *testing.T) {
	mock := sandbox.NewMockProvider()

	// Track stdin data passed to exec
	var receivedStdin []byte
	mock.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		receivedStdin = opts.Stdin
		// Simulate tar extraction success
		return providers.ExecResult{ExitCode: 0}, nil
	}

	ctx := context.Background()

	// Create sandbox
	sb, _ := mock.Create(ctx, providers.SandboxCreateOptions{Name: "transfer-test"})

	// Simulate push with stdin (tar data)
	tarData := []byte("fake-tar-archive-content")
	result, err := mock.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "tar",
		Args:    []string{"-xf", "-", "-C", "/tmp"},
		Stdin:   tarData,
	})
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Exit code: %d", result.ExitCode)
	}
	if string(receivedStdin) != string(tarData) {
		t.Errorf("Stdin not passed correctly: got %d bytes, want %d", len(receivedStdin), len(tarData))
	}

	// Verify exec was called
	if len(mock.ExecCalls) != 1 {
		t.Errorf("ExecCalls: got %d, want 1", len(mock.ExecCalls))
	}
	call := mock.ExecCalls[0]
	if call.Options.Command != "tar" {
		t.Errorf("Command: got %q, want %q", call.Options.Command, "tar")
	}
}

// TestE2E_SandboxStopForce tests force stopping a sandbox.
func TestE2E_SandboxStopForce(t *testing.T) {
	env := NewTestEnv(t)
	defer env.Cleanup()

	ctx, cancel := env.Context()
	defer cancel()

	sb, err := env.SandboxProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "e2e-stop-test",
	})
	if err != nil {
		t.Fatalf("Create sandbox: %v", err)
	}

	// Stop with force
	err = env.SandboxProvider.Stop(ctx, sb.ID, providers.SandboxStopOptions{
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// Verify status
	status, err := env.SandboxProvider.Status(ctx, sb.ID)
	if err != nil {
		// NoneProvider may return error for stopped sandbox
		t.Logf("Status after stop: %v", err)
	} else if status != providers.SandboxStatusStopped {
		t.Errorf("Status: got %v, want stopped", status)
	}

	// Cleanup
	if err := env.SandboxProvider.Delete(ctx, sb.ID, true); err != nil {
		t.Logf("Delete: %v", err)
	}
}

// TestE2E_MultiAgentCollaboration tests multiple agents sharing a sandbox via collaboration group.
func TestE2E_MultiAgentCollaboration(t *testing.T) {
	provider := sandbox.NewNoneProvider()
	pool := sandbox.NewPool(sandbox.PoolConfig{
		Name:    "collab-pool",
		MinSize: 1,
		MaxSize: 5,
	}, provider)

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		t.Fatalf("Start pool: %v", err)
	}
	defer pool.Stop(ctx)

	// Agent 1 acquires sandbox with collaboration group
	sb1, err := pool.AcquireWithOptions(ctx, sandbox.AcquireOptions{
		Agent: "@ayo",
		Group: "project-x",
	})
	if err != nil {
		t.Fatalf("Acquire @ayo: %v", err)
	}

	// Agent 2 joins the same collaboration group - should get same sandbox
	sb2, err := pool.AcquireWithOptions(ctx, sandbox.AcquireOptions{
		Agent: "@crush",
		Group: "project-x",
	})
	if err != nil {
		t.Fatalf("Acquire @crush: %v", err)
	}

	if sb1.ID != sb2.ID {
		t.Errorf("Agents in same group should share sandbox: sb1=%s, sb2=%s", sb1.ID, sb2.ID)
	}

	// Agent 3 in different group should get different sandbox
	sb3, err := pool.AcquireWithOptions(ctx, sandbox.AcquireOptions{
		Agent: "@research",
		Group: "project-y",
	})
	if err != nil {
		t.Fatalf("Acquire @research: %v", err)
	}

	if sb1.ID == sb3.ID {
		t.Errorf("Agents in different groups should have different sandboxes")
	}

	// Verify agents list
	agents := pool.GetSandboxAgents(sb1.ID)
	if len(agents) != 2 {
		t.Errorf("Expected 2 agents in sandbox, got %d", len(agents))
	}

	// Release one agent - sandbox should still be in use
	if err := pool.ReleaseAgent(ctx, sb1.ID, "@ayo"); err != nil {
		t.Fatalf("ReleaseAgent: %v", err)
	}

	agents = pool.GetSandboxAgents(sb1.ID)
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent after release, got %d", len(agents))
	}

	// Release remaining agent - sandbox should be idle
	if err := pool.ReleaseAgent(ctx, sb1.ID, "@crush"); err != nil {
		t.Fatalf("ReleaseAgent: %v", err)
	}

	status := pool.Status()
	if status.InUse != 1 { // Only sb3 should be in use
		t.Errorf("InUse: got %d, want 1", status.InUse)
	}
}

// TestE2E_JoinExistingSandbox tests an agent joining an existing sandbox by ID.
func TestE2E_JoinExistingSandbox(t *testing.T) {
	provider := sandbox.NewNoneProvider()
	pool := sandbox.NewPool(sandbox.PoolConfig{
		Name:    "join-pool",
		MinSize: 1,
		MaxSize: 5,
	}, provider)

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		t.Fatalf("Start pool: %v", err)
	}
	defer pool.Stop(ctx)

	// Primary agent acquires sandbox
	sb1, err := pool.Acquire(ctx, "@ayo")
	if err != nil {
		t.Fatalf("Acquire @ayo: %v", err)
	}

	// Secondary agent joins by sandbox ID
	sb2, err := pool.AcquireWithOptions(ctx, sandbox.AcquireOptions{
		Agent:       "@crush",
		JoinSandbox: sb1.ID,
	})
	if err != nil {
		t.Fatalf("Join sandbox: %v", err)
	}

	if sb1.ID != sb2.ID {
		t.Errorf("JoinSandbox should return same sandbox: sb1=%s, sb2=%s", sb1.ID, sb2.ID)
	}

	// Both agents should be listed
	agents := pool.GetSandboxAgents(sb1.ID)
	if len(agents) != 2 {
		t.Errorf("Expected 2 agents in sandbox, got %d: %v", len(agents), agents)
	}
}
