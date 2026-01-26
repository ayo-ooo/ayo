package flows

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/db"
)

func setupTestDB(t *testing.T) (*sql.DB, *db.Queries, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	sqlDB, queries, err := db.ConnectWithQueries(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	cleanup := func() {
		sqlDB.Close()
		os.Remove(dbPath)
	}

	return sqlDB, queries, cleanup
}

func TestHistoryService_RecordAndComplete(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := NewHistoryService(queries)

	// Create a test flow
	flow := &Flow{
		Name:   "test-flow",
		Path:   "/path/to/flow.sh",
		Dir:    "/path/to",
		Source: FlowSourceUser,
	}

	// Record start
	runID, err := svc.RecordStart(ctx, flow, `{"input": "test"}`, true, "", "")
	if err != nil {
		t.Fatalf("RecordStart: %v", err)
	}

	if runID == "" {
		t.Fatal("RunID should not be empty")
	}

	// Verify running state
	run, err := svc.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}

	if run.Status != RunStatusRunning {
		t.Errorf("Status = %v, want %v", run.Status, RunStatusRunning)
	}

	if run.FlowName != "test-flow" {
		t.Errorf("FlowName = %v, want test-flow", run.FlowName)
	}

	if run.InputJSON != `{"input": "test"}` {
		t.Errorf("InputJSON = %v, want %v", run.InputJSON, `{"input": "test"}`)
	}

	// Record completion (add small delay to ensure measurable duration)
	time.Sleep(10 * time.Millisecond)
	startedAt := run.StartedAt
	result := CompleteResult{
		Status:          RunStatusSuccess,
		ExitCode:        0,
		OutputJSON:      `{"result": "done"}`,
		StderrLog:       "some logs",
		OutputValidated: true,
	}

	completedRun, err := svc.RecordComplete(ctx, runID, result, startedAt)
	if err != nil {
		t.Fatalf("RecordComplete: %v", err)
	}

	if completedRun.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", completedRun.Status, RunStatusSuccess)
	}

	if *completedRun.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", *completedRun.ExitCode)
	}

	if completedRun.OutputJSON != `{"result": "done"}` {
		t.Errorf("OutputJSON = %v, want %v", completedRun.OutputJSON, `{"result": "done"}`)
	}

	if completedRun.DurationMs <= 0 {
		t.Error("DurationMs should be positive")
	}
}

