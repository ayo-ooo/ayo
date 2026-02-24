---
id: ayo-e2e
title: E2E guide provider neutrality and accuracy
status: closed
priority: high
assignee: "@ayo"
tags: [docs, gtm, polish, testing]
created: 2026-02-24
---

# E2E guide provider neutrality and accuracy

The E2E manual testing guide needs updates for provider neutrality and command accuracy.

## Problem

The E2E guide has hardcoded Anthropic references and some commands may not match current implementation.

## Provider Issues

### Hardcoded References

| Line | Issue |
|------|-------|
| 131-134 | Provider config shows only Anthropic |
| 144 | Model `claude-sonnet-4-20250514` hardcoded |
| 160 | Model `claude-sonnet-4-20250514` hardcoded |
| 311 | Model `claude-sonnet-4-20250514` hardcoded |

### Required Changes

Replace provider config example with:

```json
{
  "models": {
    "large": {
      "model": "your-large-model",
      "provider": "your-provider"
    },
    "small": {
      "model": "your-small-model", 
      "provider": "your-provider"
    }
  }
}
```

Add note:
```markdown
> **Note**: Replace `your-provider` and model names with your actual provider.
> Run `ayo setup` to configure interactively.
```

## Command Accuracy

### Verify These Commands

| Section | Command | Status |
|---------|---------|--------|
| Section 1 | `ayo sandbox service start` | ✓ Correct |
| Section 1 | `ayo doctor` | ✓ Correct |
| Section 2 | `ayo agent create @test-agent` | ✓ Correct |
| Section 3 | `ayo trigger schedule` | Verify syntax |
| Section 3 | `ayo trigger watch` | Verify syntax |

### Trigger Syntax Verification

Current in E2E guide:
```bash
ayo trigger schedule @daily-report "0 9 * * *" --prompt "Generate daily report"
```

Actual from `--help`:
```bash
ayo trigger schedule @backup "0 0 2 * * *"
```

Check if `--prompt` flag is correct.

## Other Improvements

### Prerequisites Section

Add more providers to the "one of" list:
- Anthropic, OpenAI, Google, OpenRouter, Ollama, etc.

### Test Isolation

Add guidance on testing with local Ollama to avoid API costs during testing.

## Acceptance Criteria

- [ ] No hardcoded Anthropic/Claude references
- [ ] Model names use placeholders with comments
- [ ] All commands verified against actual CLI
- [ ] Add local testing option (Ollama)
- [ ] Provider list complete

## Dependencies

- ayo-prov (provider neutrality)
- ayo-cmds (command accuracy)

## Notes

The E2E guide is critical for release validation. It must work for any provider.
