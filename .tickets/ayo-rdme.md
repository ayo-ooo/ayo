---
id: ayo-rdme
title: README first impression polish
status: open
priority: critical
assignee: "@ayo"
tags: [docs, gtm, polish, first-impression]
created: 2026-02-24
---

# README first impression polish

The README is the first thing users see. It must be perfect.

## Current Issues

### 1. Provider Bias (Line 23)
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```
Assumes Anthropic. Should use `ayo setup` as primary method.

### 2. Incorrect Flag (Line 57)
```bash
ayo agents create @support --description "..."
```
Flag is `-d` not `--description`. Also uses plural `agents`.

### 3. Trigger Command (Lines 144-151)
```bash
ayo triggers schedule @standup "0 9 * * 1-5"
ayo triggers watch ~/Projects @reviewer
```
Should be singular `trigger`.

### 4. Missing Features

Not mentioned in README but implemented:
- Sessions (`-c`, `-s` flags for continuation)
- Flows (multi-step workflows)
- Skills system
- Plugins
- Audit logs
- Planner system
- Share command

### 5. Missing Common Flags Section

Add section showing:
```bash
-y, --no-jodas    Auto-approve file modifications
-q, --quiet       Suppress non-essential output
--json            Output in JSON format
-c, --continue    Continue most recent session
-s, --session     Continue specific session
```

### 6. Unix Pipe Integration Not Shown

The CLI supports Unix pipes (`cat file | ayo "summarize this"`) but README doesn't show it.

## Required Changes

### Quick Start Section

**Before:**
```bash
# Configure
export ANTHROPIC_API_KEY="sk-ant-..."
ayo setup
```

**After:**
```bash
# Configure (interactive wizard, supports 10+ providers)
ayo setup

# Start the sandbox service
ayo sandbox service start

# Start chatting
ayo
```

### Agents Section

**Before:**
```bash
ayo agents create @support --description "Customer support agent"
```

**After:**
```bash
ayo agent create @support -d "Customer support agent"
```

### Triggers Section

**Before:**
```bash
ayo triggers schedule @standup "0 9 * * 1-5"
```

**After:**
```bash
ayo trigger schedule @standup "0 9 * * 1-5" --prompt "Run daily standup"
```

### Add New Sections

1. **Common Flags** section after Quick Start
2. **Unix Pipe Integration** example
3. **Sessions** section showing `-c` and `-s`
4. **More Features** section listing: flows, skills, plugins, audit

## Acceptance Criteria

- [ ] No provider-specific configuration shown (use `ayo setup`)
- [ ] All command examples use correct flags and singular forms
- [ ] Common flags documented
- [ ] Unix pipe integration shown
- [ ] Session continuation documented
- [ ] All major features at least mentioned

## Dependencies

- ayo-prov (provider neutrality)
- ayo-cmds (command accuracy)

## Notes

README should be scannable in 30 seconds. Key selling points:
1. Isolated sandbox execution
2. Persistent memory
3. Multi-agent coordination (squads)
4. Ambient automation (triggers)
