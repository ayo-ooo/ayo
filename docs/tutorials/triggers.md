# Tutorial: Event-Driven Agents with Triggers

Set up agents that run automatically on schedules and file changes. By the end, you'll have ambient agents working in the background.

**Time**: ~20 minutes  
**Prerequisites**: [First Agent Tutorial](first-agent.md) complete

## What You'll Build

- A daily standup agent that summarizes git activity
- A file watcher that reviews changed code
- Understanding of all trigger types

## Step 1: Create a Scheduled Agent

First, create an agent for daily standups:

```bash
ayo agent create @standup
```

Edit `~/.config/ayo/agents/@standup/system.md`:

```markdown
# Daily Standup Agent

You generate concise daily standup summaries.

## Output Format

```
## Daily Standup - [date]

### Completed Yesterday
- [list of completed items]

### In Progress
- [ongoing work]

### Blockers
- [any blockers, or "None"]
```

## Guidelines
- Focus on meaningful changes, not trivial commits
- Group related commits together
- Identify patterns (lots of bug fixes = quality issues)
- Keep it scannable - busy people will read this
```

## Step 2: Create a Cron Trigger

Schedule the standup for 9am every weekday:

```bash
ayo trigger create morning-standup \
  --cron "0 9 * * MON-FRI" \
  --agent @standup \
  --prompt "Summarize yesterday's git commits in ~/Projects"
```

### Cron Expression Breakdown

```
0 9 * * MON-FRI
│ │ │ │ │
│ │ │ │ └── Days: Monday through Friday
│ │ │ └──── Month: Every month
│ │ └────── Day: Every day
│ └──────── Hour: 9am
└────────── Minute: 0
```

### Using Aliases

For common schedules, use aliases:

```bash
# Every hour
ayo trigger create hourly-check --cron "@hourly" --agent @monitor

# Every day at midnight
ayo trigger create nightly-backup --cron "@daily" --agent @backup

# Every Monday at midnight
ayo trigger create weekly-report --cron "@weekly" --agent @reporter
```

## Step 3: Create a File Watch Agent

Create an agent that reviews changed files:

```bash
ayo agent create @watcher
```

Edit `~/.config/ayo/agents/@watcher/system.md`:

```markdown
# File Watch Agent

You monitor code changes and provide quick feedback.

## On File Change

When files change:
1. Identify what changed (new file, modification, deletion)
2. Quick scan for obvious issues
3. Note if changes need review

## Response Format

Keep responses brief:

```
[filename]: [status emoji] [one-line summary]
```

Status emojis:
- ✅ Looks good
- ⚠️ Needs review
- 🐛 Potential bug
- 🔒 Security concern
```

## Step 4: Create a Watch Trigger

Watch for Go file changes:

```bash
ayo trigger create go-watcher \
  --watch ~/Projects/myapp/src \
  --pattern "*.go" \
  --agent @watcher \
  --prompt "Review the changed files"
```

### Watch Options

```bash
# Watch multiple patterns
ayo trigger create code-watcher \
  --watch ./src \
  --pattern "*.go,*.ts,*.py" \
  --agent @watcher

# Specific events only
ayo trigger create new-files \
  --watch ./uploads \
  --events create \
  --agent @processor

# With debouncing (wait for changes to settle)
ayo trigger create build-watcher \
  --watch ./src \
  --debounce 5s \
  --agent @builder
```

## Step 5: Create an Interval Trigger

For recurring checks that aren't cron-based:

```bash
ayo trigger create health-check \
  --interval 30m \
  --agent @monitor \
  --prompt "Check system health and report any issues"
```

## Step 6: Create a One-Time Trigger

Schedule a single future execution:

```bash
ayo trigger create reminder \
  --once "2024-12-25 09:00" \
  --agent @ayo \
  --prompt "Remind me to send holiday greetings"
```

## Step 7: Manage Triggers

### List All Triggers

```bash
ayo trigger list
```

Example output:
```
NAME              TYPE    AGENT      SCHEDULE/PATH            ENABLED
morning-standup   cron    @standup   0 9 * * MON-FRI         true
go-watcher        watch   @watcher   ~/Projects/myapp/src    true
health-check      interval @monitor  30m                      true
```

### View Trigger Details

```bash
ayo trigger show morning-standup
```

### Disable/Enable Triggers

```bash
ayo trigger disable morning-standup
ayo trigger enable morning-standup
```

### Remove a Trigger

```bash
ayo trigger remove morning-standup
```

## Step 8: Test Triggers

### Manual Execution

Test a trigger without waiting:

```bash
ayo trigger fire morning-standup
```

### View Execution History

```bash
ayo trigger history
```

Example output:
```
TRIGGER           FIRED_AT              STATUS    DURATION
morning-standup   2024-01-15 09:00:01   success   12.3s
go-watcher        2024-01-15 08:45:22   success   3.1s
go-watcher        2024-01-15 08:42:11   success   2.8s
```

### View Specific Trigger History

```bash
ayo trigger history morning-standup --limit 10
```

## Understanding Trigger Execution

### Execution Context

When a trigger fires:
1. Daemon receives the event
2. Agent is invoked with the configured prompt
3. For watch triggers, changed files are included in context
4. Output is logged and optionally notified

### Singleton Mode

Prevent overlapping executions:

```bash
ayo trigger create slow-job \
  --cron "*/5 * * * *" \
  --agent @processor \
  --singleton \
  --prompt "Process queue"
```

If the previous run hasn't finished, the new trigger is skipped.

### Error Handling

Failed triggers are logged but don't stop future executions:

```bash
# View failed triggers
ayo trigger history --status failed
```

## Complete Example: CI/CD Agent

Create a comprehensive code quality pipeline:

```bash
# Create the agent
ayo agent create @ci

# Watch for pushes to main files
ayo trigger create ci-pipeline \
  --watch ./src \
  --pattern "*.go" \
  --debounce 10s \
  --singleton \
  --agent @ci \
  --prompt "Run tests, check for lint issues, and report results"
```

**@ci system.md**:
```markdown
# CI Agent

You run quality checks on code changes.

## Pipeline

1. **Lint**: Run golangci-lint
2. **Test**: Run go test
3. **Coverage**: Check test coverage
4. **Report**: Summarize results

## Output

```
## CI Report

### Lint: [PASS/FAIL]
[details if failed]

### Tests: [PASS/FAIL]  
[X passed, Y failed]

### Coverage: [percentage]
[files with low coverage]
```
```

## Troubleshooting

### Trigger not firing

```bash
# Check daemon is running
ayo service status

# Check trigger is enabled
ayo trigger show <name> | grep enabled

# Check trigger list
ayo trigger list
```

### Watch trigger missing events

File system watchers have limitations:
- Recursive watching may have depth limits
- Some editors use atomic saves (write to temp, rename)
- Network filesystems may not trigger events

Try with debouncing:
```bash
ayo trigger create my-watch \
  --watch ./src \
  --debounce 2s \
  --agent @watcher
```

### Cron trigger wrong timezone

Cron triggers use your system timezone. Check with:

```bash
date +%Z
```

## Next Steps

- [Memory](memory.md) - Remember context across trigger runs
- [Squads](squads.md) - Trigger squad workflows
- [Plugins](plugins.md) - Create custom trigger types

---

*You've built ambient agents! Continue to [Memory](memory.md).*
