---
id: ayo-ptgram
status: closed
deps: [ayo-pltg]
links: []
created: 2026-02-23T12:00:00Z
type: epic
priority: 1
assignee: Alex Cabrera
tags: [plugins, triggers, external-repo, chat]
---
# Epic: Telegram Bot Trigger Plugin

## Summary

Create `ayo-plugins-telegram` - a trigger plugin enabling agents to respond to Telegram messages. Telegram's Bot API is well-documented, free to use, and provides excellent real-time message delivery via long-polling or webhooks.

## The Dream

Picture this: You're on your phone, away from your desk. You open Telegram and message your agent like you'd message a friend:

> **You**: "What's the status of the production deploy?"
> **@ayo**: "Deploy completed 20 minutes ago. All health checks passing. 3 new features shipped: user profiles, dark mode, and export improvements."

> **You**: "Any issues in my GitHub notifications?"
> **@ayo**: "2 PRs need your review. 1 failing CI on the auth-refactor branch - looks like a test timeout. Want me to investigate?"

> **You**: "Yes, check the failing test"
> **@ayo**: "Found it. The test waits 30s for a response that now takes 45s due to the new validation. I can submit a PR to increase the timeout - want me to?"

This is **ambient intelligence** - your agent is always available, always context-aware, always ready to help. Not through a CLI, not through a web interface, but through the messaging app you already use every day.

## Why Telegram

- **Free API** - No cost for bot usage
- **Excellent documentation** - Clear, comprehensive Bot API
- **Real-time** - Long polling or webhook support
- **Rich features** - Inline keyboards, file sharing, groups
- **Privacy options** - Can use without phone number association
- **Cross-platform** - Desktop, mobile, web clients

## Use Cases

1. **Personal Assistant** - Chat with @ayo via Telegram
2. **Team Notifications** - Post updates to group chats
3. **Command Interface** - Run ayo commands via /commands
4. **File Processing** - Send files to agent via Telegram
5. **Alert Handler** - Receive and process alerts
6. **Remote Agent Control** - Start/stop agents remotely

## Plugin Components

```
ayo-plugins-telegram/
├── manifest.json
├── triggers/
│   └── telegram/
│       ├── trigger.json
│       └── telegram-trigger       # Binary
├── tools/
│   ├── telegram-send/
│   ├── telegram-reply/
│   ├── telegram-photo/
│   └── telegram-file/
├── agents/
│   └── @telegram-bot/
└── skills/
    └── telegram-assistant.md
```

## Trigger Specification

### Configuration Schema

```json
{
  "type": "telegram",
  "config": {
    "bot_token": {
      "type": "string",
      "required": true,
      "secret": true,
      "env": "TELEGRAM_BOT_TOKEN"
    },
    "mode": {
      "type": "string",
      "enum": ["polling", "webhook"],
      "default": "polling"
    },
    "webhook_url": {
      "type": "string",
      "description": "Required if mode=webhook"
    },
    "allowed_users": {
      "type": "array",
      "items": { "type": "integer" },
      "description": "Telegram user IDs allowed to interact"
    },
    "allowed_chats": {
      "type": "array",
      "items": { "type": "integer" },
      "description": "Telegram chat IDs to listen to"
    },
    "commands": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Bot commands to register"
    },
    "filter": {
      "type": "object",
      "properties": {
        "message_types": {
          "type": "array",
          "items": { "type": "string" },
          "default": ["text", "photo", "document"]
        },
        "ignore_bots": { "type": "boolean", "default": true }
      }
    }
  }
}
```

### Event Payload

```json
{
  "event_type": "telegram.message",
  "message_id": 12345,
  "chat": {
    "id": -100123456789,
    "type": "group",
    "title": "My Team"
  },
  "from": {
    "id": 987654321,
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "is_bot": false
  },
  "date": "2026-02-23T10:30:00Z",
  "text": "Hey @ayobot, what's the status of the deployment?",
  "entities": [
    { "type": "mention", "offset": 4, "length": 7 }
  ],
  "reply_to_message": null,
  "is_command": false,
  "command": null,
  "command_args": null
}
```

