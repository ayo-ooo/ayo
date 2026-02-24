// Package memory provides tools for agents to explicitly store and search memories.
package memory

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/memory"
)

// StoreParams are the parameters for the memory_store tool.
type StoreParams struct {
	// Content is the information to remember.
	Content string `json:"content" jsonschema:"required,description=The information to remember"`

	// Category is the type of memory.
	Category string `json:"category,omitempty" jsonschema:"enum=preference,enum=fact,enum=correction,enum=pattern,description=Category of the memory (default: fact)"`

	// Scope controls where the memory is accessible.
	// - global: Available everywhere
	// - agent: Only for the current agent
	// - path: Only in the current directory
	Scope string `json:"scope,omitempty" jsonschema:"enum=global,enum=agent,enum=path,description=Scope of the memory (default: global)"`
}

// StoreResult contains the result of a memory store operation.
type StoreResult struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (r StoreResult) String() string {
	return r.Message
}

// SearchParams are the parameters for the memory_search tool.
type SearchParams struct {
	// Query is the search query.
	Query string `json:"query" jsonschema:"required,description=Search query to find relevant memories"`

	// Limit is the maximum number of results.
	Limit int `json:"limit,omitempty" jsonschema:"description=Maximum results to return (default: 5)"`

	// Scope controls which memories to search.
	// - global: Only global memories
	// - agent: Only current agent's memories
	// - path: Only current directory's memories
	// - all: Search all accessible memories
	Scope string `json:"scope,omitempty" jsonschema:"enum=global,enum=agent,enum=path,enum=all,description=Scope to search (default: all)"`
}

// SearchResult contains the result of a memory search operation.
type SearchResult struct {
	Memories []MemoryMatch `json:"memories"`
	Message  string        `json:"message"`
}

// MemoryMatch represents a matched memory.
type MemoryMatch struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Category   string  `json:"category"`
	Similarity float32 `json:"similarity"`
}

func (r SearchResult) String() string {
	if len(r.Memories) == 0 {
		return "No matching memories found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d relevant memories:\n\n", len(r.Memories)))
	for i, m := range r.Memories {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s (%.0f%% match)\n", i+1, m.Category, m.Content, m.Similarity*100))
	}
	return sb.String()
}

// ToolConfig configures the memory tools.
type ToolConfig struct {
	// Service is the memory service to use.
	Service *memory.Service

	// AgentHandle is the current agent's handle.
	AgentHandle string

	// PathScope is the current working directory.
	PathScope string

	// SessionID is the current session ID.
	SessionID string
}

// NewStoreMemoryTool creates a tool for storing memories.
func NewStoreMemoryTool(cfg ToolConfig) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"memory_store",
		"Store important information for future reference. Use this to remember user preferences, project facts, corrections, or patterns you've learned.",
		func(ctx context.Context, params StoreParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
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

			// Create the memory
			mem := memory.Memory{
				Content:         params.Content,
				Category:        cat,
				AgentHandle:     agentHandle,
				PathScope:       pathScope,
				SourceSessionID: cfg.SessionID,
			}

			created, err := cfg.Service.Create(ctx, mem)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to store memory: %v", err)), nil
			}

			result := StoreResult{
				ID:      created.ID,
				Message: fmt.Sprintf("Stored memory [%s]: %s", cat, truncate(params.Content, 50)),
			}

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// NewSearchMemoryTool creates a tool for searching memories.
func NewSearchMemoryTool(cfg ToolConfig) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"memory_search",
		"Search for relevant information from past interactions. Use this to recall user preferences, project facts, or previously learned patterns.",
		func(ctx context.Context, params SearchParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Query == "" {
				return fantasy.NewTextErrorResponse("query is required"), nil
			}

			if cfg.Service == nil {
				return fantasy.NewTextErrorResponse("memory service not available"), nil
			}

			// Set defaults
			limit := params.Limit
			if limit == 0 {
				limit = 5
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
				// Search all scopes - include agent and path for context
				// This is the default hybrid behavior
				opts.AgentHandle = cfg.AgentHandle
				opts.PathScope = cfg.PathScope
			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid scope: %s", params.Scope)), nil
			}

			results, err := cfg.Service.Search(ctx, params.Query, opts)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("search failed: %v", err)), nil
			}

			// Convert to result format
			matches := make([]MemoryMatch, 0, len(results))
			for _, r := range results {
				matches = append(matches, MemoryMatch{
					ID:         r.Memory.ID,
					Content:    r.Memory.Content,
					Category:   string(r.Memory.Category),
					Similarity: r.Similarity,
				})
			}

			result := SearchResult{
				Memories: matches,
				Message:  fmt.Sprintf("Found %d matching memories", len(matches)),
			}

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// truncate shortens a string to the given length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
