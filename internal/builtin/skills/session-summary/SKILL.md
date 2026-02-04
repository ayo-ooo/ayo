---
name: session-summary
description: Techniques for summarizing conversation sessions. Use when creating session summaries, handoff notes, or context for resuming conversations.
compatibility: Works with all agents that have session persistence enabled
metadata:
  author: ayo
  version: "1.0"
---

# Session Summary Skill

Generate effective session summaries for conversation continuity and context handoffs.

## When to Use

Activate this skill when:
- Conversation is ending and needs a summary for later resumption
- Context window is running low and handoff is needed
- User asks for a summary of what was accomplished
- Creating notes for another agent or session to continue work
- Reviewing what happened in a previous session

## Summary Components

A good session summary includes:

### 1. Current State

What is the state right now?

```
### Current State
- Active branch: feature/new-api
- All tests passing
- 3 of 5 endpoints implemented
```

### 2. What Was Accomplished

Concrete outcomes, not process descriptions:

```
### Accomplished
- Implemented user authentication with JWT tokens
- Added rate limiting middleware (100 req/min)
- Created OpenAPI spec for auth endpoints
- Fixed race condition in session cleanup
```

### 3. What Remains

Outstanding work, ordered by priority:

```
### Remaining Work
1. Implement password reset flow (blocked on email service config)
2. Add refresh token rotation
3. Write integration tests for auth flow
4. Update API documentation
```

### 4. Key Context

Information the next session needs:

```
### Key Context
- Using `github.com/golang-jwt/jwt/v5` for JWT handling
- Token expiry is 15 minutes, refresh token is 7 days
- All auth routes are under `/api/v1/auth/`
- Password hashing uses bcrypt with cost factor 12
```

### 5. Blockers and Decisions

What's blocking progress and what decisions were made:

```
### Blockers
- Email service not configured (need SMTP credentials)
- Waiting on security review before enabling password reset

### Decisions Made
- Chose JWT over session cookies for API compatibility
- Rate limits apply per-user, not per-IP
```

## Summary Formats

### Brief Summary (default)

For quick context restoration:

```markdown
# Session Summary

**Task**: Implement user authentication
**Status**: In progress (60%)

## Completed
- JWT auth endpoints (login, logout, verify)
- Rate limiting middleware

## Next Steps
1. Password reset flow
2. Refresh token rotation

## Context
- Using golang-jwt/v5
- Tokens expire in 15 minutes
```

### Detailed Summary

For complex multi-session work:

```markdown
# Session Summary - Auth Implementation

## Current State
Branch: feature/auth (3 commits ahead of main)
Tests: 47 passing, 0 failing
Coverage: 78%

## Accomplished This Session
1. **JWT Authentication** (internal/auth/jwt.go)
   - Login endpoint with credential validation
   - Token generation with configurable expiry
   - Middleware for protected routes

2. **Rate Limiting** (internal/middleware/ratelimit.go)
   - Per-user rate limiting (100 req/min)
   - Redis-backed for distributed deployments
   - Configurable via environment variables

3. **Bug Fixes**
   - Fixed race condition in session cleanup (#123)
   - Corrected token expiry timezone handling

## Remaining Work
| Priority | Task | Blocked By |
|----------|------|------------|
| P0 | Password reset | Email config |
| P1 | Refresh rotation | - |
| P2 | Integration tests | - |

## Technical Context
- JWT library: github.com/golang-jwt/jwt/v5
- Token format: HS256, 15min access, 7d refresh
- Routes: /api/v1/auth/{login,logout,verify,refresh}
- Password: bcrypt, cost=12

## Files Modified
- internal/auth/jwt.go (new)
- internal/auth/middleware.go (new)
- internal/middleware/ratelimit.go (new)
- cmd/api/main.go (updated routes)

## Commands That Work
```bash
go test ./internal/auth/...
curl -X POST localhost:8080/api/v1/auth/login -d '{"email":"...", "password":"..."}'
```
```

### Handoff Summary

For context window exhaustion or agent delegation:

```markdown
# Handoff Summary

## Immediate Context
Currently implementing password reset flow. The email service
configuration is blocking progress.

## Resume Instructions
1. Configure SMTP credentials in .env
2. Complete internal/auth/reset.go
3. Add reset endpoint to router
4. Write tests

## Critical Information
- Reset tokens are separate from auth tokens (24h expiry)
- Reset flow: request -> email -> verify -> new password
- Must rate limit reset requests (3/hour per email)

## Files to Review
- internal/auth/reset.go (in progress)
- internal/email/sender.go (needs SMTP config)
```

## Best Practices

### Be Specific

Bad: "Worked on authentication"
Good: "Implemented JWT login endpoint with 15-minute token expiry"

### Use File References

Include file paths with line numbers when helpful:
- `internal/auth/jwt.go:45` - Token generation
- `internal/auth/middleware.go:78` - Auth middleware

### Capture Decisions

Document why, not just what:
- "Chose JWT over sessions because API will be used by mobile apps"
- "Using bcrypt cost=12 as compromise between security and latency"

### Note Blockers Clearly

Distinguish between:
- **Hard blockers**: Cannot proceed without external action
- **Soft blockers**: Could work around but waiting for clarity

### Include Working Commands

Save time on resumption:
```bash
# Build and test
go build ./... && go test ./...

# Run specific test
go test -v ./internal/auth/... -run TestLogin

# Manual testing
curl -X POST localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com", "password":"secret"}'
```

## Session Files

Sessions are stored as JSONL files:
- Location: `~/.local/share/ayo/sessions/{agent}/{month}/{id}.jsonl`
- Format: One JSON object per line (header + messages)
- Index: SQLite for fast search and listing

### Viewing Past Sessions

```bash
# List recent sessions
ayo sessions list

# Show session details
ayo sessions show <id>

# Continue a session
ayo sessions continue <id>
```

### Session Metadata

When summarizing, note session metadata:
- Session ID (for reference)
- Start time and duration
- Agent used
- Number of messages/turns

## Automatic Summarization

Some scenarios trigger automatic summary generation:
- Context window approaching limit
- Long-running sessions (>50 turns)
- Before agent handoff
- Session explicitly ended by user

## Integration with Memory

Good session summaries can become memories:
- Project-specific patterns discovered
- User preferences expressed
- Corrections made to agent behavior

Use the memory skill to persist important insights beyond the session.
