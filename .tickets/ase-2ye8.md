---
id: ase-2ye8
status: closed
deps: [ase-euxv, ase-qd2x]
links: []
created: 2026-02-09T03:07:58Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-k48b
---
# Implement flow triggers

Integrate trigger definitions from flow files with the trigger engine.

## Background

Triggers are defined in flow YAML files:
```yaml
triggers:
  - id: morning-run
    type: cron
    schedule: '0 9 * * *'
    params:
      language: spanish
    runs_before_permanent: 5
```

The daemon needs to:
1. Load triggers from flow files on startup
2. Register them with the trigger engine
3. Track stats in SQLite (trigger_stats table)
4. Handle 'runs_before_permanent' expiry

## Implementation

1. On daemon start:
   - Scan all flow files
   - Extract trigger definitions
   - Register with trigger engine

2. On flow file change (via fsnotify):
   - Re-parse flow
   - Update trigger registrations

3. When trigger fires:
   - Execute flow with specified params
   - Record run in trigger_stats
   - Check runs_before_permanent:
     - If runs_completed >= runs_before_permanent, set permanent=true
     - If not permanent and trigger has run too many times, disable it

4. Trigger callback:
   ```go
   func (d *Daemon) onFlowTrigger(flowName string, triggerID string, params map[string]any) {
       // Execute flow
       // Record stats
   }
   ```

5. Remove 'ayo triggers add' command (triggers only via flows)
   - Update existing trigger CLI to be read-only for flow triggers
   - Can still pause/resume flow triggers

## Files to modify

- internal/daemon/server.go (load flow triggers on start)
- internal/daemon/trigger_engine.go (add flow trigger type)
- internal/flows/triggers.go (new - trigger extraction)
- cmd/ayo/triggers.go (make add command flow-only aware)

## Acceptance Criteria

- Triggers loaded from flow files on startup
- Hot reload on flow file change
- Trigger stats recorded in SQLite
- runs_before_permanent works correctly
- Triggers can be paused/resumed via CLI

