---
id: ayo-pwhats
status: open
deps: [ayo-pltg]
links: []
created: 2026-02-23T12:00:00Z
type: epic
priority: 1
assignee: Alex Cabrera
tags: [plugins, triggers, external-repo, chat]
---
# Epic: WhatsApp Trigger Plugin

## Summary

Create `ayo-plugins-whatsapp` - a trigger plugin enabling agents to respond to WhatsApp messages. Uses the whatsmeow library which implements the WhatsApp Web protocol directly (no Cloud API needed).

## Why WhatsApp

- **2B+ users** - Most popular messaging app globally
- **whatsmeow** - Excellent Go library implementing WhatsApp Web protocol
- **No API costs** - Uses Web protocol, not Cloud API
- **Full features** - Text, media, groups, reactions
- **Personal account** - No business account required

## Technical Approach

Uses **whatsmeow** library (github.com/tulir/whatsmeow):
- Implements WhatsApp Web protocol
- QR code pairing with personal WhatsApp
- Full E2E encryption support
- Maintained by Matrix bridge developer

## Use Cases

1. **Personal Assistant** - Chat with @ayo via WhatsApp
2. **Family Notifications** - Send updates to family group
3. **Quick Tasks** - Send tasks via WhatsApp
4. **File Sharing** - Process files sent via WhatsApp
5. **Status Updates** - Post to WhatsApp status
6. **Multi-device** - Works alongside phone

## Plugin Components

```
ayo-plugins-whatsapp/
├── manifest.json
├── triggers/
│   └── whatsapp/
│       ├── trigger.json
│       └── whatsapp-trigger       # Binary
├── tools/
│   ├── whatsapp-send/
│   ├── whatsapp-reply/
│   ├── whatsapp-media/
│   └── whatsapp-contacts/
├── agents/
│   └── @whatsapp-bot/
└── skills/
    └── whatsapp-assistant.md
```

## Trigger Specification

### Configuration Schema

```json
{
  "type": "whatsapp",
  "config": {
    "device_store": {
      "type": "string",
      "default": "~/.local/share/ayo/whatsapp/device.db",
      "description": "Path to device/session storage"
    },
    "allowed_jids": {
      "type": "array",
      "items": { "type": "string" },
      "description": "WhatsApp JIDs allowed to interact (phone@s.whatsapp.net)"
    },
    "allowed_groups": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Group JIDs to listen to"
    },
    "respond_to_groups": {
      "type": "boolean",
      "default": false,
      "description": "Whether to respond in groups"
    },
    "mention_required": {
      "type": "boolean",
      "default": true,
      "description": "In groups, require @mention to trigger"
    },
    "filter": {
      "type": "object",
      "properties": {
        "message_types": {
          "type": "array",
          "items": { "type": "string" },
          "default": ["text", "image", "document"]
        },
        "ignore_status": { "type": "boolean", "default": true },
        "ignore_broadcast": { "type": "boolean", "default": true }
      }
    }
  }
}
```

### Pairing Flow

First run requires QR code scanning:
```
$ ayo trigger start whatsapp
WhatsApp trigger starting...
Scan this QR code with WhatsApp:

██████████████████████████████
██ ▄▄▄▄▄ ██▄▄ ▄▄█▀█ ▄▄▄▄▄ ██
██ █   █ █▄▀█▀ ▄▄▀█ █   █ ██
... (QR code)

Paired successfully! Connected as: +1234567890
Listening for messages...
```

### Event Payload

```json
{
  "event_type": "whatsapp.message",
  "message_id": "3EB0A0B0C1D2E3F4",
  "chat": {
    "jid": "1234567890@s.whatsapp.net",
    "type": "dm",
    "name": "John Doe"
  },
  "sender": {
    "jid": "1234567890@s.whatsapp.net",
    "name": "John Doe",
    "push_name": "John"
  },
  "timestamp": "2026-02-23T10:30:00Z",
  "text": "Hey, can you help me with something?",
  "quoted_message": null,
  "is_group": false,
  "is_mention": false
}
```

### Group Event

