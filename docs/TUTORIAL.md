# The Ayo System: A Comprehensive Guide to Agent-Based Computing

*A manual for understanding, constructing, and orchestrating intelligent software agents*

---

## Table of Contents

1. [Preface: What This Manual Is](#preface-what-this-manual-is)
2. [Part I: Foundations](#part-i-foundations)
   - [Chapter 1: The Philosophy of Agent-Based Systems](#chapter-1-the-philosophy-of-agent-based-systems)
   - [Chapter 2: First Principles of Ayo](#chapter-2-first-principles-of-ayo)
   - [Chapter 3: The Anatomy of an Agent](#chapter-3-the-anatomy-of-an-agent)
3. [Part II: The Architecture](#part-ii-the-architecture)
   - [Chapter 4: System Organization](#chapter-4-system-organization)
   - [Chapter 5: The Sandbox: Isolation and Safety](#chapter-5-the-sandbox-isolation-and-safety)
   - [Chapter 6: The Daemon: Background Intelligence](#chapter-6-the-daemon-background-intelligence)
4. [Part III: Working with Agents](#part-iii-working-with-agents)
   - [Chapter 7: Your First Agent](#chapter-7-your-first-agent)
   - [Chapter 8: Tools and Capabilities](#chapter-8-tools-and-capabilities)
   - [Chapter 9: Skills: Teaching Agents](#chapter-9-skills-teaching-agents)
5. [Part IV: State and Memory](#part-iv-state-and-memory)
   - [Chapter 10: Sessions: Conversational Continuity](#chapter-10-sessions-conversational-continuity)
   - [Chapter 11: Memory: Persistent Knowledge](#chapter-11-memory-persistent-knowledge)
6. [Part V: Multi-Agent Systems](#part-v-multi-agent-systems)
   - [Chapter 12: Delegation: Task Routing](#chapter-12-delegation-task-routing)
   - [Chapter 13: Chaining: Unix Pipes for Agents](#chapter-13-chaining-unix-pipes-for-agents)
   - [Chapter 14: Collaboration: Shared Sandboxes](#chapter-14-collaboration-shared-sandboxes)
7. [Part VI: Automation and Workflows](#part-vi-automation-and-workflows)
   - [Chapter 15: Flows: Composable Pipelines](#chapter-15-flows-composable-pipelines)
   - [Chapter 16: Triggers: Automated Execution](#chapter-16-triggers-automated-execution)
8. [Part VII: Advanced Topics](#part-vii-advanced-topics)
   - [Chapter 17: Plugins: Extending the System](#chapter-17-plugins-extending-the-system)
   - [Chapter 18: Trust Levels and Guardrails](#chapter-18-trust-levels-and-guardrails)
   - [Chapter 19: Inter-Agent Communication](#chapter-19-inter-agent-communication)
9. [Appendices](#appendices)
   - [Appendix A: Configuration Reference](#appendix-a-configuration-reference)
   - [Appendix B: CLI Command Reference](#appendix-b-cli-command-reference)
   - [Appendix C: Troubleshooting](#appendix-c-troubleshooting)

---

## Preface: What This Manual Is

This document serves as both introduction and reference for the Ayo system—a framework for creating, managing, and orchestrating AI agents that operate within your computing environment.

We approach this subject from first principles. Rather than simply listing commands and options, we aim to explain *why* the system is designed as it is, *what problems* it solves, and *how* its components work together. Our goal is that after reading this manual, you will understand not just how to use Ayo, but how to think about agent-based computing more generally.

The reader is assumed to have:
- Familiarity with the Unix command line
- Basic understanding of what large language models (LLMs) are
- Curiosity about how to harness AI for practical tasks

No prior experience with AI agents or prompt engineering is required.

---

# Part I: Foundations

## Chapter 1: The Philosophy of Agent-Based Systems

### 1.1 What Is an Agent?

In the context of computing, an **agent** is a software entity that can:

1. **Perceive** its environment (receive input, read files, observe state)
2. **Reason** about what to do (make decisions, form plans)
3. **Act** upon its environment (execute commands, modify files, communicate)
4. **Learn** from experience (remember outcomes, adapt behavior)

This definition distinguishes agents from simple programs. A calculator transforms input to output according to fixed rules. An agent, by contrast, exhibits goal-directed behavior that adapts to circumstances.

### 1.2 The Promise and the Peril

Large language models have made a new kind of agent possible—one that can understand natural language instructions, reason about complex problems, and generate sophisticated responses. This creates tremendous opportunity:

- **Automation of knowledge work**: Tasks that previously required human judgment can be delegated
- **Natural interfaces**: Users can express intent in plain language rather than learning specialized syntax
- **Compositional intelligence**: Simple agents can be combined to handle complex workflows

But this power brings risks:

- **Unpredictability**: LLM-based agents may behave unexpectedly
- **Security**: An agent with shell access can cause real damage
- **Trust**: How do we know an agent will do what we intend?

Ayo is designed with these tensions in mind. It provides the power of agent-based automation while maintaining human oversight and system safety.

### 1.3 The Unix Philosophy, Extended

Ayo inherits the Unix philosophy:

| Unix Principle | Ayo Application |
|----------------|-----------------|
| Do one thing well | Each agent has a focused purpose |
| Text streams as interface | JSON flows between agents via pipes |
| Small tools, composed | Simple agents combine into complex workflows |
| Files as universal abstraction | Agents are directories with configuration files |

But we extend this philosophy for the age of AI:

| Extension | Meaning |
|-----------|---------|
| Isolation by default | Agents run in sandboxes, not on the host |
| Trust is explicit | Permissions are granted, not assumed |
| Memory is structured | Agents remember, but within boundaries |
| Composition is typed | Schema validation ensures compatible data flow |

### 1.4 The Separation of Concerns

Ayo enforces a fundamental separation:

```
┌─────────────────────────────────────────────────────────────────┐
│                          HOST SYSTEM                             │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  Ayo CLI & Daemon                                           ││
│  │  • LLM API calls                                            ││
│  │  • Memory management                                         ││
│  │  • Session persistence                                       ││
│  │  • Orchestration logic                                       ││
│  └─────────────────────────────────────────────────────────────┘│
│                              │                                   │
│                              ▼                                   │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  SANDBOX CONTAINER                                          ││
│  │  • Command execution                                         ││
│  │  • File operations                                           ││
│  │  • Network access (if permitted)                             ││
│  │  • Isolated from host filesystem                             ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

The host process handles "thinking" (LLM calls, memory, orchestration). The sandbox handles "doing" (command execution, file manipulation). This separation ensures that even if an agent misbehaves, the damage is contained.

---

## Chapter 2: First Principles of Ayo

### 2.1 The Agent as a Directory

In Ayo, an agent is not a running process or a database entry. An agent is a **directory** containing files that define its behavior:

```
@code-reviewer/
├── config.json           # Configuration: model, tools, settings
├── system.md             # System prompt: personality and instructions
├── skills/               # Agent-specific skills (optional)
│   └── my-skill/
│       └── SKILL.md
├── input.jsonschema      # Input validation schema (optional)
└── output.jsonschema     # Output format schema (optional)
```

This design has several advantages:

1. **Version control**: Agent definitions live in your repository
2. **Portability**: Copy a directory to share an agent
3. **Transparency**: Open a file to understand an agent's behavior
4. **Composability**: Agents can be modified with standard text tools

The `@` prefix is convention, not syntax. It distinguishes agent handles from other identifiers and makes them easy to spot in commands and documentation.

### 2.2 Configuration Over Code

Agents are defined declaratively, not programmatically. Here is a minimal agent configuration:

```json
{
  "description": "Reviews code for bugs and best practices",
  "model": "gpt-4.1",
  "allowed_tools": ["bash"]
}
```

And its system prompt (`system.md`):

```markdown
# Code Reviewer

You are an expert code reviewer. When given code or file paths:

1. Read the files carefully
2. Identify bugs, security issues, and style problems
3. Suggest specific improvements with code examples
4. Be constructive and educational

Focus on issues that matter. Don't nitpick formatting.
```

No Python. No JavaScript. No compilation. The agent exists as soon as these files exist.

### 2.3 Tools as Capabilities

An agent without tools is a conversationalist—it can discuss but not act. **Tools** give agents the ability to affect the world:

| Tool | Capability |
|------|------------|
| `bash` | Execute shell commands |
| `memory` | Search and store persistent knowledge |
| `agent_call` | Delegate to other agents |
| `todo` | Track multi-step task progress |

Tools are opt-in. Each agent explicitly declares which tools it may use:

```json
{
  "allowed_tools": ["bash", "memory"]
}
```

This principle of **least privilege** means agents only have the capabilities they need.

### 2.4 Skills as Knowledge

While the system prompt defines an agent's core personality, **skills** provide modular knowledge that can be attached, combined, or swapped:

```
debugging/
└── SKILL.md
```

Skills are Markdown files with YAML frontmatter:

```markdown
---
name: debugging
description: Systematic techniques for finding and fixing bugs
---

# Debugging Methodology

When debugging an issue:

1. **Reproduce**: Ensure you can trigger the bug consistently
2. **Isolate**: Narrow down to the smallest failing case
3. **Instrument**: Add logging to understand state
4. **Hypothesize**: Form a theory about the cause
5. **Test**: Verify your hypothesis with targeted changes
6. **Fix**: Make the minimal change that resolves the issue
7. **Verify**: Confirm the fix and check for regressions

## Common Patterns

### NullPointerException / nil dereference
...
```

Skills follow the [agentskills spec](https://agentskills.org), an open standard for portable agent knowledge.

### 2.5 The Trust Hierarchy

Not all agents are created equal. Ayo recognizes three **trust levels**:

| Level | Meaning | Example |
|-------|---------|---------|
| `sandboxed` | Runs in isolated container | Default for new agents |
| `privileged` | Limited host access | Trusted user agents |
| `unrestricted` | Full host access | System agents only |

Trust is configured per-agent and enforced by the runtime:

```json
{
  "trust_level": "sandboxed"
}
```

**Guardrails** provide an additional layer of safety. When enabled (the default), agents receive constraints in their system prompt that discourage dangerous actions:

```
SAFETY CONSTRAINTS:
- Do not create malicious code
- Do not expose credentials or secrets
- Confirm before destructive operations
- Stay within the scope of the current project
```

---

## Chapter 3: The Anatomy of an Agent

### 3.1 The Agent Lifecycle

When you invoke an agent, the following occurs:

```
1. LOAD
   ├── Parse config.json
   ├── Read system.md
   ├── Discover and load skills
   ├── Validate schemas (if present)
   └── Resolve model and provider

2. ASSEMBLE
   ├── Build environment context (platform, date, git status)
   ├── Apply guardrails (if enabled)
   ├── Inject user prefix (~/.config/ayo/prompts/prefix.md)
   ├── Add system prompt
   ├── Inject user suffix (~/.config/ayo/prompts/suffix.md)
   ├── Append tool documentation
   └── Append skill content

3. EXECUTE
   ├── Send assembled prompt + user message to LLM
   ├── Receive response (possibly with tool calls)
   ├── Execute tool calls (in sandbox if sandboxed)
   ├── Feed results back to LLM
   └── Repeat until complete

4. PERSIST
   ├── Save session history (if enabled)
   ├── Form memories (if detected)
   └── Update access statistics
```

### 3.2 The System Prompt Assembly

The agent's system prompt is not just `system.md`. It is assembled from multiple sources:

```
┌─────────────────────────────────────────────────────────────────┐
│  FINAL SYSTEM PROMPT                                            │
├─────────────────────────────────────────────────────────────────┤
│  1. Environment Context                                         │
│     Platform: darwin/arm64                                      │
│     Date: 2026-02-09                                            │
│     Working directory: /Users/alex/myproject                    │
│     Git branch: main (clean)                                    │
├─────────────────────────────────────────────────────────────────┤
│  2. Guardrails (if enabled)                                     │
│     Safety constraints and ethical guidelines                   │
├─────────────────────────────────────────────────────────────────┤
│  3. User Prefix (if exists)                                     │
│     ~/.config/ayo/prompts/prefix.md                             │
├─────────────────────────────────────────────────────────────────┤
│  4. Agent System Prompt                                         │
│     Contents of system.md                                       │
├─────────────────────────────────────────────────────────────────┤
│  5. User Suffix (if exists)                                     │
│     ~/.config/ayo/prompts/suffix.md                             │
├─────────────────────────────────────────────────────────────────┤
│  6. Tool Documentation                                          │
│     How to use each allowed tool                                │
├─────────────────────────────────────────────────────────────────┤
│  7. Skills Content                                              │
│     Attached skill instructions                                 │
└─────────────────────────────────────────────────────────────────┘
```

This layered approach allows:
- System-wide customization via prefix/suffix
- Per-agent specialization via system.md
- Modular knowledge via skills
- Runtime context via environment injection

### 3.3 Tool Execution Flow

When an agent decides to use a tool, the following occurs:

```
Agent: "I'll check the test results"
        │
        ▼
┌───────────────────────────────────────────────────────────────┐
│  TOOL CALL: bash                                              │
│  Parameters: {"command": "go test ./...", "description": "..."}│
└───────────────────────────────────────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────────────────────────────┐
│  SANDBOX EXECUTION                                            │
│  • Command sent to container                                  │
│  • Output captured (stdout/stderr)                            │
│  • Exit code recorded                                         │
│  • Duration measured                                          │
└───────────────────────────────────────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────────────────────────────┐
│  RESULT                                                       │
│  {"stdout": "PASS\nok  mypackage 0.003s", "exit_code": 0}     │
└───────────────────────────────────────────────────────────────┘
        │
        ▼
Agent: "All tests pass. The package is working correctly."
```

The user sees both the tool call (with spinner and description) and the result (formatted output). This transparency helps users understand what the agent is doing.

### 3.4 Schema Validation

Agents can define JSON schemas for structured input and output:

**input.jsonschema**:
```json
{
  "type": "object",
  "properties": {
    "files": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Files to analyze"
    }
  },
  "required": ["files"]
}
```

When input validation is enabled:
1. Input must be valid JSON
2. JSON must match the schema
3. Invalid input produces a clear error

**output.jsonschema**:
```json
{
  "type": "object",
  "properties": {
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "file": {"type": "string"},
          "line": {"type": "integer"},
          "message": {"type": "string"}
        }
      }
    }
  }
}
```

When output validation is enabled:
1. Agent is instructed to produce JSON matching the schema
2. LLM structured output features are used when available
3. Output is validated before returning

Schemas enable **type-safe chaining**: the output of one agent can be piped to another, with confidence that the data will be compatible.

---

# Part II: The Architecture

## Chapter 4: System Organization

### 4.1 Directory Hierarchy

Ayo organizes its files across two locations following the XDG Base Directory Specification:

**Configuration** (`~/.config/ayo/`):
```
~/.config/ayo/
├── ayo.json                    # Main configuration
├── agents/                     # User-defined agents
│   └── @myagent/
│       ├── config.json
│       └── system.md
├── skills/                     # User-defined shared skills
│   └── my-skill/
│       └── SKILL.md
├── flows/                      # User-defined flows
│   └── my-flow.sh
└── prompts/                    # User customization
    ├── prefix.md               # Prepended to all agents
    └── suffix.md               # Appended to all agents
```

**Data** (`~/.local/share/ayo/`):
```
~/.local/share/ayo/
├── ayo.db                      # SQLite database (sessions, memory)
├── shares.json                 # Shared directory configuration
├── agents/                     # Built-in agents
│   └── @ayo/
├── skills/                     # Built-in skills
│   ├── ayo/
│   ├── debugging/
│   └── coding/
├── plugins/                    # Installed plugins
│   └── ayo-plugins-research/
└── sandbox/                    # Sandbox data
    ├── homes/                  # Agent home directories
    ├── shared/                 # Shared workspace
    └── workspace/              # Shared host directories
```

**Runtime** (`/tmp/ayo/` or platform equivalent):
```
/tmp/ayo/
├── daemon.sock                 # Daemon Unix socket
└── daemon.pid                  # Daemon process ID
```

### 4.2 Configuration Cascade

Configuration is resolved from multiple sources, with later sources overriding earlier ones:

```
1. Built-in defaults
       ↓
2. Global config (~/.config/ayo/ayo.json)
       ↓
3. Directory config (.ayo.json in project or parent)
       ↓
4. Agent config (config.json in agent directory)
       ↓
5. Command-line flags
```

**Example**: Model resolution

```json
// Global: ~/.config/ayo/ayo.json
{"default_model": "gpt-4.1"}

// Directory: ./project/.ayo.json
{"model": "claude-sonnet-4-20250514"}

// Agent: ~/.config/ayo/agents/@myagent/config.json
{"model": "gpt-5.2"}

// CLI
ayo @myagent -m gemini-2.5-pro "hello"
```

Result: `gemini-2.5-pro` (CLI wins)

### 4.3 The Provider Abstraction

Ayo supports multiple LLM providers through the Fantasy abstraction layer:

| Provider | Models | Auth |
|----------|--------|------|
| OpenAI | gpt-4.1, gpt-5.2, o3, etc. | `OPENAI_API_KEY` |
| Anthropic | claude-sonnet-4-20250514, claude-opus-4-20250514, etc. | `ANTHROPIC_API_KEY` |
| Google | gemini-2.5-pro, gemini-2.5-flash, etc. | `GOOGLE_API_KEY` |
| OpenRouter | Access to many models | `OPENROUTER_API_KEY` |
| Ollama | Local models | No auth (local) |

Provider selection is automatic based on model name prefixes:
- `gpt-*`, `o3-*` → OpenAI
- `claude-*` → Anthropic
- `gemini-*` → Google
- Other → OpenRouter or configured default

### 4.4 Database Schema

Ayo uses SQLite for persistent storage. Key tables:

**sessions**: Conversation history
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    agent_handle TEXT NOT NULL,
    title TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

**messages**: Individual messages within sessions
```sql
CREATE TABLE messages (
    id INTEGER PRIMARY KEY,
    session_id TEXT REFERENCES sessions(id),
    role TEXT NOT NULL,  -- 'user', 'assistant', 'tool'
    content TEXT NOT NULL,
    created_at TIMESTAMP
);
```

**memories**: Persistent knowledge
```sql
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    agent_handle TEXT,   -- NULL for global memories
    path_scope TEXT,     -- Project-scoped memories
    content TEXT NOT NULL,
    category TEXT,       -- 'preference', 'fact', 'correction', 'pattern'
    embedding BLOB,      -- Vector for semantic search
    confidence REAL,
    status TEXT,         -- 'active', 'superseded', 'archived', 'forgotten'
    created_at TIMESTAMP,
    accessed_at TIMESTAMP,
    access_count INTEGER
);
```

---

## Chapter 5: The Sandbox: Isolation and Safety

### 5.1 Why Sandbox?

When an agent executes shell commands, it operates with the full power of the command line. This power is dangerous:

- `rm -rf /` could destroy your system
- Network access could exfiltrate data
- Malformed commands could corrupt files

The **sandbox** provides isolation. Commands execute in a container that:
- Has its own filesystem (cannot access host files by default)
- Has controlled network access
- Has resource limits (CPU, memory)
- Has no access to host credentials

Even if an agent misbehaves, the damage is contained within the sandbox.

### 5.2 Sandbox Providers

Ayo uses native containerization, not Docker:

| Provider | Platform | Technology |
|----------|----------|------------|
| `apple` | macOS 26+ (Apple Silicon) | Apple Container Framework |
| `systemd-nspawn` | Linux with systemd | systemd-nspawn |
| `none` | Fallback | No isolation (host execution) |

The provider is auto-selected based on platform availability. You can verify with:

```bash
ayo doctor
```

### 5.3 The Sandbox Filesystem

Inside the sandbox, agents see a Linux filesystem (Alpine-based):

```
/
├── home/                       # Agent home directories
│   └── <agent-handle>/         # Each agent has its own home
├── shared/                     # Shared between all agents
├── workspace/                  # Host directories (via shares)
│   ├── myproject/              # Symlink to ~/Code/myproject
│   └── data/                   # Symlink to /tmp/data
├── run/ayo/                    # Runtime communication
└── ... (standard Linux filesystem)
```

### 5.4 Sharing Host Directories

The `share` system provides controlled access to host directories:

```bash
# Share a directory
ayo share ~/Code/myproject
# Now accessible at /workspace/myproject inside sandbox

# List shares
ayo share list

# Remove a share
ayo share rm myproject
```

Shares are:
- **Immediate**: No sandbox restart required
- **Read-write**: Agents can modify files
- **Symlink-based**: Minimal overhead

This model keeps host files accessible while maintaining a clear boundary between host and sandbox.

### 5.5 Agent Identity

Each agent runs as a dedicated Linux user inside the sandbox:

| Agent Handle | Linux User | Home Directory |
|--------------|------------|----------------|
| `@ayo` | `agent-ayo` | `/home/ayo` |
| `@reviewer` | `agent-reviewer` | `/home/reviewer` |

This per-agent identity enables:
- Persistent home directories (agents remember their workspace state)
- Permission separation (agents can't access each other's homes)
- Multi-agent collaboration (shared `/shared` directory)

---

## Chapter 6: The Daemon: Background Intelligence

### 6.1 Why a Daemon?

Some operations require persistence beyond a single command:

- **Sandbox pool**: Pre-warmed containers for fast startup
- **Triggers**: Cron schedules and file watchers
- **Sessions**: Managing long-running agent interactions
- **Tickets**: File-based task coordination for multi-agent workflows

The **daemon** is a background process that provides these services.

### 6.2 Starting the Daemon

The daemon starts automatically when needed. You can also manage it explicitly:

```bash
# Start in background
ayo sandbox service start

# Start in foreground (for debugging)
ayo sandbox service start -f

# Check status
ayo sandbox service status

# Stop
ayo sandbox service stop
```

### 6.3 The Sandbox Pool

Creating a container from scratch takes time. The daemon maintains a **warm pool** of pre-created sandboxes:

```
Pool Configuration:
├── MinSize: 1      # Always keep at least 1 warm
├── MaxSize: 4      # Never exceed 4 total
└── Image: alpine   # Base container image
```

When you run an agent:
1. Daemon acquires a sandbox from the pool
2. Agent operates within that sandbox
3. When done, sandbox returns to pool or is destroyed

This provides near-instant agent startup while managing resource usage.

### 6.4 Communication Protocol

The CLI communicates with the daemon via JSON-RPC over a Unix socket:

```
CLI                          Daemon
 │                              │
 │──── acquire_sandbox ────────>│
 │<─── {sandbox_id: "abc123"} ──│
 │                              │
 │──── exec(sandbox_id, cmd) ──>│
 │<─── {stdout: "...", ...} ────│
 │                              │
 │──── release_sandbox ────────>│
 │<─── {ok: true} ──────────────│
```

This architecture allows the daemon to manage sandbox lifecycle across multiple CLI invocations.

---

# Part III: Working with Agents

## Chapter 7: Your First Agent

### 7.1 The Built-in Agent

Ayo ships with one built-in agent: `@ayo`. This agent is designed to be a versatile assistant capable of handling diverse tasks:

```bash
# Start interactive chat
ayo

# Single prompt
ayo "explain how Unix pipes work"

# With file attachment
ayo -a main.go "review this code"
```

`@ayo` has access to:
- `bash` tool for command execution
- `agent_call` for delegating to specialists
- `todo` for tracking multi-step tasks
- Built-in skills for common tasks

### 7.2 Creating a Custom Agent

Let's create a specialized code reviewer agent. First, understand the structure:

```
@reviewer/
├── config.json       # Configuration
└── system.md         # System prompt
```

**Step 1**: Create the directory

```bash
mkdir -p ~/.config/ayo/agents/@reviewer
```

**Step 2**: Write the configuration

```bash
cat > ~/.config/ayo/agents/@reviewer/config.json << 'EOF'
{
  "description": "Expert code reviewer focusing on Go and Python",
  "model": "claude-sonnet-4-20250514",
  "allowed_tools": ["bash"],
  "skills": ["debugging", "coding"],
  "guardrails": true
}
EOF
```

**Step 3**: Write the system prompt

```bash
cat > ~/.config/ayo/agents/@reviewer/system.md << 'EOF'
# Code Reviewer

You are an expert code reviewer specializing in Go and Python.

## Your Process

1. **Read** the code carefully, understanding its purpose
2. **Analyze** for bugs, security issues, and anti-patterns
3. **Prioritize** findings by severity (critical > high > medium > low)
4. **Suggest** specific improvements with code examples

## Review Categories

### Security
- Input validation
- Authentication/authorization
- Data exposure
- Injection vulnerabilities

### Correctness
- Logic errors
- Edge cases
- Error handling
- Race conditions

### Maintainability
- Code clarity
- Documentation
- Test coverage
- Dependency management

## Output Format

For each finding:
1. Location (file:line)
2. Severity (critical/high/medium/low)
3. Issue description
4. Suggested fix (with code)

Be constructive. Explain *why* something is problematic, not just *what* is wrong.
EOF
```

**Step 4**: Test the agent

```bash
# List agents to verify
ayo agents list

# Show details
ayo agents show @reviewer

# Run the agent
ayo @reviewer "Review the main.go file for security issues"
```

### 7.3 Interactive Creation

For a guided experience, ask `@ayo` to help:

```bash
ayo "help me create an agent for writing unit tests"
```

`@ayo` will:
1. Ask clarifying questions about your needs
2. Suggest appropriate tools and skills
3. Draft the configuration and system prompt
4. Create the files for you

### 7.4 Agent Configuration Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | | Brief description (shown in listings) |
| `model` | string | (global) | LLM model to use |
| `allowed_tools` | string[] | `["bash"]` | Tools the agent can use |
| `skills` | string[] | `[]` | Skills to attach |
| `exclude_skills` | string[] | `[]` | Skills to exclude |
| `ignore_builtin_skills` | bool | `false` | Skip built-in skills |
| `ignore_shared_skills` | bool | `false` | Skip user shared skills |
| `guardrails` | bool | `true` | Enable safety constraints |
| `trust_level` | string | `"sandboxed"` | Trust level |
| `delegates` | object | `{}` | Task type → agent mappings |
| `memory.enabled` | bool | `false` | Enable memory for this agent |
| `sandbox.enabled` | bool | `true` | Run in sandbox |
| `sandbox.network` | bool | `true` | Allow network access |
| `sandbox.resources.cpus` | int | 2 | CPU limit |
| `sandbox.resources.memory_mb` | int | 512 | Memory limit (MB) |

---

## Chapter 8: Tools and Capabilities

### 8.1 The Tool Abstraction

A **tool** is a capability that an agent can invoke. From the agent's perspective, a tool is a function with:
- A name
- A description (explaining when to use it)
- Parameters (typed inputs)
- A return value (the result)

From the system's perspective, a tool is:
- An execution context (host, sandbox, or bridge)
- A handler that processes the call
- Security constraints

### 8.2 Built-in Tools

#### `bash`: Command Execution

The most important tool. Enables agents to run shell commands.

**Parameters**:
```json
{
  "command": "go test ./...",
  "description": "Run all tests",
  "timeout_seconds": 60,
  "working_dir": "/workspace/myproject"
}
```

**Execution**: Commands run in the sandbox (or host if unsandboxed).

**Output**:
```json
{
  "stdout": "PASS\nok  mypackage 0.003s",
  "stderr": "",
  "exit_code": 0,
  "duration": "0.4s"
}
```

#### `todo`: Task Tracking

Enables agents to track progress on multi-step tasks.

**Parameters**:
```json
{
  "todos": [
    {
      "content": "Review authentication module",
      "active_form": "Reviewing authentication module",
      "status": "completed"
    },
    {
      "content": "Check error handling",
      "active_form": "Checking error handling",
      "status": "in_progress"
    },
    {
      "content": "Write summary report",
      "active_form": "Writing summary report",
      "status": "pending"
    }
  ]
}
```

The UI displays this todo list, giving users visibility into agent progress.

#### `memory`: Knowledge Management

Enables agents to search, store, and manage persistent knowledge.

**Operations**:
- `search`: Find relevant memories by semantic similarity
- `store`: Save new information
- `list`: Show all memories
- `forget`: Remove a memory

**Example**:
```json
{
  "operation": "search",
  "query": "user's preferred programming language"
}
```

#### `agent_call`: Delegation

Enables agents to call other agents.

**Parameters**:
```json
{
  "agent": "@researcher",
  "prompt": "Find best practices for error handling in Go"
}
```

The called agent runs as a sub-agent, with its output returned to the caller.

### 8.3 Tool Categories

Some tool names are **categories** that resolve to specific implementations:

| Category | Default | Description |
|----------|---------|-------------|
| `planning` | `todo` | Task tracking |
| `shell` | `bash` | Command execution |
| `search` | (none) | Web search |

Categories allow users to swap implementations without changing agent configs:

```json
// ~/.config/ayo/ayo.json
{
  "default_tools": {
    "search": "searxng"
  }
}
```

Now any agent with `allowed_tools: ["search"]` uses `searxng`.

### 8.4 Tool Execution Contexts

Tools execute in different contexts based on security requirements:

| Context | Where | Examples |
|---------|-------|----------|
| `host` | Host process | `memory`, `agent_call`, `todo` |
| `sandbox` | Container | `bash` |
| `bridge` | Crosses boundary | `file_request`, `publish` |

**Bridge tools** enable controlled data transfer between sandbox and host:
- `file_request`: Agent requests files from host → copied to sandbox
- `publish`: Agent publishes files from sandbox → copied to host

---

## Chapter 9: Skills: Teaching Agents

### 9.1 What Skills Provide

Skills are modular knowledge modules that can be attached to agents. Unlike system prompts (which define personality), skills provide:

- **Domain knowledge**: How to debug Python, write tests, use specific frameworks
- **Procedures**: Step-by-step processes for common tasks
- **Best practices**: Guidelines and patterns

Skills are loaded at runtime and injected into the system prompt.

### 9.2 Skill Structure

```
my-skill/
├── SKILL.md              # Required: skill definition
├── scripts/              # Optional: executable code
│   └── helper.sh
├── references/           # Optional: additional docs
│   └── examples.md
└── assets/               # Optional: templates, data
    └── template.json
```

### 9.3 SKILL.md Format

```markdown
---
name: go-debugging
description: |
  Techniques for debugging Go programs.
  Use when encountering Go errors, panics, or test failures.
metadata:
  author: your-name
  version: "1.0"
compatibility: Requires Go 1.21+
---

# Go Debugging

## Quick Diagnosis

When you see a Go error:

1. Read the full stack trace bottom-to-top
2. Identify the error type and message
3. Locate your code in the stack (vs. library code)

## Common Errors

### nil pointer dereference

```go
// Problem
var user *User
fmt.Println(user.Name)  // panic: nil pointer

// Solution: Check before use
if user != nil {
    fmt.Println(user.Name)
}
```

### Test Failures

```bash
# Run specific test with verbose output
go test -v -run TestName ./package/...

# With race detection
go test -race ./...
```

## Debugging Tools

```bash
# Delve debugger
dlv debug ./cmd/main

# pprof profiling
go tool pprof http://localhost:6060/debug/pprof/profile
```
```

### 9.4 Skill Discovery

Skills are discovered from multiple locations (in priority order):

1. **Agent-specific**: `@agent/skills/`
2. **User shared**: `~/.config/ayo/skills/`
3. **Built-in**: `~/.local/share/ayo/skills/`
4. **Plugin-provided**: In installed plugins

First match wins, enabling overrides.

### 9.5 Creating a Custom Skill

```bash
# Create in shared directory
ayo skills create my-skill --shared

# Or create for a specific agent
cd ~/.config/ayo/agents/@myagent
mkdir -p skills/my-skill
```

Edit `SKILL.md` with your knowledge, then verify:

```bash
ayo skills validate my-skill
ayo skills show my-skill
```

Attach to an agent:

```json
{
  "skills": ["my-skill"]
}
```

---

# Part IV: State and Memory

## Chapter 10: Sessions: Conversational Continuity

### 10.1 What Sessions Provide

A **session** is a record of a conversation between you and an agent. Sessions enable:

- **Continuity**: Resume conversations where you left off
- **Context**: Agents remember earlier discussion
- **History**: Review past interactions

### 10.2 Session Lifecycle

```
1. START
   └── New session created with unique ID

2. CONVERSATION
   ├── User messages recorded
   ├── Assistant responses recorded
   └── Tool calls/results recorded

3. END
   └── Session saved to database

4. RESUME (optional)
   └── Load previous messages as context
```

### 10.3 Working with Sessions

```bash
# List sessions
ayo sessions list

# Show session details
ayo sessions show <id>

# Continue most recent session
ayo -c "follow up on that"

# Continue specific session
ayo -s abc123 "what about the edge cases?"

# Interactive picker
ayo sessions continue
```

### 10.4 Session Structure

Each session contains:

| Field | Description |
|-------|-------------|
| ID | Unique identifier (ULID) |
| Agent | Which agent was used |
| Title | Auto-generated or user-set |
| Messages | Conversation history |
| Created | When started |
| Updated | Last activity |

Titles are auto-generated by a small model based on conversation content.

---

## Chapter 11: Memory: Persistent Knowledge

### 11.1 The Nature of Agent Memory

Unlike sessions (which store conversation history), **memory** stores distilled knowledge that persists across sessions:

| Sessions | Memory |
|----------|--------|
| Full conversation | Key facts/preferences |
| Tied to one interaction | Available across all |
| Ephemeral context | Persistent knowledge |
| Linear history | Semantic searchable |

### 11.2 Memory Categories

| Category | Description | Example |
|----------|-------------|---------|
| `preference` | User preferences | "Prefers TypeScript for frontend" |
| `fact` | Factual information | "Project uses PostgreSQL 15" |
| `correction` | User corrections | "Don't suggest SQL in this codebase" |
| `pattern` | Observed patterns | "Usually runs tests before commits" |

### 11.3 How Memory Works

**Storage**:
1. Memorable content detected (explicit or implicit)
2. Small model extracts and categorizes
3. Embedding model creates vector representation
4. Stored in SQLite with vector for search

**Retrieval**:
1. Session starts
2. User's first message + context used as query
3. Semantic search finds relevant memories
4. Top memories injected into system prompt

### 11.4 Memory Commands

```bash
# Store manually
ayo memory store "I prefer tabs over spaces"

# Search memories
ayo memory search "formatting preferences"

# List all memories
ayo memory list

# Show memory details
ayo memory show <id>

# Forget a memory
ayo memory forget <id>

# Statistics
ayo memory stats
```

### 11.5 Automatic Memory Formation

During conversations, agents detect memorable content:

```
You: "Remember that I always want verbose test output"
Agent: "I'll remember that preference."

[Memory automatically stored]
```

Triggers include:
- Explicit "remember" requests
- Stated preferences ("I prefer...", "I like...")
- Corrections ("No, actually...", "That's wrong...")
- Project facts ("This project uses...")

### 11.6 Memory Scopes

| Scope | Applies To |
|-------|------------|
| `global` | All agents, all directories |
| `agent` | Specific agent only |
| `path` | Specific project/directory |
| `hybrid` | Combines all (default) |

Scoping ensures project-specific knowledge doesn't pollute global memory.

---

# Part V: Multi-Agent Systems

## Chapter 12: Delegation: Task Routing

### 12.1 The Delegation Pattern

Delegation allows a generalist agent to route specialized tasks to experts:

```
User: "Refactor the authentication module"
   │
   ▼
@ayo: "This is a coding task, delegating to @crush"
   │
   ▼
@crush: [performs refactoring]
   │
   ▼
User sees result
```

### 12.2 Configuring Delegates

Delegates are configured at three levels:

**Project level** (`.ayo.json`):
```json
{
  "delegates": {
    "coding": "@crush",
    "research": "@researcher"
  }
}
```

**Global level** (`~/.config/ayo/ayo.json`):
```json
{
  "delegates": {
    "coding": "@crush"
  }
}
```

**Agent level** (`config.json`):
```json
{
  "delegates": {
    "debug": "@debugger"
  }
}
```

Project overrides global; agent config applies to that agent only.

### 12.3 Task Types

| Type | Description |
|------|-------------|
| `coding` | Source code creation/modification |
| `research` | Web research and information gathering |
| `debug` | Debugging and troubleshooting |
| `test` | Test creation and execution |
| `docs` | Documentation generation |

You can define custom task types—they're just strings that map to agents.

### 12.4 How Delegation Works

1. Agent receives a task
2. Agent recognizes task type (via LLM reasoning)
3. Agent checks delegate configuration
4. If delegate exists, invokes via `agent_call` tool
5. Delegate handles task
6. Result returns to original agent

The user may or may not see the delegation, depending on the agent's communication style.

---

## Chapter 13: Chaining: Unix Pipes for Agents

### 13.1 The Chaining Philosophy

Unix pipes revolutionized computing by enabling simple programs to compose:

```bash
cat file.txt | grep pattern | wc -l
```

Ayo extends this philosophy to agents:

```bash
ayo @analyzer '{"code":"..."}' | ayo @reporter
```

Each agent:
- Receives JSON input from stdin
- Produces JSON output to stdout
- Logs UI feedback to stderr

### 13.2 Schema-Based Compatibility

For reliable chaining, agents define JSON schemas:

**Producer** (`@analyzer/output.jsonschema`):
```json
{
  "type": "object",
  "properties": {
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "file": {"type": "string"},
          "line": {"type": "integer"},
          "severity": {"type": "string"},
          "message": {"type": "string"}
        }
      }
    }
  }
}
```

**Consumer** (`@reporter/input.jsonschema`):
```json
{
  "type": "object",
  "properties": {
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "severity": {"type": "string"},
          "message": {"type": "string"}
        },
        "required": ["severity", "message"]
      }
    }
  },
  "required": ["findings"]
}
```

The system validates that producer output is compatible with consumer input.

### 13.3 Chain Discovery

```bash
# List chainable agents
ayo chain ls

# What can receive @analyzer's output?
ayo chain from @analyzer

# What can feed into @reporter?
ayo chain to @reporter

# Validate input
ayo chain validate @reporter '{"findings": [...]}'

# Generate example input
ayo chain example @reporter
```

### 13.4 Multi-Stage Pipelines

```bash
# Three-stage pipeline
ayo @code-scanner '{"path":"./src"}' \
  | ayo @issue-prioritizer \
  | ayo @report-generator

# With shell processing
ayo @analyzer '{"code":"..."}' \
  | jq '.findings | length' \
  | xargs -I {} echo "Found {} issues"
```

---

## Chapter 14: Collaboration: Shared Sandboxes

### 14.1 Multi-Agent Sandboxes

By default, each agent gets its own sandbox. But agents can share a sandbox for collaboration:

```bash
# Start a sandbox
ayo sandbox list
# → sandbox-abc123

# Join another agent to it
ayo sandbox join abc123 @researcher

# Both agents now share:
# - /shared directory
# - Network namespace
# - Process visibility
```

### 14.2 The Shared Filesystem

Inside a shared sandbox:

```
/home/ayo/           # @ayo's private home
/home/researcher/    # @researcher's private home
/shared/             # Shared between all agents
/workspace/          # Shared host directories
```

Agents can leave files in `/shared` for each other:

```
@ayo writes /shared/research-request.json
@researcher reads, processes, writes /shared/research-results.json
@ayo reads results
```

---

# Part VI: Automation and Workflows

## Chapter 15: Flows: Composable Pipelines

### 15.1 What Flows Provide

While agent chaining (pipes) handles simple linear composition, **flows** handle complex workflows:

- Multi-step sequences
- Parallel execution
- Conditional branching
- Error handling
- Scheduling

### 15.2 Shell Flows

The simplest flow type: a bash script with JSON I/O:

```bash
#!/usr/bin/env bash
# ayo:flow
# name: daily-summary
# description: Generate daily project summary

set -euo pipefail

INPUT="${1:-$(cat)}"
PROJECT=$(echo "$INPUT" | jq -r '.project // "."')

# Gather data
GIT_LOG=$(cd "$PROJECT" && git log --oneline --since="1 day ago")
TEST_STATUS=$(cd "$PROJECT" && go test ./... 2>&1 | tail -5)

# Generate summary via agent
jq -n \
  --arg git "$GIT_LOG" \
  --arg tests "$TEST_STATUS" \
  '{commits: $git, tests: $tests}' \
  | ayo @ayo "Summarize this daily activity as JSON with: headline, highlights, action_items"
```

Run with:
```bash
ayo flows run daily-summary '{"project": "~/Code/myproject"}'
```

### 15.3 YAML Flows

For complex workflows, YAML provides declarative definition:

```yaml
version: 1
name: code-review-pipeline
description: Full code review with prioritization and reporting

params:
  path:
    type: string
    default: "."
    description: Path to review

steps:
  - id: scan
    type: shell
    run: |
      find {{ params.path }} -name "*.go" -exec wc -l {} \;

  - id: analyze
    type: agent
    agent: "@ayo"
    prompt: |
      Analyze these Go files for issues:
      {{ steps.scan.stdout }}
      
      Return JSON: {findings: [{file, line, severity, message}]}
    depends_on: [scan]

  - id: prioritize
    type: agent
    agent: "@ayo"
    prompt: |
      Prioritize these findings by business impact:
      {{ steps.analyze.stdout }}
    depends_on: [analyze]

  - id: report
    type: shell
    run: |
      echo '{{ steps.prioritize.stdout }}' | jq '.findings | sort_by(.severity)'
    depends_on: [prioritize]
    when: "{{ steps.analyze.exit_code == 0 }}"
```

### 15.4 Flow Features

| Feature | Shell | YAML |
|---------|-------|------|
| JSON I/O | ✓ | ✓ |
| Schema validation | ✓ | ✓ |
| Parallel steps | Manual | ✓ |
| Dependencies | Manual | ✓ |
| Conditions | Manual | ✓ |
| Template variables | Manual | ✓ |
| Built-in triggers | ✗ | ✓ |

### 15.5 Flow Commands

```bash
# Create new flow
ayo flows new my-flow
ayo flows new my-flow --yaml

# List flows
ayo flows list

# Run flow
ayo flows run my-flow '{"input": "data"}'

# Validate YAML flow
ayo flows validate my-flow.yaml

# Show run history
ayo flows history

# Replay a run
ayo flows replay <run-id>
```

---

## Chapter 16: Triggers: Automated Execution

### 16.1 Trigger Types

Triggers automatically execute agents or flows:

| Type | Description |
|------|-------------|
| `cron` | Time-based schedules |
| `watch` | File system changes |
| `webhook` | HTTP requests |

### 16.2 Cron Triggers

```bash
# Add cron trigger
ayo triggers add \
  --cron "0 9 * * *" \
  --agent @ayo \
  --prompt "Generate daily standup summary"
```

Cron syntax: `minute hour day-of-month month day-of-week`

Examples:
- `0 9 * * *` - Daily at 9 AM
- `*/30 * * * *` - Every 30 minutes
- `0 0 * * 0` - Weekly on Sunday

### 16.3 Watch Triggers

```bash
# Watch for file changes
ayo triggers add \
  --watch ~/Code/myproject/src \
  --patterns "*.go" \
  --agent @ayo \
  --prompt "A Go file changed. Run tests and report."
```

Watch options:
- `--patterns`: Glob patterns to match
- `--events`: `create`, `modify`, `delete`
- `--debounce`: Wait time before firing (prevents rapid re-triggers)

### 16.4 Managing Triggers

```bash
# List triggers
ayo triggers list

# Show trigger details
ayo triggers show <id>

# Test trigger (run immediately)
ayo triggers test <id>

# Enable/disable
ayo triggers enable <id>
ayo triggers disable <id>

# Remove trigger
ayo triggers rm <id>
```

### 16.5 YAML Flow Triggers

YAML flows can define their own triggers:

```yaml
triggers:
  - type: cron
    schedule: "0 9 * * 1-5"  # Weekdays at 9 AM
    
  - type: watch
    path: ./src
    patterns: ["*.go"]
    debounce: 5s
```

These are registered with the daemon automatically.

---

# Part VII: Advanced Topics

## Chapter 17: Plugins: Extending the System

### 17.1 What Plugins Provide

Plugins extend Ayo with:
- New agents
- New skills
- New tools
- Custom renderers (for tool output)
- Delegate recommendations

### 17.2 Plugin Structure

```
ayo-plugins-example/
├── manifest.json         # Plugin metadata
├── agents/               # Agent definitions
│   └── @example-agent/
├── skills/               # Skill definitions
│   └── example-skill/
├── tools/                # Tool definitions
│   └── example-tool/
├── renderers/            # Custom TUI renderers
│   └── example-renderer.yaml
└── README.md
```

### 17.3 Installing Plugins

```bash
# Install from Git
ayo plugins install https://github.com/user/ayo-plugins-example

# List installed
ayo plugins list

# Show details
ayo plugins show example

# Update all plugins
ayo plugins update

# Remove
ayo plugins remove example
```

### 17.4 Creating a Plugin

**manifest.json**:
```json
{
  "name": "example",
  "version": "1.0.0",
  "description": "Example plugin demonstrating all features",
  "author": "Your Name",
  "homepage": "https://github.com/user/ayo-plugins-example",
  "agents": ["@example-agent"],
  "skills": ["example-skill"],
  "tools": ["example-tool"],
  "delegates": {
    "example": "@example-agent"
  }
}
```

Plugins are just Git repositories with this structure.

---

## Chapter 18: Trust Levels and Guardrails

### 18.1 Trust Levels

| Level | Execution | Capabilities |
|-------|-----------|--------------|
| `sandboxed` | Container | Limited to sandbox filesystem and allowed tools |
| `privileged` | Container + Host | Can access specific host paths |
| `unrestricted` | Host | Full system access (dangerous) |

Configure per-agent:
```json
{
  "trust_level": "sandboxed"
}
```

### 18.2 Guardrails

Guardrails are safety instructions injected into the system prompt:

```
SAFETY CONSTRAINTS:
- Do not create code intended for malicious purposes
- Do not expose credentials, API keys, or secrets
- Confirm with the user before destructive operations
- Stay within the scope of the current project
```

Guardrails are:
- **Enabled by default** for all agents
- **Always enabled** for `@ayo` namespace
- Can be disabled (not recommended) via `"guardrails": false`

### 18.3 The Reserved Namespace

The `@ayo` namespace is reserved for built-in agents:
- Users cannot create `@ayo` or `@ayo.*` agents
- All `@ayo` agents have guardrails enforced
- This prevents malicious impersonation

### 18.4 Security Best Practices

1. **Principle of least privilege**: Only grant tools agents need
2. **Use sandboxed trust level**: Default for all new agents
3. **Keep guardrails enabled**: Unless you have specific reason
4. **Review generated code**: Before executing in production
5. **Limit network access**: When possible via sandbox config

---

## Chapter 19: Inter-Agent Communication

### 19.1 Communication Mechanisms

| Mechanism | Use Case |
|-----------|----------|
| Agent call | Synchronous delegation |
| Chaining | Data pipeline (stdin/stdout) |
| Shared files | Asynchronous collaboration |
| Tickets | Task coordination with dependencies |

### 19.2 Agent Call (Synchronous)

The `agent_call` tool invokes another agent and waits for response:

```json
{
  "tool": "agent_call",
  "parameters": {
    "agent": "@researcher",
    "prompt": "Find best practices for X"
  }
}
```

Caller blocks until callee completes.

### 19.3 Chaining (Streaming)

Pipe agents for data transformation:

```bash
ayo @producer '{}' | ayo @transformer | ayo @consumer
```

Data flows through, with each agent processing and passing on.

### 19.4 Shared Files (Asynchronous)

Agents in shared sandbox write files for each other:

```
@coordinator: writes /shared/task-001.json
@worker-1: reads task, writes /shared/result-001.json
@worker-2: reads task, writes /shared/result-002.json
@coordinator: reads results, synthesizes
```

### 19.5 Tickets (Task Coordination)

Tickets provide persistent, file-based task coordination. Unlike real-time messaging, tickets create an auditable trail of work and support dependencies between tasks.

**Creating and assigning work:**

```bash
# Coordinator creates tickets
ayo ticket create "Implement auth module" -a @backend -s project-session
ayo ticket create "Write auth tests" -a @tester --deps auth-impl -s project-session
ayo ticket create "Review auth code" -a @reviewer --deps auth-impl -s project-session
```

**Working on tickets:**

```bash
# Agent finds ready work
ayo ticket ready -a @backend -s project-session

# Claims and starts
ayo ticket start auth-impl -s project-session

# Adds progress notes
ayo ticket note auth-impl "Completed JWT implementation" -s project-session

# Marks complete
ayo ticket close auth-impl -s project-session
```

**Handling dependencies:**

When `@backend` closes `auth-impl`, the dependent tickets (`auth-tests`, `auth-review`) become "ready" for their assignees.

```bash
# @tester can now see their ticket is ready
ayo ticket ready -a @tester -s project-session
# Shows: auth-tests (deps resolved)
```

**Benefits over other mechanisms:**

| Feature | Tickets | Agent Call | Shared Files |
|---------|---------|------------|--------------|
| Async | ✓ | ✗ | ✓ |
| Dependencies | ✓ | ✗ | Manual |
| Audit trail | ✓ | Logs | ✗ |
| Status tracking | ✓ | ✗ | Manual |
| Git-friendly | ✓ | ✗ | ✓ |

For comprehensive ticket documentation, see [Tickets](tickets.md).

### 19.6 Two-Tier Task Management

Ayo uses a two-tier approach to task management that separates near-term execution from medium/long-term planning:

| Layer | Tool | Scope | Lifetime | Purpose |
|-------|------|-------|----------|---------|
| **Near-term** | `todo` | Single session | Ephemeral | Steps to complete current work |
| **Medium/Long-term** | `ticket` | Across sessions | Persistent | Project-level work items |

**How they relate:**

```
Ticket: "Implement authentication module" (proj-a1b2)
  │
  └── Agent picks up ticket, starts session
      │
      └── Todo list (internal to this session):
          - [x] Read existing auth code
          - [x] Design JWT structure
          - [ ] Implement login endpoint
          - [ ] Write tests
```

**Todos** are the agent's internal working memory—a scratchpad for tracking immediate steps. When a session ends (agent stops, context limit reached, handoff), todos disappear.

**Tickets** are the persistent interface between sessions and agents. Progress is captured in ticket notes, ensuring continuity when:
- An agent's session ends and restarts
- Work is handed off to a different agent
- Multiple agents collaborate on related tasks

This separation keeps agents focused (todos for "what am I doing now") while maintaining project visibility (tickets for "what needs to be done").

---

# Appendices

## Appendix A: Configuration Reference

### Global Configuration (`~/.config/ayo/ayo.json`)

```json
{
  "default_model": "claude-sonnet-4-20250514",
  "provider": {
    "name": "anthropic"
  },
  "ollama_host": "http://localhost:11434",
  "embedding": {
    "model": "nomic-embed-text"
  },
  "small_model": "ministral-3:3b",
  "default_tools": {
    "search": "searxng"
  },
  "delegates": {
    "coding": "@crush"
  }
}
```

### Directory Configuration (`.ayo.json`)

```json
{
  "agent": "@ayo",
  "model": "gpt-5.2",
  "delegates": {
    "coding": "@crush"
  }
}
```

### Agent Configuration (`config.json`)

```json
{
  "description": "Agent description",
  "model": "claude-sonnet-4-20250514",
  "allowed_tools": ["bash", "memory", "agent_call"],
  "skills": ["debugging", "coding"],
  "exclude_skills": [],
  "ignore_builtin_skills": false,
  "ignore_shared_skills": false,
  "guardrails": true,
  "trust_level": "sandboxed",
  "delegates": {},
  "memory": {
    "enabled": true,
    "scope": "hybrid",
    "formation_triggers": {
      "on_correction": true,
      "on_preference": true,
      "on_project_fact": true,
      "explicit_only": false
    },
    "retrieval": {
      "auto_inject": true,
      "threshold": 0.3,
      "max_memories": 10
    }
  },
  "sandbox": {
    "enabled": true,
    "network": true,
    "resources": {
      "cpus": 2,
      "memory_mb": 512,
      "timeout": 300
    }
  }
}
```

---

## Appendix B: CLI Command Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `ayo` | Interactive chat with default agent |
| `ayo "prompt"` | Single prompt to default agent |
| `ayo @agent` | Interactive chat with specific agent |
| `ayo @agent "prompt"` | Single prompt to specific agent |
| `ayo -a file.txt "..."` | Attach file to prompt |
| `ayo -c "..."` | Continue most recent session |
| `ayo -s ID "..."` | Continue specific session |

### Agent Commands

| Command | Description |
|---------|-------------|
| `ayo agents list` | List all agents |
| `ayo agents show @name` | Show agent details |
| `ayo agents create @name` | Create new agent |
| `ayo agents update` | Update built-in agents |

### Skill Commands

| Command | Description |
|---------|-------------|
| `ayo skills list` | List all skills |
| `ayo skills show name` | Show skill details |
| `ayo skills create name` | Create new skill |
| `ayo skills validate path` | Validate skill |
| `ayo skills update` | Update built-in skills |

### Session Commands

| Command | Description |
|---------|-------------|
| `ayo sessions list` | List sessions |
| `ayo sessions show ID` | Show session details |
| `ayo sessions continue` | Resume session (picker) |
| `ayo sessions continue -l` | Resume most recent |
| `ayo sessions delete ID` | Delete session |

### Memory Commands

| Command | Description |
|---------|-------------|
| `ayo memory list` | List memories |
| `ayo memory search "query"` | Semantic search |
| `ayo memory show ID` | Show memory details |
| `ayo memory store "content"` | Store new memory |
| `ayo memory forget ID` | Forget a memory |
| `ayo memory stats` | Show statistics |
| `ayo memory clear` | Clear all memories |

### Flow Commands

| Command | Description |
|---------|-------------|
| `ayo flows list` | List flows |
| `ayo flows show name` | Show flow details |
| `ayo flows run name [input]` | Run flow |
| `ayo flows new name` | Create shell flow |
| `ayo flows new name --yaml` | Create YAML flow |
| `ayo flows validate file` | Validate YAML flow |
| `ayo flows history` | Show run history |
| `ayo flows replay ID` | Replay previous run |

### Sandbox Commands

| Command | Description |
|---------|-------------|
| `ayo sandbox list` | List sandboxes |
| `ayo sandbox show [ID]` | Show sandbox details |
| `ayo sandbox exec CMD` | Execute command |
| `ayo sandbox login` | Interactive shell |
| `ayo sandbox push SRC DEST` | Copy file to sandbox |
| `ayo sandbox pull SRC DEST` | Copy file from sandbox |
| `ayo sandbox diff SB HOST` | Show differences |
| `ayo sandbox sync SB HOST` | Sync changes to host |
| `ayo sandbox stop` | Stop sandbox |
| `ayo sandbox prune` | Remove stopped sandboxes |

### Share Commands

| Command | Description |
|---------|-------------|
| `ayo share PATH` | Share host directory |
| `ayo share PATH --as NAME` | Share with custom name |
| `ayo share list` | List shares |
| `ayo share rm NAME` | Remove share |
| `ayo share rm --all` | Remove all shares |

### Service Commands

| Command | Description |
|---------|-------------|
| `ayo sandbox service start` | Start daemon |
| `ayo sandbox service start -f` | Start in foreground |
| `ayo sandbox service stop` | Stop daemon |
| `ayo sandbox service status` | Show daemon status |

### Trigger Commands

| Command | Description |
|---------|-------------|
| `ayo triggers list` | List triggers |
| `ayo triggers show ID` | Show trigger details |
| `ayo triggers add --cron "..." --agent @a --prompt "..."` | Add cron trigger |
| `ayo triggers add --watch PATH --agent @a --prompt "..."` | Add watch trigger |
| `ayo triggers rm ID` | Remove trigger |
| `ayo triggers test ID` | Test trigger |
| `ayo triggers enable ID` | Enable trigger |
| `ayo triggers disable ID` | Disable trigger |

### Chain Commands

| Command | Description |
|---------|-------------|
| `ayo chain ls` | List chainable agents |
| `ayo chain inspect @agent` | Show schemas |
| `ayo chain from @agent` | Find compatible consumers |
| `ayo chain to @agent` | Find compatible producers |
| `ayo chain validate @agent JSON` | Validate input |
| `ayo chain example @agent` | Generate example input |

### Plugin Commands

| Command | Description |
|---------|-------------|
| `ayo plugins list` | List plugins |
| `ayo plugins install URL` | Install plugin |
| `ayo plugins show name` | Show plugin details |
| `ayo plugins update` | Update all plugins |
| `ayo plugins remove name` | Remove plugin |

### System Commands

| Command | Description |
|---------|-------------|
| `ayo setup` | Initial setup |
| `ayo doctor` | System health check |
| `ayo doctor -v` | Verbose health check |
| `ayo serve` | Start HTTP API server |

---

## Appendix C: Troubleshooting

### Agent Won't Start

1. Check configuration:
   ```bash
   ayo agents show @myagent
   ```

2. Verify model is available:
   ```bash
   ayo doctor -v
   ```

3. Check API key is set:
   ```bash
   echo $OPENAI_API_KEY  # or ANTHROPIC_API_KEY, etc.
   ```

### Sandbox Issues

1. Check sandbox status:
   ```bash
   ayo sandbox service status
   ```

2. Restart the daemon:
   ```bash
   ayo sandbox service stop
   ayo sandbox service start
   ```

3. Check for stale socket:
   ```bash
   rm -f ~/.local/share/ayo/daemon.sock
   ```

4. Verify sandbox provider:
   ```bash
   ayo doctor
   ```

### Memory Not Working

1. Check Ollama is running:
   ```bash
   ollama list
   ```

2. Verify embedding model:
   ```bash
   ollama pull nomic-embed-text
   ```

3. Check small model:
   ```bash
   ollama pull ministral-3:3b
   ```

### Trigger Not Firing

1. Verify daemon is running:
   ```bash
   ayo sandbox service status
   ```

2. Check trigger status:
   ```bash
   ayo triggers list
   ayo triggers show <id>
   ```

3. Test manually:
   ```bash
   ayo triggers test <id>
   ```

### Flow Failures

1. Check run history:
   ```bash
   ayo flows history --status=failed
   ```

2. Show run details:
   ```bash
   ayo flows history show <run-id>
   ```

3. Validate flow:
   ```bash
   ayo flows validate myflow.yaml
   ```

### Debug Mode

For detailed logging, use the `--debug` flag:

```bash
ayo --debug @myagent "test prompt"
```

This shows:
- LLM request/response details
- Tool call parameters and results
- Timing information
- Error traces

---

*This concludes the Ayo System Tutorial. For updates and community resources, visit the project repository.*
