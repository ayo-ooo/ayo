package run

import (
	"context"

	"github.com/alexcabrera/ayo/internal/session"
)

// Context keys for tool execution.
type ctxKey string

const (
	sessionIDKey ctxKey = "session_id"
	servicesKey  ctxKey = "services"
)

// WithSessionID adds the session ID to the context.
func WithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}

// GetSessionIDFromContext retrieves the session ID from the context.
func GetSessionIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(sessionIDKey).(string)
	return id
}

// WithServices adds the session services to the context.
func WithServices(ctx context.Context, svc *session.Services) context.Context {
	return context.WithValue(ctx, servicesKey, svc)
}

// GetServicesFromContext retrieves the session services from the context.
func GetServicesFromContext(ctx context.Context) *session.Services {
	svc, _ := ctx.Value(servicesKey).(*session.Services)
	return svc
}
