// Package todosdb provides the ayo-todos-db near-term planner plugin.
// This planner uses SQLite storage for better performance with large todo lists.
package todosdb

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
const PluginName = "ayo-todos-db"

// DatabaseFile is the name of the SQLite database within the planner's state directory.
const DatabaseFile = "todos.db"

// Plugin implements the PlannerPlugin interface for ayo-todos-db.
// It provides session-scoped todo list management using SQLite storage.
type Plugin struct {
	stateDir string
	db       *sql.DB
}

// New returns a factory function that creates new Plugin instances.
// This is used by the planner registry to instantiate the planner.
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

// Type returns the planner type (near-term for todos).
func (p *Plugin) Type() planners.PlannerType {
	return planners.NearTerm
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

	p.db = db
	return nil
}

// schema is the SQL schema for the todos database.
const schema = `
CREATE TABLE IF NOT EXISTS todos (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	content TEXT NOT NULL,
	active_form TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'pending',
	created_at INTEGER NOT NULL DEFAULT (unixepoch()),
	updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status);
CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at);
`

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
		p.newTodosTool(),
	}
}

// Instructions returns text to inject into agent system prompts.
func (p *Plugin) Instructions() string {
	return TodosInstructions
}

// TodosInstructions contains the system prompt instructions for the todos tool.
const TodosInstructions = `## Near-Term Task Management

Use the todos tool to track progress on complex, multi-step tasks:

**When to use:**
- Tasks requiring 3+ distinct steps
- Multi-file changes that need tracking
- Work that benefits from explicit progress tracking

**How to use:**
- Create specific, actionable todo items
- Keep exactly ONE task in_progress at a time
- Mark tasks complete IMMEDIATELY after finishing
- Update todos proactively as work progresses

**Task states:**
- pending: Not yet started
- in_progress: Currently working on (only one)
- completed: Finished successfully

**Required fields for each todo:**
- content: What needs to be done (imperative form, e.g., "Run tests")
- active_form: Present continuous form (e.g., "Running tests")
- status: One of pending, in_progress, completed

**Best practices:**
- Break complex tasks into smaller steps
- Don't batch completions; mark done immediately
- Remove irrelevant tasks from the list
- The user can see your todo list in real-time
`

// StateDir returns the directory where this planner stores its state.
func (p *Plugin) StateDir() string {
	return p.stateDir
}

// DB returns the database connection for testing.
func (p *Plugin) DB() *sql.DB {
	return p.db
}

// Register adds the ayo-todos-db plugin to the default registry.
// This is called from init() to ensure the plugin is available at startup.
func Register() {
	planners.DefaultRegistry.Register(PluginName, New())
}

func init() {
	Register()
}
