---
id: ase-d2fb
status: closed
deps: [ase-ji7h]
links: []
created: 2026-02-06T04:11:32Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-8qve
---
# Add webhook server to daemon

Add HTTP webhook server to daemon for external triggers like GitHub, CI systems, etc.

## Design

## Webhook Server
Lightweight HTTP server for receiving webhook payloads.

## Endpoints
POST /hooks/{trigger-id} - Trigger specific hook
POST /hooks/github - GitHub webhook handler
POST /hooks/generic - Generic JSON payload

## Configuration
In trigger definition:
triggers:
  - webhook: /hooks/my-trigger
    agent: @ayo
    prompt: 'Webhook received: {payload}'

## Security
- Optional secret token validation (HMAC)
- Localhost-only by default
- Configurable bind address

## Implementation
1. Add webhookServer to daemon
2. Start HTTP listener on daemon start
3. Route to trigger engine on POST
4. Parse common webhook formats (GitHub, GitLab, generic)

## Integration with Triggers
Webhook triggers register with webhook server on daemon start.
On POST, look up trigger by path, execute.

## Port Configuration
Default: random available port
Store in daemon status for CLI to query

## Acceptance Criteria

- HTTP server starts with daemon
- Webhook triggers can be registered
- GitHub/GitLab webhooks parsed
- Triggers fire on valid POST

