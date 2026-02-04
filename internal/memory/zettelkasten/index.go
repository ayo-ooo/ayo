package zettelkasten

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"

	"github.com/pressly/goose/v3"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// Index provides SQLite-based indexing for memory search.
// This index is derived from the source markdown files and can be rebuilt.
type Index struct {
	db        *sql.DB
	structure *Structure

	// Prepared statements
	stmtInsert    *sql.Stmt
	stmtUpdate    *sql.Stmt
	stmtDelete    *sql.Stmt
	stmtGet       *sql.Stmt
	stmtSearch    *sql.Stmt
	stmtSearchFTS *sql.Stmt
}

// indexSchema defines the SQLite schema for the memory index.
const indexSchema = `
-- Memory index table (derived from files)
CREATE TABLE IF NOT EXISTS memory_index (
    id TEXT PRIMARY KEY,
    category TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    agent_handle TEXT,
    path_scope TEXT,
    content TEXT NOT NULL,
    embedding BLOB,
    confidence REAL DEFAULT 1.0,
    topics TEXT, -- JSON array
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    last_accessed_at INTEGER,
    access_count INTEGER DEFAULT 0,
    supersedes_id TEXT,
    superseded_by_id TEXT,
    unclear INTEGER DEFAULT 0
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_memory_status ON memory_index(status);
CREATE INDEX IF NOT EXISTS idx_memory_category ON memory_index(category);
CREATE INDEX IF NOT EXISTS idx_memory_agent ON memory_index(agent_handle);
CREATE INDEX IF NOT EXISTS idx_memory_path ON memory_index(path_scope);
CREATE INDEX IF NOT EXISTS idx_memory_created ON memory_index(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_updated ON memory_index(updated_at DESC);

-- Full-text search virtual table
CREATE VIRTUAL TABLE IF NOT EXISTS memory_fts USING fts5(
    id UNINDEXED,
    content,
    topics,
    content='memory_index',
    content_rowid='rowid'
);

-- Triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS memory_ai AFTER INSERT ON memory_index BEGIN
    INSERT INTO memory_fts(rowid, id, content, topics)
    VALUES (new.rowid, new.id, new.content, new.topics);
END;

CREATE TRIGGER IF NOT EXISTS memory_ad AFTER DELETE ON memory_index BEGIN
    INSERT INTO memory_fts(memory_fts, rowid, id, content, topics)
    VALUES ('delete', old.rowid, old.id, old.content, old.topics);
END;

CREATE TRIGGER IF NOT EXISTS memory_au AFTER UPDATE ON memory_index BEGIN
    INSERT INTO memory_fts(memory_fts, rowid, id, content, topics)
    VALUES ('delete', old.rowid, old.id, old.content, old.topics);
    INSERT INTO memory_fts(rowid, id, content, topics)
    VALUES (new.rowid, new.id, new.content, new.topics);
END;

-- Metadata table for tracking index state
CREATE TABLE IF NOT EXISTS index_meta (
    key TEXT PRIMARY KEY,
    value TEXT
);
`

// NewIndex creates a new index for the given structure.
func NewIndex(s *Structure) *Index {
	return &Index{
		structure: s,
	}
}

// Open opens the index database.
func (idx *Index) Open(ctx context.Context) error {
	dbPath := idx.structure.IndexDB

	// Suppress goose logging
	goose.SetLogger(log.New(io.Discard, "", 0))

	db, err := sql.Open("sqlite3", "file:"+dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return fmt.Errorf("set WAL mode: %w", err)
	}

	// Create schema
	if _, err := db.ExecContext(ctx, indexSchema); err != nil {
		db.Close()
		return fmt.Errorf("create schema: %w", err)
	}

	idx.db = db

	// Prepare statements
	if err := idx.prepareStatements(ctx); err != nil {
		db.Close()
		return fmt.Errorf("prepare statements: %w", err)
	}

	return nil
}

