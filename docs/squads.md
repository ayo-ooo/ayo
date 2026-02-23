# Squads

Squads are isolated sandbox environments where multiple agents collaborate on shared work. Each squad provides a dedicated workspace with filesystem isolation, shared context, and a ticket-based coordination system.

## Overview

A squad is:
- **An isolated sandbox**: A container environment where agents execute commands
- **A shared workspace**: All agents in a squad see the same files at `/workspace`
- **A coordination space**: Agents use tickets to coordinate who does what
- **A team context**: The `SQUAD.md` file defines the team's purpose and each agent's role

```
~/.local/share/ayo/sandboxes/squads/{name}/
├── SQUAD.md               # Team constitution (shared context for all agents)
├── .tickets/              # Coordination tickets
│   ├── alpha-a1b2.md
│   └── alpha-c3d4.md
├── .context/              # Additional context files
│   └── project-brief.md
├── workspace/             # Shared code workspace
│   └── ...
└── agent-homes/           # Per-agent home directories
    ├── frontend/
    └── backend/
```

## SQUAD.md: The Team Constitution

The `SQUAD.md` file is the central document that defines a squad's identity and operating principles. Every agent in the squad receives this file in their system prompt, ensuring all team members share the same understanding of:

- **Mission**: What the squad is trying to accomplish
- **Context**: Background information, constraints, technical decisions
- **Roles**: What each agent is responsible for
- **Coordination**: How agents should work together

### File Location

```
~/.local/share/ayo/sandboxes/squads/{name}/SQUAD.md
```

The file lives in the squad's root directory, making it:
- **Visible to all agents** inside the sandbox
- **Easy to inspect** from the host system
- **Synced automatically** when changes occur

### File Format

SQUAD.md files support optional YAML frontmatter for configuration settings. The frontmatter is delimited by `---` markers:

#### Frontmatter Fields

| Field | Description | Default |
|-------|-------------|---------|
| `planners.near_term` | Near-term planner plugin for session-scoped work tracking | `ayo-todos` |
| `planners.long_term` | Long-term planner plugin for persistent coordination | `ayo-tickets` |
| `lead` | Agent handle that serves as squad lead | `@ayo` |
| `input_accepts` | Agent handle that receives direct input | Same as `lead` |

```markdown
---
planners:
  near_term: custom-todos
  long_term: custom-tickets
lead: "@architect"
input_accepts: "@frontend"
---
# Squad: {name}

## Mission

{One paragraph describing what this squad is trying to accomplish.}

## Context

{Background information all agents need: project constraints, technical decisions,
external dependencies, deadlines, or any shared knowledge.}

## Agents

### @{agent-handle}
**Role**: {Brief role description}
**Responsibilities**:
- {Specific responsibility 1}
- {Specific responsibility 2}

### @{another-agent}
**Role**: {Brief role description}
**Responsibilities**:
- {Specific responsibility 1}
- {Specific responsibility 2}

## Coordination

{How agents should work together: handoff protocols, communication patterns,
dependency chains, blocking rules.}

## Guidelines

{Specific rules or preferences for this squad: coding style, testing requirements,
commit conventions, review process.}
```

### Example SQUAD.md

```markdown
# Squad: ecommerce-auth

## Mission

Implement a secure authentication system for the e-commerce platform, including
user registration, login, password reset, and OAuth integration.

## Context

- **Framework**: Using Express.js with TypeScript
- **Database**: PostgreSQL with Prisma ORM
- **Security**: Must comply with OWASP guidelines
- **Timeline**: MVP by end of sprint (2 weeks)
- **Existing code**: Basic Express skeleton in `/workspace/src`

## Agents

### @backend
**Role**: Backend implementation
**Responsibilities**:
- Implement auth endpoints (register, login, logout, refresh)
- Design and implement database schema
- Write integration tests
- Handle security considerations (password hashing, token validation)

### @frontend
**Role**: Frontend implementation
**Responsibilities**:
- Implement login/register UI components
- Handle auth state management
- Integrate with backend API
- Write component tests

### @qa
**Role**: Quality assurance
**Responsibilities**:
- Review code changes from @backend and @frontend
- Write end-to-end tests
- Test edge cases and error handling
- Verify security requirements

## Coordination

1. **@backend** completes endpoints first, creates ticket when API is ready
2. **@frontend** waits for API ticket, then implements UI
3. **@qa** reviews changes after each agent completes their work
4. Use ticket dependencies to enforce ordering:
   - `frontend-login` depends on `backend-login`
   - `qa-review` depends on both

## Guidelines

- All code must have tests before closing tickets
- Use conventional commit messages
- Security-sensitive changes need @qa review
- Document API endpoints in `API.md`
```

