# Memory

The memory system stores facts, preferences, and patterns about you that persist across sessions, enabling agents to provide personalized and contextual responses.

## Overview

1. Agents detect memorable information during conversations
2. Memories are stored with vector embeddings for semantic search
3. Relevant memories are automatically retrieved at session start
4. Agents can use the memory tool to search, store, and manage memories

## Prerequisites

Memory requires Ollama for local LLM inference:

```bash
# Install Ollama
brew install ollama  # macOS

# Start service
ollama serve

# Pull required models
ollama pull nomic-embed-text   # For embeddings
ollama pull ministral-3:3b     # For extraction
```

Verify with:

```bash
ayo doctor
```

## Quick Start

### Store a Memory

```bash
ayo memory store "I prefer TypeScript"
```

The category is auto-detected. Override if needed:

```bash
ayo memory store "Project uses PostgreSQL 15" --category fact
```

### Search Memories

```bash
ayo memory search "programming language"
```

Finds "I prefer TypeScript" via semantic similarity.

### List Memories

```bash
ayo memory list
```

### Forget a Memory

```bash
ayo memory forget b7f3
```

## Categories

| Category | Description |
|----------|-------------|
| `preference` | User preferences (tools, styles, communication) |
| `fact` | Facts about user or project |
| `correction` | User corrections to agent behavior |
| `pattern` | Observed behavioral patterns |

Categories are auto-detected when storing via CLI or agent.

## Commands

### List

```bash
# All memories
ayo memory list

# Filter by category
ayo memory list -c preference

# Filter by agent
ayo memory list -a @ayo

# JSON output
ayo memory list --json
```

### Search

```bash
# Semantic search
ayo memory search "coding preferences"

# With threshold and limit
ayo memory search "tools" -t 0.7 -n 5

# Filter by agent
ayo memory search "setup" -a @ayo
```

### Show

```bash
# By ID prefix
ayo memory show b7f3

# Full details
ayo memory show b7f3a2e1-...
```

Shows: content, category, confidence, access count, timestamps.

### Store

```bash
# Auto-categorize
ayo memory store "I prefer tabs over spaces"

# Explicit category
ayo memory store "Project deadline is March 15" -c fact

# Scoped to agent
ayo memory store "Always use verbose output" -a @debugger
```

### Forget

```bash
# With confirmation
ayo memory forget b7f3

# Skip confirmation
ayo memory forget b7f3 -f
```

### Stats

```bash
ayo memory stats
```

Shows counts by category and agent.

### Clear

```bash
# All memories (with confirmation)
ayo memory clear

# For specific agent
ayo memory clear -a @ayo

# Skip confirmation
ayo memory clear -f
```

## Automatic Formation

During conversations, agents automatically detect memorable content:

```
You: Remember I always want verbose output
Agent: I'll remember that.
```

Exit and check:

```bash
ayo memory list
```

### Formation Triggers

Agents detect:
- Explicit "remember" requests
- Stated preferences ("I prefer...", "I like...")
- Corrections ("No, I meant...", "Actually...")
- Project facts ("This project uses...")

## Automatic Retrieval

At session start, relevant memories are retrieved based on:
- Your first message
- Current working directory
- Agent context

Retrieved memories are injected into the system prompt.

## Agent Memory Tool

Agents with `memory` in their `allowed_tools` can:

```
You: Search my memories for my coding preferences
Agent: [searches memories]
Found: "User prefers TypeScript for frontend"

You: Remember that this project uses PostgreSQL
Agent: [stores memory]
Stored as fact.
```

Memory storage is asynchronous - the agent continues immediately while storing.

## Agent Configuration

### Enable Memory

```json
{
  "allowed_tools": ["bash", "memory"]
}
```

### Memory Settings

```json
{
  "memory": {
    "enabled": true,
    "scope": "hybrid",
    "formation_triggers": {
      "on_correction": true,
      "on_preference": true,
      "on_project_fact": true,
      "explicit_only": false
    },
    "retrieval": {
      "auto_inject": true,
      "threshold": 0.3,
      "max_memories": 10
    }
  }
}
```

| Field | Description |
|-------|-------------|
| `enabled` | Enable memory for this agent |
| `scope` | `global`, `agent`, `path`, or `hybrid` |
| `formation_triggers` | When to form memories |
| `retrieval.auto_inject` | Auto-inject at session start |
| `retrieval.threshold` | Similarity threshold (0-1) |
| `retrieval.max_memories` | Max memories to inject |

## Memory Scopes

| Scope | Description |
|-------|-------------|
| `global` | Applies to all agents and directories |
| `agent` | Applies only to specific agent |
| `path` | Applies to specific project/directory |
| `hybrid` | Combines all scopes |

## Storage

Memories are stored in SQLite (`~/.local/share/ayo/ayo.db`) with:
- Content text
- Category
- Vector embedding (for search)
- Confidence score
- Access count
- Timestamps

No external vector database required.

## How It Works

1. **Extraction**: Small LLM (ministral-3:3b) analyzes content
2. **Categorization**: Same LLM assigns category
3. **Embedding**: nomic-embed-text creates vector representation
4. **Deduplication**: Semantic similarity prevents duplicates
5. **Storage**: SQLite with vector as BLOB
6. **Retrieval**: Cosine similarity search at session start

## Privacy

- All data stored locally
- No cloud sync
- Memories can be viewed, edited, or deleted anytime
- Use `ayo memory clear` to remove all memories
