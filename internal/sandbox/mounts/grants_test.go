package mounts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGrantService_LoadSave(t *testing.T) {
	// Create temp dir for test
	tmpDir, err := os.MkdirTemp("", "ayo-grants-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a service with custom path
	service := &GrantService{
		filePath: filepath.Join(tmpDir, "mounts.json"),
	}

	// Load empty (file doesn't exist)
	if err := service.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Grant a permission
	testPath := filepath.Join(tmpDir, "project")
	os.MkdirAll(testPath, 0755)
	
	if err := service.Grant(testPath, GrantModeReadWrite); err != nil {
		t.Fatalf("Grant failed: %v", err)
	}

	// Save
	if err := service.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new service and load
	service2 := &GrantService{
		filePath: filepath.Join(tmpDir, "mounts.json"),
	}
	if err := service2.Load(); err != nil {
		t.Fatalf("Load from disk failed: %v", err)
	}

	// Verify permission was persisted
	permissions := service2.List()
	if len(permissions) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(permissions))
	}
	if permissions[0].Path != testPath {
		t.Errorf("expected path %s, got %s", testPath, permissions[0].Path)
	}
	if permissions[0].Mode != GrantModeReadWrite {
		t.Errorf("expected mode readwrite, got %s", permissions[0].Mode)
	}
}

func TestGrantService_GrantRevoke(t *testing.T) {
	service := &GrantService{
		filePath: "/dev/null", // Won't save
	}
	service.Load()

	// Grant
	if err := service.Grant("/test/path", GrantModeReadOnly); err != nil {
		t.Fatalf("Grant failed: %v", err)
	}

	permissions := service.List()
	if len(permissions) != 1 {
		t.Fatalf("expected 1 permission after grant")
	}

	// Update grant (same path, different mode)
	if err := service.Grant("/test/path", GrantModeReadWrite); err != nil {
		t.Fatalf("Update grant failed: %v", err)
	}

	permissions = service.List()
	if len(permissions) != 1 {
		t.Fatalf("expected still 1 permission after update")
	}
	if permissions[0].Mode != GrantModeReadWrite {
		t.Errorf("expected mode to be updated to readwrite")
	}

	// Revoke
	if err := service.Revoke("/test/path"); err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}

	permissions = service.List()
	if len(permissions) != 0 {
		t.Fatalf("expected 0 permissions after revoke, got %d", len(permissions))
	}
}

func TestGrantService_IsGranted(t *testing.T) {
	service := &GrantService{
		filePath: "/dev/null",
	}
	service.Load()

	// Grant /Users/test
	service.Grant("/Users/test", GrantModeReadWrite)

	tests := []struct {
		path     string
		mode     GrantMode
		expected bool
	}{
		{"/Users/test", GrantModeReadOnly, true},
		{"/Users/test", GrantModeReadWrite, true},
		{"/Users/test/project", GrantModeReadWrite, true},
		{"/Users/test/project/subdir", GrantModeReadOnly, true},
		{"/Users/tester", GrantModeReadOnly, false}, // Different path
		{"/Users", GrantModeReadOnly, false},         // Parent path
		{"/other/path", GrantModeReadOnly, false},    // Unrelated path
	}

	for _, tc := range tests {
		result := service.IsGranted(tc.path, tc.mode)
		if result != tc.expected {
			t.Errorf("IsGranted(%s, %s) = %v, want %v", tc.path, tc.mode, result, tc.expected)
		}
	}
}

func TestGrantService_IsGranted_ReadOnlyMode(t *testing.T) {
	service := &GrantService{
		filePath: "/dev/null",
	}
	service.Load()

	// Grant read-only access
	service.Grant("/Users/readonly", GrantModeReadOnly)

	tests := []struct {
		path     string
		mode     GrantMode
		expected bool
	}{
		{"/Users/readonly", GrantModeReadOnly, true},
		{"/Users/readonly/file", GrantModeReadOnly, true},
		{"/Users/readonly", GrantModeReadWrite, false}, // Can't write to read-only
		{"/Users/readonly/file", GrantModeReadWrite, false},
	}

	for _, tc := range tests {
		result := service.IsGranted(tc.path, tc.mode)
		if result != tc.expected {
			t.Errorf("IsGranted(%s, %s) = %v, want %v", tc.path, tc.mode, result, tc.expected)
		}
	}
}

