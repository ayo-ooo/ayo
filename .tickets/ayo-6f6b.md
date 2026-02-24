---
id: ayo-6f6b
status: open
deps: [ayo-enaj]
links: []
created: 2026-02-23T22:16:17Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [docs, cleanup]
---
# Remove outdated documentation

Delete or archive documentation for removed features. Consolidate scattered reference docs.

## Context

After Phase 1 code removals, documentation references features that no longer exist. This ticket cleans up the docs.

## Files to Delete

### Docs for Removed Features

- `docs/flows-spec.md` - Flows removed
- `docs/flows.md` - Flows removed  
- `docs/plugins.md` - Simplify significantly or remove
- `docs/reference/` - Consolidate into cli-reference.md

## Files to Archive (Move to docs/archive/)

Keep for historical reference but move out of main docs:
- Any legacy architecture docs
- Old design docs that are now outdated

## Files to Consolidate

### CLI Reference

Consolidate scattered command docs into single file:
- `docs/cli-reference.md` - All CLI commands

### Configuration

Consolidate config docs:
- `docs/configuration.md` - All config options

## TUTORIAL.md Rewrite

The current TUTORIAL.md is too long and references removed features. Rewrite to:
- ~500 lines (down from ~2000)
- Focus on key workflows
- Remove flows/plugins complexity
- Add trigger examples

## New Structure

```
docs/
├── README.md              # Docs index
├── getting-started.md     # Quick start
├── agents.md              # Agent guide
├── squads.md              # Squad guide
├── triggers.md            # Trigger guide
├── tools.md               # Tool reference
├── configuration.md       # Config reference
├── cli-reference.md       # CLI reference
├── patterns/              # Common patterns
│   ├── watcher.md
│   ├── scheduled.md
│   └── ...
└── archive/               # Old docs
    └── ...
```

## Steps

1. Delete removed feature docs
2. Move deprecated docs to archive/
3. Consolidate reference docs
4. Update docs/README.md index
5. Verify all links work

## Acceptance Criteria

- [ ] No docs reference removed features
- [ ] Reference docs consolidated
- [ ] Archive directory created for old docs
- [ ] docs/README.md index updated
- [ ] All internal doc links work
- [ ] Clear, logical docs structure

## Testing

- Verify no broken links
- Verify no references to flows, removed plugins
- Check docs index is accurate
