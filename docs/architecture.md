# Architecture Overview

Ayo is a CLI framework for creating, managing, and orchestrating AI agents that operate within isolated sandbox environments. This document provides a unified mental model of the core primitives and how they work together.

## Core Primitives

Ayo's architecture consists of five core primitives that compose together:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              AYO SYSTEM                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   ┌─────────┐     ┌─────────┐     ┌─────────┐                           │
│   │ AGENTS  │ ←── │ SKILLS  │     │PLANNERS │                           │
│   │ @name   │     │ +domain │     │ todos   │                           │
│   └────┬────┘     └─────────┘     │ tickets │                           │
│        │                          └────┬────┘                           │
│        │   ┌──────────────────────────────────────────────────────┐     │
│        └──▶│                     SQUADS                           │     │
│            │  #team-name                                          │     │
│            │  ┌──────────┐  ┌──────────┐  ┌──────────┐            │     │
│            │  │ SQUAD.md │  │ .tickets │  │workspace │            │     │
│            │  │ mission  │  │ coord    │  │  files   │            │     │
│            │  └──────────┘  └──────────┘  └──────────┘            │     │
│            └──────────────────────────────────────────────────────┘     │
│                                    ▲                                     │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │                         FLOWS                                    │   │
│   │  step1 → step2 → step3 → ...                                    │   │
│   │  (shell or YAML pipelines)                                      │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Agents (`@name`)

The fundamental execution unit. An agent is an AI identity with:

- **Persona**: System prompt defining personality and capabilities
- **Skills**: Domain-specific instructions that can be attached
- **Planners**: Work tracking tools (todos, tickets)
- **Sandbox**: Isolated container for command execution

Agents are invoked with the `@` prefix:

```bash
ayo @backend "implement the login API"
```

### Skills (`+domain`)

Portable knowledge modules that extend agents with domain expertise:

- **Instructions**: Detailed guidance for specific tasks
- **Examples**: Concrete patterns and templates
- **Constraints**: Rules and best practices

Skills attach to agents with the `+` prefix:

```bash
ayo @ayo +typescript +react "build a todo app"
```

### Planners

Work tracking plugins that give agents the ability to manage tasks:

| Type | Scope | Persistence | Use Case |
|------|-------|-------------|----------|
| **Near-term** (`ayo-todos`) | Session | Memory only | Current work items |
| **Long-term** (`ayo-tickets`) | Cross-session | Disk files | Project coordination |

Planners provide tools (create, list, close) and instructions injected into the agent's prompt.

### Squads (`#name`)

Isolated team environments where multiple agents collaborate:

- **SQUAD.md**: Team constitution (mission, roles, coordination rules)
- **Workspace**: Shared filesystem for code and artifacts
- **Tickets**: File-based coordination between agents
- **I/O Schemas**: Typed interfaces for integration

Squads are targeted with the `#` prefix:

```bash
ayo #dev-team "implement user authentication"
ayo @backend #dev-team "focus on the API layer"
```

### Flows

Composable pipelines that orchestrate multi-step workflows:

- **Shell Flows**: Bash scripts with JSON I/O
- **YAML Flows**: Declarative with dependencies, parallelism, templates

Flows can invoke agents, squads, or shell commands:

```yaml
steps:
  - id: implement
    type: squad
    squad: "#dev-team"
    input: "${{ params.requirements }}"
```

## How Primitives Compose

The primitives layer on top of each other:

```
┌─────────────────────────────────────────┐
│              FLOWS                       │  Orchestration
│  (pipelines, dependencies, triggers)     │
├─────────────────────────────────────────┤
│              SQUADS                      │  Team isolation
│  (sandbox, workspace, SQUAD.md)          │
├─────────────────────────────────────────┤
│     AGENTS + SKILLS + PLANNERS          │  Intelligence
│  (persona, knowledge, work tracking)     │
├─────────────────────────────────────────┤
│              SANDBOX                     │  Execution
│  (containers, filesystem, tools)         │
└─────────────────────────────────────────┘
```

### Composition Examples

**Agent with Skills:**
```bash
# @ayo gains TypeScript and testing expertise
ayo @ayo +typescript +jest "write unit tests for user.ts"
```

**Agent in Squad:**
```bash
# @backend operates within #auth-team context
ayo @backend #auth-team "implement password reset"
```

**Flow with Squad Steps:**
```yaml
steps:
  - id: plan
    type: agent
    agent: "@planner"
  - id: implement
    type: squad
    squad: "#dev-team"
    input: "${{ steps.plan.output }}"
```

## Decision Tree: Choosing Primitives

Use this flowchart to decide which primitives to use:

```
                    START
                      │
                      ▼
         ┌────────────────────────┐
         │ Do I know the exact    │
         │ steps in advance?      │
         └───────────┬────────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
         ▼                       ▼
       YES                      NO
         │                       │
         ▼                       ▼
   ┌──────────┐        ┌────────────────────┐
   │ Use FLOW │        │ Is work parallel/   │
   │          │        │ collaborative?      │
   └──────────┘        └─────────┬──────────┘
                                 │
                   ┌─────────────┴─────────────┐
                   │                           │
                   ▼                           ▼
                 YES                          NO
                   │                           │
                   ▼                           ▼
           ┌───────────┐            ┌──────────────────┐
           │ Use SQUAD │            │ Single agent?    │
           │           │            └────────┬─────────┘
           └───────────┘                     │
                                ┌────────────┴────────────┐
                                │                         │
                                ▼                         ▼
                              YES                        NO
                                │                         │
                                ▼                         ▼
                        ┌───────────┐           ┌──────────────┐
                        │ Use AGENT │           │Use DELEGATION│
                        │ directly  │           │ @agent calls │
                        └───────────┘           │ another      │
                                                └──────────────┘
```

### Quick Decision Rules

| Question | Answer | Primitive |
|----------|--------|-----------|
| Do I know the exact steps? | Yes | **Flow** |
| Do I know the goal but not the steps? | Yes | **Squad dispatch** |
| Need specialized knowledge? | Yes | **Skill** |
| Need to track work across sessions? | Yes | **Planner** (tickets) |
| Need to track work within session? | Yes | **Planner** (todos) |
| Single, well-defined task? | Yes | **Agent** |
| Need quick help from another agent? | Yes | **Delegation** |

## Primitive Details

### When to Use Flows

**Use flows when:**
- The workflow has known, repeatable steps
- Output from one step feeds into the next
- You need triggers (cron, file watch)
- You want to version control the pipeline

**Flow characteristics:**
- Sequential by default, parallel when explicit
- Steps can be shell, agent, or squad
- Template variables connect steps
- Can validate I/O with JSON schemas

### When to Use Squads

**Use squads when:**
- Multiple agents need shared state
- Work is parallelizable across roles
- Agents need different perspectives (frontend, backend, QA)
- You need persistent workspace

**Squad characteristics:**
- Full sandbox isolation
- SQUAD.md defines team context
- Ticket-based coordination
- I/O schemas for typed interfaces

### When to Use Agents Directly

**Use direct agent invocation when:**
- Task is self-contained
- Single agent has necessary skills
- No shared state needed
- Quick, synchronous response needed

**Agent characteristics:**
- Fastest path to execution
- Skills add domain expertise
- Planners track work
- Can delegate to other agents

### When to Use Skills

**Attach skills when:**
- Agent needs domain expertise
- You want consistent patterns across agents
- Task requires specialized knowledge
- You want to share knowledge between agents

**Skill characteristics:**
- Portable across agents
- Stack additively
- Instructions inject into system prompt
- Can include examples and constraints

### When to Use Planners

**Configure planners when:**
- Work needs tracking
- Tasks span multiple sessions (tickets)
- Agent should manage its own work queue (todos)
- Team needs coordination (squad tickets)

**Planner characteristics:**
- Near-term: session-scoped, memory-only
- Long-term: persistent, file-based
- Inject tools and instructions
- Configurable per agent or squad

## Integration Points

### Flows → Squads

Flows can invoke squads as steps:

```yaml
steps:
  - id: implement
    type: squad
    squad: "#dev-team"
    input: "${{ params.feature }}"
    # Squad's input.jsonschema validates input
    # Squad's output.jsonschema defines output
```

Flows can:
- Start squads on-demand
- Wait for squad completion
- Validate I/O against schemas
- Pass data between squads

### Squads → Agents

Squads contain multiple agents working together:

- Each agent receives SQUAD.md context
- Agents communicate via tickets
- Squad lead (`@ayo-in-squad`) coordinates
- Agents share workspace filesystem

### Agents → Skills

Agents extend themselves with skills:

- Skills inject knowledge into system prompt
- Multiple skills can stack
- Order doesn't matter (all additive)
- Skills have no runtime overhead

### Agents → Planners

Planners give agents work tracking abilities:

- Near-term planner: todo list
- Long-term planner: ticket system
- Tools and instructions auto-inject
- Configurable per agent/squad

## See Also

| Topic | Documentation |
|-------|---------------|
| Creating agents | [agents.md](agents.md) |
| Skills system | [skills.md](skills.md) |
| Squad coordination | [squads.md](squads.md) |
| Ticket system | [tickets.md](tickets.md) |
| Flow pipelines | [flows.md](flows.md) |
| Planner plugins | [planners.md](planners.md) |
| I/O schemas | [io-schemas.md](io-schemas.md) |