### Command Event

```json
{
  "event_type": "telegram.command",
  "command": "status",
  "command_args": "production",
  "message_id": 12346,
  "chat": { "id": -100123456789, "type": "group" },
  "from": { "id": 987654321, "username": "johndoe" }
}
```

### File Event

```json
{
  "event_type": "telegram.document",
  "message_id": 12347,
  "document": {
    "file_id": "BAADBAADZQADjlcAAVJGLbpGcLPXFAI",
    "file_unique_id": "AgADZQADjlcAAVI",
    "file_name": "report.pdf",
    "mime_type": "application/pdf",
    "file_size": 102400,
    "local_path": "/tmp/telegram/report.pdf"
  },
  "caption": "Please analyze this report"
}
```

## Tool Specifications

### telegram-send

```json
{
  "name": "telegram-send",
  "description": "Send a message to a Telegram chat",
  "parameters": {
    "chat_id": { "type": "integer", "required": true },
    "text": { "type": "string", "required": true },
    "parse_mode": { "type": "string", "enum": ["Markdown", "HTML"], "default": "Markdown" },
    "reply_to": { "type": "integer", "description": "Message ID to reply to" },
    "disable_notification": { "type": "boolean", "default": false }
  }
}
```

### telegram-reply

```json
{
  "name": "telegram-reply",
  "description": "Reply to the current message (in trigger context)",
  "parameters": {
    "text": { "type": "string", "required": true },
    "parse_mode": { "type": "string", "default": "Markdown" }
  }
}
```

### telegram-photo

```json
{
  "name": "telegram-photo",
  "description": "Send a photo to a Telegram chat",
  "parameters": {
    "chat_id": { "type": "integer", "required": true },
    "photo": { "type": "string", "required": true, "description": "File path or URL" },
    "caption": { "type": "string" }
  }
}
```

### telegram-file

```json
{
  "name": "telegram-file",
  "description": "Send a file to a Telegram chat",
  "parameters": {
    "chat_id": { "type": "integer", "required": true },
    "file": { "type": "string", "required": true },
    "caption": { "type": "string" }
  }
}
```

## Agent: @telegram-bot

```markdown
# @telegram-bot

You are a Telegram bot assistant.

## Capabilities

- Respond to text messages
- Process files sent to you
- Execute commands
- Send rich messages with Markdown formatting
- Share files and images

## Guidelines

1. Keep responses concise (Telegram prefers short messages)
2. Use Markdown for formatting
3. Respond to mentions in group chats
4. Handle /commands with clear responses
5. Acknowledge file receipt before processing
6. Respect rate limits (30 messages/second)
```

## Implementation Steps

1. [ ] Create repository `ayo-plugins-telegram`
2. [ ] Implement Telegram Bot API client
3. [ ] Implement long-polling update receiver
4. [ ] Implement webhook receiver (optional)
5. [ ] Create message trigger with filtering
6. [ ] Implement file download handling
7. [ ] Implement telegram-send tool
8. [ ] Implement telegram-reply tool
9. [ ] Implement telegram-photo tool
10. [ ] Implement telegram-file tool
11. [ ] Create @telegram-bot agent
12. [ ] Add command registration
13. [ ] Write documentation
14. [ ] Add integration tests with Telegram test bot

## Dependencies

- Depends on: `ayo-pltg` (trigger plugin architecture)
- Go libraries:
  - `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram Bot API

## Security Considerations

- Bot token stored in environment variable
- User/chat allowlist for sensitive operations
- File downloads to sandboxed directory
- Rate limiting to avoid API bans
- No sensitive data logged

## Setup Instructions

1. Create bot via @BotFather on Telegram
2. Get bot token
3. Set `TELEGRAM_BOT_TOKEN` environment variable
4. Configure trigger with allowed users/chats
5. Start trigger

---

*Created: 2026-02-23*