### How SQUAD.md is Loaded

When an agent starts a session within a squad:

1. The system reads `SQUAD.md` from the squad directory
2. The content is injected into the agent's system prompt as a `<squad_context>` block
3. This happens before the agent's own persona and skills are loaded
4. Changes to `SQUAD.md` affect new sessions (existing sessions keep their original context)

**System prompt structure:**
```
<environment_context>
[Working directory, date, git status...]
</environment_context>

<squad_context>
[Contents of SQUAD.md]
</squad_context>

<agent_persona>
[Agent's own system prompt from agent.md]
</agent_persona>

<skills>
[Agent's enabled skills]
</skills>
```

## Creating a Squad

```bash
# Basic squad
ayo squad create myteam

# With specific agents
ayo squad create myteam --agents @frontend,@backend,@qa

# With workspace mount
ayo squad create myteam --workspace /path/to/project

# With output directory
ayo squad create myteam --output /path/to/output
```

After creation, edit the `SQUAD.md` file to define your team:

```bash
# View the file
cat ~/.local/share/ayo/sandboxes/squads/myteam/SQUAD.md

# Edit it
$EDITOR ~/.local/share/ayo/sandboxes/squads/myteam/SQUAD.md
```

Or use the CLI to set up the initial constitution:

```bash
# Write initial SQUAD.md from a template
ayo squad init myteam --mission "Build the auth system" \
  --agents @backend,@frontend
```

## The # Symbol: Quick Squad Dispatch

The `#` symbol provides a shorthand for sending work to squads:

```bash
# Send a prompt to a squad
ayo #myteam "implement the login feature"

# Equivalent to
ayo squad dispatch myteam "implement the login feature"
```

This syntax routes the prompt to the squad's designated input handler (configured via the `input_accepts` frontmatter field, defaulting to the squad lead).

### How Dispatch Works

1. **Parse**: The `#squad` prefix identifies the target squad
2. **Route**: The prompt is sent to the agent specified in `input_accepts`
3. **Execute**: The receiving agent processes the prompt within the squad sandbox
4. **Coordinate**: The agent may create tickets, delegate to other agents, or complete the work directly

### Combined Dispatch

You can dispatch to both an agent and a squad:

```bash
# Send to @backend within #myteam squad
ayo @backend #myteam "implement the login API"
```

When both are specified:
- The squad provides the sandbox and context (SQUAD.md)
- The specified agent receives and handles the prompt
- This overrides the squad's `input_accepts` setting for this dispatch

## Squad Lead and @ayo-in-squad

When `@ayo` operates inside a squad, it becomes `@ayo-in-squad` with a modified tool set optimized for coordination rather than direct execution.

### Why a Different Identity?

The squad lead (`@ayo` by default) serves as the orchestrator:
- **Receives dispatches**: Handles prompts sent to the squad without a specific agent target
- **Coordinates work**: Creates tickets, assigns agents, manages dependencies
- **Synthesizes results**: Combines output from multiple agents into cohesive responses

### Restricted Tools

`@ayo-in-squad` has access to:
| Tool Category | Description |
|---------------|-------------|
| **Ticket tools** | Create, assign, close, and manage tickets |
| **Agent dispatch** | Delegate work to squad members |
| **Read-only file access** | View workspace files (not modify directly) |
| **Communication** | Respond to the user, ask clarifying questions |

Notably restricted:
- **No direct file editing**: Must delegate to specialized agents
- **No bash execution**: Must use agents with sandbox access
- **No external tools**: API calls, web fetch, etc. go through agents

This separation ensures the lead coordinates rather than doing everything itself, enabling true multi-agent collaboration.

### Configuring a Different Lead

Set the `lead` field in SQUAD.md frontmatter:

```yaml
---
lead: "@architect"
input_accepts: "@architect"
---
```

The lead agent receives the squad context and coordination tools regardless of their base configuration.

## I/O Schemas

Squads can define input/output contracts using JSON Schema files, enabling type-safe integration with flows and external systems.

### Schema Files

Place schema files in the squad's root directory:

```
~/.local/share/ayo/sandboxes/squads/{name}/
├── SQUAD.md
├── input.jsonschema      # Optional: defines expected input structure
├── output.jsonschema     # Optional: defines output structure
└── ...
```

### Input Schema

