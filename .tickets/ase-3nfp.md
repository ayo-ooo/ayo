---
id: ase-3nfp
status: closed
deps: [ase-uvhc]
links: []
created: 2026-02-06T04:15:30Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-py58
---
# End-to-end tests for sandbox execution

Add E2E tests for the full sandbox execution flow from agent to bash command.

## Design

## Test Coverage
1. Agent with sandbox enabled executes in sandbox
2. Memory tool works from sandbox context
3. File operations scoped to workspace
4. Mount permissions enforced

## Test Structure
internal/integration/e2e_sandbox_test.go

## Tests
func TestE2E_AgentExecutesInSandbox(t *testing.T)
  - Create agent with sandbox enabled
  - Execute bash command
  - Verify execution was in sandbox (check hostname, paths)

func TestE2E_MemoryFromSandbox(t *testing.T)
  - Store memory on host
  - Run agent in sandbox
  - Agent searches memory
  - Verify results returned

func TestE2E_WorkspaceIsolation(t *testing.T)
  - Start session
  - Create file in workspace
  - Verify file exists at /workspaces/{session}/
  - Verify file not accessible from other sessions

func TestE2E_MountPermissions(t *testing.T)
  - Grant mount permission
  - Verify agent can access mounted path
  - Revoke permission
  - Verify agent cannot access (or mount not present)

## Environment
May need real sandbox for full E2E.
Provide mock fallback for CI.

## Acceptance Criteria

- Full execution flow tested
- Memory tool tested
- Workspace isolation verified
- Mount permissions enforced

