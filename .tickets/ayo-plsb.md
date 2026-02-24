---
id: ayo-plsb
status: open
deps: [ayo-6h19]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-plug
tags: [plugins, sandbox]
---
# Task: Sandbox Config Plugins

## Summary

Enable plugins to provide alternative sandbox configurations (harnesses). This allows specialized container setups like GPU-enabled sandboxes, specific language environments, or custom network configurations to be distributed as plugins.

## Context

Currently sandbox configuration is limited to:
- Provider selection (Apple Container vs systemd-nspawn)
- Basic options in ayo.json (network, mounts)
- Manual image specification

There's no way to:
- Package pre-configured sandbox environments
- Share specialized container setups
- Distribute environment recipes

## Use Cases

1. **GPU-enabled sandbox** - For ML/AI development agents
2. **Database sandbox** - PostgreSQL/MySQL pre-configured
3. **Node.js sandbox** - Specific Node version + common packages
4. **Python sandbox** - Conda/pip environments pre-configured
5. **Minimal sandbox** - Stripped-down for security-sensitive tasks

## Technical Approach

### Plugin Directory Structure

```
gpu-sandbox-plugin/
├── manifest.json
└── sandboxes/
    └── gpu-enabled/
        ├── sandbox.json       # Sandbox configuration
        ├── setup.sh           # Post-creation setup script
        └── packages.txt       # Packages to install
```

### Sandbox Config Schema

```json
// sandboxes/gpu-enabled/sandbox.json
{
  "name": "gpu-enabled",
  "description": "Sandbox with GPU passthrough for ML tasks",
  "base_image": "nvidia/cuda:12.0-runtime",
  "provider_requirements": {
    "gpu": true,
    "min_memory": "8G"
  },
  "mounts": [
    { "src": "/dev/nvidia0", "dst": "/dev/nvidia0", "type": "device" }
  ],
  "env": {
    "CUDA_VISIBLE_DEVICES": "0"
  },
  "packages": ["python3", "pip", "numpy", "torch"],
  "post_create": "setup.sh"
}
```

### Manifest Entry

```json
{
  "name": "gpu-sandbox-plugin",
  "version": "1.0.0",
  "components": {
    "sandbox_configs": {
      "gpu-enabled": {
        "path": "sandboxes/gpu-enabled",
        "description": "GPU-enabled sandbox for ML development",
        "requirements": ["nvidia-gpu"]
      }
    }
  }
}
```

### Usage

```json
// Agent ayo.json
{
  "agent": {
    "sandbox": {
      "config": "gpu-enabled",  // Reference plugin config
      "isolated": true
    }
  }
}
```

Or via CLI:
```bash
ayo sandbox create --config gpu-enabled my-ml-sandbox
```

## Implementation Steps

1. [ ] Design sandbox.json schema
2. [ ] Add `sandbox_configs` component to manifest
3. [ ] Implement sandbox config loading from plugins
4. [ ] Update sandbox creation to use plugin configs
5. [ ] Add config requirements validation
6. [ ] Support post-create setup scripts
7. [ ] Update `ayo sandbox create` CLI
8. [ ] Add `ayo sandbox configs` list command
9. [ ] Document sandbox config plugins

## Dependencies

- Depends on: `ayo-6h19` (foundation/ayod)
- Blocks: Specialized sandbox plugins

## Acceptance Criteria

- [ ] Plugins can define sandbox configurations
- [ ] `ayo sandbox configs` lists available configs
- [ ] Agents can reference sandbox configs by name
- [ ] Post-create scripts execute after sandbox creation
- [ ] Requirements are validated before creation
- [ ] Documentation covers sandbox config development

## Files to Create

- `internal/plugins/sandbox_configs.go` - Config loading
- `internal/sandbox/configs/` - Config resolution

## Files to Modify

- `internal/plugins/manifest.go` - Add sandbox_configs
- `internal/plugins/registry.go` - Register configs
- `internal/sandbox/manager.go` - Use plugin configs
- `cmd/ayo/sandbox.go` - Add configs command

## Security Considerations

- Device passthrough requires careful validation
- Post-create scripts run inside sandbox (not host)
- Provider requirements must be verified before creation
- Network configuration should follow existing guardrails

## Notes

- Some configs may be provider-specific (GPU only on Linux/systemd-nspawn)
- Consider layered configs (base + overrides)
- May want config inheritance for common patterns

---

*Created: 2026-02-23*
