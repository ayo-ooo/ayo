package share

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	service := NewService()
	assert.NotNil(t, service)
	assert.NotEmpty(t, service.filePath)
}

func TestService_LoadMissingFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	service := &Service{
		filePath: filepath.Join(tmpDir, "shares.json"),
	}

	err := service.Load()
	require.NoError(t, err)
	assert.NotNil(t, service.shares)
	assert.Equal(t, 1, service.shares.Version)
	assert.Empty(t, service.shares.Shares)
}

func TestService_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	service := &Service{
		filePath: filepath.Join(tmpDir, "shares.json"),
		shares: &SharesFile{
			Version: 1,
			Shares: []Share{
				{Name: "test", Path: "/tmp/test"},
			},
		},
	}

	// Save
	err := service.Save()
	require.NoError(t, err)

	// Load into new service
	service2 := &Service{
		filePath: filepath.Join(tmpDir, "shares.json"),
	}
	err = service2.Load()
	require.NoError(t, err)

	assert.Len(t, service2.shares.Shares, 1)
	assert.Equal(t, "test", service2.shares.Shares[0].Name)
}

func TestService_List(t *testing.T) {
	service := &Service{
		shares: &SharesFile{
			Version: 1,
			Shares: []Share{
				{Name: "a", Path: "/a"},
				{Name: "b", Path: "/b"},
			},
		},
	}

	list := service.List()
	assert.Len(t, list, 2)
	assert.Equal(t, "a", list[0].Name)
	assert.Equal(t, "b", list[1].Name)
}

func TestService_List_Empty(t *testing.T) {
	service := &Service{}
	list := service.List()
	assert.Nil(t, list)
}

func TestService_Get(t *testing.T) {
	service := &Service{
		shares: &SharesFile{
			Version: 1,
			Shares: []Share{
				{Name: "project", Path: "/home/user/project"},
			},
		},
	}

	// Found
	share := service.Get("project")
	require.NotNil(t, share)
	assert.Equal(t, "/home/user/project", share.Path)

	// Not found
	share = service.Get("nonexistent")
	assert.Nil(t, share)
}

func TestService_GetByPath(t *testing.T) {
	service := &Service{
		shares: &SharesFile{
			Version: 1,
			Shares: []Share{
				{Name: "project", Path: "/home/user/project"},
			},
		},
	}

	// Found
	share := service.GetByPath("/home/user/project")
	require.NotNil(t, share)
	assert.Equal(t, "project", share.Name)

	// Not found
	share = service.GetByPath("/nonexistent")
	assert.Nil(t, share)
}

func TestValidateShareName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"project", false},
		{"my-project", false},
		{"my_project", false},
		{"project123", false},
		{".hidden", false}, // allowed
		{"", true},         // empty
		{".", true},        // current dir
		{"..", true},       // parent dir
		{"has/slash", true},
		{"has\\backslash", true},
		{"-starts-dash", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateShareName(tt.name)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Add_PathNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	service := &Service{
		filePath: filepath.Join(tmpDir, "shares.json"),
		shares:   &SharesFile{Version: 1, Shares: []Share{}},
	}

	err := service.Add("/nonexistent/path", "", false, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path does not exist")
}

func TestService_Add_NameConflict(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "testdir")
	os.MkdirAll(testPath, 0755)

	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	service := &Service{
		filePath: filepath.Join(tmpDir, "shares.json"),
		shares: &SharesFile{
			Version: 1,
			Shares:  []Share{{Name: "testdir", Path: "/other"}},
		},
	}

	err := service.Add(testPath, "", false, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestService_Add_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "testdir")
	os.MkdirAll(testPath, 0755)

	service := &Service{
		filePath: filepath.Join(tmpDir, "shares.json"),
		shares:   &SharesFile{Version: 1, Shares: []Share{}},
	}

	err := service.Add(testPath, "has/slash", false, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path separators")
}

func TestService_Remove_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	service := &Service{
		filePath: filepath.Join(tmpDir, "shares.json"),
		shares:   &SharesFile{Version: 1, Shares: []Share{}},
	}

	// Should not error - idempotent
	err := service.Remove("nonexistent")
	assert.NoError(t, err)
}

func TestService_Remove_ByName(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	// Create a fake symlink
	symlinkPath := filepath.Join(workspaceDir, "project")
	os.Symlink("/tmp", symlinkPath)

	service := &Service{
		filePath: filepath.Join(tmpDir, "shares.json"),
		shares: &SharesFile{
			Version: 1,
			Shares:  []Share{{Name: "project", Path: "/tmp/project"}},
		},
	}

	err := service.Remove("project")
	require.NoError(t, err)
	assert.Empty(t, service.shares.Shares)
}

func TestService_RemoveSessionShares(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	service := &Service{
		filePath: filepath.Join(tmpDir, "shares.json"),
		shares: &SharesFile{
			Version: 1,
			Shares: []Share{
				{Name: "permanent", Path: "/a", Session: false},
				{Name: "session1", Path: "/b", Session: true, SessionID: "sess-123"},
				{Name: "session2", Path: "/c", Session: true, SessionID: "sess-123"},
				{Name: "other-session", Path: "/d", Session: true, SessionID: "sess-456"},
			},
		},
	}

	err := service.RemoveSessionShares("sess-123")
	require.NoError(t, err)

	// Should have 2 remaining: permanent and other-session
	assert.Len(t, service.shares.Shares, 2)

	names := make([]string, len(service.shares.Shares))
	for i, s := range service.shares.Shares {
		names[i] = s.Name
	}
	assert.Contains(t, names, "permanent")
	assert.Contains(t, names, "other-session")
}
