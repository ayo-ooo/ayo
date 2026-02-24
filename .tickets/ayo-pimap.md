---
id: ayo-pimap
status: open
deps: [ayo-pltg]
links: []
created: 2026-02-23T12:00:00Z
type: epic
priority: 1
assignee: Alex Cabrera
tags: [plugins, triggers, external-repo]
---
# Epic: IMAP Email Trigger Plugin

## Summary

Create `ayo-plugins-imap` - a trigger plugin that enables agents to respond to incoming emails. This is one of the most requested MCP integrations and enables powerful email automation use cases.

## The Dream

Your inbox is overwhelming. Hundreds of emails a day - newsletters you meant to read, meeting invites, sales pitches, actual important messages buried in noise. Your agent watches your inbox 24/7:

**Morning briefing** (delivered to your Telegram at 7am):
> "Good morning. You have 47 new emails. Here's what matters:
> - **Urgent**: CFO needs budget approval by noon (sent 11pm last night)  
> - **Review needed**: 2 PRs from your team (both look straightforward)
> - **FYI**: Board meeting moved to Thursday
> - **Newsletters**: Summarized 8 tech newsletters - key story is Apple's new AI announcement
> - **Archived**: 31 marketing emails, 3 notification emails"

**Throughout the day** - your agent drafts replies:
> "Invoice from AWS arrived. Amount: $2,847. 15% higher than last month due to increased compute usage. Draft reply ready for your approval."

**Proactive triage**:
> "Email from recruiter at Anthropic. Looks like a senior engineering role. Want me to draft a polite decline or save for later?"

This isn't science fiction. It's IMAP + SMTP + LLM + memory. Ayo makes it real.

## Use Cases

1. **Email Responder** - Auto-draft replies to emails matching criteria
2. **Email Classifier** - Triage incoming emails, label/move to folders
3. **Email-to-Task** - Create tasks/tickets from emails
4. **Invoice Processing** - Extract data from invoice emails
5. **Customer Support** - Handle support emails with agent assistance

## Plugin Components

```
ayo-plugins-imap/
├── manifest.json
├── triggers/
│   └── imap/
│       ├── trigger.json        # Trigger definition
│       └── imap-trigger        # Binary (Go)
├── tools/
│   ├── email-send/
│   │   ├── tool.json
│   │   └── email-send
│   ├── email-search/
│   │   ├── tool.json
│   │   └── email-search
│   └── email-move/
│       ├── tool.json
│       └── email-move
├── agents/
│   └── @email-handler/
│       ├── ayo.json
│       └── system.md
└── skills/
    └── email-triage.md
```

## Trigger Specification

### Configuration Schema

```json
{
  "type": "imap",
  "config": {
    "server": {
      "type": "string",
      "required": true,
      "description": "IMAP server hostname"
    },
    "port": {
      "type": "integer",
      "default": 993,
      "description": "IMAP port (993 for TLS)"
    },
    "username": {
      "type": "string",
      "required": true,
      "secret": false
    },
    "password": {
      "type": "string",
      "required": true,
      "secret": true,
      "env": "IMAP_PASSWORD"
    },
    "folder": {
      "type": "string",
      "default": "INBOX"
    },
    "poll_interval": {
      "type": "duration",
      "default": "1m"
    },
    "use_idle": {
      "type": "boolean",
      "default": true,
      "description": "Use IMAP IDLE for real-time notifications"
    },
    "filter": {
      "type": "object",
      "properties": {
        "from": { "type": "string" },
        "subject_contains": { "type": "string" },
        "unseen_only": { "type": "boolean", "default": true }
      }
    }
  }
}
```

### Event Payload

