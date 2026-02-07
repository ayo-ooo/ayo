# Ayo Human Manual Testing Guide

Complete manual testing checklist for ayo sandbox-first architecture.

---

## Prerequisites

```bash
# Build fresh binary
go build -o ayo ./cmd/ayo/...

# Verify version
./ayo --version
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

**Note:** All sandbox commands accept an optional `[id]`. If omitted:
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
# With ID
./ayo sandbox show <id>
# Or without (auto-selects or picker)
./ayo sandbox show
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
./ayo sandbox start <id>
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
./ayo sandbox prune --dry-run
./ayo sandbox prune
./ayo sandbox prune --homes  # Cleans agent home dirs too
```
- [ ] Dry-run shows what would be removed
- [ ] Prune removes stopped sandboxes
- [ ] --homes cleans ~/.local/share/ayo/homes/

---

## 4. File Transfer

### 4.1 Push File to Sandbox
```bash
echo "test content" > /tmp/testfile.txt
./ayo sandbox push <id> /tmp/testfile.txt /tmp/testfile.txt
./ayo sandbox exec <id> cat /tmp/testfile.txt
```
- [ ] File appears in sandbox
- [ ] Content matches

### 4.2 Pull File from Sandbox
```bash
./ayo sandbox exec <id> "echo 'from sandbox' > /tmp/sandbox-file.txt"
./ayo sandbox pull <id> /tmp/sandbox-file.txt /tmp/pulled-file.txt
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
./ayo sandbox push <id> /tmp/test-workdir /workspace/test
./ayo sandbox exec <id> "echo 'modified' > /workspace/test/file.txt"

# Show diff
./ayo sandbox diff <id> /tmp/test-workdir
```
- [ ] Shows differences between sandbox and host

### 4.4 Working Copy Sync
```bash
./ayo sandbox sync <id> /tmp/test-workdir
cat /tmp/test-workdir/file.txt
```
- [ ] Changes synced back to host
- [ ] File now shows "modified"

---

## 5. Mount Management

### 5.1 Grant Access
```bash
./ayo mount grant /tmp/test-mount
# OR using alias:
./ayo mount add /tmp/test-mount
```
- [ ] Mount added successfully

### 5.2 List Mounts
```bash
./ayo mount list
```
- [ ] Shows granted paths with permissions
- [ ] JSON output works: `./ayo mount list --json`

### 5.3 Revoke Access
```bash
./ayo mount revoke /tmp/test-mount
# OR using alias:
./ayo mount rm /tmp/test-mount
./ayo mount list
```
- [ ] Mount removed from list

---

## 6. Trigger Management

### 6.1 List Triggers
```bash
./ayo triggers list
```
- [ ] Shows empty table or existing triggers

### 6.2 Add Cron Trigger
```bash
./ayo triggers add --cron "*/5 * * * *" --agent ayo --prompt "Check status"
./ayo triggers list
```
- [ ] Trigger appears in list
- [ ] Shows cron schedule

### 6.3 Add Watch Trigger
```bash
./ayo triggers add --watch /tmp --agent ayo --prompt "File changed"
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
./ayo sandbox join <id> secondagent
```
- [ ] Agent added to sandbox
- [ ] User created inside container

### 7.2 List Agents in Sandbox
```bash
./ayo sandbox users <id>
```
- [ ] Shows all agents in sandbox

### 7.3 Verify Agent Isolation
```bash
./ayo sandbox exec <id> --user secondagent whoami
./ayo sandbox exec <id> --user secondagent pwd
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
./ayo sandbox exec <pool-id> cat /tmp/agent-test.txt
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

## Cleanup

```bash
./ayo service stop
./ayo sandbox prune --all --force
rm -rf /tmp/test-workdir /tmp/testfile.txt /tmp/pulled-file.txt
```

---

## Test Results Summary

| Section | Status | Notes |
|---------|--------|-------|
| Pre-Flight | | |
| Daemon | | |
| Sandbox Lifecycle | | |
| File Transfer | | |
| Mount Management | | |
| Trigger Management | | |
| Multi-Agent | | |
| Agent Execution | | |
| Sessions | | |
| Debug Scripts | | |

**Issues Found:**
- 

**Notes:**
- 