// prepareStatements prepares commonly used statements.
func (idx *Index) prepareStatements(ctx context.Context) error {
	var err error

	idx.stmtInsert, err = idx.db.PrepareContext(ctx, `
		INSERT OR REPLACE INTO memory_index (
			id, category, status, agent_handle, path_scope, content, embedding,
			confidence, topics, created_at, updated_at, last_accessed_at,
			access_count, supersedes_id, superseded_by_id, unclear
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}

	idx.stmtUpdate, err = idx.db.PrepareContext(ctx, `
		UPDATE memory_index SET
			category = ?, status = ?, agent_handle = ?, path_scope = ?,
			content = ?, embedding = ?, confidence = ?, topics = ?,
			updated_at = ?, unclear = ?
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare update: %w", err)
	}

	idx.stmtDelete, err = idx.db.PrepareContext(ctx, `DELETE FROM memory_index WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("prepare delete: %w", err)
	}

	idx.stmtGet, err = idx.db.PrepareContext(ctx, `SELECT * FROM memory_index WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("prepare get: %w", err)
	}

	idx.stmtSearchFTS, err = idx.db.PrepareContext(ctx, `
		SELECT m.id, m.category, m.status, m.agent_handle, m.path_scope, 
		       m.content, m.embedding, m.confidence, m.topics,
		       m.created_at, m.updated_at, m.unclear,
		       bm25(memory_fts) as rank
		FROM memory_fts f
		JOIN memory_index m ON f.id = m.id
		WHERE memory_fts MATCH ?
		  AND m.status = ?
		ORDER BY rank
		LIMIT ?
	`)
	if err != nil {
		return fmt.Errorf("prepare search FTS: %w", err)
	}

	return nil
}

// Close closes the index database.
func (idx *Index) Close() error {
	if idx.stmtInsert != nil {
		idx.stmtInsert.Close()
	}
	if idx.stmtUpdate != nil {
		idx.stmtUpdate.Close()
	}
	if idx.stmtDelete != nil {
		idx.stmtDelete.Close()
	}
	if idx.stmtGet != nil {
		idx.stmtGet.Close()
	}
	if idx.stmtSearchFTS != nil {
		idx.stmtSearchFTS.Close()
	}
	if idx.db != nil {
		return idx.db.Close()
	}
	return nil
}

// IndexEntry represents a memory in the index.
type IndexEntry struct {
	ID             string
	Category       string
	Status         string
	AgentHandle    sql.NullString
	PathScope      sql.NullString
	Content        string
	Embedding      []byte
	Confidence     float64
	Topics         sql.NullString // JSON array
	CreatedAt      int64
	UpdatedAt      int64
	LastAccessedAt sql.NullInt64
	AccessCount    int64
	SupersedesID   sql.NullString
	SupersededByID sql.NullString
	Unclear        bool
}

// IndexFromMemoryFile creates an IndexEntry from a MemoryFile.
func IndexFromMemoryFile(mf *MemoryFile) IndexEntry {
	fm := mf.Frontmatter

	entry := IndexEntry{
		ID:         fm.ID,
		Category:   fm.Category,
		Status:     fm.Status,
		Content:    mf.Content,
		Confidence: fm.Confidence,
		CreatedAt:  fm.Created.Unix(),
		UpdatedAt:  fm.Updated.Unix(),
		Unclear:    fm.Unclear.Flagged,
	}

	if fm.Scope.Agent != "" {
		entry.AgentHandle = sql.NullString{String: fm.Scope.Agent, Valid: true}
	}
	if fm.Scope.Path != "" {
		entry.PathScope = sql.NullString{String: fm.Scope.Path, Valid: true}
	}
	if len(fm.Topics) > 0 {
		// Simple JSON array encoding
		entry.Topics = sql.NullString{String: topicsToJSON(fm.Topics), Valid: true}
	}
	if !fm.Access.LastAccessed.IsZero() {
		entry.LastAccessedAt = sql.NullInt64{Int64: fm.Access.LastAccessed.Unix(), Valid: true}
	}
	entry.AccessCount = fm.Access.AccessCount

	if fm.Supersession.Supersedes != "" {
		entry.SupersedesID = sql.NullString{String: fm.Supersession.Supersedes, Valid: true}
	}
	if fm.Supersession.SupersededBy != "" {
		entry.SupersededByID = sql.NullString{String: fm.Supersession.SupersededBy, Valid: true}
	}

	return entry
}

