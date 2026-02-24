---
id: ayo-1ryh
status: closed
deps: []
links: []
created: 2026-02-23T22:15:11Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [removal, flows]
---
# Remove YAML flow executor

Delete YAML-based flow execution. Keep flow discovery and DAG inspection.

## Context

YAML flows were an attempt at complex multi-step pipelines. In practice:
- They duplicate what shell scripts do
- They add complexity without clear benefit
- Users prefer simple agent invocations

We keep DAG inspection for visualizing flow dependencies.

## Files to Delete

- `internal/flows/yaml_executor.go` (~500 lines)
- `internal/flows/yaml_validate.go` (~200 lines)

## Files to Keep

- `internal/flows/discover.go` - Find flow definitions
- `internal/flows/parse.go` - Parse flow YAML for inspection
- `internal/flows/dag.go` - DAG construction and visualization

## Files to Modify

- `internal/daemon/server.go` - Remove flow execution RPC
- `cmd/ayo/flows.go` - Remove `run` and `execute` subcommands (see ayo-fwye)

## Verification Steps

1. Delete executor and validate files
2. Remove execution code paths
3. Run `go build ./...`
4. Run `ayo flow list` - should work
5. Run `ayo flow graph` - should work
6. Run `ayo flow run` - should return error or not exist

## Acceptance Criteria

- [ ] Executor files deleted
- [ ] `ayo flow list` works
- [ ] `ayo flow graph` works
- [ ] Flow execution no longer available
- [ ] Tests pass
