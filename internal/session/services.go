package session

import (
	"context"
	"database/sql"

	"github.com/alexcabrera/ayo/internal/db"
)

// Services provides access to all session-related services.
type Services struct {
	db       *sql.DB
	queries  *db.Queries
	Sessions *SessionService
	Messages *MessageService
	Edges    *EdgeService
}

// NewServices creates a new Services instance from a database connection.
func NewServices(database *sql.DB, queries *db.Queries) *Services {
	return &Services{
		db:       database,
		queries:  queries,
		Sessions: NewSessionService(queries),
		Messages: NewMessageService(queries),
		Edges:    NewEdgeService(queries),
	}
}

// Close closes the database connection and prepared queries.
func (s *Services) Close() error {
	if err := s.queries.Close(); err != nil {
		return err
	}
	return s.db.Close()
}

// Queries returns the underlying database queries for use by other services.
func (s *Services) Queries() *db.Queries {
	return s.queries
}

// Connect opens a database connection, runs migrations, and returns Services.
func Connect(ctx context.Context, dbPath string) (*Services, error) {
	database, queries, err := db.ConnectWithQueries(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	return NewServices(database, queries), nil
}
