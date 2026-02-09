# Ayo Human Manual Testing Guide

Complete manual testing checklist for ayo sandbox-first architecture.

---

## Prerequisites

```bash
# Build fresh binary
go build -o .local/bin/ayo ./cmd/ayo/...

# Verify version
./.local/bin/ayo --version
```

**Required:**
- macOS 26+ on Apple Silicon (for Apple Container)
- OR Linux with systemd (for systemd-nspawn)

---

## 1. Pre-Flight Checks

### 1.1 System Info
```bash
./debug/system-info.sh
```
- [ ] Platform detected correctly
- [ ] Container runtime available

### 1.2 Doctor Check
```bash
./ayo doctor
```
- [ ] Ayo Version shows
- [ ] Platform: darwin/arm64 (or linux/amd64)
- [ ] Sandbox Provider: apple or systemd-nspawn
- [ ] Database: OK

---

## 2. Service Operations

### 2.1 Start Service
```bash
./ayo sandbox service start
./ayo sandbox service status
```
- [ ] Service shows "running"
- [ ] PID, Uptime visible
- [ ] Sandbox Pool shows counts

### 2.2 Stop Service
```bash
./ayo sandbox service stop
./ayo sandbox service status
```
- [ ] Service shows "not running"

### 2.3 Restart for remaining tests
```bash
./ayo sandbox service start
```

---

## 3. Sandbox Lifecycle

**Note:** All sandbox commands accept an optional `--id <id>` flag. If omitted:
- With 1 sandbox: auto-selects it
- With multiple: shows interactive picker

### 3.1 List Sandboxes
```bash
./ayo sandbox list
```
- [ ] Shows table of sandboxes (may be empty initially)
- [ ] JSON output works: `./ayo sandbox list --json`

### 3.2 Show Sandbox Details
```bash
# Auto-select or picker
./ayo sandbox show
# Or with explicit ID
./ayo sandbox show --id <id>
```
- [ ] Shows ID, Name, Status, Created, Agents

### 3.3 Execute Command in Sandbox
```bash
./ayo sandbox exec cat /etc/os-release
```
- [ ] Shows Alpine Linux info (not host OS)
- [ ] Proves execution is sandboxed

### 3.4 Interactive Shell
```bash
./ayo sandbox login
```
- [ ] Opens shell inside container
- [ ] `whoami` shows root
- [ ] `exit` returns to host
- [ ] Ctrl+C works to escape

### 3.5 Start/Stop Sandbox
```bash
./ayo sandbox stop
./ayo sandbox list  # Should show stopped
./ayo sandbox start --id <id>
./ayo sandbox list  # Should show running
```
- [ ] Stop changes status
- [ ] Start restores status

### 3.6 Resource Stats
```bash
./ayo sandbox stats
```
- [ ] Shows CPU, Memory, PIDs
- [ ] JSON output works: `./ayo sandbox stats --json`

### 3.7 Prune Sandboxes
```bash
./ayo sandbox prune              # Prune stopped sandboxes (prompts for confirmation)
./ayo sandbox prune --force      # Skip confirmation
./ayo sandbox prune --all        # Also remove running sandboxes
./ayo sandbox prune --homes      # Also clean agent home dirs
```
- [ ] Prune removes stopped sandboxes
- [ ] --all removes running sandboxes too
- [ ] --homes cleans ~/.local/share/ayo/homes/

---

## 4. File Transfer

### 4.1 Push File to Sandbox
```bash
echo "test content" > /tmp/testfile.txt
./ayo sandbox push /tmp/testfile.txt /tmp/testfile.txt
./ayo sandbox exec cat /tmp/testfile.txt
```
- [ ] File appears in sandbox
- [ ] Content matches

### 4.2 Pull File from Sandbox
```bash
./ayo sandbox exec sh -c "echo 'from sandbox' > /tmp/sandbox-file.txt"
./ayo sandbox pull /tmp/sandbox-file.txt /tmp/pulled-file.txt
cat /tmp/pulled-file.txt
```
- [ ] File pulled to host
- [ ] Content matches

### 4.3 Working Copy Diff
```bash
# Create a test directory
mkdir -p /tmp/test-workdir
echo "original" > /tmp/test-workdir/file.txt

# Push to sandbox and modify
./ayo sandbox push /tmp/test-workdir /workspace/test
./ayo sandbox exec sh -c "echo 'modified' > /workspace/test/file.txt"

# Show diff (sandbox-path first, then host-path)
./ayo sandbox diff /workspace/test /tmp/test-workdir
```
- [ ] Shows differences between sandbox and host

### 4.4 Working Copy Sync
```bash
# Sync from sandbox back to host (sandbox-path first, then host-path)
./ayo sandbox sync /workspace/test /tmp/test-workdir
cat /tmp/test-workdir/file.txt
```
- [ ] Changes synced back to host
- [ ] File now shows "modified"

---

## 5. Mount Management

