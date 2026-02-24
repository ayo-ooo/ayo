# Security Configuration Guide

Complete reference for ayo's security model and guardrails.

## Security Model Overview

Ayo's security is built on three principles:

1. **Isolation**: Agents run in sandboxes, not on your host
2. **Explicit Permission**: Host modifications require approval
3. **Auditability**: All actions are logged

## Sandbox Isolation

### What's Isolated

| Resource | Isolation |
|----------|-----------|
| Filesystem | Container rootfs, separate from host |
| Processes | Namespace isolation |
| Network | Disabled by default |
| Users | Container users, not host users |

### What's Shared

| Resource | Access |
|----------|--------|
| Host files | Read-only mount at `/mnt/{username}/` |
| Output | Write to `/output/`, syncs to host |
| Time | System clock (read-only) |

### Trust Levels

| Level | Sandbox | Host Access | Guardrails |
|-------|---------|-------------|------------|
| `sandboxed` | Yes | Read-only mount | Yes |
| `privileged` | Yes | Read + file_request | Yes |
| `unrestricted` | No | Full access | No |

**Default**: `sandboxed`

Configure in agent `config.json`:

```json
{
  "trust_level": "sandboxed"
}
```

## file_request Flow

### How It Works

1. Agent wants to modify host file
2. Agent calls `file_request` tool
3. Request appears in terminal
4. User reviews and approves/denies
5. If approved, file is written

### Approval Prompt

```
┌─────────────────────────────────────────────┐
│ @ayo wants to write:                        │
│   ~/Projects/app/main.go                    │
│                                             │
│ Reason: Fixed authentication bug            │
│                                             │
│ [Y]es  [N]o  [D]iff  [A]lways for session   │
└─────────────────────────────────────────────┘
```

### Approval Options

| Key | Action | Persistence |
|-----|--------|-------------|
| `Y` | Approve this request | Single request |
| `N` | Deny request | Single request |
| `D` | View diff | - |
| `A` | Approve this and similar | Session |

### Session Caching

"Always" (`A`) caches approval for:
- Same file path
- Same directory pattern
- Current session only

**Important**: Cache is NOT persisted to disk for security.

## --no-jodas Mode

### What It Does

Auto-approves all file_request prompts without asking.

### Usage

```bash
# CLI flag
ayo --no-jodas "refactor everything"

# Short flag
ayo -y "make changes"
```

### Configuration

Global config:
```json
{
  "permissions": {
    "no_jodas": true
  }
}
```

Agent config:
```json
{
  "permissions": {
    "auto_approve": true
  }
}
```

### Precedence

1. Session cache (highest)
2. CLI flag (`--no-jodas`)
3. Agent config (`permissions.auto_approve`)
4. Global config (`permissions.no_jodas`)

**Warning**: Use with caution. Agents can modify any file.

## Auto-Approve Patterns

### Configuration

```json
{
  "permissions": {
    "auto_approve_patterns": [
      "./build/*",
      "./dist/*",
      "./tmp/*",
      "./output/*"
    ]
  }
}
```

### Pattern Syntax

Uses doublestar glob patterns:

| Pattern | Matches |
|---------|---------|
| `*.txt` | Any .txt file |
| `./build/*` | Files in build/ |
| `./src/**/*.go` | All .go files in src/ recursively |
| `**/test/*` | Any test/ directory |

## Blocked Patterns

### Default Blocked

These patterns are ALWAYS blocked, regardless of approval:

```go
DefaultBlockedPatterns = []string{
    ".git/*",           // Git internals
    ".env*",            // Environment files
    "**/secrets/*",     // Secret directories
    "**/*.key",         // Private keys
    "**/*.pem",         // Certificates
    "**/id_rsa*",       // SSH keys
    "**/.ssh/*",        // SSH directory
    "**/credentials*",  // Credentials files
    "**/.aws/*",        // AWS config
    "**/.kube/*",       // Kubernetes config
}
```

### Custom Blocked Patterns

Add in config:

```json
{
  "permissions": {
    "blocked_patterns": [
      "**/production/*",
      "**/.secrets/*"
    ]
  }
}
```

## Guardrails

### What Are Guardrails

Safety prompts injected before and after agent prompts to constrain behavior.

