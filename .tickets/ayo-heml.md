---
id: ayo-heml
status: closed
deps: [ayo-hscm, ayo-pimap]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-hitl
tags: [human-in-the-loop, email]
---
# Task: Implement Email Input Handler

## Summary

Create an email-based input handler that sends structured requests via email and parses responses. This enables agents to reach third-party humans who may not have access to ayo.

## Email Format

### Outgoing Request

```
From: finance-assistant@company.com (configurable)
To: sarah@acme.com
Subject: Action Required: Invoice Approval #INV-2024-0847

Hi Sarah,

I'm processing invoices for Acme Corp and need your approval for:

Vendor: CloudHost Inc
Amount: $4,847.00
Due: March 15, 2026

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Please reply with one of the following:

• APPROVE - Authorize payment
• REJECT - Decline payment  
• HOLD - Delay decision

You can also reply with any questions.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Best regards,
Finance Assistant
```

### Response Parsing

The handler monitors for replies and parses:
1. Keyword responses (APPROVE, REJECT, HOLD)
2. Free-text responses (treated as notes/questions)
3. Attachments (for file fields)

## Implementation

### EmailInputHandler

```go
type EmailInputHandler struct {
    smtp *SMTPClient
    imap *IMAPClient
}

func (h *EmailInputHandler) Send(req *InputRequest) (string, error) {
    // Generate email from schema
    // Send via SMTP
    // Return message ID for tracking
}

func (h *EmailInputHandler) WaitForResponse(ctx context.Context, messageID string) (*InputResponse, error) {
    // Monitor IMAP for replies
    // Parse response
    // Return values
}
```

### Keyword Extraction

For select fields, generate keyword mapping:
```go
keywords := map[string]string{
    "APPROVE": "approve",
    "REJECT":  "reject",
    "HOLD":    "hold",
    "YES":     "true",
    "NO":      "false",
}
```

Parse reply body for keywords (case-insensitive, first match wins).

## Security

- Recipient allowlist required
- Reply-to threading to prevent spoofing
- Rate limiting on outbound
- Audit logging

## Files to Create

- `internal/hitl/email.go` - Handler
- `internal/hitl/email_test.go` - Tests

## Acceptance Criteria

- [ ] Can send input request emails
- [ ] Can parse keyword responses
- [ ] Can handle free-text replies
- [ ] Threading works correctly
- [ ] Recipient allowlist enforced
- [ ] Timeout handled gracefully
