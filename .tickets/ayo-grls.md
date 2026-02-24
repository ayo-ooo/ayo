---
id: ayo-grls
status: open
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-pv3a
tags: [guardrails, config]
---
# Refactor guardrails to layered model

Restructure guardrails into a layered architecture that works with sandboxing and supports experimentation.

## Current State

| Component | Type |
|-----------|------|
| LegacyGuardrails (6 rules) | Hardcoded prompt |
| Sandwich pattern | Configurable files |
| AdversarialPatterns (17 regexes) | Hardcoded |
| Trust levels | Configurable |

## Target Architecture

```
L1: INFRASTRUCTURE (Sandbox)
├─ Filesystem isolation (read-only host mount)
├─ Network controls (per-agent)
├─ Process isolation (Unix users)
└─ Resource limits

L2: PROTOCOL (Host Daemon)
├─ file_request approval flow
├─ Audit logging
├─ Trust level enforcement
└─ Adversarial input detection

L3: PROMPT (Agent System)
├─ Sandwich pattern (configurable)
├─ Per-agent guardrails (ayo.json)
└─ Squad constitution (SQUAD.md)

L4: BEHAVIORAL (Optional)
├─ Output filters
├─ Rate limiting
└─ Human-in-the-loop
```

## Configuration

Add to ayo.json schema:

```json
{
  "agent": {
    "guardrails": {
      "enabled": true,
      "level": "standard",  // "minimal", "standard", "strict"
      "sandbox": {
        "network": true,
        "filesystem": "readonly"
      },
      "prompt": {
        "use_sandwich": true,
        "custom_prefix": "path/to/prefix.txt"
      },
      "permissions": {
        "auto_approve": false,
        "allowed_paths": ["~/Projects/*"],
        "denied_paths": ["~/.ssh", "~/.aws"]
      }
    }
  }
}
```

## Levels

| Level | Description |
|-------|-------------|
| `minimal` | L1+L2 only (sandbox + protocol) |
| `standard` | L1+L2+L3 (default) |
| `strict` | All layers + additional behavioral filters |

## Implementation

### Files to Modify

- `internal/guardrails/` - Restructure into layers
- `internal/agent/agent.go` - Load guardrails config
- Schema files - Add guardrails section

### New Files

- `internal/guardrails/layers.go` - Layer abstraction
- `internal/guardrails/sandbox.go` - L1 helpers
- `internal/guardrails/protocol.go` - L2 helpers

## Experimentation Support

- `guardrails.enabled: false` disables L3/L4
- L1/L2 always active (infrastructure safety)
- Custom sandwich files for prompt experimentation
- Per-agent override of any setting

## Testing

- Test each layer independently
- Test level configurations
- Test custom sandwich files
- Test per-agent overrides
