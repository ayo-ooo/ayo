---
id: ase-o8c9
status: closed
deps: []
links: []
created: 2026-02-09T03:08:18Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-fked
---
# Implement trust levels for agents

Add trust level field to agent configuration with three levels: sandboxed, privileged, unrestricted.

## Background

Trust levels control agent capabilities and visibility:

| Level | Guardrails | Sandbox | @ayo can orchestrate |
|-------|------------|---------|---------------------|
| sandboxed (default) | Yes | Full | Yes |
| privileged | Yes | Host access | Yes |
| unrestricted | No | None | **No** |

Key behavior: Unrestricted agents are invisible to @ayo's capability discovery and cannot be orchestrated.

## Implementation

1. Add trust field to agent.json schema:
   ```json
   {
     "trust": "sandboxed"  // or 'privileged' or 'unrestricted'
   }
   ```

2. Default to 'sandboxed' if not specified

3. When creating unrestricted agent, require explicit flag:
   ```bash
   ayo agents new my-unsafe --unrestricted
   # Warning: Unrestricted agents run without safety guardrails
   # and cannot be orchestrated by @ayo. Continue? [y/N]
   ```

4. Update agent loading to parse trust level

5. Update sandbox creation based on trust:
   - sandboxed: Full sandbox, no host mounts
   - privileged: Sandbox with host mounts (from grants)
   - unrestricted: No sandbox, runs on host

6. Update capability discovery to exclude unrestricted agents

7. Update CLI display:
   ```
   $ ayo agents list
     @ayo           builtin    sandboxed
     @researcher    plugin     sandboxed
     @my-unsafe     user       ⚠ unrestricted
   ```

## Files to modify

- internal/agent/agent.go (add TrustLevel field)
- internal/agent/load.go (parse trust level)
- cmd/ayo/agents.go (add --unrestricted flag, update display)
- internal/sandbox/ (respect trust level)
- internal/capabilities/ (exclude unrestricted)

## Acceptance Criteria

- Trust level parsed from agent.json
- Default is sandboxed
- Unrestricted creation requires confirmation
- Sandbox respects trust level
- Unrestricted agents excluded from capabilities
- CLI shows trust level clearly

