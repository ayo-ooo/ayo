---
id: ase-odwt
status: closed
deps: [ase-alok]
links: []
created: 2026-02-06T04:11:12Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-rzhr
---
# Add IRC bridge to daemon

The daemon should monitor IRC logs from the sandbox and can bridge messages to external systems.

## Design

## IRC Bridge Components
1. IRCLogWatcher - watches /var/log/irc/ in sandbox
2. MessageParser - parses IRC log lines
3. NotificationRouter - routes important messages

## Use Cases
1. @-mention routing: Detect @agent mentions, trigger agent sessions
2. Urgency detection: High-priority messages can notify user
3. External bridge: Optional webhook for external notifications

## Implementation
1. Watch sandbox's /var/log/irc/ directory via fsnotify
2. Tail new log files as they appear
3. Parse messages looking for patterns:
   - @{agent} mentions
   - Priority markers
   - Trigger patterns
4. Route accordingly:
   - Agent mention → inject into agent's next session
   - User mention → push notification (future)
   - Trigger match → execute trigger

## Message Injection
When agent has pending messages from IRC:
- Store in daemon memory
- On agent session start, inject as context
- 'You have messages: ...'

## Future Enhancement
Could also write to IRC on agent's behalf if daemon has IRC client.

## Acceptance Criteria

- IRC logs monitored in real-time
- @agent mentions detected
- Pending messages injected at session start
- Optional webhook for external notification

