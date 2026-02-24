package hostwrite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateHostPath(t *testing.T) {
	cfg := ToolConfig{
		BaseDir:      "/home/user",
		BlockedPaths: DefaultBlockedPaths(),
	}

	tests := []struct {
		name    string
		relPath string
		fullPath string
		wantErr bool
	}{
		{
			name:     "valid path",
			relPath:  "Projects/app/main.go",
			fullPath: "/home/user/Projects/app/main.go",
			wantErr:  false,
		},
		{
			name:     "path traversal",
			relPath:  "../etc/passwd",
			fullPath: "/home/user/../etc/passwd",
			wantErr:  true,
		},
		{
			name:     "blocked ssh key",
			relPath:  ".ssh/id_rsa",
			fullPath: "/home/user/.ssh/id_rsa",
			wantErr:  true,
		},
		{
			name:     "blocked pem file",
			relPath:  "secrets/cert.pem",
			fullPath: "/home/user/secrets/cert.pem",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHostPath(tt.relPath, tt.fullPath, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHostPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateHostPath_AllowedPaths(t *testing.T) {
	cfg := ToolConfig{
		BaseDir:      "/home/user",
		BlockedPaths: DefaultBlockedPaths(),
		AllowedPaths: []string{"Projects/", "Documents/"},
	}

	tests := []struct {
		name    string
		relPath string
		wantErr bool
	}{
		{
			name:    "allowed projects",
			relPath: "Projects/app/main.go",
			wantErr: false,
		},
		{
			name:    "allowed documents",
			relPath: "Documents/notes.txt",
			wantErr: false,
		},
		{
			name:    "not allowed desktop",
			relPath: "Desktop/file.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := filepath.Join(cfg.BaseDir, tt.relPath)
			err := validateHostPath(tt.relPath, fullPath, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHostPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name     string
		baseDir  string
		path     string
		expected string
	}{
		{
			name:     "relative path",
			baseDir:  "/home/user",
			path:     "Projects/app/main.go",
			expected: "/home/user/Projects/app/main.go",
		},
		{
			name:     "absolute path",
			baseDir:  "/home/user",
			path:     "/tmp/file.txt",
			expected: "/tmp/file.txt",
		},
		{
			name:     "path with dots",
			baseDir:  "/home/user",
			path:     "Projects/./app/main.go",
			expected: "/home/user/Projects/app/main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePath(tt.baseDir, tt.path)
			if result != tt.expected {
				t.Errorf("resolvePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWriteAndDeleteFile(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Test create
	testFile := filepath.Join(tmpDir, "subdir", "test.txt")
	content := "Hello, World!"
	
	err := writeFile(testFile, content)
	if err != nil {
		t.Fatalf("writeFile() error = %v", err)
	}
	
	// Verify file exists
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != content {
		t.Errorf("content = %v, want %v", string(data), content)
	}
	
	// Test update
	newContent := "Updated content"
	err = writeFile(testFile, newContent)
	if err != nil {
		t.Fatalf("writeFile() update error = %v", err)
	}
	
	data, _ = os.ReadFile(testFile)
	if string(data) != newContent {
		t.Errorf("updated content = %v, want %v", string(data), newContent)
	}
	
	// Test delete
	err = deleteFile(testFile)
	if err != nil {
		t.Fatalf("deleteFile() error = %v", err)
	}
	
	// Verify file deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestDeleteFile_NotExists(t *testing.T) {
	err := deleteFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// MockApprovalHandler implements ApprovalHandler for testing.
type MockApprovalHandler struct {
	Approved bool
	Requests []ApprovalRequest
}

func (m *MockApprovalHandler) RequestApproval(ctx context.Context, req ApprovalRequest) (bool, error) {
	m.Requests = append(m.Requests, req)
	return m.Approved, nil
}

func TestDefaultBlockedPaths(t *testing.T) {
	blocked := DefaultBlockedPaths()
	if len(blocked) == 0 {
		t.Error("expected default blocked paths")
	}
	
	// Should include ssh
	hasSSH := false
	for _, p := range blocked {
		if p == ".ssh/*" {
			hasSSH = true
			break
		}
	}
	if !hasSSH {
		t.Error("expected .ssh/* in blocked paths")
	}
}
