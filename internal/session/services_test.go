package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexcabrera/ayo/internal/db"
)

func setupTestDB(t *testing.T) (*Services, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "ayo-session-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	ctx := context.Background()

	services, err := Connect(ctx, dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to connect: %v", err)
	}

	cleanup := func() {
		services.Close()
		os.RemoveAll(tmpDir)
	}

	return services, cleanup
}

func TestSessionServiceCreate(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	session, err := svc.Sessions.Create(ctx, CreateParams{
		AgentHandle: "@ayo",
		Title:       "Test Session",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if session.ID == "" {
		t.Error("expected non-empty ID")
	}
	if session.AgentHandle != "@ayo" {
		t.Errorf("AgentHandle = %q, want %q", session.AgentHandle, "@ayo")
	}
	if session.Title != "Test Session" {
		t.Errorf("Title = %q, want %q", session.Title, "Test Session")
	}
	if session.CreatedAt == 0 {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestSessionServiceCreateDefaultTitle(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	session, err := svc.Sessions.Create(ctx, CreateParams{
		AgentHandle: "@ayo",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if session.Title != "Untitled Session" {
		t.Errorf("Title = %q, want %q", session.Title, "Untitled Session")
	}
}

func TestSessionServiceGet(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, _ := svc.Sessions.Create(ctx, CreateParams{
		AgentHandle: "@ayo",
		Title:       "Test",
	})

	got, err := svc.Sessions.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestSessionServiceGetNotFound(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := svc.Sessions.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestSessionServiceGetByPrefix(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})

	// Search by first 8 chars of UUID
	prefix := created.ID[:8]
	sessions, err := svc.Sessions.GetByPrefix(ctx, prefix)
	if err != nil {
		t.Fatalf("GetByPrefix failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("got %d sessions, want 1", len(sessions))
	}
	if sessions[0].ID != created.ID {
		t.Errorf("ID = %q, want %q", sessions[0].ID, created.ID)
	}
}

func TestSessionServiceList(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	}

	sessions, err := svc.Sessions.List(ctx, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("got %d sessions, want 3", len(sessions))
	}
}

func TestSessionServiceListByAgent(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@other"})

	sessions, err := svc.Sessions.ListByAgent(ctx, "@ayo", 10)
	if err != nil {
		t.Fatalf("ListByAgent failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("got %d sessions, want 2", len(sessions))
	}
}

func TestSessionServiceSearch(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo", Title: "Fix authentication"})
	svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo", Title: "Fix login problem"})
	svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo", Title: "Add new feature"})

	sessions, err := svc.Sessions.Search(ctx, "login", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Errorf("got %d sessions, want 1", len(sessions))
	}
}

func TestSessionServiceUpdateTitle(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, _ := svc.Sessions.Create(ctx, CreateParams{
		AgentHandle: "@ayo",
		Title:       "Original",
	})

	err := svc.Sessions.UpdateTitle(ctx, created.ID, "Updated Title")
	if err != nil {
		t.Fatalf("UpdateTitle failed: %v", err)
	}

	got, _ := svc.Sessions.Get(ctx, created.ID)
	if got.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", got.Title, "Updated Title")
	}
}

func TestSessionServiceDelete(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})

	err := svc.Sessions.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = svc.Sessions.Get(ctx, created.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestSessionServiceCount(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	count, _ := svc.Sessions.Count(ctx)
	if count != 0 {
		t.Errorf("initial count = %d, want 0", count)
	}

	svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})

	count, _ = svc.Sessions.Count(ctx)
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

// MessageService tests

func TestMessageServiceCreate(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	session, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})

	msg, err := svc.Messages.Create(ctx, CreateMessageParams{
		SessionID: session.ID,
		Role:      RoleUser,
		Parts:     []ContentPart{TextContent{Text: "Hello"}},
		Model:     "gpt-4",
		Provider:  "openai",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if msg.ID == "" {
		t.Error("expected non-empty ID")
	}
	if msg.Role != RoleUser {
		t.Errorf("Role = %v, want %v", msg.Role, RoleUser)
	}
	if msg.TextContent() != "Hello" {
		t.Errorf("TextContent = %q, want %q", msg.TextContent(), "Hello")
	}
}

func TestMessageServiceCreateUpdatesSessionCount(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	session, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})

	svc.Messages.Create(ctx, CreateMessageParams{
		SessionID: session.ID,
		Role:      RoleUser,
		Parts:     []ContentPart{TextContent{Text: "Hello"}},
	})

	// Refresh session to check message count
	updated, _ := svc.Sessions.Get(ctx, session.ID)
	if updated.MessageCount != 1 {
		t.Errorf("MessageCount = %d, want 1", updated.MessageCount)
	}
}

func TestMessageServiceList(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	session, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})

	svc.Messages.Create(ctx, CreateMessageParams{
		SessionID: session.ID,
		Role:      RoleUser,
		Parts:     []ContentPart{TextContent{Text: "Hello"}},
	})
	svc.Messages.Create(ctx, CreateMessageParams{
		SessionID: session.ID,
		Role:      RoleAssistant,
		Parts:     []ContentPart{TextContent{Text: "Hi there!"}},
	})

	msgs, err := svc.Messages.List(ctx, session.ID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2", len(msgs))
	}

	// Should be ordered by creation time
	if msgs[0].Role != RoleUser || msgs[1].Role != RoleAssistant {
		t.Error("messages not in expected order")
	}
}

func TestMessageServiceUpdate(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	session, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	msg, _ := svc.Messages.Create(ctx, CreateMessageParams{
		SessionID: session.ID,
		Role:      RoleAssistant,
		Parts:     []ContentPart{TextContent{Text: "Hello"}},
	})

	// Update with more content
	newParts := []ContentPart{
		TextContent{Text: "Hello, updated!"},
		Finish{Reason: FinishReasonStop, Time: 123},
	}
	err := svc.Messages.Update(ctx, msg.ID, newParts, 123)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, _ := svc.Messages.Get(ctx, msg.ID)
	if updated.TextContent() != "Hello, updated!" {
		t.Errorf("TextContent = %q, want %q", updated.TextContent(), "Hello, updated!")
	}
	if !updated.IsFinished() {
		t.Error("expected message to be finished")
	}
}

func TestMessageServiceDelete(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	session, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	msg, _ := svc.Messages.Create(ctx, CreateMessageParams{
		SessionID: session.ID,
		Role:      RoleUser,
		Parts:     []ContentPart{TextContent{Text: "Hello"}},
	})

	err := svc.Messages.Delete(ctx, msg.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = svc.Messages.Get(ctx, msg.ID)
	if err == nil {
		t.Error("expected error after delete")
	}

	// Check session message count was decremented
	updated, _ := svc.Sessions.Get(ctx, session.ID)
	if updated.MessageCount != 0 {
		t.Errorf("MessageCount = %d, want 0", updated.MessageCount)
	}
}

func TestMessageServiceCascadeDelete(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	session, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	msg, _ := svc.Messages.Create(ctx, CreateMessageParams{
		SessionID: session.ID,
		Role:      RoleUser,
		Parts:     []ContentPart{TextContent{Text: "Hello"}},
	})

	// Delete session should cascade to messages
	svc.Sessions.Delete(ctx, session.ID)

	_, err := svc.Messages.Get(ctx, msg.ID)
	if err == nil {
		t.Error("message should be deleted with session")
	}
}

// EdgeService tests

func TestEdgeServiceCreate(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	parent, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	child, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo.research"})

	err := svc.Edges.Create(ctx, parent.ID, child.ID, EdgeTypeAgentCall, "")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	children, _ := svc.Edges.GetChildren(ctx, parent.ID)
	if len(children) != 1 {
		t.Fatalf("got %d children, want 1", len(children))
	}
	if children[0].ChildID != child.ID {
		t.Errorf("ChildID = %q, want %q", children[0].ChildID, child.ID)
	}
	if children[0].EdgeType != EdgeTypeAgentCall {
		t.Errorf("EdgeType = %v, want %v", children[0].EdgeType, EdgeTypeAgentCall)
	}
}

func TestEdgeServiceGetParents(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	parent1, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	parent2, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@other"})
	child, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo.research"})

	svc.Edges.Create(ctx, parent1.ID, child.ID, EdgeTypeAgentCall, "")
	svc.Edges.Create(ctx, parent2.ID, child.ID, EdgeTypeChain, "")

	parents, err := svc.Edges.GetParents(ctx, child.ID)
	if err != nil {
		t.Fatalf("GetParents failed: %v", err)
	}

	if len(parents) != 2 {
		t.Errorf("got %d parents, want 2", len(parents))
	}
}

func TestEdgeServiceDelete(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	parent, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	child, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo.research"})

	svc.Edges.Create(ctx, parent.ID, child.ID, EdgeTypeAgentCall, "")
	svc.Edges.Delete(ctx, parent.ID, child.ID)

	children, _ := svc.Edges.GetChildren(ctx, parent.ID)
	if len(children) != 0 {
		t.Errorf("got %d children after delete, want 0", len(children))
	}
}

func TestEdgeServiceCascadeDelete(t *testing.T) {
	svc, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	parent, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo"})
	child, _ := svc.Sessions.Create(ctx, CreateParams{AgentHandle: "@ayo.research"})

	svc.Edges.Create(ctx, parent.ID, child.ID, EdgeTypeAgentCall, "")

	// Delete parent session should cascade to edges
	svc.Sessions.Delete(ctx, parent.ID)

	// Child session should still exist
	_, err := svc.Sessions.Get(ctx, child.ID)
	if err != nil {
		t.Error("child session should still exist")
	}

	// But edge should be gone
	parents, _ := svc.Edges.GetParents(ctx, child.ID)
	if len(parents) != 0 {
		t.Errorf("got %d parents after cascade delete, want 0", len(parents))
	}
}

// Services integration tests

func TestServicesConnect(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-session-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ctx := context.Background()

	svc, err := Connect(ctx, dbPath)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer svc.Close()

	// Verify all services are initialized
	if svc.Sessions == nil {
		t.Error("Sessions service not initialized")
	}
	if svc.Messages == nil {
		t.Error("Messages service not initialized")
	}
	if svc.Edges == nil {
		t.Error("Edges service not initialized")
	}
}

// Suppress unused import warning from db package
var _ = db.Querier(nil)