func TestGrantService_GetGrant(t *testing.T) {
	service := &GrantService{
		filePath: "/dev/null",
	}
	service.Load()

	service.Grant("/test/path", GrantModeReadWrite)

	// Exact match
	grant := service.GetGrant("/test/path")
	if grant == nil {
		t.Fatal("expected to find grant")
	}
	if grant.Mode != GrantModeReadWrite {
		t.Errorf("expected readwrite mode")
	}

	// No match
	grant = service.GetGrant("/other/path")
	if grant != nil {
		t.Error("expected nil for ungranted path")
	}
}

func TestIsUnderPath(t *testing.T) {
	tests := []struct {
		checkPath string
		basePath  string
		expected  bool
	}{
		{"/Users/alex/Code", "/Users/alex/Code", true},
		{"/Users/alex/Code/project", "/Users/alex/Code", true},
		{"/Users/alex/Code/project/src", "/Users/alex/Code", true},
		{"/Users/alex/Codebase", "/Users/alex/Code", false},
		{"/Users/alex", "/Users/alex/Code", false},
		{"/Other/path", "/Users/alex/Code", false},
	}

	for _, tc := range tests {
		result := isUnderPath(tc.checkPath, tc.basePath)
		if result != tc.expected {
			t.Errorf("isUnderPath(%s, %s) = %v, want %v", tc.checkPath, tc.basePath, result, tc.expected)
		}
	}
}

func TestResolveProjectMounts(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		configDir string
		mounts    map[string]string
		wantErr   bool
		wantCount int
		checkPath string // path to verify
	}{
		{
			name:      "empty mounts",
			configDir: tmpDir,
			mounts:    nil,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:      "relative path",
			configDir: tmpDir,
			mounts:    map[string]string{".": "readwrite"},
			wantErr:   false,
			wantCount: 1,
			checkPath: tmpDir,
		},
		{
			name:      "relative subdir",
			configDir: tmpDir,
			mounts:    map[string]string{"./subdir": "readonly"},
			wantErr:   false,
			wantCount: 1,
			checkPath: filepath.Join(tmpDir, "subdir"),
		},
		{
			name:      "absolute path",
			configDir: tmpDir,
			mounts:    map[string]string{"/absolute/path": "readonly"},
			wantErr:   false,
			wantCount: 1,
			checkPath: "/absolute/path",
		},
		{
			name:      "invalid mode",
			configDir: tmpDir,
			mounts:    map[string]string{".": "invalid"},
			wantErr:   true,
		},
		{
			name:      "multiple mounts",
			configDir: tmpDir,
			mounts: map[string]string{
				".":             "readwrite",
				"../other":      "readonly",
				"/absolute/dir": "readonly",
			},
			wantErr:   false,
			wantCount: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mounts, err := ResolveProjectMounts(tc.configDir, tc.mounts)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(mounts) != tc.wantCount {
				t.Errorf("expected %d mounts, got %d", tc.wantCount, len(mounts))
			}
			if tc.checkPath != "" && len(mounts) > 0 {
				found := false
				for _, m := range mounts {
					if m.Path == tc.checkPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to find path %s in mounts", tc.checkPath)
				}
			}
		})
	}
}

func TestResolveProjectMounts_HomePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory available")
	}

	mounts, err := ResolveProjectMounts("/tmp", map[string]string{
		"~/Documents": "readonly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount, got %d", len(mounts))
	}

	expectedPath := filepath.Join(homeDir, "Documents")
	if mounts[0].Path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, mounts[0].Path)
	}
}

