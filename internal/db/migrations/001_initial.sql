-- +goose Up

-- Sessions (conversations)
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    agent_handle TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT 'Untitled Session',
    
    -- Structured I/O
    input_schema TEXT,
    output_schema TEXT,
    structured_input TEXT,
    structured_output TEXT,
    
    -- Chain context
    chain_depth INTEGER NOT NULL DEFAULT 0,
    chain_source TEXT,
    
    -- Stats
    message_count INTEGER NOT NULL DEFAULT 0,
    
    -- Timestamps (Unix seconds)
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    finished_at INTEGER
);

CREATE INDEX idx_sessions_agent ON sessions(agent_handle);
CREATE INDEX idx_sessions_updated ON sessions(updated_at DESC);

-- Messages
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL,
    parts TEXT NOT NULL DEFAULT '[]',
    model TEXT,
    provider TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    finished_at INTEGER,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_session ON messages(session_id);
CREATE INDEX idx_messages_created ON messages(session_id, created_at);

-- Session edges (DAG for parent-child relationships)
CREATE TABLE session_edges (
    parent_id TEXT NOT NULL,
    child_id TEXT NOT NULL,
    edge_type TEXT NOT NULL,
    trigger_message_id TEXT,
    created_at INTEGER NOT NULL,
    PRIMARY KEY (parent_id, child_id),
    FOREIGN KEY (parent_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (child_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (trigger_message_id) REFERENCES messages(id) ON DELETE SET NULL
);

CREATE INDEX idx_edges_parent ON session_edges(parent_id);
CREATE INDEX idx_edges_child ON session_edges(child_id);

-- +goose StatementBegin
CREATE TRIGGER update_session_on_message_insert
AFTER INSERT ON messages
BEGIN
    UPDATE sessions SET 
        message_count = message_count + 1,
        updated_at = strftime('%s', 'now')
    WHERE id = NEW.session_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER update_session_on_message_delete
AFTER DELETE ON messages
BEGIN
    UPDATE sessions SET 
        message_count = message_count - 1,
        updated_at = strftime('%s', 'now')
    WHERE id = OLD.session_id;
END;
-- +goose StatementEnd

-- +goose Down

DROP TRIGGER IF EXISTS update_session_on_message_insert;
DROP TRIGGER IF EXISTS update_session_on_message_delete;
DROP INDEX IF EXISTS idx_edges_child;
DROP INDEX IF EXISTS idx_edges_parent;
DROP INDEX IF EXISTS idx_messages_created;
DROP INDEX IF EXISTS idx_messages_session;
DROP INDEX IF EXISTS idx_sessions_updated;
DROP INDEX IF EXISTS idx_sessions_agent;
DROP TABLE IF EXISTS session_edges;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS sessions;