Defines the structure of data the squad expects to receive:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "feature_name": {
      "type": "string",
      "description": "Name of the feature to implement"
    },
    "priority": {
      "type": "string",
      "enum": ["low", "medium", "high"],
      "default": "medium"
    },
    "requirements": {
      "type": "array",
      "items": { "type": "string" }
    }
  },
  "required": ["feature_name"]
}
```

### Output Schema

Defines the structure of data the squad produces:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["success", "failure", "partial"]
    },
    "files_changed": {
      "type": "array",
      "items": { "type": "string" }
    },
    "summary": {
      "type": "string"
    }
  },
  "required": ["status", "summary"]
}
```

### Schema Validation

When schemas are present:

1. **Input validation**: Dispatched data is validated against `input.jsonschema` before processing
2. **Output validation**: Squad output is validated against `output.jsonschema` before returning
3. **Error handling**: Validation failures return detailed error messages

### Use with Flows

Schemas enable squads to participate in typed flows:

```yaml
# flow.yaml
name: feature-pipeline
steps:
  - id: plan
    agent: "@planner"
    output_schema: "planning-output.jsonschema"
  
  - id: implement
    squad: "#dev-team"
    input: "${{ steps.plan.output }}"
    # Squad's input.jsonschema validates the planning output
    # Squad's output.jsonschema defines what flows to the next step
  
  - id: review
    agent: "@reviewer"
    input: "${{ steps.implement.output }}"
```

### Schema Commands

```bash
# Validate input against squad schema
ayo squad validate myteam --input data.json

# View squad schemas
ayo squad schema myteam show

# Set schema from file
ayo squad schema myteam set --input schema.json --output out-schema.json
```

## Squad Lifecycle

### Create

```bash
ayo squad create alpha --agents @frontend,@backend
```

Creates:
- Squad sandbox directory
- Empty `SQUAD.md` template
- `.tickets/` directory
- `.context/` directory
- `workspace/` directory

### Start

```bash
ayo squad start alpha
```

Starts the squad's sandbox container. Agents can now be spawned into it.

### Add Work

```bash
# Create tickets for agents
ayo squad ticket alpha create "Implement login API" -a @backend
ayo squad ticket alpha create "Build login form" -a @frontend --deps login-api

# Or use the ticket watcher to auto-assign
ayo squad watch alpha
```

### Monitor

```bash
# List all squads
ayo squad list

# Show squad status
ayo squad status alpha

# List tickets in squad
ayo squad ticket alpha list
```

### Stop

```bash
ayo squad stop alpha
```

Stops the sandbox but preserves all state (workspace, tickets, SQUAD.md).

### Destroy

```bash
ayo squad destroy alpha
```

Removes the sandbox and optionally deletes all data.

## Multi-Squad Membership

An agent can belong to multiple squads simultaneously. Each squad provides different context:

```
@backend agent
├── Squad: ecommerce-auth
│   └── SQUAD.md defines auth-specific responsibilities
├── Squad: payment-integration
│   └── SQUAD.md defines payment-specific responsibilities
└── Squad: api-refactor
    └── SQUAD.md defines refactoring responsibilities
```

When spawned into a squad, the agent receives that squad's `SQUAD.md`. The same agent persona adapts to different team contexts.

## Context Files

The `.context/` directory holds additional files that agents should know about:

```
.context/
├── project-brief.md     # Project overview
├── api-spec.yaml        # API specification
├── architecture.md      # System design
└── decisions/           # Decision records
    ├── 001-auth-flow.md
    └── 002-db-choice.md
```

These files are not automatically injected into system prompts (to avoid context overflow), but agents can read them with file tools.

To add context:

```bash
# Copy a file into squad context
cp ~/docs/api-spec.yaml ~/.local/share/ayo/sandboxes/squads/alpha/.context/

# Or create it directly
echo "# Project Brief" > ~/.local/share/ayo/sandboxes/squads/alpha/.context/brief.md
```

## Syncing SQUAD.md Changes

When you edit `SQUAD.md`:

| When | Effect |
|------|--------|
| **Before squad start** | All agents get updated context |
| **After squad start, new agents** | New agents get updated context |
| **After squad start, running agents** | Running agents keep old context until restart |

To refresh a running agent's context:

```bash
# Stop and restart specific agent
ayo squad agent alpha stop @backend
ayo squad agent alpha start @backend

# Or restart entire squad
ayo squad stop alpha
ayo squad start alpha
```

## File System Permissions

Inside the sandbox:

| Path | Permissions | Purpose |
|------|-------------|---------|
| `/workspace` | Read-write for all agents | Shared code workspace |
| `~/.tickets` | Read-write for all agents | Coordination tickets |
| `/context` | Read-only for all agents | Reference materials |
| `/home/{agent}` | Per-agent read-write | Agent-specific files |

The `SQUAD.md` file is mounted read-only so agents can reference it but cannot modify the team constitution.

