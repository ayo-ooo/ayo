# Troubleshooting

Comprehensive guide to diagnosing and resolving common ayo issues.

## Diagnostic Tools

### ayo doctor

The primary diagnostic tool:

```bash
ayo doctor
```

**Output sections:**

| Section | Checks |
|---------|--------|
| System | OS version, architecture |
| Daemon | Status, socket, PID |
| Sandbox | Provider availability, permissions |
| LLM | API connectivity, credentials |
| Memory | Database integrity, embedding provider |
| Plugins | Loading, conflicts |

### Debug Logging

Enable verbose output:

```bash
AYO_DEBUG=1 ayo run code "Fix the bug"
```

**Log levels:**

| Level | Variable | Output |
|-------|----------|--------|
| Normal | (default) | Minimal |
| Debug | `AYO_DEBUG=1` | Verbose |
| Trace | `AYO_TRACE=1` | Very verbose |

### Daemon Logs

View daemon activity:

```bash
tail -f ~/.local/share/ayo/daemon.log
```

### Debug Scripts

Located in `debug/` directory:

| Script | Purpose |
|--------|---------|
| `system-info.sh` | Host system information |
| `sandbox-status.sh` | Container status and health |
| `daemon-status.sh` | Daemon process and socket status |

## Common Issues

### Daemon Issues

#### Daemon Won't Start

**Symptom:** `ayo service start` hangs or returns error

**Diagnosis:**
```bash
# Check for existing socket
ls -la ~/.local/share/ayo/daemon.sock

# Check for running daemon
ps aux | grep ayod

# Check PID file
cat ~/.local/share/ayo/daemon.pid
```

**Solutions:**

1. **Remove stale socket:**
   ```bash
   rm -f ~/.local/share/ayo/daemon.sock
   ayo service start
   ```

2. **Kill orphaned process:**
   ```bash
   kill $(cat ~/.local/share/ayo/daemon.pid)
   rm ~/.local/share/ayo/daemon.pid
   ayo service start
   ```

3. **Check permissions:**
   ```bash
   ls -la ~/.local/share/ayo/
   # Should be owned by your user
   ```

4. **Check logs:**
   ```bash
   cat ~/.local/share/ayo/daemon.log | tail -50
   ```

#### Daemon Connection Refused

**Symptom:** "connection refused" errors

**Solutions:**

1. **Start daemon:**
   ```bash
   ayo service start
   ```

2. **Check socket exists:**
   ```bash
   ls ~/.local/share/ayo/daemon.sock
   ```

3. **Verify daemon running:**
   ```bash
   ayo service status
   ```

### Sandbox Issues

#### Sandbox Creation Fails

**Symptom:** "failed to create sandbox" or "provider not available"

**Diagnosis:**
```bash
ayo doctor
# Check "Sandbox" section
```

**Solutions by platform:**

**macOS:**
```bash
# Verify macOS version (requires 26+)
sw_vers

# Check containerization entitlement
# App must be signed with container entitlement
```

**Linux:**
```bash
# Check systemd-nspawn availability
which systemd-nspawn

# Install if missing (Debian/Ubuntu)
sudo apt install systemd-container

# Check permissions
ls -la /var/lib/machines/
```

#### Sandbox Timeout

**Symptom:** Operations take too long, timeout errors

**Solutions:**

1. **Check system resources:**
   ```bash
   # Memory
   free -h
   
   # Disk space
   df -h ~/.local/share/ayo/sandboxes/
   ```

2. **Increase timeout:**
   ```json
   // ~/.config/ayo/config.json
   {
     "sandbox": {
       "timeout": 120
     }
   }
   ```

3. **Warm up sandbox pool:**
   ```bash
   ayo service start
   # Pool warms automatically
   ```

#### Sandbox Exec Fails

**Symptom:** Commands fail inside sandbox

**Diagnosis:**
```bash
# Get sandbox shell
ayo squad shell my-squad

# Try command manually
ls -la /workspace
```

**Solutions:**

1. **Check ayod running:**
   ```bash
   # Inside sandbox
   ls /run/ayod.sock
   ```

2. **Check user exists:**
   ```bash
   # Inside sandbox
   id agent-name
   ```

3. **Check permissions:**
   ```bash
   # Inside sandbox
   ls -la /workspace
   ```

### Agent Issues

#### Agent Not Found

