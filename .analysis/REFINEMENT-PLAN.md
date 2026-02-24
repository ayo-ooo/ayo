# GTM Refinement Plan

## Executive Summary

This document outlines a comprehensive refinement plan for GTM readiness. Every change is planned before execution to ensure clean, well-factored code.

---

## 1. LARGE FILE SPLITTING

### 1.1 internal/agent/agent.go (1320 lines → 5 files)

**Current structure analysis:**
- Lines 1-250: Types, constants, Config structs
- Lines 251-400: Guardrails, trust level methods
- Lines 400-650: Loading functions (ListHandles, Load, loadFromDir)
- Lines 650-750: Context builders (delegate, env)
- Lines 750-900: Save functions
- Lines 900-1100: Schema loading and validation
- Lines 1100-1320: Compatibility checking

**Target files:**
| File | Contents | Lines |
|------|----------|-------|
| `types.go` | TrustLevel, Config, ModelConfig, MemoryConfig, SandboxConfig structs | ~250 |
| `trust.go` | Trust level methods, guardrails logic | ~150 |
| `loading.go` | ListHandles, Load, loadFromDir, resolveModel | ~250 |
| `saving.go` | Save, SaveWithSchemas, copySchemaFile | ~100 |
| `schema.go` | Schema loading, validation, compatibility | ~250 |
| `agent.go` | Agent struct, core methods, context builders | ~320 |

### 1.2 internal/daemon/server.go (1170 lines → 5 files)

**Current structure analysis:**
- Lines 1-150: Server struct, constructor, lifecycle
- Lines 150-400: RPC method registration, core handlers
- Lines 400-600: Session handlers
- Lines 600-800: Sandbox handlers
- Lines 800-1000: Memory handlers
- Lines 1000-1170: Utility functions

**Target files:**
| File | Contents | Lines |
|------|----------|-------|
| `server.go` | Server struct, lifecycle, registration | ~200 |
| `handlers_session.go` | Session RPC handlers | ~200 |
| `handlers_sandbox.go` | Sandbox RPC handlers | ~200 |
| `handlers_memory.go` | Memory RPC handlers | ~200 |
| `handlers_agent.go` | Agent RPC handlers | ~200 |
| `handlers_util.go` | Shared handler utilities | ~170 |

### 1.3 Other Large Files (assess for splitting)

| File | Lines | Action |
|------|-------|--------|
| cmd/ayo/sandbox.go | 1651 | Consider splitting by subcommand |
| internal/run/run.go | 1410 | Consider splitting by phase |
| cmd/ayo/memory.go | 1380 | Consider splitting by subcommand |
| cmd/ayo/triggers.go | 1118 | Consider splitting by trigger type |
| cmd/ayo/flows.go | 1100 | Consider splitting by subcommand |
| internal/ui/ui.go | 1072 | Consider splitting by component |

---

## 2. INTERFACE CONSOLIDATION

### 2.1 AgentInvoker (2 definitions → 1)

**Current locations:**
- `internal/daemon/invoker.go:16` 
- `internal/squads/invoker.go:10`

**Action:** Create `internal/interfaces/invoker.go` with canonical definition, update both locations to import.

### 2.2 Logger (2 definitions)

**Current locations:**
- `internal/audit/audit.go:46`
- `internal/daemon/trigger_loader.go:79`

**Action:** Create `internal/interfaces/logger.go` with canonical definition.

### 2.3 FormRenderer (2 definitions)

**Current locations:**
- `internal/tools/humaninput/humaninput.go:35`
- `internal/hitl/timeout.go:39`

**Action:** Create `internal/interfaces/form.go` with canonical definition.

---

## 3. DEAD CODE REMOVAL

### 3.1 Deprecated Commands

| File | Issue | Action |
|------|-------|--------|
| cmd/ayo/chain.go | Deprecated in favor of `flow schema` | Remove entirely |

### 3.2 Unused Functions (from gopls)

