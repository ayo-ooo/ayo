package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileLogger_LogAndQuery(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	// Create logger
	logger, err := NewFileLoggerWithPath(logPath)
	if err != nil {
		t.Fatalf("NewFileLoggerWithPath: %v", err)
	}
	defer logger.Close()

	// Log some entries
	entries := []Entry{
		{
			Timestamp:   time.Now().UTC(),
			Agent:       "@ayo",
			Session:     "session1",
			Action:      ActionCreate,
			Path:        "/home/user/file1.txt",
			Approval:    ApprovalUserApproved,
			Size:        100,
			ContentHash: "sha256:abc",
		},
		{
			Timestamp:   time.Now().UTC(),
			Agent:       "@crush",
			Session:     "session1",
			Action:      ActionUpdate,
			Path:        "/home/user/file2.txt",
			Approval:    ApprovalNoJodas,
			Size:        200,
			ContentHash: "sha256:def",
		},
		{
			Timestamp:   time.Now().UTC(),
			Agent:       "@ayo",
			Session:     "session2",
			Action:      ActionDelete,
			Path:        "/home/user/file3.txt",
			Approval:    ApprovalSessionCache,
		},
	}

	for _, entry := range entries {
		if err := logger.Log(entry); err != nil {
			t.Fatalf("Log: %v", err)
		}
	}

	// Query all
	results, err := logger.Query(Filter{})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 entries, got %d", len(results))
	}

	// Query by agent
	results, err = logger.Query(Filter{Agent: "@ayo"})
	if err != nil {
		t.Fatalf("Query by agent: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 entries for @ayo, got %d", len(results))
	}

	// Query by session
	results, err = logger.Query(Filter{Session: "session1"})
	if err != nil {
		t.Fatalf("Query by session: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 entries for session1, got %d", len(results))
	}

	// Query by action
	results, err = logger.Query(Filter{Action: ActionDelete})
	if err != nil {
		t.Fatalf("Query by action: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 delete entry, got %d", len(results))
	}

	// Query with limit
	results, err = logger.Query(Filter{Limit: 1})
	if err != nil {
		t.Fatalf("Query with limit: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 entry with limit, got %d", len(results))
	}
}

func TestFileLogger_Rotation(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	// Create logger with small max size
	logger, err := NewFileLoggerWithPath(logPath)
	if err != nil {
		t.Fatalf("NewFileLoggerWithPath: %v", err)
	}
	logger.maxSize = 500 // Small size to trigger rotation
	logger.maxBackups = 2
	defer logger.Close()

	// Write enough entries to trigger rotation
	for i := 0; i < 20; i++ {
		entry := Entry{
			Timestamp: time.Now().UTC(),
			Agent:     "@test",
			Session:   "session",
			Action:    ActionCreate,
			Path:      "/home/user/file.txt",
			Approval:  ApprovalUserApproved,
			Size:      1000,
		}
		if err := logger.Log(entry); err != nil {
			t.Fatalf("Log: %v", err)
		}
	}

	// Check that backup files exist
	if _, err := os.Stat(logPath + ".1"); os.IsNotExist(err) {
		t.Error("expected backup file .1 to exist")
	}
}

func TestLogFileModification(t *testing.T) {
	// Create temp directory and set up logger
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	logger, err := NewFileLoggerWithPath(logPath)
	if err != nil {
		t.Fatalf("NewFileLoggerWithPath: %v", err)
	}
	
	// Replace global logger temporarily
	loggerMu.Lock()
	oldLogger := defaultLogger
	defaultLogger = logger
	loggerMu.Unlock()
	defer func() {
		loggerMu.Lock()
		defaultLogger = oldLogger
		loggerMu.Unlock()
		logger.Close()
	}()

	// Log via convenience function
	err = LogFileModification("@agent", "sess123", ActionCreate, "/path/to/file", ApprovalUserApproved, 500, "sha256:xyz")
	if err != nil {
		t.Fatalf("LogFileModification: %v", err)
	}

	// Query to verify
	results, err := logger.Query(Filter{Agent: "@agent"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(results))
	}
	
	entry := results[0]
	if entry.Agent != "@agent" {
		t.Errorf("expected agent @agent, got %s", entry.Agent)
	}
	if entry.Session != "sess123" {
		t.Errorf("expected session sess123, got %s", entry.Session)
	}
	if entry.Size != 500 {
		t.Errorf("expected size 500, got %d", entry.Size)
	}
}

func TestFilter_TimeRange(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	logger, err := NewFileLoggerWithPath(logPath)
	if err != nil {
		t.Fatalf("NewFileLoggerWithPath: %v", err)
	}
	defer logger.Close()

	now := time.Now().UTC()
	
	// Log entries at different times
	logger.Log(Entry{
		Timestamp: now.Add(-2 * time.Hour),
		Agent:     "@test",
		Action:    ActionCreate,
	})
	logger.Log(Entry{
		Timestamp: now.Add(-1 * time.Hour),
		Agent:     "@test",
		Action:    ActionUpdate,
	})
	logger.Log(Entry{
		Timestamp: now,
		Agent:     "@test",
		Action:    ActionDelete,
	})

	// Query since 90 minutes ago
	results, err := logger.Query(Filter{
		Since: now.Add(-90 * time.Minute),
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 entries since 90 min ago, got %d", len(results))
	}
}
