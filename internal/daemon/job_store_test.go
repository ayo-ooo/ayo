package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteJobStore_CRUD(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewSQLiteJobStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	t.Run("create and get", func(t *testing.T) {
		job := &ScheduledJob{
			ID:       "test-job-1",
			Name:     "Test Job",
			Type:     JobTypeCron,
			Schedule: json.RawMessage(`{"cron": "0 9 * * *"}`),
			Agent:    "@ayo",
			Prompt:   "Run daily check",
			Enabled:  true,
		}

		if err := store.Create(job); err != nil {
			t.Fatalf("create failed: %v", err)
		}

		got, err := store.Get("test-job-1")
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}

		if got.Name != "Test Job" {
			t.Errorf("expected name 'Test Job', got %q", got.Name)
		}
		if got.Agent != "@ayo" {
			t.Errorf("expected agent '@ayo', got %q", got.Agent)
		}
		if !got.Enabled {
			t.Error("expected job to be enabled")
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := store.Get("nonexistent")
		if err != ErrJobNotFound {
			t.Errorf("expected ErrJobNotFound, got %v", err)
		}
	})

	t.Run("create duplicate fails", func(t *testing.T) {
		job := &ScheduledJob{
			ID:       "test-job-1",
			Name:     "Duplicate",
			Type:     JobTypeCron,
			Schedule: json.RawMessage(`{}`),
			Agent:    "@ayo",
			Enabled:  true,
		}

		err := store.Create(job)
		if err != ErrJobExists {
			t.Errorf("expected ErrJobExists, got %v", err)
		}
	})

	t.Run("update", func(t *testing.T) {
		job, _ := store.Get("test-job-1")
		job.Name = "Updated Name"
		job.Enabled = false

		if err := store.Update(job); err != nil {
			t.Fatalf("update failed: %v", err)
		}

		got, _ := store.Get("test-job-1")
		if got.Name != "Updated Name" {
			t.Errorf("expected name 'Updated Name', got %q", got.Name)
		}
		if got.Enabled {
			t.Error("expected job to be disabled")
		}
	})

	t.Run("update not found", func(t *testing.T) {
		job := &ScheduledJob{
			ID:   "nonexistent",
			Name: "Does not exist",
		}

		err := store.Update(job)
		if err != ErrJobNotFound {
			t.Errorf("expected ErrJobNotFound, got %v", err)
		}
	})

	t.Run("list", func(t *testing.T) {
		// Create another job
		job2 := &ScheduledJob{
			ID:       "test-job-2",
			Name:     "Second Job",
			Type:     JobTypeInterval,
			Schedule: json.RawMessage(`{"every": "5m"}`),
			Agent:    "@bot",
			Enabled:  true,
		}
		store.Create(job2)

		jobs, err := store.List()
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}

		if len(jobs) < 2 {
			t.Errorf("expected at least 2 jobs, got %d", len(jobs))
		}
	})

	t.Run("delete", func(t *testing.T) {
		if err := store.Delete("test-job-1"); err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		_, err := store.Get("test-job-1")
		if err != ErrJobNotFound {
			t.Errorf("expected job to be deleted, got %v", err)
		}
	})

	t.Run("delete not found", func(t *testing.T) {
		err := store.Delete("nonexistent")
		if err != ErrJobNotFound {
			t.Errorf("expected ErrJobNotFound, got %v", err)
		}
	})
}

func TestSQLiteJobStore_JobRuns(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewSQLiteJobStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create a job first
	job := &ScheduledJob{
		ID:       "job-with-runs",
		Name:     "Job with runs",
		Type:     JobTypeCron,
		Schedule: json.RawMessage(`{"cron": "0 * * * *"}`),
		Agent:    "@ayo",
		Enabled:  true,
	}
	store.Create(job)

	t.Run("record and get runs", func(t *testing.T) {
		now := time.Now()
		finished := now.Add(5 * time.Minute)
		run := &JobRun{
			JobID:     "job-with-runs",
			StartedAt: now,
			FinishedAt: &finished,
			Status:    JobRunStatusSuccess,
		}

		if err := store.RecordRun(run); err != nil {
			t.Fatalf("record run failed: %v", err)
		}

		if run.ID == 0 {
			t.Error("expected run ID to be set")
		}

		runs, err := store.GetRecentRuns("job-with-runs", 10)
		if err != nil {
			t.Fatalf("get runs failed: %v", err)
		}

		if len(runs) != 1 {
			t.Errorf("expected 1 run, got %d", len(runs))
		}
		if runs[0].Status != JobRunStatusSuccess {
			t.Errorf("expected status 'success', got %q", runs[0].Status)
		}
	})

	t.Run("update run", func(t *testing.T) {
		runs, _ := store.GetRecentRuns("job-with-runs", 1)
		run := runs[0]
		run.Status = JobRunStatusFailed
		run.ErrorMessage = "something went wrong"

		if err := store.UpdateRun(run); err != nil {
			t.Fatalf("update run failed: %v", err)
		}

		updated, _ := store.GetRecentRuns("job-with-runs", 1)
		if updated[0].Status != JobRunStatusFailed {
			t.Errorf("expected status 'failed', got %q", updated[0].Status)
		}
		if updated[0].ErrorMessage != "something went wrong" {
			t.Errorf("unexpected error message: %q", updated[0].ErrorMessage)
		}
	})
}