**Symptom:** "agent @name not found"

**Diagnosis:**
```bash
# List available agents
ayo agent list

# Check agent path
ls ~/.config/ayo/agents/
```

**Solutions:**

1. **Check name spelling** (case-sensitive)

2. **Check agent files exist:**
   ```bash
   ls ~/.config/ayo/agents/name/
   # Should contain agent.md or ayo.json
   ```

3. **Check plugin provides agent:**
   ```bash
   ayo plugin list
   ```

4. **Validate agent config:**
   ```bash
   ayo agent show name
   ```

#### Agent Hangs

**Symptom:** Agent doesn't respond

**Solutions:**

1. **Check LLM connectivity:**
   ```bash
   ayo doctor
   # Check "LLM" section
   ```

2. **Check API key:**
   ```bash
   # Check for any configured provider API keys
   env | grep -E "_API_KEY$"
   # At least one should be set
   ```

3. **Try with debug:**
   ```bash
   AYO_DEBUG=1 ayo run agent "test"
   ```

### Memory Issues

#### Memory Search Returns Nothing

**Symptom:** `ayo memory search` returns empty

**Diagnosis:**
```bash
# Check memories exist
ayo memory list

# Check memory count
ayo memory list --json | jq length
```

**Solutions:**

1. **Verify memories exist:**
   ```bash
   ayo memory list
   # Should show memories
   ```

2. **Check embedding provider:**
   ```bash
   ayo doctor
   # Check "Memory" section
   ```

3. **Rebuild index:**
   ```bash
   ayo memory reindex
   ```

4. **Check query syntax:**
   ```bash
   # Semantic search, not keyword
   ayo memory search "concept I'm looking for"
   ```

#### Memory Database Corruption

**Symptom:** SQLite errors, memory operations fail

**Solutions:**

1. **Check database:**
   ```bash
   sqlite3 ~/.local/share/ayo/memory/index.db "PRAGMA integrity_check;"
   ```

2. **Backup and rebuild:**
   ```bash
   # Backup
   cp -r ~/.local/share/ayo/memory ~/.local/share/ayo/memory.bak
   
   # Rebuild from zettelkasten
   ayo memory reindex
   ```

### file_request Issues

#### Agent Can't Write to Host

**Symptom:** file_request denied or not appearing

**Diagnosis:**
```bash
# Check audit log
ayo audit list

# Check permissions config
cat ~/.config/ayo/config.json | jq '.permissions'
```

**Solutions:**

1. **Check terminal for prompt** - approval required

2. **Use --no-jodas flag:**
   ```bash
   ayo run code "task" --no-jodas
   ```

3. **Check config permissions:**
   ```json
   // ~/.config/ayo/config.json
   {
     "permissions": {
       "file_request": true,
       "auto_approve": false
     }
   }
   ```

4. **Check path restrictions:**
   ```json
   {
     "permissions": {
       "allowed_paths": ["/home/user/projects"]
     }
   }
   ```

### Trigger Issues

#### Trigger Not Firing

**Symptom:** Scheduled trigger doesn't run

**Diagnosis:**
```bash
# Check trigger exists and enabled
ayo trigger list

# Check trigger details
ayo trigger show my-trigger

# Check history
ayo trigger history my-trigger
```

**Solutions:**

1. **Verify daemon running:**
   ```bash
   ayo service status
   ```

2. **Check trigger enabled:**
   ```bash
   ayo trigger list
   # "enabled" column should be true
   ```

3. **Validate cron expression:**
   ```bash
   # Use standard cron format
   # minute hour day month weekday
   ayo trigger schedule --schedule "0 9 * * *" ...
   ```

4. **Test manually:**
   ```bash
   ayo trigger fire my-trigger
   ```

5. **Check agent exists:**
   ```bash
   ayo agent show trigger-agent
   ```

#### Watch Trigger Misses Files

**Symptom:** File changes not detected

**Solutions:**

1. **Check path configuration:**
   ```yaml
   triggers:
     - name: my-watch
       type: watch
       path: /absolute/path  # Use absolute path
       patterns: ["*.go"]
   ```

2. **Check pattern syntax:**
   ```yaml
   patterns: ["*.go", "*.md"]  # Glob patterns
   ```

3. **Check recursive setting:**
   ```yaml
   recursive: true  # Watch subdirectories
   ```

