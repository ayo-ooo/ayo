# Tutorial: File Watcher Triggers

This tutorial shows you how to set up file watch triggers that automatically run agents or flows when files change.

## What You'll Build

A file watching system that:
1. Monitors a directory for changes
2. Triggers agent actions on file modifications
3. Sends notifications via Matrix
4. Integrates with flows for complex workflows

## Prerequisites

- ayo installed and configured
- Sandbox service running (`ayo sandbox service start`)
- Mount access to the directory you want to watch

## Step 1: Understanding Watch Triggers

Watch triggers monitor the filesystem and fire when specified events occur:

- **create** - New file created
- **modify** - File content changed
- **delete** - File removed
- **rename** - File renamed

By default, triggers fire on `create` and `modify`.

## Step 2: Create a Simple Watch Trigger

### Via CLI

```bash
# Watch for any changes in ./src
ayo trigger watch ./src @ayo

# Watch with specific patterns
ayo trigger watch ./src @ayo "*.go" "*.mod"

# Watch recursively
ayo trigger watch ./src @ayo "*.go" --recursive

# Watch with custom prompt
ayo trigger watch ./src @ayo "*.go" \
  --prompt "A Go file changed. Check for syntax errors and run tests."
```

### Verify the Trigger

```bash
# List triggers
ayo trigger list

# Show details
ayo trigger show <trigger-id>
```

## Step 3: Watch Trigger in YAML Flows

Embed triggers directly in flow files:

```yaml
# ~/.config/ayo/flows/auto-build.yaml
version: 1
name: auto-build
description: Automatically build and test on code changes

triggers:
  - id: code-change
    type: watch
    path: ./src
    patterns: ["*.go"]
    recursive: true
    events: [create, modify]
    debounce: 5s  # Wait 5s of quiet before triggering

steps:
  - id: build
    type: shell
    run: go build ./...

  - id: test
    type: shell
    run: go test ./... -short
    depends_on: [build]
    continue_on_error: true

  - id: notify
    type: shell
    run: |
      if [ "{{ steps.test.exit_code }}" = "0" ]; then
        ayo matrix send dev-updates "✅ Build and tests passed"
      else
        ayo matrix send dev-updates "❌ Tests failed"
      fi
    depends_on: [test]
```

## Step 4: Pattern Matching

### Glob Patterns

Watch triggers support glob patterns:

| Pattern | Matches |
|---------|---------|
| `*.go` | All Go files |
| `*.{go,mod}` | Go files and go.mod |
| `*_test.go` | Test files only |
| `!*_test.go` | Exclude test files |
| `**/*.go` | Go files in all subdirectories |

### Examples

```bash
# Watch only test files
ayo trigger watch ./src @ayo "*_test.go"

# Watch multiple patterns
ayo trigger watch ./src @ayo "*.go" "*.yaml" "*.json"

# Watch docs
ayo trigger watch ./docs @ayo "*.md" --recursive
```

## Step 5: Debouncing

Debouncing prevents rapid-fire triggers when multiple files change quickly (like during a `git checkout`).

### In YAML

```yaml
triggers:
  - id: code-change
    type: watch
    path: ./src
    patterns: ["*.go"]
    debounce: 10s  # Wait 10 seconds of quiet
```

### Via CLI

```bash
# Debounce is automatic (default 2s)
# Changes within 2s are batched
ayo trigger watch ./src @ayo "*.go"
```

## Step 6: Event Filtering

Control which events trigger actions:

```yaml
triggers:
  - id: new-files-only
    type: watch
    path: ./uploads
    events: [create]  # Only new files

  - id: modifications-only
    type: watch
    path: ./config
    events: [modify]  # Only changes to existing files

  - id: deletions
    type: watch
    path: ./tmp
    events: [delete]  # Cleanup trigger
```

## Step 7: Practical Examples

### Auto-Format on Save

```yaml
version: 1
name: auto-format
description: Format code on save

triggers:
  - type: watch
    path: ./src
    patterns: ["*.go"]
    events: [modify]

steps:
  - id: format
    type: shell
    run: gofmt -w ./src/

  - id: lint
    type: shell
    run: golangci-lint run ./src/...
    continue_on_error: true
```

### Documentation Watcher

