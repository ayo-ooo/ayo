package session

import (
	"context"
	"database/sql"
	"time"

	"github.com/alexcabrera/ayo/internal/db"
	"github.com/google/uuid"
)

// Session represents a conversation session.
type Session struct {
	ID               string
	AgentHandle      string
	Title            string
	InputSchema      string
	OutputSchema     string
	StructuredInput  string
	StructuredOutput string
	ChainDepth       int64
	ChainSource      string
	MessageCount     int64
	CreatedAt        int64
	UpdatedAt        int64
	FinishedAt       int64
}

// SessionService provides operations on sessions.
type SessionService struct {
	q *db.Queries
}

// NewSessionService creates a new session service.
func NewSessionService(q *db.Queries) *SessionService {
	return &SessionService{q: q}
}

// CreateParams contains parameters for creating a session.
type CreateParams struct {
	AgentHandle      string
	Title            string
	InputSchema      string
	OutputSchema     string
	StructuredInput  string
	ChainDepth       int64
	ChainSource      string
}

// Create creates a new session.
func (s *SessionService) Create(ctx context.Context, params CreateParams) (Session, error) {
	title := params.Title
	if title == "" {
		title = "Untitled Session"
	}

	dbSession, err := s.q.CreateSession(ctx, db.CreateSessionParams{
		ID:              uuid.New().String(),
		AgentHandle:     params.AgentHandle,
		Title:           title,
		InputSchema:     toNullString(params.InputSchema),
		OutputSchema:    toNullString(params.OutputSchema),
		StructuredInput: toNullString(params.StructuredInput),
		ChainDepth:      params.ChainDepth,
		ChainSource:     toNullString(params.ChainSource),
	})
	if err != nil {
		return Session{}, err
	}
	return sessionFromDB(dbSession), nil
}

// Get retrieves a session by ID.
func (s *SessionService) Get(ctx context.Context, id string) (Session, error) {
	dbSession, err := s.q.GetSession(ctx, id)
	if err != nil {
		return Session{}, err
	}
	return sessionFromDB(dbSession), nil
}

// GetByPrefix finds sessions matching an ID prefix.
func (s *SessionService) GetByPrefix(ctx context.Context, prefix string) ([]Session, error) {
	dbSessions, err := s.q.GetSessionByPrefix(ctx, toNullString(prefix))
	if err != nil {
		return nil, err
	}
	return sessionsFromDB(dbSessions), nil
}

// List returns sessions ordered by most recent first.
func (s *SessionService) List(ctx context.Context, limit int64) ([]Session, error) {
	if limit <= 0 {
		limit = 50
	}
	dbSessions, err := s.q.ListSessions(ctx, limit)
	if err != nil {
		return nil, err
	}
	return sessionsFromDB(dbSessions), nil
}

// ListByAgent returns sessions for a specific agent.
func (s *SessionService) ListByAgent(ctx context.Context, agentHandle string, limit int64) ([]Session, error) {
	if limit <= 0 {
		limit = 50
	}
	dbSessions, err := s.q.ListSessionsByAgent(ctx, db.ListSessionsByAgentParams{
		AgentHandle: agentHandle,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}
	return sessionsFromDB(dbSessions), nil
}

// Search finds sessions by title substring.
func (s *SessionService) Search(ctx context.Context, query string, limit int64) ([]Session, error) {
	if limit <= 0 {
		limit = 50
	}
	dbSessions, err := s.q.SearchSessionsByTitle(ctx, db.SearchSessionsByTitleParams{
		Query: toNullString(query),
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	return sessionsFromDB(dbSessions), nil
}

// UpdateTitle updates a session's title.
func (s *SessionService) UpdateTitle(ctx context.Context, id, title string) error {
	return s.q.UpdateSessionTitle(ctx, db.UpdateSessionTitleParams{
		ID:    id,
		Title: title,
	})
}

// Finish marks a session as finished.
func (s *SessionService) Finish(ctx context.Context, id string, structuredOutput string) (Session, error) {
	now := time.Now().Unix()
	dbSession, err := s.q.UpdateSession(ctx, db.UpdateSessionParams{
		ID:               id,
		Title:            "", // Will be preserved by the query
		StructuredOutput: toNullString(structuredOutput),
		FinishedAt:       sql.NullInt64{Int64: now, Valid: true},
	})
	if err != nil {
		return Session{}, err
	}
	return sessionFromDB(dbSession), nil
}

// Delete deletes a session and its messages.
func (s *SessionService) Delete(ctx context.Context, id string) error {
	return s.q.DeleteSession(ctx, id)
}

// Count returns the total number of sessions.
func (s *SessionService) Count(ctx context.Context) (int64, error) {
	return s.q.CountSessions(ctx)
}

// CountByAgent returns the number of sessions for an agent.
func (s *SessionService) CountByAgent(ctx context.Context, agentHandle string) (int64, error) {
	return s.q.CountSessionsByAgent(ctx, agentHandle)
}

func sessionFromDB(d db.Session) Session {
	return Session{
		ID:               d.ID,
		AgentHandle:      d.AgentHandle,
		Title:            d.Title,
		InputSchema:      d.InputSchema.String,
		OutputSchema:     d.OutputSchema.String,
		StructuredInput:  d.StructuredInput.String,
		StructuredOutput: d.StructuredOutput.String,
		ChainDepth:       d.ChainDepth,
		ChainSource:      d.ChainSource.String,
		MessageCount:     d.MessageCount,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
		FinishedAt:       d.FinishedAt.Int64,
	}
}

func sessionsFromDB(ds []db.Session) []Session {
	sessions := make([]Session, len(ds))
	for i, d := range ds {
		sessions[i] = sessionFromDB(d)
	}
	return sessions
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
