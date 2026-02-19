// Package kanban provides the ayo-kanban long-term planner plugin.
// This planner implements a kanban board with columns, WIP limits, and card movement.
package kanban

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/fantasy"
	"github.com/alexcabrera/ayo/internal/planners"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// PluginName is the identifier for this planner in the registry.
const PluginName = "ayo-kanban"

// DatabaseFile is the name of the SQLite database within the planner's state directory.
const DatabaseFile = "kanban.db"

// Default columns for a kanban board.
var DefaultColumns = []string{"backlog", "ready", "in_progress", "review", "done"}

// Default WIP (Work In Progress) limits for columns.
// 0 means no limit.
var DefaultWIPLimits = map[string]int{
	"backlog":     0,
	"ready":       5,
	"in_progress": 3,
	"review":      3,
	"done":        0,
}

// Plugin implements the PlannerPlugin interface for ayo-kanban.
// It provides kanban-style work management for agents.
type Plugin struct {
	stateDir string
	db       *sql.DB
}

// New returns a factory function that creates new Plugin instances.
func New() planners.PlannerFactory {
	return func(ctx planners.PlannerContext) (planners.PlannerPlugin, error) {
		return &Plugin{
			stateDir: ctx.StateDir,
		}, nil
	}
}

// Name returns the unique identifier for this planner.
func (p *Plugin) Name() string {
	return PluginName
}

// Type returns the planner type (long-term for kanban).
func (p *Plugin) Type() planners.PlannerType {
	return planners.LongTerm
}

// Init initializes the planner, opening the SQLite database.
func (p *Plugin) Init(ctx context.Context) error {
	// Ensure state directory exists
	if err := os.MkdirAll(p.stateDir, 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	// Open database
	dbPath := filepath.Join(p.stateDir, DatabaseFile)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return fmt.Errorf("enable WAL mode: %w", err)
	}

	// Run migration
	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return fmt.Errorf("run migration: %w", err)
	}

	// Initialize default columns if empty
	if err := p.ensureDefaultColumns(ctx, db); err != nil {
		db.Close()
		return fmt.Errorf("init default columns: %w", err)
	}

	p.db = db
	return nil
}

// schema is the SQL schema for the kanban database.
const schema = `
CREATE TABLE IF NOT EXISTS columns (
	name TEXT PRIMARY KEY,
	position INTEGER NOT NULL,
	wip_limit INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS cards (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	column_name TEXT NOT NULL REFERENCES columns(name),
	priority INTEGER NOT NULL DEFAULT 0,
	position INTEGER NOT NULL DEFAULT 0,
	created_at INTEGER NOT NULL DEFAULT (unixepoch()),
	updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_cards_column ON cards(column_name);
CREATE INDEX IF NOT EXISTS idx_cards_priority ON cards(priority DESC);
`

// ensureDefaultColumns creates default columns if the board is empty.
func (p *Plugin) ensureDefaultColumns(ctx context.Context, db *sql.DB) error {
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM columns").Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil // Already has columns
	}

	// Insert default columns
	stmt, err := db.PrepareContext(ctx, "INSERT INTO columns (name, position, wip_limit) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, col := range DefaultColumns {
		limit := DefaultWIPLimits[col]
		if _, err := stmt.ExecContext(ctx, col, i, limit); err != nil {
			return err
		}
	}

	return nil
}

// Close releases any resources held by the planner.
func (p *Plugin) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Tools returns the fantasy tools that this planner provides.
func (p *Plugin) Tools() []fantasy.AgentTool {
	return []fantasy.AgentTool{
		p.newKanbanTool(),
	}
}

// Instructions returns text to inject into agent system prompts.
func (p *Plugin) Instructions() string {
	return KanbanInstructions
}

// KanbanInstructions contains the system prompt instructions for the kanban tool.
const KanbanInstructions = `## Kanban Board Management

Use the kanban tool for visual work management with columns and cards.

**Board structure:**
- Columns: backlog → ready → in_progress → review → done
- Cards: Work items that move through columns
- WIP limits: Maximum cards per column (enforced for in_progress, review)

**Actions:**
- board: View entire board with all columns and cards
- add: Create a new card in a column (default: backlog)
- move: Move a card to a different column
- update: Update card title or description
- remove: Delete a card

**When to use:**
- Visualizing work flow
- Managing work in progress
- Tracking items through stages
- Coordinating multi-step work

**Best practices:**
- Pull cards from left to right as work progresses
- Respect WIP limits to avoid overload
- Keep backlog prioritized
- Move done items regularly to maintain flow
- Add descriptions for context
`

// StateDir returns the directory where this planner stores its state.
func (p *Plugin) StateDir() string {
	return p.stateDir
}

// DB returns the database connection for testing.
func (p *Plugin) DB() *sql.DB {
	return p.db
}

// Register adds the ayo-kanban plugin to the default registry.
func Register() {
	planners.DefaultRegistry.Register(PluginName, New())
}

func init() {
	Register()
}
