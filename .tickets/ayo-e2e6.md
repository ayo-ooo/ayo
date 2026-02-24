---
id: ayo-e2e6
status: open
deps: [ayo-e2e5]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 4 - Memory System

## Summary

Write Section 4 of the E2E Manual Testing Guide covering the memory storage, search, and context system.

## Content Requirements

### Prerequisites
- Embedding model configured (e.g., ollama with nomic-embed-text)
- Verify: `./ayo doctor` shows embedding provider configured

### Store Facts
```bash
# Store various facts
./ayo memory store "This E2E test uses PostgreSQL 15"
./ayo memory store "The API runs on port 8080"
./ayo memory store "Project lead is Alice"
./ayo memory store "Deployment happens on Fridays"

# Expected: Each command confirms storage
```

### List Memories
```bash
./ayo memory list

# Expected: Shows all stored memories with IDs and timestamps
```

### Search Memories
```bash
# Semantic search
./ayo memory search "database"
# Expected: Returns PostgreSQL fact

./ayo memory search "who is the lead"
# Expected: Returns Alice fact

./ayo memory search "when do we deploy"
# Expected: Returns Friday fact
```

### Memory in Conversations
```bash
# Test that memory context enhances conversations
./ayo "What database does this project use?"
# Expected: Agent mentions PostgreSQL (retrieved from memory)

./ayo "What port is the API on?"
# Expected: Agent mentions 8080
```

### Memory Scopes (if applicable)
```bash
# Global memory (default)
./ayo memory store "Global fact"

# Agent-specific memory
./ayo memory store --scope @tester "Tester-specific fact"

# Verify scope
./ayo memory list --scope @tester
```

### Memory Deletion
```bash
# Get memory ID from list
./ayo memory list

# Delete specific memory
./ayo memory rm <memory-id>

# Verify deletion
./ayo memory list
```

### Memory Export/Import (if applicable)
```bash
# Export
./ayo memory export > /tmp/memories.json

# Clear memories
./ayo memory clear --confirm

# Import
./ayo memory import < /tmp/memories.json

# Verify
./ayo memory list
```

### Verification Criteria
- [ ] Memory storage works
- [ ] Memory listing works
- [ ] Semantic search returns relevant results
- [ ] Memories appear in conversation context
- [ ] Memory deletion works
- [ ] Memory scopes work (if applicable)

## Acceptance Criteria

- [ ] Section written in guide
- [ ] All memory operations documented
- [ ] Search quality verified
- [ ] Context injection verified
- [ ] Cleanup instructions included
