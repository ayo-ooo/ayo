---
id: ase-uvhc
status: closed
deps: [ase-1wnw]
links: []
created: 2026-02-06T04:12:22Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-py58
---
# Create session workspace in sandbox

Each session should have a dedicated workspace directory in the sandbox for isolated work.

## Design

## Session Workspace
/workspaces/{session-id}/
├── mounted/          # Symlinks to mounted paths
├── scratch/          # Temp files, agent outputs
└── shared/           # Inter-agent files for this session

## Lifecycle
1. On session start: Create workspace directory
2. Set as default working directory for agent
3. Apply mounts to mounted/ subdirectory
4. On session end: Archive or cleanup

## Implementation
1. Add createSessionWorkspace() to Runner
2. Call when session created
3. Pass workspace path to sandbox executor
4. Set WORKSPACE env var in sandbox

## Environment Variables
Inject into sandbox environment:
- WORKSPACE=/workspaces/{session-id}
- SESSION_ID={session-id}
- AGENT={agent-handle}

## Cleanup Policy
Configurable in ayo.json:
- workspace_retention: '7d' (delete after 7 days)
- workspace_archive: true (compress before delete)

## Integration with Working Copy
If using working copy model, session workspace is the target.

## Acceptance Criteria

- Workspace created on session start
- Correct environment variables set
- Workspace cleanup works
- Mounted paths accessible