Mounts grant sandboxed agents access to host filesystem paths. Grants persist across sessions.

### 5.1 Add Mount (Read-Write)
```bash
mkdir -p /tmp/test-mount
./ayo mount add /tmp/test-mount
```
- [ ] Shows: `✓ Granted readwrite access to /tmp/test-mount`

### 5.2 Add Mount (Read-Only)
```bash
mkdir -p /tmp/readonly-mount
./ayo mount add /tmp/readonly-mount --ro
```
- [ ] Shows: `✓ Granted readonly access to /tmp/readonly-mount`

### 5.3 Add Mount (Non-Existent Path)
```bash
./ayo mount add /tmp/does-not-exist
```
- [ ] Shows warning: `! Path does not exist: /tmp/does-not-exist`
- [ ] Still grants the mount (for future use)

### 5.4 Add Mount (Home Directory Expansion)
```bash
./ayo mount add ~/test-mount-dir --ro
./ayo mount list
```
- [ ] Path is expanded to full home directory path
- [ ] Shows in list with correct mode

### 5.5 List Mounts
```bash
./ayo mount list
```
- [ ] Shows table with PATH, MODE, GRANTED columns
- [ ] All added mounts appear in list

### 5.6 List Mounts (JSON)
```bash
./ayo mount list --json
```
- [ ] Valid JSON array output
- [ ] Each entry has path, mode, granted_at fields

### 5.7 Remove Mount
```bash
./ayo mount rm /tmp/test-mount
```
- [ ] Shows: `✓ Revoked access to /tmp/test-mount`

### 5.8 Remove Non-Existent Mount
```bash
./ayo mount rm /tmp/never-granted
```
- [ ] Shows warning: `! Path not granted: /tmp/never-granted`
- [ ] No error (graceful handling)

### 5.9 Remove All Mounts
```bash
./ayo mount rm --all
./ayo mount list
```
- [ ] Shows: `✓ Revoked N grant(s)`
- [ ] List shows no grants

### 5.10 Verify Sandbox Access
```bash
# Re-add mount for testing
mkdir -p /tmp/sandbox-mount-test
echo "host content" > /tmp/sandbox-mount-test/file.txt
./ayo mount add /tmp/sandbox-mount-test

# Verify sandbox can access it (requires running sandbox)
./ayo sandbox exec cat /tmp/sandbox-mount-test/file.txt
```
- [ ] Sandbox can read file from mounted path
- [ ] Content matches host file

### 5.11 Cleanup
```bash
./ayo mount rm /tmp/sandbox-mount-test
./ayo mount rm ~/test-mount-dir
./ayo mount rm /tmp/does-not-exist
rm -rf /tmp/sandbox-mount-test /tmp/test-mount /tmp/readonly-mount
```
- [ ] All test mounts removed

---

## 6. Trigger Management

### 6.1 List Triggers
```bash
./ayo triggers list
```
- [ ] Shows empty table or existing triggers

### 6.2 Add Cron Trigger
```bash
./ayo triggers add --type cron --schedule "*/5 * * * *" --agent @ayo --prompt "Check status"
./ayo triggers list
```
- [ ] Trigger appears in list
- [ ] Shows cron schedule

### 6.3 Add Watch Trigger
```bash
./ayo triggers add --type watch --path /tmp --agent @ayo --prompt "File changed"
./ayo triggers list
```
- [ ] Watch trigger appears
- [ ] Shows watched path

### 6.4 Show Trigger Details
```bash
./ayo triggers show <id>
```
- [ ] Shows full trigger config

### 6.5 Enable/Disable Trigger
```bash
./ayo triggers disable <id>
./ayo triggers show <id>  # Should show disabled
./ayo triggers enable <id>
./ayo triggers show <id>  # Should show enabled
```
- [ ] Disable changes status
- [ ] Enable restores status

### 6.6 Test Trigger
```bash
./ayo triggers test <id>
```
- [ ] Trigger fires manually
- [ ] Agent executes prompt

### 6.7 Remove Trigger
```bash
./ayo triggers rm <id>
./ayo triggers list
```
- [ ] Trigger removed from list

---

## 7. Multi-Agent Sandbox

### 7.1 Join Agent to Sandbox
```bash
./ayo sandbox join secondagent
# Or with explicit sandbox ID
./ayo sandbox join secondagent --id <id>
```
- [ ] Agent added to sandbox
- [ ] User created inside container

### 7.2 List Agents in Sandbox
```bash
./ayo sandbox users
# Or with explicit sandbox ID
./ayo sandbox users --id <id>
```
- [ ] Shows all agents in sandbox

### 7.3 Verify Agent Isolation
```bash
./ayo sandbox exec --user secondagent whoami
./ayo sandbox exec --user secondagent pwd
```
- [ ] Runs as correct user
- [ ] Home directory is /home/secondagent

---

## 8. Agent Execution

### 8.1 Basic Agent Run
```bash
./ayo @ayo "What OS am I running on? Run uname -a"
```
- [ ] Agent responds
- [ ] Shows Linux kernel (sandboxed)

