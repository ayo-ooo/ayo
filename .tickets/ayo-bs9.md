---
id: ayo-bs9
status: open
deps: [ayo-bs1, ayo-bs2, ayo-bs3, ayo-bs4, ayo-bs5, ayo-bs6, ayo-bs7]
links: []
created: 2026-03-11T18:00:00Z
type: epic
priority: 9
assignee: Alex Cabrera
tags: [build-system, documentation, user-guide]
---
# Phase 9: Documentation

Complete documentation for the build system. Users should be able to understand and use the system without contacting maintainers.

## Context

Good documentation is critical for adoption. Users need:
1. Clear getting started guide
2. Complete configuration reference
3. Skills development guide
4. Tools development guide
5. Schema validation guide
6. Examples and tutorials

## Tasks

### 9.1 Update PLAN.md with Final Architecture
- [ ] Verify architecture description
- [ ] Update diagrams
- [ ] Add implementation notes
- [ ] Document decisions made
- [ ] Link to tickets

### 9.2 Write Getting Started Guide
- [ ] Prerequisites and installation
- [ ] Create first agent with fresh
- [ ] Build agent with build command
- [ ] Run agent
- [ ] Troubleshooting common issues

### 9.3 Write Configuration Guide
- [ ] config.toml reference
- [ ] All configuration options explained
- [ ] Environment variable overrides
- [ ] Configuration examples
- [ ] Migration guide from old format

### 9.4 Write Skills Development Guide
- [ ] agentskills.io spec introduction
- [ ] Create a skill
- [ ] Skill structure and format
- [ ] Frontmatter fields explained
- [ ] Optional directories
- [ ] Skill best practices
- [ ] Skill examples

### 9.5 Write Tools Development Guide
- [ ] What are tools
- [ ] Create a custom tool
- [ ] Tool requirements
- [ ] Tool discovery
- [ ] Tool execution
- [ ] Built-in tools reference
- [ ] Tool examples

### 9.6 Write Schema Validation Guide
- [ ] JSON Schema introduction
- [ ] Create input.jsonschema
- [ ] Create output.jsonschema
- [ ] Schema to CLI mapping
- [ ] Validation behavior
- [ ] Common schema patterns
- [ ] Schema examples

### 9.7 Update README.md
- [ ] Build system overview
- [ ] Quick start
- [ ] Key features
- [ ] Links to detailed docs
- [ ] Examples
- [ ] Contributing guide

### 9.8 Update AGENTS.md
- [ ] Agent project structure
- [ ] Agent configuration
- [ ] Agent examples
- [ ] Best practices
- [ ] Troubleshooting

### 9.9 Write Build Command Documentation
- [ ] Command usage
- [ ] Options and flags
- [ ] Build output
- [ ] Cross-platform builds
- [ ] Build optimization
- [ ] Examples

### 9.10 Write Migration Guide
- [ ] Migrating from plugin system
- [ ] Migrating from old config format
- [ ] Migrating skills
- [ ] Migrating tools
- [ ] Breaking changes
- [ ] Migration checklist

### 9.11 Write Examples and Tutorials
- [ ] Simple hello agent
- [ ] Agent with skills
- [ ] Agent with tools
- [ ] Agent with structured I/O
- [ ] Complex real-world agent

### 9.12 Write Troubleshooting Guide
- [ ] Common build errors
- [ ] Common runtime errors
- [ ] Configuration issues
- [ ] Platform-specific issues
- [ ] How to get help

## Technical Details

### Documentation Structure

```
docs/
├── getting-started.md
├── configuration/
│   ├── overview.md
│   ├── reference.md
│   └── migration.md
├── skills/
│   ├── overview.md
│   ├── spec.md
│   └── development.md
├── tools/
│   ├── overview.md
│   ├── builtin.md
│   └── development.md
├── schemas/
│   ├── overview.md
│   ├── input.md
│   └── output.md
├── build-system.md
├── examples/
│   ├── hello-agent.md
│   ├── agent-with-skills.md
│   ├── agent-with-tools.md
│   └── agent-with-schemas.md
└── troubleshooting.md
```

### Agent Example Structure

```markdown
# Example: Hello Agent

## Agent Structure
...
## config.toml
...
## Build Command
...
## Run Command
...
## Output
...
```

### Documentation Standards

- Clear, concise language
- Code examples for all features
- Copy-pasteable snippets
- Screenshots where helpful
- Links between related docs
- Versioning notes

## Deliverables

- [ ] Updated PLAN.md
- [ ] Getting started guide
- [ ] Configuration guide
- [ ] Skills development guide
- [ ] Tools development guide
- [ ] Schema validation guide
- [ ] Updated README.md
- [ ] Updated AGENTS.md
- [ ] Build command documentation
- [ ] Migration guide
- [ ] Examples and tutorials
- [ ] Troubleshooting guide
- [ ] All docs reviewed and edited

## Acceptance Criteria

1. User can create first agent in 5 minutes
2. All configuration options documented
3. All features have examples
4. Common issues covered in troubleshooting
5. Docs are easy to navigate
6. Docs are kept in sync with code

## Dependencies

All previous phases (bs1-bs7) must be complete to document accurately.

## Out of Scope

- Video tutorials
- API documentation (internal APIs not public)
- Plugin documentation (plugins removed in bs1)

## Risks

- **Docs Out of Sync**: Code changes may make docs obsolete
  - **Mitigation**: Review docs with each phase, add to PR checklist
- **Missing Edge Cases**: Docs may not cover rare scenarios
  - **Mitigation**: Expand docs based on user questions

## Notes

Good documentation is as important as good code.