## Best Practices

### Keep SQUAD.md Focused

Include:
- Mission (1-2 paragraphs max)
- Key context (technical constraints, deadlines)
- Agent roles (2-3 bullet points per agent)
- Coordination rules (how to hand off work)

Avoid:
- Full API documentation (put in `.context/`)
- Detailed code examples (put in workspace)
- Project history (keep brief or link to docs)

### Define Clear Handoffs

Specify how work flows between agents:

```markdown
## Coordination

1. @backend creates API endpoint
2. @backend creates ticket for @frontend when ready
3. @frontend implements UI
4. @qa reviews after each component is done
```

### Use Ticket Dependencies

Encode handoff requirements in ticket dependencies rather than prose:

```bash
# Frontend can't start until backend is done
ayo squad ticket alpha create "Login form" -a @frontend --deps backend-login-api
```

### Version SQUAD.md

If your workspace is a git repo, consider tracking `SQUAD.md`:

```bash
cd ~/.local/share/ayo/sandboxes/squads/alpha
git init
git add SQUAD.md
git commit -m "Initial squad constitution"
```

## CLI Reference

### Squad Management

| Command | Description |
|---------|-------------|
| `ayo squad create NAME` | Create a new squad |
| `ayo squad list` | List all squads |
| `ayo squad status NAME` | Show squad details |
| `ayo squad start NAME` | Start squad sandbox |
| `ayo squad stop NAME` | Stop squad sandbox |
| `ayo squad destroy NAME` | Remove squad entirely |

### SQUAD.md Management

| Command | Description |
|---------|-------------|
| `ayo squad init NAME` | Generate SQUAD.md template |
| `ayo squad show NAME` | Display SQUAD.md contents |
| `ayo squad edit NAME` | Open SQUAD.md in editor |

### Agent Management

| Command | Description |
|---------|-------------|
| `ayo squad agent NAME add @agent` | Add agent to squad |
| `ayo squad agent NAME remove @agent` | Remove agent from squad |
| `ayo squad agent NAME start @agent` | Start agent session |
| `ayo squad agent NAME stop @agent` | Stop agent session |
| `ayo squad agent NAME list` | List agents in squad |

### Ticket Management

| Command | Description |
|---------|-------------|
| `ayo squad ticket NAME create` | Create ticket in squad |
| `ayo squad ticket NAME list` | List squad tickets |
| `ayo squad ticket NAME show ID` | Show ticket details |
| `ayo squad ticket NAME close ID` | Close ticket |

### Work Coordination with Tickets

Squad agents use ticket tools to coordinate work. When the squad lead (or orchestrator) receives a complex task:

1. **Create tickets** for each work item using `ticket_create`
2. **Assign agents** via the `assignee` field (e.g., `@backend`, `@frontend`)
3. **Track progress** with `ticket_list` to see what's done and what's blocked
4. **Add notes** using `ticket_note` for updates and context
5. **Close tickets** after review with `ticket_close`

**Example workflow:**
```
User: "Build a user registration system"

Squad lead:
1. ticket_create("Implement registration API", assignee: "@backend", priority: 1)
2. ticket_create("Design registration form", assignee: "@frontend", deps: [api-ticket])
3. ticket_create("Write E2E tests", assignee: "@qa", deps: [api-ticket, form-ticket])

Each agent picks up their assigned tickets and works in the shared workspace.
```

**Available ticket tools:**
| Tool | Description |
|------|-------------|
| `ticket_create` | Create a new ticket with title, description, assignee |
| `ticket_list` | List tickets filtered by status or assignee |
| `ticket_start` | Mark ticket as in_progress |
| `ticket_close` | Mark ticket as completed |
| `ticket_block` | Mark ticket as blocked with reason |
| `ticket_note` | Add a note/update to a ticket |

## Comparison to Other Systems

| Feature | Squads | Flows | Delegation |
|---------|--------|-------|------------|
| **Isolation** | Full sandbox | Shared process | Shared process |
| **Parallelism** | Multiple agents | Sequential steps | Single call |
| **Coordination** | Tickets | Chained output | Direct call |
| **Persistence** | Disk workspace | None | None |
| **Context** | SQUAD.md | Flow definition | Agent config |

**Use Squads when:**
- Multiple agents need to work on shared files
- Work is parallelizable across agents
- You need persistent workspace state
- Agents need isolation from each other

**Use Flows when:**
- Work is sequential (output of one → input of next)
- No shared mutable state needed
- Simple pipeline without branching

**Use Delegation when:**
- One agent needs quick help from another
- Subtask is well-defined and synchronous
- No persistent state needed
