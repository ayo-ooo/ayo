---
id: smft-odvh
status: closed
deps: [smft-ega3]
links: []
created: 2026-02-12T23:53:28Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-w6k3
---
# AGENTS.md: Add debugging quick reference

Add a debugging quick reference to AGENTS.md.

**File:** AGENTS.md
**Insert after:** Code Conventions section

**Content to add:**
```markdown
## Debugging

Scripts in \`debug/\`:
- \`system-info.sh\` - Host system information
- \`sandbox-status.sh\` - Container status
- \`daemon-status.sh\` - Service status

Common issues:
- **Daemon won't start:** \`rm -f ~/.local/share/ayo/daemon.sock\` then restart
- **Sandbox not working:** Check \`ayo doctor\` output
- **Share not mounting:** Verify path exists, check \`ayo share list\`
```

