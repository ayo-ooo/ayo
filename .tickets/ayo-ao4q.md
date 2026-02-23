---
id: ayo-ao4q
status: open
deps: []
links: []
created: 2026-02-23T23:13:09Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [sandbox, agents]
---
# Implement shared sandbox with per-agent home directories

Establish the @ayo sandbox as the default shared workspace for all agents.

## Design

```
@AYO SANDBOX (shared by default):
/home/
├── ayo/              # @ayo home
├── crush/            # @crush home (created on first use)
├── reviewer/         # @reviewer home
└── {agent-name}/     # New agents get directories automatically

/mnt/{user}/          # Host home (read-only)
/workspace/           # Shared workspace
/output/              # Safe write zone
```

## Key Behaviors

1. **All agents share @ayo sandbox by default**
   - When you run `ayo @crush "write code"`, @crush executes in @ayo's sandbox
   - File handoff between agents is natural (same filesystem)
   
2. **Per-agent home directories**
   - Each agent's `$HOME` is set to `/home/{agent-name}`
   - Directory created on first use
   - Agents can still access other agents' home dirs if needed

3. **Isolation opt-in**
   - Agents can request isolation via `sandbox.isolated: true` in ayo.json
   - Squads always get their own sandbox

## Implementation

### Files to Modify

1. `internal/sandbox/providers/apple.go`
   - Modify `CreateSandbox` to handle shared sandbox mode
   - Add `EnsureAgentHome(sandboxID, agentName)` function
   - Update `Execute` to set correct HOME env var

2. `internal/sandbox/providers/linux.go`
   - Same changes for systemd-nspawn provider

3. `internal/sandbox/sandbox.go`
   - Add `GetSharedSandboxID()` function
   - Add logic to determine shared vs isolated

### New Functions

```go
// Get or create the shared @ayo sandbox
func (p *Provider) GetSharedSandbox() (string, error)

// Ensure /home/{agent} exists in sandbox
func (p *Provider) EnsureAgentHome(sandboxID, agentName string) error

// Execute command with agent's HOME set
func (p *Provider) ExecuteAsAgent(sandboxID, agentName string, cmd []string) error
```

### Sandbox Lifecycle

1. On first agent invocation, create @ayo sandbox if not exists
2. Create `/home/{agent-name}` directory in sandbox
3. Execute commands with `HOME=/home/{agent-name}`
4. Sandbox persists across sessions (not ephemeral)

## Testing

- Test creating home directories for new agents
- Test file visibility between agents
- Test isolated agents don't share sandbox
- Test sandbox persistence across daemon restarts
