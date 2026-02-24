# Ticket Worker Pattern

The ticket worker pattern creates an agent that continuously processes work items from a ticket queue. This enables autonomous, prioritized task execution.

## Overview

```
Ticket Queue → Worker Polls → Picks Task → Executes → Closes Ticket
     ↑                                          |
     └──────────────────────────────────────────┘
                  (continuous loop)
```

## Basic Setup

```bash
# Poll for work every 5 minutes
ayo trigger schedule @worker "*/5 * * * *" \
  --prompt "Check for ready tickets assigned to you. Pick the highest priority one and complete it." \
  --singleton
```

## How It Works

1. **Trigger fires** at interval (e.g., every 5 minutes)
2. **Agent queries** for ready tickets: `ayo ticket ready --assignee @worker`
3. **Agent picks** highest priority ticket
4. **Agent works** on the ticket
5. **Agent closes** ticket when done
6. **Cycle repeats** on next trigger

## Configuration Options

| Option | Description | Recommended |
|--------|-------------|-------------|
| `--singleton` | Prevent concurrent runs | Always use |
| `--timeout` | Maximum execution time | Set based on task complexity |
| Interval | How often to poll | 1-15 minutes |

## Examples

### Code Review Worker

Automatically review ready PRs:

```bash
ayo trigger schedule @reviewer "*/10 * * * *" \
  --prompt "Check for 'ready-for-review' tickets assigned to you.
    For each ticket:
    1. Review the associated code changes
    2. Leave comments on issues found
    3. Approve if no issues, or request changes
    4. Update ticket status appropriately" \
  --singleton \
  --timeout 30m
```

### Bug Fix Worker

Process bug reports:

```bash
ayo trigger schedule @fixer "*/15 * * * *" \
  --prompt "Check for bug tickets assigned to you.
    For the highest priority bug:
    1. Analyze the bug report
    2. Find the root cause
    3. Implement a fix
    4. Write tests for the fix
    5. Close the ticket with a summary" \
  --singleton \
  --timeout 1h
```

### Documentation Worker

Keep docs updated:

```bash
ayo trigger schedule @documenter "0 * * * *" \
  --prompt "Check for documentation tickets.
    For each ready ticket:
    1. Identify what needs documenting
    2. Write or update documentation
    3. Ensure examples work
    4. Close ticket with changes made" \
  --singleton \
  --timeout 30m
```

### Test Writer

Generate missing tests:

```bash
ayo trigger schedule @tester "*/30 * * * *" \
  --prompt "Check for 'needs-tests' tickets.
    For the highest priority:
    1. Analyze the code that needs testing
    2. Write comprehensive test cases
    3. Ensure tests pass
    4. Close ticket with test summary" \
  --singleton \
  --timeout 45m
```

## Creating Work Tickets

Workers need tickets to process:

```bash
# Create work items
ayo ticket create "Review PR #123" --assignee @reviewer --priority high
ayo ticket create "Fix login bug" --assignee @fixer --priority critical
ayo ticket create "Document API endpoints" --assignee @documenter
```

### With Dependencies

```bash
# Design must complete before implementation
DESIGN=$(ayo ticket create "Design user schema" --assignee @architect --json | jq -r '.id')
ayo ticket create "Implement user API" --assignee @worker --depends-on $DESIGN
```

The worker won't see "Implement user API" until "Design user schema" is closed.

## Agent Prompt Best Practices

### Query for Work

```
Good: "Check for ready tickets assigned to you using the ticket tool"
Bad:  "Look for work to do"
```

### Handle No Work

```
Good: "If no tickets are ready, respond with 'No work available' and exit"
Bad:  (no guidance - agent might hallucinate work)
```

### Update Status

```
Good: "Start the ticket before working, close when done with a summary"
Bad:  "Work on tickets"
```

### Prioritize Correctly

```
Good: "Pick the highest priority ticket that is ready"
Bad:  "Pick any ticket"
```

## Singleton Mode (Critical)

**Always use singleton mode for ticket workers:**

```bash
ayo trigger schedule @worker "* * * * *" \
  --singleton  # REQUIRED
```

Without singleton:
```
Minute 1: Worker starts on ticket A (takes 5 min)
Minute 2: ANOTHER worker starts, picks same ticket A!
Minute 3: ANOTHER worker starts... chaos!
```

With singleton:
```
Minute 1: Worker starts on ticket A
Minute 2-5: Trigger skipped (worker running)
Minute 6: Worker finishes, ready for next trigger
```

## Squad-Based Workers

For complex tickets, use squads:

```bash
# Create feature squad
ayo squad create feature-team -a @designer,@developer,@tester

# Worker dispatches to squad
ayo trigger schedule @coordinator "*/15 * * * *" \
  --prompt "Check for feature tickets. 
    Dispatch each to #feature-team with appropriate subtasks." \
  --singleton
```

## Monitoring Workers

### Check Worker Status

```bash
# List triggers
ayo trigger list

# See worker's recent sessions
ayo session list --agent @worker

# View specific run
ayo session show <session-id>
```

### Check Ticket Progress

```bash
# Open tickets
ayo ticket list --status open

# Blocked tickets
ayo ticket blocked

# Recently closed
ayo ticket list --status closed --limit 10
```

## Troubleshooting

### Worker Not Processing Tickets

1. Check trigger is enabled:
   ```bash
   ayo trigger show <id>
   ```

2. Verify tickets exist:
   ```bash
   ayo ticket ready --assignee @worker
   ```

3. Check worker sessions for errors:
   ```bash
   ayo session list --agent @worker
   ```

### Worker Picking Wrong Tickets

- Verify ticket assignment: `ayo ticket show <id>`
- Check priority values
- Review agent prompt for selection logic

### Worker Timing Out

- Increase timeout value
- Break large tasks into smaller tickets
- Add progress checkpoints to prompt

### Duplicate Work

- **Enable singleton mode**
- Check for multiple workers on same queue

## Best Practices

1. **Always use singleton mode** to prevent concurrent runs
2. **Set realistic timeouts** based on task complexity
3. **Clear ticket assignments** so workers know what to pick
4. **Use dependencies** to enforce task ordering
5. **Monitor early** to catch issues before they compound
6. **Start with longer intervals** (15+ min) and reduce as needed

## See Also

- [Tickets Guide](../guides/triggers.md)
- [Scheduled Pattern](scheduled.md)
- [Squads Guide](../guides/squads.md)
