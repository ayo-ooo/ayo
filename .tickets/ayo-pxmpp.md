---
id: ayo-pxmpp
status: open
deps: [ayo-pltg]
links: []
created: 2026-02-23T12:00:00Z
type: epic
priority: 3
assignee: Alex Cabrera
tags: [plugins, triggers, external-repo, open-standards]
---
# Epic: XMPP/Jabber Trigger Plugin

## Summary

Create `ayo-plugins-xmpp` - a trigger plugin for XMPP (Jabber) messaging. XMPP is an open standard (RFC 6120/6121) for real-time messaging, used by many services and self-hosted servers.

## Open Standards Focus

- **XMPP Core (RFC 6120)** - Transport protocol
- **XMPP IM (RFC 6121)** - Instant messaging extension
- **XEP-0045** - Multi-User Chat (MUC)
- **XEP-0313** - Message Archive Management
- Federated, decentralized protocol

## Use Cases

1. **Self-hosted Chat Bot** - Respond to messages on Prosody/ejabberd
2. **Team Chat Integration** - Connect to company XMPP server
3. **IoT Notifications** - Receive XMPP messages from IoT devices
4. **Alert Handler** - Process monitoring alerts via XMPP
5. **Conference Bot** - Participate in MUC rooms

## Plugin Components

```
ayo-plugins-xmpp/
├── manifest.json
├── triggers/
│   └── xmpp/
│       ├── trigger.json
│       └── xmpp-trigger       # Binary
├── tools/
│   ├── xmpp-send/
│   ├── xmpp-presence/
│   └── xmpp-roster/
├── agents/
│   └── @xmpp-bot/
└── skills/
    └── chat-assistant.md
```

## Supported Servers

Any XMPP-compliant server:
- **Prosody** (lightweight, Lua-based)
- **ejabberd** (Erlang, enterprise-grade)
- **Openfire** (Java)
- **Conversations.im** (via any server)
- **Movim** (social network on XMPP)

## Trigger Specification

### Configuration Schema

```json
{
  "type": "xmpp",
  "config": {
    "jid": {
      "type": "string",
      "required": true,
      "description": "Full JID (user@domain/resource)"
    },
    "password": {
      "type": "string",
      "required": true,
      "secret": true,
      "env": "XMPP_PASSWORD"
    },
    "server": {
      "type": "string",
      "description": "Override server (for SRV record failures)"
    },
    "port": {
      "type": "integer",
      "default": 5222
    },
    "use_tls": {
      "type": "boolean",
      "default": true
    },
    "muc_rooms": {
      "type": "array",
      "items": { "type": "string" },
      "description": "MUC rooms to join (room@conference.server)"
    },
    "filter": {
      "type": "object",
      "properties": {
        "from_jids": { "type": "array", "items": { "type": "string" } },
        "message_contains": { "type": "string" },
        "ignore_own": { "type": "boolean", "default": true }
      }
    }
  }
}
```

### Event Payload

```json
{
  "event_type": "xmpp.message",
  "message_id": "abc123",
  "from": {
    "jid": "user@example.com/mobile",
    "bare_jid": "user@example.com",
    "resource": "mobile"
  },
  "to": {
    "jid": "bot@example.com/ayo"
  },
  "type": "chat",
  "body": "Hey, can you help me with something?",
  "subject": null,
  "thread": "thread-123",
  "timestamp": "2026-02-23T10:30:00Z",
  "is_muc": false,
  "muc_room": null
}
```

### MUC Event Payload

```json
{
  "event_type": "xmpp.muc_message",
  "message_id": "def456",
  "from": {
    "jid": "room@conference.example.com/nickname",
    "room": "room@conference.example.com",
    "nickname": "nickname"
  },
  "body": "@bot please summarize today's discussion",
  "is_mention": true,
  "timestamp": "2026-02-23T10:30:00Z"
}
```

## Tool Specifications

### xmpp-send

```json
{
  "name": "xmpp-send",
  "description": "Send XMPP message",
  "parameters": {
    "to": { "type": "string", "required": true, "description": "JID or MUC room" },
    "body": { "type": "string", "required": true },
    "type": { "type": "string", "enum": ["chat", "groupchat"], "default": "chat" },
    "thread": { "type": "string" }
  }
}
```

### xmpp-presence

```json
{
  "name": "xmpp-presence",
  "description": "Set XMPP presence status",
  "parameters": {
    "show": { "type": "string", "enum": ["available", "away", "dnd", "xa"], "default": "available" },
    "status": { "type": "string", "description": "Status message" }
  }
}
```

### xmpp-roster

```json
{
  "name": "xmpp-roster",
  "description": "List XMPP contacts",
  "parameters": {
    "online_only": { "type": "boolean", "default": false }
  }
}
```

## Agent: @xmpp-bot

```markdown
# @xmpp-bot

You are an XMPP chat bot that responds to messages.

## Capabilities

- Respond to direct messages
- Participate in MUC rooms
- Set presence status
- Handle mentions in group chats

## Guidelines

1. Respond promptly to messages
2. Be concise in chat format
3. Use threading for multi-turn conversations
4. Set appropriate presence when busy
5. Handle offline messages gracefully
```

## Implementation Steps

1. [ ] Create repository `ayo-plugins-xmpp`
2. [ ] Implement XMPP client (using mellium.im/xmpp)
3. [ ] Implement connection management and reconnection
4. [ ] Implement MUC support (XEP-0045)
5. [ ] Create message trigger
6. [ ] Implement xmpp-send tool
7. [ ] Implement xmpp-presence tool
8. [ ] Implement xmpp-roster tool
9. [ ] Create @xmpp-bot agent
10. [ ] Write documentation
11. [ ] Add tests with local Prosody server

## Dependencies

- Depends on: `ayo-pltg` (trigger plugin architecture)
- Go libraries:
  - `mellium.im/xmpp` - Modern Go XMPP library

## Security Considerations

- TLS required by default
- Password stored in environment variable
- Support SASL authentication
- Handle certificate verification

---

*Created: 2026-02-23*
