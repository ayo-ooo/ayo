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