// Insert adds or replaces a memory in the index.
func (idx *Index) Insert(ctx context.Context, entry IndexEntry) error {
	_, err := idx.stmtInsert.ExecContext(ctx,
		entry.ID, entry.Category, entry.Status,
		entry.AgentHandle, entry.PathScope,
		entry.Content, entry.Embedding, entry.Confidence,
		entry.Topics, entry.CreatedAt, entry.UpdatedAt,
		entry.LastAccessedAt, entry.AccessCount,
		entry.SupersedesID, entry.SupersededByID, entry.Unclear,
	)
	return err
}

// Delete removes a memory from the index.
func (idx *Index) Delete(ctx context.Context, id string) error {
	_, err := idx.stmtDelete.ExecContext(ctx, id)
	return err
}

// SearchFTS performs full-text search.
func (idx *Index) SearchFTS(ctx context.Context, query string, status string, limit int) ([]IndexEntry, error) {
	if status == "" {
		status = "active"
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := idx.stmtSearchFTS.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []IndexEntry
	for rows.Next() {
		var e IndexEntry
		var rank float64
		if err := rows.Scan(
			&e.ID, &e.Category, &e.Status, &e.AgentHandle, &e.PathScope,
			&e.Content, &e.Embedding, &e.Confidence, &e.Topics,
			&e.CreatedAt, &e.UpdatedAt, &e.Unclear, &rank,
		); err != nil {
			return nil, err
		}
		results = append(results, e)
	}

	return results, rows.Err()
}

// Rebuild rebuilds the entire index from source files.
func (idx *Index) Rebuild(ctx context.Context) error {
	// Start transaction
	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing data
	if _, err := tx.ExecContext(ctx, "DELETE FROM memory_index"); err != nil {
		return fmt.Errorf("clear index: %w", err)
	}

	// Load all memory files
	files, err := idx.structure.ListAllMemories()
	if err != nil {
		return fmt.Errorf("list memories: %w", err)
	}

	// Insert each memory
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO memory_index (
			id, category, status, agent_handle, path_scope, content, embedding,
			confidence, topics, created_at, updated_at, last_accessed_at,
			access_count, supersedes_id, superseded_by_id, unclear
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, path := range files {
		mf, err := ParseFile(path)
		if err != nil {
			// Skip files that can't be parsed
			continue
		}

		entry := IndexFromMemoryFile(mf)
		if _, err := stmt.ExecContext(ctx,
			entry.ID, entry.Category, entry.Status,
			entry.AgentHandle, entry.PathScope,
			entry.Content, entry.Embedding, entry.Confidence,
			entry.Topics, entry.CreatedAt, entry.UpdatedAt,
			entry.LastAccessedAt, entry.AccessCount,
			entry.SupersedesID, entry.SupersededByID, entry.Unclear,
		); err != nil {
			return fmt.Errorf("insert %s: %w", entry.ID, err)
		}
	}

	return tx.Commit()
}

// Count returns the number of entries in the index.
func (idx *Index) Count(ctx context.Context) (int64, error) {
	var count int64
	err := idx.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM memory_index").Scan(&count)
	return count, err
}

// CountByStatus returns the count of entries by status.
func (idx *Index) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := idx.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM memory_index WHERE status = ?", status,
	).Scan(&count)
	return count, err
}

// UpdateEmbedding updates the embedding for a memory.
func (idx *Index) UpdateEmbedding(ctx context.Context, id string, embedding []byte) error {
	_, err := idx.db.ExecContext(ctx,
		"UPDATE memory_index SET embedding = ? WHERE id = ?",
		embedding, id,
	)
	return err
}

// GetMissingEmbeddings returns IDs of entries without embeddings.
func (idx *Index) GetMissingEmbeddings(ctx context.Context, limit int) ([]string, error) {
	rows, err := idx.db.QueryContext(ctx,
		"SELECT id FROM memory_index WHERE (embedding IS NULL OR length(embedding) = 0) AND status = 'active' LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// topicsToJSON converts a slice of topics to a simple JSON array string.
func topicsToJSON(topics []string) string {
	if len(topics) == 0 {
		return "[]"
	}
	result := "["
	for i, t := range topics {
		if i > 0 {
			result += ","
		}
		result += `"` + t + `"`
	}
	result += "]"
	return result
}