```json
{
  "event_type": "whatsapp.message",
  "message_id": "3EB0A0B0C1D2E3F5",
  "chat": {
    "jid": "1234567890-1234567890@g.us",
    "type": "group",
    "name": "Family Group"
  },
  "sender": {
    "jid": "1234567890@s.whatsapp.net",
    "name": "John Doe"
  },
  "text": "@1234567890 what's the weather today?",
  "is_group": true,
  "is_mention": true,
  "mentioned_jids": ["1234567890@s.whatsapp.net"]
}
```

### Media Event

```json
{
  "event_type": "whatsapp.media",
  "message_id": "3EB0A0B0C1D2E3F6",
  "media_type": "document",
  "media": {
    "mime_type": "application/pdf",
    "file_name": "report.pdf",
    "file_size": 102400,
    "local_path": "/tmp/whatsapp/report.pdf",
    "sha256": "abc123..."
  },
  "caption": "Please review this report"
}
```

## Tool Specifications

### whatsapp-send

```json
{
  "name": "whatsapp-send",
  "description": "Send a WhatsApp message",
  "parameters": {
    "jid": { "type": "string", "required": true, "description": "Recipient JID" },
    "text": { "type": "string", "required": true }
  }
}
```

### whatsapp-reply

```json
{
  "name": "whatsapp-reply",
  "description": "Reply to the current message (in trigger context)",
  "parameters": {
    "text": { "type": "string", "required": true },
    "quote": { "type": "boolean", "default": true, "description": "Quote original message" }
  }
}
```

### whatsapp-media

```json
{
  "name": "whatsapp-media",
  "description": "Send media via WhatsApp",
  "parameters": {
    "jid": { "type": "string", "required": true },
    "file": { "type": "string", "required": true, "description": "Path to file" },
    "caption": { "type": "string" },
    "as_document": { "type": "boolean", "default": false }
  }
}
```

### whatsapp-contacts

```json
{
  "name": "whatsapp-contacts",
  "description": "List WhatsApp contacts",
  "parameters": {
    "search": { "type": "string", "description": "Search filter" }
  }
}
```

## Agent: @whatsapp-bot

```markdown
# @whatsapp-bot

You are a WhatsApp assistant.

## Capabilities

- Respond to text messages
- Process media files
- Send images and documents
- Handle group mentions
- Quote and reply to messages

## Guidelines

1. Keep responses concise (WhatsApp style)
2. Use emojis appropriately
3. In groups, only respond to mentions
4. Acknowledge media before processing
5. Respect conversation context
6. Don't spam - reasonable response rate
```

## Implementation Steps

1. [ ] Create repository `ayo-plugins-whatsapp`
2. [ ] Implement whatsmeow client wrapper
3. [ ] Implement QR code pairing flow
4. [ ] Implement device/session persistence
5. [ ] Create message trigger with filtering
6. [ ] Implement media download handling
7. [ ] Implement whatsapp-send tool
8. [ ] Implement whatsapp-reply tool
9. [ ] Implement whatsapp-media tool
10. [ ] Implement whatsapp-contacts tool
11. [ ] Create @whatsapp-bot agent
12. [ ] Handle reconnection and session recovery
13. [ ] Write documentation
14. [ ] Add tests

## Dependencies

- Depends on: `ayo-pltg` (trigger plugin architecture)
- Go libraries:
  - `github.com/tulir/whatsmeow` - WhatsApp Web protocol
  - `go.mau.fi/whatsmeow/store/sqlstore` - Session storage

## Security Considerations

- Session stored locally (E2E encrypted)
- JID allowlist for message processing
- Media downloads to sandboxed directory
- No message content logged by default
- Follows WhatsApp ToS (personal use)

## Limitations

- Uses personal WhatsApp account
- WhatsApp may update protocol (library maintained)
- Not for high-volume/business use
- One device per trigger instance

## Setup Instructions

1. Install plugin: `ayo plugin install ayo-plugins-whatsapp`
2. Configure trigger in ayo.json
3. Run `ayo trigger start whatsapp`
4. Scan QR code with WhatsApp mobile app
5. Start chatting!

---

*Created: 2026-02-23*
