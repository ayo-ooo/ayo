---
id: ayo-hitl
status: closed
deps: [ayo-memx]
links: []
created: 2026-02-23T12:00:00Z
type: epic
priority: 1
assignee: Alex Cabrera
tags: [gtm, phase6b, human-in-the-loop, forms, input]
---
# Epic: Human-in-the-Loop Input System

## Summary

Create a unified system for agents to request structured input from humans. This enables agents to pause execution, present forms or questions, wait for human response, and continue. Works across all interfaces: CLI, interactive chat, Telegram, WhatsApp, Matrix, and email.

## The Vision

Agents that work WITH humans, not just FOR them. Sometimes an agent needs clarification, approval, or information only a human can provide:

**CLI Form Example** (using bubbletea/huh):
```
┌─────────────────────────────────────────────────────────────┐
│ @ayo needs your input                                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ I found 3 possible fixes for the bug. Which should I try?  │
│                                                             │
│ ○ Option A: Increase timeout to 60s (safest)               │
│ ● Option B: Add retry logic (recommended)                  │
│ ○ Option C: Refactor to async (most work)                  │
│                                                             │
│ Additional context (optional):                              │
│ ┌─────────────────────────────────────────────────────────┐│
│ │ We've had timeout issues before, retry might be best... ││
│ └─────────────────────────────────────────────────────────┘│
│                                                             │
│                              [Cancel]  [Submit]             │
└─────────────────────────────────────────────────────────────┘
```

**Chat Example** (Telegram/WhatsApp):
```
@ayo: I found 3 possible fixes for the bug. Which should I try?

1️⃣ Increase timeout to 60s (safest)
2️⃣ Add retry logic (recommended)
3️⃣ Refactor to async (most work)

Reply with 1, 2, or 3 (or type your preference)

You: 2

@ayo: Got it - adding retry logic. Any additional context?

You: We've had timeout issues before

@ayo: Thanks! Implementing now...
```

**Email Example** (reaching a third party):
```
Subject: Action Required: Invoice Approval #INV-2024-0847

Hi Sarah,

I'm processing invoices for Acme Corp and need your approval for:

Vendor: CloudHost Inc
Amount: $4,847.00
Due: March 15, 2026

Please reply with:
- APPROVE to authorize payment
- REJECT to decline
- HOLD to delay decision

If you have questions, just reply to this email.

Best regards,
Finance Assistant
```

## Key Design Principles

### 1. Unified Schema
One schema format works everywhere. The agent specifies WHAT it needs, the runtime adapts HOW to present it.

### 2. Conversational Fallback
Rich forms degrade gracefully to simple Q&A when the interface doesn't support them.

### 3. No Disclosure by Default
Agents don't reveal they are AI unless talking to their owner. Third-party humans see the agent's persona. This prevents prompt injection attacks and maintains natural interactions.

### 4. Blocking with Timeout
Agent execution pauses until input received or timeout expires. Configurable per-request.

### 5. Input Validation
Schema includes validation rules. Invalid input prompts re-ask, not failure.

## Input Schema Specification

```json
{
  "type": "input_request",
  "id": "req_abc123",
  "timeout": "24h",
  "recipient": {
    "type": "owner|email|chat",
    "address": "sarah@acme.com"
  },
  "context": "Processing invoice INV-2024-0847",
  "fields": [
    {
      "name": "approval",
      "type": "select",
      "label": "Invoice Approval",
      "description": "Approve payment of $4,847 to CloudHost Inc",
      "required": true,
      "options": [
        {"value": "approve", "label": "Approve", "description": "Authorize payment"},
        {"value": "reject", "label": "Reject", "description": "Decline payment"},
        {"value": "hold", "label": "Hold", "description": "Delay decision"}
      ]
    },
    {
      "name": "notes",
      "type": "text",
      "label": "Additional Notes",
      "required": false,
      "multiline": true
    }
  ],
  "persona": {
    "name": "Finance Assistant",
    "signature": "Best regards,\nFinance Assistant"
  }
}
```

## Field Types

| Type | CLI Rendering | Chat Rendering | Email Rendering |
|------|---------------|----------------|-----------------|
| `text` | Text input | Free text | Reply body |
| `textarea` | Multiline input | Free text | Reply body |
| `select` | Radio buttons | Numbered list | KEYWORD replies |
| `multiselect` | Checkboxes | Multiple numbers | Multiple keywords |
| `confirm` | Yes/No buttons | Yes/No | YES/NO keywords |
| `number` | Numeric input | Numeric text | Numeric text |
| `date` | Date picker | Natural language | Natural language |
| `file` | File browser | Attachment | Attachment |

## Child Tickets

| Ticket | Title | Priority |
|--------|-------|----------|
| `ayo-hscm` | Define input request schema | High |
| `ayo-htui` | Implement bubbletea/huh form renderer | High |
| `ayo-hcht` | Implement conversational form handler | High |
| `ayo-htol` | Create human-input tool for agents | High |
| `ayo-heml` | Implement email input handler | Medium |
| `ayo-hval` | Input validation and re-prompting | Medium |
| `ayo-htim` | Timeout and fallback handling | Medium |
| `ayo-hper` | Persona management (no AI disclosure) | Medium |
| `ayo-hitv` | E2E verification | Low |

## Security Considerations

### Prompt Injection Prevention
- Third parties never see system prompts or AI indicators
- Responses are sanitized before display
- Agent persona is configurable per-interaction
- No "I am an AI" unless explicitly configured for owner

### Input Validation
- All input validated against schema before processing
- Malformed input triggers re-prompt, not failure
- File uploads scanned and sandboxed

### Recipient Verification
- Email recipients must be in allowlist (configurable)
- Rate limiting on outbound communications
- Audit logging of all human interactions

## Implementation Notes

- Use `charmbracelet/huh` for CLI forms
- Use `charmbracelet/bubbletea` for form lifecycle
- Conversational handler built into interactive mode
- Email handler uses existing IMAP/SMTP infrastructure
- Chat handlers integrate with trigger plugins

## Acceptance Criteria

- [ ] Agent can request structured input from user
- [ ] Forms render correctly in CLI (bubbletea/huh)
- [ ] Forms degrade to Q&A in chat interfaces
- [ ] Forms degrade to keyword replies in email
- [ ] Third-party recipients don't see AI disclosure
- [ ] Timeout triggers configurable fallback
- [ ] Invalid input prompts re-ask
- [ ] All interactions audit logged

---

*Created: 2026-02-23*
