package daemon

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// JobType represents the type of scheduled job.
type JobType string

const (
	JobTypeCron     JobType = "cron"
	JobTypeDaily    JobType = "daily"
	JobTypeWeekly   JobType = "weekly"
	JobTypeMonthly  JobType = "monthly"
	JobTypeOnce     JobType = "once"
	JobTypeInterval JobType = "interval"
)

// ScheduledJob represents a persistent scheduled job.
type ScheduledJob struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Type       JobType         `json:"type"`
	Schedule   json.RawMessage `json:"schedule"`
	Agent      string          `json:"agent"`
	Prompt     string          `json:"prompt,omitempty"`
	OutputPath string          `json:"output_path,omitempty"`
	Singleton  bool            `json:"singleton"`
	Enabled    bool            `json:"enabled"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// JobRunStatus represents the status of a job run.
type JobRunStatus string

const (
	JobRunStatusRunning   JobRunStatus = "running"
	JobRunStatusSuccess   JobRunStatus = "success"
	JobRunStatusFailed    JobRunStatus = "failed"
	JobRunStatusCancelled JobRunStatus = "cancelled"
)

// JobRun represents a single execution of a scheduled job.
type JobRun struct {
	ID             int64        `json:"id"`
	JobID          string       `json:"job_id"`
	StartedAt      time.Time    `json:"started_at"`
	FinishedAt     *time.Time   `json:"finished_at,omitempty"`
	Status         JobRunStatus `json:"status"`
	ErrorMessage   string       `json:"error_message,omitempty"`
	OutputLocation string       `json:"output_location,omitempty"`
}

// JobStore defines the interface for job persistence.
type JobStore interface {
	// CRUD operations
	Create(job *ScheduledJob) error
	Get(id string) (*ScheduledJob, error)
	List() ([]*ScheduledJob, error)
	Update(job *ScheduledJob) error
	Delete(id string) error

	// Run history
	RecordRun(run *JobRun) error
	UpdateRun(run *JobRun) error
	GetRecentRuns(jobID string, limit int) ([]*JobRun, error)

	// Lifecycle
	LoadAllEnabled() ([]*ScheduledJob, error)
	Close() error
}

// SQLiteJobStore implements JobStore using SQLite.
type SQLiteJobStore struct {
	db     *sql.DB
	mu     sync.RWMutex
	dbPath string
}

// JobStoreErrors
var (
	ErrJobNotFound  = errors.New("job not found")
	ErrJobExists    = errors.New("job already exists")
	ErrStoreNotOpen = errors.New("job store not open")
)

// DefaultJobDBPath returns the default path for the jobs database.
func DefaultJobDBPath() string {
	return filepath.Join(paths.DataDir(), "jobs.db")
}

// NewSQLiteJobStore creates a new SQLite-backed job store.
func NewSQLiteJobStore(dbPath string) (*SQLiteJobStore, error) {
	if dbPath == "" {
		dbPath = DefaultJobDBPath()
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	store := &SQLiteJobStore{
		db:     db,
		dbPath: dbPath,
	}

	if err := store.init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	return store, nil
}

// init creates the database schema.
func (s *SQLiteJobStore) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS scheduled_jobs (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		schedule TEXT NOT NULL,
		agent TEXT NOT NULL,
		prompt TEXT,
		output_path TEXT,
		singleton BOOLEAN DEFAULT false,
		enabled BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS job_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id TEXT NOT NULL,
		started_at TIMESTAMP NOT NULL,
		finished_at TIMESTAMP,
		status TEXT NOT NULL,
		error_message TEXT,
		output_location TEXT,
		FOREIGN KEY (job_id) REFERENCES scheduled_jobs(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_job_runs_job_id ON job_runs(job_id);
	CREATE INDEX IF NOT EXISTS idx_job_runs_started_at ON job_runs(started_at);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Create inserts a new scheduled job.
func (s *SQLiteJobStore) Create(job *ScheduledJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return ErrStoreNotOpen
	}

	now := time.Now()
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	job.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO scheduled_jobs (id, name, type, schedule, agent, prompt, output_path, singleton, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.Name, job.Type, job.Schedule, job.Agent, job.Prompt, job.OutputPath, job.Singleton, job.Enabled, job.CreatedAt, job.UpdatedAt)

	if err != nil {
		if isConstraintError(err) {
			return ErrJobExists
		}
		return fmt.Errorf("insert job: %w", err)
	}

	return nil
}

// Get retrieves a job by ID.
func (s *SQLiteJobStore) Get(id string) (*ScheduledJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, ErrStoreNotOpen
	}

	job := &ScheduledJob{}
	err := s.db.QueryRow(`
		SELECT id, name, type, schedule, agent, prompt, output_path, singleton, enabled, created_at, updated_at
		FROM scheduled_jobs WHERE id = ?
	`, id).Scan(&job.ID, &job.Name, &job.Type, &job.Schedule, &job.Agent, &job.Prompt, &job.OutputPath, &job.Singleton, &job.Enabled, &job.CreatedAt, &job.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query job: %w", err)
	}

	return job, nil
}

// List returns all scheduled jobs.
func (s *SQLiteJobStore) List() ([]*ScheduledJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, ErrStoreNotOpen
	}

	rows, err := s.db.Query(`
		SELECT id, name, type, schedule, agent, prompt, output_path, singleton, enabled, created_at, updated_at
		FROM scheduled_jobs ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*ScheduledJob
	for rows.Next() {
		job := &ScheduledJob{}
		if err := rows.Scan(&job.ID, &job.Name, &job.Type, &job.Schedule, &job.Agent, &job.Prompt, &job.OutputPath, &job.Singleton, &job.Enabled, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// Update updates an existing job.
func (s *SQLiteJobStore) Update(job *ScheduledJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return ErrStoreNotOpen
	}

	job.UpdatedAt = time.Now()

	result, err := s.db.Exec(`
		UPDATE scheduled_jobs 
		SET name = ?, type = ?, schedule = ?, agent = ?, prompt = ?, output_path = ?, singleton = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`, job.Name, job.Type, job.Schedule, job.Agent, job.Prompt, job.OutputPath, job.Singleton, job.Enabled, job.UpdatedAt, job.ID)

	if err != nil {
		return fmt.Errorf("update job: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	return nil
}

// Delete removes a job by ID.
func (s *SQLiteJobStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return ErrStoreNotOpen
	}

	result, err := s.db.Exec("DELETE FROM scheduled_jobs WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	return nil
}

// RecordRun creates a new job run record.
func (s *SQLiteJobStore) RecordRun(run *JobRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return ErrStoreNotOpen
	}

	result, err := s.db.Exec(`
		INSERT INTO job_runs (job_id, started_at, finished_at, status, error_message, output_location)
		VALUES (?, ?, ?, ?, ?, ?)
	`, run.JobID, run.StartedAt, run.FinishedAt, run.Status, run.ErrorMessage, run.OutputLocation)

	if err != nil {
		return fmt.Errorf("insert run: %w", err)
	}

	lastID, _ := result.LastInsertId()
	run.ID = lastID

	// Prune old runs (keep last 100 per job)
	_, _ = s.db.Exec(`
		DELETE FROM job_runs WHERE job_id = ? AND id NOT IN (
			SELECT id FROM job_runs WHERE job_id = ? ORDER BY started_at DESC LIMIT 100
		)
	`, run.JobID, run.JobID)

	return nil
}

// UpdateRun updates an existing job run.
func (s *SQLiteJobStore) UpdateRun(run *JobRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return ErrStoreNotOpen
	}

	_, err := s.db.Exec(`
		UPDATE job_runs SET finished_at = ?, status = ?, error_message = ?, output_location = ?
		WHERE id = ?
	`, run.FinishedAt, run.Status, run.ErrorMessage, run.OutputLocation, run.ID)

	if err != nil {
		return fmt.Errorf("update run: %w", err)
	}

	return nil
}

// GetRecentRuns returns the most recent runs for a job.
func (s *SQLiteJobStore) GetRecentRuns(jobID string, limit int) ([]*JobRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, ErrStoreNotOpen
	}

	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.Query(`
		SELECT id, job_id, started_at, finished_at, status, error_message, output_location
		FROM job_runs WHERE job_id = ? ORDER BY started_at DESC LIMIT ?
	`, jobID, limit)
	if err != nil {
		return nil, fmt.Errorf("query runs: %w", err)
	}
	defer rows.Close()

	var runs []*JobRun
	for rows.Next() {
		run := &JobRun{}
		if err := rows.Scan(&run.ID, &run.JobID, &run.StartedAt, &run.FinishedAt, &run.Status, &run.ErrorMessage, &run.OutputLocation); err != nil {
			return nil, fmt.Errorf("scan run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, rows.Err()
}

// LoadAllEnabled returns all enabled jobs.
func (s *SQLiteJobStore) LoadAllEnabled() ([]*ScheduledJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, ErrStoreNotOpen
	}

	rows, err := s.db.Query(`
		SELECT id, name, type, schedule, agent, prompt, output_path, singleton, enabled, created_at, updated_at
		FROM scheduled_jobs WHERE enabled = true ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query enabled jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*ScheduledJob
	for rows.Next() {
		job := &ScheduledJob{}
		if err := rows.Scan(&job.ID, &job.Name, &job.Type, &job.Schedule, &job.Agent, &job.Prompt, &job.OutputPath, &job.Singleton, &job.Enabled, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// Close closes the database connection.
func (s *SQLiteJobStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return nil
	}

	err := s.db.Close()
	s.db = nil
	return err
}

// isConstraintError checks if an error is a constraint violation.
func isConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "UNIQUE constraint failed") || contains(errStr, "PRIMARY KEY constraint failed")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
