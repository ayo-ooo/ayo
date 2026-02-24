// Package memory provides tools for agents to manage Zettelkasten memory notes.
package memory

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/memory/zettelkasten"
)

// NoteCreateParams are the parameters for the memory_note_create tool.
type NoteCreateParams struct {
	// Title is the note title.
	Title string `json:"title" jsonschema:"required,description=Title of the memory note"`

	// Content is the note content.
	Content string `json:"content" jsonschema:"required,description=Content of the memory note"`

	// Category is the type of memory.
	Category string `json:"category" jsonschema:"required,enum=preference,enum=fact,enum=correction,enum=pattern,description=Category of the memory"`

	// Tags are optional tags for the note.
	Tags []string `json:"tags,omitempty" jsonschema:"description=Tags for organizing the note"`

	// Links are IDs of related notes to link to.
	Links []string `json:"links,omitempty" jsonschema:"description=IDs of related notes to link to"`

	// Scope controls where the memory is accessible.
	Scope string `json:"scope,omitempty" jsonschema:"enum=global,enum=agent,enum=path,description=Scope of the memory (default: global)"`
}

// NoteCreateResult contains the result of a note create operation.
type NoteCreateResult struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Links   []string `json:"links,omitempty"`
	Message string   `json:"message"`
}

func (r NoteCreateResult) String() string {
	return r.Message
}

// NoteLinkParams are the parameters for the memory_note_link tool.
type NoteLinkParams struct {
	// FromID is the source note ID.
	FromID string `json:"from_id" jsonschema:"required,description=ID of the source note"`

	// ToID is the target note ID.
	ToID string `json:"to_id" jsonschema:"required,description=ID of the target note"`

	// Relationship describes the type of link.
	Relationship string `json:"relationship,omitempty" jsonschema:"description=Type of relationship (e.g., 'supersedes', 'relates_to', 'contradicts')"`
}

// NoteLinkResult contains the result of a note link operation.
type NoteLinkResult struct {
	FromID       string `json:"from_id"`
	ToID         string `json:"to_id"`
	Relationship string `json:"relationship,omitempty"`
	Message      string `json:"message"`
}

func (r NoteLinkResult) String() string {
	return r.Message
}

// NoteSearchParams are the parameters for the memory_note_search tool.
type NoteSearchParams struct {
	// Query is the semantic search query.
	Query string `json:"query,omitempty" jsonschema:"description=Semantic search query"`

	// Tags filters by tags.
	Tags []string `json:"tags,omitempty" jsonschema:"description=Filter by tags"`

	// Category filters by category.
	Category string `json:"category,omitempty" jsonschema:"enum=preference,enum=fact,enum=correction,enum=pattern,description=Filter by category"`

	// Scope filters by scope.
	Scope string `json:"scope,omitempty" jsonschema:"enum=global,enum=agent,enum=path,enum=all,description=Scope to search (default: all)"`

	// IncludeLinked includes linked notes in results.
	IncludeLinked bool `json:"include_linked,omitempty" jsonschema:"description=Include linked notes in results (default: true)"`

	// Limit is the maximum number of results.
	Limit int `json:"limit,omitempty" jsonschema:"description=Maximum results to return (default: 10)"`
}

// NoteSearchResult contains the result of a note search operation.
type NoteSearchResult struct {
	Notes   []NoteMatch `json:"notes"`
	Message string      `json:"message"`
}

// NoteMatch represents a matched note.
type NoteMatch struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags,omitempty"`
	Similarity float32  `json:"similarity,omitempty"`
}

func (r NoteSearchResult) String() string {
	if len(r.Notes) == 0 {
		return "No matching notes found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d notes:\n\n", len(r.Notes)))
	for i, n := range r.Notes {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, n.Category, n.Title))
		if len(n.Content) > 100 {
			sb.WriteString(fmt.Sprintf("   %s...\n", n.Content[:100]))
		} else {
			sb.WriteString(fmt.Sprintf("   %s\n", n.Content))
		}
	}
	return sb.String()
}

// NoteReadParams are the parameters for the memory_note_read tool.
type NoteReadParams struct {
	// ID is the note ID to read.
	ID string `json:"id" jsonschema:"required,description=ID of the note to read"`

	// IncludeLinks includes linked notes.
	IncludeLinks bool `json:"include_links,omitempty" jsonschema:"description=Include linked notes (default: true)"`
}

