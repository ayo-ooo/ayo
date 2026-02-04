# Manual Test Plan for Ayo Sandbox

This document provides a comprehensive manual test plan for validating the ayo sandbox implementation. Execute these commands in order, verifying expected behavior at each step.

## Prerequisites

Before starting, ensure:

1. You have Go 1.21+ installed
2. Docker is installed and running (for sandbox tests)
3. Ollama is installed with required models (for memory tests)
4. You're in the project root directory

```bash
cd /Users/alexcabrera/Code/ayo/ayo-sandbox
```

## Phase 1: Build and Install

### 1.1 Clean Build

```bash
# Clean any previous builds
go clean ./...

# Build all packages
go build ./...
```

**Expected:** No errors, clean build output.

### 1.2 Install Binary

```bash
# Install using the install script
./install.sh
```

**Expected:** Binary installed to `.local/bin/ayo`, built-ins installed to `.local/share/ayo/`.

### 1.3 Verify Installation

```bash
# Check version
.local/bin/ayo --version

# Check help
.local/bin/ayo --help
```

**Expected:** Version displayed, help text shows all commands.

---

## Phase 2: Unit Tests

### 2.1 Run All Unit Tests

```bash
go test ./internal/... 2>&1 | tee test-results.txt
```

**Expected:** All tests pass except known credential test failures in `internal/config`.

### 2.2 Run Tests with Coverage

```bash
make test-coverage
```

**Expected:** Coverage report generated, no new test failures.

### 2.3 Run Specific Package Tests

```bash
# Sandbox package
go test -v ./internal/sandbox/...

# Run package
go test -v ./internal/run/...

# Flows package
go test -v ./internal/flows/...

# Memory package
go test -v ./internal/memory/...

# UI packages
go test -v ./internal/ui/...
```

**Expected:** All tests pass in each package.

---

## Phase 3: Agent System

### 3.1 List Agents

```bash
.local/bin/ayo agents list
```

**Expected:** Shows `@ayo` built-in agent.

### 3.2 Show Agent Details

```bash
.local/bin/ayo agents show @ayo
```

**Expected:** Displays @ayo configuration, skills, and allowed tools.

### 3.3 Create Test Agent

```bash
.local/bin/ayo agents create @test-agent -m gpt-4o -n
```

**Expected:** Agent created successfully in `~/.config/ayo/agents/@test-agent/`.

### 3.4 Verify Agent Creation

```bash
.local/bin/ayo agents list
.local/bin/ayo agents show @test-agent
```

**Expected:** @test-agent appears in list with correct configuration.

### 3.5 Delete Test Agent

```bash
rm -rf ~/.config/ayo/agents/@test-agent
.local/bin/ayo agents list
```

**Expected:** @test-agent no longer appears.

---

## Phase 4: Skills System

### 4.1 List Skills

```bash
.local/bin/ayo skills list
```

**Expected:** Shows built-in skills (debugging, session-summary, ayo, project-summary).

### 4.2 Show Skill Details

```bash
.local/bin/ayo skills show debugging
.local/bin/ayo skills show ayo
```

**Expected:** Skill content and metadata displayed.

### 4.3 Create Test Skill

```bash
.local/bin/ayo skills create test-skill
```

**Expected:** Skill created in `~/.config/ayo/skills/test-skill/`.

### 4.4 Validate Skill

```bash
.local/bin/ayo skills validate ~/.config/ayo/skills/test-skill
```

**Expected:** Validation passes.

### 4.5 Cleanup Test Skill

```bash
rm -rf ~/.config/ayo/skills/test-skill
```

---

## Phase 5: Session Management

### 5.1 List Sessions

```bash
.local/bin/ayo sessions list
```

**Expected:** Shows existing sessions (or empty list if none).

### 5.2 Run Single Prompt (Creates Session)

```bash
.local/bin/ayo "echo hello from ayo test"
```

**Expected:** Agent responds, session created.

### 5.3 Verify Session Created

```bash
.local/bin/ayo sessions list
```

**Expected:** New session appears in list.

### 5.4 Show Session Details

