---
id: ayo-enaj
status: open
deps: [ayo-ydub, ayo-8nn8, ayo-rdao, ayo-tha0, ayo-1ryh, ayo-ieiy, ayo-fwye, ayo-qbsu]
links: []
created: 2026-02-23T22:16:35Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-6h19
tags: [cleanup, lint, verification]
---
# Clean up dead code and verify removals

Final verification that all removals are complete and no dead code remains.

## Context

This is the **last task** in the removal sequence. All other removal tickets must be complete first.

## Tasks

1. Run `go vet ./...` - fix any issues
2. Run `golangci-lint run` - fix warnings
3. Search for orphaned references:
   ```bash
   grep -r "server\." --include="*.go" internal/
   grep -r "webhook" --include="*.go" .
   grep -r "yaml_executor" --include="*.go" .
   ```
4. Remove any remaining dead code
5. Update go.mod - remove unused dependencies
6. Run `go mod tidy`

## Verification Steps

1. `go build ./...` - passes
2. `go test ./...` - passes
3. `golangci-lint run` - no errors
4. `go mod tidy` - no changes (already tidy)

## Acceptance Criteria

- [ ] Zero linter errors
- [ ] Zero vet errors
- [ ] All tests pass
- [ ] go.mod is tidy
- [ ] No references to removed packages
- [ ] Line count reduced by ~5000 lines
