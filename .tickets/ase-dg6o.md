---
id: ase-dg6o
status: closed
deps: [ase-et8m, ase-frc8, ase-gbox]
links: []
created: 2026-02-10T01:35:47Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-8d04
---
# Add manual test cases for share system to MANUAL_TEST.md

Document manual testing procedures for the share system.

## File to Modify
- MANUAL_TEST.md (or create if doesn't exist)

## Dependencies
- ase-et8m (share add works)
- ase-frc8 (share list works)
- ase-gbox (share rm works)

## Content to Add

### Section: Share System

```markdown
## Share System

The share system allows sharing host directories with sandboxed agents
without requiring sandbox restart.

### Prerequisites
- Daemon running: `ayo service start`
- At least one sandbox available

### Test: Basic Share Workflow

1. Create a test directory:
   ```bash
   mkdir -p /tmp/share-test
   echo "test content" > /tmp/share-test/file.txt
   ```

2. Share the directory:
   ```bash
   ayo share /tmp/share-test
   ```
   Expected: ✓ Shared /tmp/share-test → /workspace/share-test

3. List shares:
   ```bash
   ayo share list
   ```
   Expected: Shows share-test with path and timestamp

4. Verify visibility in sandbox:
   ```bash
   ayo sandbox exec <id> ls /workspace/
   ```
   Expected: Shows share-test directory

5. Read file through sandbox:
   ```bash
   ayo sandbox exec <id> cat /workspace/share-test/file.txt
   ```
   Expected: "test content"

6. Write file through sandbox:
   ```bash
   ayo sandbox exec <id> 'echo "new file" > /workspace/share-test/new.txt'
   cat /tmp/share-test/new.txt
   ```
   Expected: "new file" visible on host

7. Remove share:
   ```bash
   ayo share rm share-test
   ```
   Expected: ✓ Removed share 'share-test'

8. Verify removal in sandbox:
   ```bash
   ayo sandbox exec <id> ls /workspace/
   ```
   Expected: share-test no longer listed

### Test: Custom Name with --as

```bash
ayo share /tmp/share-test --as myproject
ayo share list
# Should show: myproject → /tmp/share-test
ayo share rm myproject
```

### Test: Name Conflict Detection

```bash
ayo share /tmp/share-test --as foo
ayo share /tmp/other-dir --as foo
# Expected error: share 'foo' already exists, use --as to specify a different name
ayo share rm foo
```

### Test: Path Expansion

```bash
# Test ~/ expansion
ayo share ~/Desktop --as desktop
ayo share list
# Should show absolute path like /Users/alex/Desktop
ayo share rm desktop

# Test relative paths
cd /tmp
ayo share . --as current
ayo share list
# Should show /tmp
ayo share rm current
```

### Test: Remove by Path

```bash
ayo share /tmp/share-test --as myshare
ayo share rm /tmp/share-test  # Remove by original path
# Expected: ✓ Removed share 'myshare'
```

### Test: JSON Output

```bash
ayo share /tmp/share-test --json
# Expected: JSON with name, path, workspace_path

ayo share list --json
# Expected: JSON array of shares

ayo share rm share-test --json
# Expected: JSON with removed field
```

### Test: Path Validation

```bash
ayo share /nonexistent/path
# Expected error: path does not exist
```

### Test: No Restart Required

1. Start sandbox and leave it running
2. Add multiple shares
3. Verify each is immediately visible without restart
4. Remove shares
5. Verify immediately removed without restart

### Cleanup

```bash
rm -rf /tmp/share-test
ayo share rm --all
```
```

## Acceptance Criteria

- [ ] All test cases documented
- [ ] Prerequisites section included
- [ ] Expected output for each command
- [ ] Cleanup steps provided
- [ ] Edge cases covered (conflicts, paths, JSON)

