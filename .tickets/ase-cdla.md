---
id: ase-cdla
status: open
deps: [ase-yqtq]
links: []
created: 2026-02-09T03:06:12Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-gw5j
---
# Implement agent refinement based on usage

Enable @ayo to refine agents it created based on usage patterns and outcomes.

## Background

@ayo-created agents improve over time:
- After successful uses, @ayo may add clarifying instructions
- After failures, @ayo may adjust approach
- Refinements are tracked in SQLite for transparency

## Implementation

1. Add `ayo agents refine` command:
   ```bash
   ayo agents refine science-researcher \
     --append-system 'When discussing biology, cite recent papers.' \
     --note 'User prefers academic sources'
   ```

2. Track refinements in ayo_created_agents.refinement_notes (JSON array):
   ```json
   [
     {"timestamp": "...", "change": "append", "content": "...", "reason": "..."},
     ...
   ]
   ```

3. Increment system_prompt_version on each refinement

4. Add @ayo skill for refinement decisions:
   ```markdown
   ## Refining Agents
   
   After using an agent you created, consider refinement if:
   - The output needed correction that a prompt change could prevent
   - User gave feedback on preferences
   - You notice a pattern in how the agent should behave
   
   Use: `ayo agents refine <name> --append-system '...' --note '...'`
   
   Always include a note explaining why you're refining.
   ```

5. Only allow refinement of @ayo-created agents (check SQLite)

## Files to modify/create

- cmd/ayo/agents.go (add refine subcommand)
- internal/agent/refine.go (new)
- internal/database/repository.go (add refinement tracking)
- internal/builtin/skills/ayo/SKILL.md (add refinement guidance)

## Acceptance Criteria

- Refinement modifies AGENT.md correctly
- Refinement history tracked in SQLite
- Version incremented
- Only works on @ayo-created agents
- @ayo skill provides clear guidance

