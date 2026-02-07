---
id: ase-ji7h
status: closed
deps: [ase-fw7m]
links: []
created: 2026-02-06T04:10:52Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-rzhr
---
# Add trigger engine to daemon

Implement the trigger engine in the daemon that watches for trigger conditions and spawns agent sessions.

## Design

## Trigger Engine Components
1. TriggerRegistry - stores registered triggers
2. TriggerWatcher - monitors conditions
3. TriggerExecutor - spawns sessions when conditions met

## Trigger Sources
- Agent config.json (triggers field)
- Flow frontmatter (triggers field)
- Global ayo.json (triggers field)

## Trigger Types (Phase 1)
- cron: time-based using robfig/cron
- watch: file system changes using fsnotify

## Trigger Lifecycle
1. On daemon start, load all triggers from sources
2. Start watchers for each trigger type
3. On condition match:
   a. Create context with trigger metadata
   b. Spawn agent session with injected prompt
   c. Log trigger execution
4. On daemon stop, cleanup watchers

## Data Structures
type Trigger struct {
    ID        string
    Type      string // cron, watch, webhook, irc
    Agent     string
    Config    map[string]any
    Prompt    string
    Source    string // path to config that defined it
}

type TriggerEvent struct {
    TriggerID  string
    FiredAt    time.Time
    Context    map[string]any
}

## Acceptance Criteria

- Triggers loaded from all sources on daemon start
- Cron triggers fire on schedule
- Watch triggers fire on file changes
- Agent sessions spawned with context injection
- Trigger history logged

