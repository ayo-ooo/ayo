package run

import (
	"context"
	"testing"

	"github.com/alexcabrera/ayo/internal/session"
)

func TestWithSessionID(t *testing.T) {
	ctx := context.Background()
	sessionID := "test-session-123"

	ctx = WithSessionID(ctx, sessionID)
	got := GetSessionIDFromContext(ctx)

	if got != sessionID {
		t.Errorf("GetSessionIDFromContext() = %q, want %q", got, sessionID)
	}
}

func TestGetSessionIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	got := GetSessionIDFromContext(ctx)

	if got != "" {
		t.Errorf("GetSessionIDFromContext() = %q, want empty string", got)
	}
}

func TestWithServices(t *testing.T) {
	ctx := context.Background()

	// Use nil services for basic test - just testing context storage
	var svc *session.Services
	ctx = WithServices(ctx, svc)
	got := GetServicesFromContext(ctx)

	if got != svc {
		t.Error("GetServicesFromContext() did not return expected services")
	}
}

func TestGetServicesFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	got := GetServicesFromContext(ctx)

	if got != nil {
		t.Error("GetServicesFromContext() should return nil for empty context")
	}
}

func TestContextChaining(t *testing.T) {
	ctx := context.Background()
	sessionID := "session-456"

	// Add both values
	ctx = WithSessionID(ctx, sessionID)
	ctx = WithServices(ctx, nil)

	// Both should be retrievable
	gotID := GetSessionIDFromContext(ctx)
	if gotID != sessionID {
		t.Errorf("GetSessionIDFromContext() = %q, want %q", gotID, sessionID)
	}

	gotSvc := GetServicesFromContext(ctx)
	if gotSvc != nil {
		t.Error("GetServicesFromContext() should return nil")
	}
}
