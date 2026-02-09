---
id: ase-cy5l
status: open
deps: [ase-0oyk]
links: []
created: 2026-02-09T03:10:53Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-cjpe
---
# Implement ayo agents capabilities CLI

## Background

Users need visibility into what capabilities the system has inferred for each agent. This enables debugging when @ayo selects unexpected agents, and allows users to understand how their agents are perceived by the system.

## Why This Matters

Without this CLI:
- Users can't see what @ayo thinks an agent can do
- Debugging agent selection is opaque
- Users can't manually refresh stale capabilities
- No way to verify capability inference is working

## Implementation Details

### Commands

```bash
# List capabilities for a specific agent
ayo agents capabilities <agent-name>

# List all capabilities across all agents
ayo agents capabilities --all

# Search capabilities by description
ayo agents capabilities --search "code review"

# Refresh capabilities for an agent (re-run inference)
ayo agents capabilities refresh <agent-name>

# Refresh all agents
ayo agents capabilities refresh --all
```

### Output Formats

**Default (human-readable):**
```
Capabilities for @code-reviewer:

  code-review (confidence: 0.95)
    Reviews source code for issues and improvements
    Source: system_prompt

  security-analysis (confidence: 0.87)
    Identifies security vulnerabilities in code
    Source: skill:security-scanner

  best-practices (confidence: 0.72)
    Suggests coding best practices and patterns
    Source: system_prompt

Last updated: 2 hours ago
Input hash: a1b2c3d4...
```

**JSON (--json flag):**
```json
{
  "agent": "@code-reviewer",
  "capabilities": [
    {
      "name": "code-review",
      "description": "Reviews source code for issues and improvements",
      "confidence": 0.95,
      "source": "system_prompt"
    }
  ],
  "last_updated": "2024-01-15T10:30:00Z",
  "input_hash": "a1b2c3d4..."
}
```

### Search Output

```bash
$ ayo agents capabilities --search "review code"

Agents matching "review code":

  @code-reviewer (similarity: 0.94)
    code-review: Reviews source code for issues and improvements

  @security-auditor (similarity: 0.78)
    code-audit: Audits code for security compliance

  @tech-lead (similarity: 0.65)
    pr-review: Reviews pull requests and provides feedback
```

### Files to Modify

1. `cmd/ayo/agents.go` - Add `capabilities` subcommand
2. `cmd/ayo/agents_capabilities.go` - Implementation (new file)
3. Use `internal/capabilities/` package for backend

### Implementation Notes

- Use `globalOutput` for JSON/quiet mode support
- Color-code confidence levels (green >0.8, yellow 0.5-0.8, red <0.5)
- Show "No capabilities inferred" message for agents without any
- Include stale warning if capabilities older than agent definition hash

## Acceptance Criteria

- [ ] ayo agents capabilities <agent> shows inferred capabilities
- [ ] --all flag lists capabilities for all agents
- [ ] --search flag does semantic search across capabilities
- [ ] refresh subcommand re-runs inference
- [ ] --json flag outputs structured JSON
- [ ] Confidence color-coded in terminal output
- [ ] Shows last updated timestamp and input hash
- [ ] Stale warning when hash mismatch detected

