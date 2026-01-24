package session

import (
	"context"
	"database/sql"
	"time"

	"github.com/alexcabrera/ayo/internal/db"
	"github.com/google/uuid"
)

// Session source constants.
const (
	SourceAyo         = "ayo"           // Session created by ayo
	SourceCrush       = "crush"         // Session created by standalone Crush
	SourceCrushViaAyo = "crush-via-ayo" // Session created by Crush invoked through ayo
)

// Session represents a conversation session.
type Session struct {
	ID               string
	AgentHandle      string
	Title            string
	Source           string // Session source: ayo, crush, crush-via-ayo
	InputSchema      string
	OutputSchema     string
	StructuredInput  string
	StructuredOutput string
	ChainDepth       int64
	ChainSource      string
	MessageCount     int64
	Plan             Plan
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
	Source           string // Defaults to SourceAyo if empty
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

	source := params.Source
	if source == "" {
		source = SourceAyo
	}

	dbSession, err := s.q.CreateSession(ctx, db.CreateSessionParams{
		ID:              uuid.New().String(),
		AgentHandle:     params.AgentHandle,
		Title:           title,
		Source:          source,
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

// ListBySource returns sessions for a specific source.
func (s *SessionService) ListBySource(ctx context.Context, source string, limit int64) ([]Session, error) {
	if limit <= 0 {
		limit = 50
	}
	dbSessions, err := s.q.ListSessionsBySource(ctx, db.ListSessionsBySourceParams{
		Source: source,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	return sessionsFromDB(dbSessions), nil
}

// CountBySource returns the number of sessions for a source.
func (s *SessionService) CountBySource(ctx context.Context, source string) (int64, error) {
	return s.q.CountSessionsBySource(ctx, source)
}

// UpdatePlan updates a session's plan.
func (s *SessionService) UpdatePlan(ctx context.Context, id string, plan Plan) (Session, error) {
	planJSON, err := marshalPlan(plan)
	if err != nil {
		return Session{}, err
	}
	dbSession, err := s.q.UpdateSessionPlan(ctx, db.UpdateSessionPlanParams{
		ID:   id,
		Plan: toNullString(planJSON),
	})
	if err != nil {
		return Session{}, err
	}
	return sessionFromDB(dbSession), nil
}

func sessionFromDB(d db.Session) Session {
	plan, _ := unmarshalPlan(d.Plan.String)
	return Session{
		ID:               d.ID,
		AgentHandle:      d.AgentHandle,
		Title:            d.Title,
		Source:           d.Source,
		InputSchema:      d.InputSchema.String,
		OutputSchema:     d.OutputSchema.String,
		StructuredInput:  d.StructuredInput.String,
		StructuredOutput: d.StructuredOutput.String,
		ChainDepth:       d.ChainDepth,
		ChainSource:      d.ChainSource.String,
		MessageCount:     d.MessageCount,
		Plan:             plan,
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
