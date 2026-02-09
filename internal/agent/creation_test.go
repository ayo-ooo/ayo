package agent

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/db"
)

// setupTestDB creates a temporary SQLite database for testing.
func setupTestDB(t *testing.T) (*sql.DB, db.Querier) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	// Create schema (matching internal/db/migrations/003_orchestration.sql)
	schema := `
	CREATE TABLE IF NOT EXISTS ayo_created_agents (
		agent_id TEXT PRIMARY KEY,
		agent_handle TEXT NOT NULL UNIQUE,
		created_by TEXT NOT NULL,
		creation_reason TEXT,
		original_prompt TEXT NOT NULL,
		current_prompt_hash TEXT,
		invocation_count INTEGER NOT NULL DEFAULT 0,
		success_count INTEGER NOT NULL DEFAULT 0,
		failure_count INTEGER NOT NULL DEFAULT 0,
		last_used_at INTEGER,
		refinement_count INTEGER NOT NULL DEFAULT 0,
		confidence REAL NOT NULL DEFAULT 0.0,
		is_archived INTEGER NOT NULL DEFAULT 0,
		promoted_to TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS agent_refinements (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		previous_prompt TEXT NOT NULL,
		new_prompt TEXT NOT NULL,
		reason TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		FOREIGN KEY (agent_id) REFERENCES ayo_created_agents(agent_id)
	);
	`
	if _, err := sqlDB.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return sqlDB, db.New(sqlDB)
}

// setupTestConfig creates a temporary config for testing.
func setupTestConfig(t *testing.T) config.Config {
	t.Helper()

	tmpDir := t.TempDir()

	// Create agents directory
	agentsDir := filepath.Join(tmpDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("create agents dir: %v", err)
	}

	return config.Config{
		AgentsDir: agentsDir,
	}
}

// TestCreateAgent_Basic tests basic agent creation.
func TestCreateAgent_Basic(t *testing.T) {
	ctx := context.Background()
	_, q := setupTestDB(t)
	cfg := setupTestConfig(t)

	result, err := CreateAgent(ctx, cfg, q, CreateOptions{
		Handle:         "test-agent",
		SystemPrompt:   "You are a helpful test assistant.",
		Description:    "Test agent",
		CreatedBy:      "@ayo",
		CreationReason: "Unit testing",
	})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	// Handle should be normalized with @ prefix
	if result.Agent.Handle != "@test-agent" {
		t.Errorf("handle = %q, want '@test-agent'", result.Agent.Handle)
	}
	if result.AgentID == "" {
		t.Error("AgentID should be set for @ayo-created agents")
	}
	if !result.IsAyoCreated {
		t.Error("IsAyoCreated should be true")
	}

	// Verify agent file was created (directory uses normalized handle with @)
	agentDir := filepath.Join(cfg.AgentsDir, "@test-agent")
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		t.Error("agent directory should exist")
	}
}

// TestCreateAgent_UserCreated tests agent creation by user (not @ayo).
func TestCreateAgent_UserCreated(t *testing.T) {
	ctx := context.Background()
	_, q := setupTestDB(t)
	cfg := setupTestConfig(t)

	result, err := CreateAgent(ctx, cfg, q, CreateOptions{
		Handle:       "user-agent",
		SystemPrompt: "You are a user-created agent.",
		CreatedBy:    "user",
	})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	if result.IsAyoCreated {
		t.Error("IsAyoCreated should be false for user-created agents")
	}
	if result.AgentID != "" {
		t.Error("AgentID should be empty for user-created agents")
	}
}

// TestCreateAgent_ReservedHandle tests that reserved handles are rejected.
func TestCreateAgent_ReservedHandle(t *testing.T) {
	ctx := context.Background()
	_, q := setupTestDB(t)
	cfg := setupTestConfig(t)

	_, err := CreateAgent(ctx, cfg, q, CreateOptions{
		Handle:       "@ayo",
		SystemPrompt: "Trying to override @ayo",
		CreatedBy:    "@ayo",
	})
	if err == nil {
		t.Error("should reject reserved handle @ayo")
	}
}

