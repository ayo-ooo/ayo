---
id: ase-alok
status: closed
deps: [ase-fb0m]
links: []
created: 2026-02-06T04:09:55Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-ka3q
---
# Install and configure ngircd in sandbox

Install ngircd IRC server in the Alpine sandbox and configure it to start automatically. This enables inter-agent communication.

## Design

## ngircd Setup
1. Install via apk: 'apk add ngircd'
2. Configure /etc/ngircd/ngircd.conf with:
   - Listen on localhost:6667
   - No password required
   - Logging to /var/log/irc/
   - Auto-create channels

## Startup
Add ngircd to sandbox startup sequence. Options:
- Add to SetupCommands in Create()
- Use a startup script that runs on container start
- Supervisord or similar (may be overkill)

## Channels
- #general - all agents, broadcasts
- #session-{id} - created per session
- Private messages via /msg @agent

## IRC Log Directory
/var/log/irc/ - persisted for git sync and user access

## Acceptance Criteria

- ngircd starts when sandbox starts
- Agents can connect via nc or irssi
- Logs written to /var/log/irc/
- Server survives container restarts

