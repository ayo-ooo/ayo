# Ayo Memory System

Ayo stores facts about you locally and uses them across sessions.

## Check Setup

Run `ayo doctor`. Confirm Ollama Service, Embedding Model, and Small Model show OK.

## Store a Memory

```bash
ayo memory store "I prefer TypeScript"
```

The category is auto-detected. "I prefer TypeScript" becomes a `preference`.

> A small LLM (ministral-3:3b) analyzes the content and assigns the appropriate category.

Override with `--category` if needed:

```bash
ayo memory store "Project uses PostgreSQL 15" --category fact
```

Categories: `preference`, `fact`, `correction`, `pattern`

## List and View

```bash
ayo memory list
ayo memory show           # interactive picker
ayo memory show b7        # or use ID prefix
```

## Search

```bash
ayo memory search "programming language"
```

Finds "I prefer TypeScript" because it understands meaning, not keywords.

> Search uses vector embeddings (nomic-embed-text) to match by semantic similarity.

## Automatic Formation

During chat, say something memorable:

```
Remember I always want verbose output
```

Exit and run `ayo memory list` to see what was stored.

> The same LLM extracts memorable content from conversations automatically.

## Retrieval

Start a new session. Relevant memories load automatically based on your first message.

> Memories are retrieved via semantic search and injected into the system prompt.

## Agent Memory Tool

Agents can access memory directly:

```
Search my memories for preferences
Store that this project uses PostgreSQL 15
```

Memory storage is asynchronous - the agent continues immediately while the memory is stored in the background. You'll see status updates:

```
◇ Storing memory...
◆ Memory stored
```

## Management

```bash
ayo memory stats           # Show statistics
ayo memory forget          # Interactive picker
ayo memory forget b7       # Or use ID prefix
ayo memory clear           # Delete all
ayo memory list --json     # JSON output
```

## How It Works

> 1. Small LLM analyzes content and assigns category
> 2. Memories are converted to vector embeddings
> 3. Duplicates are detected via semantic similarity before storing
> 4. At session start, relevant memories are retrieved and injected
> 5. Everything stored locally in SQLite
