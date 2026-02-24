# Watcher Pattern

The watcher pattern triggers agents in response to file system changes. This is ideal for reactive automation like auto-formatting, linting, or test running.

## Overview

```
File Change → Trigger → Agent Runs → Output
     ↑                        |
     └────────────────────────┘
         (continuous loop)
```

## Basic Setup

```bash
# Watch a directory for changes
ayo trigger watch ~/Code/myproject @linter \
  --prompt "Check the changed files for issues and fix them"
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--pattern` | File glob pattern (e.g., `*.go`) | `*` |
| `--debounce` | Wait time before triggering | `500ms` |
| `--singleton` | Prevent concurrent runs | `false` |
| `--recursive` | Watch subdirectories | `true` |

## Examples

### Auto-Formatter

Format code on every save:

```bash
ayo trigger watch ~/Code/myproject @formatter \
  --prompt "Format the changed files using the project's configured formatter" \
  --pattern "*.{go,ts,py}" \
  --debounce 1s \
  --singleton
```

### Test Runner

Run tests when source files change:

```bash
ayo trigger watch ~/Code/myproject/src @tester \
  --prompt "Run tests related to the changed files" \
  --pattern "*.go" \
  --debounce 2s \
  --singleton
```

### Documentation Generator

Update docs when source changes:

```bash
ayo trigger watch ~/Code/myproject/src @documenter \
  --prompt "Update documentation for any changed public APIs" \
  --pattern "*.go" \
  --debounce 5s
```

### Build Trigger

Rebuild on source changes:

```bash
ayo trigger watch ~/Code/myproject @builder \
  --prompt "Rebuild the project and report any errors" \
  --pattern "*.{go,mod,sum}" \
  --singleton
```

## Agent Prompt Best Practices

### Be Specific About Context

```
Good: "Lint the changed Go files and fix any issues found"
Bad:  "Check files"
```

### Specify Output Expectations

```
Good: "Format changed files and report what was fixed"
Bad:  "Format stuff"
```

### Include Error Handling

```
Good: "Run tests. If any fail, analyze the failure and suggest fixes."
Bad:  "Run tests"
```

## Debouncing

Debouncing prevents rapid-fire triggers during active editing:

```
Without debounce (problematic):
  save → trigger → save → trigger → save → trigger
  
With debounce (better):
  save → save → save → [wait 500ms] → trigger
```

**Recommended debounce values:**
- Formatting: 500ms - 1s
- Linting: 1s - 2s
- Tests: 2s - 5s
- Builds: 2s - 5s

## Singleton Mode

Use singleton mode to prevent overlapping runs:

```bash
ayo trigger watch ~/Code @slow-analyzer \
  --prompt "Deep code analysis" \
  --singleton
```

Without singleton:
```
trigger → agent starts (slow)
trigger → second agent starts (overlap!)
trigger → third agent starts (chaos!)
```

With singleton:
```
trigger → agent starts
trigger → queued (wait)
agent finishes → queued trigger runs
```

## Combining with Squads

Watch triggers can dispatch to squads:

```bash
# Create a code review squad
ayo squad create code-review -a @linter,@security,@docs

# Watch for changes
ayo trigger watch ~/Code/myproject "#code-review" \
  --prompt "Review the changed files for quality, security, and documentation"
```

## Troubleshooting

### Trigger Not Firing

1. Check trigger status:
   ```bash
   ayo trigger list
   ayo trigger show <id>
   ```

2. Verify watch path exists:
   ```bash
   ls -la ~/Code/myproject
   ```

3. Check daemon is running:
   ```bash
   ayo service status
   ```

### Too Many Triggers

- Increase debounce time
- Use more specific file patterns
- Enable singleton mode

### Agent Errors

Check recent sessions:
```bash
ayo session list
ayo session show <id>
```

## Best Practices

1. **Start with long debounce** (2-5s) and reduce if needed
2. **Use singleton mode** for slow operations
3. **Scope patterns tightly** to relevant file types
4. **Test manually first** before automating
5. **Monitor early** to catch issues

## See Also

- [Triggers Guide](../guides/triggers.md)
- [Scheduled Pattern](scheduled.md)
- [Monitor Pattern](monitor.md)
