# Tutorial: Daily Digest Flow

This tutorial walks you through building an automated daily digest flow that summarizes your activity and sends it via Matrix.

## What You'll Build

A YAML flow that:
1. Collects git commits from the past day
2. Gathers recently modified files
3. Summarizes the activity using an AI agent
4. Posts the digest to a Matrix room

## Prerequisites

- ayo installed and configured
- Sandbox service running (`ayo sandbox service start`)
- At least one API key configured

## Step 1: Create the Flow File

Create a new YAML flow:

```bash
mkdir -p ~/.config/ayo/flows
```

Create `~/.config/ayo/flows/daily-digest.yaml`:

```yaml
version: 1
name: daily-digest
description: Generate and post a daily activity digest

params:
  days:
    type: integer
    default: 1
  room:
    type: string
    default: "daily-updates"

steps:
  - id: git-commits
    type: shell
    run: |
      git log --oneline --since="{{ params.days }} days ago" 2>/dev/null || echo "No git repository"

  - id: recent-files
    type: shell
    run: |
      find . -name "*.go" -o -name "*.md" -o -name "*.yaml" | head -20 | xargs ls -lt 2>/dev/null | head -10

  - id: generate-summary
    type: agent
    agent: "@ayo"
    prompt: |
      Create a brief daily digest from this activity:
      
      ## Recent Commits
      {{ steps.git-commits.stdout }}
      
      ## Recently Modified Files
      {{ steps.recent-files.stdout }}
      
      Format as a concise summary with:
      - Key accomplishments
      - Files touched
      - Suggested next steps
    depends_on: [git-commits, recent-files]

  - id: post-digest
    type: shell
    run: |
      ayo matrix send {{ params.room }} "$(cat << 'EOF'
      📊 Daily Digest
      
      {{ steps.generate-summary.output }}
      EOF
      )"
    depends_on: [generate-summary]
```

## Step 2: Validate the Flow

Check for syntax errors:

```bash
ayo flows validate ~/.config/ayo/flows/daily-digest.yaml
```

Expected output:
```
✓ Flow 'daily-digest' is valid
  4 steps defined
  Dependencies form valid DAG
```

## Step 3: Create the Matrix Room

Set up the notification room:

```bash
# Create the room
ayo matrix create daily-updates

# Verify it exists
ayo matrix rooms
```

## Step 4: Test the Flow Manually

Run the flow with default parameters:

```bash
cd /path/to/your/project
ayo flows run daily-digest
```

Run with custom parameters:

```bash
ayo flows run daily-digest --param days=7 --param room=weekly-updates
```

## Step 5: Add Automation

### Option A: Trigger in the Flow

Add a trigger section to the YAML:

```yaml
version: 1
name: daily-digest
description: Generate and post a daily activity digest

triggers:
  - id: morning-digest
    type: cron
    schedule: "0 9 * * *"  # 9am daily
    params:
      days: 1
      room: daily-updates

  - id: weekly-digest
    type: cron
    schedule: "0 9 * * MON"  # Monday 9am
    params:
      days: 7
      room: weekly-updates

params:
  # ... rest of flow
```

### Option B: External Trigger

Create a trigger from the CLI:

```bash
ayo trigger schedule @ayo "every day at 9am" \
  --prompt "Run: ayo flows run daily-digest"
```

## Step 6: View History

Check past runs:

```bash
# List recent runs
ayo flows history --flow daily-digest

# View statistics
ayo flows stats daily-digest
```

## Customization Ideas

### Add Code Quality Metrics

```yaml
- id: test-coverage
  type: shell
  run: go test -cover ./... 2>&1 | grep "coverage:"
```

### Include TODO Items

```yaml
- id: todos
  type: shell
  run: grep -r "TODO:" --include="*.go" . | head -10
```

### Email Instead of Matrix

```yaml
- id: send-email
  type: shell
  run: |
    echo "{{ steps.generate-summary.output }}" | \
    mail -s "Daily Digest" you@example.com
  depends_on: [generate-summary]
```

## Troubleshooting

### Flow not finding git repository

Make sure you run the flow from a git repository:
```bash
cd /path/to/your/repo
ayo flows run daily-digest
```

### Matrix room not found

Verify the room exists and you have access:
```bash
ayo matrix rooms
ayo matrix create daily-updates
```

### Agent step timing out

Increase the timeout in the step:
```yaml
- id: generate-summary
  type: agent
  agent: "@ayo"
  timeout: 5m  # Increase from default
  # ...
```

## Complete Flow

Here's the final, production-ready flow:

```yaml
version: 1
name: daily-digest
description: Generate and post a daily activity digest
created_at: 2026-02-08T12:00:00Z

triggers:
  - id: morning
    type: cron
    schedule: "0 9 * * *"
    params:
      days: 1
      room: daily-updates

params:
  days:
    type: integer
    default: 1
  room:
    type: string
    default: "daily-updates"

steps:
  - id: git-commits
    type: shell
    run: git log --oneline --since="{{ params.days }} days ago" 2>/dev/null || echo "No git repository"
    timeout: 30s

  - id: recent-files
    type: shell
    run: find . -name "*.go" -o -name "*.md" | xargs ls -lt 2>/dev/null | head -10
    timeout: 30s

  - id: test-status
    type: shell
    run: go test ./... -short -json 2>&1 | tail -5 || echo "No tests"
    timeout: 2m
    continue_on_error: true

  - id: generate-summary
    type: agent
    agent: "@ayo"
    prompt: |
      Create a brief daily digest:
      
      Commits: {{ steps.git-commits.stdout }}
      Files: {{ steps.recent-files.stdout }}
      Tests: {{ steps.test-status.stdout }}
      
      Format: accomplishments, files touched, next steps
    depends_on: [git-commits, recent-files, test-status]
    timeout: 3m

  - id: post-digest
    type: shell
    run: |
      ayo matrix send {{ params.room }} "📊 Daily Digest

      {{ steps.generate-summary.output }}"
    depends_on: [generate-summary]
```

## Next Steps

- [TUTORIAL-code-review.md](TUTORIAL-code-review.md) - Build a code review pipeline
- [TUTORIAL-file-watcher.md](TUTORIAL-file-watcher.md) - Set up file watch triggers
