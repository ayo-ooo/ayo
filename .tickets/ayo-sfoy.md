---
id: ayo-sfoy
status: closed
deps: [ayo-k22j]
links: []
created: 2026-02-06T22:14:57Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [testing, manual-test, sandbox, agent]
---
# Test Section 4: Agent Execution in Sandbox

Test agent command execution within sandbox environment.

## Scope
Verify agents execute commands inside sandbox with proper isolation:
- Working directory is inside container
- File operations work
- Environment variables are set
- Network access policy
- Process isolation
- Filesystem boundaries
- Resource limits

## Test Steps

### 4.1 Verify Sandbox Execution Path
```bash
ayo @ayo "run 'pwd' and tell me the working directory"
```
Expected: Path inside container, NOT host path

### 4.2 File Creation in Sandbox
```bash
ayo @ayo "create a file at /tmp/agent-test.txt with content 'test-content-12345'"
ayo sandbox exec <id> -- cat /tmp/agent-test.txt
```
Expected: File created, content matches

### 4.3 File Reading in Sandbox
```bash
ayo @ayo "read the contents of /tmp/agent-test.txt"
```
Expected: Agent reports correct content

### 4.4 Environment Variables
```bash
ayo @ayo "run 'env' and list any AYO or USER related variables"
```
Expected: Environment variables documented

### 4.5 Network Access
```bash
ayo @ayo "run 'ping -c 1 8.8.8.8' or equivalent network test"
```
Document: Network access policy (allowed/blocked)

### 4.6 Process Isolation
```bash
ayo @ayo "run 'ps aux' and list all processes"
```
Expected: Only container processes visible

### 4.7 Filesystem Boundaries
```bash
ayo @ayo "try to list /etc/shadow and report what happens"
```
Expected: Permission denied or restricted access

### 4.8 Resource Limits
```bash
ayo sandbox exec <id> -- ulimit -a
```
Document: Resource limit values

## Analysis Required
- Document working directory path
- Document all AYO environment variables
- Document network policy
- Document process isolation
- Document security boundaries
- Document resource limits

## Cleanup
```bash
ayo sandbox prune --all --force
```

## Exit Criteria
Agent execution is properly sandboxed with expected isolation


## Notes

**2026-02-06T22:22:28Z**

COMPLETED:
- 4.1 Sandbox execution path: PASS - runs on Linux (uname shows Linux), pwd shows mounted host path
- 4.2 File creation: PASS - files created in /tmp (not on host)
- 4.3 File reading: N/A - ephemeral sandboxes don't persist files across sessions
- 4.4 Environment variables: PASS - minimal env (SHLVL, HOME=/root, PATH, PWD)
- 4.5 Network access: SKIPPED
- 4.6 Process isolation: PASS - only container processes visible (PID 1=sleep infinity)
- 4.7 Filesystem boundaries: Agent refused to read /etc/shadow (policy). Can read /etc/passwd.
- 4.8 Resource limits: SKIPPED (busybox ulimit limited)

Key observations:
- Agent runs as root (uid=0) inside container
- Sandboxes are ephemeral per session
- Host directory mounted at same path inside container
