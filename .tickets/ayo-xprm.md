---
id: ayo-xprm
status: open
deps: []
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [gtm, phase1, prompts]
---
# Task: Externalized Prompts System

## Summary

Remove ALL hardcoded prompts from the codebase. All system prompts, guardrails, sandwich patterns, and instruction text must be sourced at runtime from `~/.local/share/ayo/prompts/`.

## Context

Current state has hardcoded prompts scattered throughout:
- `LegacyGuardrails` with 6 hardcoded rules
- Sandwich pattern prefix/suffix
- Agent system prompt templates
- Tool usage instructions
- Error message templates

Problems with hardcoding:
1. Users can't inspect what prompts are being used
2. Can't customize without modifying source
3. Hurts plugin system - plugins can't override prompts
4. Prevents alternative harness experimentation
5. Opaque behavior

## Solution

All prompts live in `~/.local/share/ayo/prompts/` with clear organization:

```
~/.local/share/ayo/prompts/
├── system/
│   ├── base.md              # Base system prompt for all agents
│   ├── tool-usage.md        # How to use tools
│   ├── memory-usage.md      # How to use memory
│   └── planning.md          # Planning/reasoning guidance
├── guardrails/
│   ├── default.md           # Default guardrails
│   ├── safety.md            # Safety-critical rules
│   └── sandbox-aware.md     # Sandbox-specific rules
├── sandwich/
│   ├── prefix.md            # Conversation prefix
│   └── suffix.md            # Conversation suffix
├── agents/
│   └── @ayo/
│       └── system.md        # @ayo specific prompt
├── errors/
│   ├── tool-failed.md       # Tool failure message
│   └── permission-denied.md # Permission error
└── templates/
    ├── agent-response.md    # Response formatting
    └── delegation.md        # Delegation instructions
```

## Prompt Loading

```go
package prompts

type PromptLoader struct {
    baseDir string
    cache   map[string]string
    mu      sync.RWMutex
}

func NewPromptLoader() *PromptLoader {
    return &PromptLoader{
        baseDir: filepath.Join(xdg.DataHome, "ayo", "prompts"),
        cache:   make(map[string]string),
    }
}

// Load returns prompt content, with caching
func (l *PromptLoader) Load(path string) (string, error) {
    l.mu.RLock()
    if cached, ok := l.cache[path]; ok {
        l.mu.RUnlock()
        return cached, nil
    }
    l.mu.RUnlock()
    
    fullPath := filepath.Join(l.baseDir, path)
    content, err := os.ReadFile(fullPath)
    if err != nil {
        return "", fmt.Errorf("prompt not found: %s", path)
    }
    
    l.mu.Lock()
    l.cache[path] = string(content)
    l.mu.Unlock()
    
    return string(content), nil
}

// MustLoad panics if prompt not found (for required prompts)
func (l *PromptLoader) MustLoad(path string) string {
    content, err := l.Load(path)
    if err != nil {
        panic(fmt.Sprintf("required prompt missing: %s", path))
    }
    return content
}

// Refresh clears cache (for development/testing)
func (l *PromptLoader) Refresh() {
    l.mu.Lock()
    l.cache = make(map[string]string)
    l.mu.Unlock()
}
```

## Default Prompts Installation

On first run or `ayo doctor --fix`, install default prompts:

```go
func InstallDefaultPrompts(force bool) error {
    promptsDir := filepath.Join(xdg.DataHome, "ayo", "prompts")
    
    // Check if already installed
    if !force && dirExists(promptsDir) {
        return nil
    }
    
    // Extract embedded defaults
    return fs.WalkDir(embeddedPrompts, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil || d.IsDir() {
            return err
        }
        
        destPath := filepath.Join(promptsDir, path)
        os.MkdirAll(filepath.Dir(destPath), 0755)
        
        content, _ := embeddedPrompts.ReadFile(path)
        return os.WriteFile(destPath, content, 0644)
    })
}
```

## Embedded Defaults

Keep defaults embedded for installation, but NEVER use them at runtime:

```go
//go:embed prompts/*
var embeddedPrompts embed.FS
```

## Plugin Prompt Overrides

Plugins can provide prompt overrides in their manifest:

```json
{
  "name": "my-plugin",
  "prompts": {
    "guardrails/default.md": "prompts/my-guardrails.md",
    "system/tool-usage.md": "prompts/my-tool-usage.md"
  }
}
```

Plugin prompts are layered:
1. User prompts (highest priority)
2. Plugin prompts
3. Default prompts (lowest priority)

## Prompt CLI Commands

```bash
# List all prompts
ayo prompt list

# View a prompt
ayo prompt show system/base.md

# Edit a prompt (opens in $EDITOR)
ayo prompt edit guardrails/default.md

# Reset a prompt to default
ayo prompt reset guardrails/default.md

# Reset all prompts
ayo prompt reset --all

# Validate prompts (check for missing required)
ayo prompt validate
```

## Required Prompts

These must exist or agent fails to start:
- `system/base.md`
- `system/tool-usage.md`
- `guardrails/default.md`

## Implementation Steps

1. [ ] Create `internal/prompts/` package
2. [ ] Implement PromptLoader with caching
3. [ ] Create embedded default prompts
4. [ ] Implement InstallDefaultPrompts
5. [ ] Add plugin prompt overlay support
6. [ ] Replace ALL hardcoded prompts with loader calls
7. [ ] Add `ayo prompt` CLI commands
8. [ ] Add `ayo doctor` check for prompts
9. [ ] Document prompt customization
10. [ ] Update agent loading to use prompt loader
11. [ ] Test prompt override layering

## Migration

Search and replace all hardcoded prompts:

```bash
# Find hardcoded prompt strings
grep -rn "You are" internal/
grep -rn "Always " internal/
grep -rn "Never " internal/
grep -rn "guardrail" internal/
```

## Dependencies

- Depends on: None
- Blocks: `ayo-grls` (guardrails refactor)

## Acceptance Criteria

- [ ] Zero hardcoded prompts in codebase
- [ ] All prompts in `~/.local/share/ayo/prompts/`
- [ ] Users can view all prompts via CLI
- [ ] Users can edit prompts
- [ ] Plugins can override prompts
- [ ] `ayo doctor` validates prompt installation
- [ ] Documentation explains prompt customization

---

*Created: 2026-02-23*
