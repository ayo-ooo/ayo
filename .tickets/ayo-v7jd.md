---
id: ayo-v7jd
status: open
deps: [ayo-spy5]
links: []
created: 2026-02-23T22:16:25Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [docs, examples]
---
# Add examples and recipes

Create an examples/ directory with ready-to-use configurations. Include agent definitions, squad setups, and trigger configurations that users can copy and modify.

## Context

Users learn best from examples. This ticket creates a curated set of example configurations demonstrating common use cases.

## Directory Structure

```
examples/
├── README.md               # Overview and index
├── agents/
│   ├── code-reviewer/
│   │   ├── ayo.json
│   │   ├── AGENT.md
│   │   └── skills/
│   │       └── review-guidelines.md
│   ├── project-analyzer/
│   │   ├── ayo.json
│   │   └── AGENT.md
│   └── daily-summarizer/
│       ├── ayo.json
│       └── AGENT.md
├── squads/
│   ├── dev-team/
│   │   ├── ayo.json
│   │   └── SQUAD.md
│   └── research-team/
│       ├── ayo.json
│       └── SQUAD.md
└── triggers/
    ├── auto-lint.yaml
    ├── morning-report.yaml
    └── file-watcher.yaml
```

## Example Details

### Code Reviewer Agent

```json
// examples/agents/code-reviewer/ayo.json
{
  "version": "1",
  "agent": {
    "description": "Reviews code changes for quality and style",
    "model": "claude-sonnet-4-5-20250929",
    "tools": ["bash", "view", "grep"],
    "skills": ["review-guidelines"]
  }
}
```

### Project Analyzer Agent

Purpose: Analyze a codebase and generate reports
- Identifies tech stack
- Counts lines of code
- Finds potential issues

### Daily Summarizer Agent

Purpose: Generate daily summaries
- Git activity
- Calendar events
- Open tasks

### Dev Team Squad

Purpose: Coordinated development team
- @architect (lead)
- @frontend
- @backend
- @qa

### Research Team Squad

Purpose: Research and documentation
- @researcher (lead)
- @writer
- @editor

### Auto-Lint Trigger

Purpose: Lint files on save
- File watch trigger
- Runs on *.go, *.ts changes
- Singleton mode

### Morning Report Trigger

Purpose: Daily morning briefing
- Daily trigger at 9am
- Summarizes overnight activity

### File Watcher Trigger

Purpose: React to file changes
- Watch specified directory
- Debouncing configured

## README Content

```markdown
# ayo Examples

Ready-to-use configurations for common use cases.

## Agents

| Example | Use Case |
|---------|----------|
| code-reviewer | Review PRs and code changes |
| project-analyzer | Analyze codebase structure |
| daily-summarizer | Generate daily summaries |

## Squads

| Example | Use Case |
|---------|----------|
| dev-team | Coordinated development |
| research-team | Research and documentation |

## Triggers

| Example | Use Case |
|---------|----------|
| auto-lint | Lint on file save |
| morning-report | Daily briefing |
| file-watcher | React to file changes |

## Using Examples

Copy to your config directory:
\`\`\`bash
cp -r examples/agents/code-reviewer ~/.config/ayo/agents/
\`\`\`

Or use as templates:
\`\`\`bash
ayo agent create --from examples/agents/code-reviewer my-reviewer
\`\`\`
```

## Acceptance Criteria

- [ ] All examples are complete and tested
- [ ] README provides clear overview
- [ ] Each example has clear purpose
- [ ] Configuration files are valid JSON/YAML
- [ ] AGENT.md and SQUAD.md files included where needed
- [ ] Examples demonstrate key features
- [ ] Copy-paste usage instructions work

## Testing

- Test each example can be loaded by ayo
- Verify configuration syntax is valid
- Test copy-paste instructions work
