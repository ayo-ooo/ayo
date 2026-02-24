---
id: ayo-pwhk
status: closed
deps: [ayo-pltg]
links: []
created: 2026-02-23T12:00:00Z
type: epic
priority: 1
assignee: Alex Cabrera
tags: [plugins, triggers, external-repo]
---
# Epic: Webhook Trigger Plugin

## Summary

Create `ayo-plugins-webhook` - a comprehensive webhook receiver plugin that enables external services to trigger agents. While ayo has basic webhook support in the daemon, this plugin provides a more robust, configurable, and secure webhook receiver.

## Use Cases

1. **GitHub Webhooks** - Trigger on PR, push, issue events
2. **Stripe Webhooks** - Process payment events
3. **Slack Events** - Respond to Slack messages
4. **Generic Webhooks** - Accept webhooks from any service
5. **n8n/Zapier Integration** - Bridge automation platforms

## Plugin Components

```
ayo-plugins-webhook/
├── manifest.json
├── triggers/
│   └── webhook/
│       ├── trigger.json
│       └── webhook-server       # Binary
├── tools/
│   └── webhook-send/
│       ├── tool.json
│       └── webhook-send
├── agents/
│   └── @webhook-handler/
└── skills/
    ├── github-events.md
    └── stripe-events.md
```

## Trigger Specification

### Configuration Schema

```json
{
  "type": "webhook",
  "config": {
    "port": {
      "type": "integer",
      "default": 8080
    },
    "path": {
      "type": "string",
      "default": "/webhook",
      "description": "URL path for this webhook"
    },
    "secret": {
      "type": "string",
      "secret": true,
      "description": "Webhook signature secret"
    },
    "signature_header": {
      "type": "string",
      "default": "X-Hub-Signature-256",
      "description": "Header containing signature"
    },
    "signature_algorithm": {
      "type": "string",
      "enum": ["sha256", "sha1", "none"],
      "default": "sha256"
    },
    "allowed_ips": {
      "type": "array",
      "items": { "type": "string" },
      "description": "IP whitelist (optional)"
    },
    "transform": {
      "type": "string",
      "description": "JQ expression to transform payload"
    }
  }
}
```

### Event Payload

```json
{
  "event_type": "webhook.received",
  "webhook_id": "github-pr",
  "received_at": "2026-02-23T10:30:00Z",
  "source_ip": "192.30.252.1",
  "headers": {
    "X-GitHub-Event": "pull_request",
    "X-GitHub-Delivery": "abc123"
  },
  "body": {
    "action": "opened",
    "pull_request": {
      "number": 42,
      "title": "Add new feature",
      "body": "This PR adds...",
      "user": { "login": "developer" },
      "base": { "ref": "main" },
      "head": { "ref": "feature-branch" }
    }
  }
}
```

## Pre-configured Webhook Templates

### GitHub

```yaml
name: github-webhooks
type: webhook
config:
  path: /webhooks/github
  secret: ${GITHUB_WEBHOOK_SECRET}
  signature_header: X-Hub-Signature-256
  signature_algorithm: sha256
routing:
  - event_header: X-GitHub-Event
    mappings:
      pull_request: "@code-reviewer"
      issues: "@issue-handler"
      push: "@ci-agent"
```

### Stripe

```yaml
name: stripe-webhooks
type: webhook
config:
  path: /webhooks/stripe
  secret: ${STRIPE_WEBHOOK_SECRET}
  signature_header: Stripe-Signature
  signature_algorithm: stripe  # Stripe-specific signing
routing:
  - field: type
    mappings:
      "payment_intent.succeeded": "@payment-handler"
      "customer.subscription.*": "@subscription-handler"
```

### Slack Events

```yaml
name: slack-events
type: webhook
config:
  path: /webhooks/slack
  secret: ${SLACK_SIGNING_SECRET}
  signature_header: X-Slack-Signature
  signature_algorithm: slack
  # Slack challenge handling
  challenge_response: true
```

## Tool: webhook-send

```json
{
  "name": "webhook-send",
  "description": "Send HTTP webhook/request",
  "parameters": {
    "url": { "type": "string", "required": true },
    "method": { "type": "string", "default": "POST" },
    "headers": { "type": "object" },
    "body": { "type": "object" },
    "timeout": { "type": "integer", "default": 30 }
  }
}
```

## Implementation Steps

1. [ ] Create repository `ayo-plugins-webhook`
2. [ ] Implement webhook server with signature validation
3. [ ] Support multiple signature algorithms (HMAC, Stripe, etc.)
4. [ ] Implement IP whitelisting
5. [ ] Add JQ-based payload transformation
6. [ ] Create webhook-send tool
7. [ ] Create @webhook-handler agent
8. [ ] Add GitHub event skill
9. [ ] Add Stripe event skill
10. [ ] Implement challenge-response for Slack
11. [ ] Add ngrok/tunnel integration for local dev
12. [ ] Write documentation

## Dependencies

- Depends on: `ayo-pltg` (trigger plugin architecture)
- Blocks: None

## Security Considerations

- Signature validation required by default
- IP whitelisting for sensitive endpoints
- Rate limiting to prevent DoS
- Request logging for audit
- HTTPS recommended (via reverse proxy)

## Integration with Daemon

The webhook trigger can either:
1. Run its own HTTP server on a configured port
2. Register routes with daemon's existing server

Prefer option 1 for isolation, but support 2 for simplicity.

---

*Created: 2026-02-23*
