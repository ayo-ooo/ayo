# Tutorial: Multi-Agent Coordination with Squads

Build a development squad where frontend and backend agents collaborate on features. By the end, you'll understand how agents coordinate through tickets and shared workspaces.

**Time**: ~30 minutes  
**Prerequisites**: [First Agent Tutorial](first-agent.md) complete

## What You'll Build

A `dev-team` squad with:
- `@backend` - Implements API endpoints
- `@frontend` - Builds UI components
- Ticket-based coordination between agents

## Step 1: Create the Squad

```bash
ayo squad create dev-team
```

This creates the squad directory at `~/.local/share/ayo/sandboxes/squads/dev-team/`:

```
dev-team/
├── SQUAD.md           # Team constitution
├── ayo.json           # Squad configuration
├── workspace/         # Shared code workspace
└── .tickets/          # Coordination tickets
```

## Step 2: Create the Agents

Create specialized agents for the squad:

```bash
# Backend agent
ayo agents create @backend

# Frontend agent  
ayo agents create @frontend
```

### Configure @backend

Edit `~/.config/ayo/agents/@backend/system.md`:

```markdown
# Backend Developer

You are a backend developer specializing in Go APIs.

## Responsibilities
- Design and implement REST API endpoints
- Write database queries and migrations
- Create unit tests for business logic
- Document API endpoints

## Guidelines
- Use standard library when possible
- Follow Go idioms and best practices
- Return proper HTTP status codes
- Include error handling and logging

## When working in a squad
- Create tickets for frontend integration points
- Document request/response schemas
- Coordinate timing with @frontend
```

### Configure @frontend

Edit `~/.config/ayo/agents/@frontend/system.md`:

```markdown
# Frontend Developer

You are a frontend developer specializing in React/TypeScript.

## Responsibilities
- Build React components
- Integrate with backend APIs
- Write component tests
- Ensure responsive design

## Guidelines
- Use TypeScript for type safety
- Follow React hooks best practices
- Handle loading and error states
- Keep components small and focused

## When working in a squad
- Wait for API endpoints before integration
- Create tickets for UI/UX decisions
- Coordinate with @backend on data formats
```

## Step 3: Add Agents to the Squad

```bash
ayo squad add-agent dev-team @backend
ayo squad add-agent dev-team @frontend
```

Verify agents are added:

```bash
ayo squad show dev-team
```

## Step 4: Write the Constitution

Edit `~/.local/share/ayo/sandboxes/squads/dev-team/SQUAD.md`:

```markdown
---
name: dev-team
planners:
  near_term: ayo-todos
  long_term: ayo-tickets
agents:
  - "@backend"
  - "@frontend"
---

# Squad: dev-team

## Mission

Build features efficiently through coordinated frontend and backend development.

## Agents

### @backend
The backend developer handles:
- API design and implementation
- Database operations
- Server-side business logic
- API documentation

### @frontend
The frontend developer handles:
- UI component development
- API integration
- User experience
- Responsive design

## Workflow

1. **Planning**: Break feature into backend and frontend tasks
2. **API First**: @backend implements and documents API endpoints
3. **Integration**: @frontend builds UI using the API
4. **Review**: Both agents review each other's work
5. **Testing**: Integration testing across the stack

## Coordination Rules

- Create tickets for all tasks
- @backend creates "api-ready" tickets when endpoints are complete
- @frontend waits for "api-ready" before starting integration
- Use the shared `/workspace/` for all code
- Document decisions in tickets

## Communication

When handing off work:
1. Create a ticket with clear requirements
2. Include example request/response if API-related
3. Note any blockers or dependencies
4. Update ticket status when complete
```

## Step 5: Start the Squad

```bash
ayo squad start dev-team
```

This creates an isolated sandbox where both agents will work.

## Step 6: Send a Task

Dispatch a feature request to the squad:

```bash
ayo "#dev-team" "Build a user registration feature with email/password signup"
```

The squad will:
1. Plan the work and create tickets
2. Route tasks to appropriate agents
3. Coordinate through ticket handoffs

## Step 7: Monitor Progress

### View Tickets

```bash
ayo squad ticket dev-team list
```

Example output:
```
ID         STATUS        ASSIGNEE    TITLE
reg-001    in_progress   @backend    Create POST /api/users endpoint
reg-002    blocked       @frontend   Build registration form (deps: reg-001)
reg-003    open          @backend    Add email verification
```

### View Ready Tickets

```bash
ayo squad ticket dev-team ready
```

Shows tickets with no blockers that can be worked on.

### Shell Into the Sandbox

Inspect the squad's workspace:

```bash
ayo squad shell dev-team
```

Navigate to see the code:
```bash
ls /workspace/
cat .tickets/*.md
```

## Step 8: Continue Development

Send follow-up tasks:

```bash
ayo "#dev-team" "Add password strength validation"
ayo "#dev-team" @frontend "Make the form mobile-responsive"
```

The `@agent` syntax routes directly to a specific agent.

## Understanding Ticket Flow

### Ticket Lifecycle

```
open → in_progress → review → closed
         ↓
       blocked (waiting on dependency)
```

### Ticket Dependencies

When @backend creates a ticket:

```markdown
---
id: api-001
status: closed
assignee: "@backend"
---
# POST /api/users endpoint

Endpoint is ready at POST /api/users
Request: { email, password }
Response: { id, email, token }
```

The frontend ticket can now unblock:

```markdown
---
id: ui-001
status: in_progress
assignee: "@frontend"
deps: [api-001]
---
# Registration form

Integrate with /api/users endpoint...
```

## Complete Example

Your squad structure:

```
~/.local/share/ayo/sandboxes/squads/dev-team/
├── SQUAD.md
├── ayo.json
├── workspace/
│   ├── backend/
│   │   └── api/
│   └── frontend/
│       └── components/
├── .tickets/
│   ├── reg-001.md
│   └── reg-002.md
└── .context/
    └── session.json
```

## Troubleshooting

### Squad not starting

```bash
# Check daemon status
ayo sandbox service status

# Check squad status
ayo squad show dev-team

# View logs
tail -f ~/.local/share/ayo/daemon.log
```

### Agent not receiving tasks

Verify the agent is added to the squad:

```bash
ayo squad show dev-team | grep agents
```

Check SQUAD.md frontmatter includes the agent:

```yaml
agents:
  - "@backend"
  - "@frontend"
```

### Tickets not syncing

Tickets are stored in `.tickets/` inside the squad sandbox:

```bash
ayo squad shell dev-team
ls .tickets/
```

## Next Steps

- [Triggers](triggers.md) - Auto-run squad on events
- [Memory](memory.md) - Share context across sessions
- [Plugins](plugins.md) - Add custom tools to your squad

---

*You've built a multi-agent squad! Continue to [Triggers](triggers.md).*
