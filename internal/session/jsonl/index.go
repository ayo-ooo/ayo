package jsonl

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// Index provides SQLite-based indexing for JSONL session files.
// The index is derived from session headers and can be fully rebuilt from files.
type Index struct {
	db        *sql.DB
	structure *Structure
}

// OpenIndex opens or creates the session index database.
func OpenIndex(structure *Structure) (*Index, error) {
	if err := structure.Initialize(); err != nil {
		return nil, fmt.Errorf("initialize structure: %w", err)
	}

	db, err := sql.Open("sqlite3", structure.IndexDB)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	idx := &Index{db: db, structure: structure}
	if err := idx.ensureSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure schema: %w", err)
	}

	return idx, nil
}

// Close closes the index database.
func (i *Index) Close() error {
	return i.db.Close()
}

func (i *Index) ensureSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS session_index (
		id TEXT PRIMARY KEY,
		agent_handle TEXT NOT NULL,
		title TEXT NOT NULL,
		source TEXT NOT NULL,
		file_path TEXT NOT NULL,
		message_count INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		finished_at INTEGER,
		chain_depth INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_session_agent ON session_index(agent_handle);
	CREATE INDEX IF NOT EXISTS idx_session_created ON session_index(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_session_source ON session_index(source);
	`
	_, err := i.db.Exec(schema)
	return err
}

// Rebuild drops and recreates the index from all session files.
func (i *Index) Rebuild() (*IndexResult, error) {
	result := &IndexResult{}

	// Clear existing index
	if _, err := i.db.Exec("DELETE FROM session_index"); err != nil {
		return nil, fmt.Errorf("clear index: %w", err)
	}

	// Walk all session files
	paths, err := i.structure.ListAllSessions()
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	for _, path := range paths {
		header, err := ReadSessionHeader(path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", path, err))
			continue
		}

		if err := i.indexSession(header, path); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", path, err))
			continue
		}

		result.Indexed++
	}

	return result, nil
}

// IndexSession adds or updates a session in the index.
func (i *Index) IndexSession(header *SessionHeader, path string) error {
	return i.indexSession(header, path)
}

func (i *Index) indexSession(header *SessionHeader, path string) error {
	var finishedAt *int64
	if header.FinishedAt != nil {
		ts := header.FinishedAt.Unix()
		finishedAt = &ts
	}

	_, err := i.db.Exec(`
		INSERT OR REPLACE INTO session_index 
		(id, agent_handle, title, source, file_path, message_count, created_at, updated_at, finished_at, chain_depth)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		header.ID,
		header.AgentHandle,
		header.Title,
		header.Source,
		path,
		header.MessageCount,
		header.CreatedAt.Unix(),
		header.UpdatedAt.Unix(),
		finishedAt,
		header.ChainDepth,
	)
	return err
}

// RemoveSession removes a session from the index.
func (i *Index) RemoveSession(sessionID string) error {
	_, err := i.db.Exec("DELETE FROM session_index WHERE id = ?", sessionID)
	return err
}

// Count returns the total number of indexed sessions.
func (i *Index) Count() (int, error) {
	var count int
	err := i.db.QueryRow("SELECT COUNT(*) FROM session_index").Scan(&count)
	return count, err
}

// IndexResult contains the result of an indexing operation.
type IndexResult struct {
	Indexed int
	Errors  []string
}

// SessionIndexEntry represents a session in the index.
type SessionIndexEntry struct {
	ID           string
	AgentHandle  string
	Title        string
	Source       string
	FilePath     string
	MessageCount int
	CreatedAt    int64
	UpdatedAt    int64
	FinishedAt   *int64
	ChainDepth   int
}

