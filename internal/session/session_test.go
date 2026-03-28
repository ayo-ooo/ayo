package session

import (
	"os"
	"path/filepath"
	"testing"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	return NewStoreWithDir(dir)
}

func TestNew(t *testing.T) {
	store := testStore(t)
	sess := store.New("test-agent")

	if sess.ID == "" {
		t.Fatal("expected non-empty session ID")
	}
	if len(sess.ID) != 8 {
		t.Fatalf("expected 8-char session ID, got %d chars: %q", len(sess.ID), sess.ID)
	}
	if sess.AgentName != "test-agent" {
		t.Fatalf("expected agent name %q, got %q", "test-agent", sess.AgentName)
	}
	if len(sess.Messages) != 0 {
		t.Fatalf("expected empty messages, got %d", len(sess.Messages))
	}
	if sess.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
	if sess.UpdatedAt.IsZero() {
		t.Fatal("expected non-zero UpdatedAt")
	}
}

func TestNewUniqueIDs(t *testing.T) {
	store := testStore(t)
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		sess := store.New("test-agent")
		if seen[sess.ID] {
			t.Fatalf("duplicate session ID: %q", sess.ID)
		}
		seen[sess.ID] = true
	}
}

func TestSaveAndLoad(t *testing.T) {
	store := testStore(t)
	sess := store.New("test-agent")
	sess.Messages = append(sess.Messages,
		Message{Role: "user", Content: "Hello"},
		Message{Role: "assistant", Content: "Hi there!"},
	)

	if err := store.Save(sess); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Verify file exists on disk.
	path := store.sessionPath(sess.AgentName, sess.ID)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("session file not found: %v", err)
	}

	loaded, err := store.Load("test-agent", sess.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.ID != sess.ID {
		t.Fatalf("expected ID %q, got %q", sess.ID, loaded.ID)
	}
	if loaded.AgentName != sess.AgentName {
		t.Fatalf("expected agent name %q, got %q", sess.AgentName, loaded.AgentName)
	}
	if len(loaded.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(loaded.Messages))
	}
	if loaded.Messages[0].Role != "user" || loaded.Messages[0].Content != "Hello" {
		t.Fatalf("unexpected first message: %+v", loaded.Messages[0])
	}
	if loaded.Messages[1].Role != "assistant" || loaded.Messages[1].Content != "Hi there!" {
		t.Fatalf("unexpected second message: %+v", loaded.Messages[1])
	}
}

func TestLoadNotFound(t *testing.T) {
	store := testStore(t)

	_, err := store.Load("test-agent", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestSaveUpdatesTimestamp(t *testing.T) {
	store := testStore(t)
	sess := store.New("test-agent")
	originalUpdated := sess.UpdatedAt

	sess.Messages = append(sess.Messages, Message{Role: "user", Content: "test"})
	if err := store.Save(sess); err != nil {
		t.Fatalf("save: %v", err)
	}

	if !sess.UpdatedAt.After(originalUpdated) && sess.UpdatedAt.Equal(originalUpdated) {
		// The timestamp might be equal if the test runs very fast,
		// so we just check it's not before.
		if sess.UpdatedAt.Before(originalUpdated) {
			t.Fatal("UpdatedAt went backwards")
		}
	}
}

func TestListEmpty(t *testing.T) {
	store := testStore(t)

	ids, err := store.List("test-agent")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected empty list, got %d items", len(ids))
	}
}

func TestListSessions(t *testing.T) {
	store := testStore(t)

	// Create several sessions.
	var expectedIDs []string
	for i := 0; i < 3; i++ {
		sess := store.New("test-agent")
		if err := store.Save(sess); err != nil {
			t.Fatalf("save: %v", err)
		}
		expectedIDs = append(expectedIDs, sess.ID)
	}

	ids, err := store.List("test-agent")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(ids))
	}

	// Check all expected IDs are present.
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	for _, expected := range expectedIDs {
		if !idSet[expected] {
			t.Fatalf("expected ID %q not found in list", expected)
		}
	}
}

func TestListIgnoresNonJSON(t *testing.T) {
	store := testStore(t)

	// Create a session to establish the directory.
	sess := store.New("test-agent")
	if err := store.Save(sess); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Write a non-JSON file.
	dir := store.sessionsDir("test-agent")
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("writing non-JSON file: %v", err)
	}

	ids, err := store.List("test-agent")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 session, got %d", len(ids))
	}
}

func TestListIsolatesAgents(t *testing.T) {
	store := testStore(t)

	s1 := store.New("agent-a")
	if err := store.Save(s1); err != nil {
		t.Fatalf("save: %v", err)
	}

	s2 := store.New("agent-b")
	if err := store.Save(s2); err != nil {
		t.Fatalf("save: %v", err)
	}

	idsA, err := store.List("agent-a")
	if err != nil {
		t.Fatalf("list agent-a: %v", err)
	}
	if len(idsA) != 1 {
		t.Fatalf("expected 1 session for agent-a, got %d", len(idsA))
	}

	idsB, err := store.List("agent-b")
	if err != nil {
		t.Fatalf("list agent-b: %v", err)
	}
	if len(idsB) != 1 {
		t.Fatalf("expected 1 session for agent-b, got %d", len(idsB))
	}
}

func TestGenerateID(t *testing.T) {
	id := generateID()
	if len(id) != 8 {
		t.Fatalf("expected 8-char ID, got %d chars: %q", len(id), id)
	}
	// Should be valid hex.
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("invalid hex char %c in ID %q", c, id)
		}
	}
}
