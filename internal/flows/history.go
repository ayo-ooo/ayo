package flows

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/alexcabrera/ayo/internal/db"
)

// FlowRun represents a flow execution record.
type FlowRun struct {
	ID              string
	FlowName        string
	FlowPath        string
	FlowSource      FlowSource
	Status          RunStatus
	ExitCode        *int
	ErrorMessage    string
	InputJSON       string
	OutputJSON      string
	StderrLog       string
	StartedAt       time.Time
	FinishedAt      *time.Time
	DurationMs      int64
	ParentRunID     string
	SessionID       string
	InputValidated  bool
	OutputValidated bool
}

// RunFilter contains optional filters for listing runs.
type RunFilter struct {
	FlowName  string
	Status    RunStatus
	SessionID string
	Limit     int64
}

// HistoryService manages flow run history.
type HistoryService struct {
	queries *db.Queries
}

// NewHistoryService creates a new history service.
func NewHistoryService(queries *db.Queries) *HistoryService {
	return &HistoryService{queries: queries}
}

// RecordStart creates a new running flow record and returns its ID.
func (h *HistoryService) RecordStart(ctx context.Context, flow *Flow, input string, inputValidated bool, parentRunID, sessionID string) (string, error) {
	id := ulid.Make().String()

	params := db.CreateFlowRunParams{
		ID:             id,
		FlowName:       flow.Name,
		FlowPath:       flow.Path,
		FlowSource:     string(flow.Source),
		InputJson:      toNullString(input),
		InputValidated: boolToInt64(inputValidated),
		StartedAt:      time.Now().UnixMilli(),
		ParentRunID:    toNullString(parentRunID),
		SessionID:      toNullString(sessionID),
	}

	_, err := h.queries.CreateFlowRun(ctx, params)
	if err != nil {
		return "", err
	}

	return id, nil
}

// CompleteResult contains the result of a completed flow run.
type CompleteResult struct {
	Status          RunStatus
	ExitCode        int
	ErrorMessage    string
	OutputJSON      string
	StderrLog       string
	OutputValidated bool
}

// RecordComplete updates a flow run with completion data.
func (h *HistoryService) RecordComplete(ctx context.Context, runID string, result CompleteResult, startedAt time.Time) (*FlowRun, error) {
	now := time.Now()
	durationMs := now.Sub(startedAt).Milliseconds()

	params := db.CompleteFlowRunParams{
		ID:              runID,
		Status:          string(result.Status),
		ExitCode:        sql.NullInt64{Int64: int64(result.ExitCode), Valid: true},
		ErrorMessage:    toNullString(result.ErrorMessage),
		OutputJson:      toNullString(result.OutputJSON),
		StderrLog:       toNullString(result.StderrLog),
		OutputValidated: boolToInt64(result.OutputValidated),
		FinishedAt:      sql.NullInt64{Int64: now.UnixMilli(), Valid: true},
		DurationMs:      sql.NullInt64{Int64: durationMs, Valid: true},
	}

	dbRun, err := h.queries.CompleteFlowRun(ctx, params)
	if err != nil {
		return nil, err
	}

	return dbFlowRunToFlowRun(dbRun), nil
}

