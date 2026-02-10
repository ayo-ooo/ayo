---
id: ase-qs8a
status: closed
deps: [ase-ionw, ase-1f71]
links: []
created: 2026-02-10T01:35:04Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-8d04
---
# Unit tests for share package

Create comprehensive unit tests for the internal/share package.

## File to Create
- internal/share/share_test.go

## Dependencies
- ase-ionw (Share types exist)
- ase-1f71 (Add/Remove exist)

## Test Cases

### ShareService basics
```go
func TestNewShareService(t *testing.T) {
    // Verify service is created with correct file path
}

func TestShareService_LoadMissingFile(t *testing.T) {
    // Load() should succeed with empty shares when file doesn't exist
}

func TestShareService_LoadExistingFile(t *testing.T) {
    // Load() should parse existing shares.json
}

func TestShareService_SaveCreatesDirectory(t *testing.T) {
    // Save() should create parent directory if missing
}
```

### Share operations
```go
func TestShareService_Add(t *testing.T) {
    // Add creates symlink and persists
}

func TestShareService_Add_AutoName(t *testing.T) {
    // Name derived from path basename when not specified
}

func TestShareService_Add_CustomName(t *testing.T) {
    // --as flag sets custom name
}

func TestShareService_Add_PathNotExists(t *testing.T) {
    // Error when path doesn't exist
}

func TestShareService_Add_NameConflict(t *testing.T) {
    // Error when name already exists
}

func TestShareService_Add_InvalidName(t *testing.T) {
    // Error for names with / or \
}

func TestShareService_Remove_ByName(t *testing.T) {
    // Remove by workspace name
}

func TestShareService_Remove_ByPath(t *testing.T) {
    // Remove by original host path
}

func TestShareService_Remove_NotFound(t *testing.T) {
    // Idempotent - no error if not found
}

func TestShareService_List(t *testing.T) {
    // Returns copy of all shares
}

func TestShareService_Get(t *testing.T) {
    // Get by name returns share or nil
}

func TestShareService_GetByPath(t *testing.T) {
    // Get by path returns share or nil
}
```

### Symlink verification
```go
func TestShareService_Add_CreatesSymlink(t *testing.T) {
    // Verify symlink exists in workspace dir after Add
}

func TestShareService_Remove_DeletesSymlink(t *testing.T) {
    // Verify symlink removed after Remove
}

func TestShareService_SymlinkTarget(t *testing.T) {
    // Verify symlink points to correct host path
}
```

### Session shares
```go
func TestShareService_SessionShare(t *testing.T) {
    // Session flag is stored correctly
}

func TestShareService_RemoveSessionShares(t *testing.T) {
    // Only removes shares with matching session ID
}
```

### Name validation
```go
func TestValidateShareName(t *testing.T) {
    cases := []struct {
        name    string
        wantErr bool
    }{
        {"project", false},
        {"my-project", false},
        {"my_project", false},
        {"", true},             // empty
        {"../escape", true},    // path traversal
        {"has/slash", true},    // contains separator
        {"-starts-dash", true}, // starts with dash
        {".hidden", false},     // allowed? TBD
    }
}
```

## Test Helpers
- Create temp directory for workspace
- Override WorkspaceDir() for testing
- Clean up after tests

## Example Test Structure
```go
func TestShareService_Add(t *testing.T) {
    // Setup temp directories
    tmpDir := t.TempDir()
    workspaceDir := filepath.Join(tmpDir, "workspace")
    os.MkdirAll(workspaceDir, 0755)
    
    // Create a path to share
    sharePath := filepath.Join(tmpDir, "myproject")
    os.MkdirAll(sharePath, 0755)
    
    // Override workspace dir for test
    origWorkspaceDir := sync.WorkspaceDir
    sync.SetWorkspaceDirForTest(workspaceDir)
    defer sync.SetWorkspaceDirForTest("")
    
    // Create service with temp file
    service := &share.ShareService{...}
    
    // Test Add
    err := service.Add(sharePath, "", false, "")
    require.NoError(t, err)
    
    // Verify symlink
    symlinkPath := filepath.Join(workspaceDir, "myproject")
    target, err := os.Readlink(symlinkPath)
    require.NoError(t, err)
    assert.Equal(t, sharePath, target)
}
```

## Acceptance Criteria

- [ ] All test cases implemented
- [ ] Tests use temp directories
- [ ] No side effects on real filesystem
- [ ] Tests pass on CI
- [ ] >80% code coverage for share package

