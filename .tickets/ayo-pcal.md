---
id: ayo-pcal
status: closed
deps: [ayo-pltg]
links: []
created: 2026-02-23T12:00:00Z
type: epic
priority: 2
assignee: Alex Cabrera
tags: [plugins, triggers, external-repo, open-standards]
---
# Epic: CalDAV Calendar Trigger Plugin

## Summary

Create `ayo-plugins-caldav` - a trigger and tool plugin for CalDAV calendar integrations. CalDAV is an open standard (RFC 4791) for calendar access, supported by Google Calendar, Apple Calendar, Fastmail, Nextcloud, and most self-hosted calendar servers.

## Open Standards Focus

This plugin prioritizes the CalDAV open standard over proprietary APIs:
- **CalDAV (RFC 4791)** - Primary protocol for all operations
- **iCalendar (RFC 5545)** - Event format
- Works with any CalDAV-compliant server
- No vendor lock-in

## Use Cases

1. **Meeting Prep** - Generate meeting briefs before scheduled meetings
2. **Standup Automation** - Trigger standup report generation at scheduled times
3. **Calendar Blocking** - Auto-block focus time based on patterns
4. **Meeting Notes** - Draft meeting notes/agendas before meetings
5. **Reminder Agent** - Smart reminders with context

## Plugin Components

```
ayo-plugins-caldav/
├── manifest.json
├── triggers/
│   └── caldav/
│       ├── trigger.json
│       └── caldav-trigger       # Binary
├── tools/
│   ├── calendar-list/
│   ├── calendar-create/
│   ├── calendar-update/
│   └── calendar-search/
├── agents/
│   └── @calendar-assistant/
└── skills/
    └── meeting-prep.md
```

## Supported Servers

Any CalDAV-compliant server:
- **Google Calendar** (CalDAV endpoint)
- **Apple iCloud Calendar**
- **Fastmail**
- **Nextcloud Calendar**
- **Radicale** (self-hosted)
- **Baïkal** (self-hosted)
- **DAViCal** (self-hosted)
- **Zimbra**

## Trigger Specification

### Configuration Schema

```json
{
  "type": "caldav",
  "config": {
    "server_url": {
      "type": "string",
      "required": true,
      "description": "CalDAV server URL"
    },
    "username": {
      "type": "string",
      "required": true
    },
    "password": {
      "type": "string",
      "required": true,
      "secret": true,
      "env": "CALDAV_PASSWORD"
    },
    "calendars": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Calendar paths to watch (default: all)"
    },
    "trigger_before": {
      "type": "duration",
      "default": "15m",
      "description": "Trigger this much before event start"
    },
    "poll_interval": {
      "type": "duration",
      "default": "5m",
      "description": "How often to check for changes"
    },
    "filter": {
      "type": "object",
      "properties": {
        "title_contains": { "type": "string" },
        "has_attendees": { "type": "boolean" },
        "is_recurring": { "type": "boolean" }
      }
    }
  }
}
```

### Event Payload

```json
{
  "event_type": "calendar.event_starting",
  "event_uid": "abc123@example.com",
  "calendar_path": "/calendars/user/default/",
  "title": "Weekly Team Sync",
  "description": "Weekly sync to discuss progress...",
  "start": "2026-02-23T14:00:00Z",
  "end": "2026-02-23T15:00:00Z",
  "location": "Conference Room A",
  "organizer": {
    "name": "Jane Smith",
    "email": "jane@company.com"
  },
  "attendees": [
    { "name": "John Doe", "email": "john@company.com", "partstat": "ACCEPTED" }
  ],
  "is_recurring": true,
  "rrule": "FREQ=WEEKLY;BYDAY=MO",
  "raw_icalendar": "BEGIN:VCALENDAR\n..."
}
```

## Tool Specifications

### calendar-list

```json
{
  "name": "calendar-list",
  "description": "List upcoming calendar events via CalDAV",
  "parameters": {
    "calendar_path": { "type": "string", "description": "Calendar path (default: all)" },
    "start": { "type": "string", "description": "ISO 8601 datetime" },
    "end": { "type": "string" },
    "limit": { "type": "integer", "default": 10 }
  }
}
```

### calendar-create

```json
{
  "name": "calendar-create",
  "description": "Create a calendar event via CalDAV",
  "parameters": {
    "calendar_path": { "type": "string" },
    "title": { "type": "string", "required": true },
    "description": { "type": "string" },
    "start": { "type": "string", "required": true },
    "end": { "type": "string", "required": true },
    "location": { "type": "string" },
    "attendees": { "type": "array", "items": { "type": "string" } },
    "reminders": { "type": "array", "items": { "type": "integer" } }
  }
}
```

## Agent: @calendar-assistant

```markdown
# @calendar-assistant

You help manage calendar and meeting preparation using CalDAV.

## Capabilities

- Query upcoming events from any CalDAV server
- Create, update, delete events
- Generate meeting briefs
- Find available time slots
- Works with self-hosted and cloud calendars

## Guidelines

1. Respect existing commitments when scheduling
2. Consider time zones (use ISO 8601)
3. Default meeting length: 30 minutes
4. Preserve iCalendar data when updating
5. Add buffer time between back-to-back meetings
```

## Implementation Steps

1. [ ] Create repository `ayo-plugins-caldav`
2. [ ] Implement CalDAV client using Go (github.com/emersion/go-webdav)
3. [ ] Implement calendar discovery (PROPFIND)
4. [ ] Implement event listing (REPORT calendar-query)
5. [ ] Implement event creation (PUT)
6. [ ] Create trigger with pre-event timing
7. [ ] Parse and generate iCalendar format
8. [ ] Implement calendar-list tool
9. [ ] Implement calendar-create tool
10. [ ] Implement calendar-update tool
11. [ ] Create @calendar-assistant agent
12. [ ] Create meeting-prep skill
13. [ ] Write documentation
14. [ ] Add integration tests with test CalDAV server (Radicale)

## Dependencies

- Depends on: `ayo-pltg` (trigger plugin architecture)
- Go libraries:
  - `github.com/emersion/go-webdav` - CalDAV client
  - `github.com/emersion/go-ical` - iCalendar parsing

## Security Considerations

- Passwords stored in environment variables
- TLS required for all connections
- Support app-specific passwords for cloud providers
- No OAuth required - CalDAV uses HTTP Basic Auth

---

*Created: 2026-02-23*
