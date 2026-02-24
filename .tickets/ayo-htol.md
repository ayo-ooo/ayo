---
id: ayo-htol
status: closed
deps: [ayo-hscm, ayo-htui, ayo-hcht]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-hitl
tags: [human-in-the-loop, tools]
---
# Task: Create human-input Tool for Agents

## Summary

Create a built-in tool that allows agents to request structured input from humans. This is the primary interface for agents to invoke the human-in-the-loop system.

## Tool Specification

```json
{
  "name": "human_input",
  "description": "Request structured input from a human. Use when you need approval, clarification, or information only a human can provide.",
  "parameters": {
    "type": "object",
    "properties": {
      "context": {
        "type": "string",
        "description": "Brief explanation of why you need this input"
      },
      "fields": {
        "type": "array",
        "description": "Fields to collect from the human",
        "items": {
          "type": "object",
          "properties": {
            "name": {"type": "string"},
            "type": {"type": "string", "enum": ["text", "select", "confirm", "number"]},
            "label": {"type": "string"},
            "required": {"type": "boolean"},
            "options": {"type": "array"}
          }
        }
      },
      "recipient": {
        "type": "string",
        "description": "Who to ask: 'owner' (default), or email address"
      },
      "timeout": {
        "type": "string",
        "description": "How long to wait, e.g., '5m', '1h', '24h'"
      }
    },
    "required": ["context", "fields"]
  }
}
```

## Tool Behavior

1. Agent calls `human_input` with schema
2. Runtime detects current interface (CLI, chat, email)
3. Appropriate renderer presents form
4. Execution blocks until response or timeout
5. Response returned to agent

### Context Detection

| Context | Renderer |
|---------|----------|
| Interactive CLI | bubbletea/huh form |
| Interactive chat | Conversational Q&A |
| Triggered (chat plugin) | Conversational via trigger channel |
| Email recipient | Email with keyword replies |
| Timeout | Configurable fallback |

## Implementation

### Tool Handler

```go
func handleHumanInput(ctx context.Context, params HumanInputParams) (any, error) {
    req := buildInputRequest(params)
    
    // Detect interface and get appropriate renderer
    renderer := getRenderer(ctx)
    
    // Block until response
    response, err := renderer.Render(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return response.Values, nil
}
```

## Files to Create/Modify

- `internal/tools/human_input.go` - Tool implementation
- `internal/tools/human_input_test.go` - Tests
- `internal/builtin/tools.go` - Register tool

## Acceptance Criteria

- [ ] Tool callable by any agent
- [ ] Correct renderer selected per interface
- [ ] Execution blocks until response
- [ ] Timeout returns error or fallback
- [ ] Response values returned to agent
- [ ] Tool documented in agent prompts
