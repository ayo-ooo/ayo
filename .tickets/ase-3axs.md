---
id: ase-3axs
status: closed
deps: [ase-nvjz, ase-c1y8, ase-et8m]
links: []
created: 2026-02-10T01:35:24Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-8d04
---
# Integration test: share visibility in sandbox

Create integration tests verifying that shares are immediately visible inside running sandboxes.

## Context
The key value proposition of the share system is that symlinks are instantly visible without sandbox restart. This test verifies that end-to-end.

## Dependencies
- ase-nvjz (sandbox manager has workspace mount)
- ase-c1y8 (pool has workspace mount)
- ase-et8m (share add command works)

## Test Scenarios

### Test 1: Share visible after creation
1. Start a sandbox
2. Create a share on host: `ayo share /tmp/testshare`
3. Verify /workspace/testshare exists inside sandbox
4. Clean up

### Test 2: File operations through share
1. Start a sandbox
2. Share a test directory
3. Create file inside sandbox at /workspace/test/newfile.txt
4. Verify file exists on host at shared path
5. Modify file on host
6. Verify changes visible inside sandbox
7. Clean up

### Test 3: Share removal immediately reflected
1. Start a sandbox
2. Create share
3. Verify visible in sandbox
4. Remove share: `ayo share rm test`
5. Verify /workspace/test no longer exists in sandbox
6. Clean up

### Test 4: Multiple shares
1. Start sandbox
2. Create multiple shares
3. Verify all visible at /workspace/{name}
4. Remove one, verify others still work
5. Clean up

## Test Implementation
```go
func TestShareVisibleInSandbox(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Create temp directory to share
    shareDir := t.TempDir()
    testFile := filepath.Join(shareDir, "test.txt")
    os.WriteFile(testFile, []byte("hello"), 0644)

    // Start sandbox (use test helper or daemon client)
    // ...

    // Create share
    service := share.NewShareService()
    service.Load()
    err := service.Add(shareDir, "testshare", false, "")
    require.NoError(t, err)

    // Execute ls inside sandbox to verify
    output, err := sandbox.Exec("ls /workspace/testshare")
    require.NoError(t, err)
    assert.Contains(t, output, "test.txt")

    // Read file through sandbox
    output, err = sandbox.Exec("cat /workspace/testshare/test.txt")
    require.NoError(t, err)
    assert.Equal(t, "hello", strings.TrimSpace(output))

    // Clean up
    service.Remove("testshare")
}
```

## File Location
- Test file: internal/share/integration_test.go or tests/share_integration_test.go

## Prerequisites
- Sandbox provider available (apple container or systemd-nspawn)
- Daemon running for sandbox management
- May need build tag for integration tests

## Skip Conditions
- Skip on CI if no container runtime available
- Skip with testing.Short()

## Acceptance Criteria

- [ ] Test share visible after creation
- [ ] Test file read/write through share
- [ ] Test share removal immediately reflected
- [ ] Test multiple simultaneous shares
- [ ] Tests skip gracefully when sandbox unavailable
- [ ] Tests clean up after themselves