### Default Guardrails

`~/.local/share/ayo/prompts/defaults/guardrails/default.md`:

```markdown
# Safety Guidelines

You are an AI assistant operating in a sandboxed environment.

## Boundaries

- Never execute commands that could harm the host system
- Never access or exfiltrate sensitive data
- Always explain what you're doing before doing it
- Request explicit permission for significant changes

## Prohibited Actions

- Accessing credentials or secrets
- Network requests to unknown endpoints
- Modifying system files
- Installing untrusted software
```

### Prompt Injection Order

1. System base prompt
2. **Guardrails prefix**
3. Agent system.md
4. Squad constitution (if applicable)
5. Skills
6. **Guardrails suffix**

### Customizing Guardrails

Override at `~/.config/ayo/prompts/guardrails/`:

```
prompts/
├── guardrails/
│   ├── default.md      # Override default
│   └── @reviewer.md    # Agent-specific
└── sandwich/
    ├── prefix.md       # Before agent prompt
    └── suffix.md       # After agent prompt
```

### Disabling Guardrails

For `unrestricted` agents, guardrails are disabled:

```json
{
  "trust_level": "unrestricted",
  "guardrails": false
}
```

**Warning**: Only for trusted, well-tested agents.

## Audit Logging

### What's Logged

| Event | Data |
|-------|------|
| File modification | Path, agent, timestamp, approval method |
| Command execution | Command, agent, timestamp |
| File request | Path, approved/denied, reason |
| Agent invocation | Agent, prompt, session |

### Log Location

```
~/.local/share/ayo/audit.log
```

### Log Format

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "event": "file_modified",
  "agent": "@ayo",
  "path": "~/Projects/app/main.go",
  "approval_method": "user_approved",
  "session_id": "sess_abc123"
}
```

### Viewing Audit Log

```bash
# List recent entries
ayo audit list

# Filter by agent
ayo audit list --agent @ayo

# Filter by path
ayo audit list --path ~/Projects/

# Filter by time
ayo audit list --since "24h"
```

### Audit Retention

Default: 90 days

Configure:
```json
{
  "audit": {
    "retention_days": 90,
    "max_size_mb": 100
  }
}
```

## Network Security

### Default: Disabled

```json
{
  "sandbox": {
    "network": false
  }
}
```

### Enabling Network

```json
{
  "sandbox": {
    "network": true
  }
}
```

**Risk**: Agent can make arbitrary network requests.

### Network Restrictions (Future)

```json
{
  "sandbox": {
    "network": {
      "enabled": true,
      "allowed_hosts": [
        "api.github.com",
        "pypi.org"
      ]
    }
  }
}
```

## Best Practices

### For Personal Use

1. Use `sandboxed` trust level
2. Review file_request prompts
3. Use `--no-jodas` sparingly
4. Review audit logs periodically

### For Shared Systems

1. Never use `unrestricted` trust
2. Configure blocked patterns for sensitive paths
3. Enable audit logging with rotation
4. Restrict network access

### For Production/Automation

1. Use `privileged` with auto_approve_patterns
2. Limit patterns to specific directories
3. Monitor audit logs
4. Use dedicated sandbox images
5. Disable network unless required

## Incident Response

### If Suspicious Activity Detected

1. **Stop the agent**:
   ```bash
   ayo sandbox service stop
   ```

2. **Review audit log**:
   ```bash
   ayo audit list --since "1h"
   ```

3. **Check modified files**:
   ```bash
   git status
   git diff
   ```

4. **Revoke any cached approvals**:
   ```bash
   ayo sandbox service restart
   ```

5. **Review agent configuration**:
   ```bash
   ayo agents show @name
   ```

### Recovery

1. Restore files from backup or git
2. Review and tighten permissions
3. Add patterns to `blocked_patterns`
4. Consider reducing trust level

## Security Checklist

- [ ] Trust levels appropriate for each agent
- [ ] Blocked patterns cover sensitive files
- [ ] Network disabled unless needed
- [ ] Audit logging enabled
- [ ] Regular audit log review
- [ ] `--no-jodas` used sparingly
- [ ] Guardrails not disabled unnecessarily
- [ ] Agent prompts reviewed for safety
