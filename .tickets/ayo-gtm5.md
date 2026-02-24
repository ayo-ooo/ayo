---
id: ayo-gtm5
status: open
deps: []
links: []
created: 2026-02-24T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-gtm1
tags: [gtm, refactoring, prompts]
---
# Task: Externalize Remaining Hardcoded Prompts

## Summary

PLAN.md requires zero hardcoded prompt strings in the codebase. Currently, several prompts are still hardcoded as Go constants.

## Current Hardcoded Prompts

### 1. `internal/guardrails/defaults.go`

| Constant | Description | Target File |
|----------|-------------|-------------|
| `defaultPrefix` | Security sandwich prefix | `prompts/sandwich/prefix.md` |
| `defaultSuffix` | Security sandwich suffix | `prompts/sandwich/suffix.md` |
| `defaultLegacyGuardrails` | Legacy 6-rule guardrails | `prompts/guardrails/legacy.md` |

### 2. `internal/plugins/patterns.go`

| Item | Description | Target File |
|------|-------------|-------------|
| `AdversarialPatterns` | Regex patterns for input filtering | `prompts/adversarial/patterns.txt` |

## Implementation Plan

### Step 1: Create Prompt Files

```
~/.local/share/ayo/prompts/
├── sandwich/
│   ├── prefix.md      # Move defaultPrefix
│   └── suffix.md      # Move defaultSuffix
├── guardrails/
│   └── legacy.md      # Move defaultLegacyGuardrails
└── adversarial/
    └── patterns.txt   # Move AdversarialPatterns (one per line)
```

### Step 2: Update Code

1. Modify `LegacyGuardrails()` to load from file with embedded fallback
2. Modify `GetPrefix()` / `GetSuffix()` to load from files
3. Modify adversarial pattern loading to use external file
4. Keep embedded defaults as fallbacks for fresh installs

### Step 3: Update Installer

Ensure `ayo setup` copies default prompts to user directory.

## Implementation Steps

1. [ ] Create `prompts/sandwich/prefix.md` in embedded defaults
2. [ ] Create `prompts/sandwich/suffix.md` in embedded defaults
3. [ ] Create `prompts/guardrails/legacy.md` in embedded defaults
4. [ ] Create `prompts/adversarial/patterns.txt` in embedded defaults
5. [ ] Update `LegacyGuardrails()` to load external → fallback to embedded
6. [ ] Update sandwich loading to use prompts loader
7. [ ] Update adversarial patterns to load from file
8. [ ] Update `ayo setup` to install prompts
9. [ ] Test with and without installed prompts
10. [ ] Document prompt customization in docs

## Code Changes

```go
// Before
const defaultLegacyGuardrails = `<guardrails>...`

func LegacyGuardrails() string {
    return prompts.Default().LoadOrDefault(prompts.PathGuardrailsDefault, defaultLegacyGuardrails)
}

// After - embeds the default but prefers user file
//go:embed defaults/guardrails/legacy.md
var embeddedLegacyGuardrails string

func LegacyGuardrails() string {
    return prompts.Default().LoadOrDefault(prompts.PathGuardrailsLegacy, embeddedLegacyGuardrails)
}
```

## Acceptance Criteria

- [ ] No hardcoded prompt strings in Go source (only `//go:embed`)
- [ ] `ayo setup` installs all default prompts
- [ ] Prompts work correctly without installed files (fallback)
- [ ] Users can customize prompts by editing files
- [ ] Documentation explains prompt customization
