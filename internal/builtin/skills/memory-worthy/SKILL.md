---
name: memory-worthy
description: Identifying information worth remembering. Use when extracting memorable content from conversations for long-term storage.
compatibility: Requires memory tool and small model for extraction
metadata:
  author: ayo
  version: "1.0"
---

# Memory-Worthy Extraction Skill

Identify and extract information from conversations that should be stored as long-term memories.

## When to Use

This skill activates during conversation analysis to identify:
- Explicit preferences ("I prefer...", "I always...")
- Corrections to agent behavior ("No, I meant...", "That's wrong...")
- Facts about user or project ("I work at...", "This project uses...")
- Behavioral patterns observed over time

## What Makes Information Memory-Worthy

### High Value (Store Immediately)

| Signal | Example | Category |
|--------|---------|----------|
| Explicit preference | "I prefer tabs over spaces" | preference |
| Direct correction | "No, use npm not yarn" | correction |
| Stated fact | "Our API is at api.example.com" | fact |
| Requested memory | "Remember that I like..." | preference |

### Medium Value (Consider Storing)

| Signal | Example | Category |
|--------|---------|----------|
| Repeated behavior | Always runs tests first | pattern |
| Implicit preference | Consistently uses short variable names | pattern |
| Project convention | Tests are in __tests__ folder | fact |

### Low Value (Do Not Store)

- Session-specific context (current task details)
- Widely known best practices
- Temporary information
- Sensitive data (credentials, secrets)
- Information that changes frequently

## Extraction Guidelines

### Be Concise

Store the essence, not full sentences:

```
Good: "Prefers table-driven tests in Go"
Bad: "The user mentioned they really like to use table-driven tests when writing Go code"
```

### Be Specific

Include actionable detail:

```
Good: "Uses Go 1.22 with generics"
Bad: "Uses Go"
```

### Be Accurate

Only store what was explicitly stated or clearly implied:

```
Good: "Prefers concise code comments"
Bad: "Doesn't like documentation" (inference, may be wrong)
```

### Avoid Duplicates

Before storing, consider if similar information already exists. Update rather than create if possible.

## Category Selection

| Category | Use When |
|----------|----------|
| `preference` | User expresses how they like things done |
| `fact` | Objective, verifiable information |
| `correction` | User corrects previous agent behavior |
| `pattern` | Observed repeated behavior (not explicitly stated) |

## Source Attribution

When extracting memories, include source context:
- Session ID where information was learned
- Approximate time/date
- Whether explicit or inferred

## Scope Considerations

| Scope | Store When |
|-------|------------|
| Global | Applies to all contexts (personal preferences) |
| Agent | Specific to how this agent should behave |
| Path | Project-specific conventions |

## Conflict Prevention

Before storing, check for conflicts:
- Does this contradict existing memory?
- Is this more recent/authoritative than existing info?
- Should old memory be updated or replaced?

## Output Format

When identifying memory-worthy content, structure as:

```json
{
  "content": "Concise statement of what to remember",
  "category": "preference|fact|correction|pattern",
  "confidence": "high|medium",
  "source": "Relevant quote or context"
}
```

## Examples

### Input
> "No, don't use the default Go formatter - we use gofumpt in this project"

### Output
```json
{
  "content": "Project uses gofumpt instead of default gofmt",
  "category": "correction",
  "confidence": "high",
  "source": "User correction about formatting"
}
```

### Input
> "I've been working on this API for the past year"

### Output
```json
{
  "content": "User has year+ experience with current API",
  "category": "fact",
  "confidence": "medium",
  "source": "User statement about experience"
}
```