| Location | Issue | Action |
|----------|-------|--------|
| cmd/ayo/doctor.go:366 | `dirExists` unused | Remove |
| cmd/ayo/root.go:557 | `knownSubcommands` unused | Remove |
| Various plugin.go | Unused `cfgPath` params | Fix signature or use |

---

## 4. GOPLS MODERNIZATION (Remaining hints)

### 4.1 slices.Contains improvements

| Location | Current | Target |
|----------|---------|--------|
| agent.go:471 | Manual loop | `slices.Contains()` |
| agent.go:937 | Manual loop | `slices.Contains()` |

### 4.2 omitzero hints (informational)

These are hints about `omitempty` on struct fields - they work but have no effect. Consider using `omitzero` tag for Go 1.24+.

### 4.3 stringsseq improvements

| Location | Current | Target |
|----------|---------|--------|
| flows.go:322 | `strings.Split` in loop | `strings.SplitSeq` |
| flows.go:775 | `strings.Split` in loop | `strings.SplitSeq` |

### 4.4 rangeint improvements

| Location | Current | Target |
|----------|---------|--------|
| flows.go:334 | `for i := 0; i < n; i++` | `for i := range n` |

---

## 5. PROMPT EXTERNALIZATION

### 5.1 Hardcoded prompts in guardrails/defaults.go

| Constant | Lines | Target File |
|----------|-------|-------------|
| `defaultPrefix` | 10-22 | `prompts/guardrails/prefix.md` |
| `defaultSuffix` | 23-107 | `prompts/guardrails/suffix.md` |
| `defaultLegacyGuardrails` | 108+ | `prompts/guardrails/legacy.md` |

### 5.2 Implementation approach

1. Create `internal/prompts/embed/` directory with markdown files
2. Use `//go:embed` to include defaults
3. Add runtime loading with fallback to embedded
4. Update `ayo setup` to optionally copy to user directory

---

## 6. CONSISTENCY IMPROVEMENTS

### 6.1 Error message formatting

Audit all error messages for consistency:
- Lowercase first letter (Go convention)
- No trailing punctuation
- Include context

### 6.2 Logging consistency

Ensure all packages use consistent logging patterns.

### 6.3 JSON tag consistency

Audit all structs for consistent JSON tag usage.

---

## 7. EXECUTION ORDER

| Phase | Tasks | Risk |
|-------|-------|------|
| 1 | Remove dead code (chain.go, unused functions) | Low |
| 2 | gopls modernization (slices.Contains, stringsseq, rangeint) | Low |
| 3 | Interface consolidation | Medium |
| 4 | Prompt externalization | Medium |
| 5 | Split agent/agent.go | High |
| 6 | Split daemon/server.go | High |
| 7 | Split large cmd files | Medium |
| 8 | Final verification | N/A |

---

## 8. VERIFICATION CHECKLIST

After each phase:
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
- [ ] No new gopls warnings
- [ ] Functionality verified

---

## 9. FILES TO CREATE

```
internal/interfaces/
├── invoker.go      # AgentInvoker interface
├── logger.go       # Logger interface  
└── form.go         # FormRenderer interface

internal/agent/
├── agent.go        # (refactored - core only)
├── types.go        # (new)
├── trust.go        # (new)
├── loading.go      # (new)
├── saving.go       # (new)
└── schema.go       # (new)

internal/daemon/
├── server.go       # (refactored - core only)
├── handlers_session.go  # (new)
├── handlers_sandbox.go  # (new)
├── handlers_memory.go   # (new)
├── handlers_agent.go    # (new)
└── handlers_util.go     # (new)

internal/prompts/
├── embed/
│   └── guardrails/
│       ├── prefix.md
│       ├── suffix.md
│       └── legacy.md
└── loader.go       # Runtime loading with embedded fallback
```

## 10. FILES TO DELETE

```
cmd/ayo/chain.go    # Deprecated command
```