// NoteReadResult contains the result of a note read operation.
type NoteReadResult struct {
	ID       string      `json:"id"`
	Title    string      `json:"title"`
	Content  string      `json:"content"`
	Category string      `json:"category"`
	Tags     []string    `json:"tags,omitempty"`
	Links    []NoteLink  `json:"links,omitempty"`
	Backlink []NoteLink  `json:"backlinks,omitempty"`
	Message  string      `json:"message"`
}

// NoteLink represents a link to another note.
type NoteLink struct {
	ID           string `json:"id"`
	Title        string `json:"title,omitempty"`
	Relationship string `json:"relationship,omitempty"`
}

func (r NoteReadResult) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", r.Title))
	sb.WriteString(fmt.Sprintf("**ID:** %s | **Category:** %s\n\n", r.ID, r.Category))
	if len(r.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(r.Tags, ", ")))
	}
	sb.WriteString(r.Content)
	if len(r.Links) > 0 {
		sb.WriteString("\n\n## Links\n")
		for _, l := range r.Links {
			if l.Relationship != "" {
				sb.WriteString(fmt.Sprintf("- [[%s]] (%s): %s\n", l.ID, l.Relationship, l.Title))
			} else {
				sb.WriteString(fmt.Sprintf("- [[%s]]: %s\n", l.ID, l.Title))
			}
		}
	}
	if len(r.Backlink) > 0 {
		sb.WriteString("\n## Backlinks\n")
		for _, l := range r.Backlink {
			sb.WriteString(fmt.Sprintf("- [[%s]]: %s\n", l.ID, l.Title))
		}
	}
	return sb.String()
}

// NoteUpdateParams are the parameters for the memory_note_update tool.
type NoteUpdateParams struct {
	// ID is the note ID to update.
	ID string `json:"id" jsonschema:"required,description=ID of the note to update"`

	// Content is the new content (optional).
	Content string `json:"content,omitempty" jsonschema:"description=New content for the note"`

	// Tags are new tags (optional).
	Tags []string `json:"tags,omitempty" jsonschema:"description=New tags for the note"`

	// AddLinks adds links to other notes.
	AddLinks []string `json:"add_links,omitempty" jsonschema:"description=IDs of notes to link to"`

	// RemoveLinks removes links to other notes.
	RemoveLinks []string `json:"remove_links,omitempty" jsonschema:"description=IDs of notes to unlink"`
}

// NoteUpdateResult contains the result of a note update operation.
type NoteUpdateResult struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (r NoteUpdateResult) String() string {
	return r.Message
}

// NoteToolConfig configures the note tools.
type NoteToolConfig struct {
	// Service is the memory service to use.
	Service *memory.Service

	// Index is the Zettelkasten index for link management.
	Index *zettelkasten.Index

	// AgentHandle is the current agent's handle.
	AgentHandle string

	// PathScope is the current working directory.
	PathScope string

	// SessionID is the current session ID.
	SessionID string
}

