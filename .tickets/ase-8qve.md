---
id: ase-8qve
status: closed
deps: [ase-95o4, ase-rzhr]
links: []
created: 2026-02-06T04:09:14Z
type: epic
priority: 1
assignee: Alex Cabrera
parent: ase-95o4
---
# Trigger System

Implement the trigger system for autonomous agent execution based on cron, file changes, webhooks, and IRC events.

## Design

## Trigger Types
1. cron - time-based scheduling
2. watch - host filesystem changes
3. webhook - HTTP endpoints
4. irc - IRC channel patterns

## Configuration
Triggers defined in:
- Agent config.json
- Flow frontmatter
- Global ayo.json

## Trigger Lifecycle
1. Daemon watches for trigger conditions
2. Condition met → spawn agent session
3. Inject trigger context into prompt
4. Agent executes autonomously
5. Session ends, results logged

## CLI Commands
- ayo triggers list
- ayo triggers add
- ayo triggers remove <id>
- ayo triggers history

## Acceptance Criteria

- Cron triggers fire on schedule
- File watch triggers on host changes
- Webhook triggers via HTTP POST
- IRC triggers on message patterns
- Trigger history tracked