```json
{
  "event_type": "email.received",
  "message_id": "<unique-id@mail.example.com>",
  "from": {
    "name": "John Doe",
    "address": "john@example.com"
  },
  "to": ["recipient@example.com"],
  "cc": [],
  "subject": "Urgent: Need help with deployment",
  "date": "2026-02-23T10:30:00Z",
  "body_text": "Plain text body...",
  "body_html": "<html>...</html>",
  "attachments": [
    {
      "filename": "report.pdf",
      "content_type": "application/pdf",
      "size": 102400,
      "path": "/tmp/attachments/report.pdf"
    }
  ],
  "headers": {
    "Reply-To": "john@example.com",
    "X-Priority": "1"
  }
}
```

## Tool Specifications

### email-send

```json
{
  "name": "email-send",
  "description": "Send an email via SMTP",
  "parameters": {
    "to": { "type": "array", "items": { "type": "string" }, "required": true },
    "cc": { "type": "array", "items": { "type": "string" } },
    "bcc": { "type": "array", "items": { "type": "string" } },
    "subject": { "type": "string", "required": true },
    "body": { "type": "string", "required": true },
    "html": { "type": "boolean", "default": false },
    "reply_to": { "type": "string", "description": "Message-ID to reply to" },
    "attachments": { "type": "array", "items": { "type": "string" } }
  }
}
```

### email-search

```json
{
  "name": "email-search",
  "description": "Search emails in mailbox",
  "parameters": {
    "folder": { "type": "string", "default": "INBOX" },
    "from": { "type": "string" },
    "subject": { "type": "string" },
    "since": { "type": "string", "description": "ISO 8601 date" },
    "before": { "type": "string" },
    "unseen": { "type": "boolean" },
    "limit": { "type": "integer", "default": 20 }
  }
}
```

### email-move

```json
{
  "name": "email-move",
  "description": "Move email to folder",
  "parameters": {
    "message_id": { "type": "string", "required": true },
    "destination": { "type": "string", "required": true },
    "mark_read": { "type": "boolean", "default": false }
  }
}
```

## Agent: @email-handler

```markdown
# @email-handler

You are an email handling agent. You process incoming emails and take appropriate actions.

## Capabilities

- Read and understand email content
- Draft responses
- Categorize emails
- Extract key information
- Move emails to appropriate folders

## Guidelines

1. Always check sender authenticity before taking actions
2. Never send emails without explicit confirmation unless configured
3. Be cautious with attachments
4. Preserve professional tone in responses
5. Log all actions taken
```

## Implementation Steps

1. [ ] Create repository `ayo-plugins-imap`
2. [ ] Implement IMAP connection manager (Go)
3. [ ] Implement IDLE support for real-time notifications
4. [ ] Implement polling fallback for servers without IDLE
5. [ ] Create trigger JSON-RPC interface
6. [ ] Implement email-send tool (SMTP)
7. [ ] Implement email-search tool
8. [ ] Implement email-move tool
9. [ ] Create @email-handler agent
10. [ ] Create email-triage skill
11. [ ] Write documentation
12. [ ] Add integration tests with test mailbox

## Dependencies

- Depends on: `ayo-pltg` (trigger plugin architecture)
- Blocks: None

## Security Considerations

- Passwords stored in environment variables or secrets manager
- TLS required for all connections
- OAuth2 support for Gmail/Outlook (optional enhancement)
- Sandboxed execution for attachment processing
- Rate limiting to prevent spam

## Example Configuration

```yaml
# ~/.config/ayo/triggers/support-inbox.yaml
name: support-inbox
type: imap
config:
  server: imap.gmail.com
  username: support@mycompany.com
  password: ${SUPPORT_IMAP_PASSWORD}
  folder: INBOX
  use_idle: true
  filter:
    unseen_only: true
agent: "@email-handler"
prompt_template: |
  New support email received:
  
  From: {{.From.Name}} <{{.From.Address}}>
  Subject: {{.Subject}}
  Date: {{.Date}}
  
  Body:
  {{.BodyText}}
  
  Please analyze this email and:
  1. Classify priority (low/medium/high/urgent)
  2. Identify the issue type
  3. Draft a helpful response
  4. Suggest which folder to move it to
```

---

*Created: 2026-02-23*
