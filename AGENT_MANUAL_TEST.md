# Ayo Agent Manual Testing Guide

This document provides exhaustive testing procedures for AI agents to verify all ayo functionality. Each test step is atomic and includes setup, execution, verification, and cleanup phases.

---

## Testing Protocol

For each test section:
1. **Add todos** for all steps in the section
2. **Execute** each step and collect output
3. **Analyze** output against expected results
4. **Record** any failures as new tickets
5. **Cleanup** artifacts before next section

---

## Table of Contents

1. [Environment Setup](#1-environment-setup)
2. [Pre-Flight Verification](#2-pre-flight-verification)
3. [Sandbox Lifecycle](#3-sandbox-lifecycle)
4. [Agent Execution in Sandbox](#4-agent-execution-in-sandbox)
5. [Mount System](#5-mount-system)
6. [Daemon Operations](#6-daemon-operations)
7. [Trigger System](#7-trigger-system)
8. [Sessions and Persistence](#8-sessions-and-persistence)
9. [IRC Inter-Agent Communication](#9-irc-inter-agent-communication)
10. [Backup and Sync](#10-backup-and-sync)
11. [Error Recovery](#11-error-recovery)
12. [Full System Teardown](#12-full-system-teardown)

---

## 1. Environment Setup

### 1.1 Build Fresh Binary

**Setup:**
```bash
go build -o ayo ./cmd/ayo
```

**Verification:**
- Binary exists at `./ayo`
- `./ayo --version` returns version string

**Cleanup:** None

### 1.2 Clean State Preparation

**Setup:**
```bash
# Stop any running daemon
ayo daemon stop 2>/dev/null || true

# Prune all sandboxes
ayo sandbox prune --all --force 2>/dev/null || true
```

**Verification:**
- `ayo sandbox list` shows no sandboxes
- `ayo status` shows daemon not running

**Cleanup:** None (this IS cleanup)

### 1.3 Verify Docker Available

**Setup:** None

**Execution:**
```bash
docker info >/dev/null 2>&1 && echo "Docker OK" || echo "Docker FAIL"
```

**Verification:**
- Output is "Docker OK"

**Cleanup:** None

---

## 2. Pre-Flight Verification

### 2.1 System Info Collection

**Execution:**
```bash
./debug/system-info.sh
```

**Verification:**
- OS detected correctly
- Docker shows as running
- ayo version displayed
- Config paths shown

**Analysis:** Record OS, Docker version, ayo version for test report.

### 2.2 Doctor Check

**Execution:**
```bash
ayo doctor
```

**Verification:**
- Ayo Version displayed
- Paths section shows OK or expected WARN
- Ollama section (if applicable) shows status
- Database section shows status
- Configuration section shows model settings
- Sandbox provider shown

**Analysis:**
- Note any FAIL status
- Note any unexpected WARN status
- Verify sandbox provider is correct (apple-container or linux)

### 2.3 Doctor Verbose

**Execution:**
```bash
ayo doctor -v
```

**Verification:**
- Additional sandbox execution test shown
- More detailed diagnostics displayed

---

## 3. Sandbox Lifecycle

### 3.1 List Sandboxes (Empty State)

**Precondition:** Clean state from 1.2

**Execution:**
```bash
ayo sandbox list
```

**Expected:** "No active sandboxes" or empty table

**Verification:** No running sandboxes listed

### 3.2 Create Sandbox via Agent Command

**Execution:**
```bash
ayo @ayo "run 'echo sandbox-test-marker' and report the output"
```

**Verification:**
- Agent responds successfully
- Output contains "sandbox-test-marker"
- No errors in output

**Analysis:** Record response time and any warnings.

### 3.3 Verify Sandbox Created

**Execution:**
```bash
ayo sandbox list
```

**Verification:**
- At least one sandbox listed
- Status is "running"
- Name contains "ayo"

**Analysis:** Record sandbox ID for subsequent tests.

### 3.4 Inspect Sandbox Details

**Execution:**
```bash
ayo sandbox show <sandbox-id>
```

**Verification:**
- ID matches expected
- Status is "running"
- Image shown (note which image)
- Created timestamp present
- Mounts section present

**Analysis:**
- Record image name (expected: alpine-based or busybox)
- Record mount configuration
- Note any unexpected values

### 3.5 Execute Command in Sandbox

**Execution:**
```bash
ayo sandbox exec <sandbox-id> -- uname -a
```

**Verification:**
- Command executes without error
- Output shows Linux kernel (container OS)

### 3.6 Execute Command as Specific User

**Execution:**
```bash
ayo sandbox exec <sandbox-id> -- whoami
ayo sandbox exec <sandbox-id> -- id
```

**Verification:**
- User identity shown
- UID/GID displayed

**Analysis:** Record user context for security verification.

### 3.7 Stop Sandbox

**Execution:**
```bash
ayo sandbox stop <sandbox-id>
```

**Verification:**
- Confirmation message displayed
- `ayo sandbox list` shows status as "stopped"

### 3.8 Restart Sandbox

**Execution:**
```bash
ayo sandbox start <sandbox-id>
```

**Verification:**
- Sandbox restarts successfully
- Status returns to "running"

### 3.9 Prune Stopped Sandboxes

**Setup:**
```bash
ayo sandbox stop <sandbox-id>
```

**Execution:**
```bash
ayo sandbox prune --force
```

**Verification:**
- Confirmation of removal
- `ayo sandbox list` no longer shows the sandbox

### 3.10 Prune All Sandboxes

**Execution:**
```bash
ayo sandbox prune --all --force
```

**Verification:**
- All sandboxes removed
- `ayo sandbox list` shows empty

---

## 4. Agent Execution in Sandbox

### 4.1 Verify Sandbox Execution Path

**Execution:**
```bash
ayo @ayo "run 'pwd' and tell me the working directory"
```

**Verification:**
- Path is inside container (not host path)
- Path format matches container filesystem

**Analysis:** Record working directory path.

### 4.2 File Creation in Sandbox

**Execution:**
```bash
ayo @ayo "create a file at /tmp/agent-test.txt with content 'test-content-12345'"
```

**Verification:**
- Agent confirms file creation
- No permission errors

**Secondary Verification:**
```bash
ayo sandbox exec <sandbox-id> -- cat /tmp/agent-test.txt
```
- Content matches "test-content-12345"

### 4.3 File Reading in Sandbox

**Execution:**
```bash
ayo @ayo "read the contents of /tmp/agent-test.txt"
```

**Verification:**
- Agent reports correct file contents
- No errors

### 4.4 Environment Variables

**Execution:**
```bash
ayo @ayo "run 'env' and list any AYO or USER related variables"
```

**Verification:**
- Environment variables shown
- Identify agent-specific variables

**Analysis:** Document all ayo-related environment variables.

### 4.5 Network Access from Sandbox

**Execution:**
```bash
ayo @ayo "run 'ping -c 1 8.8.8.8' or equivalent network test"
```

**Verification:**
- Network access works OR is intentionally blocked
- Document network policy

### 4.6 Process Isolation

**Execution:**
```bash
ayo @ayo "run 'ps aux' and list all processes"
```

**Verification:**
- Only container processes visible
- Host processes NOT visible

### 4.7 Filesystem Boundaries

**Execution:**
```bash
ayo @ayo "try to list /etc/shadow and report what happens"
```

**Verification:**
- Permission denied or restricted access
- Agent cannot read sensitive files

### 4.8 Resource Limits

**Execution:**
```bash
ayo sandbox exec <sandbox-id> -- cat /sys/fs/cgroup/memory/memory.limit_in_bytes 2>/dev/null || echo "cgroup v2"
ayo sandbox exec <sandbox-id> -- ulimit -a
```

**Verification:**
- Resource limits are set
- Document limit values

---

## 5. Mount System

### 5.1 List Mounts (Initial State)

**Execution:**
```bash
ayo mount list
```

**Verification:**
- Shows current mounts (may be empty)
- Format is readable

### 5.2 Add Mount (Read-Write)

**Execution:**
```bash
ayo mount add /tmp/ayo-mount-test --reason "Agent testing"
```

**Setup (if needed):**
```bash
mkdir -p /tmp/ayo-mount-test
echo "host-file-content" > /tmp/ayo-mount-test/host-file.txt
```

**Verification:**
- Mount added successfully
- `ayo mount list` shows the mount

### 5.3 Verify Mount in Sandbox

**Precondition:** Start new sandbox after mount added

**Execution:**
```bash
ayo @ayo "list files in /tmp/ayo-mount-test"
```

**Verification:**
- Agent can see host-file.txt
- Path is accessible

### 5.4 Read Mounted File from Sandbox

**Execution:**
```bash
ayo @ayo "read /tmp/ayo-mount-test/host-file.txt"
```

**Verification:**
- Content matches "host-file-content"

### 5.5 Write to Mounted Directory from Sandbox

**Execution:**
```bash
ayo @ayo "create file /tmp/ayo-mount-test/sandbox-file.txt with content 'from-sandbox'"
```

**Verification (on host):**
```bash
cat /tmp/ayo-mount-test/sandbox-file.txt
```
- File exists on host
- Content is "from-sandbox"

### 5.6 Add Mount (Read-Only)

**Execution:**
```bash
ayo mount add /tmp/ayo-mount-readonly --readonly --reason "RO test"
```

**Setup:**
```bash
mkdir -p /tmp/ayo-mount-readonly
echo "readonly-content" > /tmp/ayo-mount-readonly/ro-file.txt
```

**Verification:**
- Mount added with read-only flag

### 5.7 Verify Read-Only Enforcement

**Execution:**
```bash
ayo @ayo "try to create a file in /tmp/ayo-mount-readonly and report what happens"
```

**Verification:**
- Write fails with permission/read-only error
- Read-only is enforced

### 5.8 Remove Mount

**Execution:**
```bash
ayo mount rm /tmp/ayo-mount-test
```

**Verification:**
- Mount removed
- `ayo mount list` no longer shows it

### 5.9 Cleanup Mount Test Artifacts

**Cleanup:**
```bash
rm -rf /tmp/ayo-mount-test /tmp/ayo-mount-readonly
ayo mount rm /tmp/ayo-mount-readonly 2>/dev/null || true
```

---

## 6. Daemon Operations

### 6.1 Start Daemon

**Precondition:** Daemon not running

**Execution:**
```bash
ayo daemon start
```

**Verification:**
- Success message displayed
- Process running in background

### 6.2 Verify Daemon Status

**Execution:**
```bash
ayo status
```

**Verification:**
- CLI Version shown
- Daemon: running
- PID displayed
- Uptime shown

### 6.3 Daemon Process Verification

**Execution:**
```bash
pgrep -f "ayo daemon"
ps aux | grep "ayo daemon"
```

**Verification:**
- Process exists
- Single daemon instance running

### 6.4 Daemon Socket Verification

**Execution:**
```bash
ls -la /tmp/ayo/daemon.sock 2>/dev/null || ls -la ~/.local/share/ayo/daemon.sock 2>/dev/null
```

**Verification:**
- Socket file exists
- Permissions are correct

### 6.5 List Daemon Sessions

**Execution:**
```bash
ayo daemon sessions
```

**Verification:**
- Command executes without error
- Shows sessions or "no sessions" message

### 6.6 Daemon Logs

**Execution:**
```bash
ayo daemon logs 2>/dev/null || ./debug/daemon-status.sh --logs
```

**Verification:**
- Logs accessible
- No error messages in recent logs

### 6.7 Stop Daemon

**Execution:**
```bash
ayo daemon stop
```

**Verification:**
- Confirmation message
- `ayo status` shows daemon not running
- `pgrep -f "ayo daemon"` returns nothing

### 6.8 Daemon Restart

**Execution:**
```bash
ayo daemon start
ayo daemon restart 2>/dev/null || (ayo daemon stop && ayo daemon start)
```

**Verification:**
- Daemon restarts successfully
- New PID assigned

---

## 7. Trigger System

### 7.1 List Triggers (Empty)

**Precondition:** Daemon running

**Execution:**
```bash
ayo triggers list
```

**Verification:**
- Empty list or "no triggers" message

### 7.2 Add Cron Trigger

**Execution:**
```bash
ayo triggers add --type cron --agent @ayo --schedule "*/5 * * * *" --prompt "echo trigger-test"
```

**Verification:**
- Trigger ID returned
- Success message

**Record:** Trigger ID for subsequent tests

### 7.3 List Triggers (With Cron)

**Execution:**
```bash
ayo triggers list
```

**Verification:**
- Trigger appears in list
- Schedule shown correctly
- Agent shown correctly

### 7.4 Test Trigger Manually

**Execution:**
```bash
ayo triggers test <trigger-id>
```

**Verification:**
- Trigger fires
- Agent executes
- Output shown

### 7.5 Add Watch Trigger

**Setup:**
```bash
mkdir -p /tmp/ayo-watch-test
```

**Execution:**
```bash
ayo triggers add --type watch --agent @ayo --path /tmp/ayo-watch-test --patterns "*.txt" --prompt "file changed"
```

**Verification:**
- Trigger ID returned
- Success message

### 7.6 Test Watch Trigger

**Execution:**
```bash
touch /tmp/ayo-watch-test/test.txt
sleep 2
```

**Verification:**
- Check daemon logs for trigger fire
- Agent responds to file change

### 7.7 Disable Trigger

**Execution:**
```bash
ayo triggers disable <trigger-id>
```

**Verification:**
- Trigger disabled
- List shows disabled status

### 7.8 Enable Trigger

**Execution:**
```bash
ayo triggers enable <trigger-id>
```

**Verification:**
- Trigger enabled
- List shows enabled status

### 7.9 Remove Trigger

**Execution:**
```bash
ayo triggers rm <trigger-id>
```

**Verification:**
- Trigger removed
- No longer in list

### 7.10 Cleanup Watch Trigger Artifacts

**Cleanup:**
```bash
rm -rf /tmp/ayo-watch-test
ayo triggers list | grep -q "ayo-watch-test" && ayo triggers rm <id> || true
```

---

## 8. Sessions and Persistence

### 8.1 List Sessions (Initial)

**Execution:**
```bash
ayo sessions list
```

**Verification:**
- Shows sessions or empty message
- Format is readable

### 8.2 Create Interactive Session

**Execution:**
```bash
echo "remember the number 42" | ayo @ayo
```

**Verification:**
- Session created
- Agent responds

### 8.3 List Sessions (After Creation)

**Execution:**
```bash
ayo sessions list
```

**Verification:**
- New session appears
- Timestamp shown
- Message count shown

### 8.4 Continue Session with --latest

**Execution:**
```bash
echo "what number did I ask you to remember?" | ayo sessions continue --latest
```

**Verification:**
- Session resumed
- Agent remembers "42"

### 8.5 Show Session Details

**Execution:**
```bash
ayo sessions show --latest
```

**Verification:**
- Session ID shown
- Agent shown
- Full conversation history displayed

### 8.6 Session by ID

**Execution:**
```bash
ayo sessions show <session-id>
```

**Verification:**
- Correct session displayed
- All messages shown

### 8.7 Delete Session

**Execution:**
```bash
ayo sessions delete --latest --force
```

**Verification:**
- Session deleted
- No longer in list

### 8.8 Session Persistence Across Restarts

**Execution:**
```bash
echo "remember test-persistence" | ayo @ayo
ayo daemon stop 2>/dev/null || true
ayo daemon start 2>/dev/null || true
ayo sessions list
```

**Verification:**
- Session persists after daemon restart
- Data stored in SQLite

---

## 9. IRC Inter-Agent Communication

### 9.1 Verify IRC Server Running

**Precondition:** Sandbox running

**Execution:**
```bash
./debug/irc-status.sh 2>/dev/null || ayo sandbox exec <id> -- pgrep ngircd
```

**Verification:**
- IRC server process exists OR
- IRC not configured (document this)

### 9.2 Check IRC Configuration

**Execution:**
```bash
ayo sandbox exec <id> -- cat /etc/ngircd.conf 2>/dev/null || echo "No ngircd config"
```

**Verification:**
- Configuration exists OR
- Feature not implemented

### 9.3 Check IRC Ports

**Execution:**
```bash
ayo sandbox exec <id> -- netstat -tlnp 2>/dev/null | grep 6667 || echo "Port 6667 not listening"
```

**Verification:**
- Port 6667 listening OR
- Feature not implemented

### 9.4 Agent IRC Users

**Execution:**
```bash
ayo sandbox exec <id> -- cat /etc/passwd | grep agent-
```

**Verification:**
- Agent users exist
- User naming convention correct

---

## 10. Backup and Sync

### 10.1 Backup Status

**Execution:**
```bash
ayo backup status 2>/dev/null || echo "Backup command not available"
```

**Verification:**
- Status shown OR command not implemented

### 10.2 Initialize Backup

**Execution:**
```bash
ayo backup init 2>/dev/null || echo "Backup init not available"
```

**Verification:**
- Git repo initialized in data dir OR
- Command not implemented

### 10.3 Create Backup

**Execution:**
```bash
ayo backup create 2>/dev/null || echo "Backup create not available"
```

**Verification:**
- Backup created OR
- Command not implemented

### 10.4 List Backups

**Execution:**
```bash
ayo backup list 2>/dev/null || echo "Backup list not available"
```

**Verification:**
- Backups listed OR
- Command not implemented

### 10.5 Sync Status

**Execution:**
```bash
ayo sync status 2>/dev/null || echo "Sync not available"
```

**Verification:**
- Sync status shown OR
- Command not implemented

---

## 11. Error Recovery

### 11.1 Stale Socket Recovery

**Setup:**
```bash
ayo daemon stop
# Leave socket file if exists
```

**Execution:**
```bash
ayo daemon start
```

**Verification:**
- Daemon handles stale socket
- Starts successfully

### 11.2 Orphan Container Recovery

**Setup:**
```bash
# Create orphan container state (if possible)
```

**Execution:**
```bash
ayo sandbox list
ayo sandbox prune --all --force
```

**Verification:**
- Orphans detected
- Cleanup successful

### 11.3 Database Recovery

**Execution:**
```bash
ayo doctor
```

**Verification:**
- Database status shown
- Recovery path documented if issues found

---

## 12. Full System Teardown

### 12.1 Stop All Services

**Execution:**
```bash
ayo daemon stop 2>/dev/null || true
```

**Verification:**
- Daemon stopped

### 12.2 Remove All Sandboxes

**Execution:**
```bash
ayo sandbox prune --all --force
```

**Verification:**
- All sandboxes removed
- `ayo sandbox list` empty

### 12.3 Verify Clean State

**Execution:**
```bash
ayo sandbox list
ayo status
ayo triggers list 2>/dev/null || true
```

**Verification:**
- No sandboxes
- Daemon not running
- No active triggers

### 12.4 Remove Test Artifacts

**Execution:**
```bash
rm -rf /tmp/ayo-mount-test /tmp/ayo-mount-readonly /tmp/ayo-watch-test
rm -f /tmp/ayo-test-*
```

**Verification:**
- Temp files removed
- Clean filesystem state

---

## Test Summary Template

After all tests complete, generate:

```
## Test Results

**Date:** YYYY-MM-DD HH:MM
**Ayo Version:** X.X.X
**Platform:** darwin/linux arm64/amd64

### Section Results

| Section | Tests | Passed | Failed | Skipped |
|---------|-------|--------|--------|---------|
| 1. Environment Setup | X | X | X | X |
| 2. Pre-Flight | X | X | X | X |
| ... | ... | ... | ... | ... |

### Failed Tests

1. Test X.Y: Description
   - Expected: ...
   - Actual: ...
   - Ticket: #XXX

### Issues Created

- Ticket #XXX: Description
- Ticket #YYY: Description

### Notes

- Any observations or recommendations
```
