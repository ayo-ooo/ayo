---
id: ayo-e2e2
status: open
deps: []
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 0 - Clean State & Prerequisites

## Summary

Write Section 0 of the E2E Manual Testing Guide covering clean state management and system prerequisites.

## Content Requirements

### System Requirements
- macOS 26+ (Tahoe) with Apple Silicon
- Go 1.24+ (for building from source)
- Container runtime available (`container list` works)
- At least 8GB RAM recommended
- Disk space requirements (~500MB for sandboxes)

### Environment Verification Script
```bash
#!/bin/bash
# verify-prerequisites.sh

echo "=== Ayo E2E Testing Prerequisites Check ==="

# macOS version
sw_vers -productVersion

# Go version  
go version

# Container runtime
container list 2>/dev/null && echo "✓ Container runtime available"

# Disk space
df -h ~/.local/share/ | head -2

# RAM
sysctl -n hw.memsize | awk '{print $1/1024/1024/1024 " GB"}'

echo "=== Done ==="
```

### Clean State Script
```bash
#!/bin/bash
# clean-state.sh

echo "=== Cleaning Ayo State ==="

# Stop daemon if running
pkill -f ayod 2>/dev/null || true
rm -f ~/.local/share/ayo/daemon.sock
rm -f ~/.local/share/ayo/daemon.pid

# Remove all data directories
rm -rf ~/.local/share/ayo/sandboxes/
rm -rf ~/.local/share/ayo/sessions/
rm -rf ~/.local/share/ayo/memory/
rm -rf ~/.local/share/ayo/prompts/

# Remove config (optional - uncomment if needed)
# rm -rf ~/.config/ayo/

echo "✓ Clean state achieved"
echo "Ready for fresh E2E testing"
```

### Provider Configuration
- Required: API key for at least one provider
- Required: Embedding model for memory system
- Example configuration:
  ```json
  {
    "providers": {
      "anthropic": { "api_key": "..." },
      "ollama": { "base_url": "http://localhost:11434" }
    },
    "defaults": {
      "provider": "anthropic",
      "model": "claude-sonnet-4-20250514",
      "embedding_provider": "ollama",
      "embedding_model": "nomic-embed-text"
    }
  }
  ```

### Verification Criteria
- [ ] macOS 26+ confirmed
- [ ] Go 1.24+ installed
- [ ] Container runtime works
- [ ] Clean state script runs without errors
- [ ] Provider API key available and valid
- [ ] Embedding model available

## Acceptance Criteria

- [ ] Section written in guide
- [ ] verify-prerequisites.sh script included
- [ ] clean-state.sh script included
- [ ] Provider configuration documented
- [ ] All verification criteria listed
