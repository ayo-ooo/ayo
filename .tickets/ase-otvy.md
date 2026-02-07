---
id: ase-otvy
status: closed
deps: [ase-1wnw]
links: []
created: 2026-02-06T04:12:13Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-py58
---
# Intercept host-side tools before sandbox dispatch

Some tools (memory, agent_call) should execute on the host, not in the sandbox. Implement tool interception to route these correctly.

## Design

## Host-Side Tools
- memory: Queries host's ayo.db, uses embeddings
- agent_call: Orchestrates agent sessions on host
- publish: Copies files from sandbox to host
- file_request: Copies files from host to sandbox

## Implementation
1. Categorize tools as 'host' or 'sandbox'
2. In tool handler, check category before dispatch
3. Host tools execute locally with access to daemon services
4. Sandbox tools dispatch to sandbox.Exec()

## Tool Metadata
Add ExecutionContext to tool definitions:
type ToolExecutionContext string
const (
    ToolExecHost    ToolExecutionContext = 'host'
    ToolExecSandbox ToolExecutionContext = 'sandbox'
    ToolExecBridge  ToolExecutionContext = 'bridge' // needs both
)

## Memory Tool from Sandbox
When agent in sandbox calls 'memory search X':
1. Tool call comes to host
2. Host executes memory search
3. Results returned to LLM
Agent never directly accesses ayo.db.

## Bridge Tools
file_request and publish need both:
- Read/write host filesystem
- Read/write sandbox filesystem
These execute on host but use sandbox.Exec for sandbox-side operations.

## Acceptance Criteria

- Memory tool works from sandbox context
- agent_call works from sandbox context
- Tool routing is correct and tested
- Bridge tools can access both filesystems

