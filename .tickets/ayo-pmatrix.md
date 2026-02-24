---
id: ayo-pmatrix
status: closed
deps: [ayo-pltg]
links: []
created: 2026-02-23T12:00:00Z
type: epic
priority: 2
assignee: Alex Cabrera
tags: [plugins, triggers, external-repo, chat, open-standards]
---
# Epic: Matrix Protocol Trigger Plugin

## Summary

Create `ayo-plugins-matrix` - a trigger plugin for the Matrix protocol. Matrix is an open standard for decentralized, federated communication, making it ideal for privacy-conscious users and organizations.

## Why Matrix (Open Standard)

- **Open protocol** - Fully documented, federated
- **Self-hostable** - Run your own server (Synapse, Dendrite)
- **Bridges** - Connect to Slack, Discord, IRC, etc.
- **E2E encryption** - Verified encryption with Olm/Megolm
- **Rich ecosystem** - Element, FluffyChat, Nheko clients
- **No vendor lock-in** - Any Matrix server works

## Use Cases

1. **Self-hosted Assistant** - Chat with @ayo on your Matrix server
2. **Team Integration** - Bridge to corporate Matrix/Element
3. **Privacy-first** - E2E encrypted conversations
4. **Multi-room Bot** - Participate in multiple rooms
5. **Federated Access** - Interact across Matrix servers
6. **Bridge Gateway** - Reach IRC, Slack via Matrix bridges

## Plugin Components

```
ayo-plugins-matrix/
├── manifest.json
├── triggers/
│   └── matrix/
│       ├── trigger.json
│       └── matrix-trigger       # Binary
├── tools/
│   ├── matrix-send/
│   ├── matrix-reply/
│   ├── matrix-react/
│   └── matrix-rooms/
├── agents/
│   └── @matrix-bot/
└── skills/
    └── matrix-assistant.md
```

## Trigger Specification

### Configuration Schema

```json
{
  "type": "matrix",
  "config": {
    "homeserver": {
      "type": "string",
      "required": true,
      "description": "Matrix homeserver URL (e.g., https://matrix.org)"
    },
    "user_id": {
      "type": "string",
      "required": true,
      "description": "Full Matrix user ID (@user:server)"
    },
    "access_token": {
      "type": "string",
      "required": true,
      "secret": true,
      "env": "MATRIX_ACCESS_TOKEN"
    },
    "device_id": {
      "type": "string",
      "description": "Device ID for E2E encryption"
    },
    "rooms": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Room IDs to listen to (default: all joined)"
    },
    "encryption": {
      "type": "object",
      "properties": {
        "enabled": { "type": "boolean", "default": true },
        "verify_devices": { "type": "boolean", "default": false },
        "store_path": { "type": "string", "default": "~/.local/share/ayo/matrix/crypto" }
      }
    },
    "filter": {
      "type": "object",
      "properties": {
        "message_types": {
          "type": "array",
          "items": { "type": "string" },
          "default": ["m.text", "m.image", "m.file"]
        },
        "ignore_own": { "type": "boolean", "default": true },
        "mention_required": { "type": "boolean", "default": false }
      }
    }
  }
}
```

### Event Payload

```json
{
  "event_type": "matrix.message",
  "event_id": "$abc123:matrix.org",
  "room": {
    "room_id": "!abc123:matrix.org",
    "name": "General",
    "is_direct": false
  },
  "sender": {
    "user_id": "@john:matrix.org",
    "display_name": "John Doe",
    "avatar_url": "mxc://..."
  },
  "timestamp": "2026-02-23T10:30:00Z",
  "content": {
    "msgtype": "m.text",
    "body": "Hey @ayo, what's the status?",
    "formatted_body": "Hey <a href=\"...\">@ayo</a>, what's the status?",
    "format": "org.matrix.custom.html"
  },
  "is_encrypted": true,
  "is_mention": true,
  "reply_to": null
}
```

### Reply Event

