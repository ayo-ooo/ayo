-- +goose Up

-- Agents created dynamically by @ayo
CREATE TABLE ayo_created_agents (
    agent_id TEXT PRIMARY KEY,              -- Agent identifier (e.g., 'science-researcher')
    agent_handle TEXT NOT NULL,             -- Full handle (e.g., '@science-researcher')
    
    -- Creation context
    created_by TEXT NOT NULL,               -- '@ayo' or other orchestrating agent
    creation_reason TEXT,                   -- Why the agent was created
    original_prompt TEXT NOT NULL,          -- Initial system prompt
    
    -- Learning metrics
    invocation_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,
    last_used_at INTEGER,                   -- Unix timestamp
    
    -- Evolution tracking
    refinement_count INTEGER NOT NULL DEFAULT 0,
    current_prompt_hash TEXT,               -- SHA256 of current system prompt
    
    -- Lifecycle
    confidence REAL NOT NULL DEFAULT 0.0,   -- 0.0-1.0, increases with success
    is_archived INTEGER NOT NULL DEFAULT 0, -- 1 = hidden but not deleted
    promoted_to TEXT,                       -- New handle if promoted
    
    -- Timestamps (Unix seconds)
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX idx_ayo_agents_handle ON ayo_created_agents(agent_handle);
CREATE INDEX idx_ayo_agents_creator ON ayo_created_agents(created_by);
CREATE INDEX idx_ayo_agents_archived ON ayo_created_agents(is_archived);
CREATE INDEX idx_ayo_agents_confidence ON ayo_created_agents(confidence DESC);
CREATE INDEX idx_ayo_agents_used ON ayo_created_agents(last_used_at DESC);

-- Refinement history for @ayo-created agents
CREATE TABLE agent_refinements (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    
    -- What changed
    previous_prompt TEXT NOT NULL,
    new_prompt TEXT NOT NULL,
    reason TEXT NOT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL,
    
    FOREIGN KEY (agent_id) REFERENCES ayo_created_agents(agent_id) ON DELETE CASCADE
);

CREATE INDEX idx_refinements_agent ON agent_refinements(agent_id);
CREATE INDEX idx_refinements_created ON agent_refinements(created_at DESC);

-- Trigger execution statistics (for flow triggers)
CREATE TABLE trigger_stats (
    trigger_id TEXT PRIMARY KEY,            -- Unique trigger identifier
    flow_name TEXT NOT NULL,                -- Associated flow
    trigger_type TEXT NOT NULL,             -- 'cron' or 'watch'
    
    -- Execution stats
    run_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,
    
    -- Permanence tracking
    runs_before_permanent INTEGER,          -- NULL = already permanent
    is_permanent INTEGER NOT NULL DEFAULT 0,
    
    -- Timing
    last_run_at INTEGER,                    -- Unix timestamp
    last_success_at INTEGER,
    last_failure_at INTEGER,
    avg_duration_ms INTEGER,
    
    -- Timestamps
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX idx_trigger_stats_flow ON trigger_stats(flow_name);
CREATE INDEX idx_trigger_stats_type ON trigger_stats(trigger_type);
CREATE INDEX idx_trigger_stats_permanent ON trigger_stats(is_permanent);

-- Inferred agent capabilities (for semantic search)
CREATE TABLE agent_capabilities (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,                 -- Agent this belongs to
    
    -- Capability info
    name TEXT NOT NULL,                     -- Short name: 'code-review', 'summarization'
    description TEXT NOT NULL,              -- Longer explanation
    confidence REAL NOT NULL,               -- 0.0-1.0, how confident we are
    source TEXT NOT NULL,                   -- 'system_prompt', 'skill', 'schema'
    
    -- Vector embedding for semantic search (serialized float32 array)
    embedding BLOB,
    
    -- Cache management
    input_hash TEXT NOT NULL,               -- Hash of inference inputs for invalidation
    
    -- Timestamps
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX idx_capabilities_agent ON agent_capabilities(agent_id);
CREATE INDEX idx_capabilities_name ON agent_capabilities(name);
CREATE INDEX idx_capabilities_confidence ON agent_capabilities(confidence DESC);
CREATE INDEX idx_capabilities_hash ON agent_capabilities(input_hash);

-- +goose Down

DROP INDEX IF EXISTS idx_capabilities_hash;
DROP INDEX IF EXISTS idx_capabilities_confidence;
DROP INDEX IF EXISTS idx_capabilities_name;
DROP INDEX IF EXISTS idx_capabilities_agent;
DROP TABLE IF EXISTS agent_capabilities;

DROP INDEX IF EXISTS idx_trigger_stats_permanent;
DROP INDEX IF EXISTS idx_trigger_stats_type;
DROP INDEX IF EXISTS idx_trigger_stats_flow;
DROP TABLE IF EXISTS trigger_stats;

DROP INDEX IF EXISTS idx_refinements_created;
DROP INDEX IF EXISTS idx_refinements_agent;
DROP TABLE IF EXISTS agent_refinements;

DROP INDEX IF EXISTS idx_ayo_agents_used;
DROP INDEX IF EXISTS idx_ayo_agents_confidence;
DROP INDEX IF EXISTS idx_ayo_agents_archived;
DROP INDEX IF EXISTS idx_ayo_agents_creator;
DROP INDEX IF EXISTS idx_ayo_agents_handle;
DROP TABLE IF EXISTS ayo_created_agents;
