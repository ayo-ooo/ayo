package squads

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
)

// WorkspaceInitType represents the type of workspace initialization.
type WorkspaceInitType string

const (
	// WorkspaceInitEmpty creates an empty workspace directory.
	WorkspaceInitEmpty WorkspaceInitType = "empty"

	// WorkspaceInitGit clones a git repository into the workspace.
	WorkspaceInitGit WorkspaceInitType = "git"

	// WorkspaceInitCopy copies a directory into the workspace.
	WorkspaceInitCopy WorkspaceInitType = "copy"

	// WorkspaceInitLink creates a symlink to an existing directory.
	WorkspaceInitLink WorkspaceInitType = "link"
)

// WorkspaceInit specifies how to initialize a squad workspace.
type WorkspaceInit struct {
	// Type is the initialization method.
	Type WorkspaceInitType

	// Source is the source path or URL for git/copy/link types.
	// Ignored for "empty" type.
	Source string

	// Branch is the git branch to clone (optional, for git type only).
	Branch string
}

// InitWorkspace initializes the workspace for a squad.
func InitWorkspace(squadName string, init WorkspaceInit) error {
	workspaceDir := paths.SquadWorkspaceDir(squadName)

	// Ensure parent directory exists
	squadDir := paths.SquadDir(squadName)
	if err := os.MkdirAll(squadDir, 0o755); err != nil {
		return fmt.Errorf("create squad directory: %w", err)
	}

	debug.Log("initializing workspace", "squad", squadName, "type", init.Type, "source", init.Source)

	switch init.Type {
	case WorkspaceInitEmpty, "":
		return initEmptyWorkspace(workspaceDir)

	case WorkspaceInitGit:
		return initGitWorkspace(workspaceDir, init.Source, init.Branch)

	case WorkspaceInitCopy:
		return initCopyWorkspace(workspaceDir, init.Source)

	case WorkspaceInitLink:
		return initLinkWorkspace(workspaceDir, init.Source)

	default:
		return fmt.Errorf("unknown workspace init type: %s", init.Type)
	}
}

// initEmptyWorkspace creates an empty workspace directory.
func initEmptyWorkspace(workspaceDir string) error {
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		return fmt.Errorf("create workspace directory: %w", err)
	}

	// Create a README to indicate this is a squad workspace
	readme := filepath.Join(workspaceDir, "README.md")
	content := `# Squad Workspace

This is the shared workspace for squad agents.

All agents in the squad can read and write files here.
`
	if err := os.WriteFile(readme, []byte(content), 0o644); err != nil {
		debug.Log("failed to create workspace README", "error", err)
		// Not fatal, continue
	}

	return nil
}

// initGitWorkspace clones a git repository into the workspace.
func initGitWorkspace(workspaceDir, repoURL, branch string) error {
	if repoURL == "" {
		return fmt.Errorf("git URL is required")
	}

	// Remove existing workspace if it exists
	if err := os.RemoveAll(workspaceDir); err != nil {
		return fmt.Errorf("remove existing workspace: %w", err)
	}

	// Build git clone command
	args := []string{"clone"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, "--depth", "1") // Shallow clone by default
	args = append(args, repoURL, workspaceDir)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	debug.Log("cloning git repository", "url", repoURL, "branch", branch)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Configure git for squad use
	return configureGitWorkspace(workspaceDir)
}

// configureGitWorkspace sets up git configuration for the workspace.
func configureGitWorkspace(workspaceDir string) error {
	// Set git user for commits (can be overridden by agents)
	configs := map[string]string{
		"user.name":  "ayo-squad",
		"user.email": "squad@ayo.local",
	}

	for key, value := range configs {
		cmd := exec.Command("git", "-C", workspaceDir, "config", key, value)
		if err := cmd.Run(); err != nil {
			debug.Log("failed to set git config", "key", key, "error", err)
			// Not fatal, continue
		}
	}

	return nil
}

// initCopyWorkspace copies a directory into the workspace.
func initCopyWorkspace(workspaceDir, srcDir string) error {
	if srcDir == "" {
		return fmt.Errorf("source directory is required")
	}

	// Expand home directory
	if srcDir[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		srcDir = filepath.Join(home, srcDir[1:])
	}

	// Resolve to absolute path
	srcDir, err := filepath.Abs(srcDir)
	if err != nil {
		return fmt.Errorf("resolve source path: %w", err)
	}

	// Verify source exists and is a directory
	info, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("source directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source is not a directory: %s", srcDir)
	}

	// Remove existing workspace if it exists
	if err := os.RemoveAll(workspaceDir); err != nil {
		return fmt.Errorf("remove existing workspace: %w", err)
	}

	debug.Log("copying directory", "src", srcDir, "dst", workspaceDir)

	// Copy the directory
	return copyDir(srcDir, workspaceDir)
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		// Copy file
		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Preserve file mode
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// initLinkWorkspace creates a symlink to an existing directory.
func initLinkWorkspace(workspaceDir, srcDir string) error {
	if srcDir == "" {
		return fmt.Errorf("source directory is required")
	}

	// Expand home directory
	if srcDir[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		srcDir = filepath.Join(home, srcDir[1:])
	}

	// Resolve to absolute path
	srcDir, err := filepath.Abs(srcDir)
	if err != nil {
		return fmt.Errorf("resolve source path: %w", err)
	}

	// Verify source exists and is a directory
	info, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("source directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source is not a directory: %s", srcDir)
	}

	// Remove existing workspace if it exists
	if err := os.RemoveAll(workspaceDir); err != nil {
		return fmt.Errorf("remove existing workspace: %w", err)
	}

	debug.Log("creating symlink", "src", srcDir, "dst", workspaceDir)

	// Create symlink
	if err := os.Symlink(srcDir, workspaceDir); err != nil {
		return fmt.Errorf("create symlink: %w", err)
	}

	return nil
}

// WorkspaceExists returns true if the workspace directory exists.
func WorkspaceExists(squadName string) bool {
	workspaceDir := paths.SquadWorkspaceDir(squadName)
	info, err := os.Stat(workspaceDir)
	return err == nil && info.IsDir()
}

// WorkspaceIsLink returns true if the workspace is a symlink.
func WorkspaceIsLink(squadName string) bool {
	workspaceDir := paths.SquadWorkspaceDir(squadName)
	info, err := os.Lstat(workspaceDir)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// WorkspaceIsGit returns true if the workspace is a git repository.
func WorkspaceIsGit(squadName string) bool {
	workspaceDir := paths.SquadWorkspaceDir(squadName)
	gitDir := filepath.Join(workspaceDir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}