// GetRun retrieves a flow run by ID or ID prefix.
func (h *HistoryService) GetRun(ctx context.Context, idOrPrefix string) (*FlowRun, error) {
	// Try exact match first
	dbRun, err := h.queries.GetFlowRun(ctx, idOrPrefix)
	if err == nil {
		return dbFlowRunToFlowRun(dbRun), nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Try prefix match
	runs, err := h.queries.GetFlowRunByPrefix(ctx, sql.NullString{String: idOrPrefix, Valid: true})
	if err != nil {
		return nil, err
	}

	if len(runs) == 0 {
		return nil, sql.ErrNoRows
	}

	if len(runs) > 1 {
		return nil, errors.New("ambiguous run ID prefix: multiple matches")
	}

	return dbFlowRunToFlowRun(runs[0]), nil
}

// ListRuns retrieves flow runs with optional filters.
func (h *HistoryService) ListRuns(ctx context.Context, filter RunFilter) ([]*FlowRun, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	var dbRuns []db.FlowRun
	var err error

	switch {
	case filter.FlowName != "":
		dbRuns, err = h.queries.ListFlowRunsByName(ctx, db.ListFlowRunsByNameParams{
			FlowName: filter.FlowName,
			Limit:    limit,
		})
	case filter.Status != "":
		dbRuns, err = h.queries.ListFlowRunsByStatus(ctx, db.ListFlowRunsByStatusParams{
			Status: string(filter.Status),
			Limit:  limit,
		})
	case filter.SessionID != "":
		dbRuns, err = h.queries.ListFlowRunsBySession(ctx, sql.NullString{String: filter.SessionID, Valid: true})
	default:
		dbRuns, err = h.queries.ListFlowRuns(ctx, limit)
	}

	if err != nil {
		return nil, err
	}

	runs := make([]*FlowRun, len(dbRuns))
	for i, dbRun := range dbRuns {
		runs[i] = dbFlowRunToFlowRun(dbRun)
	}

	return runs, nil
}

// GetLastRun retrieves the most recent run for a flow.
func (h *HistoryService) GetLastRun(ctx context.Context, flowName string) (*FlowRun, error) {
	dbRun, err := h.queries.GetLastFlowRun(ctx, flowName)
	if err != nil {
		return nil, err
	}

	return dbFlowRunToFlowRun(dbRun), nil
}

// DeleteRun deletes a flow run by ID.
func (h *HistoryService) DeleteRun(ctx context.Context, id string) error {
	return h.queries.DeleteFlowRun(ctx, id)
}

// PruneByAge deletes runs older than the given duration.
func (h *HistoryService) PruneByAge(ctx context.Context, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge).UnixMilli()
	return h.queries.PruneFlowRunsByAge(ctx, cutoff)
}

// PruneByCount keeps only the most recent N runs.
func (h *HistoryService) PruneByCount(ctx context.Context, keepCount int64) error {
	return h.queries.PruneFlowRunsByCount(ctx, keepCount)
}

// Prune applies both age and count retention policies.
func (h *HistoryService) Prune(ctx context.Context, maxAgeDays int, maxCount int64) error {
	if maxAgeDays > 0 {
		if err := h.PruneByAge(ctx, time.Duration(maxAgeDays)*24*time.Hour); err != nil {
			return err
		}
	}

	if maxCount > 0 {
		if err := h.PruneByCount(ctx, maxCount); err != nil {
			return err
		}
	}

	return nil
}

// CountRuns returns the total number of flow runs.
func (h *HistoryService) CountRuns(ctx context.Context) (int64, error) {
	return h.queries.CountFlowRuns(ctx)
}

// CountRunsByName returns the number of runs for a specific flow.
func (h *HistoryService) CountRunsByName(ctx context.Context, flowName string) (int64, error) {
	return h.queries.CountFlowRunsByName(ctx, flowName)
}

// CountRunsByStatus returns the number of runs with a specific status.
func (h *HistoryService) CountRunsByStatus(ctx context.Context, status RunStatus) (int64, error) {
	return h.queries.CountFlowRunsByStatus(ctx, string(status))
}

// Helper functions

func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func int64ToBool(i int64) bool {
	return i != 0
}

func dbFlowRunToFlowRun(dbRun db.FlowRun) *FlowRun {
	run := &FlowRun{
		ID:              dbRun.ID,
		FlowName:        dbRun.FlowName,
		FlowPath:        dbRun.FlowPath,
		FlowSource:      FlowSource(dbRun.FlowSource),
		Status:          RunStatus(dbRun.Status),
		ErrorMessage:    dbRun.ErrorMessage.String,
		InputJSON:       dbRun.InputJson.String,
		OutputJSON:      dbRun.OutputJson.String,
		StderrLog:       dbRun.StderrLog.String,
		StartedAt:       time.UnixMilli(dbRun.StartedAt),
		ParentRunID:     dbRun.ParentRunID.String,
		SessionID:       dbRun.SessionID.String,
		InputValidated:  int64ToBool(dbRun.InputValidated),
		OutputValidated: int64ToBool(dbRun.OutputValidated),
	}

	if dbRun.ExitCode.Valid {
		exitCode := int(dbRun.ExitCode.Int64)
		run.ExitCode = &exitCode
	}

	if dbRun.FinishedAt.Valid {
		finishedAt := time.UnixMilli(dbRun.FinishedAt.Int64)
		run.FinishedAt = &finishedAt
	}

	if dbRun.DurationMs.Valid {
		run.DurationMs = dbRun.DurationMs.Int64
	}

	return run
}
