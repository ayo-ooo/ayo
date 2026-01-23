package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/alexcabrera/ayo/internal/memory"
)

// MemoryContext represents retrieved memories for prompt injection.
type MemoryContext struct {
	Memories []memory.SearchResult
	Section  string // Formatted section for injection
}

// BuildMemoryContext retrieves relevant memories and formats them for prompt injection.
func BuildMemoryContext(ctx context.Context, svc *memory.Service, agentHandle, pathScope, query string, cfg MemoryConfig) (*MemoryContext, error) {
	if svc == nil || !cfg.Enabled {
		return nil, nil
	}

	threshold := cfg.Retrieval.Threshold
	if threshold == 0 {
		threshold = 0.5
	}

	maxMems := cfg.Retrieval.MaxMemories
	if maxMems == 0 {
		maxMems = 10
	}

	// Determine agent filter based on scope
	var agentFilter string
	switch cfg.Scope {
	case "agent":
		agentFilter = agentHandle
	case "global":
		agentFilter = "" // Search all
	case "hybrid", "":
		agentFilter = "" // Search all, let similarity sort
	}

	results, err := svc.Search(ctx, query, memory.SearchOptions{
		AgentHandle: agentFilter,
		PathScope:   pathScope,
		Threshold:   threshold,
		Limit:       maxMems,
	})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	// Format the memory section
	section := formatMemorySection(results, agentHandle)

	return &MemoryContext{
		Memories: results,
		Section:  section,
	}, nil
}

// formatMemorySection formats retrieved memories for prompt injection.
func formatMemorySection(results []memory.SearchResult, agentHandle string) string {
	if len(results) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<user_context>\n")
	sb.WriteString("The following memories were retrieved from previous interactions with this user.\n")
	sb.WriteString("Use this context to provide more personalized and contextual responses.\n\n")

	for i, r := range results {
		// Format: category, content, source info
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, r.Memory.Category, r.Memory.Content))

		// Add scope info if relevant
		var meta []string
		if r.Memory.AgentHandle != "" && r.Memory.AgentHandle != agentHandle {
			meta = append(meta, fmt.Sprintf("from: %s", r.Memory.AgentHandle))
		}
		if r.Memory.PathScope != "" {
			meta = append(meta, fmt.Sprintf("path: %s", r.Memory.PathScope))
		}
		if len(meta) > 0 {
			sb.WriteString(fmt.Sprintf("   (%s)\n", strings.Join(meta, ", ")))
		}
	}

	sb.WriteString("</user_context>\n")
	return sb.String()
}

// InjectMemoryContext adds memory context to the system prompt.
func InjectMemoryContext(systemPrompt string, memoryCtx *MemoryContext) string {
	if memoryCtx == nil || memoryCtx.Section == "" {
		return systemPrompt
	}

	// Insert memory context after the system prompt but before any tools/skills
	return systemPrompt + "\n\n" + memoryCtx.Section
}