```bash
# Get the most recent session ID from the list
.local/bin/ayo sessions list | head -5
# Then show details (replace SESSION_ID)
# .local/bin/ayo sessions show SESSION_ID
```

**Expected:** Session messages and metadata displayed.

---

## Phase 6: Memory System

### 6.1 Check Ollama Status

```bash
ollama list
```

**Expected:** Shows installed models. Required: `nomic-embed-text`, `qwen2.5:3b` (or similar small model).

### 6.2 Install Required Models (if missing)

```bash
ollama pull nomic-embed-text
ollama pull qwen2.5:3b
```

### 6.3 List Memories

```bash
.local/bin/ayo memory list
```

**Expected:** Shows existing memories (or empty list).

### 6.4 Store a Memory

```bash
.local/bin/ayo memory store "I prefer TypeScript over JavaScript for new projects"
```

**Expected:** Memory stored with auto-detected category.

### 6.5 Search Memories

```bash
.local/bin/ayo memory search "programming language preferences"
```

**Expected:** Returns the stored memory with similarity score.

### 6.6 Show Memory Stats

```bash
.local/bin/ayo memory stats
```

**Expected:** Shows memory count, categories, and embedding status.

### 6.7 Forget Memory

```bash
# Get memory ID from list
.local/bin/ayo memory list
# Forget it (replace MEMORY_ID)
# .local/bin/ayo memory forget MEMORY_ID
```

**Expected:** Memory soft-deleted.

---

## Phase 7: Flows System

### 7.1 List Flows

```bash
.local/bin/ayo flows list
```

**Expected:** Shows available flows (may be empty).

### 7.2 Create Test Flow

```bash
.local/bin/ayo flows new test-flow
```

**Expected:** Flow created in `.ayo/flows/test-flow.sh`.

### 7.3 Show Flow Details

```bash
.local/bin/ayo flows show test-flow
```

**Expected:** Flow metadata and script content displayed.

### 7.4 Validate Flow

```bash
.local/bin/ayo flows validate .ayo/flows/test-flow.sh
```

**Expected:** Validation passes.

### 7.5 Run Flow

```bash
.local/bin/ayo flows run test-flow '{"message": "hello"}'
```

**Expected:** Flow executes and returns result.

### 7.6 Check Flow History

```bash
.local/bin/ayo flows history
```

**Expected:** Shows the test flow run.

### 7.7 Cleanup Test Flow

```bash
rm -f .ayo/flows/test-flow.sh
```

---

## Phase 8: Sandbox System

### 8.1 Check Docker Status

```bash
docker info > /dev/null 2>&1 && echo "Docker is running" || echo "Docker is NOT running"
```

**Expected:** "Docker is running"

### 8.2 Run Sandbox Unit Tests

```bash
go test -v ./internal/sandbox/...
```

**Expected:** All sandbox tests pass.

### 8.3 Run Mount Tests

```bash
go test -v ./internal/sandbox/mounts/...
```

**Expected:** All mount tests pass.

### 8.4 Test Docker Provider (if Docker available)

```bash
# This is tested via integration tests
go test -v ./internal/integration/... -run TestDocker
```

**Expected:** Docker integration tests pass (or skip if Docker unavailable).

### 8.5 Test Apple Container Provider (macOS 15+ only)

```bash
# Check if Apple Container is available
which container && echo "Apple Container available" || echo "Apple Container not available"

# Run Apple Container tests (will skip if not available)
go test -v ./internal/sandbox/... -run TestApple
```

**Expected:** Tests pass or skip appropriately.

---

## Phase 9: Interactive Chat

### 9.1 Start Interactive Session

```bash
.local/bin/ayo
```

**Expected:** Chat interface starts with header "Chat with @ayo".

### 9.2 Test Basic Conversation

In the interactive session:
```
> What is 2 + 2?
```

**Expected:** Agent responds with "4" or similar.

### 9.3 Test Tool Usage

In the interactive session:
```
> Run the command: echo "Hello from bash"
```

**Expected:** Agent uses bash tool, shows spinner, displays output.

### 9.4 Test Multi-turn Conversation

```
> Remember that my favorite color is blue
> What did I just tell you about colors?
```

**Expected:** Agent recalls the color preference.

