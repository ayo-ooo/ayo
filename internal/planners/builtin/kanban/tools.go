package kanban

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/fantasy"
)

// ToolName is the name of the kanban tool.
const ToolName = "kanban"

// ToolDescription describes what the kanban tool does.
const ToolDescription = "Manages a kanban board for visual work tracking with columns and cards."

// KanbanParams are the parameters for the kanban tool.
type KanbanParams struct {
	// Action is the operation to perform: board, add, move, update, remove
	Action string `json:"action" jsonschema:"required,enum=board,enum=add,enum=move,enum=update,enum=remove,description=Action to perform"`

	// CardID is the card ID for move/update/remove actions
	CardID int64 `json:"card_id,omitempty" jsonschema:"description=Card ID for move/update/remove actions"`

	// Title is the card title for add/update actions
	Title string `json:"title,omitempty" jsonschema:"description=Card title for add/update actions"`

	// Description is the card description for add/update actions
	Description string `json:"description,omitempty" jsonschema:"description=Card description for add/update actions"`

	// Column is the target column for add/move actions
	Column string `json:"column,omitempty" jsonschema:"description=Target column (backlog\\, ready\\, in_progress\\, review\\, done)"`

	// Priority is the card priority (higher = more important)
	Priority int `json:"priority,omitempty" jsonschema:"description=Card priority (higher = more important)"`
}