### 8.2 File Operations in Sandbox
```bash
./ayo @ayo "Create a file /tmp/agent-test.txt with 'hello world'"
./ayo sandbox exec cat /tmp/agent-test.txt
```
- [ ] Agent creates file
- [ ] File exists in sandbox (not on host)

---

## 9. Sessions

### 9.1 List Sessions
```bash
./ayo sessions list
```
- [ ] Shows previous agent sessions

### 9.2 Continue Session
```bash
./ayo sessions continue --latest
```
- [ ] Resumes previous conversation

### 9.3 Delete Session
```bash
./ayo sessions delete <id>
./ayo sessions list
```
- [ ] Session removed

---

## 10. Debug Scripts

### 10.1 Sandbox Status
```bash
./debug/sandbox-status.sh
./debug/sandbox-status.sh --verbose
```
- [ ] Shows container status
- [ ] Verbose includes processes

### 10.2 Daemon Status
```bash
./debug/daemon-status.sh
./debug/daemon-status.sh --logs
```
- [ ] Shows daemon info
- [ ] Logs option shows recent logs

### 10.3 Mount Check
```bash
./debug/mount-check.sh
```
- [ ] Verifies mount permissions

### 10.4 Collect All
```bash
./debug/collect-all.sh | head -100
```
- [ ] Comprehensive diagnostic output

---

## 11. Matrix Communication

### 11.1 Status Check
```bash
./ayo matrix status
./ayo matrix status --json
```
- [ ] Shows Conduit status (running/stopped)
- [ ] Shows broker connection status
- [ ] JSON output works

### 11.2 Room Management
```bash
# List rooms
./ayo matrix rooms

# Create a test room
./ayo matrix create test-room

# List rooms again
./ayo matrix rooms
```
- [ ] Rooms listed (may be empty initially)
- [ ] Room creation succeeds
- [ ] New room appears in list

### 11.3 Messaging
```bash
# Send a message to the test room
./ayo matrix send test-room "Hello, world!"

# Read messages
./ayo matrix read test-room

# Read with limit
./ayo matrix read test-room 5
```
- [ ] Message sends successfully
- [ ] Read shows sent message
- [ ] Limit works correctly

### 11.4 Room Membership
```bash
# Show who is in the room
./ayo matrix who test-room

# Invite another agent (if available)
./ayo matrix invite test-room @ayo
```
- [ ] Who shows current member(s)
- [ ] Invite works (or shows appropriate error if agent doesn't exist)

### 11.5 Matrix from Sandbox
```bash
# Enter sandbox login shell
./ayo sandbox login

# Inside sandbox, verify socket mount
ls -la /run/ayo/

# Test matrix commands work from sandbox
ayo matrix status
ayo matrix rooms
```
- [ ] /run/ayo/ directory exists with sockets
- [ ] Matrix commands work from inside sandbox

---

## 12. Flow System

### 12.1 List Flows
```bash
./ayo flows list
./ayo flows list --json
```
- [ ] Shows available flows
- [ ] JSON output works

### 12.2 Create Shell Flow
```bash
./ayo flows new test-flow
cat ~/.config/ayo/flows/test-flow.sh
```
- [ ] Flow file created
- [ ] Has ayo:flow frontmatter
- [ ] Has shebang and set -euo pipefail

### 12.3 Run Shell Flow
```bash
./ayo flows run test-flow '{"message": "hello"}'
```
- [ ] Flow executes
- [ ] Output is JSON

### 12.4 Create YAML Flow
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
EOF
```
- [ ] YAML file created with valid syntax

### 12.5 Validate Flow
```bash
./ayo flows validate /tmp/test-yaml-flow.yaml
```
- [ ] Shows valid or errors

### 12.6 Flow History
```bash
./ayo flows history
./ayo flows history --flow test-flow
```
- [ ] Shows run history
- [ ] Filter by flow works

### 12.7 Flow Stats
```bash
./ayo flows stats
./ayo flows stats test-flow
```
- [ ] Shows execution statistics
- [ ] Per-flow stats work

### 12.8 Remove Flow
```bash
./ayo flows rm test-flow --force
./ayo flows list
```
- [ ] Flow removed
- [ ] No longer in list

---

## Cleanup

```bash
./ayo sandbox service stop
./ayo sandbox prune --all --force
rm -rf /tmp/test-workdir /tmp/testfile.txt /tmp/pulled-file.txt
```

---

## Test Results Summary

| Section | Status | Notes |
|---------|--------|-------|
| Pre-Flight | | |
| Service Operations | | |
| Sandbox Lifecycle | | |
| File Transfer | | |
| Mount Management | | |
| Trigger Management | | |
| Multi-Agent | | |
| Agent Execution | | |
| Sessions | | |
| Debug Scripts | | |
| Matrix Communication | | |
| Flow System | | |

**Issues Found:**
- 

**Notes:**
- 
