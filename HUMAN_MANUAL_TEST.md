# Ayo Human Manual Testing Guide

This guide walks you through testing the ayo sandbox-first architecture. I will run commands and report results. **You only need to respond if you have feedback or want changes** — silence means approval.

---

## Communication Protocol

- **I run commands** and show you the output
- **You review** and respond ONLY if something needs attention
- **No response = approved** — I continue to the next section
- **Questions are marked with ❓** — these require your input

---

## Table of Contents

1. [Pre-Flight Checks](#1-pre-flight-checks)
2. [Sandbox Lifecycle](#2-sandbox-lifecycle)
3. [Agent Execution](#3-agent-execution)
4. [Daemon Operations](#4-daemon-operations)
5. [Sessions](#5-sessions)

---

## 1. Pre-Flight Checks

### 1.1 System Requirements

I will run `./debug/system-info.sh` and verify:
- Docker is running (required for Apple Container)
- ayo binary exists
- Config paths are correct

### 1.2 Doctor Check

I will run `ayo doctor` and verify all checks pass.

**Expected results:**
- Ayo Version: 0.3.0+
- Platform: darwin/arm64 (or your platform)
- Sandbox Provider: apple-container (macOS) or linux equivalent
- Database: OK
- All critical paths OK

**Known acceptable warnings:**
- Config File: WARN (using defaults) - this is fine if you haven't created a config

---

## 2. Sandbox Lifecycle

### 2.1 Initial State

I will run `ayo sandbox list` to show current sandbox state.

**Note:** Without daemon running, this may show "No active sandboxes".

### 2.2 Create Sandbox via Agent

I will run `ayo @ayo "echo hello from sandbox"` and verify:
- Agent responds successfully
- Command executes inside Linux container (not on host)

**Note:** Per-session sandboxes are ephemeral - they're created and destroyed with each agent run.

### 2.3 Daemon-Managed Sandbox

With daemon running (`ayo daemon start`), a persistent sandbox is maintained in the pool.

I will verify:
- `ayo status` shows Sandbox Pool with at least 1 sandbox
- `ayo sandbox list` shows running sandbox

### 2.4 Sandbox Commands

Available commands:
- `ayo sandbox list` - List active sandboxes
- `ayo sandbox show <id>` - Show sandbox details
- `ayo sandbox exec <id> -- <cmd>` - Execute command in sandbox
- `ayo sandbox start <id>` - Start a stopped sandbox
- `ayo sandbox stop <id>` - Stop a running sandbox
- `ayo sandbox prune [--all] [--force]` - Remove stopped sandboxes

---

## 3. Agent Execution

### 3.1 Verify Sandboxed Execution

I will have the agent run commands and verify:
- `uname -a` shows Linux kernel (not Darwin)
- `pwd` shows mounted host directory path
- Process isolation works (only container processes visible)

**Expected behavior:**
- Agent runs as root (uid=0) inside container
- Host directory is mounted at same path inside container
- Sandboxes are ephemeral per session (files in /tmp don't persist)

### 3.2 File Operations

I will have the agent:
- Create files in /tmp (works, but ephemeral)
- Access files in mounted host directory (works)
- Verify isolation (files in /tmp don't appear on host)

---

## 4. Daemon Operations

### 4.1 Start/Stop Daemon

Commands:
- `ayo daemon start` - Start daemon (runs in background)
- `ayo daemon stop` - Stop daemon (terminates process)
- `ayo status` - Check daemon status

**Expected status output:**
- CLI Version
- Daemon: running/not running
- PID, Uptime, Memory (when running)
- Sandbox Pool: Total, Idle, In Use

### 4.2 Known Limitations

- `ayo daemon logs` - Not implemented
- `ayo daemon sessions` - Not implemented

---

## 5. Sessions

### 5.1 Session Management

Commands:
- `ayo sessions list` - List all sessions
- `ayo sessions show <id>` - Show session details
- `ayo sessions continue <id>` - Continue a session
- `ayo sessions continue --latest` - Continue most recent session
- `ayo sessions delete <id> [--force]` - Delete a session

### 5.2 Session Persistence

Sessions are stored in SQLite database and persist across restarts.

---

## Features Not Yet Implemented

The following features are documented in AGENTS.md but not yet implemented in the CLI:

| Feature | Status |
|---------|--------|
| `ayo mount` commands | Not implemented |
| `ayo triggers` commands | Not implemented |
| `ayo backup` commands | Not implemented |
| `ayo sync` commands | Not implemented |
| IRC inter-agent communication | Not implemented (no ngircd in sandbox) |

---

## Post-Testing Summary

After all tests complete, I will provide:
1. Summary of passed/failed tests
2. List of any new issues found
3. Current working features

**❓ Are there specific areas you want deeper testing on?**
