# Migration Guide

This guide covers migrating from older versions of ayo to the current version with sandbox and Zettelkasten memory support.

## Version Compatibility

| From Version | To Version | Migration Complexity |
|--------------|------------|---------------------|
| 0.1.x | 0.2.x | Low |
| 0.2.x | 0.3.x | Medium |

## Migrating from SQLite Sessions to JSONL

### Automatic Migration

Run the built-in migration command:

```bash
ayo sessions migrate
```

This converts sessions from `~/.local/share/ayo/ayo.db` to JSONL files in `~/.local/share/ayo/sessions/`.

### Migration Options

```bash
# Migrate with overwrite existing
ayo sessions migrate --overwrite

# Rebuild search index after migration
ayo sessions reindex
```

### Manual Migration

If automatic migration fails:

1. Export from SQLite:
```bash
sqlite3 ~/.local/share/ayo/ayo.db ".dump sessions"
```

2. Create JSONL files manually following the format in `internal/session/jsonl/FORMAT.md`

## Migrating to Zettelkasten Memory

### Understanding the Change

**Old system:** Memories stored in SQLite as rows with embeddings as BLOBs.

**New system:** Memories stored as Markdown files with TOML frontmatter, indexed in SQLite.

### Directory Structure

```
~/.local/share/ayo/memory/
├── index.md              # Auto-generated overview
├── preferences/          # User preferences
│   └── go-testing.md
├── facts/                # Project/user facts
│   └── api-endpoint.md
├── corrections/          # Behavior corrections
└── .index.sqlite         # Search index (derived)
```

### File Format

Each memory file:

```markdown
+++
id = "mem_01HX..."
created = 2024-01-15T10:30:00Z
category = "preference"
topics = ["go", "testing"]
source = "session:abc123"
status = "active"
+++

User strongly prefers table-driven tests in Go.
```

### Migration Steps

1. **Backup existing memories:**
```bash
cp ~/.local/share/ayo/ayo.db ~/.local/share/ayo/ayo.db.backup
```

2. **Export memories:**
```bash
sqlite3 ~/.local/share/ayo/ayo.db \
  "SELECT id, category, content FROM memories WHERE deleted = 0"
```

3. **Create memory files:**
```bash
mkdir -p ~/.local/share/ayo/memory/preferences
mkdir -p ~/.local/share/ayo/memory/facts
mkdir -p ~/.local/share/ayo/memory/corrections
mkdir -p ~/.local/share/ayo/memory/patterns
```

4. **Rebuild index:**
```bash
ayo memory reindex
```

## Migrating Agent Configurations

### Sandbox Configuration

Add sandbox config to agents that need isolation:

```json
{
  "sandbox": {
    "enabled": true,
    "provider": "docker",
    "image": "golang:1.22",
    "network": true,
    "mounts": [
      {"source": ".", "target": "/workspace"}
    ]
  }
}
```

### Memory Configuration

Update memory settings in agent config:

```json
{
  "memory": {
    "enabled": true,
    "scope": "hybrid",
    "retrieval": {
      "auto_inject": true,
      "threshold": 0.3,
      "max_memories": 10
    }
  }
}
```

## Provider Configuration

### New Provider Section

Add to `~/.config/ayo/ayo.json`:

```json
{
  "providers": {
    "sandbox": {
      "type": "docker",
      "pool": {
        "min_size": 1,
        "max_size": 4
      }
    },
    "embedding": {
      "type": "ollama",
      "model": "nomic-embed-text"
    },
    "memory": {
      "type": "zettelkasten"
    }
  }
}
```

## Daemon Setup

### First Time Setup

The daemon auto-starts when needed. For manual control:

```bash
# Check status
ayo status

# Start manually
ayo daemon start

# Stop
ayo daemon stop
```

### Systemd Service (Linux)

Create `/etc/systemd/user/ayo-daemon.service`:

```ini
[Unit]
Description=Ayo Daemon
After=network.target

[Service]
ExecStart=%h/.local/bin/ayo daemon start --foreground
Restart=on-failure

[Install]
WantedBy=default.target
```

Enable:
```bash
systemctl --user enable ayo-daemon
systemctl --user start ayo-daemon
```

### LaunchAgent (macOS)

Create `~/Library/LaunchAgents/land.charm.ayo.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>land.charm.ayo</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/ayo</string>
        <string>daemon</string>
        <string>start</string>
        <string>--foreground</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

Load:
```bash
launchctl load ~/Library/LaunchAgents/land.charm.ayo.plist
```

## Troubleshooting Migration

### Sessions Not Appearing

```bash
# Rebuild index
ayo sessions reindex

# Check session files exist
ls ~/.local/share/ayo/sessions/
```

### Memories Not Found

```bash
# Rebuild memory index
ayo memory reindex

# Check memory files exist
ls ~/.local/share/ayo/memory/
```

### Daemon Issues

```bash
# Check daemon status
ayo status

# View daemon logs
AYO_DEBUG=1 ayo daemon start --foreground
```

## Breaking Changes

### v0.2.x to v0.3.x

1. **Session storage format changed** - Use `ayo sessions migrate`
2. **Memory tool parameters changed** - `operation` replaced with `action`
3. **Provider configuration moved** - Now under `providers` key in config

### Deprecated Features

- Direct SQLite memory access (use memory tool instead)
- `--sandbox` CLI flag (use agent config instead)
- Old memory tool operations (`store_memory`, `search_memory`)

## Rolling Back

If migration fails:

1. Restore SQLite backup:
```bash
cp ~/.local/share/ayo/ayo.db.backup ~/.local/share/ayo/ayo.db
```

2. Remove new directories:
```bash
rm -rf ~/.local/share/ayo/sessions/
rm -rf ~/.local/share/ayo/memory/
```

3. Downgrade ayo binary to previous version
