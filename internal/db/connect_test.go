package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestConnect(t *testing.T) {
	// Create a temp directory for the test database
	tmpDir, err := os.MkdirTemp("", "ayo-db-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ctx := context.Background()

	// Test connection and migrations
	db, err := Connect(ctx, dbPath)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	tables := []string{"sessions", "messages", "session_edges", "goose_db_version"}
	for _, table := range tables {
		var name string
		err := db.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s not found: %v", table, err)
		}
	}

	// Verify indexes exist
	indexes := []string{"idx_sessions_agent", "idx_sessions_updated", "idx_messages_session", "idx_messages_created", "idx_edges_parent", "idx_edges_child"}
	for _, idx := range indexes {
		var name string
		err := db.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&name)
		if err != nil {
			t.Errorf("index %s not found: %v", idx, err)
		}
	}
}

func TestConnectWithQueries(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-db-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ctx := context.Background()

	db, queries, err := ConnectWithQueries(ctx, dbPath)
	if err != nil {
		t.Fatalf("ConnectWithQueries failed: %v", err)
	}
	defer db.Close()
	defer queries.Close()

	// Test that queries work
	count, err := queries.CountSessions(ctx)
	if err != nil {
		t.Errorf("CountSessions failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 sessions, got %d", count)
	}
}

func TestConnectCreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-db-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use a nested path that doesn't exist
	dbPath := filepath.Join(tmpDir, "nested", "subdir", "test.db")
	ctx := context.Background()

	db, err := Connect(ctx, dbPath)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()

	// Verify the directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}

func TestConnectEmptyPath(t *testing.T) {
	ctx := context.Background()
	_, err := Connect(ctx, "")
	if err == nil {
		t.Error("expected error for empty path")
	}
}
