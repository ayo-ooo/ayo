// Package squads provides squad management for agent team coordination.
package squads

import (
	"time"
)

// DispatchStatus represents the status of a dispatch to a squad.
type DispatchStatus string

const (
	DispatchStatusPending   DispatchStatus = "pending"
	DispatchStatusCompleted DispatchStatus = "completed"
	DispatchStatusFailed    DispatchStatus = "failed"
)

// DispatchRecord tracks a dispatch operation for @ayo's planner.
// This allows @ayo to answer "what am I waiting on?" and track
// delegated work to squads.
type DispatchRecord struct {
	// ID is a unique identifier for this dispatch.
	ID string `json:"id"`

	// SquadName is the target squad.
	SquadName string `json:"squad_name"`

	// Input contains the dispatch input that was sent.
	Input DispatchInput `json:"input"`

	// RoutedTo is the agent that received the input within the squad.
	RoutedTo string `json:"routed_to"`

	// Status is the current status of the dispatch.
	Status DispatchStatus `json:"status"`

	// SquadTicketID is a reference to the ticket created in the squad's planner.
	// This links @ayo's tracking to the squad's internal coordination.
	SquadTicketID string `json:"squad_ticket_id,omitempty"`

	// CreatedAt is when the dispatch was initiated.
	CreatedAt time.Time `json:"created_at"`

	// CompletedAt is when the dispatch completed (or failed).
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Result contains the dispatch result when completed.
	Result *DispatchResult `json:"result,omitempty"`

	// Error contains error details when status is "failed".
	Error string `json:"error,omitempty"`
}

// DispatchTracker tracks dispatch operations for @ayo's planner.
// This interface allows different implementations (in-memory, file-based, etc.)
type DispatchTracker interface {
	// Track records a new dispatch operation.
	Track(record DispatchRecord) error

	// Complete marks a dispatch as completed with result.
	Complete(id string, result *DispatchResult) error

	// Fail marks a dispatch as failed with error.
	Fail(id string, errMsg string) error

	// Get retrieves a dispatch record by ID.
	Get(id string) (*DispatchRecord, error)

	// ListPending returns all pending dispatches.
	ListPending() ([]DispatchRecord, error)

	// List returns all dispatches, optionally filtered by status.
	List(status *DispatchStatus) ([]DispatchRecord, error)
}

// InMemoryDispatchTracker is a simple in-memory implementation of DispatchTracker.
// Used for testing and as a fallback when no persistent tracker is configured.
type InMemoryDispatchTracker struct {
	records map[string]DispatchRecord
}

// NewInMemoryDispatchTracker creates a new in-memory dispatch tracker.
func NewInMemoryDispatchTracker() *InMemoryDispatchTracker {
	return &InMemoryDispatchTracker{
		records: make(map[string]DispatchRecord),
	}
}

// Track records a new dispatch operation.
func (t *InMemoryDispatchTracker) Track(record DispatchRecord) error {
	t.records[record.ID] = record
	return nil
}

// Complete marks a dispatch as completed with result.
func (t *InMemoryDispatchTracker) Complete(id string, result *DispatchResult) error {
	record, ok := t.records[id]
	if !ok {
		return nil // Ignore unknown IDs
	}
	now := time.Now()
	record.Status = DispatchStatusCompleted
	record.CompletedAt = &now
	record.Result = result
	t.records[id] = record
	return nil
}

// Fail marks a dispatch as failed with error.
func (t *InMemoryDispatchTracker) Fail(id string, errMsg string) error {
	record, ok := t.records[id]
	if !ok {
		return nil // Ignore unknown IDs
	}
	now := time.Now()
	record.Status = DispatchStatusFailed
	record.CompletedAt = &now
	record.Error = errMsg
	t.records[id] = record
	return nil
}

// Get retrieves a dispatch record by ID.
func (t *InMemoryDispatchTracker) Get(id string) (*DispatchRecord, error) {
	record, ok := t.records[id]
	if !ok {
		return nil, nil
	}
	return &record, nil
}

// ListPending returns all pending dispatches.
func (t *InMemoryDispatchTracker) ListPending() ([]DispatchRecord, error) {
	status := DispatchStatusPending
	return t.List(&status)
}

// List returns all dispatches, optionally filtered by status.
func (t *InMemoryDispatchTracker) List(status *DispatchStatus) ([]DispatchRecord, error) {
	var result []DispatchRecord
	for _, record := range t.records {
		if status == nil || record.Status == *status {
			result = append(result, record)
		}
	}
	return result, nil
}
