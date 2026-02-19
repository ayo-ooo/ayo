package kanban

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNewKanbanTool(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	tool := p.newKanbanTool()

	info := tool.Info()
	if info.Name != ToolName {
		t.Errorf("tool name = %q, want %q", info.Name, ToolName)
	}
	if info.Description != ToolDescription {
		t.Errorf("tool description = %q, want %q", info.Description, ToolDescription)
	}
}

func TestHandleKanban_NotInitialized(t *testing.T) {
	p := &Plugin{}

	params := KanbanParams{Action: "board"}

	resp, err := p.handleKanban(context.Background(), params)
	if err != nil {
		t.Fatalf("handleKanban() failed: %v", err)
	}

	if !resp.IsError {
		t.Error("expected error response for uninitialized plugin")
	}
}

func TestHandleKanban_Board(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	params := KanbanParams{Action: "board"}

	resp, err := p.handleKanban(ctx, params)
	if err != nil {
		t.Fatalf("handleKanban() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	var result BoardResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(result.Columns) != len(DefaultColumns) {
		t.Errorf("got %d columns, want %d", len(result.Columns), len(DefaultColumns))
	}
}

func TestHandleKanban_Add(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	params := KanbanParams{
		Action:      "add",
		Title:       "Test card",
		Description: "Test description",
		Priority:    1,
	}

	resp, err := p.handleKanban(ctx, params)
	if err != nil {
		t.Fatalf("handleKanban() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	var result CardResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Card == nil {
		t.Fatal("expected card in response")
	}
	if result.Card.Title != "Test card" {
		t.Errorf("card title = %q, want %q", result.Card.Title, "Test card")
	}
	if result.Card.Column != "backlog" {
		t.Errorf("card column = %q, want %q", result.Card.Column, "backlog")
	}
}

func TestHandleKanban_AddToSpecificColumn(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	params := KanbanParams{
		Action: "add",
		Title:  "Ready card",
		Column: "ready",
	}

	resp, err := p.handleKanban(ctx, params)
	if err != nil {
		t.Fatalf("handleKanban() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	var result CardResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Card.Column != "ready" {
		t.Errorf("card column = %q, want %q", result.Card.Column, "ready")
	}
}

func TestHandleKanban_Move(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	// Add a card first
	addParams := KanbanParams{
		Action: "add",
		Title:  "Card to move",
	}
	addResp, _ := p.handleKanban(ctx, addParams)
	var addResult CardResult
	json.Unmarshal([]byte(addResp.Content), &addResult)

	// Move the card
	moveParams := KanbanParams{
		Action: "move",
		CardID: addResult.Card.ID,
		Column: "in_progress",
	}

	resp, err := p.handleKanban(ctx, moveParams)
	if err != nil {
		t.Fatalf("handleKanban() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	var result CardResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Card.Column != "in_progress" {
		t.Errorf("card column = %q, want %q", result.Card.Column, "in_progress")
	}
}

func TestHandleKanban_MoveWIPLimit(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	// Add cards up to WIP limit for in_progress (3)
	for i := 0; i < 3; i++ {
		addParams := KanbanParams{
			Action: "add",
			Title:  "Card",
			Column: "in_progress",
		}
		p.handleKanban(ctx, addParams)
	}

	// Add one more in backlog
	addParams := KanbanParams{
		Action: "add",
		Title:  "Card to move",
	}
	addResp, _ := p.handleKanban(ctx, addParams)
	var addResult CardResult
	json.Unmarshal([]byte(addResp.Content), &addResult)

	// Try to move it to in_progress (should fail)
	moveParams := KanbanParams{
		Action: "move",
		CardID: addResult.Card.ID,
		Column: "in_progress",
	}

	resp, err := p.handleKanban(ctx, moveParams)
	if err != nil {
		t.Fatalf("handleKanban() failed: %v", err)
	}

	if !resp.IsError {
		t.Error("expected error response for WIP limit exceeded")
	}
}

func TestHandleKanban_Update(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	// Add a card first
	addParams := KanbanParams{
		Action: "add",
		Title:  "Original title",
	}
	addResp, _ := p.handleKanban(ctx, addParams)
	var addResult CardResult
	json.Unmarshal([]byte(addResp.Content), &addResult)

	// Update the card
	updateParams := KanbanParams{
		Action:      "update",
		CardID:      addResult.Card.ID,
		Title:       "Updated title",
		Description: "New description",
	}

	resp, err := p.handleKanban(ctx, updateParams)
	if err != nil {
		t.Fatalf("handleKanban() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	var result CardResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Card.Title != "Updated title" {
		t.Errorf("card title = %q, want %q", result.Card.Title, "Updated title")
	}
	if result.Card.Description != "New description" {
		t.Errorf("card description = %q, want %q", result.Card.Description, "New description")
	}
}

func TestHandleKanban_Remove(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	// Add a card first
	addParams := KanbanParams{
		Action: "add",
		Title:  "Card to remove",
	}
	addResp, _ := p.handleKanban(ctx, addParams)
	var addResult CardResult
	json.Unmarshal([]byte(addResp.Content), &addResult)

	// Remove the card
	removeParams := KanbanParams{
		Action: "remove",
		CardID: addResult.Card.ID,
	}

	resp, err := p.handleKanban(ctx, removeParams)
	if err != nil {
		t.Fatalf("handleKanban() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	// Verify card was removed
	cards, err := p.ListCards(ctx)
	if err != nil {
		t.Fatalf("ListCards() failed: %v", err)
	}

	for _, card := range cards {
		if card.ID == addResult.Card.ID {
			t.Error("card should have been removed")
		}
	}
}

func TestHandleKanban_UnknownAction(t *testing.T) {
	p := &Plugin{stateDir: t.TempDir()}
	ctx := context.Background()
	if err := p.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Close()

	params := KanbanParams{Action: "unknown"}

	resp, err := p.handleKanban(ctx, params)
	if err != nil {
		t.Fatalf("handleKanban() failed: %v", err)
	}

	if !resp.IsError {
		t.Error("expected error response for unknown action")
	}
}
