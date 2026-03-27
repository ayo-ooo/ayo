# CLI Integration Flow

This document describes the execution flow for generated agent CLIs, including how interactive forms are triggered and how errors are handled.

## Execution Flow

```
Agent invoked
    │
    ▼
Parse CLI args against schema
    │
    ▼
All required args present? ──Yes──► Run agent
    │
   No
    │
    ▼
TTY available? ──No──► Fail: missing required args (styled error)
    │
   Yes
    │
    ▼
Agent allows interactive? ──No──► Fail: missing required args (styled error)
    │
   Yes
    │
    ▼
Generate inline form from schema
    │
    ▼
Pre-populate with CLI-provided values
    │
    ▼
Show form (huh), validate inline
    │
    ▼
Submit ──► Run agent with merged inputs
```

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| No TTY (piped input) | Fail with styled error listing required args |
| `--non-interactive` flag | Skip form, fail if missing required args |
| Partial CLI flags | Show form with pre-populated values |
| All required flags provided | Skip form entirely, run directly |
| Agent has `interactive: false` | Fail immediately if missing args |
| User cancels form (Ctrl+C) | Exit code 1, no error message |
| No `input.jsonschema` | Run agent without form (passthrough) |

## Error Message Format

When required arguments are missing and no form can be shown:

```
Missing required arguments:

  --prompt    What would you like the agent to do?
  --scope     Where should the agent operate?

Run with --help for usage information.
```

Error messages are styled with `lipgloss` for consistent visual appearance:
- Title in bold
- Flag names in a distinct color
- Descriptions in a dimmed color

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (missing args, user cancelled, etc.) |
| 2 | Invalid arguments (parsing error) |

## Flag Behavior

### `--non-interactive`

Bypasses interactive form even when available:

```bash
# If prompt is required and not provided:
my-agent --non-interactive
# Output: Error: missing required argument --prompt

# With all required args:
my-agent --prompt "Hello" --non-interactive
# Runs immediately without form
```

### `--help`

Shows usage information including all available flags:

```bash
my-agent --help
```

### JSON Input

Agents can also accept JSON input directly:

```bash
# From argument
my-agent '{"prompt": "Hello", "scope": "project"}'

# From stdin
echo '{"prompt": "Hello"}' | my-agent -
```

JSON input is merged with CLI flags (flags take precedence).

## Pre-population Example

Given an agent with required `prompt` and optional `scope`:

```bash
# No args: form shown for both fields
my-agent

# Partial: form shown for prompt only, scope pre-filled
my-agent --scope file

# Full: no form, runs directly
my-agent --prompt "Review this code" --scope file
```

## Performance Requirements

- Form should render in under 100ms for typical schemas (<20 fields)
- Validation should be instant (no network calls)
- Pre-population should not delay form display
