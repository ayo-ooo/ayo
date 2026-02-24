---
id: ayo-hcht
status: closed
deps: [ayo-hscm]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-hitl
tags: [human-in-the-loop, chat, conversational]
---
# Task: Implement Conversational Form Handler

## Summary

Create a handler that presents InputRequest schemas as conversational Q&A sequences. This works in interactive chat mode and through chat trigger plugins (Telegram, WhatsApp, Matrix).

## Conversational Flow

For a multi-field form, the handler:
1. Presents the context/intro
2. Asks each required field as a question
3. Validates responses inline
4. Re-asks on invalid input
5. Asks optional fields (with skip option)
6. Confirms completion

### Example Flow

Schema:
```json
{
  "fields": [
    {"name": "choice", "type": "select", "options": [...]},
    {"name": "notes", "type": "text", "required": false}
  ]
}
```

Conversation:
```
Agent: I need your input on something.

I found 3 possible fixes. Which should I try?
1️⃣ Increase timeout (safest)
2️⃣ Add retry logic (recommended)
3️⃣ Refactor to async (most work)

Reply with 1, 2, or 3

User: 2

Agent: Got it - "Add retry logic"

Any additional notes? (or reply "skip")

User: skip

Agent: Thanks! Proceeding with option 2.
```

## Implementation

### ConversationalFormHandler

```go
type ConversationalFormHandler struct {
    sendMessage func(string) error
    getMessage  func() (string, error)
}

func (h *ConversationalFormHandler) Run(ctx context.Context, req *InputRequest) (*InputResponse, error) {
    response := &InputResponse{Values: make(map[string]any)}
    
    for _, field := range req.Fields {
        value, err := h.askField(ctx, field)
        if err != nil {
            return nil, err
        }
        response.Values[field.Name] = value
    }
    
    return response, nil
}
```

### Field Presentation

| Type | Presentation |
|------|--------------|
| `select` | Numbered list with emoji |
| `multiselect` | Numbered list, accept multiple |
| `confirm` | "Yes or No?" |
| `text` | Free text prompt |
| `number` | Numeric prompt |
| `date` | Natural language ("tomorrow", "March 15") |

## Files to Create

- `internal/hitl/conversational.go` - Handler
- `internal/hitl/conversational_test.go` - Tests

## Acceptance Criteria

- [ ] All field types have conversational representation
- [ ] Invalid input triggers re-prompt
- [ ] Optional fields can be skipped
- [ ] Timeout handled gracefully
- [ ] Works in interactive chat mode
- [ ] Works with chat trigger plugins