func TestHistoryService_ListRuns(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := NewHistoryService(queries)

	// Create multiple test flows
	flow1 := &Flow{Name: "flow-a", Path: "/path/flow-a.sh", Dir: "/path", Source: FlowSourceUser}
	flow2 := &Flow{Name: "flow-b", Path: "/path/flow-b.sh", Dir: "/path", Source: FlowSourceProject}

	// Create runs
	for i := 0; i < 3; i++ {
		runID, _ := svc.RecordStart(ctx, flow1, "{}", false, "", "")
		status := RunStatusSuccess
		if i == 1 {
			status = RunStatusFailed
		}
		svc.RecordComplete(ctx, runID, CompleteResult{Status: status, ExitCode: 0}, time.Now())
	}

	for i := 0; i < 2; i++ {
		runID, _ := svc.RecordStart(ctx, flow2, "{}", false, "", "")
		svc.RecordComplete(ctx, runID, CompleteResult{Status: RunStatusSuccess, ExitCode: 0}, time.Now())
	}

	// Test list all
	runs, err := svc.ListRuns(ctx, RunFilter{Limit: 100})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 5 {
		t.Errorf("ListRuns count = %d, want 5", len(runs))
	}

	// Test filter by flow name
	runs, err = svc.ListRuns(ctx, RunFilter{FlowName: "flow-a", Limit: 100})
	if err != nil {
		t.Fatalf("ListRuns with flow filter: %v", err)
	}
	if len(runs) != 3 {
		t.Errorf("ListRuns by flow count = %d, want 3", len(runs))
	}

	// Test filter by status
	runs, err = svc.ListRuns(ctx, RunFilter{Status: RunStatusFailed, Limit: 100})
	if err != nil {
		t.Fatalf("ListRuns with status filter: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("ListRuns by status count = %d, want 1", len(runs))
	}

	// Test limit
	runs, err = svc.ListRuns(ctx, RunFilter{Limit: 2})
	if err != nil {
		t.Fatalf("ListRuns with limit: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("ListRuns with limit count = %d, want 2", len(runs))
	}
}

func TestHistoryService_GetLastRun(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := NewHistoryService(queries)

	flow := &Flow{Name: "last-flow", Path: "/path/flow.sh", Dir: "/path", Source: FlowSourceUser}

	// Create runs with small delays
	var lastRunID string
	for i := 0; i < 3; i++ {
		runID, _ := svc.RecordStart(ctx, flow, "{}", false, "", "")
		svc.RecordComplete(ctx, runID, CompleteResult{Status: RunStatusSuccess, ExitCode: 0}, time.Now())
		lastRunID = runID
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get last run
	run, err := svc.GetLastRun(ctx, "last-flow")
	if err != nil {
		t.Fatalf("GetLastRun: %v", err)
	}

	if run.ID != lastRunID {
		t.Errorf("GetLastRun ID = %v, want %v", run.ID, lastRunID)
	}
}

func TestHistoryService_GetRunByPrefix(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := NewHistoryService(queries)

	flow := &Flow{Name: "prefix-flow", Path: "/path/flow.sh", Dir: "/path", Source: FlowSourceUser}
	runID, _ := svc.RecordStart(ctx, flow, "{}", false, "", "")
	svc.RecordComplete(ctx, runID, CompleteResult{Status: RunStatusSuccess, ExitCode: 0}, time.Now())

	// Get by exact ID
	run, err := svc.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("GetRun exact: %v", err)
	}
	if run.ID != runID {
		t.Errorf("GetRun ID = %v, want %v", run.ID, runID)
	}

	// Get by prefix (first 8 chars)
	prefix := runID[:8]
	run, err = svc.GetRun(ctx, prefix)
	if err != nil {
		t.Fatalf("GetRun prefix: %v", err)
	}
	if run.ID != runID {
		t.Errorf("GetRun by prefix ID = %v, want %v", run.ID, runID)
	}
}

func TestHistoryService_DeleteRun(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := NewHistoryService(queries)

	flow := &Flow{Name: "delete-flow", Path: "/path/flow.sh", Dir: "/path", Source: FlowSourceUser}
	runID, _ := svc.RecordStart(ctx, flow, "{}", false, "", "")
	svc.RecordComplete(ctx, runID, CompleteResult{Status: RunStatusSuccess, ExitCode: 0}, time.Now())

	// Verify it exists
	_, err := svc.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("GetRun before delete: %v", err)
	}

	// Delete
	if err := svc.DeleteRun(ctx, runID); err != nil {
		t.Fatalf("DeleteRun: %v", err)
	}

	// Verify it's gone
	_, err = svc.GetRun(ctx, runID)
	if err == nil {
		t.Error("GetRun after delete should return error")
	}
}

func TestHistoryService_PruneByAge(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := NewHistoryService(queries)

	flow := &Flow{Name: "prune-flow", Path: "/path/flow.sh", Dir: "/path", Source: FlowSourceUser}

	// Create runs
	for i := 0; i < 5; i++ {
		runID, _ := svc.RecordStart(ctx, flow, "{}", false, "", "")
		svc.RecordComplete(ctx, runID, CompleteResult{Status: RunStatusSuccess, ExitCode: 0}, time.Now())
	}

	// Prune with 0 days (should keep all since they're new)
	if err := svc.PruneByAge(ctx, 24*time.Hour); err != nil {
		t.Fatalf("PruneByAge: %v", err)
	}

	runs, _ := svc.ListRuns(ctx, RunFilter{Limit: 100})
	if len(runs) != 5 {
		t.Errorf("After prune count = %d, want 5", len(runs))
	}
}

func TestHistoryService_PruneByCount(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := NewHistoryService(queries)

	flow := &Flow{Name: "prune-count-flow", Path: "/path/flow.sh", Dir: "/path", Source: FlowSourceUser}

	// Create runs
	for i := 0; i < 10; i++ {
		runID, _ := svc.RecordStart(ctx, flow, "{}", false, "", "")
		svc.RecordComplete(ctx, runID, CompleteResult{Status: RunStatusSuccess, ExitCode: 0}, time.Now())
	}

	// Prune to keep only 3
	if err := svc.PruneByCount(ctx, 3); err != nil {
		t.Fatalf("PruneByCount: %v", err)
	}

	runs, _ := svc.ListRuns(ctx, RunFilter{Limit: 100})
	if len(runs) != 3 {
		t.Errorf("After prune count = %d, want 3", len(runs))
	}
}

func TestHistoryService_CountRuns(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := NewHistoryService(queries)

	flow := &Flow{Name: "count-flow", Path: "/path/flow.sh", Dir: "/path", Source: FlowSourceUser}

	// Create runs
	for i := 0; i < 5; i++ {
		runID, _ := svc.RecordStart(ctx, flow, "{}", false, "", "")
		status := RunStatusSuccess
		if i%2 == 0 {
			status = RunStatusFailed
		}
		svc.RecordComplete(ctx, runID, CompleteResult{Status: status, ExitCode: 0}, time.Now())
	}

	// Total count
	count, err := svc.CountRuns(ctx)
	if err != nil {
		t.Fatalf("CountRuns: %v", err)
	}
	if count != 5 {
		t.Errorf("CountRuns = %d, want 5", count)
	}

	// Count by name
	count, err = svc.CountRunsByName(ctx, "count-flow")
	if err != nil {
		t.Fatalf("CountRunsByName: %v", err)
	}
	if count != 5 {
		t.Errorf("CountRunsByName = %d, want 5", count)
	}

	// Count by status
	count, err = svc.CountRunsByStatus(ctx, RunStatusFailed)
	if err != nil {
		t.Fatalf("CountRunsByStatus: %v", err)
	}
	if count != 3 {
		t.Errorf("CountRunsByStatus(failed) = %d, want 3", count)
	}
}

func TestHistoryService_ParentAndSession(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := NewHistoryService(queries)

	flow := &Flow{Name: "linked-flow", Path: "/path/flow.sh", Dir: "/path", Source: FlowSourceUser}

	// Create parent run
	parentID, err := svc.RecordStart(ctx, flow, "{}", false, "", "")
	if err != nil {
		t.Fatalf("RecordStart parent: %v", err)
	}
	svc.RecordComplete(ctx, parentID, CompleteResult{Status: RunStatusSuccess, ExitCode: 0}, time.Now())

	// Create child run with parent ID (no session since it requires a real session in the DB)
	childID, err := svc.RecordStart(ctx, flow, "{}", false, parentID, "")
	if err != nil {
		t.Fatalf("RecordStart child: %v", err)
	}
	svc.RecordComplete(ctx, childID, CompleteResult{Status: RunStatusSuccess, ExitCode: 0}, time.Now())

	// Verify child has parent
	child, err := svc.GetRun(ctx, childID)
	if err != nil {
		t.Fatalf("GetRun child: %v", err)
	}
	if child.ParentRunID != parentID {
		t.Errorf("ParentRunID = %v, want %v", child.ParentRunID, parentID)
	}
}
