---
id: ase-hkao
status: closed
deps: [ase-sjha]
links: []
created: 2026-02-09T03:03:53Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-jep0
---
# Add user_sessions and user_messages tables

Create SQLite tables to store user conversations with agents. These are direct user-to-agent messages (not inter-agent coordination).

## Background

User conversations are stored in SQLite because:
- They're the user's personal data, searchable for memory context
- Different from agent-agent coordination (which lives in Matrix)
- Enable queries like 'what did I ask @researcher last week?'

## Schema

```sql
CREATE TABLE user_sessions (
  id TEXT PRIMARY KEY,
  agent_handle TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_messages (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES user_sessions(id),
  role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_sessions_agent ON user_sessions(agent_handle);
CREATE INDEX idx_user_messages_session ON user_messages(session_id);
```

## Implementation

1. Add migration file in internal/database/migrations/
2. Add Go types in internal/database/models/
3. Add repository methods: CreateSession, AddMessage, GetSessionMessages, SearchMessages
4. Wire into existing session handling code

## Files to modify

- internal/database/migrations/NNN_user_sessions.sql (new)
- internal/database/models/session.go (new or extend)
- internal/database/repository.go (add methods)

## Acceptance Criteria

- Migration creates tables successfully
- Repository methods work correctly
- Existing session code uses new tables
- Index on agent_handle for efficient queries