```yaml
version: 1
name: docs-watcher
description: Process documentation changes

triggers:
  - type: watch
    path: ./docs
    patterns: ["*.md"]
    recursive: true

steps:
  - id: check-links
    type: shell
    run: |
      # Find broken links in markdown
      grep -r '\[.*\](.*\.md)' docs/ | while read line; do
        file=$(echo "$line" | grep -oP '\(.*?\.md\)')
        if [ ! -f "docs/${file:1:-1}" ]; then
          echo "Broken link: $line"
        fi
      done

  - id: generate-toc
    type: shell
    run: |
      # Generate table of contents
      find docs/ -name "*.md" | sort | while read f; do
        title=$(head -1 "$f" | sed 's/# //')
        echo "- [$title]($f)"
      done > docs/TOC.md
```

### Config Change Notifier

```yaml
version: 1
name: config-watcher
description: Alert on configuration changes

triggers:
  - type: watch
    path: ./config
    patterns: ["*.yaml", "*.json", "*.env"]
    events: [create, modify, delete]

steps:
  - id: diff
    type: shell
    run: git diff --no-color ./config/ 2>/dev/null || echo "No git"

  - id: notify
    type: agent
    agent: "@ayo"
    prompt: |
      A configuration file was changed:
      
      {{ steps.diff.stdout }}
      
      Summarize what changed and flag any potential issues.
    depends_on: [diff]

  - id: alert
    type: shell
    run: |
      ayo matrix send config-alerts "⚙️ Config changed:
      
      {{ steps.notify.output }}"
    depends_on: [notify]
```

## Step 8: Managing Watch Triggers

### List All Triggers

```bash
ayo trigger list
```

Output:
```
ID        TYPE   AGENT   PATH         PATTERNS    STATUS
abc123    watch  @ayo    ./src        *.go        enabled
def456    watch  @ayo    ./docs       *.md        enabled
ghi789    cron   @ayo    -            -           enabled
```

### Disable/Enable

```bash
# Temporarily disable
ayo trigger disable abc123

# Re-enable
ayo trigger enable abc123
```

### Test Manually

```bash
# Simulate trigger firing
ayo trigger test abc123
```

### Remove

```bash
ayo trigger rm abc123
```

## Step 9: Troubleshooting

### Trigger Not Firing

1. **Check service is running:**
   ```bash
   ayo sandbox service status
   ```

2. **Verify trigger is enabled:**
   ```bash
   ayo trigger show <id>
   ```

3. **Check path is accessible:**
   ```bash
   ayo mount list
   ls -la /path/to/watched/dir
   ```

4. **Check daemon logs:**
   ```bash
   ./debug/daemon-status.sh --logs
   ```

### Too Many Triggers

If changes cause trigger storms:

1. Increase debounce time
2. Use more specific patterns
3. Filter events more narrowly
4. Exclude directories (node_modules, .git, etc.)

### Performance Issues

For large directories:

```yaml
triggers:
  - type: watch
    path: ./src
    patterns: ["*.go"]
    recursive: true
    # Exclude noisy directories
    exclude: ["vendor/", "testdata/", "*.generated.go"]
```

## Complete Example: Development Workflow

```yaml
version: 1
name: dev-workflow
description: Complete development workflow automation

triggers:
  # Watch source code
  - id: code
    type: watch
    path: ./src
    patterns: ["*.go"]
    recursive: true
    debounce: 3s

  # Watch tests
  - id: tests
    type: watch
    path: ./src
    patterns: ["*_test.go"]
    debounce: 3s

  # Watch config
  - id: config
    type: watch
    path: ./config
    patterns: ["*.yaml"]

steps:
  # Build
  - id: build
    type: shell
    run: go build -o bin/app ./cmd/...
    timeout: 2m

  # Run tests for changed files
  - id: test
    type: shell
    run: go test ./... -short -count=1
    depends_on: [build]
    timeout: 5m
    continue_on_error: true

  # Lint
  - id: lint
    type: shell
    run: golangci-lint run --new-from-rev=HEAD~1
    depends_on: [build]
    continue_on_error: true

  # Report
  - id: report
    type: shell
    run: |
      status="✅ All checks passed"
      if [ "{{ steps.test.exit_code }}" != "0" ]; then
        status="❌ Tests failed"
      elif [ "{{ steps.lint.exit_code }}" != "0" ]; then
        status="⚠️ Lint warnings"
      fi
      ayo matrix send dev-channel "$status"
    depends_on: [test, lint]
```

## Next Steps

- [TUTORIAL-daily-digest.md](TUTORIAL-daily-digest.md) - Build a daily digest flow
- [TUTORIAL-code-review.md](TUTORIAL-code-review.md) - Create a code review pipeline