// NewNoteCreateTool creates a tool for creating Zettelkasten notes.
func NewNoteCreateTool(cfg NoteToolConfig) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"memory_note_create",
		"Create a new Zettelkasten memory note with optional links to other notes. Use this to store structured information that can be linked and searched.",
		func(ctx context.Context, params NoteCreateParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Title == "" {
				return fantasy.NewTextErrorResponse("title is required"), nil
			}
			if params.Content == "" {
				return fantasy.NewTextErrorResponse("content is required"), nil
			}

			if cfg.Service == nil {
				return fantasy.NewTextErrorResponse("memory service not available"), nil
			}

			// Parse category
			cat := memory.CategoryFact
			switch params.Category {
			case "preference":
				cat = memory.CategoryPreference
			case "correction":
				cat = memory.CategoryCorrection
			case "pattern":
				cat = memory.CategoryPattern
			case "fact", "":
				cat = memory.CategoryFact
			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid category: %s", params.Category)), nil
			}

			// Determine scope
			agentHandle := ""
			pathScope := ""
			switch params.Scope {
			case "agent":
				agentHandle = cfg.AgentHandle
			case "path":
				pathScope = cfg.PathScope
			case "global", "":
				// Both empty = global scope
			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid scope: %s", params.Scope)), nil
			}

			// Create the memory with title as prefix
			// Include tags in content for searchability
			fullContent := fmt.Sprintf("# %s\n\n%s", params.Title, params.Content)
			if len(params.Tags) > 0 {
				fullContent = fmt.Sprintf("# %s\n\nTags: %s\n\n%s", params.Title, strings.Join(params.Tags, ", "), params.Content)
			}
			mem := memory.Memory{
				Content:         fullContent,
				Category:        cat,
				AgentHandle:     agentHandle,
				PathScope:       pathScope,
				SourceSessionID: cfg.SessionID,
			}

			created, err := cfg.Service.Create(ctx, mem)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to create note: %v", err)), nil
			}

			// Create links if provided and index is available
			if cfg.Index != nil && len(params.Links) > 0 {
				for _, linkID := range params.Links {
					if err := cfg.Index.AddNoteLink(ctx, created.ID, linkID, "relates_to"); err != nil {
						// Log but don't fail
						continue
					}
				}
			}

			result := NoteCreateResult{
				ID:      created.ID,
				Title:   params.Title,
				Links:   params.Links,
				Message: fmt.Sprintf("Created note [%s] %s with ID %s", cat, params.Title, created.ID),
			}

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// NewNoteLinkTool creates a tool for linking memory notes.
func NewNoteLinkTool(cfg NoteToolConfig) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"memory_note_link",
		"Create a bidirectional link between two memory notes. Use this to connect related information.",
		func(ctx context.Context, params NoteLinkParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.FromID == "" {
				return fantasy.NewTextErrorResponse("from_id is required"), nil
			}
			if params.ToID == "" {
				return fantasy.NewTextErrorResponse("to_id is required"), nil
			}

			if cfg.Index == nil {
				return fantasy.NewTextErrorResponse("note linking not available"), nil
			}

			// Create the link
			rel := params.Relationship
			if rel == "" {
				rel = "relates_to"
			}

			if err := cfg.Index.AddNoteLink(ctx, params.FromID, params.ToID, rel); err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to link notes: %v", err)), nil
			}

			result := NoteLinkResult{
				FromID:       params.FromID,
				ToID:         params.ToID,
				Relationship: rel,
				Message:      fmt.Sprintf("Linked %s -> %s (%s)", params.FromID, params.ToID, rel),
			}

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// NewNoteSearchTool creates a tool for searching memory notes.
func NewNoteSearchTool(cfg NoteToolConfig) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"memory_note_search",
		"Search memory notes semantically or by metadata. Returns notes matching the query with optional linked notes.",
		func(ctx context.Context, params NoteSearchParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Query == "" && len(params.Tags) == 0 && params.Category == "" {
				return fantasy.NewTextErrorResponse("at least one search criteria (query, tags, or category) is required"), nil
			}

			if cfg.Service == nil {
				return fantasy.NewTextErrorResponse("memory service not available"), nil
			}

			// Set defaults
			limit := params.Limit
			if limit == 0 {
				limit = 10
			}

			// Build search options
			opts := memory.SearchOptions{
				Limit: limit,
			}

			// Apply scope filtering
			switch params.Scope {
			case "global":
				// Empty agent/path = global only
			case "agent":
				opts.AgentHandle = cfg.AgentHandle
			case "path":
				opts.PathScope = cfg.PathScope
			case "all", "":
				opts.AgentHandle = cfg.AgentHandle
				opts.PathScope = cfg.PathScope
			}

			// Apply category filter
			if params.Category != "" {
				switch params.Category {
				case "preference":
					opts.Categories = []memory.Category{memory.CategoryPreference}
				case "fact":
					opts.Categories = []memory.Category{memory.CategoryFact}
				case "correction":
					opts.Categories = []memory.Category{memory.CategoryCorrection}
				case "pattern":
					opts.Categories = []memory.Category{memory.CategoryPattern}
				}
			}

			query := params.Query
			if query == "" && len(params.Tags) > 0 {
				query = strings.Join(params.Tags, " ")
			}

			results, err := cfg.Service.Search(ctx, query, opts)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("search failed: %v", err)), nil
			}

			// Convert to result format
			notes := make([]NoteMatch, 0, len(results))
			for _, r := range results {
				// Extract title from content (first line starting with #)
				title := extractTitle(r.Memory.Content)
				notes = append(notes, NoteMatch{
					ID:         r.Memory.ID,
					Title:      title,
					Content:    r.Memory.Content,
					Category:   string(r.Memory.Category),
					Tags:       extractTags(r.Memory.Content),
					Similarity: r.Similarity,
				})
			}

			result := NoteSearchResult{
				Notes:   notes,
				Message: fmt.Sprintf("Found %d matching notes", len(notes)),
			}

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// NewNoteReadTool creates a tool for reading a specific memory note.
func NewNoteReadTool(cfg NoteToolConfig) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"memory_note_read",
		"Read a specific memory note by ID, optionally including its links and backlinks.",
		func(ctx context.Context, params NoteReadParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.ID == "" {
				return fantasy.NewTextErrorResponse("id is required"), nil
			}

			if cfg.Service == nil {
				return fantasy.NewTextErrorResponse("memory service not available"), nil
			}

			// Get the memory
			mem, err := cfg.Service.Get(ctx, params.ID)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("note not found: %v", err)), nil
			}

			result := NoteReadResult{
				ID:       mem.ID,
				Title:    extractTitle(mem.Content),
				Content:  mem.Content,
				Category: string(mem.Category),
				Tags:     extractTags(mem.Content),
			}

			// Get links if requested and index is available
			includeLinks := params.IncludeLinks
			if !params.IncludeLinks && params.ID != "" {
				includeLinks = true // Default to true
			}

			if cfg.Index != nil && includeLinks {
				outgoing, incoming, err := cfg.Index.GetAllNoteLinks(ctx, params.ID)
				if err == nil {
					for _, l := range outgoing {
						link := NoteLink{
							ID:           l.ToNoteID,
							Relationship: l.Relationship,
						}
						// Try to get title
						if linkedMem, err := cfg.Service.Get(ctx, l.ToNoteID); err == nil {
							link.Title = extractTitle(linkedMem.Content)
						}
						result.Links = append(result.Links, link)
					}
					for _, l := range incoming {
						link := NoteLink{
							ID:           l.FromNoteID,
							Relationship: l.Relationship,
						}
						// Try to get title
						if linkedMem, err := cfg.Service.Get(ctx, l.FromNoteID); err == nil {
							link.Title = extractTitle(linkedMem.Content)
						}
						result.Backlink = append(result.Backlink, link)
					}
				}
			}

			result.Message = fmt.Sprintf("Read note %s", params.ID)

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// NewNoteUpdateTool creates a tool for updating memory notes.
func NewNoteUpdateTool(cfg NoteToolConfig) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"memory_note_update",
		"Update an existing memory note's content, tags, or links.",
		func(ctx context.Context, params NoteUpdateParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.ID == "" {
				return fantasy.NewTextErrorResponse("id is required"), nil
			}

			if cfg.Service == nil {
				return fantasy.NewTextErrorResponse("memory service not available"), nil
			}

			// Get existing memory
			mem, err := cfg.Service.Get(ctx, params.ID)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("note not found: %v", err)), nil
			}

			// Apply updates
			if params.Content != "" {
				// Preserve title if present
				title := extractTitle(mem.Content)
				if title != "" {
					mem.Content = fmt.Sprintf("# %s\n\n%s", title, params.Content)
				} else {
					mem.Content = params.Content
				}
			}

			// Update tags in content if provided
			if len(params.Tags) > 0 {
				title := extractTitle(mem.Content)
				contentWithoutTitle := strings.TrimPrefix(mem.Content, "# "+title+"\n\n")
				// Remove old tags line if present
				if idx := strings.Index(contentWithoutTitle, "Tags: "); idx == 0 {
					if newlineIdx := strings.Index(contentWithoutTitle, "\n\n"); newlineIdx > 0 {
						contentWithoutTitle = contentWithoutTitle[newlineIdx+2:]
					}
				}
				mem.Content = fmt.Sprintf("# %s\n\nTags: %s\n\n%s", title, strings.Join(params.Tags, ", "), contentWithoutTitle)
			}

			// Update the memory
			if err := cfg.Service.Update(ctx, mem); err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to update note: %v", err)), nil
			}

			// Handle link changes if index is available
			if cfg.Index != nil {
				for _, linkID := range params.AddLinks {
					_ = cfg.Index.AddNoteLink(ctx, params.ID, linkID, "relates_to")
				}
				for _, linkID := range params.RemoveLinks {
					_ = cfg.Index.RemoveNoteLink(ctx, params.ID, linkID)
				}
			}

			result := NoteUpdateResult{
				ID:      params.ID,
				Message: fmt.Sprintf("Updated note %s", params.ID),
			}

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// extractTitle extracts the title from markdown content (first # heading).
func extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	// Return first 50 chars if no heading
	if len(content) > 50 {
		return content[:50] + "..."
	}
	return content
}

// extractTags extracts tags from content (looks for "Tags: tag1, tag2" line).
func extractTags(content string) []string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Tags: ") {
			tagsStr := strings.TrimPrefix(line, "Tags: ")
			tags := strings.Split(tagsStr, ", ")
			result := make([]string, 0, len(tags))
			for _, t := range tags {
				t = strings.TrimSpace(t)
				if t != "" {
					result = append(result, t)
				}
			}
			return result
		}
	}
	return nil
}
