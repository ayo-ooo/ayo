---
id: am-x7qw
status: open
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

## Design

In the entity indexing code, truncate the text sent to the embedding model to fit within context limits. Options: 1) Truncate to first N characters/tokens, 2) Extract key sections (description, capabilities) for embedding, 3) Use a chunking strategy and average embeddings.

## Acceptance Criteria

- 'ayo index rebuild' completes without errors for @ayo
- @ayo appears in 'ayo index search' results
- Embedding quality is sufficient for semantic matching

