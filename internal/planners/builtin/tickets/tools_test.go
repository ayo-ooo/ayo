package tickets

import (
	"context"
	"testing"

	"github.com/alexcabrera/ayo/internal/planners"
	"github.com/alexcabrera/ayo/internal/tickets"
)

// setupTestPlugin creates a plugin with a real service for testing.
func setupTestPlugin(t *testing.T) *Plugin {
	t.Helper()
	tempDir := t.TempDir()

	factory := New()
	ctx := planners.PlannerContext{
		SandboxName: "test-sandbox",
		SandboxDir:  "/sandbox",
		StateDir:    tempDir,
	}

	plugin, err := factory(ctx)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	return plugin.(*Plugin)
}

func TestHandleCreate(t *testing.T) {
	p := setupTestPlugin(t)

	tests := []struct {
		name      string
		params    CreateParams
		wantError bool
	}{
		{
			name:      "missing title",
			params:    CreateParams{},
			wantError: true,
		},
		{
			name: "minimal ticket",
			params: CreateParams{
				Title: "Test ticket",
			},
			wantError: false,
		},
		{
			name: "full ticket",
			params: CreateParams{
				Title:       "Full ticket",
				Description: "A detailed description",
				Type:        "feature",
				Priority:    1,
				Assignee:    "alice",
				Tags:        []string{"backend", "urgent"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := p.handleCreate(context.Background(), tt.params)
			if err != nil {
				t.Fatalf("handleCreate returned error: %v", err)
			}

			if tt.wantError {
				if !resp.IsError {
					t.Error("expected error response")
				}
			} else {
				if resp.IsError {
					t.Errorf("unexpected error: %s", resp.Content)
				}
			}
		})
	}
}

func TestHandleCreate_Uninitialized(t *testing.T) {
	p := &Plugin{} // No service
	resp, err := p.handleCreate(context.Background(), CreateParams{Title: "Test"})
	if err != nil {
		t.Fatalf("handleCreate returned error: %v", err)
	}
	if !resp.IsError {
		t.Error("expected error response for uninitialized plugin")
	}
}

func TestHandleList(t *testing.T) {
	p := setupTestPlugin(t)

	// Create some tickets first
	_, _ = p.handleCreate(context.Background(), CreateParams{Title: "Ticket 1", Type: "task"})
	_, _ = p.handleCreate(context.Background(), CreateParams{Title: "Ticket 2", Type: "bug"})
	_, _ = p.handleCreate(context.Background(), CreateParams{Title: "Ticket 3", Type: "task", Assignee: "alice"})

	tests := []struct {
		name   string
		params ListParams
	}{
		{
			name:   "list all",
			params: ListParams{},
		},
		{
			name:   "filter by type",
			params: ListParams{Type: "task"},
		},
		{
			name:   "filter by assignee",
			params: ListParams{Assignee: "alice"},
		},
		{
			name:   "filter by status",
			params: ListParams{Status: "open"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := p.handleList(context.Background(), tt.params)
			if err != nil {
				t.Fatalf("handleList returned error: %v", err)
			}
			if resp.IsError {
				t.Errorf("unexpected error: %s", resp.Content)
			}
		})
	}
}

func TestHandleList_Uninitialized(t *testing.T) {
	p := &Plugin{}
	resp, err := p.handleList(context.Background(), ListParams{})
	if err != nil {
		t.Fatalf("handleList returned error: %v", err)
	}
	if !resp.IsError {
		t.Error("expected error response for uninitialized plugin")
	}
}

func TestHandleStart(t *testing.T) {
	p := setupTestPlugin(t)

	// Start with empty ID should fail
	resp, err := p.handleStart(context.Background(), TicketIDParams{ID: ""})
	if err != nil {
		t.Fatalf("handleStart returned error: %v", err)
	}
	if !resp.IsError {
		t.Error("expected error for empty ID")
	}
}

func TestHandleStart_Uninitialized(t *testing.T) {
	p := &Plugin{}
	resp, err := p.handleStart(context.Background(), TicketIDParams{ID: "test"})
	if err != nil {
		t.Fatalf("handleStart returned error: %v", err)
	}
	if !resp.IsError {
		t.Error("expected error response for uninitialized plugin")
	}
}

func TestHandleClose(t *testing.T) {
	p := setupTestPlugin(t)

	// Close with empty ID should fail
	resp, err := p.handleClose(context.Background(), CloseParams{ID: ""})
	if err != nil {
		t.Fatalf("handleClose returned error: %v", err)
	}
	if !resp.IsError {
		t.Error("expected error for empty ID")
	}
}

func TestHandleClose_Uninitialized(t *testing.T) {
	p := &Plugin{}
	resp, err := p.handleClose(context.Background(), CloseParams{ID: "test"})
	if err != nil {
		t.Fatalf("handleClose returned error: %v", err)
	}
	if !resp.IsError {
		t.Error("expected error response for uninitialized plugin")
	}
}

func TestHandleBlock(t *testing.T) {
	p := setupTestPlugin(t)

	// Block with empty ID should fail
	resp, err := p.handleBlock(context.Background(), TicketIDParams{ID: ""})
	if err != nil {
		t.Fatalf("handleBlock returned error: %v", err)
	}
	if !resp.IsError {
		t.Error("expected error for empty ID")
	}
}

func TestHandleBlock_Uninitialized(t *testing.T) {
	p := &Plugin{}
	resp, err := p.handleBlock(context.Background(), TicketIDParams{ID: "test"})
	if err != nil {
		t.Fatalf("handleBlock returned error: %v", err)
	}
	if !resp.IsError {
		t.Error("expected error response for uninitialized plugin")
	}
}

func TestHandleNote(t *testing.T) {
	p := setupTestPlugin(t)

	tests := []struct {
		name      string
		params    NoteParams
		wantError bool
	}{
		{
			name:      "empty ID",
			params:    NoteParams{ID: "", Content: "note"},
			wantError: true,
		},
		{
			name:      "empty content",
			params:    NoteParams{ID: "test", Content: ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := p.handleNote(context.Background(), tt.params)
			if err != nil {
				t.Fatalf("handleNote returned error: %v", err)
			}
			if tt.wantError && !resp.IsError {
				t.Error("expected error response")
			}
		})
	}
}

func TestHandleNote_Uninitialized(t *testing.T) {
	p := &Plugin{}
	resp, err := p.handleNote(context.Background(), NoteParams{ID: "test", Content: "note"})
	if err != nil {
		t.Fatalf("handleNote returned error: %v", err)
	}
	if !resp.IsError {
		t.Error("expected error response for uninitialized plugin")
	}
}

func TestTicketWorkflow(t *testing.T) {
	p := setupTestPlugin(t)

	// Create a ticket
	createResp, err := p.handleCreate(context.Background(), CreateParams{
		Title:       "Workflow test",
		Description: "Test the full workflow",
		Type:        "task",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if createResp.IsError {
		t.Fatalf("create returned error: %s", createResp.Content)
	}

	// List to get tickets
	listResp, err := p.handleList(context.Background(), ListParams{Status: "open"})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listResp.IsError {
		t.Fatalf("list returned error: %s", listResp.Content)
	}

	// Get the ticket from the service directly to get its ID
	ticketList, err := p.service.List("", tickets.Filter{})
	if err != nil {
		t.Fatalf("service list failed: %v", err)
	}
	if len(ticketList) == 0 {
		t.Fatal("no tickets found")
	}

	ticketID := ticketList[0].ID

	// Start the ticket
	startResp, err := p.handleStart(context.Background(), TicketIDParams{ID: ticketID})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if startResp.IsError {
		t.Fatalf("start returned error: %s", startResp.Content)
	}

	// Add a note
	noteResp, err := p.handleNote(context.Background(), NoteParams{ID: ticketID, Content: "Working on it"})
	if err != nil {
		t.Fatalf("note failed: %v", err)
	}
	if noteResp.IsError {
		t.Fatalf("note returned error: %s", noteResp.Content)
	}

	// Close with message
	closeResp, err := p.handleClose(context.Background(), CloseParams{ID: ticketID, Message: "Done!"})
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if closeResp.IsError {
		t.Fatalf("close returned error: %s", closeResp.Content)
	}

	// Verify ticket is closed
	ticket, err := p.service.Get("", ticketID)
	if err != nil {
		t.Fatalf("get ticket failed: %v", err)
	}
	if string(ticket.Status) != "closed" {
		t.Errorf("ticket status = %q, want %q", ticket.Status, "closed")
	}
}

func TestJsonResponse(t *testing.T) {
	resp, err := jsonResponse(map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("jsonResponse returned error: %v", err)
	}
	if resp.IsError {
		t.Error("expected non-error response")
	}
	if resp.Content != `{"key":"value"}` {
		t.Errorf("unexpected response content: %s", resp.Content)
	}
}