// Card represents a kanban card.
type Card struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Column      string `json:"column"`
	Priority    int    `json:"priority"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// Column represents a kanban column with its cards.
type Column struct {
	Name     string `json:"name"`
	WIPLimit int    `json:"wip_limit"`
	Cards    []Card `json:"cards"`
}

// BoardResult contains the board state.
type BoardResult struct {
	Message string   `json:"message"`
	Columns []Column `json:"columns"`
}

// CardResult contains the result of a card operation.
type CardResult struct {
	Message string `json:"message"`
	Card    *Card  `json:"card,omitempty"`
}

// newKanbanTool creates the kanban tool for this plugin instance.
func (p *Plugin) newKanbanTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ToolName,
		ToolDescription,
		func(ctx context.Context, params KanbanParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return p.handleKanban(ctx, params)
		},
	)
}

// handleKanban processes a kanban tool invocation.
func (p *Plugin) handleKanban(ctx context.Context, params KanbanParams) (fantasy.ToolResponse, error) {
	if p.db == nil {
		return fantasy.NewTextErrorResponse("plugin not initialized"), nil
	}

	switch params.Action {
	case "board":
		return p.handleBoard(ctx)
	case "add":
		return p.handleAdd(ctx, params)
	case "move":
		return p.handleMove(ctx, params)
	case "update":
		return p.handleUpdate(ctx, params)
	case "remove":
		return p.handleRemove(ctx, params)
	default:
		return fantasy.NewTextErrorResponse(fmt.Sprintf("unknown action: %s", params.Action)), nil
	}
}

// handleBoard returns the current board state.
func (p *Plugin) handleBoard(ctx context.Context) (fantasy.ToolResponse, error) {
	// Get columns
	colRows, err := p.db.QueryContext(ctx, "SELECT name, wip_limit FROM columns ORDER BY position")
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("query columns: %v", err)), nil
	}
	defer colRows.Close()

	var columns []Column
	for colRows.Next() {
		var col Column
		if err := colRows.Scan(&col.Name, &col.WIPLimit); err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("scan column: %v", err)), nil
		}
		col.Cards = []Card{}
		columns = append(columns, col)
	}

	// Get cards for each column
	cardRows, err := p.db.QueryContext(ctx, `
		SELECT id, title, description, column_name, priority, created_at, updated_at
		FROM cards ORDER BY priority DESC, position
	`)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("query cards: %v", err)), nil
	}
	defer cardRows.Close()

	// Map cards to columns
	colIndex := make(map[string]int)
	for i, col := range columns {
		colIndex[col.Name] = i
	}

	for cardRows.Next() {
		var card Card
		if err := cardRows.Scan(&card.ID, &card.Title, &card.Description, &card.Column, &card.Priority, &card.CreatedAt, &card.UpdatedAt); err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("scan card: %v", err)), nil
		}
		if idx, ok := colIndex[card.Column]; ok {
			columns[idx].Cards = append(columns[idx].Cards, card)
		}
	}

	result := BoardResult{
		Message: "Kanban board",
		Columns: columns,
	}

	// Build text summary
	var sb strings.Builder
	sb.WriteString("## Kanban Board\n\n")
	for _, col := range columns {
		wipInfo := ""
		if col.WIPLimit > 0 {
			wipInfo = fmt.Sprintf(" [%d/%d]", len(col.Cards), col.WIPLimit)
		}
		sb.WriteString(fmt.Sprintf("### %s%s\n", col.Name, wipInfo))
		if len(col.Cards) == 0 {
			sb.WriteString("  (empty)\n")
		} else {
			for _, card := range col.Cards {
				sb.WriteString(fmt.Sprintf("  - [#%d] %s", card.ID, card.Title))
				if card.Priority > 0 {
					sb.WriteString(fmt.Sprintf(" (P%d)", card.Priority))
				}
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	result.Message = sb.String()

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("marshal result: %v", err)), nil
	}

	return fantasy.NewTextResponse(string(jsonResult)), nil
}

// handleAdd creates a new card.
func (p *Plugin) handleAdd(ctx context.Context, params KanbanParams) (fantasy.ToolResponse, error) {
	if params.Title == "" {
		return fantasy.NewTextErrorResponse("title is required"), nil
	}

	column := params.Column
	if column == "" {
		column = "backlog"
	}

	// Verify column exists
	var exists int
	if err := p.db.QueryRowContext(ctx, "SELECT 1 FROM columns WHERE name = ?", column).Scan(&exists); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid column: %s", column)), nil
	}

	// Get next position
	var maxPos int
	p.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(position), 0) FROM cards WHERE column_name = ?", column).Scan(&maxPos)

	// Insert card
	result, err := p.db.ExecContext(ctx, `
		INSERT INTO cards (title, description, column_name, priority, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, unixepoch(), unixepoch())
	`, params.Title, params.Description, column, params.Priority, maxPos+1)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("insert card: %v", err)), nil
	}

	id, _ := result.LastInsertId()

	cardResult := CardResult{
		Message: fmt.Sprintf("Created card #%d in %s: %s", id, column, params.Title),
		Card: &Card{
			ID:          id,
			Title:       params.Title,
			Description: params.Description,
			Column:      column,
			Priority:    params.Priority,
		},
	}

	jsonResult, err := json.Marshal(cardResult)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("marshal result: %v", err)), nil
	}

	return fantasy.NewTextResponse(string(jsonResult)), nil
}

// handleMove moves a card to a different column.
func (p *Plugin) handleMove(ctx context.Context, params KanbanParams) (fantasy.ToolResponse, error) {
	if params.CardID == 0 {
		return fantasy.NewTextErrorResponse("card_id is required"), nil
	}
	if params.Column == "" {
		return fantasy.NewTextErrorResponse("column is required"), nil
	}

	// Get current card
	var card Card
	err := p.db.QueryRowContext(ctx, `
		SELECT id, title, description, column_name, priority, created_at, updated_at
		FROM cards WHERE id = ?
	`, params.CardID).Scan(&card.ID, &card.Title, &card.Description, &card.Column, &card.Priority, &card.CreatedAt, &card.UpdatedAt)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("card not found: %d", params.CardID)), nil
	}

	// Check WIP limit for target column
	var wipLimit, currentCount int
	err = p.db.QueryRowContext(ctx, "SELECT wip_limit FROM columns WHERE name = ?", params.Column).Scan(&wipLimit)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid column: %s", params.Column)), nil
	}

	if wipLimit > 0 {
		p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cards WHERE column_name = ?", params.Column).Scan(&currentCount)
		if currentCount >= wipLimit {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("WIP limit reached for %s (%d/%d)", params.Column, currentCount, wipLimit)), nil
		}
	}

	// Get next position in target column
	var maxPos int
	p.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(position), 0) FROM cards WHERE column_name = ?", params.Column).Scan(&maxPos)

	// Move card
	_, err = p.db.ExecContext(ctx, `
		UPDATE cards SET column_name = ?, position = ?, updated_at = unixepoch() WHERE id = ?
	`, params.Column, maxPos+1, params.CardID)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("move card: %v", err)), nil
	}

	oldColumn := card.Column
	card.Column = params.Column

	cardResult := CardResult{
		Message: fmt.Sprintf("Moved card #%d from %s to %s: %s", card.ID, oldColumn, params.Column, card.Title),
		Card:    &card,
	}

	jsonResult, err := json.Marshal(cardResult)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("marshal result: %v", err)), nil
	}

	return fantasy.NewTextResponse(string(jsonResult)), nil
}

// handleUpdate updates a card's details.
func (p *Plugin) handleUpdate(ctx context.Context, params KanbanParams) (fantasy.ToolResponse, error) {
	if params.CardID == 0 {
		return fantasy.NewTextErrorResponse("card_id is required"), nil
	}

	// Build update query
	var updates []string
	var args []any

	if params.Title != "" {
		updates = append(updates, "title = ?")
		args = append(args, params.Title)
	}
	if params.Description != "" {
		updates = append(updates, "description = ?")
		args = append(args, params.Description)
	}
	if params.Priority != 0 {
		updates = append(updates, "priority = ?")
		args = append(args, params.Priority)
	}

	if len(updates) == 0 {
		return fantasy.NewTextErrorResponse("nothing to update"), nil
	}

	updates = append(updates, "updated_at = unixepoch()")
	args = append(args, params.CardID)

	query := fmt.Sprintf("UPDATE cards SET %s WHERE id = ?", strings.Join(updates, ", "))
	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("update card: %v", err)), nil
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("card not found: %d", params.CardID)), nil
	}

	// Get updated card
	var card Card
	p.db.QueryRowContext(ctx, `
		SELECT id, title, description, column_name, priority, created_at, updated_at
		FROM cards WHERE id = ?
	`, params.CardID).Scan(&card.ID, &card.Title, &card.Description, &card.Column, &card.Priority, &card.CreatedAt, &card.UpdatedAt)

	cardResult := CardResult{
		Message: fmt.Sprintf("Updated card #%d: %s", card.ID, card.Title),
		Card:    &card,
	}

	jsonResult, err := json.Marshal(cardResult)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("marshal result: %v", err)), nil
	}

	return fantasy.NewTextResponse(string(jsonResult)), nil
}

// handleRemove deletes a card.
func (p *Plugin) handleRemove(ctx context.Context, params KanbanParams) (fantasy.ToolResponse, error) {
	if params.CardID == 0 {
		return fantasy.NewTextErrorResponse("card_id is required"), nil
	}

	// Get card info before deleting
	var card Card
	err := p.db.QueryRowContext(ctx, `
		SELECT id, title, description, column_name, priority, created_at, updated_at
		FROM cards WHERE id = ?
	`, params.CardID).Scan(&card.ID, &card.Title, &card.Description, &card.Column, &card.Priority, &card.CreatedAt, &card.UpdatedAt)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("card not found: %d", params.CardID)), nil
	}

	// Delete card
	_, err = p.db.ExecContext(ctx, "DELETE FROM cards WHERE id = ?", params.CardID)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("delete card: %v", err)), nil
	}

	cardResult := CardResult{
		Message: fmt.Sprintf("Removed card #%d: %s", card.ID, card.Title),
		Card:    &card,
	}

	jsonResult, err := json.Marshal(cardResult)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("marshal result: %v", err)), nil
	}

	return fantasy.NewTextResponse(string(jsonResult)), nil
}

// ListCards returns all cards in the database.
func (p *Plugin) ListCards(ctx context.Context) ([]Card, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := p.db.QueryContext(ctx, `
		SELECT id, title, description, column_name, priority, created_at, updated_at
		FROM cards ORDER BY column_name, priority DESC, position
	`)
	if err != nil {
		return nil, fmt.Errorf("query cards: %w", err)
	}
	defer rows.Close()

	var cards []Card
	for rows.Next() {
		var c Card
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Column, &c.Priority, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		cards = append(cards, c)
	}

	return cards, rows.Err()
}
