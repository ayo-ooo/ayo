package tickets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServiceCreate(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	ticket, err := svc.Create("session1", CreateOptions{
		Title:       "Test ticket",
		Description: "Test description",
		Type:        TypeTask,
		Priority:    1,
		Assignee:    "@coder",
		Tags:        []string{"test"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if ticket.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if ticket.Title != "Test ticket" {
		t.Errorf("Title = %q, want %q", ticket.Title, "Test ticket")
	}
	if ticket.Status != StatusOpen {
		t.Errorf("Status = %q, want %q", ticket.Status, StatusOpen)
	}

	// Verify file exists
	if _, err := os.Stat(ticket.FilePath); err != nil {
		t.Errorf("Ticket file not created: %v", err)
	}
}

func TestServiceGet(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	created, err := svc.Create("session1", CreateOptions{
		Title: "Test ticket",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get by full ID
	ticket, err := svc.Get("session1", created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if ticket.ID != created.ID {
		t.Errorf("ID = %q, want %q", ticket.ID, created.ID)
	}

	// Get by partial ID (last 4 chars)
	partialID := created.ID[len(created.ID)-4:]
	ticket2, err := svc.Get("session1", partialID)
	if err != nil {
		t.Fatalf("Get by partial ID failed: %v", err)
	}
	if ticket2.ID != created.ID {
		t.Errorf("Partial match ID = %q, want %q", ticket2.ID, created.ID)
	}
}

func TestServiceStatusTransitions(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	ticket, _ := svc.Create("session1", CreateOptions{Title: "Test"})

	// Start
	if err := svc.Start("session1", ticket.ID); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	ticket, _ = svc.Get("session1", ticket.ID)
	if ticket.Status != StatusInProgress {
		t.Errorf("After Start: Status = %q, want %q", ticket.Status, StatusInProgress)
	}
	if ticket.Started == nil {
		t.Error("Started timestamp should be set")
	}

	// Close
	if err := svc.Close("session1", ticket.ID); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	ticket, _ = svc.Get("session1", ticket.ID)
	if ticket.Status != StatusClosed {
		t.Errorf("After Close: Status = %q, want %q", ticket.Status, StatusClosed)
	}
	if ticket.Closed == nil {
		t.Error("Closed timestamp should be set")
	}

	// Reopen
	if err := svc.Reopen("session1", ticket.ID); err != nil {
		t.Fatalf("Reopen failed: %v", err)
	}
	ticket, _ = svc.Get("session1", ticket.ID)
	if ticket.Status != StatusOpen {
		t.Errorf("After Reopen: Status = %q, want %q", ticket.Status, StatusOpen)
	}

	// Block
	if err := svc.Block("session1", ticket.ID); err != nil {
		t.Fatalf("Block failed: %v", err)
	}
	ticket, _ = svc.Get("session1", ticket.ID)
	if ticket.Status != StatusBlocked {
		t.Errorf("After Block: Status = %q, want %q", ticket.Status, StatusBlocked)
	}
}

func TestServiceList(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	// Create multiple tickets
	svc.Create("session1", CreateOptions{Title: "Task 1", Type: TypeTask, Assignee: "@coder"})
	svc.Create("session1", CreateOptions{Title: "Task 2", Type: TypeTask, Assignee: "@reviewer"})
	svc.Create("session1", CreateOptions{Title: "Bug 1", Type: TypeBug, Assignee: "@coder"})

	// List all
	all, err := svc.List("session1", Filter{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("List all: got %d, want 3", len(all))
	}

	// Filter by type
	tasks, _ := svc.List("session1", Filter{Type: TypeTask})
	if len(tasks) != 2 {
		t.Errorf("List tasks: got %d, want 2", len(tasks))
	}

	// Filter by assignee
	coderTasks, _ := svc.List("session1", Filter{Assignee: "@coder"})
	if len(coderTasks) != 2 {
		t.Errorf("List @coder: got %d, want 2", len(coderTasks))
	}
}

func TestServiceReadyAndBlocked(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	// Create tickets with dependencies
	t1, _ := svc.Create("session1", CreateOptions{Title: "Task 1", Assignee: "@coder"})
	_, _ = svc.Create("session1", CreateOptions{Title: "Task 2", Assignee: "@coder", Deps: []string{t1.ID}})
	_, _ = svc.Create("session1", CreateOptions{Title: "Task 3", Assignee: "@coder"})

	// Check ready - t1 and t3 should be ready
	ready, err := svc.Ready("session1", "@coder")
	if err != nil {
		t.Fatalf("Ready failed: %v", err)
	}
	if len(ready) != 2 {
		t.Errorf("Ready: got %d, want 2", len(ready))
	}

	// Check blocked - t2 should be blocked
	blocked, err := svc.Blocked("session1", "@coder")
	if err != nil {
		t.Fatalf("Blocked failed: %v", err)
	}
	if len(blocked) != 1 {
		t.Errorf("Blocked: got %d, want 1", len(blocked))
	}

	// Close t1 - t2 should become ready
	svc.Close("session1", t1.ID)
	ready2, _ := svc.Ready("session1", "@coder")
	blocked2, _ := svc.Blocked("session1", "@coder")

	// t2 and t3 should be ready (t1 is closed)
	if len(ready2) != 2 {
		t.Errorf("Ready after close: got %d, want 2", len(ready2))
	}
	if len(blocked2) != 0 {
		t.Errorf("Blocked after close: got %d, want 0", len(blocked2))
	}
}

func TestServiceDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	t1, _ := svc.Create("session1", CreateOptions{Title: "Task 1"})
	t2, _ := svc.Create("session1", CreateOptions{Title: "Task 2"})

	// Add dependency
	if err := svc.AddDep("session1", t2.ID, t1.ID); err != nil {
		t.Fatalf("AddDep failed: %v", err)
	}

	t2, _ = svc.Get("session1", t2.ID)
	if len(t2.Deps) != 1 || t2.Deps[0] != t1.ID {
		t.Errorf("Deps = %v, want [%s]", t2.Deps, t1.ID)
	}

	// Remove dependency
	if err := svc.RemoveDep("session1", t2.ID, t1.ID); err != nil {
		t.Fatalf("RemoveDep failed: %v", err)
	}

	t2, _ = svc.Get("session1", t2.ID)
	if len(t2.Deps) != 0 {
		t.Errorf("Deps after remove = %v, want []", t2.Deps)
	}
}

func TestServiceCycleDetection(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	t1, _ := svc.Create("session1", CreateOptions{Title: "Task 1"})
	t2, _ := svc.Create("session1", CreateOptions{Title: "Task 2"})
	t3, _ := svc.Create("session1", CreateOptions{Title: "Task 3"})

	// Create chain: t3 -> t2 -> t1
	svc.AddDep("session1", t2.ID, t1.ID)
	svc.AddDep("session1", t3.ID, t2.ID)

	// Try to create cycle: t1 -> t3 (would create t1 -> t3 -> t2 -> t1)
	err := svc.AddDep("session1", t1.ID, t3.ID)
	if err == nil {
		t.Error("Expected cycle detection error")
	}
}

func TestServiceNotes(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	ticket, _ := svc.Create("session1", CreateOptions{Title: "Test"})

	// Add notes
	svc.AddNote("session1", ticket.ID, "First note")
	svc.AddNote("session1", ticket.ID, "Second note")

	ticket, _ = svc.Get("session1", ticket.ID)
	if len(ticket.Notes) != 2 {
		t.Errorf("Notes count = %d, want 2", len(ticket.Notes))
	}
	if ticket.Notes[0].Content != "First note" {
		t.Errorf("Note 0 content = %q, want %q", ticket.Notes[0].Content, "First note")
	}
}

func TestServiceAssignment(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	ticket, _ := svc.Create("session1", CreateOptions{Title: "Test"})

	// Assign
	if err := svc.Assign("session1", ticket.ID, "@coder"); err != nil {
		t.Fatalf("Assign failed: %v", err)
	}
	ticket, _ = svc.Get("session1", ticket.ID)
	if ticket.Assignee != "@coder" {
		t.Errorf("Assignee = %q, want %q", ticket.Assignee, "@coder")
	}

	// Unassign
	if err := svc.Unassign("session1", ticket.ID); err != nil {
		t.Fatalf("Unassign failed: %v", err)
	}
	ticket, _ = svc.Get("session1", ticket.ID)
	if ticket.Assignee != "" {
		t.Errorf("Assignee after unassign = %q, want empty", ticket.Assignee)
	}
}

func TestServiceDelete(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	ticket, _ := svc.Create("session1", CreateOptions{Title: "Test"})
	path := ticket.FilePath

	// Verify exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Ticket file should exist")
	}

	// Delete
	if err := svc.Delete("session1", ticket.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Ticket file should be deleted")
	}
}

func TestServiceDepTree(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	t1, _ := svc.Create("session1", CreateOptions{Title: "Root"})
	t2, _ := svc.Create("session1", CreateOptions{Title: "Child 1"})
	t3, _ := svc.Create("session1", CreateOptions{Title: "Child 2"})

	svc.AddDep("session1", t1.ID, t2.ID)
	svc.AddDep("session1", t1.ID, t3.ID)

	tree, err := svc.DepTree("session1", t1.ID)
	if err != nil {
		t.Fatalf("DepTree failed: %v", err)
	}

	if tree.Ticket.ID != t1.ID {
		t.Errorf("Root ID = %q, want %q", tree.Ticket.ID, t1.ID)
	}
	if len(tree.Children) != 2 {
		t.Errorf("Children count = %d, want 2", len(tree.Children))
	}
}

func TestEmptySessionList(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewService(tmpDir)

	// List from non-existent session should return empty, not error
	tickets, err := svc.List("nonexistent", Filter{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tickets) != 0 {
		t.Errorf("Expected empty list, got %d", len(tickets))
	}
}

func TestServiceTicketsDir(t *testing.T) {
	svc := NewService("/base/dir")
	expected := filepath.Join("/base/dir", "session1", ".tickets")
	if got := svc.ticketsDir("session1"); got != expected {
		t.Errorf("ticketsDir = %q, want %q", got, expected)
	}
}
