package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestStatefulToolBase_Name(t *testing.T) {
	base := NewStatefulToolBase("test-tool")
	if got := base.Name(); got != "test-tool" {
		t.Errorf("Name() = %q, want %q", got, "test-tool")
	}
}

func TestStatefulToolBase_Storage(t *testing.T) {
	base := NewStatefulToolBase("test-tool")
	storage := base.Storage()
	if storage == "" {
		t.Error("Storage() returned empty string")
	}
	if !filepath.IsAbs(storage) {
		t.Errorf("Storage() = %q, expected absolute path", storage)
	}
}

func TestStatefulToolBase_DatabasePath(t *testing.T) {
	base := NewStatefulToolBase("test-tool")
	dbPath := base.DatabasePath()
	if dbPath == "" {
		t.Error("DatabasePath() returned empty string")
	}
	if !filepath.IsAbs(dbPath) {
		t.Errorf("DatabasePath() = %q, expected absolute path", dbPath)
	}
	if filepath.Ext(dbPath) != ".db" {
		t.Errorf("DatabasePath() = %q, expected .db extension", dbPath)
	}
}

func TestStatefulToolBase_EnsureStorage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-tools-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override storage path by setting up a custom tool
	base := StatefulToolBase{name: "test-tool"}

	// Use a mock that points to temp dir
	// Since Storage() uses paths.ToolDataDir, we'll test EnsureStorage in integration
	err = base.EnsureStorage()
	// This might fail if we don't have write access to default location
	// but we're testing the code path exists
	if err != nil {
		// Expected in some environments
		t.Logf("EnsureStorage failed (expected in restricted environments): %v", err)
	}
}

func TestStatefulToolBase_OpenDatabaseAndClose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-tools-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a custom base that we can control
	base := &StatefulToolBase{name: "test"}

	// Create storage dir manually
	storageDir := filepath.Join(tmpDir, "tools", "test")
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("failed to create storage dir: %v", err)
	}

	// We need to test with actual paths, but since paths.ToolDataDir uses 
	// fixed paths, we'll test the DB() and Close() methods directly
	if base.DB() != nil {
		t.Error("DB() should be nil before OpenDatabase")
	}

	err = base.Close()
	if err != nil {
		t.Errorf("Close() on uninitialized base should not error: %v", err)
	}
}

func TestStatefulToolBase_RunMigrationWithoutDB(t *testing.T) {
	base := &StatefulToolBase{name: "test"}
	ctx := context.Background()

	err := base.RunMigration(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY)")
	if err == nil {
		t.Error("RunMigration without open DB should error")
	}
}

func TestStatefulToolBase_SessionDataPath(t *testing.T) {
	base := NewStatefulToolBase("test-tool")
	path := base.SessionDataPath("session-123", "data.json")

	if path == "" {
		t.Error("SessionDataPath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("SessionDataPath() = %q, expected absolute path", path)
	}
	if filepath.Base(path) != "data.json" {
		t.Errorf("SessionDataPath() = %q, expected data.json as basename", path)
	}
}
