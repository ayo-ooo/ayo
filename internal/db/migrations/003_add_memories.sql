-- +goose Up

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

-- Indexes for efficient queries
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
