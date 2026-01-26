-- +goose Up

-- Flow runs (execution history)
CREATE TABLE flow_runs (
    id TEXT PRIMARY KEY,                    -- ULID for sortable unique ID
    
    -- Flow identification
    flow_name TEXT NOT NULL,                -- Name from frontmatter
    flow_path TEXT NOT NULL,                -- Absolute path at time of execution
    flow_source TEXT NOT NULL,              -- 'project', 'user', 'builtin'
    
    -- Execution status
    status TEXT NOT NULL DEFAULT 'running', -- 'running', 'success', 'failed', 'timeout'
    exit_code INTEGER,                      -- Exit code (null while running)
    error_message TEXT,                     -- Error description if failed
    
    -- Input/Output
    input_json TEXT,                        -- JSON input (nullable)
    output_json TEXT,                       -- JSON output from stdout (nullable)
    stderr_log TEXT,                        -- Captured stderr
    
    -- Timing (Unix milliseconds for precision)
    started_at INTEGER NOT NULL,
    finished_at INTEGER,
    duration_ms INTEGER,
    
    -- Relationships
    parent_run_id TEXT,                     -- If triggered by another flow
    session_id TEXT,                        -- If triggered from a session
    
    -- Validation
    input_validated INTEGER NOT NULL DEFAULT 0,  -- 1 if input was validated
    output_validated INTEGER NOT NULL DEFAULT 0, -- 1 if output was validated
    
    FOREIGN KEY (parent_run_id) REFERENCES flow_runs(id) ON DELETE SET NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
);

CREATE INDEX idx_flow_runs_name ON flow_runs(flow_name);
CREATE INDEX idx_flow_runs_status ON flow_runs(status);
CREATE INDEX idx_flow_runs_started ON flow_runs(started_at DESC);
CREATE INDEX idx_flow_runs_parent ON flow_runs(parent_run_id);
CREATE INDEX idx_flow_runs_session ON flow_runs(session_id);

-- +goose Down

DROP INDEX IF EXISTS idx_flow_runs_session;
DROP INDEX IF EXISTS idx_flow_runs_parent;
DROP INDEX IF EXISTS idx_flow_runs_started;
DROP INDEX IF EXISTS idx_flow_runs_status;
DROP INDEX IF EXISTS idx_flow_runs_name;
DROP TABLE IF EXISTS flow_runs;
