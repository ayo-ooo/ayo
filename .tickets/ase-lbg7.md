---
id: ase-lbg7
status: closed
deps: []
links: []
created: 2026-02-09T03:07:05Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-k48b
---
# Define flow YAML spec

Define and document the YAML specification for flow files.

## Background

Flows are saved orchestration patterns that @ayo creates after proving a pattern works. They're YAML files that support:
- Shell commands and agent invocations
- Conditional steps
- Input/output schemas
- Trigger definitions

## Spec

```yaml
# ~/.config/ayo/flows/daily-digest.yaml
version: 1
name: daily-digest
description: Summarize notes, translate, format as email
created_by: '@ayo'  # or 'user'
created_at: 2026-02-08T20:30:00Z

# Optional input/output schemas for validation
input:
  type: object
  properties:
    language: { type: string, default: 'english' }

output:
  type: object
  properties:
    email: { type: string }

# Execution steps
steps:
  - id: gather
    type: shell
    run: |
      find ~/notes -name '*.md' -mtime -1 -exec cat {} \;
    
  - id: summarize
    type: agent
    agent: '@summarizer'
    prompt: 'Extract key points from these notes'
    input: '{{ steps.gather.stdout }}'
    
  - id: translate
    type: agent
    agent: '@translator'
    context: 'Translate to {{ params.language }}'  # Optional context
    prompt: 'Translate the following'
    input: '{{ steps.summarize.output }}'
    when: '{{ params.language != "english" }}'  # Conditional
    
  - id: format
    type: agent
    agent: '@formatter'
    prompt: 'Format as a professional email'
    input: '{{ steps.translate.output // steps.summarize.output }}'

# Triggers (optional)
triggers:
  - id: morning-run
    type: cron
    schedule: '0 9 * * *'
    params:
      language: spanish
    runs_before_permanent: 5
      
  - id: note-change
    type: watch
    path: ~/notes
    patterns: ['*.md']
    runs_before_permanent: 10
```

## Step types

| Type | Fields | Description |
|------|--------|-------------|
| shell | run | Execute shell command, capture stdout/stderr |
| agent | agent, prompt, context?, input? | Invoke agent with prompt |

## Template syntax

- `{{ params.X }}` - input parameters
- `{{ steps.ID.stdout }}` - shell step output
- `{{ steps.ID.output }}` - agent step output
- `{{ steps.A.output // steps.B.output }}` - fallback (A or B)
- `{{ env.VAR }}` - environment variable

## Conditionals

- `when: '{{ condition }}'` - skip step if false
- Expressions: ==, !=, &&, ||, !, comparisons

## Implementation

1. Create flow schema definition (JSON Schema or Go struct)
2. Create docs/flows-spec.md with full documentation
3. Add schema validation

## Files to create

- internal/flows/spec.go (Go types)
- internal/flows/schema.json (JSON Schema for validation)
- docs/flows-spec.md (documentation)

## Acceptance Criteria

- Spec is complete and unambiguous
- JSON Schema validates example flows
- Go types match schema
- Documentation is clear

