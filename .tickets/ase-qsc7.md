---
id: ase-qsc7
status: closed
deps: [ase-kuef]
links: []
created: 2026-02-09T03:10:15Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-cjpe
---
# Implement capability inference via LLM

## Background

When @ayo needs to select an agent for a task, it needs to understand what each agent is capable of. Rather than requiring manual capability declarations, we infer capabilities by analyzing the agent's system prompt, installed skills, and any schema definitions using an LLM.

## Why This Matters

Users create agents with natural language prompts like "You are a code reviewer who focuses on security issues." The system needs to automatically understand:
- This agent is good at: code review, security analysis, vulnerability detection
- This agent is NOT good at: writing code, project management, data analysis

This enables @ayo to intelligently route tasks to appropriate agents.

## Implementation Details

### Capability Inference Flow

```
Agent Definition → LLM Analysis → Structured Capabilities → SQLite Storage
```

### Capability Schema

```go
// internal/capabilities/types.go
type Capability struct {
    ID          string    // UUID
    AgentID     string    // Agent this belongs to
    Name        string    // Short name: "code-review", "summarization"
    Description string    // Longer explanation
    Confidence  float64   // 0.0-1.0 - how confident are we
    Source      string    // "system_prompt", "skill", "schema"
    Embedding   []float32 // Vector for semantic search
}

type InferenceInput struct {
    SystemPrompt  string   // Agent's system prompt
    SkillNames    []string // Installed skills
    SkillContents []string // SKILL.md contents
    SchemaJSON    string   // Any JSON schema definitions
}
```

### LLM Analysis Prompt

```
Analyze the following agent definition and infer its capabilities.

SYSTEM PROMPT:
{{ .SystemPrompt }}

INSTALLED SKILLS:
{{ range .Skills }}
- {{ .Name }}: {{ .Description }}
{{ end }}

Return a JSON array of capabilities:
[
  {
    "name": "short-kebab-case-name",
    "description": "What this agent can do",
    "confidence": 0.95
  }
]

Focus on:
1. Primary purpose (highest confidence)
2. Secondary abilities (medium confidence)
3. Implied abilities from skills (confidence based on skill relevance)

Do NOT infer capabilities the agent explicitly denies or restricts.
```

### Files to Create

1. `internal/capabilities/inference.go` - LLM-based inference logic
2. `internal/capabilities/types.go` - Capability types and schemas
3. `internal/capabilities/repository.go` - SQLite storage (uses ase-kuef table)
4. `internal/capabilities/inference_test.go` - Tests with sample prompts

### Cache Invalidation

Store a hash of the inference inputs. Re-run inference only when:
- Agent system prompt changes
- Skills are installed/uninstalled
- User explicitly requests refresh (`ayo agents capabilities refresh <agent>`)

```go
type CapabilityCache struct {
    AgentID     string
    InputHash   string    // SHA256 of inference inputs
    InferredAt  time.Time
}
```

## Acceptance Criteria

- [ ] InferenceInput struct captures all relevant agent info
- [ ] LLM prompt produces structured capability JSON
- [ ] Capabilities stored in SQLite with embeddings
- [ ] Cache invalidation based on input hash
- [ ] Confidence scores reflect source reliability
- [ ] Unit tests with diverse agent prompts
- [ ] Handles agents with no clear capabilities gracefully

