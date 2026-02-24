-- +goose Up

-- Add squad_name column for squad-scoped memories
ALTER TABLE memories ADD COLUMN squad_name TEXT;

-- Index for squad queries
CREATE INDEX idx_memories_squad ON memories(squad_name, status);

-- +goose Down

DROP INDEX IF EXISTS idx_memories_squad;
-- SQLite doesn't support DROP COLUMN directly, but the index removal is sufficient
-- for rollback purposes since the column won't be used
