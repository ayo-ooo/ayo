# ayo – Agents You Orchestrate

**ayo** is a CLI framework for creating, managing, and orchestrating AI agents that work autonomously in isolated sandbox environments. Design specialized agents, teach them skills, chain them into workflows, and let them coordinate through a ticket-based task system.

Whether you're triaging customer emails, coordinating team communications, processing documents, or automating business workflows—ayo provides the infrastructure to deploy AI agents that actually get work done.

---

## Table of Contents

- [Why ayo?](#why-ayo)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Real-World Examples](#real-world-examples)
- [Documentation](#documentation)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [CLI Reference](#cli-reference)
- [License](#license)

---

## Why ayo?

Most AI tools stop at conversation. ayo goes further:

| Challenge | ayo Solution |
|-----------|--------------|
| Agents that just chat | Agents that execute commands in isolated sandboxes |
| Single-purpose assistants | Specialized agents that delegate to each other |
| Lost context between sessions | Persistent memory and session resumption |
| Manual task management | Ticket-based coordination with dependencies |
| Ad-hoc automation | Declarative flows with scheduling and triggers |
| Security concerns | Sandboxed execution with explicit trust levels |

### Design Philosophy

ayo extends the Unix philosophy to agent-based computing:

- **Do one thing well**: Each agent has a focused purpose
- **Text streams as interface**: JSON flows between agents via pipes
- **Small tools, composed**: Simple agents combine into complex workflows
- **Files as universal abstraction**: Agents are directories, tickets are markdown
- **Isolation by default**: Agents run in containers, not on your host
- **Trust is explicit**: Permissions are granted, not assumed

---

## Quick Start

```bash
# Install
go install github.com/alexcabrera/ayo/cmd/ayo@latest

# Configure API key
export ANTHROPIC_API_KEY="sk-ant-..."

# Run setup
ayo setup

# Start chatting
ayo
```

That's it. You now have a working AI assistant with shell access in a sandboxed environment.

### Your First Agent

```bash
# Create a specialized agent
ayo agents create @triager

# Edit its system prompt
$EDITOR ~/.config/ayo/agents/@triager/system.md
```

```markdown
# Email Triage Agent

You categorize and prioritize incoming emails.

## Categories
- urgent: Requires immediate response
- important: Respond within 24 hours
- routine: Respond when convenient
- spam: Ignore or delete

## Output Format
For each email, output JSON:
{"category": "...", "priority": 1-5, "summary": "...", "suggested_action": "..."}
```

```bash
# Use your agent
cat emails.json | ayo @triager "Categorize these emails"
```

---

## Core Concepts

### Agents

Agents are AI assistants defined as directories containing configuration and system prompts:

```
@support-rep/
├── config.json     # Model, tools, settings
└── system.md       # Behavior instructions
```

Use any agent by prefixing with `@`:

```bash
ayo @support-rep "Draft a response to this complaint"
```

📖 **Deep dive**: [Agents Guide](docs/agents.md)

### Skills

Skills are reusable instruction modules that extend agent capabilities. Add domain knowledge without duplicating prompts:

```bash
ayo skill install customer-service
```

```json
// In agent config.json
{
  "skills": ["customer-service", "email-etiquette"]
}
```

📖 **Deep dive**: [Skills Guide](docs/skills.md)

### Tickets

Tickets coordinate work between agents and across sessions. They're markdown files with dependencies, assignments, and status tracking:

```bash
# Create work items
ayo ticket create "Process Q1 expense reports" -a @finance
ayo ticket create "Review processed expenses" -a @auditor --deps <first-id>

# Check what's ready
ayo ticket ready
```

When an agent closes a ticket, blocked dependents automatically become available.

📖 **Deep dive**: [Tickets Guide](docs/tickets.md)

### Squads

Squads are isolated team sandboxes where multiple agents collaborate under a shared constitution:

```bash
ayo squad create support-team -a @triager,@responder,@escalator
```

The `SQUAD.md` file defines roles, workflows, and coordination rules—automatically injected into every agent's context.

📖 **Deep dive**: [Squads Guide](docs/squads.md)

### Flows

Flows orchestrate multi-step workflows with parallel execution, conditional branching, and error handling:

```yaml
# customer-onboarding.yaml
name: customer-onboarding
steps:
  - name: verify
    agent: "@verifier"
    prompt: "Verify customer identity from {{.input.documents}}"
    
  - name: setup
    agent: "@provisioner"
    depends_on: [verify]
    prompt: "Create accounts for verified customer"
    
  - name: welcome
    agent: "@communicator"
    depends_on: [setup]
    prompt: "Send welcome package to {{.input.email}}"
```

```bash
ayo flow run customer-onboarding '{"email": "new@customer.com", "documents": [...]}'
```

📖 **Deep dive**: [Flows Guide](docs/flows.md)

### Memory

Semantic memory persists facts and preferences across sessions:

```bash
# Store organizational knowledge
ayo memory store "Our SLA requires response within 4 hours for urgent tickets"
ayo memory store "Use formal tone for enterprise customers"

# Agents automatically retrieve relevant memories
ayo @support-rep "Draft response to Acme Corp complaint"
# → Agent retrieves SLA and tone preferences from memory
```

📖 **Deep dive**: [Memory Guide](docs/memory.md)

### Triggers

Automate agent execution with schedules, file watchers, or webhooks:

```bash
# Daily report at 9am
ayo trigger create daily-digest \
  --cron "0 9 * * *" \
  --agent @reporter \
  --prompt "Generate daily team status report"

# Process new files automatically
ayo trigger create invoice-processor \
  --watch ~/Invoices/*.pdf \
  --agent @accounting \
  --prompt "Process invoice: {{.path}}"
```

---

## Real-World Examples

### Example 1: Customer Support Triage

Build an automated support triage system that categorizes, prioritizes, and routes customer inquiries.

**Agents:**
- `@triager`: Categorizes incoming messages
- `@responder`: Drafts initial responses
- `@escalator`: Handles complex issues

**Setup:**

```bash
# Create the squad
ayo squad create support-team -a @triager,@responder,@escalator

# Edit the team constitution
$EDITOR ~/.local/share/ayo/sandboxes/squads/support-team/SQUAD.md
```

```markdown
# Customer Support Squad

## Mission
Provide fast, accurate, empathetic customer support.

## Roles

### @triager
- Categorize incoming tickets: billing, technical, general, spam
- Assign priority 1-5 (1=urgent)
- Route to @responder or @escalator based on complexity

### @responder
- Handle routine inquiries (priority 3-5)
- Use templates from /workspace/templates/
- Escalate if issue requires account access

### @escalator
- Handle complex issues (priority 1-2)
- Coordinate with internal teams via notes
- Has elevated permissions for account lookup

## Workflow
1. @triager processes incoming messages
2. Creates tickets with category, priority, and routing
3. @responder or @escalator picks up assigned tickets
4. Response drafted → ticket closed
```

**Process emails:**

```bash
# Pipe in emails
cat new-emails.json | ayo @triager "Process these support requests"

# Check ticket queue
ayo ticket list

# View what needs attention
ayo ticket ready -a @responder
```

---

### Example 2: Document Processing Pipeline

Automate processing of incoming documents with extraction, validation, and filing.

**Flow definition:**

```yaml
# document-pipeline.yaml
name: document-pipeline
description: Process incoming documents

steps:
  - name: classify
    agent: "@classifier"
    prompt: "Classify document type: {{.input.file}}"
    
  - name: extract
    agent: "@extractor"
    depends_on: [classify]
    prompt: |
      Document type: {{.steps.classify.output.type}}
      Extract key fields from: {{.input.file}}
    
  - name: validate
    agent: "@validator"
    depends_on: [extract]
    prompt: "Validate extracted data: {{.steps.extract.output}}"
    
  - name: file
    agent: "@filer"
    depends_on: [validate]
    prompt: |
      File document based on:
      Type: {{.steps.classify.output.type}}
      Data: {{.steps.extract.output}}
```

**Automate with file watcher:**

```bash
ayo trigger create incoming-docs \
  --watch ~/Incoming/*.pdf \
  --flow document-pipeline \
  --input '{"file": "{{.path}}"}'
```

Now any PDF dropped in `~/Incoming/` gets automatically processed.

---

### Example 3: Team Communication Logger

Create an agent that monitors Slack channels and maintains a structured log of decisions and action items.

**Agent setup:**

```bash
ayo agents create @meeting-notes
```

```markdown
# Meeting Notes Agent

You extract and structure key information from meeting transcripts and chat logs.

## Extract
- Decisions made (with context)
- Action items (with assignee and deadline)
- Open questions
- Key discussion points

## Output Format
```json
{
  "date": "YYYY-MM-DD",
  "participants": ["..."],
  "decisions": [{"decision": "...", "context": "...", "made_by": "..."}],
  "action_items": [{"task": "...", "assignee": "...", "deadline": "..."}],
  "open_questions": ["..."],
  "summary": "..."
}
```
```

**Daily digest flow:**

```yaml
# team-digest.yaml
name: team-digest
steps:
  - name: collect
    agent: "@collector"
    prompt: "Gather today's Slack messages from #team channel"
    
  - name: extract
    agent: "@meeting-notes"
    depends_on: [collect]
    prompt: "Extract decisions and action items: {{.steps.collect.output}}"
    
  - name: distribute
    agent: "@communicator"
    depends_on: [extract]
    prompt: |
      Send daily digest to team:
      {{.steps.extract.output}}
```

```bash
ayo trigger create daily-digest --cron "0 18 * * 1-5" --flow team-digest
```

---

### Example 4: Expense Report Processing

Automate expense report review with multi-stage validation.

**Squad setup:**

```bash
ayo squad create expense-team -a @submitter,@reviewer,@approver
```

**SQUAD.md:**

```markdown
# Expense Processing Squad

## Workflow
1. @submitter receives expense report, extracts line items
2. @reviewer validates against policy, flags issues
3. @approver makes final decision

## Policy
- Meals: max $75/person
- Travel: requires pre-approval over $500
- Supplies: itemized receipts required
- Entertainment: manager approval required

## Output
Each expense gets a ticket with:
- Line item breakdown
- Policy compliance status
- Approval recommendation
```

**Process reports:**

```bash
# Submit expense report
ayo ticket create "Process Alex's Q1 expenses" -a @submitter

# In the ticket description, include the expense data
ayo ticket note <id> "Expense report attached: /shared/expenses/alex-q1.pdf"
```

---

### Example 5: Content Review Pipeline

Multi-agent content review before publication.

```yaml
# content-review.yaml
name: content-review
steps:
  - name: grammar
    agent: "@editor"
    prompt: "Review for grammar and style: {{.input.content}}"
    
  - name: factcheck
    agent: "@researcher"
    prompt: "Verify claims and facts: {{.input.content}}"
    parallel: true  # Runs alongside grammar
    
  - name: legal
    agent: "@compliance"
    prompt: "Check for legal/compliance issues: {{.input.content}}"
    parallel: true  # Runs alongside grammar
    
  - name: synthesize
    agent: "@reviewer"
    depends_on: [grammar, factcheck, legal]
    prompt: |
      Compile review results:
      Grammar: {{.steps.grammar.output}}
      Facts: {{.steps.factcheck.output}}
      Legal: {{.steps.legal.output}}
      
      Provide final recommendation.
```

---

## Documentation

### Getting Started

| Guide | Description |
|-------|-------------|
| [Tutorial](docs/TUTORIAL.md) | Comprehensive walkthrough of ayo concepts |
| [Getting Started](docs/getting-started.md) | Installation and first steps |

### Core Guides

| Guide | Description |
|-------|-------------|
| [Agents](docs/agents.md) | Creating and managing AI agents |
| [Skills](docs/skills.md) | Extending agents with reusable instructions |
| [Tools](docs/tools.md) | Tool system (bash, memory, delegation) |
| [Memory](docs/memory.md) | Persistent facts and preferences |
| [Sessions](docs/sessions.md) | Conversation persistence and resumption |

### Multi-Agent Systems

| Guide | Description |
|-------|-------------|
| [Squads](docs/squads.md) | Team sandboxes with SQUAD.md constitutions |
| [Tickets](docs/tickets.md) | File-based task coordination |
| [Flows](docs/flows.md) | Composable agent pipelines |
| [Planners](docs/planners.md) | Work coordination plugins (todos, tickets) |
| [Delegation](docs/delegation.md) | Task routing to specialists |

### Reference

| Guide | Description |
|-------|-------------|
| [CLI Reference](docs/cli-reference.md) | Complete command documentation |
| [Configuration](docs/configuration.md) | Config files and directories |
| [Plugins](docs/plugins.md) | Extending ayo |
| [Flows Specification](docs/flows-spec.md) | YAML flow schema |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                            HOST                                  │
├─────────────────────────────────────────────────────────────────┤
│  CLI (ayo)                                                       │
│  • LLM API calls via Fantasy abstraction                        │
│  • Memory management (SQLite + embeddings)                       │
│  • Session persistence                                           │
│  • Agent orchestration                                           │
├─────────────────────────────────────────────────────────────────┤
│  Daemon (background service)                                     │
│  • Sandbox pool management                                       │
│  • Trigger engine (cron, watch, webhook)                        │
│  • Ticket watcher (spawns agents on assignment)                 │
├─────────────────────────────────────────────────────────────────┤
│  Sandbox (isolated container)                                    │
│  • Command execution                                             │
│  • File operations                                               │
│  • Per-agent home directories                                    │
│  • Shared workspace via bind mounts                              │
└─────────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
~/.config/ayo/                    # Configuration
├── ayo.json                      # Main config
├── agents/                       # User-defined agents
│   └── @myagent/
│       ├── config.json
│       └── system.md
├── skills/                       # User-defined skills
├── flows/                        # User-defined flows
└── prompts/                      # Custom prefix/suffix

~/.local/share/ayo/               # Data
├── ayo.db                        # SQLite database
├── agents/                       # Built-in agents
├── skills/                       # Built-in skills
├── plugins/                      # Installed plugins
├── sessions/                     # Session data
│   └── {session-id}/.tickets/    # Per-session tickets
└── sandboxes/
    ├── squads/                   # Squad sandboxes
    └── pool/                     # Pre-warmed containers
```

---

## Installation

### Requirements

- Go 1.21+
- macOS 26+ (Apple Silicon) or Linux with systemd
- API key for Anthropic, OpenAI, or OpenRouter

### Install

```bash
go install github.com/alexcabrera/ayo/cmd/ayo@latest
```

### Configure

```bash
# Set API key
export ANTHROPIC_API_KEY="sk-ant-..."

# Run setup
ayo setup

# Verify installation
ayo doctor
```

---

## Configuration

**Main config**: `~/.config/ayo/ayo.json`

```json
{
  "default_model": "claude-sonnet-4-20250514",
  "provider": {
    "name": "anthropic"
  },
  "delegates": {
    "research": "@researcher",
    "writing": "@writer"
  }
}
```

**Environment variables**:

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `OPENAI_API_KEY` | OpenAI API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |
| `AYO_CONFIG` | Custom config path |
| `AYO_HOME` | Custom data directory |

---

## CLI Reference

```bash
# Chat
ayo                              # Interactive chat
ayo "prompt"                     # Single prompt
ayo @agent "prompt"              # Specific agent
ayo -a file.txt "analyze"        # With attachment
ayo -c "follow up"               # Continue session

# Agents
ayo agents list                  # List all agents
ayo agents show @name            # Show details
ayo agents create @name          # Create new agent

# Tickets
ayo ticket list                  # List tickets
ayo ticket create "title"        # Create ticket
ayo ticket start <id>            # Start working
ayo ticket close <id>            # Complete ticket
ayo ticket ready                 # Show available work

# Squads
ayo squad create name            # Create team sandbox
ayo squad list                   # List squads
ayo squad start name             # Start sandbox

# Flows
ayo flow list                    # List flows
ayo flow run name [input]        # Execute flow
ayo flow new name                # Create flow

# System
ayo setup                        # Initial setup
ayo doctor                       # Health check
ayo sandbox service start        # Start daemon
```

📖 **Complete reference**: [CLI Reference](docs/reference/README.md)

---

## License

MIT
