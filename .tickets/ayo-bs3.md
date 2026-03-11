---
id: ayo-bs3
status: open
deps: [ayo-bs2]
links: []
created: 2026-03-11T18:00:00Z
type: epic
priority: 3
assignee: Alex Cabrera
tags: [build-system, skills, agentskills-io]
---
# Phase 3: Skills System Enhancement

Provide excellent skills support using the agentskills.io standard. Leverage the crush project implementation as reference.

## Context

We have a basic skills system (internal/skills/) but need to enhance it for the build system context:

1. Refine discovery based on crush's implementation
2. Improve parsing and validation
3. Build skills prompt for system prompt injection
4. Support skills with optional directories (scripts/, references/, assets/)

## Tasks

### 3.1 Refine Skills Discovery
- [ ] Implement concurrent discovery (like crush uses fastwalk)
- [ ] Support symlinked skill directories
- [ ] Implement priority-based deduplication
- [ ] Add discovery caching
- [ ] Follow agentskills.io spec exactly

### 3.2 Improve Skills Parsing
- [ ] Parse YAML frontmatter correctly
- [ ] Extract all optional fields (license, compatibility, metadata)
- [ ] Validate skill names match directory names
- [ ] Validate against agentskills.io spec
- [ ] Provide clear error messages

### 3.3 Skills Prompt Building
- [ ] Convert skills to XML format
- [ ] Inject into system prompt at correct location
- [ ] Include skill descriptions and locations
- [ ] Handle skills with dependencies
- [ ] Support skill deprecation warnings

### 3.4 Support Optional Skill Directories
- [ ] Detect scripts/ directory
- [ ] Detect references/ directory
- [ ] Detect assets/ directory
- [ ] Include in skill metadata
- [ ] Document optional directories

### 3.5 Skills Error Handling
- [ ] Graceful handling of invalid skills
- [ ] Warning system for non-critical issues
- [ ] Clear error messages for blocking issues
- [ ] Recovery from parsing errors

## Technical Details

### agentskills.io Spec Compliance

**Required Fields**:
- name: Alphanumeric with hyphens, matches directory
- description: 1-1024 characters

**Optional Fields**:
- license: SPDX identifier
- compatibility: Platform compatibility info
- allowed-tools: Tool requirements
- metadata: Key-value pairs

**File Structure**:
```
skills/skill-name/
├── SKILL.md           # Required
├── scripts/           # Optional
├── references/        # Optional
└── assets/            # Optional
```

### Prompt Injection Format

```xml
<available_skills>
  <skill>
    <name>skill-name</name>
    <description>Skill description</description>
    <location>/path/to/SKILL.md</location>
    <license>MIT</license>
    <has_scripts>true</has_scripts>
  </skill>
</available_skills>
```

## Deliverables

- [ ] Skills discovery matches crush implementation
- [ ] All agentskills.io spec features supported
- [ ] Skills prompt generation works correctly
- [ ] Optional skill directories detected and reported
- [ ] Comprehensive error handling
- [ ] Test coverage > 85% for skills code
- [ ] Skills documentation updated

## Acceptance Criteria

1. Skills in agent's skills/ directory are discovered
2. Skills in shared skills directories are discovered
3. Invalid skills show clear error messages
4. Skills inject correctly into system prompt
5. Optional directories detected correctly
6. Performance: <100ms for 50 skills

## Dependencies

- **ayo-bs2**: Build system core (to embed skills)
- **Crush Reference**: Use ./.read-only/crush/internal/skills/ as reference

## Out of Scope

- Skill execution (skills are passive, instructions only)
- Skill dependencies (future enhancement)
- Dynamic skill loading at runtime

## Risks

- **Spec Drift**: May diverge from agentskills.io spec
  - **Mitigation**: Reference spec directly, use crush as implementation guide

## Notes

The crush project has excellent skills implementation. Use it as the gold standard.