func TestSQLiteJobStore_LoadAllEnabled(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewSQLiteJobStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create enabled and disabled jobs
	enabled1 := &ScheduledJob{
		ID:       "enabled-1",
		Name:     "Enabled 1",
		Type:     JobTypeCron,
		Schedule: json.RawMessage(`{}`),
		Agent:    "@ayo",
		Enabled:  true,
	}
	enabled2 := &ScheduledJob{
		ID:       "enabled-2",
		Name:     "Enabled 2",
		Type:     JobTypeInterval,
		Schedule: json.RawMessage(`{}`),
		Agent:    "@bot",
		Enabled:  true,
	}
	disabled := &ScheduledJob{
		ID:       "disabled-1",
		Name:     "Disabled",
		Type:     JobTypeCron,
		Schedule: json.RawMessage(`{}`),
		Agent:    "@ayo",
		Enabled:  false,
	}

	store.Create(enabled1)
	store.Create(enabled2)
	store.Create(disabled)

	jobs, err := store.LoadAllEnabled()
	if err != nil {
		t.Fatalf("load enabled failed: %v", err)
	}

	if len(jobs) != 2 {
		t.Errorf("expected 2 enabled jobs, got %d", len(jobs))
	}

	// Check disabled is not included
	for _, j := range jobs {
		if j.ID == "disabled-1" {
			t.Error("disabled job should not be included")
		}
	}
}

func TestSQLiteJobStore_Close(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewSQLiteJobStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Close the store
	if err := store.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// Operations should fail after close
	_, err = store.Get("any")
	if err != ErrStoreNotOpen {
		t.Errorf("expected ErrStoreNotOpen after close, got %v", err)
	}

	// Double close should be safe
	if err := store.Close(); err != nil {
		t.Errorf("double close should succeed, got %v", err)
	}
}

func TestSQLiteJobStore_Persistence(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "persist.db")

	// Create store and add a job
	store1, err := NewSQLiteJobStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	job := &ScheduledJob{
		ID:       "persistent-job",
		Name:     "Persists across restarts",
		Type:     JobTypeCron,
		Schedule: json.RawMessage(`{"cron": "0 9 * * *"}`),
		Agent:    "@ayo",
		Enabled:  true,
	}
	store1.Create(job)
	store1.Close()

	// Re-open and verify job exists
	store2, err := NewSQLiteJobStore(dbPath)
	if err != nil {
		t.Fatalf("failed to re-open store: %v", err)
	}
	defer store2.Close()

	got, err := store2.Get("persistent-job")
	if err != nil {
		t.Fatalf("job should persist: %v", err)
	}
	if got.Name != "Persists across restarts" {
		t.Errorf("unexpected name: %q", got.Name)
	}
}

func TestDefaultJobDBPath(t *testing.T) {
	path := DefaultJobDBPath()
	if path == "" {
		t.Error("expected non-empty path")
	}

	// Should end with jobs.db
	if !contains(path, "jobs.db") {
		t.Errorf("expected path to contain 'jobs.db', got %q", path)
	}
}

func TestJobTypes(t *testing.T) {
	types := []JobType{
		JobTypeCron,
		JobTypeDaily,
		JobTypeWeekly,
		JobTypeMonthly,
		JobTypeOnce,
		JobTypeInterval,
	}

	for _, jt := range types {
		if jt == "" {
			t.Error("job type should not be empty")
		}
	}
}

func TestJobRunStatuses(t *testing.T) {
	statuses := []JobRunStatus{
		JobRunStatusRunning,
		JobRunStatusSuccess,
		JobRunStatusFailed,
		JobRunStatusCancelled,
	}

	for _, s := range statuses {
		if s == "" {
			t.Error("job run status should not be empty")
		}
	}
}

func TestIsConstraintError(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{os.ErrNotExist, false},
	}

	for _, tt := range tests {
		got := isConstraintError(tt.err)
		if got != tt.expected {
			t.Errorf("isConstraintError(%v) = %v, want %v", tt.err, got, tt.expected)
		}
	}
}
