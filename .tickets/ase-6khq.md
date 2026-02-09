---
id: ase-6khq
status: open
deps: []
links: []
created: 2026-02-07T03:24:58Z
type: epic
priority: 2
assignee: Alex Cabrera
parent: ase-zlew
---
# Trigger CLI Redesign

Redesign trigger CLI for simpler, more intuitive UX.

Current (verbose):
  ayo triggers add --type cron --agent @backup --schedule "0 0 2 * * *"
  ayo triggers add --type watch --agent @build --path ./src --patterns "*.go"

Proposed structure:
  ayo trigger                    # list triggers (default action)
  ayo trigger schedule @agent "schedule"   # create cron trigger
  ayo trigger watch <path> @agent [patterns] # create watch trigger
  ayo trigger show <id>          # show details
  ayo trigger rm <id>            # remove
  ayo trigger test <id>          # fire manually
  ayo trigger enable/disable <id>

Key changes:
1. Singular `trigger` (not `triggers`)
2. `schedule` subcommand for cron (clearer than `cron`)
3. `watch` subcommand for filesystem
4. Positional args instead of verbose flags
5. ID picker when omitted
6. Short ID prefix matching

