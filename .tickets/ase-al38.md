---
id: ase-al38
status: closed
deps: [ase-ncrx]
links: []
created: 2026-02-09T03:27:03Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-k48b
---
# Implement flow history and stats CLI

## Background

Users need visibility into flow execution history and statistics. This is separate from running flows - it's about observing past and ongoing executions.

## Commands

```bash
# List recent flow runs
ayo flows history                     # Last 20 runs
ayo flows history --limit 50          # More runs
ayo flows history --flow daily-digest # Filter by flow name
ayo flows history --status failed     # Filter by status
ayo flows history --since 24h         # Time filter
ayo flows history --json              # JSON output

# Show details of specific run
ayo flows run-show <run-id>           # Full details
ayo flows run-show <run-id> --json

# Flow statistics
ayo flows stats                       # All flows
ayo flows stats daily-digest          # Specific flow
ayo flows stats --json
```

## Output Examples

### flows history

```
ID           FLOW           STATUS    DURATION   STARTED
f3a2b1c0     daily-digest   ✓ success 2m 34s     2 hours ago
e4c3d2b1     process-data   ✗ failed  45s        3 hours ago
d5e4f3c2     daily-digest   ✓ success 2m 12s     1 day ago

Showing 3 of 47 runs. Use --limit for more.
```

### flows run-show

```
Run ID: f3a2b1c0
Flow: daily-digest
Status: success
Started: 2026-02-08T20:30:00Z
Duration: 2m 34s
Triggered by: cron (morning-run)

Parameters:
  language: spanish

Steps:
  ✓ gather      12s     exit 0
  ✓ summarize   1m 20s  @summarizer responded
  ✓ translate   45s     @translator responded
  ✓ format      17s     @formatter responded

Output:
  [truncated, use --full for complete output]
```

### flows stats

```
Flow: daily-digest
Total Runs: 47
Success Rate: 95.7% (45/47)
Avg Duration: 2m 28s
Last Run: 2 hours ago (success)

Trigger Stats:
  morning-run (cron): 45 runs, 95.6% success
  manual: 2 runs, 100% success
```

## Implementation

### Data Sources

- flow_runs table for execution history
- trigger_stats table for trigger-specific stats

### Files to Modify

1. `cmd/ayo/flows.go` - Add history, run-show, stats subcommands
2. Create `internal/flows/stats.go` - Stats calculation

## Acceptance Criteria

- [ ] ayo flows history lists recent runs
- [ ] Filtering by flow, status, time works
- [ ] ayo flows run-show shows full run details
- [ ] Step-by-step execution visible
- [ ] ayo flows stats shows aggregate statistics
- [ ] Trigger-specific stats included
- [ ] JSON output for all commands
- [ ] Color-coded status indicators

