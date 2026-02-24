---
id: ayo-gtm1
status: open
deps: []
links: []
created: 2026-02-24T12:00:00Z
type: epic
priority: 0
assignee: Alex Cabrera
tags: [gtm, polish, refactoring]
---
# Epic: GTM Final Polish

## Summary

This epic tracks all remaining non-testing work required to achieve GTM readiness. It focuses on refactoring, code quality, prompt externalization, and human-facing polish.

## Analysis

### From PLAN.md Success Criteria

| Criterion | Current Status | Work Needed |
|-----------|---------------|-------------|
| Documentation | ✓ Complete | None |
| No dead code | ⚠ Partial | File splitting, interface consolidation |
| Mental model clarity | ⚠ Partial | Prompt externalization |
| Code quality | ⚠ Partial | gopls modernization |

### Scope (Non-Testing Work)

1. **Large File Splitting** - agent.go (1320 lines), server.go (1170 lines)
2. **Interface Consolidation** - AgentInvoker defined in 9+ locations
3. **Prompt Externalization** - Move hardcoded prompts to files
4. **gopls Modernization** - Fix 172+ hints
5. **Cron Migration Cleanup** - Remove robfig/cron if gocron is complete
6. **Dead Code Removal** - chat.go and other remnants

## Children

| Ticket | Description | Priority |
|--------|-------------|----------|
| ayo-gtm2 | Split internal/agent/agent.go | High |
| ayo-gtm3 | Split internal/daemon/server.go | High |
| ayo-gtm4 | Consolidate AgentInvoker interface | Medium |
| ayo-gtm5 | Externalize remaining hardcoded prompts | High |
| ayo-gtm6 | gopls modernization (interface{} → any, etc.) | Low |
| ayo-gtm7 | Remove cmd/ayo/chat.go and dead code | Low |
| ayo-gtm8 | Verify and remove robfig/cron if unused | Low |

## Acceptance Criteria

- [ ] No file > 500 lines in core packages
- [ ] AgentInvoker defined in one location only
- [ ] Zero hardcoded prompt strings in Go code
- [ ] Zero gopls modernization hints
- [ ] No unused dependencies
