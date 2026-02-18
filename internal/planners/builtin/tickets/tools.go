package tickets

import (
	"context"
	"encoding/json"
	"fmt"

	"charm.land/fantasy"
	"github.com/alexcabrera/ayo/internal/tickets"
)

// Tool names
const (
	ToolTicketCreate = "ticket_create"
	ToolTicketList   = "ticket_list"
	ToolTicketStart  = "ticket_start"
	ToolTicketClose  = "ticket_close"
	ToolTicketBlock  = "ticket_block"
	ToolTicketNote   = "ticket_note"
)

// Tool descriptions
const (
	descCreate = "Create a new ticket for tracking work"
	descList   = "List tickets with optional filtering by status, type, or assignee"
	descStart  = "Start working on a ticket (changes status to in_progress)"
	descClose  = "Close a completed ticket"
	descBlock  = "Mark a ticket as blocked"
	descNote   = "Add a timestamped note to a ticket"
)

// CreateParams are the parameters for ticket_create.
type CreateParams struct {
	Title       string   `json:"title" jsonschema:"required,description=Short descriptive title for the ticket"`
	Description string   `json:"description,omitempty" jsonschema:"description=Detailed description of the work"`
	Type        string   `json:"type,omitempty" jsonschema:"enum=task,enum=feature,enum=bug,enum=chore,enum=epic,description=Ticket type (default: task)"`
	Priority    int      `json:"priority,omitempty" jsonschema:"minimum=0,maximum=4,description=Priority 0-4 (0=highest\\, default=2)"`
	Assignee    string   `json:"assignee,omitempty" jsonschema:"description=Person assigned to the ticket"`
	Deps        []string `json:"deps,omitempty" jsonschema:"description=IDs of tickets this depends on"`
	Parent      string   `json:"parent,omitempty" jsonschema:"description=Parent ticket ID (for sub-tasks)"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description=Tags for categorization"`
}

// ListParams are the parameters for ticket_list.
type ListParams struct {
	Status   string   `json:"status,omitempty" jsonschema:"enum=open,enum=in_progress,enum=blocked,enum=closed,description=Filter by status"`
	Type     string   `json:"type,omitempty" jsonschema:"enum=task,enum=feature,enum=bug,enum=chore,enum=epic,description=Filter by type"`
	Assignee string   `json:"assignee,omitempty" jsonschema:"description=Filter by assignee"`
	Tags     []string `json:"tags,omitempty" jsonschema:"description=Filter by tags (must have all)"`
}

// TicketIDParams is used for operations that only need a ticket ID.
type TicketIDParams struct {
	ID string `json:"id" jsonschema:"required,description=Ticket ID (full or partial)"`
}

// NoteParams are the parameters for ticket_note.
type NoteParams struct {
	ID      string `json:"id" jsonschema:"required,description=Ticket ID (full or partial)"`
	Content string `json:"content" jsonschema:"required,description=Note content to add"`
}

// CloseParams are the parameters for ticket_close.
type CloseParams struct {
	ID      string `json:"id" jsonschema:"required,description=Ticket ID (full or partial)"`
	Message string `json:"message,omitempty" jsonschema:"description=Optional closing message/summary"`
}

