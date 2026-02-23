---
id: am-x7qw
status: closed
deps: []
links: []
created: 2026-02-20T02:50:07Z
type: bug
priority: 2
assignee: Alex Cabrera
tags: [index, embedding]
---
# Truncate @ayo system prompt before embedding to fit context window

When running 'ayo index rebuild', embedding @ayo fails with error: 'the input length exceeds the context length'. The @ayo agent's system prompt is too long for nomic-embed-text's context window (8192 tokens). This causes @ayo to be missing from semantic search results.

## Root Cause

Truncation code exists but the limit may be too high:
- Current limit: `maxEmbeddingChars = 24000` (internal/capabilities/index.go:286)
- Assumes ~4 chars/token, but nomic-embed-text has 8192 token limit
- For code/markdown with many special characters, actual ratio is ~2-3 chars/token
- @ayo system prompt is ~307 lines, ~10KB (internal/builtin/agents/@ayo/system.md)

## Key Code Locations

| File | Lines | Purpose |
|------|-------|---------|
| cmd/ayo/index.go | 91-185 | `ayo index rebuild` command |
| cmd/ayo/index.go | 138 | Calls `idx.GetAgentEmbedding()` |
| internal/capabilities/index.go | 213-246 | `GetAgentEmbedding()` embeds description + system prompt |
| internal/capabilities/index.go | 286 | `maxEmbeddingChars = 24000` |
| internal/capabilities/index.go | 288-304 | `truncateForEmbedding()` already implemented |
| internal/ollama/embed.go | 12-16 | `EmbedRequest` missing `truncate` field |

## Implementation Plan

### Option 1: Lower the client-side limit (Recommended, simplest)

```go
// internal/capabilities/index.go, line 286
// Reduce from 24000 to 16000 chars (~2 chars/token conservative estimate)
const maxEmbeddingChars = 16000
```

### Option 2: Add server-side truncation via Ollama API

```go
// internal/ollama/embed.go
type EmbedRequest struct {
    Model    string   `json:"model"`
    Input    []string `json:"input"`
    Truncate *bool    `json:"truncate,omitempty"` // ADD THIS
}

func (c *Client) EmbedBatch(ctx context.Context, model string, texts []string) ([][]float32, error) {
    truncate := true // Enable server-side truncation
    reqBody := EmbedRequest{
        Model:    model,
        Input:    texts,
        Truncate: &truncate,
    }
    // ...
}
```

### Option 3: Extract key sections for embedding

For agents with very long system prompts, embed only semantically relevant parts:
- Description
- First few paragraphs (identity/purpose)
- Skill/capability mentions
- Skip code examples, long instruction sections

## Recommended Fix

Combine Option 1 + Option 2:

1. Lower `maxEmbeddingChars` to 16000 for safety
2. Add `truncate: true` to Ollama API call as defense-in-depth
3. Add a unit test that verifies @ayo can be embedded

## Files to Modify

| File | Change |
|------|--------|
| internal/capabilities/index.go | Line 286: change 24000 → 16000 |
| internal/ollama/embed.go | Add `Truncate` field to `EmbedRequest` |
| internal/capabilities/index_test.go | Add test for large input truncation |

## Test Case

```go
func TestTruncateForEmbedding_LargeInput(t *testing.T) {
    large := strings.Repeat("x", 50000)
    truncated := truncateForEmbedding(large)
    
    if len(truncated) > maxEmbeddingChars {
        t.Errorf("truncateForEmbedding returned %d chars, want <= %d", 
            len(truncated), maxEmbeddingChars)
    }
}
```

## Acceptance Criteria

- 'ayo index rebuild' completes without errors for @ayo
- @ayo appears in 'ayo index search' results
- Embedding quality is sufficient for semantic matching
