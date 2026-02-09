---
id: ase-1hkx
status: closed
deps: [ase-yqtq]
links: []
created: 2026-02-09T03:06:42Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-gw5j
---
# Implement agent promotion and archival

Allow users to promote @ayo-created agents to user-owned, and @ayo to archive low-use agents.

## Background

Agent lifecycle:
- @ayo creates agent → tracked as @ayo-created
- Agent is used, metrics accumulate
- Success: user may promote to take ownership
- Low use: @ayo may archive (hidden but not deleted)

## Implementation

### Promotion

```bash
ayo agents promote science-researcher my-science-helper
```

1. Check agent is @ayo-created (in SQLite table)
2. Rename agent directory
3. Update agent.json (remove any @ayo metadata)
4. Update SQLite: set promoted_to field
5. Agent is now user-owned, @ayo can't refine it

### Archival

```bash
ayo agents archive old-helper      # Hide agent
ayo agents unarchive old-helper    # Bring back
ayo agents list --archived         # Show archived
```

1. Set archived=true in SQLite
2. Archived agents:
   - Hidden from normal `ayo agents list`
   - Still exist on disk
   - Can be unarchived
   - NOT included in capability search

3. @ayo can archive agents with:
   - 0 uses in last 30 days
   - Low confidence score
   - Include in @ayo skill guidance

### CLI updates

```bash
ayo agents list                  # Normal + @ayo-created
ayo agents list --archived       # Include archived
ayo agents show <name>           # Shows creation info if @ayo-created
```

## Files to modify/create

- cmd/ayo/agents.go (add promote, archive, unarchive)
- internal/agent/lifecycle.go (new)
- internal/database/repository.go (archive/promote methods)
- internal/builtin/skills/ayo/SKILL.md (archive guidance)

## Acceptance Criteria

- Promotion renames and transfers ownership
- Archival hides from list and capabilities
- Unarchival restores
- @ayo skill explains when to archive

