---
id: ase-enaz
status: closed
deps: [ase-2msm]
links: []
created: 2026-02-06T04:15:46Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-ka3q
---
# Update sandbox skill documentation

Update the sandbox skill with new IRC messaging, file sharing, and directory structure information.

## Design

## Skill Location
internal/builtin/skills/sandbox/SKILL.md

## Content to Add
1. Directory structure explanation
2. IRC messaging instructions
3. File sharing between agents
4. Environment variables
5. Helper commands (msg, irc-log)

## Directory Structure Section
/home/{agent}/     - Your home directory
/shared/           - Shared files (all agents can access)
/workspaces/{id}/  - Current session workspace
/mnt/host/         - Mounted host files

## IRC Section
Inter-agent communication via IRC:
- Send message: msg '#general' 'your message'
- Send DM: msg '@crush' 'your message'
- Read logs: irc-log [channel] [lines]
- Your nick is your agent handle

## File Sharing Section
To share files with other agents:
1. Copy to /shared/ for permanent sharing
2. Copy to /shared/ for session sharing
3. Notify via IRC: msg '#general' 'File ready at /shared/output.json'

## Environment Variables
  - Current session workspace path
 - Current session ID
      - Your agent handle

## Acceptance Criteria

- Skill document updated
- IRC instructions clear
- File sharing documented
- Examples provided