4. **Check debounce:**
   ```yaml
   debounce: "1s"  # Batch rapid changes
   ```

### Plugin Issues

#### Plugin Conflicts

**Symptom:** Strange behavior, wrong component used

**Diagnosis:**
```bash
# List all plugins
ayo plugin list

# Check for duplicates
ayo agent list
# Look for same name from different plugins
```

**Solutions:**

1. **Understand resolution order:**
   - User components (`~/.config/ayo/`)
   - Installed plugins
   - Built-in components

2. **Remove conflicting plugin:**
   ```bash
   ayo plugin remove conflicting-plugin
   ```

3. **Override with user component:**
   ```bash
   # Create user agent with same name
   mkdir -p ~/.config/ayo/agents/conflicting-name
   ```

#### Plugin Load Fails

**Symptom:** "failed to load plugin" errors

**Solutions:**

1. **Check manifest.json:**
   ```bash
   cat ~/.config/ayo/plugins/my-plugin/manifest.json | jq .
   ```

2. **Validate manifest:**
   ```bash
   ayo plugin validate ~/.config/ayo/plugins/my-plugin
   ```

3. **Check dependencies:**
   ```bash
   # Check required binaries exist
   which required-binary
   ```

4. **Reinstall plugin:**
   ```bash
   ayo plugin remove my-plugin
   ayo plugin install source-url
   ```

### Squad Issues

#### Squad Won't Start

**Symptom:** `ayo squad start` fails

**Solutions:**

1. **Check SQUAD.md exists:**
   ```bash
   cat ~/.local/share/ayo/sandboxes/squads/my-squad/SQUAD.md
   ```

2. **Check agent configurations:**
   ```bash
   ayo squad show my-squad
   ```

3. **Check sandbox provider:**
   ```bash
   ayo doctor
   ```

#### Squad RPC Errors

**Symptom:** RPC errors when communicating with squad

**Solutions:**

1. **Restart daemon:**
   ```bash
   ayo service restart
   ```

2. **Check sandbox running:**
   ```bash
   ayo squad list
   # Should show "running" status
   ```

3. **Check logs:**
   ```bash
   tail -f ~/.local/share/ayo/daemon.log
   ```

## Performance Tuning

### Slow First Response

**Cause:** Model loading, sandbox creation

**Solutions:**

1. **Use warm pool:**
   ```json
   // config.json
   {
     "sandbox": {
       "pool_size": 3
     }
   }
   ```

2. **Pre-warm on daemon start:**
   ```bash
   ayo service start
   # Pool initializes automatically
   ```

### High Memory Usage

**Cause:** Large models, many sandboxes

**Solutions:**

1. **Limit concurrent sandboxes:**
   ```json
   {
     "sandbox": {
       "max_concurrent": 5
     }
   }
   ```

2. **Use smaller models:**
   ```json
   {
     "provider": "anthropic",
     "model": "a-smaller-model"
   }
   ```

3. **Clean up old sandboxes:**
   ```bash
   ayo sandbox cleanup
   ```

### Trigger Latency

**Cause:** Polling interval, event batching

**Solutions:**

1. **Reduce poll interval:**
   ```yaml
   triggers:
     - type: interval
       every: "30s"  # More frequent
   ```

2. **Adjust debounce:**
   ```yaml
   triggers:
     - type: watch
       debounce: "500ms"  # Faster response
   ```

## Getting Help

### Gather Diagnostic Info

Before reporting issues:

```bash
# Run diagnostics
ayo doctor > doctor-output.txt 2>&1

# Get system info
./debug/system-info.sh > system-info.txt

# Get recent logs
tail -500 ~/.local/share/ayo/daemon.log > daemon-logs.txt
```

### Resources

| Resource | Link |
|----------|------|
| GitHub Issues | https://github.com/alexcabrera/ayo/issues |
| Documentation | See `docs/` directory |
| Community | *Coming soon* |

### Reporting Bugs

Include:

1. **ayo version:** `ayo --version`
2. **OS and version:** `uname -a` or `sw_vers`
3. **`ayo doctor` output**
4. **Steps to reproduce**
5. **Expected vs actual behavior**
6. **Relevant logs** (sanitize sensitive data)

## See Also

- [Architecture](architecture.md) - Understanding system internals
- [Extending Ayo](extending.md) - Custom development guide
- [CLI Reference](../reference/cli.md) - Command documentation
