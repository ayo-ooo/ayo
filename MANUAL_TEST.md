# Ayo Sandbox Manual Testing Guide

This guide walks through comprehensive manual testing of the ayo sandbox-first architecture. Each section includes commands to run, expected outcomes, common failure modes, and diagnostic commands to gather debug information.

**Before You Begin:**
1. Ensure Docker is running
2. Build ayo from source: `go build -o ayo ./cmd/ayo`
3. Add to PATH or use `./ayo` for commands below

---

## Table of Contents

1. [Pre-Flight Checks](#1-pre-flight-checks)
2. [Sandbox Lifecycle](#2-sandbox-lifecycle)
3. [Agent Execution in Sandbox](#3-agent-execution-in-sandbox)
4. [Mount System](#4-mount-system)
5. [Daemon and Background Agents](#5-daemon-and-background-agents)
6. [Trigger System](#6-trigger-system)
7. [Sessions and Persistence](#7-sessions-and-persistence)
8. [IRC Inter-Agent Communication](#8-irc-inter-agent-communication)
9. [Sync and Backup](#9-sync-and-backup)
10. [Error Recovery](#10-error-recovery)

---

## 1. Pre-Flight Checks

### 1.1 Verify System Requirements

**Command:**
```bash
./debug/system-info.sh
```

**Expected Output:**
- OS: macOS or Linux
- Docker: Installed and running
- ayo: Version displayed, config/data paths shown

**Possible Failures:**
| Symptom | Cause | Fix |
|---------|-------|-----|
| Docker not running | Docker daemon stopped | Start Docker Desktop |
| ayo command not found | Not in PATH | Use `./ayo` or add to PATH |
| Config dir missing | First run | Run `ayo setup` |

**Diagnostic Command (paste output back):**
```bash
./debug/system-info.sh --json | pbcopy
```

### 1.2 Verify Ayo Doctor

**Command:**
```bash
ayo doctor
```

**Expected Output:**
- All checks should show OK (green)
- Database accessible
- Config file valid

**Possible Failures:**
| Symptom | Cause | Fix |
|---------|-------|-----|
| Database error | Corrupted SQLite | Delete `~/.local/share/ayo/ayo.db` |
| No providers configured | Missing API keys | Run `ayo setup` |
| Ollama not available | Ollama not running | Start Ollama (optional) |

**Diagnostic Command:**
```bash
(ayo doctor; echo "---"; ./debug/system-info.sh) | pbcopy
```

---

## 2. Sandbox Lifecycle

### 2.1 List Sandboxes (Empty State)

**Command:**
```bash
ayo sandbox list
```

**Expected Output:**
```
No active sandboxes
```

**Diagnostic Command:**
```bash
./debug/sandbox-status.sh | pbcopy
```

### 2.2 Create Sandbox by Starting Agent

**Command:**
```bash
ayo @ayo "echo hello from sandbox"
```

**Expected Output:**
- Agent responds with acknowledgment
- No errors about sandbox creation
- Response mentions "hello from sandbox" context

**What's Being Tested:**
- Sandbox container creation
- Alpine image pull (first run)
- Agent user creation
- Bash tool execution inside container

**Possible Failures:**
| Symptom | Cause | Fix |
|---------|-------|-----|
| Image pull failed | Network issue | Check internet, retry |
| Container start failed | Docker resources | `docker system prune` |
| Permission denied | Docker socket | Check Docker permissions |
| Timeout | Slow image pull | Wait, increase timeout |

**Diagnostic Command:**
```bash
(./debug/sandbox-status.sh --verbose; echo "---"; docker logs $(docker ps -q --filter name=ayo-sandbox) 2>&1 | tail -50) | pbcopy
```

### 2.3 Verify Sandbox Exists

**Command:**
```bash
ayo sandbox list
```

**Expected Output:**
- At least one sandbox listed
- Status shows "running"
- Name contains "ayo-sandbox"

**Diagnostic Command:**
```bash
ayo sandbox list --json | pbcopy
```

### 2.4 Inspect Sandbox Details

**Command:**
```bash
ayo sandbox show <id>
```
*(Replace `<id>` with ID from sandbox list)*

**Expected Output:**
- Container ID and name
- Status: running
- Image: shows Alpine-based image
- Mounts: listed if any
- Created timestamp

**Diagnostic Command:**
```bash
(ayo sandbox list --json; echo "---"; docker inspect $(docker ps -q --filter name=ayo-sandbox) 2>&1) | pbcopy
```

### 2.5 Execute Command in Sandbox

**Command:**
```bash
ayo sandbox exec <id> cat /etc/os-release
```

**Expected Output:**
```
NAME="Alpine Linux"
ID=alpine
VERSION_ID=...
```

**What's Being Tested:**
- Direct container command execution
- Container responsiveness

**Diagnostic Command:**
```bash
./debug/sandbox-exec.sh cat /etc/os-release | pbcopy
```

### 2.6 Stop Sandbox

**Command:**
```bash
ayo sandbox stop <id>
```

**Expected Output:**
- Confirmation message
- Sandbox stops (verify with `ayo sandbox list`)

**Diagnostic Command:**
```bash
(ayo sandbox list; echo "---"; docker ps -a --filter name=ayo-sandbox) | pbcopy
```

### 2.7 Prune Stopped Sandboxes

**Command:**
```bash
ayo sandbox prune
```

**Expected Output:**
- Confirmation or count of removed containers

**Diagnostic Command:**
```bash
docker ps -a --filter name=ayo-sandbox | pbcopy
```

---

## 3. Agent Execution in Sandbox

### 3.1 Simple Bash Command

**Command:**
```bash
ayo @ayo "run 'pwd' and tell me the current directory"
```

**Expected Output:**
- Agent executes `pwd` in sandbox
- Reports path like `/home/agent-ayo` or `/workspace/...`
- NOT a host path

**What's Being Tested:**
- Bash tool routes to sandbox
- Working directory is inside container

**Possible Failures:**
| Symptom | Cause | Fix |
|---------|-------|-----|
| Shows host path | Sandbox not enabled | Check agent config |
| Command not found | Missing in Alpine | Install via apk |
| Permission denied | User permissions | Check agent user creation |

**Diagnostic Command:**
```bash
(./debug/sandbox-exec.sh --as ayo pwd; ./debug/sandbox-exec.sh --as ayo whoami) | pbcopy
```

### 3.2 File Creation in Sandbox

**Command:**
```bash
ayo @ayo "create a file called test.txt with 'hello world' in your home directory"
```

**Expected Output:**
- Agent confirms file creation
- No errors

**Verify:**
```bash
./debug/sandbox-exec.sh --as ayo cat ~/test.txt
```

Should output: `hello world`

**Diagnostic Command:**
```bash
./debug/sandbox-exec.sh --as ayo "ls -la ~ && cat ~/test.txt 2>/dev/null" | pbcopy
```

### 3.3 Install Package (Alpine apk)

**Command:**
```bash
ayo @ayo "install jq using apk and then run 'jq --version'"
```

**Expected Output:**
- Agent runs `apk add jq`
- Shows jq version output

**What's Being Tested:**
- Package installation works
- Agent has necessary permissions

**Possible Failures:**
| Symptom | Cause | Fix |
|---------|-------|-----|
| Permission denied on apk | Agent not in sudoers | Check user setup |
| Network unreachable | Container network | Check Docker network |
| Package not found | Wrong package name | Use correct Alpine pkg name |

**Diagnostic Command:**
```bash
./debug/sandbox-exec.sh "apk info -v 2>&1; cat /etc/apk/repositories" | pbcopy
```

### 3.4 Verify Agent Isolation

**Command:**
```bash
ayo @ayo "try to read /etc/shadow"
```

**Expected Output:**
- Permission denied or similar error
- Agent should NOT be able to read sensitive files as regular user

**What's Being Tested:**
- Privilege isolation works
- Agent runs as non-root

**Diagnostic Command:**
```bash
./debug/sandbox-exec.sh --as ayo "id; cat /etc/shadow 2>&1" | pbcopy
```

---

## 4. Mount System

### 4.1 List Current Mounts

**Command:**
```bash
ayo mount list
```

**Expected Output:**
- List of mounted paths (or empty if none)
- Shows permissions (rw/ro)

**Diagnostic Command:**
```bash
./debug/mount-check.sh | pbcopy
```

### 4.2 Add a Mount

**Command:**
```bash
ayo mount add $(pwd) --reason "Testing mount system"
```

**Expected Output:**
- Confirmation message
- Mount added to mounts.json

**Verify:**
```bash
ayo mount list
cat ~/.local/share/ayo/mounts.json
```

**Diagnostic Command:**
```bash
(ayo mount list; echo "---"; cat ~/.local/share/ayo/mounts.json 2>/dev/null) | pbcopy
```

### 4.3 Access Mounted Directory from Agent

**Command:**
```bash
ayo @ayo "list files in the current project directory"
```

**Expected Output:**
- Agent can see files from the mounted host directory
- Lists actual project files

**What's Being Tested:**
- Mount propagates to sandbox
- Agent can read mounted files

**Possible Failures:**
| Symptom | Cause | Fix |
|---------|-------|-----|
| Cannot access path | Mount not active | Restart sandbox |
| Empty directory | Mount path wrong | Check mount config |
| Permission denied | Mount permissions | Check ro/rw setting |

**Diagnostic Command:**
```bash
./debug/mount-check.sh --test-write | pbcopy
```

### 4.4 Write to Mounted Directory

**Command:**
```bash
ayo @ayo "create a file called .ayo-test-file in the project root with content 'test'"
```

**Expected Output:**
- Agent confirms creation
- File appears on host filesystem

**Verify (on host):**
```bash
cat .ayo-test-file
rm .ayo-test-file
```

**Diagnostic Command:**
```bash
(ls -la .ayo-test-file 2>&1; ./debug/mount-check.sh) | pbcopy
```

### 4.5 Remove Mount

**Command:**
```bash
ayo mount rm $(pwd)
```

**Expected Output:**
- Confirmation message
- Mount removed from list

**Diagnostic Command:**
```bash
(ayo mount list; cat ~/.local/share/ayo/mounts.json 2>/dev/null) | pbcopy
```

---

## 5. Daemon and Background Agents

### 5.1 Start Daemon

**Command:**
```bash
ayo daemon start
```

**Expected Output:**
- "Daemon started" message or similar
- Process runs in background

**Verify:**
```bash
ayo status
```

**Possible Failures:**
| Symptom | Cause | Fix |
|---------|-------|-----|
| Address already in use | Daemon already running | `ayo daemon stop` first |
| Permission denied | Socket path issue | Check XDG_RUNTIME_DIR |
| Crash immediately | Check logs | `ayo daemon logs` |

**Diagnostic Command:**
```bash
./debug/daemon-status.sh --logs | pbcopy
```

### 5.2 Check Daemon Status

**Command:**
```bash
ayo status
```

**Expected Output:**
- CLI Version shown
- Daemon: running
- PID, Version, Uptime displayed
- Sandbox Pool status

**Diagnostic Command:**
```bash
ayo status --json | pbcopy
```

### 5.3 List Daemon Sessions

**Command:**
```bash
ayo daemon sessions
```

**Expected Output:**
- List of active/recent sessions managed by daemon
- Or "no sessions" message

**Diagnostic Command:**
```bash
./debug/daemon-status.sh | pbcopy
```

### 5.4 Stop Daemon

**Command:**
```bash
ayo daemon stop
```

**Expected Output:**
- Confirmation message
- Daemon process terminates

**Verify:**
```bash
ayo status
```
Should show daemon not running.

**Diagnostic Command:**
```bash
(ayo status 2>&1; pgrep -f "ayo daemon" 2>&1) | pbcopy
```

---

## 6. Trigger System

### 6.1 List Triggers (Empty)

**Command:**
```bash
ayo triggers list
```

**Expected Output:**
- "No triggers registered" or empty list

**Diagnostic Command:**
```bash
ayo triggers list --json | pbcopy
```

### 6.2 Add Cron Trigger

**Prerequisites:** Daemon must be running.

**Command:**
```bash
ayo triggers add --type cron --agent @ayo --schedule "*/5 * * * *" --prompt "say hello"
```

**Expected Output:**
- Trigger ID returned
- Confirmation message

**Verify:**
```bash
ayo triggers list
```

**Possible Failures:**
| Symptom | Cause | Fix |
|---------|-------|-----|
| Daemon not running | Need daemon | `ayo daemon start` |
| Invalid cron syntax | Bad schedule format | Use valid cron expression |
| Agent not found | Wrong handle | Check `ayo agents list` |

**Diagnostic Command:**
```bash
(ayo triggers list --json; ./debug/daemon-status.sh) | pbcopy
```

### 6.3 Test Trigger

**Command:**
```bash
ayo triggers test <trigger-id>
```

**Expected Output:**
- Trigger fires immediately
- Agent executes the prompt
- Output shown

**Diagnostic Command:**
```bash
(ayo triggers test <id> 2>&1; ./debug/daemon-status.sh --logs) | pbcopy
```

### 6.4 Add Watch Trigger

**Command:**
```bash
ayo triggers add --type watch --agent @ayo --path $(pwd) --patterns "*.txt" --prompt "a file changed"
```

**Expected Output:**
- Trigger ID returned

**Test by creating a file:**
```bash
touch test-watch.txt
```

**Verify:** Agent should respond to the file change (check daemon logs).

**Diagnostic Command:**
```bash
(ayo triggers list --json; touch test-trigger.txt; sleep 2; ./debug/daemon-status.sh --logs) | pbcopy
```

### 6.5 Remove Trigger

**Command:**
```bash
ayo triggers rm <trigger-id>
```

**Expected Output:**
- Confirmation message
- Trigger removed from list

**Diagnostic Command:**
```bash
ayo triggers list | pbcopy
```

---

## 7. Sessions and Persistence

### 7.1 Start Interactive Session

**Command:**
```bash
ayo @ayo
```

Then type: "remember that my favorite color is blue" and exit with Ctrl+D.

**Expected Output:**
- Interactive chat starts
- Agent responds
- Session saved

**Diagnostic Command:**
```bash
ayo sessions list --json | head -20 | pbcopy
```

### 7.2 List Sessions

**Command:**
```bash
ayo sessions list
```

**Expected Output:**
- Recent sessions listed
- Shows agent, title, message count, time

**Diagnostic Command:**
```bash
ayo sessions list --json | pbcopy
```

### 7.3 Continue Session with --latest

**Command:**
```bash
ayo sessions continue --latest
```

Then ask: "what is my favorite color?"

**Expected Output:**
- Resumes most recent session
- Agent should remember "blue" from previous message

**What's Being Tested:**
- Session persistence
- --latest flag functionality
- Memory/context carry-over

**Diagnostic Command:**
```bash
(ayo sessions list -n 5 --json; ayo sessions show --latest --json 2>/dev/null) | pbcopy
```

### 7.4 Show Session Details

**Command:**
```bash
ayo sessions show --latest
```

**Expected Output:**
- Session metadata (ID, agent, title)
- Full conversation history

**Diagnostic Command:**
```bash
ayo sessions show --latest | pbcopy
```

### 7.5 Delete Session

**Command:**
```bash
ayo sessions delete --latest --force
```

**Expected Output:**
- Confirmation message
- Session removed

**Diagnostic Command:**
```bash
ayo sessions list | pbcopy
```

---

## 8. IRC Inter-Agent Communication

### 8.1 Check IRC Server Status

**Prerequisites:** Sandbox must be running.

**Command:**
```bash
./debug/irc-status.sh
```

**Expected Output:**
- ngircd: Running
- Configuration shown
- Ports 6667/6697 listening

**Possible Failures:**
| Symptom | Cause | Fix |
|---------|-------|-----|
| ngircd not running | Not started | Image issue, rebuild |
| Port not listening | Config error | Check ngircd.conf |
| No log file | Logging disabled | Check config |

**Diagnostic Command:**
```bash
./debug/irc-status.sh --messages 50 | pbcopy
```

### 8.2 View IRC Messages

**Command:**
```bash
ayo messages
```

**Expected Output:**
- IRC log viewer interface
- Or recent messages listed

**Diagnostic Command:**
```bash
./debug/sandbox-exec.sh "cat /var/log/ngircd.log 2>/dev/null || echo 'No log'" | pbcopy
```

### 8.3 Check Agent IRC Users

**Command:**
```bash
./debug/sandbox-exec.sh cat /etc/passwd | grep agent-
```

**Expected Output:**
- Agent users listed (e.g., agent-ayo:x:1000:1000:...)

**Diagnostic Command:**
```bash
./debug/sandbox-exec.sh "cat /etc/passwd | grep agent-; ls -la /home/" | pbcopy
```

---

## 9. Sync and Backup

### 9.1 Initialize Backup

**Command:**
```bash
ayo backup init
```

**Expected Output:**
- Git repository initialized in data directory
- Or "already initialized" message

**Diagnostic Command:**
```bash
(ayo backup status 2>&1; ls -la ~/.local/share/ayo/.git 2>&1) | pbcopy
```

### 9.2 Create Backup

**Command:**
```bash
ayo backup create
```

**Expected Output:**
- Backup created message
- Git commit made

**Diagnostic Command:**
```bash
(ayo backup list 2>&1; git -C ~/.local/share/ayo log --oneline -5 2>&1) | pbcopy
```

### 9.3 List Backups

**Command:**
```bash
ayo backup list
```

**Expected Output:**
- List of backup commits with dates

**Diagnostic Command:**
```bash
ayo backup list --json | pbcopy
```

### 9.4 Sync Status

**Command:**
```bash
ayo sync status
```

**Expected Output:**
- Current sync state
- Remote configuration (if any)
- Pending changes

**Diagnostic Command:**
```bash
(ayo sync status 2>&1; git -C ~/.local/share/ayo remote -v 2>&1) | pbcopy
```

---

## 10. Error Recovery

### 10.1 Sandbox Won't Start

**Symptoms:** Container fails to create or start.

**Diagnostic Commands:**
```bash
# Full system check
./debug/collect-all.sh | pbcopy
```

**Recovery Steps:**
1. Prune old containers: `docker system prune`
2. Remove ayo images: `docker rmi $(docker images -q '*ayo*')`
3. Restart Docker
4. Try again

### 10.2 Daemon Won't Connect

**Symptoms:** `ayo status` shows daemon running but cannot connect.

**Diagnostic Commands:**
```bash
./debug/daemon-status.sh --logs | pbcopy
```

**Recovery Steps:**
1. Kill daemon: `pkill -f 'ayo daemon'`
2. Remove socket: `rm -f /tmp/ayo/daemon.sock`
3. Restart: `ayo daemon start`

### 10.3 Agent Can't Access Mounted Files

**Symptoms:** Agent reports file not found for host files.

**Diagnostic Commands:**
```bash
./debug/mount-check.sh --test-write | pbcopy
```

**Recovery Steps:**
1. Verify mount: `ayo mount list`
2. Stop sandbox: `ayo sandbox stop <id>`
3. Re-add mount: `ayo mount add <path>`
4. Start new session

### 10.4 Complete Reset

**⚠️ CAUTION: This will delete all ayo data!**

```bash
# Stop everything
ayo daemon stop
ayo sandbox prune

# Remove all containers
docker rm -f $(docker ps -aq --filter name=ayo-sandbox)

# Remove data (DESTRUCTIVE)
rm -rf ~/.local/share/ayo
rm -rf ~/.config/ayo

# Reinstall
ayo setup
```

---

## Diagnostic Script Reference

| Script | Purpose | When to Use |
|--------|---------|-------------|
| `./debug/system-info.sh` | Host system info | First step in any debugging |
| `./debug/sandbox-status.sh` | Container status | Sandbox issues |
| `./debug/daemon-status.sh` | Daemon health | Background agent issues |
| `./debug/sandbox-exec.sh` | Run commands in sandbox | Test container access |
| `./debug/irc-status.sh` | IRC server health | Inter-agent communication |
| `./debug/mount-check.sh` | Mount verification | File access issues |
| `./debug/collect-all.sh` | Everything | Bug reports |

All scripts support `--json` for machine-readable output.

---

## Reporting Issues

When reporting issues, always include:

1. **Full diagnostic report:**
   ```bash
   ./debug/collect-all.sh | pbcopy
   ```

2. **Exact command that failed**

3. **Expected vs actual behavior**

4. **Steps to reproduce**

Paste the clipboard contents into the issue or chat for analysis.