### 9.5 Exit Interactive Session

Press `Ctrl+C` twice to exit.

**Expected:** Clean exit, session saved.

---

## Phase 10: Piped/Non-Interactive Mode

### 10.1 Single Prompt Mode

```bash
.local/bin/ayo "What is the capital of France?"
```

**Expected:** Agent responds and exits.

### 10.2 Piped Input

```bash
echo "Summarize this: Hello world" | .local/bin/ayo
```

**Expected:** Agent processes piped input and responds.

### 10.3 JSON Output (with output schema agent)

```bash
# This requires an agent with output schema
# Skip if no such agent exists
```

---

## Phase 11: System Diagnostics

### 11.1 Run Doctor

```bash
.local/bin/ayo doctor
```

**Expected:** Shows system health status.

### 11.2 Verbose Doctor

```bash
.local/bin/ayo doctor -v
```

**Expected:** Detailed diagnostics including model list.

---

## Phase 12: Daemon (if implemented)

### 12.1 Check Daemon Status

```bash
.local/bin/ayo daemon status 2>/dev/null || echo "Daemon not running or not implemented"
```

### 12.2 Start Daemon

```bash
.local/bin/ayo daemon start 2>/dev/null || echo "Daemon start not available"
```

### 12.3 Stop Daemon

```bash
.local/bin/ayo daemon stop 2>/dev/null || echo "Daemon stop not available"
```

---

## Phase 13: Chain System

### 13.1 List Chainable Agents

```bash
.local/bin/ayo chain ls
```

**Expected:** Shows agents with input/output schemas (may be empty).

### 13.2 Inspect Agent Schemas

```bash
.local/bin/ayo chain inspect @ayo 2>/dev/null || echo "No schemas defined for @ayo"
```

---

## Phase 14: Plugin System

### 14.1 List Plugins

```bash
.local/bin/ayo plugins list
```

**Expected:** Shows installed plugins (may be empty).

---

## Phase 15: Cleanup

### 15.1 Remove Test Data

```bash
# Remove test sessions (optional - be careful)
# .local/bin/ayo sessions list | grep test | xargs -I {} .local/bin/ayo sessions delete {}

# Clear test memories
# .local/bin/ayo memory clear --force
```

### 15.2 Verify Clean State

```bash
.local/bin/ayo doctor
```

---

## Troubleshooting

### Build Failures

```bash
# Clean and rebuild
go clean -cache
go mod tidy
go build ./...
```

### Test Failures

```bash
# Run specific failing test with verbose output
go test -v -run TestName ./path/to/package/...
```

### Sandbox Issues

```bash
# Check Docker
docker ps
docker info

# Reset Docker (if needed)
docker system prune -f
```

### Memory/Embedding Issues

```bash
# Check Ollama
ollama list
ollama ps

# Restart Ollama
ollama stop
ollama serve &
```

### Database Issues

```bash
# Check database location
ls -la ~/.local/share/ayo/ayo.db

# Reset database (WARNING: destroys all data)
# rm ~/.local/share/ayo/ayo.db
```

---

## Test Results Checklist

Use this checklist to track test completion:

- [ ] Phase 1: Build and Install
- [ ] Phase 2: Unit Tests
- [ ] Phase 3: Agent System
- [ ] Phase 4: Skills System
- [ ] Phase 5: Session Management
- [ ] Phase 6: Memory System
- [ ] Phase 7: Flows System
- [ ] Phase 8: Sandbox System
- [ ] Phase 9: Interactive Chat
- [ ] Phase 10: Piped/Non-Interactive Mode
- [ ] Phase 11: System Diagnostics
- [ ] Phase 12: Daemon
- [ ] Phase 13: Chain System
- [ ] Phase 14: Plugin System
- [ ] Phase 15: Cleanup

---

## Reporting Issues

When reporting issues found during testing:

1. Note the exact phase and step number
2. Copy the full command executed
3. Copy the complete error output
4. Note any relevant system state (Docker running, Ollama status, etc.)
5. Check if the issue is reproducible

Create a ticket with:
```bash
cd .tickets
# Create ticket with details
```

Or report to the coding agent for immediate fix assistance.
