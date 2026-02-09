package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestDB creates an in-memory database for testing with all migrations run.
func setupTestDB(t *testing.T) (*sql.DB, *Queries) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "ayo-db-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	dbPath := filepath.Join(tmpDir, "test.db")
	ctx := context.Background()

	db, queries, err := ConnectWithQueries(ctx, dbPath)
	if err != nil {
		t.Fatalf("ConnectWithQueries failed: %v", err)
	}
	t.Cleanup(func() {
		queries.Close()
		db.Close()
	})

	return db, queries
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// AyoAgent Tests

func TestAyoAgents_CreateAndGet(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create an agent
	err := queries.CreateAyoAgent(ctx, CreateAyoAgentParams{
		AgentID:           "agent-1",
		AgentHandle:       "@test-agent",
		CreatedBy:         "@ayo",
		CreationReason:    nullString("User requested a helper"),
		OriginalPrompt:    "You are a helpful assistant.",
		CurrentPromptHash: nullString("hash123"),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		t.Fatalf("CreateAyoAgent failed: %v", err)
	}

	// Get by ID
	agent, err := queries.GetAyoAgent(ctx, "agent-1")
	if err != nil {
		t.Fatalf("GetAyoAgent failed: %v", err)
	}
	if agent.AgentHandle != "@test-agent" {
		t.Errorf("expected handle @test-agent, got %s", agent.AgentHandle)
	}
	if agent.CreatedBy != "@ayo" {
		t.Errorf("expected created_by @ayo, got %s", agent.CreatedBy)
	}
	if agent.IsArchived != 0 {
		t.Error("expected agent to not be archived")
	}
	if agent.PromotedTo.Valid {
		t.Error("expected agent to not be promoted")
	}

	// Get by handle
	agent2, err := queries.GetAyoAgentByHandle(ctx, "@test-agent")
	if err != nil {
		t.Fatalf("GetAyoAgentByHandle failed: %v", err)
	}
	if agent2.AgentID != "agent-1" {
		t.Errorf("expected agent-1, got %s", agent2.AgentID)
	}
}

func TestAyoAgents_NotFound(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()

	_, err := queries.GetAyoAgent(ctx, "nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestAyoAgents_ListAndArchive(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create multiple agents
	for i, name := range []string{"@agent-a", "@agent-b", "@agent-c"} {
		err := queries.CreateAyoAgent(ctx, CreateAyoAgentParams{
			AgentID:           name,
			AgentHandle:       name,
			CreatedBy:         "@ayo",
			OriginalPrompt:    "Prompt " + name,
			CurrentPromptHash: nullString("hash"),
			CreatedAt:         now + int64(i),
			UpdatedAt:         now + int64(i),
		})
		if err != nil {
			t.Fatalf("CreateAyoAgent %s failed: %v", name, err)
		}
	}

	// List active agents
	agents, err := queries.ListAyoAgents(ctx)
	if err != nil {
		t.Fatalf("ListAyoAgents failed: %v", err)
	}
	if len(agents) != 3 {
		t.Errorf("expected 3 agents, got %d", len(agents))
	}

	// Archive one agent
	err = queries.ArchiveAyoAgent(ctx, ArchiveAyoAgentParams{
		AgentID:   "@agent-b",
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("ArchiveAyoAgent failed: %v", err)
	}

	// List active agents again
	agents, err = queries.ListAyoAgents(ctx)
	if err != nil {
		t.Fatalf("ListAyoAgents failed: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("expected 2 active agents, got %d", len(agents))
	}

	// List archived agents
	archived, err := queries.ListArchivedAyoAgents(ctx)
	if err != nil {
		t.Fatalf("ListArchivedAyoAgents failed: %v", err)
	}
	if len(archived) != 1 {
		t.Errorf("expected 1 archived agent, got %d", len(archived))
	}
	if archived[0].AgentHandle != "@agent-b" {
		t.Errorf("expected @agent-b to be archived, got %s", archived[0].AgentHandle)
	}
}

func TestAyoAgents_PromoteAndUnarchive(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create an agent
	err := queries.CreateAyoAgent(ctx, CreateAyoAgentParams{
		AgentID:           "agent-1",
		AgentHandle:       "@test",
		CreatedBy:         "@ayo",
		OriginalPrompt:    "Original",
		CurrentPromptHash: nullString("hash"),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		t.Fatalf("CreateAyoAgent failed: %v", err)
	}

	// Promote the agent
	err = queries.PromoteAyoAgent(ctx, PromoteAyoAgentParams{
		AgentID:    "agent-1",
		PromotedTo: nullString("user-agents/test"),
		UpdatedAt:  time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("PromoteAyoAgent failed: %v", err)
	}

	// Verify promoted
	agent, err := queries.GetAyoAgent(ctx, "agent-1")
	if err != nil {
		t.Fatalf("GetAyoAgent failed: %v", err)
	}
	if !agent.PromotedTo.Valid || agent.PromotedTo.String != "user-agents/test" {
		t.Errorf("expected promoted_to 'user-agents/test', got %v", agent.PromotedTo)
	}

	// Archive and unarchive
	err = queries.ArchiveAyoAgent(ctx, ArchiveAyoAgentParams{
		AgentID:   "agent-1",
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("ArchiveAyoAgent failed: %v", err)
	}

	err = queries.UnarchiveAyoAgent(ctx, UnarchiveAyoAgentParams{
		AgentID:   "agent-1",
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("UnarchiveAyoAgent failed: %v", err)
	}

	agent, err = queries.GetAyoAgent(ctx, "agent-1")
	if err != nil {
		t.Fatalf("GetAyoAgent failed: %v", err)
	}
	if agent.IsArchived != 0 {
		t.Error("expected agent to not be archived after unarchive")
	}
}

func TestAyoAgents_Refinements(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create agent first
	err := queries.CreateAyoAgent(ctx, CreateAyoAgentParams{
		AgentID:           "agent-1",
		AgentHandle:       "@test",
		CreatedBy:         "@ayo",
		OriginalPrompt:    "Original prompt",
		CurrentPromptHash: nullString("hash1"),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		t.Fatalf("CreateAyoAgent failed: %v", err)
	}

	// Create refinements
	for i, reason := range []string{"Better tone", "More detail", "Fixed bugs"} {
		err := queries.CreateAgentRefinement(ctx, CreateAgentRefinementParams{
			ID:             "ref-" + string(rune('a'+i)),
			AgentID:        "agent-1",
			PreviousPrompt: "Previous " + string(rune('a'+i)),
			NewPrompt:      "New " + string(rune('a'+i)),
			Reason:         reason,
			CreatedAt:      now + int64(i),
		})
		if err != nil {
			t.Fatalf("CreateAgentRefinement failed: %v", err)
		}
	}

	// List refinements
	refinements, err := queries.ListAgentRefinements(ctx, "agent-1")
	if err != nil {
		t.Fatalf("ListAgentRefinements failed: %v", err)
	}
	if len(refinements) != 3 {
		t.Errorf("expected 3 refinements, got %d", len(refinements))
	}

	// Get latest refinement
	latest, err := queries.GetLatestRefinement(ctx, "agent-1")
	if err != nil {
		t.Fatalf("GetLatestRefinement failed: %v", err)
	}
	if latest.Reason != "Fixed bugs" {
		t.Errorf("expected 'Fixed bugs', got %s", latest.Reason)
	}
}

func TestAyoAgents_Delete(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create agent
	err := queries.CreateAyoAgent(ctx, CreateAyoAgentParams{
		AgentID:           "agent-1",
		AgentHandle:       "@test",
		CreatedBy:         "@ayo",
		OriginalPrompt:    "Prompt",
		CurrentPromptHash: nullString("hash"),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		t.Fatalf("CreateAyoAgent failed: %v", err)
	}

	// Delete
	err = queries.DeleteAyoAgent(ctx, "agent-1")
	if err != nil {
		t.Fatalf("DeleteAyoAgent failed: %v", err)
	}

	// Verify deleted
	_, err = queries.GetAyoAgent(ctx, "agent-1")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

// FlowRun Tests

func TestFlowRuns_CreateAndGet(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create a flow run
	run, err := queries.CreateFlowRun(ctx, CreateFlowRunParams{
		ID:         "run-1",
		FlowName:   "daily-digest",
		FlowPath:   "/flows/daily-digest.yaml",
		FlowSource: "user",
		InputJson:  nullString(`{"key": "value"}`),
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("CreateFlowRun failed: %v", err)
	}
	if run.ID != "run-1" {
		t.Errorf("expected run-1, got %s", run.ID)
	}
	if run.Status != "running" {
		t.Errorf("expected running, got %s", run.Status)
	}

	// Get the run
	retrieved, err := queries.GetFlowRun(ctx, "run-1")
	if err != nil {
		t.Fatalf("GetFlowRun failed: %v", err)
	}
	if retrieved.FlowName != "daily-digest" {
		t.Errorf("expected daily-digest, got %s", retrieved.FlowName)
	}
}

func TestFlowRuns_Complete(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create a flow run
	_, err := queries.CreateFlowRun(ctx, CreateFlowRunParams{
		ID:         "run-1",
		FlowName:   "build",
		FlowPath:   "/flows/build.yaml",
		FlowSource: "project",
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("CreateFlowRun failed: %v", err)
	}

	// Complete the run
	finished := now + 5
	duration := int64(5000) // 5 seconds in ms
	completed, err := queries.CompleteFlowRun(ctx, CompleteFlowRunParams{
		ID:         "run-1",
		Status:     "success",
		ExitCode:   sql.NullInt64{Int64: 0, Valid: true},
		OutputJson: nullString(`{"result": "ok"}`),
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		DurationMs: sql.NullInt64{Int64: duration, Valid: true},
	})
	if err != nil {
		t.Fatalf("CompleteFlowRun failed: %v", err)
	}
	if completed.Status != "success" {
		t.Errorf("expected success, got %s", completed.Status)
	}
	if !completed.ExitCode.Valid || completed.ExitCode.Int64 != 0 {
		t.Errorf("expected exit code 0, got %v", completed.ExitCode)
	}
}

func TestFlowRuns_ListByName(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create runs for different flows
	flows := []string{"build", "test", "build", "deploy", "build"}
	for i, name := range flows {
		_, err := queries.CreateFlowRun(ctx, CreateFlowRunParams{
			ID:         "run-" + string(rune('0'+i)),
			FlowName:   name,
			FlowPath:   "/flows/" + name + ".yaml",
			FlowSource: "user",
			StartedAt:  now + int64(i),
		})
		if err != nil {
			t.Fatalf("CreateFlowRun %d failed: %v", i, err)
		}
	}

	// List all
	all, err := queries.ListFlowRuns(ctx, 100)
	if err != nil {
		t.Fatalf("ListFlowRuns failed: %v", err)
	}
	if len(all) != 5 {
		t.Errorf("expected 5 runs, got %d", len(all))
	}

	// List by name
	buildRuns, err := queries.ListFlowRunsByName(ctx, ListFlowRunsByNameParams{
		FlowName: "build",
		Limit:    100,
	})
	if err != nil {
		t.Fatalf("ListFlowRunsByName failed: %v", err)
	}
	if len(buildRuns) != 3 {
		t.Errorf("expected 3 build runs, got %d", len(buildRuns))
	}
}

func TestFlowRuns_ListByStatus(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create runs and complete them with different statuses
	for i, status := range []string{"success", "failed", "success", "success"} {
		run, err := queries.CreateFlowRun(ctx, CreateFlowRunParams{
			ID:         "run-" + string(rune('0'+i)),
			FlowName:   "test-flow",
			FlowPath:   "/flows/test.yaml",
			FlowSource: "user",
			StartedAt:  now + int64(i),
		})
		if err != nil {
			t.Fatalf("CreateFlowRun %d failed: %v", i, err)
		}

		_, err = queries.CompleteFlowRun(ctx, CompleteFlowRunParams{
			ID:         run.ID,
			Status:     status,
			FinishedAt: sql.NullInt64{Int64: now + int64(i) + 1, Valid: true},
		})
		if err != nil {
			t.Fatalf("CompleteFlowRun %d failed: %v", i, err)
		}
	}

	// Count by status
	successCount, err := queries.CountFlowRunsByStatus(ctx, "success")
	if err != nil {
		t.Fatalf("CountFlowRunsByStatus failed: %v", err)
	}
	if successCount != 3 {
		t.Errorf("expected 3 success runs, got %d", successCount)
	}

	// List by status
	failedRuns, err := queries.ListFlowRunsByStatus(ctx, ListFlowRunsByStatusParams{
		Status: "failed",
		Limit:  100,
	})
	if err != nil {
		t.Fatalf("ListFlowRunsByStatus failed: %v", err)
	}
	if len(failedRuns) != 1 {
		t.Errorf("expected 1 failed run, got %d", len(failedRuns))
	}
}

func TestFlowRuns_Delete(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create a run
	_, err := queries.CreateFlowRun(ctx, CreateFlowRunParams{
		ID:         "run-1",
		FlowName:   "test",
		FlowPath:   "/flows/test.yaml",
		FlowSource: "user",
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("CreateFlowRun failed: %v", err)
	}

	// Delete it
	err = queries.DeleteFlowRun(ctx, "run-1")
	if err != nil {
		t.Fatalf("DeleteFlowRun failed: %v", err)
	}

	// Verify deleted
	_, err = queries.GetFlowRun(ctx, "run-1")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestFlowRuns_Count(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create some runs
	for i := 0; i < 5; i++ {
		_, err := queries.CreateFlowRun(ctx, CreateFlowRunParams{
			ID:         "run-" + string(rune('0'+i)),
			FlowName:   "test",
			FlowPath:   "/flows/test.yaml",
			FlowSource: "user",
			StartedAt:  now + int64(i),
		})
		if err != nil {
			t.Fatalf("CreateFlowRun %d failed: %v", i, err)
		}
	}

	// Count all
	count, err := queries.CountFlowRuns(ctx)
	if err != nil {
		t.Fatalf("CountFlowRuns failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 runs, got %d", count)
	}

	// Count by name
	byName, err := queries.CountFlowRunsByName(ctx, "test")
	if err != nil {
		t.Fatalf("CountFlowRunsByName failed: %v", err)
	}
	if byName != 5 {
		t.Errorf("expected 5 test runs, got %d", byName)
	}
}

// Capabilities Tests

func TestCapabilities_CreateAndGet(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create a capability
	err := queries.CreateCapability(ctx, CreateCapabilityParams{
		ID:          "cap-1",
		AgentID:     "agent-1",
		Name:        "code-review",
		Description: "Reviews code for bugs and style issues",
		Confidence:  0.9,
		Source:      "inferred",
		InputHash:   "hash123",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("CreateCapability failed: %v", err)
	}

	// Get capabilities by agent
	caps, err := queries.GetCapabilitiesByAgent(ctx, "agent-1")
	if err != nil {
		t.Fatalf("GetCapabilitiesByAgent failed: %v", err)
	}
	if len(caps) != 1 {
		t.Errorf("expected 1 capability, got %d", len(caps))
	}
	if caps[0].Name != "code-review" {
		t.Errorf("expected code-review, got %s", caps[0].Name)
	}
}

func TestCapabilities_Search(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create multiple capabilities
	capabilities := []struct {
		agent string
		name  string
		desc  string
	}{
		{"agent-1", "code-review", "Reviews code for issues"},
		{"agent-1", "refactoring", "Refactors code for clarity"},
		{"agent-2", "documentation", "Writes documentation"},
		{"agent-2", "testing", "Writes unit tests"},
	}

	for i, cap := range capabilities {
		err := queries.CreateCapability(ctx, CreateCapabilityParams{
			ID:          "cap-" + string(rune('0'+i)),
			AgentID:     cap.agent,
			Name:        cap.name,
			Description: cap.desc,
			Confidence:  0.8,
			Source:      "inferred",
			InputHash:   "hash",
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		if err != nil {
			t.Fatalf("CreateCapability %s failed: %v", cap.name, err)
		}
	}

	// Search by name pattern
	results, err := queries.SearchCapabilitiesByName(ctx, SearchCapabilitiesByNameParams{
		Name:        "%code%",
		Description: "%code%", // Match code in name or description
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("SearchCapabilitiesByName failed: %v", err)
	}
	// Should match "code-review" (has "code" in name) and "refactoring" (has "code" in description)
	if len(results) != 2 {
		t.Errorf("expected 2 results for %%code%%, got %d", len(results))
	}

	// Get by name
	cap, err := queries.GetCapabilityByName(ctx, GetCapabilityByNameParams{
		AgentID: "agent-2",
		Name:    "testing",
	})
	if err != nil {
		t.Fatalf("GetCapabilityByName failed: %v", err)
	}
	if cap.Description != "Writes unit tests" {
		t.Errorf("wrong description: %s", cap.Description)
	}
}

func TestCapabilities_Delete(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create capabilities for an agent
	for i, name := range []string{"cap1", "cap2", "cap3"} {
		err := queries.CreateCapability(ctx, CreateCapabilityParams{
			ID:          "cap-" + name,
			AgentID:     "agent-1",
			Name:        name,
			Description: "Description",
			Confidence:  0.8,
			Source:      "inferred",
			InputHash:   "hash",
			CreatedAt:   now + int64(i),
			UpdatedAt:   now + int64(i),
		})
		if err != nil {
			t.Fatalf("CreateCapability failed: %v", err)
		}
	}

	// Delete all for agent
	err := queries.DeleteCapabilitiesByAgent(ctx, "agent-1")
	if err != nil {
		t.Fatalf("DeleteCapabilitiesByAgent failed: %v", err)
	}

	// Verify deleted
	caps, err := queries.GetCapabilitiesByAgent(ctx, "agent-1")
	if err != nil {
		t.Fatalf("GetCapabilitiesByAgent failed: %v", err)
	}
	if len(caps) != 0 {
		t.Errorf("expected 0 capabilities, got %d", len(caps))
	}
}

func TestCapabilities_ListAll(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create capabilities across agents
	for i := 0; i < 5; i++ {
		err := queries.CreateCapability(ctx, CreateCapabilityParams{
			ID:          "cap-" + string(rune('0'+i)),
			AgentID:     "agent-" + string(rune('a'+i)),
			Name:        "cap-" + string(rune('0'+i)),
			Description: "Description " + string(rune('0'+i)),
			Confidence:  0.9,
			Source:      "inferred",
			InputHash:   "hash",
			CreatedAt:   now + int64(i),
			UpdatedAt:   now + int64(i),
		})
		if err != nil {
			t.Fatalf("CreateCapability failed: %v", err)
		}
	}

	// List all
	all, err := queries.ListAllCapabilities(ctx)
	if err != nil {
		t.Fatalf("ListAllCapabilities failed: %v", err)
	}
	if len(all) != 5 {
		t.Errorf("expected 5 capabilities, got %d", len(all))
	}
}

func TestCapabilities_GetByHash(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Create capabilities with same hash (simulates same agent config)
	for i := 0; i < 3; i++ {
		err := queries.CreateCapability(ctx, CreateCapabilityParams{
			ID:          "cap-" + string(rune('0'+i)),
			AgentID:     "agent-1",
			Name:        "cap-" + string(rune('0'+i)),
			Description: "Description",
			Confidence:  0.8,
			Source:      "inferred",
			InputHash:   "same-hash",
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		if err != nil {
			t.Fatalf("CreateCapability failed: %v", err)
		}
	}

	// Get by hash
	caps, err := queries.GetCapabilitiesByHash(ctx, "same-hash")
	if err != nil {
		t.Fatalf("GetCapabilitiesByHash failed: %v", err)
	}
	if len(caps) != 3 {
		t.Errorf("expected 3 capabilities with same hash, got %d", len(caps))
	}
}