// TestCreateAgent_WithSkillsAndTools tests agent creation with skills and tools.
func TestCreateAgent_WithSkillsAndTools(t *testing.T) {
	ctx := context.Background()
	_, q := setupTestDB(t)
	cfg := setupTestConfig(t)

	result, err := CreateAgent(ctx, cfg, q, CreateOptions{
		Handle:       "configured-agent",
		SystemPrompt: "You are a configured agent.",
		Skills:       []string{"coding", "file-operations"},
		AllowedTools: []string{"bash", "edit"},
		CreatedBy:    "@ayo",
	})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	// Reload agent to verify config was saved
	loaded, err := Load(cfg, "@configured-agent")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Config.Skills) != 2 {
		t.Errorf("skills count = %d, want 2", len(loaded.Config.Skills))
	}
	if len(loaded.Config.AllowedTools) != 2 {
		t.Errorf("allowed_tools count = %d, want 2", len(loaded.Config.AllowedTools))
	}

	_ = result // silence unused warning
}

// TestRefineAgent_Append tests appending to an agent's prompt.
func TestRefineAgent_Append(t *testing.T) {
	ctx := context.Background()
	_, q := setupTestDB(t)
	cfg := setupTestConfig(t)

	// First create an agent
	_, err := CreateAgent(ctx, cfg, q, CreateOptions{
		Handle:       "refine-test",
		SystemPrompt: "Original prompt.",
		CreatedBy:    "@ayo",
	})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	// Refine with append (handle must include @)
	err = RefineAgent(ctx, cfg, q, RefinementOptions{
		AgentHandle:  "@refine-test",
		AppendPrompt: "Additional instruction.",
		Reason:       "Testing append",
		UpdateOnDisk: true,
	})
	if err != nil {
		t.Fatalf("RefineAgent: %v", err)
	}

	// Verify updated prompt
	loaded, err := Load(cfg, "@refine-test")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	expected := "Original prompt.\n\nAdditional instruction."
	if loaded.System != expected {
		t.Errorf("prompt = %q, want %q", loaded.System, expected)
	}
}

// TestRefineAgent_Replace tests replacing an agent's prompt.
func TestRefineAgent_Replace(t *testing.T) {
	ctx := context.Background()
	_, q := setupTestDB(t)
	cfg := setupTestConfig(t)

	// First create an agent
	_, err := CreateAgent(ctx, cfg, q, CreateOptions{
		Handle:       "replace-test",
		SystemPrompt: "Original prompt.",
		CreatedBy:    "@ayo",
	})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	// Refine with replacement
	err = RefineAgent(ctx, cfg, q, RefinementOptions{
		AgentHandle:  "@replace-test",
		NewPrompt:    "Completely new prompt.",
		Reason:       "Testing replacement",
		UpdateOnDisk: true,
	})
	if err != nil {
		t.Fatalf("RefineAgent: %v", err)
	}

	// Verify updated prompt
	loaded, err := Load(cfg, "@replace-test")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.System != "Completely new prompt." {
		t.Errorf("prompt = %q, want 'Completely new prompt.'", loaded.System)
	}
}

// TestArchiveAgent tests archiving an agent.
func TestArchiveAgent(t *testing.T) {
	ctx := context.Background()
	sqlDB, q := setupTestDB(t)
	cfg := setupTestConfig(t)

	// Create agent
	result, err := CreateAgent(ctx, cfg, q, CreateOptions{
		Handle:       "archive-test",
		SystemPrompt: "Test agent",
		CreatedBy:    "@ayo",
	})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	// Archive it (use @-prefixed handle)
	err = ArchiveAgent(ctx, q, "@archive-test")
	if err != nil {
		t.Fatalf("ArchiveAgent: %v", err)
	}

	// Verify archived in database
	var isArchived int
	err = sqlDB.QueryRow(`SELECT is_archived FROM ayo_created_agents WHERE agent_id = ?`, result.AgentID).Scan(&isArchived)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if isArchived != 1 {
		t.Error("agent should be archived")
	}
}

