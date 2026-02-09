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
6. [Service Operations](#6-service-operations)
7. [Trigger System](#7-trigger-system)
8. [Sessions and Persistence](#8-sessions-and-persistence)
9. [Matrix Inter-Agent Communication](#9-matrix-inter-agent-communication)
10. [Backup and Sync](#10-backup-and-sync)
11. [Error Recovery](#11-error-recovery)
12. [Full System Teardown](#12-full-system-teardown)
13. [Flow System](#13-flow-system)

---

## 1. Environment Setup

### 1.1 Build Fresh Binary

**Setup:**
```bash
go build -o .local/bin/ayo ./cmd/ayo/...
```

**Verification:**
- Binary exists at `./.local/bin/ayo`
- `./.local/bin/ayo --version` returns version string

**Cleanup:** None

### 1.2 Clean State Preparation

**Setup:**
```bash
# Stop any running service
ayo sandbox service stop 2>/dev/null || true

# Prune all sandboxes
ayo sandbox prune --all --force 2>/dev/null || true
```

**Verification:**
- `ayo sandbox list` shows no sandboxes
- `ayo status` shows service not running

**Cleanup:** None (this IS cleanup)

### 1.3 Verify Container Runtime Available

**Setup:** None

**Execution:**
```bash
# On macOS with Apple Container
container --version >/dev/null 2>&1 && echo "Apple Container OK" || echo "Apple Container FAIL"

# On Linux (if applicable)
# systemctl status systemd-nspawn >/dev/null 2>&1 && echo "systemd-nspawn OK" || echo "systemd-nspawn FAIL"
```

**Verification:**
- Output is "Apple Container OK" (on macOS)

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
- Container runtime shows as available
- ayo version displayed
- Config paths shown

**Analysis:** Record OS, container runtime version, ayo version for test report.

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
- Verify sandbox provider is correct (apple-container or systemd-nspawn)

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

**Note:** All sandbox commands accept an optional `--id <id>` flag. If omitted:
- With 1 sandbox: auto-selects it
- With multiple: shows interactive picker

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
ayo sandbox show
# Or with explicit ID:
ayo sandbox show --id <sandbox-id>
```

**Verification:**
- ID matches expected
- Status is "running"
- Image shown (note which image)
- Created timestamp present
- Mounts section present

**Analysis:**
- Record image name (expected: Alpine-based)
- Record mount configuration
- Note any unexpected values

### 3.5 Execute Command in Sandbox

**Execution:**
```bash
ayo sandbox exec uname -a
# Or with explicit ID:
ayo sandbox exec --id <sandbox-id> uname -a
```

**Verification:**
- Command executes without error
- Output shows Linux kernel (container OS)

### 3.6 Execute Command as Specific User

**Execution:**
```bash
ayo sandbox exec whoami
ayo sandbox exec --user ayo whoami
ayo sandbox exec id
```

**Verification:**
- User identity shown
- UID/GID displayed

**Analysis:** Record user context for security verification.

### 3.7 Stop Sandbox

**Execution:**
```bash
ayo sandbox stop
# Or with explicit ID:
ayo sandbox stop --id <sandbox-id>
```

**Verification:**
- Confirmation message displayed
- `ayo sandbox list` shows status as "stopped"

### 3.8 Restart Sandbox

**Execution:**
```bash
ayo sandbox start --id <sandbox-id>
```

**Verification:**
- Sandbox restarts successfully
- Status returns to "running"

### 3.9 Prune Stopped Sandboxes

**Setup:**
```bash
ayo sandbox stop
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
ayo sandbox exec cat /tmp/agent-test.txt
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
ayo sandbox exec cat /sys/fs/cgroup/memory/memory.limit_in_bytes 2>/dev/null || echo "cgroup v2"
ayo sandbox exec ulimit -a
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

**Setup (if needed):**
```bash
mkdir -p /tmp/ayo-mount-test
echo "host-file-content" > /tmp/ayo-mount-test/host-file.txt
```

**Execution:**
```bash
ayo mount add /tmp/ayo-mount-test
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

**Setup:**
```bash
mkdir -p /tmp/ayo-mount-readonly
echo "readonly-content" > /tmp/ayo-mount-readonly/ro-file.txt
```

**Execution:**
```bash
ayo mount add /tmp/ayo-mount-readonly --ro
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

## 6. Service Operations

### 6.1 Start Service

**Precondition:** Service not running

**Execution:**
```bash
ayo sandbox service start
```

**Verification:**
- Success message displayed
- Process running in background

### 6.2 Verify Service Status

**Execution:**
```bash
ayo sandbox service status
# Or:
ayo status
```

**Verification:**
- CLI Version shown
- Service: running
- PID displayed
- Uptime shown

### 6.3 Service Process Verification

**Execution:**
```bash
pgrep -f "ayo"
ps aux | grep "ayo"
```

**Verification:**
- Process exists
- Single service instance running

### 6.4 Service Socket Verification

**Execution:**
```bash
ls -la /tmp/ayo/daemon.sock 2>/dev/null || ls -la ~/.local/share/ayo/daemon.sock 2>/dev/null
```

**Verification:**
- Socket file exists
- Permissions are correct

### 6.5 Stop Service

**Execution:**
```bash
ayo sandbox service stop
```

**Verification:**
- Confirmation message
- `ayo status` shows service not running
- `pgrep -f "ayo"` returns nothing (or only test process)

### 6.6 Service Restart

**Execution:**
```bash
ayo sandbox service start
ayo sandbox service stop && ayo sandbox service start
```

**Verification:**
- Service restarts successfully
- New PID assigned

---

## 7. Trigger System

### 7.1 List Triggers (Empty)

**Precondition:** Service running

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
ayo sandbox service stop 2>/dev/null || true
ayo sandbox service start 2>/dev/null || true
ayo sessions list
```

**Verification:**
- Session persists after service restart
- Data stored in SQLite

---

## 9. Matrix Inter-Agent Communication

### 9.1 Verify Matrix/Conduit Running

**Precondition:** Sandbox service running

**Execution:**
```bash
ayo matrix status
ayo matrix status --json
```

**Verification:**
- Conduit server status shown (running/stopped)
- Broker connection status shown
- JSON output format valid

### 9.2 Room Management

**Execution:**
```bash
# List existing rooms
ayo matrix rooms

# Create a test room
ayo matrix create test-agent-room

# List rooms again
ayo matrix rooms
```

**Verification:**
- Initial room list displayed (may be empty)
- Room creation succeeds with confirmation
- New room appears in list

### 9.3 Messaging

**Execution:**
```bash
# Send a message
ayo matrix send test-agent-room "Test message from agent"

# Read messages
ayo matrix read test-agent-room

# Read with limit
ayo matrix read test-agent-room 5
```

**Verification:**
- Message sends successfully
- Read shows the sent message
- Limit parameter works correctly

### 9.4 Room Membership

**Execution:**
```bash
# Show room members
ayo matrix who test-agent-room

# Invite another agent (if available)
ayo matrix invite test-agent-room @ayo
```

**Verification:**
- Current members listed
- Invite works or shows appropriate error

### 9.5 Matrix from Inside Sandbox

**Execution:**
```bash
# Login to sandbox
ayo sandbox login

# Inside sandbox, verify socket mount exists
ls -la /run/ayo/

# Test matrix commands work from sandbox
ayo matrix status
ayo matrix rooms

# Exit sandbox
exit
```

**Verification:**
- /run/ayo/ directory exists with sockets
- Matrix commands work from inside sandbox
- Socket communication successful

### 9.6 Cleanup Matrix Test Artifacts

**Cleanup:**
```bash
# No persistent cleanup needed - rooms persist for future tests
# Or manually delete room if cleanup desired
```

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
ayo sandbox service stop
# Leave socket file if exists
```

**Execution:**
```bash
ayo sandbox service start
```

**Verification:**
- Service handles stale socket
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
ayo sandbox service stop 2>/dev/null || true
```

**Verification:**
- Service stopped

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
- Service not running
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

## 13. Flow System

### 13.1 List Flows (Initial State)

**Execution:**
```bash
ayo flows list
ayo flows list --json
```

**Verification:**
- Shows available flows (may be empty)
- JSON output format valid

### 13.2 Create Shell Flow

**Execution:**
```bash
ayo flows new test-agent-flow
cat ~/.config/ayo/flows/test-agent-flow.sh
```

**Verification:**
- Flow file created at expected path
- Has ayo:flow frontmatter header
- Has shebang and set -euo pipefail

### 13.3 Run Shell Flow

**Execution:**
```bash
ayo flows run test-agent-flow '{"message": "hello from agent"}'
```

**Verification:**
- Flow executes successfully
- Output is JSON format
- No errors in execution

### 13.4 Create YAML Flow

**Setup:**
```bash
cat > /tmp/test-yaml-flow.yaml << 'EOF'
version: 1
name: test-yaml-flow
description: Test multi-step flow

steps:
  - id: step1
    type: shell
    run: echo "Hello from step 1"

  - id: step2
    type: shell
    run: echo "Step 2 received: {{ steps.step1.stdout }}"
    depends_on: [step1]

  - id: step3
    type: shell
    run: echo "Final output"
    depends_on: [step2]
EOF
```

**Verification:**
- YAML file created with valid syntax
- All steps defined correctly

### 13.5 Validate YAML Flow

**Execution:**
```bash
ayo flows validate /tmp/test-yaml-flow.yaml
```

**Verification:**
- Shows validation result (valid or errors)
- Identifies any structural issues

### 13.6 Flow History

**Execution:**
```bash
ayo flows history
ayo flows history --flow test-agent-flow
```

**Verification:**
- Shows run history
- Filter by flow name works

### 13.7 Flow Stats

**Execution:**
```bash
ayo flows stats
ayo flows stats test-agent-flow
```

**Verification:**
- Shows execution statistics
- Per-flow stats work correctly

### 13.8 Flow Triggers

**Execution:**
```bash
# List flows that have triggers defined
ayo flows list --with-triggers 2>/dev/null || ayo flows list
```

**Verification:**
- Flows with triggers identified OR
- Feature shows in listing

### 13.9 Remove Flow

**Execution:**
```bash
ayo flows rm test-agent-flow --force
ayo flows list
```

**Verification:**
- Flow removed from list
- No longer appears

### 13.10 Cleanup Flow Artifacts

**Cleanup:**
```bash
rm -f /tmp/test-yaml-flow.yaml
ayo flows rm test-agent-flow 2>/dev/null || true
```

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
