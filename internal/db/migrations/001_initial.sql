-- +goose Up

-- Sessions (conversations)
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    agent_handle TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT 'Untitled Session',
    source TEXT NOT NULL DEFAULT 'ayo',  -- 'ayo', 'crush', 'crush-via-ayo'
    
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
CREATE INDEX idx_sessions_source ON sessions(source);

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

-- Memories table for agent memory storage
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    
    -- Scope
    agent_handle TEXT,              -- NULL = global, "@ayo" = agent-specific
    path_scope TEXT,                -- NULL = global, "/path/to/project" = path-scoped
    
    -- Content
    content TEXT NOT NULL,          -- The memory itself (natural language)
    category TEXT NOT NULL DEFAULT 'fact',  -- 'preference', 'fact', 'correction', 'pattern'
    
    -- Embedding (stored as BLOB for pure Go similarity search)
    embedding BLOB,                 -- Serialized float32 vector
    
    -- Provenance
    source_session_id TEXT,         -- Which session created this
    source_message_id TEXT,         -- Specific message if applicable
    
    -- Timestamps (Unix seconds)
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    
    -- Lifecycle
    confidence REAL DEFAULT 1.0,    -- How certain the memory is accurate
    last_accessed_at INTEGER,       -- For recency weighting
    access_count INTEGER DEFAULT 0, -- Usage tracking
    
    -- Supersession chain
    supersedes_id TEXT,             -- Points to memory this replaces
    superseded_by_id TEXT,          -- Points to memory that replaced this
    supersession_reason TEXT,       -- Why it was superseded
    
    -- Status
    status TEXT DEFAULT 'active',   -- 'active', 'superseded', 'archived', 'forgotten'
    
    FOREIGN KEY (supersedes_id) REFERENCES memories(id) ON DELETE SET NULL,
    FOREIGN KEY (superseded_by_id) REFERENCES memories(id) ON DELETE SET NULL,
    FOREIGN KEY (source_session_id) REFERENCES sessions(id) ON DELETE SET NULL,
    FOREIGN KEY (source_message_id) REFERENCES messages(id) ON DELETE SET NULL
);

CREATE INDEX idx_memories_agent ON memories(agent_handle, status);
CREATE INDEX idx_memories_path ON memories(path_scope, status);
CREATE INDEX idx_memories_status ON memories(status);
CREATE INDEX idx_memories_category ON memories(category);
CREATE INDEX idx_memories_supersedes ON memories(supersedes_id);
CREATE INDEX idx_memories_superseded_by ON memories(superseded_by_id);
CREATE INDEX idx_memories_created ON memories(created_at DESC);
CREATE INDEX idx_memories_accessed ON memories(last_accessed_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_memories_accessed;
DROP INDEX IF EXISTS idx_memories_created;
DROP INDEX IF EXISTS idx_memories_superseded_by;
DROP INDEX IF EXISTS idx_memories_supersedes;
DROP INDEX IF EXISTS idx_memories_category;
DROP INDEX IF EXISTS idx_memories_status;
DROP INDEX IF EXISTS idx_memories_path;
DROP INDEX IF EXISTS idx_memories_agent;
DROP TABLE IF EXISTS memories;
DROP TRIGGER IF EXISTS update_session_on_message_insert;
DROP TRIGGER IF EXISTS update_session_on_message_delete;
DROP INDEX IF EXISTS idx_edges_child;
DROP INDEX IF EXISTS idx_edges_parent;
DROP INDEX IF EXISTS idx_messages_created;
DROP INDEX IF EXISTS idx_messages_session;
DROP INDEX IF EXISTS idx_sessions_source;
DROP INDEX IF EXISTS idx_sessions_updated;
DROP INDEX IF EXISTS idx_sessions_agent;
DROP TABLE IF EXISTS session_edges;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS sessions;