func TestValidateProjectMounts(t *testing.T) {
	service := &GrantService{filePath: "/dev/null"}
	service.Load()
	service.Grant("/allowed/path", GrantModeReadWrite)
	service.Grant("/readonly/path", GrantModeReadOnly)

	tests := []struct {
		name           string
		mounts         []ConfigMount
		wantViolations int
	}{
		{
			name:           "empty mounts",
			mounts:         nil,
			wantViolations: 0,
		},
		{
			name: "all allowed",
			mounts: []ConfigMount{
				{Path: "/allowed/path", Mode: GrantModeReadWrite},
				{Path: "/readonly/path", Mode: GrantModeReadOnly},
			},
			wantViolations: 0,
		},
		{
			name: "not granted",
			mounts: []ConfigMount{
				{Path: "/not/granted", Mode: GrantModeReadWrite},
			},
			wantViolations: 1,
		},
		{
			name: "readonly request on readonly grant",
			mounts: []ConfigMount{
				{Path: "/readonly/path", Mode: GrantModeReadOnly},
			},
			wantViolations: 0,
		},
		{
			name: "readwrite request on readonly grant",
			mounts: []ConfigMount{
				{Path: "/readonly/path", Mode: GrantModeReadWrite},
			},
			wantViolations: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			violations := ValidateProjectMounts(tc.mounts, service)
			if len(violations) != tc.wantViolations {
				t.Errorf("expected %d violations, got %d: %v", tc.wantViolations, len(violations), violations)
			}
		})
	}
}

func TestMergeMounts(t *testing.T) {
	service := &GrantService{filePath: "/dev/null"}
	service.Load()
	service.Grant("/Users/test/project", GrantModeReadWrite)
	service.Grant("/Users/test/docs", GrantModeReadWrite)

	tests := []struct {
		name          string
		cliMounts     map[string]GrantMode
		projectMounts []ConfigMount
		checkPath     string
		expectMode    GrantMode
		expectSource  string
	}{
		{
			name:          "grant only",
			cliMounts:     nil,
			projectMounts: nil,
			checkPath:     "/Users/test/project",
			expectMode:    GrantModeReadWrite,
			expectSource:  "grants",
		},
		{
			name:      "project restricts grant",
			cliMounts: nil,
			projectMounts: []ConfigMount{
				{Path: "/Users/test/project", Mode: GrantModeReadOnly},
			},
			checkPath:    "/Users/test/project",
			expectMode:   GrantModeReadOnly,
			expectSource: "project",
		},
		{
			name:      "cli restricts grant",
			cliMounts: map[string]GrantMode{"/Users/test/project": GrantModeReadOnly},
			projectMounts: []ConfigMount{
				{Path: "/Users/test/project", Mode: GrantModeReadWrite},
			},
			checkPath:    "/Users/test/project",
			expectMode:   GrantModeReadOnly,
			expectSource: "cli",
		},
		{
			name:      "project cannot grant new access",
			cliMounts: nil,
			projectMounts: []ConfigMount{
				{Path: "/Users/test/not-granted", Mode: GrantModeReadWrite},
			},
			checkPath:    "/Users/test/not-granted",
			expectMode:   "", // Not in result
			expectSource: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := MergeMounts(tc.cliMounts, tc.projectMounts, service)

			var found *MergedMount
			for i := range result {
				if result[i].Path == tc.checkPath {
					found = &result[i]
					break
				}
			}

			if tc.expectMode == "" {
				if found != nil {
					t.Errorf("expected path %s not to be in result", tc.checkPath)
				}
				return
			}

			if found == nil {
				t.Fatalf("expected to find path %s in result", tc.checkPath)
			}
			if found.Mode != tc.expectMode {
				t.Errorf("expected mode %s, got %s", tc.expectMode, found.Mode)
			}
			if found.Source != tc.expectSource {
				t.Errorf("expected source %s, got %s", tc.expectSource, found.Source)
			}
		})
	}
}