// TestUnarchiveAgent tests unarchiving an agent.
func TestUnarchiveAgent(t *testing.T) {
	ctx := context.Background()
	sqlDB, q := setupTestDB(t)
	cfg := setupTestConfig(t)

	// Create and archive agent
	result, err := CreateAgent(ctx, cfg, q, CreateOptions{
		Handle:       "unarchive-test",
		SystemPrompt: "Test agent",
		CreatedBy:    "@ayo",
	})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	err = ArchiveAgent(ctx, q, "@unarchive-test")
	if err != nil {
		t.Fatalf("ArchiveAgent: %v", err)
	}

	// Unarchive it
	err = UnarchiveAgent(ctx, q, "@unarchive-test")
	if err != nil {
		t.Fatalf("UnarchiveAgent: %v", err)
	}

	// Verify not archived
	var isArchived int
	err = sqlDB.QueryRow(`SELECT is_archived FROM ayo_created_agents WHERE agent_id = ?`, result.AgentID).Scan(&isArchived)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if isArchived != 0 {
		t.Error("agent should not be archived after unarchive")
	}
}

// TestPromoteAgent tests promoting an agent to a new handle.
func TestPromoteAgent(t *testing.T) {
	ctx := context.Background()
	sqlDB, q := setupTestDB(t)
	cfg := setupTestConfig(t)

	// Create agent
	result, err := CreateAgent(ctx, cfg, q, CreateOptions{
		Handle:       "promote-source",
		SystemPrompt: "Promotable agent",
		CreatedBy:    "@ayo",
	})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	// Promote it (use @-prefixed handles)
	err = PromoteAgent(ctx, cfg, q, "@promote-source", "promoted-agent")
	if err != nil {
		t.Fatalf("PromoteAgent: %v", err)
	}

	// Verify new agent exists on disk
	promoted, err := Load(cfg, "@promoted-agent")
	if err != nil {
		t.Fatalf("Load promoted: %v", err)
	}
	if promoted.System != "Promotable agent" {
		t.Errorf("promoted system prompt incorrect")
	}

	// Verify database records promotion
	var promotedTo string
	err = sqlDB.QueryRow(`SELECT promoted_to FROM ayo_created_agents WHERE agent_id = ?`, result.AgentID).Scan(&promotedTo)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if promotedTo != "@promoted-agent" {
		t.Errorf("promoted_to = %q, want '@promoted-agent'", promotedTo)
	}
}

// TestRefineAgent_History tests that refinement history is recorded.
func TestRefineAgent_History(t *testing.T) {
	ctx := context.Background()
	sqlDB, q := setupTestDB(t)
	cfg := setupTestConfig(t)

	// Create and refine multiple times
	result, err := CreateAgent(ctx, cfg, q, CreateOptions{
		Handle:       "history-test",
		SystemPrompt: "Version 1",
		CreatedBy:    "@ayo",
	})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	for i, version := range []string{"Version 2", "Version 3"} {
		err = RefineAgent(ctx, cfg, q, RefinementOptions{
			AgentHandle:  "@history-test",
			NewPrompt:    version,
			Reason:       "Refinement " + string(rune('1'+i)),
			UpdateOnDisk: true,
		})
		if err != nil {
			t.Fatalf("RefineAgent %d: %v", i+1, err)
		}
	}

	// Count refinement history
	var count int
	err = sqlDB.QueryRow(`SELECT COUNT(*) FROM agent_refinements WHERE agent_id = ?`, result.AgentID).Scan(&count)
	if err != nil {
		t.Fatalf("query refinements: %v", err)
	}
	if count != 2 {
		t.Errorf("refinement count = %d, want 2", count)
	}
}