```json
{
  "event_type": "matrix.message",
  "event_id": "$def456:matrix.org",
  "content": {
    "msgtype": "m.text",
    "body": "> Original message\n\nMy reply",
    "m.relates_to": {
      "m.in_reply_to": {
        "event_id": "$abc123:matrix.org"
      }
    }
  },
  "reply_to": {
    "event_id": "$abc123:matrix.org",
    "sender": "@john:matrix.org",
    "body": "Original message"
  }
}
```

### File Event

```json
{
  "event_type": "matrix.file",
  "event_id": "$ghi789:matrix.org",
  "content": {
    "msgtype": "m.file",
    "body": "report.pdf",
    "filename": "report.pdf",
    "info": {
      "mimetype": "application/pdf",
      "size": 102400
    },
    "url": "mxc://matrix.org/abc123"
  },
  "local_path": "/tmp/matrix/report.pdf"
}
```

## Tool Specifications

### matrix-send

```json
{
  "name": "matrix-send",
  "description": "Send a message to a Matrix room",
  "parameters": {
    "room_id": { "type": "string", "required": true },
    "body": { "type": "string", "required": true },
    "formatted_body": { "type": "string", "description": "HTML formatted body" },
    "msgtype": { "type": "string", "default": "m.text" }
  }
}
```

### matrix-reply

```json
{
  "name": "matrix-reply",
  "description": "Reply to the current message (in trigger context)",
  "parameters": {
    "body": { "type": "string", "required": true },
    "formatted_body": { "type": "string" }
  }
}
```

### matrix-react

```json
{
  "name": "matrix-react",
  "description": "React to a message with an emoji",
  "parameters": {
    "event_id": { "type": "string", "required": true },
    "room_id": { "type": "string", "required": true },
    "emoji": { "type": "string", "required": true }
  }
}
```

### matrix-rooms

```json
{
  "name": "matrix-rooms",
  "description": "List joined Matrix rooms",
  "parameters": {
    "include_dm": { "type": "boolean", "default": true }
  }
}
```

## Agent: @matrix-bot

```markdown
# @matrix-bot

You are a Matrix protocol bot.

## Capabilities

- Respond to text messages
- Handle E2E encrypted rooms
- React to messages with emojis
- Process file attachments
- Work across federated servers

## Guidelines

1. Support Matrix reply format
2. Use HTML formatting when appropriate
3. Respect room permissions
4. Handle encryption gracefully
5. React to acknowledge when appropriate
6. Work with bridged messages
```

## Implementation Steps

1. [ ] Create repository `ayo-plugins-matrix`
2. [ ] Implement Matrix client using mautrix-go
3. [ ] Implement E2E encryption (libolm)
4. [ ] Implement sync loop for events
5. [ ] Create message trigger
6. [ ] Implement media download from MXC URLs
7. [ ] Implement matrix-send tool
8. [ ] Implement matrix-reply tool
9. [ ] Implement matrix-react tool
10. [ ] Implement matrix-rooms tool
11. [ ] Create @matrix-bot agent
12. [ ] Handle device verification (optional)
13. [ ] Write documentation
14. [ ] Add tests with local Synapse server

## Dependencies

- Depends on: `ayo-pltg` (trigger plugin architecture)
- Go libraries:
  - `maunium.net/go/mautrix` - Matrix client library
  - `maunium.net/go/mautrix/crypto` - E2E encryption

## Security Considerations

- Access token stored in environment variable
- E2E encryption enabled by default
- Crypto store for Olm sessions
- Device verification for high-security rooms
- Room membership validation

## Self-Hosting

Works with any Matrix homeserver:
```yaml
# Example with self-hosted Synapse
triggers:
  - name: matrix-home
    type: matrix
    config:
      homeserver: https://matrix.mycompany.com
      user_id: "@ayo:mycompany.com"
      access_token: ${MATRIX_TOKEN}
```

---

*Created: 2026-02-23*
