---
id: ase-79vs
status: closed
deps: []
links: []
created: 2026-02-07T03:25:29Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-6khq
---
# Add trigger ID picker and prefix matching

Improve trigger ID handling for better UX.

Changes:
1. Support short ID prefix matching (like sandbox commands)
   ayo trigger show trig  # matches trig_abc123

2. Add interactive picker when ID omitted
   ayo trigger rm         # shows picker if multiple triggers

3. Auto-select when only one trigger exists
   ayo trigger test       # auto-selects if only 1 trigger

Apply to: show, rm, test, enable, disable

## Acceptance Criteria

- Short ID prefixes work: `ayo trigger show tri`
- Picker shown when ID omitted and multiple exist
- Auto-select when only one trigger
- Clear error when no triggers exist

