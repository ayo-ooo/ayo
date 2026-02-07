---
id: ayo-athk
status: closed
deps: [ayo-sfoy]
links: []
created: 2026-02-06T22:15:09Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [testing, manual-test, mounts]
---
# Test Section 5: Mount System

Test host filesystem mount system for sandbox access.

## Scope
Test mount operations:
- List current mounts
- Add read-write mounts
- Add read-only mounts
- Verify mount access from sandbox
- Verify read-only enforcement
- Remove mounts

## Setup
```bash
mkdir -p /tmp/ayo-mount-test
echo 'host-file-content' > /tmp/ayo-mount-test/host-file.txt
mkdir -p /tmp/ayo-mount-readonly
echo 'readonly-content' > /tmp/ayo-mount-readonly/ro-file.txt
```

## Test Steps

### 5.1 List Mounts (Initial)
```bash
ayo mount list
```
Document: Current mount state

### 5.2 Add Mount (Read-Write)
```bash
ayo mount add /tmp/ayo-mount-test --reason 'Agent testing'
ayo mount list
```
Expected: Mount added successfully

### 5.3 Verify Mount in Sandbox
```bash
ayo @ayo "list files in /tmp/ayo-mount-test"
```
Expected: Agent can see host-file.txt

### 5.4 Read Mounted File
```bash
ayo @ayo "read /tmp/ayo-mount-test/host-file.txt"
```
Expected: Content matches 'host-file-content'

### 5.5 Write to Mounted Directory
```bash
ayo @ayo "create file /tmp/ayo-mount-test/sandbox-file.txt with content 'from-sandbox'"
cat /tmp/ayo-mount-test/sandbox-file.txt
```
Expected: File created on host with correct content

### 5.6 Add Mount (Read-Only)
```bash
ayo mount add /tmp/ayo-mount-readonly --readonly --reason 'RO test'
ayo mount list
```
Expected: Mount added with read-only flag

### 5.7 Verify Read-Only Enforcement
```bash
ayo @ayo "try to create a file in /tmp/ayo-mount-readonly and report what happens"
```
Expected: Write fails with permission error

### 5.8 Remove Mount
```bash
ayo mount rm /tmp/ayo-mount-test
ayo mount list
```
Expected: Mount removed from list

## Cleanup
```bash
ayo mount rm /tmp/ayo-mount-test 2>/dev/null || true
ayo mount rm /tmp/ayo-mount-readonly 2>/dev/null || true
rm -rf /tmp/ayo-mount-test /tmp/ayo-mount-readonly
ayo sandbox prune --all --force
```

## Exit Criteria
Mount system works correctly with proper read/write enforcement


## Notes

**2026-02-06T22:23:13Z**

BLOCKED - 'ayo mount' CLI commands not implemented (ticket ayo-2s3e).

Current state:
- Working directory auto-mounted via virtiofs
- Agent can access host files in mounted path
- No CLI to manage mounts

Skipped tests:
- 5.1-5.9 All mount management tests

Observed:
- Sandbox shows virtiofs mount at /Users/acabrera/Code/ayo-ooo/ayo
