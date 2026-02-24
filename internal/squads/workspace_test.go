package squads

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitEmptyWorkspace(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")

	err := initEmptyWorkspace(workspaceDir)
	if err != nil {
		t.Fatalf("initEmptyWorkspace failed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(workspaceDir)
	if err != nil {
		t.Fatalf("workspace directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("workspace is not a directory")
	}

	// Verify README exists
	readme := filepath.Join(workspaceDir, "README.md")
	if _, err := os.Stat(readme); err != nil {
		t.Error("README.md not created")
	}
}

func TestInitCopyWorkspace(t *testing.T) {
	// Create temp source directory
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	subDir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create temp workspace directory
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")

	err := initCopyWorkspace(workspaceDir, srcDir)
	if err != nil {
		t.Fatalf("initCopyWorkspace failed: %v", err)
	}

	// Verify files were copied
	if _, err := os.Stat(filepath.Join(workspaceDir, "test.txt")); err != nil {
		t.Error("test.txt not copied")
	}
	if _, err := os.Stat(filepath.Join(workspaceDir, "subdir", "nested.txt")); err != nil {
		t.Error("subdir/nested.txt not copied")
	}

	// Verify content
	content, err := os.ReadFile(filepath.Join(workspaceDir, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello" {
		t.Errorf("content mismatch: got %q, want %q", string(content), "hello")
	}
}

func TestInitLinkWorkspace(t *testing.T) {
	// Create temp source directory
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create temp workspace directory (parent only)
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")

	err := initLinkWorkspace(workspaceDir, srcDir)
	if err != nil {
		t.Fatalf("initLinkWorkspace failed: %v", err)
	}

	// Verify symlink was created
	info, err := os.Lstat(workspaceDir)
	if err != nil {
		t.Fatalf("workspace not created: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("workspace is not a symlink")
	}

	// Verify file is accessible through symlink
	content, err := os.ReadFile(filepath.Join(workspaceDir, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello" {
		t.Errorf("content mismatch: got %q, want %q", string(content), "hello")
	}
}

func TestCopyDir_SkipsGit(t *testing.T) {
	// Create temp source directory with .git
	srcDir := t.TempDir()
	gitDir := filepath.Join(srcDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte("git stuff"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "code.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Copy to temp destination
	dstDir := filepath.Join(t.TempDir(), "dst")
	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Verify .git was NOT copied
	if _, err := os.Stat(filepath.Join(dstDir, ".git")); !os.IsNotExist(err) {
		t.Error(".git directory should not be copied")
	}

	// Verify code.go was copied
	if _, err := os.Stat(filepath.Join(dstDir, "code.go")); err != nil {
		t.Error("code.go should be copied")
	}
}

func TestWorkspaceInit_Validation(t *testing.T) {
	tests := []struct {
		name    string
		init    WorkspaceInit
		wantErr bool
	}{
		{
			name:    "empty source for git",
			init:    WorkspaceInit{Type: WorkspaceInitGit, Source: ""},
			wantErr: true,
		},
		{
			name:    "empty source for copy",
			init:    WorkspaceInit{Type: WorkspaceInitCopy, Source: ""},
			wantErr: true,
		},
		{
			name:    "empty source for link",
			init:    WorkspaceInit{Type: WorkspaceInitLink, Source: ""},
			wantErr: true,
		},
		{
			name:    "nonexistent source for copy",
			init:    WorkspaceInit{Type: WorkspaceInitCopy, Source: "/nonexistent/path"},
			wantErr: true,
		},
		{
			name:    "nonexistent source for link",
			init:    WorkspaceInit{Type: WorkspaceInitLink, Source: "/nonexistent/path"},
			wantErr: true,
		},
	}

	tmpDir := t.TempDir()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspaceDir := filepath.Join(tmpDir, tt.name)
			var err error

			switch tt.init.Type {
			case WorkspaceInitGit:
				err = initGitWorkspace(workspaceDir, tt.init.Source, tt.init.Branch)
			case WorkspaceInitCopy:
				err = initCopyWorkspace(workspaceDir, tt.init.Source)
			case WorkspaceInitLink:
				err = initLinkWorkspace(workspaceDir, tt.init.Source)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestInitCopyWorkspace_FileNotDir(t *testing.T) {
	// Source is a file, not a directory
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(srcFile, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err := initCopyWorkspace(workspaceDir, srcFile)
	if err == nil {
		t.Error("expected error when source is a file, not directory")
	}
}

func TestInitLinkWorkspace_FileNotDir(t *testing.T) {
	// Source is a file, not a directory
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(srcFile, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err := initLinkWorkspace(workspaceDir, srcFile)
	if err == nil {
		t.Error("expected error when source is a file, not directory")
	}
}

func TestCopyFile(t *testing.T) {
	// Test copyFile directly
	tmpDir := t.TempDir()

	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "subdir", "dst.txt")

	content := "test content here"
	if err := os.WriteFile(srcFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// copyFile should create parent directory
	err := copyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify content
	data, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("content mismatch: got %q, want %q", string(data), content)
	}
}

func TestWorkspaceInitType_Constants(t *testing.T) {
	// Verify the constant values are as expected
	if WorkspaceInitEmpty != "empty" {
		t.Errorf("WorkspaceInitEmpty = %q, want %q", WorkspaceInitEmpty, "empty")
	}
	if WorkspaceInitGit != "git" {
		t.Errorf("WorkspaceInitGit = %q, want %q", WorkspaceInitGit, "git")
	}
	if WorkspaceInitCopy != "copy" {
		t.Errorf("WorkspaceInitCopy = %q, want %q", WorkspaceInitCopy, "copy")
	}
	if WorkspaceInitLink != "link" {
		t.Errorf("WorkspaceInitLink = %q, want %q", WorkspaceInitLink, "link")
	}
}

func TestInitCopyWorkspace_HomeTilde(t *testing.T) {
	// Test that tilde expansion works
	// We can't actually test ~ expansion without a real home dir,
	// but we can verify the code path handles it

	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")

	// This will fail because ~/nonexistent doesn't exist, but it tests the tilde expansion path
	err := initCopyWorkspace(workspaceDir, "~/nonexistent-test-dir-12345")
	if err == nil {
		t.Error("expected error for nonexistent tilde path")
	}
}

func TestInitLinkWorkspace_HomeTilde(t *testing.T) {
	// Test that tilde expansion works
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")

	err := initLinkWorkspace(workspaceDir, "~/nonexistent-test-dir-12345")
	if err == nil {
		t.Error("expected error for nonexistent tilde path")
	}
}

func TestInitCopyWorkspace_OverwritesExisting(t *testing.T) {
	// Create source with content
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "new.txt"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create workspace with different content
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspaceDir, "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Copy should overwrite
	err := initCopyWorkspace(workspaceDir, srcDir)
	if err != nil {
		t.Fatalf("initCopyWorkspace failed: %v", err)
	}

	// Old file should be gone
	if _, err := os.Stat(filepath.Join(workspaceDir, "old.txt")); !os.IsNotExist(err) {
		t.Error("old.txt should have been removed")
	}

	// New file should exist
	if _, err := os.Stat(filepath.Join(workspaceDir, "new.txt")); err != nil {
		t.Error("new.txt should exist")
	}
}
