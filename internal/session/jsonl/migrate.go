package jsonl

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/session"
)

// MigrateResult contains the result of a migration operation.
type MigrateResult struct {
	SessionsMigrated int
	SessionsSkipped  int
	MessagesMigrated int
	Errors           []string
}

// MigrateFromSQLite migrates sessions from SQLite database to JSONL files.
// It reads all sessions and their messages from the database and writes
// them to JSONL files in the session directory structure.
func MigrateFromSQLite(ctx context.Context, q *db.Queries, structure *Structure, overwrite bool) (*MigrateResult, error) {
	// Initialize structure
	if err := structure.Initialize(); err != nil {
		return nil, fmt.Errorf("initialize structure: %w", err)
	}

	result := &MigrateResult{}

	// Create session and message services
	sessionSvc := session.NewSessionService(q)
	messageSvc := session.NewMessageService(q)

	// Get all sessions (use a large limit)
	sessions, err := sessionSvc.List(ctx, 100000)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	for _, sess := range sessions {
		// Check if file already exists
		createdAt := time.Unix(sess.CreatedAt, 0)
		path := structure.SessionPath(sess.AgentHandle, sess.ID, createdAt)

		if !overwrite {
			if _, err := os.Stat(path); err == nil {
				result.SessionsSkipped++
				continue
			}
		}

		// Get messages for this session
		messages, err := messageSvc.List(ctx, sess.ID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("session %s: list messages: %v", sess.ID, err))
			continue
		}

		// Create writer
		writer, err := NewWriter(structure, sess)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("session %s: create writer: %v", sess.ID, err))
			continue
		}

		// Write all messages
		for _, msg := range messages {
			if err := writer.WriteMessage(msg); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("session %s, message %s: write: %v", sess.ID, msg.ID, err))
			} else {
				result.MessagesMigrated++
			}
		}

		// Finish the session if it was finished
		if sess.FinishedAt > 0 {
			if err := writer.Finish(sess.StructuredOutput); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("session %s: finish: %v", sess.ID, err))
			}
		}

		if err := writer.Close(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("session %s: close: %v", sess.ID, err))
		}

		result.SessionsMigrated++
	}

	return result, nil
}

// MigrateSession migrates a single session from SQLite to JSONL.
func MigrateSession(ctx context.Context, q *db.Queries, structure *Structure, sessionID string, overwrite bool) error {
	sessionSvc := session.NewSessionService(q)
	messageSvc := session.NewMessageService(q)

	// Get session
	sess, err := sessionSvc.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	// Check if file already exists
	createdAt := time.Unix(sess.CreatedAt, 0)
	path := structure.SessionPath(sess.AgentHandle, sess.ID, createdAt)

	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("session file already exists: %s", path)
		}
	}

	// Get messages
	messages, err := messageSvc.List(ctx, sess.ID)
	if err != nil {
		return fmt.Errorf("list messages: %w", err)
	}

	// Create writer
	writer, err := NewWriter(structure, sess)
	if err != nil {
		return fmt.Errorf("create writer: %w", err)
	}
	defer writer.Close()

	// Write messages
	for _, msg := range messages {
		if err := writer.WriteMessage(msg); err != nil {
			return fmt.Errorf("write message %s: %w", msg.ID, err)
		}
	}

	// Finish if needed
	if sess.FinishedAt > 0 {
		if err := writer.Finish(sess.StructuredOutput); err != nil {
			return fmt.Errorf("finish: %w", err)
		}
	}

	return writer.Close()
}