// List returns sessions matching the given filters, ordered by creation time (newest first).
func (i *Index) List(agentHandle, source string, limit int) ([]SessionIndexEntry, error) {
	query := "SELECT id, agent_handle, title, source, file_path, message_count, created_at, updated_at, finished_at, chain_depth FROM session_index WHERE 1=1"
	var args []any

	if agentHandle != "" {
		query += " AND agent_handle = ?"
		args = append(args, agentHandle)
	}
	if source != "" {
		query += " AND source = ?"
		args = append(args, source)
	}

	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := i.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []SessionIndexEntry
	for rows.Next() {
		var e SessionIndexEntry
		if err := rows.Scan(&e.ID, &e.AgentHandle, &e.Title, &e.Source, &e.FilePath, &e.MessageCount, &e.CreatedAt, &e.UpdatedAt, &e.FinishedAt, &e.ChainDepth); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// Search searches sessions by title.
func (i *Index) Search(query string, limit int) ([]SessionIndexEntry, error) {
	sqlQuery := "SELECT id, agent_handle, title, source, file_path, message_count, created_at, updated_at, finished_at, chain_depth FROM session_index WHERE title LIKE ? ORDER BY created_at DESC"
	args := []any{"%" + query + "%"}

	if limit > 0 {
		sqlQuery += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := i.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []SessionIndexEntry
	for rows.Next() {
		var e SessionIndexEntry
		if err := rows.Scan(&e.ID, &e.AgentHandle, &e.Title, &e.Source, &e.FilePath, &e.MessageCount, &e.CreatedAt, &e.UpdatedAt, &e.FinishedAt, &e.ChainDepth); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// Get retrieves a session by ID.
func (i *Index) Get(sessionID string) (*SessionIndexEntry, error) {
	var e SessionIndexEntry
	err := i.db.QueryRow(`
		SELECT id, agent_handle, title, source, file_path, message_count, created_at, updated_at, finished_at, chain_depth 
		FROM session_index WHERE id = ?`, sessionID).Scan(
		&e.ID, &e.AgentHandle, &e.Title, &e.Source, &e.FilePath, &e.MessageCount, &e.CreatedAt, &e.UpdatedAt, &e.FinishedAt, &e.ChainDepth,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// GetByPrefix finds sessions matching an ID prefix.
func (i *Index) GetByPrefix(prefix string) ([]SessionIndexEntry, error) {
	rows, err := i.db.Query(`
		SELECT id, agent_handle, title, source, file_path, message_count, created_at, updated_at, finished_at, chain_depth 
		FROM session_index WHERE id LIKE ? ORDER BY created_at DESC`, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []SessionIndexEntry
	for rows.Next() {
		var e SessionIndexEntry
		if err := rows.Scan(&e.ID, &e.AgentHandle, &e.Title, &e.Source, &e.FilePath, &e.MessageCount, &e.CreatedAt, &e.UpdatedAt, &e.FinishedAt, &e.ChainDepth); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// RemoveOrphanedEntries removes index entries for sessions that no longer exist as files.
func (i *Index) RemoveOrphanedEntries() (int, error) {
	rows, err := i.db.Query("SELECT id, file_path FROM session_index")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var orphans []string
	for rows.Next() {
		var id, path string
		if err := rows.Scan(&id, &path); err != nil {
			continue
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			orphans = append(orphans, id)
		}
	}

	if len(orphans) == 0 {
		return 0, nil
	}

	for _, id := range orphans {
		if _, err := i.db.Exec("DELETE FROM session_index WHERE id = ?", id); err != nil {
			return 0, err
		}
	}

	return len(orphans), nil
}

// DeleteIndex removes the index database file.
func DeleteIndex(structure *Structure) error {
	return os.Remove(structure.IndexDB)
}

// IndexExists returns true if the index database exists.
func IndexExists(structure *Structure) bool {
	_, err := os.Stat(structure.IndexDB)
	return err == nil
}

// EnsureIndex opens the index, creating and rebuilding if necessary.
func EnsureIndex(structure *Structure) (*Index, error) {
	idx, err := OpenIndex(structure)
	if err != nil {
		return nil, err
	}

	// Check if index is empty
	count, err := idx.Count()
	if err != nil {
		idx.Close()
		return nil, err
	}

	// Rebuild if empty
	if count == 0 {
		if _, err := idx.Rebuild(); err != nil {
			idx.Close()
			return nil, err
		}
	}

	return idx, nil
}

// SessionFileFromEntry reads the full session from an index entry.
func SessionFileFromEntry(entry *SessionIndexEntry) (*Reader, error) {
	return NewReader(entry.FilePath)
}

// Sync checks for new files and indexes them, removes orphaned entries.
func (i *Index) Sync() (*SyncResult, error) {
	result := &SyncResult{}

	// Get indexed sessions
	indexed := make(map[string]string) // id -> path
	rows, err := i.db.Query("SELECT id, file_path FROM session_index")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, path string
		rows.Scan(&id, &path)
		indexed[id] = path
	}
	rows.Close()

	// Scan for files
	paths, err := i.structure.ListAllSessions()
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		// Extract session ID from filename
		base := filepath.Base(path)
		sessionID := base[:len(base)-6] // remove .jsonl

		if existingPath, ok := indexed[sessionID]; ok {
			// Already indexed, check if path changed
			if existingPath != path {
				// Update path
				if _, err := i.db.Exec("UPDATE session_index SET file_path = ? WHERE id = ?", path, sessionID); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("update %s: %v", sessionID, err))
				}
			}
			delete(indexed, sessionID)
			continue
		}

		// New file, index it
		header, err := ReadSessionHeader(path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("read %s: %v", path, err))
			continue
		}

		if err := i.indexSession(header, path); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("index %s: %v", path, err))
			continue
		}

		result.Added++
	}

	// Remove orphaned entries (remaining in indexed map)
	for id := range indexed {
		if _, err := i.db.Exec("DELETE FROM session_index WHERE id = ?", id); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("remove orphan %s: %v", id, err))
			continue
		}
		result.Removed++
	}

	return result, nil
}

// SyncResult contains the result of a sync operation.
type SyncResult struct {
	Added   int
	Removed int
	Errors  []string
}
