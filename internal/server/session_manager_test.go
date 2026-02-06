package server

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/config"
)

func newTestSessionManager() *SessionManager {
	return NewSessionManager(SessionManagerConfig{
		Logger:      slog.Default(),
		Config:      config.Config{},
		IdleTimeout: 0, // No auto-stop in tests
	})
}

func TestSessionManager_StartStop(t *testing.T) {
	mgr := newTestSessionManager()
	ctx := context.Background()

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := mgr.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestSessionManager_List_Empty(t *testing.T) {
	mgr := newTestSessionManager()

	sessions := mgr.List()
	if len(sessions) != 0 {
		t.Errorf("expected empty list, got %d sessions", len(sessions))
	}
}

func TestSessionManager_Get_NotFound(t *testing.T) {
	mgr := newTestSessionManager()

	_, err := mgr.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestSessionManager_GetByAgent_NotFound(t *testing.T) {
	mgr := newTestSessionManager()

	_, err := mgr.GetByAgent("@test")
	if err == nil {
		t.Error("expected error for nonexistent agent session")
	}
}

func TestSessionManager_Sleep_NotFound(t *testing.T) {
	mgr := newTestSessionManager()
	ctx := context.Background()

	err := mgr.Sleep(ctx, "@test")
	if err == nil {
		t.Error("expected error when no session to sleep")
	}
}

func TestSessionManager_StopSession_NotFound(t *testing.T) {
	mgr := newTestSessionManager()
	ctx := context.Background()

	err := mgr.StopSession(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestSessionManager_Inject_NotFound(t *testing.T) {
	mgr := newTestSessionManager()
	ctx := context.Background()

	_, err := mgr.Inject(ctx, "nonexistent", "hello")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestSessionManager_HandleNormalization(t *testing.T) {
	mgr := newTestSessionManager()

	// GetByAgent should normalize handles
	_, err := mgr.GetByAgent("test") // No @ prefix
	if err == nil {
		t.Error("expected error for nonexistent agent session")
	}

	// Error message should have normalized handle
	if err.Error() != "no active session for agent: @test" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSessionManager_CleanupStopped(t *testing.T) {
	mgr := newTestSessionManager()

	// Manually add a stopped session
	mgr.mu.Lock()
	mgr.sessions["test-id"] = &managedSession{
		AgentSession: AgentSession{
			ID:          "test-id",
			AgentHandle: "@test",
			Status:      SessionStatusStopped,
		},
	}
	mgr.mu.Unlock()

	// Verify it's there
	mgr.mu.RLock()
	if _, ok := mgr.sessions["test-id"]; !ok {
		mgr.mu.RUnlock()
		t.Fatal("session should exist before cleanup")
	}
	mgr.mu.RUnlock()

	// Cleanup
	mgr.CleanupStopped()

	// Verify it's gone
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	if _, ok := mgr.sessions["test-id"]; ok {
		t.Error("stopped session should be cleaned up")
	}
}

func TestSessionManager_MarkActive(t *testing.T) {
	mgr := newTestSessionManager()

	initialTime := time.Now().Add(-time.Hour)

	// Manually add a session with old LastActive time
	mgr.mu.Lock()
	mgr.sessions["test-id"] = &managedSession{
		AgentSession: AgentSession{
			ID:          "test-id",
			AgentHandle: "@test",
			Status:      SessionStatusIdle,
			LastActive:  initialTime,
		},
	}
	mgr.mu.Unlock()

	// Mark active
	mgr.MarkActive("test-id")

	// Check LastActive was updated
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	sess := mgr.sessions["test-id"]
	if sess.LastActive.Before(time.Now().Add(-time.Second)) {
		t.Error("LastActive should be updated to recent time")
	}
}

func TestSessionManager_IdleTimeout(t *testing.T) {
	// Create manager with short idle timeout
	mgr := NewSessionManager(SessionManagerConfig{
		Logger:      slog.Default(),
		Config:      config.Config{},
		IdleTimeout: 50 * time.Millisecond,
	})

	ctx := context.Background()
	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer mgr.Stop(ctx)

	// Manually add an idle session with old LastActive
	mgr.mu.Lock()
	mgr.sessions["test-id"] = &managedSession{
		AgentSession: AgentSession{
			ID:          "test-id",
			AgentHandle: "@test",
			Status:      SessionStatusIdle,
			LastActive:  time.Now().Add(-time.Hour), // Very old
		},
	}
	mgr.mu.Unlock()

	// Wait for idle check to run (IdleTimeout / 2 interval)
	time.Sleep(100 * time.Millisecond)

	// Check session is stopped
	mgr.mu.RLock()
	sess := mgr.sessions["test-id"]
	mgr.mu.RUnlock()

	if sess.Status != SessionStatusStopped {
		t.Errorf("expected session to be stopped due to idle timeout, got %s", sess.Status)
	}
}

func TestSessionManager_GenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	if id1 == "" {
		t.Error("session ID should not be empty")
	}

	// IDs should be unique (different timestamps)
	if id1 == id2 {
		t.Error("session IDs should be unique")
	}

	// Should have proper prefix
	if len(id1) < 5 || id1[:5] != "sess_" {
		t.Errorf("session ID should have 'sess_' prefix, got: %s", id1)
	}
}