// TicketResult represents a ticket in tool responses.
type TicketResult struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Status   string   `json:"status"`
	Type     string   `json:"type"`
	Priority int      `json:"priority"`
	Assignee string   `json:"assignee,omitempty"`
	Deps     []string `json:"deps,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// newCreateTool creates the ticket_create tool.
func (p *Plugin) newCreateTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ToolTicketCreate,
		descCreate,
		func(ctx context.Context, params CreateParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return p.handleCreate(ctx, params)
		},
	)
}

// newListTool creates the ticket_list tool.
func (p *Plugin) newListTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ToolTicketList,
		descList,
		func(ctx context.Context, params ListParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return p.handleList(ctx, params)
		},
	)
}

// newStartTool creates the ticket_start tool.
func (p *Plugin) newStartTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ToolTicketStart,
		descStart,
		func(ctx context.Context, params TicketIDParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return p.handleStart(ctx, params)
		},
	)
}

// newCloseTool creates the ticket_close tool.
func (p *Plugin) newCloseTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ToolTicketClose,
		descClose,
		func(ctx context.Context, params CloseParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return p.handleClose(ctx, params)
		},
	)
}

// newBlockTool creates the ticket_block tool.
func (p *Plugin) newBlockTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ToolTicketBlock,
		descBlock,
		func(ctx context.Context, params TicketIDParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return p.handleBlock(ctx, params)
		},
	)
}

// newNoteTool creates the ticket_note tool.
func (p *Plugin) newNoteTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ToolTicketNote,
		descNote,
		func(ctx context.Context, params NoteParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return p.handleNote(ctx, params)
		},
	)
}

// handleCreate processes ticket_create invocations.
func (p *Plugin) handleCreate(ctx context.Context, params CreateParams) (fantasy.ToolResponse, error) {
	if p.service == nil {
		return fantasy.NewTextErrorResponse("plugin not initialized"), nil
	}

	if params.Title == "" {
		return fantasy.NewTextErrorResponse("title is required"), nil
	}

	opts := tickets.CreateOptions{
		Title:       params.Title,
		Description: params.Description,
		Type:        tickets.Type(params.Type),
		Priority:    params.Priority,
		Assignee:    params.Assignee,
		Deps:        params.Deps,
		Parent:      params.Parent,
		Tags:        params.Tags,
	}

	// Direct mode service uses empty session ID
	ticket, err := p.service.Create("", opts)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("create ticket: %v", err)), nil
	}

	result := TicketResult{
		ID:       ticket.ID,
		Title:    ticket.Title,
		Status:   string(ticket.Status),
		Type:     string(ticket.Type),
		Priority: ticket.Priority,
		Assignee: ticket.Assignee,
		Deps:     ticket.Deps,
		Tags:     ticket.Tags,
	}

	return jsonResponse(map[string]any{
		"message": fmt.Sprintf("Created ticket %s", ticket.ID),
		"ticket":  result,
	})
}

// handleList processes ticket_list invocations.
func (p *Plugin) handleList(ctx context.Context, params ListParams) (fantasy.ToolResponse, error) {
	if p.service == nil {
		return fantasy.NewTextErrorResponse("plugin not initialized"), nil
	}

	filter := tickets.Filter{
		Status:   tickets.Status(params.Status),
		Type:     tickets.Type(params.Type),
		Assignee: params.Assignee,
		Tags:     params.Tags,
	}

	ticketList, err := p.service.List("", filter)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("list tickets: %v", err)), nil
	}

	results := make([]TicketResult, len(ticketList))
	for i, t := range ticketList {
		results[i] = TicketResult{
			ID:       t.ID,
			Title:    t.Title,
			Status:   string(t.Status),
			Type:     string(t.Type),
			Priority: t.Priority,
			Assignee: t.Assignee,
			Deps:     t.Deps,
			Tags:     t.Tags,
		}
	}

	return jsonResponse(map[string]any{
		"count":   len(results),
		"tickets": results,
	})
}

// handleStart processes ticket_start invocations.
func (p *Plugin) handleStart(ctx context.Context, params TicketIDParams) (fantasy.ToolResponse, error) {
	if p.service == nil {
		return fantasy.NewTextErrorResponse("plugin not initialized"), nil
	}

	if params.ID == "" {
		return fantasy.NewTextErrorResponse("id is required"), nil
	}

	if err := p.service.Start("", params.ID); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("start ticket: %v", err)), nil
	}

	return jsonResponse(map[string]any{
		"message": fmt.Sprintf("Started ticket %s", params.ID),
		"id":      params.ID,
		"status":  "in_progress",
	})
}

// handleClose processes ticket_close invocations.
func (p *Plugin) handleClose(ctx context.Context, params CloseParams) (fantasy.ToolResponse, error) {
	if p.service == nil {
		return fantasy.NewTextErrorResponse("plugin not initialized"), nil
	}

	if params.ID == "" {
		return fantasy.NewTextErrorResponse("id is required"), nil
	}

	// Add closing note if message provided
	if params.Message != "" {
		if err := p.service.AddNote("", params.ID, params.Message); err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("add closing note: %v", err)), nil
		}
	}

	if err := p.service.Close("", params.ID); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("close ticket: %v", err)), nil
	}

	return jsonResponse(map[string]any{
		"message": fmt.Sprintf("Closed ticket %s", params.ID),
		"id":      params.ID,
		"status":  "closed",
	})
}

// handleBlock processes ticket_block invocations.
func (p *Plugin) handleBlock(ctx context.Context, params TicketIDParams) (fantasy.ToolResponse, error) {
	if p.service == nil {
		return fantasy.NewTextErrorResponse("plugin not initialized"), nil
	}

	if params.ID == "" {
		return fantasy.NewTextErrorResponse("id is required"), nil
	}

	if err := p.service.Block("", params.ID); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("block ticket: %v", err)), nil
	}

	return jsonResponse(map[string]any{
		"message": fmt.Sprintf("Blocked ticket %s", params.ID),
		"id":      params.ID,
		"status":  "blocked",
	})
}

// handleNote processes ticket_note invocations.
func (p *Plugin) handleNote(ctx context.Context, params NoteParams) (fantasy.ToolResponse, error) {
	if p.service == nil {
		return fantasy.NewTextErrorResponse("plugin not initialized"), nil
	}

	if params.ID == "" {
		return fantasy.NewTextErrorResponse("id is required"), nil
	}

	if params.Content == "" {
		return fantasy.NewTextErrorResponse("content is required"), nil
	}

	if err := p.service.AddNote("", params.ID, params.Content); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("add note: %v", err)), nil
	}

	return jsonResponse(map[string]any{
		"message": fmt.Sprintf("Added note to ticket %s", params.ID),
		"id":      params.ID,
	})
}

// jsonResponse marshals a value to JSON and returns it as a text response.
func jsonResponse(v any) (fantasy.ToolResponse, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("marshal response: %v", err)), nil
	}
	return fantasy.NewTextResponse(string(data)), nil
}
