---
id: ayo-bs7
status: open
deps: [ayo-bs2]
links: []
created: 2026-03-11T18:00:00Z
type: epic
priority: 7
assignee: Alex Cabrera
tags: [build-system, configuration, validation]
---
# Phase 7: Configuration System

Robust configuration parsing and validation for both build-time and runtime. Ensures config.toml is comprehensive, validated, and supports necessary overrides.

## Context

The configuration system needs to handle:
1. Build-time configuration (agent definition, skills, tools)
2. Runtime configuration (model settings, CLI behavior)
3. Environment variable overrides
4. Configuration validation and error handling
5. Migration path from old config formats

## Tasks

### 7.1 Complete config.toml Schema
- [ ] Finalize all configuration sections
- [ ] Add validation rules for each field
- [ ] Document all configuration options
- [ ] Support nested configuration structures
- [ ] Add configuration examples

### 7.2 Implement Config Validation
- [ ] Validate required fields
- [ ] Validate field types and ranges
- [ ] Validate enum values
- [ ] Check for deprecated fields
- [ ] Provide clear validation errors

### 7.3 Configuration Merging
- [ ] Implement default config loading
- [ ] Merge agent config with defaults
- [ ] Handle conflicting settings
- [ ] Support partial configs
- [ ] Document merge precedence

### 7.4 Environment Variable Overrides
- [ ] Map config fields to env vars
- [ ] Parse env var values correctly
- [ ] Support boolean, string, number types
- [ ] Document env var naming convention
- [ ] Add examples

### 7.5 Configuration Migration
- [ ] Detect old config formats
- [ ] Provide migration warnings
- [ ] Auto-migrate compatible configs
- [ ] Document breaking changes
- [ ] Add migration guide

### 7.6 Model Provider Discovery
- [ ] Discover providers from environment variables (OPENAI_API_KEY, ANTHROPIC_API_KEY, etc.)
- [ ] Support custom providers from config
- [ ] Parse provider configurations (base_url, type, models)
- [ ] Validate provider credentials
- [ ] Provider priority and deduplication

### 7.7 Model Selection UI
- [ ] Implement TUI model selection dialog (bubbletea/huh)
- [ ] Show available providers and their models
- [ ] Support filtering/search
- [ ] Show recently used models
- [ ] Handle multiple providers from env vars
- [ ] Show nice UI to pick model when multiple options exist
- [ ] Show nice UI to pick model when no options configured

### 7.8 First-Run Onboarding
- [ ] Detect first run of built executable
- [ ] Show model selection dialog on first run
- [ ] Prompt for API keys if needed
- [ ] Allow model provider selection
- [ ] Mark agent as initialized after setup
- [ ] Persist model selection

### 7.9 API Key Management
- [ ] Support API key input in UI
- [ ] Store API keys securely (env vars only, no local storage)
- [ ] Allow editing API keys (ctrl+e like crush)
- [ ] Support OAuth tokens for providers that need them
- [ ] Re-auth prompt on auth errors

### 7.10 Configuration Testing
- [ ] Unit tests for validation
- [ ] Unit tests for merging
- [ ] Unit tests for env overrides
- [ ] Unit tests for migration
- [ ] Unit tests for provider discovery
- [ ] Unit tests for model selection UI
- [ ] Test coverage > 90%

## Technical Details

### config.toml Structure

```toml
[agent]
name = "my-agent"
description = "My AI assistant"
model = "claude-3-5-sonnet"
max_tokens = 4096
temperature = 0.7

[cli]
mode = "hybrid"            # freeform | hybrid | structured
description = "My agent CLI"

[cli.flags]
# Flags auto-generated from input.jsonschema
# Can be manually overridden if needed

[agent.tools]
allowed = ["bash", "file_read", "file_write"]
builtin_enabled = true

[agent.skills]
include = []              # Explicit skill list (empty = all)
exclude = []
ignore_builtin = false

[agent.memory]
enabled = false
type = "sqlite"
path = "memory.db"

[agent.output]
validate_schema = true
retry_on_failure = true
max_retries = 3
```

### Environment Variable Mapping

```bash
AYO_AGENT_MODEL=claude-3-5-sonnet
AYO_CLI_MODE=hybrid
AYO_MEMORY_ENABLED=true

# Model providers (discovered from environment)
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
GEMINI_API_KEY=...
```

### Provider Configuration

Providers are discovered from:
1. Environment variables (OPENAI_API_KEY, ANTHROPIC_API_KEY, etc.)
2. Custom providers defined in config
3. Predefined provider list (from schema/registry)

Provider config structure:
```toml
[providers.openai]
id = "openai"
name = "OpenAI"
type = "openai"              # openai, anthropic, gemini, etc.
api_key = "$OPENAI_API_KEY"   # Can be env var or literal
base_url = "https://api.openai.com/v1"
models = [
  { id = "gpt-4o", name = "GPT-4o", context_window = 128000 },
  { id = "gpt-4o-mini", name = "GPT-4o Mini", context_window = 128000 }
]
```

### Model Selection UI

When running agent:
1. Check if model is configured
2. If multiple providers have API keys → show selection UI
3. If no providers configured → show selection + API key input UI
4. On first run → show onboarding flow
5. Store selection for subsequent runs (in config or env)

UI features:
- List providers and models
- Filter/search models
- Show recently used models
- Press ctrl+e to edit API key
- Support keyboard navigation
- Use bubbletea/huh for TUI

### Validation Rules

- agent.name: Required, alphanumeric with hyphens
- agent.model: Required, must be valid model identifier
- agent.temperature: Optional, 0.0-2.0
- cli.mode: Optional, must be one of: freeform, hybrid, structured
- agent.tools.allowed: Optional, array of valid tool names

## Deliverables

- [ ] Complete config.toml schema
- [ ] Config validation implementation
- [ ] Config merging logic
- [ ] Environment variable support
- [ ] Configuration migration path
- [ ] Model provider discovery system
- [ ] Model selection TUI (bubbletea/huh)
- [ ] First-run onboarding flow
- [ ] API key management
- [ ] Test coverage > 90%
- [ ] Configuration documentation
- [ ] Model selection documentation

## Acceptance Criteria

1. All config.toml fields are validated
2. Invalid configs show clear error messages
3. Default configs merge correctly
4. Environment variables override config
5. Old configs migrate with warnings
6. Documentation covers all options
7. Model providers discovered from env vars
8. Model selection UI shows when multiple providers exist
9. Model selection UI shows when no providers configured
10. First-run onboarding works correctly
11. API keys can be edited in UI (ctrl+e)
12. No local storage of API keys (env vars only)

## Dependencies

- **ayo-bs2**: Build system core (uses config at build time)
- **ayo-bs6**: Runtime execution (uses config at runtime)

## Out of Scope

- Hot-reloading of configuration
- Remote configuration loading
- Dynamic configuration via API

## Risks

- **Breaking Changes**: New config format may break existing agents
  - **Mitigation**: Provide migration path, clear documentation

## Notes

Configuration is the foundation - make it robust and well-documented.
