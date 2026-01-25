-- +goose Up

-- Add source column to track where sessions originate
-- Values: 'ayo' (default), 'crush', 'crush-via-ayo'
ALTER TABLE sessions ADD COLUMN source TEXT NOT NULL DEFAULT 'ayo';

CREATE INDEX idx_sessions_source ON sessions(source);

-- +goose Down

DROP INDEX IF EXISTS idx_sessions_source;
-- SQLite doesn't support DROP COLUMN, so we need to recreate the table
-- For simplicity, we leave the column in place during down migration
