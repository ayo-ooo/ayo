---
id: ayo-yh4d
status: closed
deps: [ayo-egy0]
links: []
created: 2026-02-06T22:15:33Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [testing, manual-test, triggers]
---
# Test Section 7: Trigger System

Test trigger system for automated agent execution.

## Scope
Test trigger operations:
- List triggers (empty state)
- Create cron trigger
- Create watch trigger
- Test trigger manually
- Enable/disable triggers
- Remove triggers

## Precondition
Daemon must be running

## Setup
```bash
ayo daemon start
mkdir -p /tmp/ayo-watch-test
```

## Test Steps

### 7.1 List Triggers (Empty)
```bash
ayo triggers list
```
Expected: Empty list or 'no triggers' message

### 7.2 Add Cron Trigger
```bash
ayo triggers add --type cron --agent @ayo --schedule '*/5 * * * *' --prompt 'echo trigger-test'
```
Expected: Trigger ID returned, success message
Record: Trigger ID for subsequent tests

### 7.3 List Triggers (With Cron)
```bash
ayo triggers list
```
Expected: Trigger in list, schedule shown, agent shown

### 7.4 Test Trigger Manually
```bash
ayo triggers test <trigger-id>
```
Expected: Trigger fires, agent executes, output shown

### 7.5 Add Watch Trigger
```bash
ayo triggers add --type watch --agent @ayo --path /tmp/ayo-watch-test --patterns '*.txt' --prompt 'file changed'
```
Expected: Trigger ID returned

### 7.6 Test Watch Trigger
```bash
touch /tmp/ayo-watch-test/test.txt
sleep 2
```
Verify: Check daemon logs for trigger fire

### 7.7 Disable Trigger
```bash
ayo triggers disable <trigger-id>
ayo triggers list
```
Expected: Trigger disabled, list shows disabled status

### 7.8 Enable Trigger
```bash
ayo triggers enable <trigger-id>
ayo triggers list
```
Expected: Trigger enabled, list shows enabled status

### 7.9 Remove Trigger
```bash
ayo triggers rm <trigger-id>
ayo triggers list
```
Expected: Trigger removed, no longer in list

## Cleanup
```bash
rm -rf /tmp/ayo-watch-test
# Remove any remaining test triggers
ayo triggers list 2>/dev/null | grep -E 'trigger-test|ayo-watch' && echo 'Remove remaining triggers manually'
ayo daemon stop
```

## Exit Criteria
Trigger system works for both cron and watch types


## Notes

**2026-02-06T22:25:22Z**

BLOCKED - 'ayo triggers' CLI commands not implemented (ticket ayo-f6ac).

Skipped all tests:
- 7.1-7.10 All trigger management tests
