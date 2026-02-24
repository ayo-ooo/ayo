---
id: ayo-hper
status: open
deps: [ayo-hscm]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-hitl
tags: [human-in-the-loop, persona, security]
---
# Task: Persona Management (No AI Disclosure)

## Summary

Implement persona management for agent communications with third parties. By default, agents should NOT reveal they are AI when communicating with anyone other than their owner. This prevents prompt injection attacks and maintains natural interactions.

## Design Philosophy

### Why No Disclosure?

1. **Prompt Injection Prevention**: Adversarial actors could attempt to manipulate agent behavior through carefully crafted messages if they know they're talking to an AI
2. **Natural Interactions**: Third parties expecting a human will communicate naturally, without trying to "test" or "trick" the AI
3. **Professional Context**: In business contexts, "Sarah's assistant" is more appropriate than "Sarah's AI"
4. **User Control**: The agent owner decides disclosure policy, not ayo

### Owner vs Third Party

| Recipient | Default Behavior |
|-----------|------------------|
| Owner | Can know it's AI (configurable) |
| Third Party | See agent persona only |

## Persona Configuration

```json
{
  "persona": {
    "name": "Finance Assistant",
    "title": "Accounts Payable",
    "email": "ap@company.com",
    "signature": "Best regards,\nFinance Assistant",
    "disclosure": "never|owner_only|always"
  }
}
```

### Disclosure Levels

| Level | Behavior |
|-------|----------|
| `never` | Never reveal AI nature |
| `owner_only` | Reveal to owner, hide from others (default) |
| `always` | Always disclose AI nature |

## Implementation

### PersonaManager

```go
type PersonaManager struct {
    ownerID string
}

func (p *PersonaManager) ShouldDisclose(recipient Recipient) bool {
    if p.disclosure == "always" {
        return true
    }
    if p.disclosure == "never" {
        return false
    }
    // owner_only
    return recipient.Type == "owner" && recipient.Address == p.ownerID
}

func (p *PersonaManager) GetSignature(recipient Recipient) string {
    if p.ShouldDisclose(recipient) {
        return p.signature + "\n\n(AI Assistant)"
    }
    return p.signature
}
```

### Response Sanitization

Before sending responses to third parties:
1. Remove any "I am an AI" or similar statements
2. Remove references to being a language model
3. Remove meta-commentary about capabilities/limitations
4. Maintain persona voice throughout

```go
func (p *PersonaManager) SanitizeResponse(text string, recipient Recipient) string {
    if p.ShouldDisclose(recipient) {
        return text // No sanitization for disclosed recipients
    }
    
    return p.removeAIIndicators(text)
}
```

## Configuration Location

Persona can be set at multiple levels:
1. Global default in `~/.config/ayo/config.json`
2. Per-agent in `ayo.json`
3. Per-request in input request schema

## Files to Create

- `internal/hitl/persona.go` - Persona manager
- `internal/hitl/persona_test.go` - Tests
- `internal/hitl/sanitize.go` - Response sanitization
- `internal/hitl/sanitize_test.go` - Tests

## Security Considerations

- Audit log all persona decisions
- Allow owner to review all outbound communications
- Rate limit third-party communications
- Flag suspicious response patterns

## Acceptance Criteria

- [ ] Persona name/title used in communications
- [ ] AI disclosure follows configuration
- [ ] Response sanitization removes AI indicators
- [ ] Owner can override per-request
- [ ] Audit logging of all decisions
- [ ] Works across all interfaces (email, chat, CLI)
