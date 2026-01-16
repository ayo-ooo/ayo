package session

import (
	"context"
	"database/sql"

	"github.com/alexcabrera/ayo/internal/db"
	"github.com/google/uuid"
)

// MessageService provides operations on messages.
type MessageService struct {
	q *db.Queries
}

// NewMessageService creates a new message service.
func NewMessageService(q *db.Queries) *MessageService {
	return &MessageService{q: q}
}

// CreateMessageParams contains parameters for creating a message.
type CreateMessageParams struct {
	SessionID string
	Role      MessageRole
	Parts     []ContentPart
	Model     string
	Provider  string
}

// Create creates a new message.
func (s *MessageService) Create(ctx context.Context, params CreateMessageParams) (Message, error) {
	partsJSON, err := MarshalParts(params.Parts)
	if err != nil {
		return Message{}, err
	}

	dbMsg, err := s.q.CreateMessage(ctx, db.CreateMessageParams{
		ID:        uuid.New().String(),
		SessionID: params.SessionID,
		Role:      string(params.Role),
		Parts:     string(partsJSON),
		Model:     toNullString(params.Model),
		Provider:  toNullString(params.Provider),
	})
	if err != nil {
		return Message{}, err
	}
	return messageFromDB(dbMsg)
}

// Get retrieves a message by ID.
func (s *MessageService) Get(ctx context.Context, id string) (Message, error) {
	dbMsg, err := s.q.GetMessage(ctx, id)
	if err != nil {
		return Message{}, err
	}
	return messageFromDB(dbMsg)
}

// List returns all messages for a session ordered by creation time.
func (s *MessageService) List(ctx context.Context, sessionID string) ([]Message, error) {
	dbMsgs, err := s.q.ListMessagesBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return messagesFromDB(dbMsgs)
}

// Update updates a message's parts and optionally marks it finished.
func (s *MessageService) Update(ctx context.Context, id string, parts []ContentPart, finishedAt int64) error {
	partsJSON, err := MarshalParts(parts)
	if err != nil {
		return err
	}

	var finishedAtNull sql.NullInt64
	if finishedAt > 0 {
		finishedAtNull = sql.NullInt64{Int64: finishedAt, Valid: true}
	}

	return s.q.UpdateMessage(ctx, db.UpdateMessageParams{
		ID:         id,
		Parts:      string(partsJSON),
		FinishedAt: finishedAtNull,
	})
}

// Delete deletes a message.
func (s *MessageService) Delete(ctx context.Context, id string) error {
	return s.q.DeleteMessage(ctx, id)
}

// DeleteBySession deletes all messages for a session.
func (s *MessageService) DeleteBySession(ctx context.Context, sessionID string) error {
	return s.q.DeleteMessagesBySession(ctx, sessionID)
}

// Count returns the number of messages in a session.
func (s *MessageService) Count(ctx context.Context, sessionID string) (int64, error) {
	return s.q.CountMessagesBySession(ctx, sessionID)
}

func messageFromDB(d db.Message) (Message, error) {
	parts, err := UnmarshalParts([]byte(d.Parts))
	if err != nil {
		return Message{}, err
	}

	return Message{
		ID:         d.ID,
		SessionID:  d.SessionID,
		Role:       MessageRole(d.Role),
		Parts:      parts,
		Model:      d.Model.String,
		Provider:   d.Provider.String,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
		FinishedAt: d.FinishedAt.Int64,
	}, nil
}

func messagesFromDB(ds []db.Message) ([]Message, error) {
	messages := make([]Message, len(ds))
	for i, d := range ds {
		m, err := messageFromDB(d)
		if err != nil {
			return nil, err
		}
		messages[i] = m
	}
	return messages, nil
}
