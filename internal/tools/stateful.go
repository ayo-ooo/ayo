package tools

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/paths"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// StatefulTool extends AgentTool with persistent storage capabilities.
// Tools that need to persist data across sessions should implement this interface.
type StatefulTool interface {
	fantasy.AgentTool

	// Storage returns the tool's data directory path.
	// Example: ~/.local/share/ayo/tools/todo/
	Storage() string

	// DatabasePath returns the path to the tool's SQLite database.
	// Example: ~/.local/share/ayo/tools/todo/todo.db
	DatabasePath() string

	// Init initializes the tool's storage (creates DB, runs migrations).
	// Called once when the tool is first loaded in a session.
	Init(ctx context.Context) error

	// Close releases any resources held by the tool.
	Close() error
}

// StatefulToolBase provides common implementation for stateful tools.
// Embed this in your tool implementation to get standard storage behavior.
type StatefulToolBase struct {
	name string
	db   *sql.DB
}

// NewStatefulToolBase creates a new base with the given tool name.
func NewStatefulToolBase(name string) StatefulToolBase {
	return StatefulToolBase{name: name}
}

// Name returns the tool name.
func (t *StatefulToolBase) Name() string {
	return t.name
}

// Storage returns the tool's data directory path.
func (t *StatefulToolBase) Storage() string {
	return paths.ToolDataDir(t.name)
}

// DatabasePath returns the path to the tool's SQLite database.
func (t *StatefulToolBase) DatabasePath() string {
	return paths.ToolDatabasePath(t.name)
}

// EnsureStorage creates the tool's data directory if it doesn't exist.
func (t *StatefulToolBase) EnsureStorage() error {
	return os.MkdirAll(t.Storage(), 0o755)
}

// OpenDatabase opens the tool's SQLite database.
// Call this from your Init() implementation.
func (t *StatefulToolBase) OpenDatabase(ctx context.Context) (*sql.DB, error) {
	if err := t.EnsureStorage(); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}

	dbPath := t.DatabasePath()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}

	t.db = db
	return db, nil
}

// DB returns the open database connection.
// Returns nil if Init() hasn't been called.
func (t *StatefulToolBase) DB() *sql.DB {
	return t.db
}

// Close closes the database connection.
func (t *StatefulToolBase) Close() error {
	if t.db != nil {
		return t.db.Close()
	}
	return nil
}

// RunMigration executes a SQL migration string.
// Call this from your Init() after OpenDatabase().
func (t *StatefulToolBase) RunMigration(ctx context.Context, schema string) error {
	if t.db == nil {
		return fmt.Errorf("database not opened")
	}
	_, err := t.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("run migration: %w", err)
	}
	return nil
}

// SessionDataPath returns a path for session-specific data files.
// Useful for tools that need to store non-DB data per session.
func (t *StatefulToolBase) SessionDataPath(sessionID, filename string) string {
	return filepath.Join(t.Storage(), "sessions", sessionID, filename)
}
