package squads

import (
	"testing"
	"time"
)

func TestInMemoryDispatchTracker(t *testing.T) {
	tracker := NewInMemoryDispatchTracker()

	// Create a dispatch record
	record := DispatchRecord{
		ID:        "dispatch-001",
		SquadName: "dev-team",
		Input: DispatchInput{
			Prompt: "Implement login feature",
		},
		RoutedTo:  "@backend",
		Status:    DispatchStatusPending,
		CreatedAt: time.Now(),
	}

	// Track the dispatch
	if err := tracker.Track(record); err != nil {
		t.Fatalf("Track failed: %v", err)
	}

	// Get the record
	got, err := tracker.Get("dispatch-001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil {
		t.Fatal("Expected record, got nil")
	}
	if got.SquadName != "dev-team" {
		t.Errorf("SquadName = %q, want %q", got.SquadName, "dev-team")
	}
	if got.Status != DispatchStatusPending {
		t.Errorf("Status = %q, want %q", got.Status, DispatchStatusPending)
	}

	// List pending
	pending, err := tracker.ListPending()
	if err != nil {
		t.Fatalf("ListPending failed: %v", err)
	}
	if len(pending) != 1 {
		t.Errorf("ListPending returned %d records, want 1", len(pending))
	}

	// Complete the dispatch
	result := &DispatchResult{
		Raw:      "Feature implemented",
		RoutedTo: "@backend",
	}
	if err := tracker.Complete("dispatch-001", result); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	// Verify completion
	got, _ = tracker.Get("dispatch-001")
	if got.Status != DispatchStatusCompleted {
		t.Errorf("Status = %q, want %q", got.Status, DispatchStatusCompleted)
	}
	if got.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
	if got.Result == nil {
		t.Error("Result should be set")
	}

	// Pending should now be empty
	pending, _ = tracker.ListPending()
	if len(pending) != 0 {
		t.Errorf("ListPending returned %d records, want 0", len(pending))
	}
}

func TestInMemoryDispatchTracker_Fail(t *testing.T) {
	tracker := NewInMemoryDispatchTracker()

	record := DispatchRecord{
		ID:        "dispatch-fail",
		SquadName: "test-team",
		Status:    DispatchStatusPending,
		CreatedAt: time.Now(),
	}
	tracker.Track(record)

	// Fail the dispatch
	if err := tracker.Fail("dispatch-fail", "connection timeout"); err != nil {
		t.Fatalf("Fail failed: %v", err)
	}

	got, _ := tracker.Get("dispatch-fail")
	if got.Status != DispatchStatusFailed {
		t.Errorf("Status = %q, want %q", got.Status, DispatchStatusFailed)
	}
	if got.Error != "connection timeout" {
		t.Errorf("Error = %q, want %q", got.Error, "connection timeout")
	}
}

func TestInMemoryDispatchTracker_List(t *testing.T) {
	tracker := NewInMemoryDispatchTracker()

	// Add multiple dispatches with different statuses
	tracker.Track(DispatchRecord{ID: "d1", Status: DispatchStatusPending})
	tracker.Track(DispatchRecord{ID: "d2", Status: DispatchStatusPending})
	tracker.Track(DispatchRecord{ID: "d3", Status: DispatchStatusCompleted})

	// List all
	all, _ := tracker.List(nil)
	if len(all) != 3 {
		t.Errorf("List(nil) returned %d records, want 3", len(all))
	}

	// List pending only
	pending := DispatchStatusPending
	pendingList, _ := tracker.List(&pending)
	if len(pendingList) != 2 {
		t.Errorf("List(pending) returned %d records, want 2", len(pendingList))
	}

	// List completed only
	completed := DispatchStatusCompleted
	completedList, _ := tracker.List(&completed)
	if len(completedList) != 1 {
		t.Errorf("List(completed) returned %d records, want 1", len(completedList))
	}
}

func TestInMemoryDispatchTracker_UnknownID(t *testing.T) {
	tracker := NewInMemoryDispatchTracker()

	// Get unknown ID should return nil
	got, err := tracker.Get("unknown")
	if err != nil {
		t.Errorf("Get should not error for unknown ID: %v", err)
	}
	if got != nil {
		t.Errorf("Get should return nil for unknown ID, got %+v", got)
	}

	// Complete unknown ID should not error
	if err := tracker.Complete("unknown", nil); err != nil {
		t.Errorf("Complete should not error for unknown ID: %v", err)
	}

	// Fail unknown ID should not error
	if err := tracker.Fail("unknown", "test"); err != nil {
		t.Errorf("Fail should not error for unknown ID: %v", err)
	}
}
