# Tutorial: Memory System Deep Dive

Build agents that remember your preferences and project context. By the end, you'll have persistent, personalized AI assistants.

**Time**: ~15 minutes  
**Prerequisites**: [Getting Started](../getting-started.md) complete

## What You'll Build

- An agent that remembers your coding preferences
- Project-specific context that persists across sessions
- Understanding of memory scopes and categories

## Step 1: Enable Memory

Memory is enabled by default for most agents. Check your agent's config:

```json
{
  "memory": {
    "enabled": true,
    "scope": "agent"
  }
}
```

## Step 2: Store Your First Memory

### Via CLI

```bash
ayo memory store "I prefer TypeScript over JavaScript for new projects"
```

### Via Natural Conversation

Just tell the agent:

```bash
ayo "Remember that I always want tabs, not spaces, in my code"
```

Agents automatically recognize preferences and store them.

## Step 3: View Stored Memories

### List All Memories

```bash
ayo memory list
```

Example output:
```
ID          CATEGORY     CREATED      CONTENT
mem_abc123  preference   2024-01-15   I prefer TypeScript over JavaScript
mem_def456  correction   2024-01-14   Use tabs, not spaces
mem_ghi789  fact         2024-01-13   API endpoint is /api/v2/users
```

### Filter by Category

```bash
ayo memory list --category preference
ayo memory list --category fact
ayo memory list --category correction
```

### Search Semantically

```bash
ayo memory search "coding style preferences"
```

## Step 4: Understand Memory Categories

### Preferences

User preferences that guide agent behavior:

```bash
ayo memory store --category preference "Use dark theme in all code examples"
ayo memory store --category preference "Keep explanations concise"
```

### Facts

Factual information about projects or environment:

```bash
ayo memory store --category fact "Production database is PostgreSQL 15"
ayo memory store --category fact "CI/CD runs on GitHub Actions"
```

### Corrections

Corrections to agent behavior:

```bash
ayo memory store --category correction "Don't suggest console.log, use the logger"
ayo memory store --category correction "Our API returns camelCase, not snake_case"
```

### Patterns

Observed behavioral patterns (usually auto-detected):

```bash
ayo memory store --category pattern "User usually wants tests after implementation"
```

## Step 5: Use Memory Scopes

### Global Scope

Available to all agents, everywhere:

```bash
ayo memory store --scope global "My timezone is US/Eastern"
```

### Agent Scope

Available only to a specific agent:

```bash
ayo memory store --scope agent --agent @reviewer "Focus on security in this codebase"
```

### Path Scope

Available when working in a specific directory:

```bash
ayo memory store --scope path --path ~/Projects/myapp "This project uses React 18"
```

### Squad Scope

Available to all agents in a squad:

```bash
ayo memory store --scope squad --squad dev-team "Team standup is at 10am daily"
```

## Step 6: Memory Retrieval in Action

When you start a conversation, relevant memories are automatically retrieved:

```bash
ayo "Help me write a function to parse user input"
```

The agent will:
1. Search memories relevant to "function", "parsing", "user input"
2. Include matching memories in its context
3. Apply your preferences (TypeScript, tabs, etc.)

## Step 7: Manage Memories

### View Memory Details

```bash
ayo memory show mem_abc123
```

### Forget a Memory

Soft-delete (can be recovered):

```bash
ayo memory forget mem_abc123
```

### Permanently Delete

```bash
ayo memory delete mem_abc123
```

### Link Related Memories

```bash
ayo memory link mem_abc123 mem_def456
```

### Merge Duplicates

Find and merge similar memories:

```bash
ayo memory merge
```

## Step 8: Export and Import

### Backup Memories

```bash
ayo memory export > ~/memories-backup.json
```

### Restore Memories

```bash
ayo memory import ~/memories-backup.json
```

### Export Options

```bash
# Include embeddings (larger file, faster import)
ayo memory export --include-embeddings > backup.json

# Export specific scope
ayo memory export --scope agent --agent @reviewer > reviewer-memories.json
```

## How Memory Works

### Storage Architecture

```
~/.local/share/ayo/memory/
├── facts/
│   └── api-endpoint-v2.md
├── preferences/
│   └── typescript-preference.md
├── corrections/
│   └── no-console-log.md
├── patterns/
│   └── testing-after-impl.md
└── .index.sqlite
```

### Zettelkasten Format

Each memory is a markdown file with TOML frontmatter:

```markdown
+++
id = "mem_abc123"
category = "preference"
scope = "global"
created = "2024-01-15T10:30:00Z"
confidence = 0.95
+++

I prefer TypeScript over JavaScript for new projects.
```

### Semantic Search

Memories are embedded as vectors for semantic search:

1. Your query is embedded
2. Cosine similarity finds relevant memories
3. Top matches are included in context

## Complete Example: Personalized Agent

Create an agent that heavily uses memory:

```bash
ayo agent create @assistant
```

**config.json**:
```json
{
  "model": "your-model",
  "description": "Personalized coding assistant",
  "memory": {
    "enabled": true,
    "scope": "global",
    "retrieval": {
      "limit": 10,
      "threshold": 0.7
    }
  }
}
```

**system.md**:
```markdown
# Personal Assistant

You are a personalized coding assistant.

## Memory Usage

Always check your memories for:
- User preferences (coding style, tools, frameworks)
- Project facts (architecture, conventions)
- Previous corrections (things to avoid)

Apply memories proactively - don't ask about preferences
you already know.

## Learning

When the user corrects you or states a preference:
1. Acknowledge the correction
2. Store it as a memory
3. Apply it going forward
```

Populate initial memories:

```bash
ayo memory store --category preference "Use functional React components"
ayo memory store --category preference "Always add TypeScript types"
ayo memory store --category fact "Primary language is TypeScript"
ayo memory store --category correction "Don't suggest class components"
```

Now the agent will personalize responses:

```bash
ayo @assistant "Write a component to display a user profile"
```

The agent will automatically use functional components with TypeScript types.

## Troubleshooting

### Memories not being retrieved

Check retrieval settings in agent config:

```json
{
  "memory": {
    "retrieval": {
      "limit": 10,
      "threshold": 0.7
    }
  }
}
```

Lower threshold = more memories retrieved (less relevant).

### Search returning nothing

Rebuild the search index:

```bash
ayo memory reindex
```

### Too many irrelevant memories

Increase threshold or reduce limit:

```json
{
  "retrieval": {
    "limit": 5,
    "threshold": 0.85
  }
}
```

## Next Steps

- [Triggers](triggers.md) - Use memories across scheduled runs
- [Squads](squads.md) - Share memories between agents
- [Plugins](plugins.md) - Custom memory providers

---

*You've mastered the memory system! Continue to [Plugins](plugins.md).*
