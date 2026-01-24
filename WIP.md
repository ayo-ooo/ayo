# WIP: Memory Deduplication

## Status: Complete, Ready for Testing

## Problem
When a user says "remember that i prefer typescript over javascript", multiple triggers fire:
- `TriggerExplicit` (matches "remember that")
- `TriggerPreference` (matches "i prefer")

This created duplicate memories with the same content but different categories.

## Solution Implemented
Added semantic deduplication in `FormationService.processFormation()`:

1. Before creating a memory, search for similar existing memories
2. If similarity ≥ 0.95 (exact duplicate) → skip creation, return existing memory
3. If similarity ≥ 0.85 (similar) → supersede the old memory
4. Otherwise → create new memory

## Files Changed

### `internal/memory/formation.go`
- Added constants `ExactDuplicateThreshold` (0.95) and `SupersedeThreshold` (0.85)
- Modified `processFormation()` to check for duplicates before creating
- Extended `FormationResult` struct with:
  - `Deduplicated bool` - true if exact duplicate found
  - `Superseded bool` - true if similar memory was superseded
  - `SupersededID string` - ID of superseded memory

### `internal/memory/formation_test.go`
- Added `TestFormationDeduplication_ExactDuplicate`
- Added `TestFormationDeduplication_SimilarSupersedes`
- Added `TestFormationDeduplication_DifferentContent`
- Added helper `deterministicEmbedder` for consistent test embeddings

## Tests
All tests pass:
```
go test ./... ✓
```

## Next Steps
1. Manual testing with real usage to verify thresholds are appropriate
2. Consider adding UI feedback when a memory is deduplicated vs created fresh
3. May want to tune thresholds (0.85/0.95) based on real-world usage
