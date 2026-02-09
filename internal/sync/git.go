// Package sync provides git-based synchronization for sandbox state.
package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alexcabrera/ayo/internal/paths"
)

// sandboxDirOverride allows tests to override the sandbox directory.
var sandboxDirOverride string

// SandboxDir returns the sandbox state directory.
// Dev mode: {repo}/.local/share/ayo/sandbox
// Production: ~/.local/share/ayo/sandbox
func SandboxDir() string {
	if sandboxDirOverride != "" {
		return sandboxDirOverride
	}
	return filepath.Join(paths.DataDir(), "sandbox")
}

// HomesDir returns the directory for agent home directories within sandbox.
func HomesDir() string {
	return filepath.Join(SandboxDir(), "homes")
}

// SharedDir returns the shared files directory within sandbox.
func SharedDir() string {
	return filepath.Join(SandboxDir(), "shared")
}

// IsInitialized checks if the sandbox directory has been initialized as a git repo.
func IsInitialized() bool {
	gitDir := filepath.Join(SandboxDir(), ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// Init initializes the sandbox directory as a git repository.
// Creates the directory structure and initial commit.
func Init() error {
	sandboxDir := SandboxDir()

	// Create sandbox directory if it doesn't exist
	if err := os.MkdirAll(sandboxDir, 0755); err != nil {
		return fmt.Errorf("create sandbox directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{
		HomesDir(),
		SharedDir(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
		// Create .gitkeep to ensure empty directories are tracked
		gitkeep := filepath.Join(dir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			if err := os.WriteFile(gitkeep, []byte(""), 0644); err != nil {
				return fmt.Errorf("create .gitkeep in %s: %w", dir, err)
			}
		}
	}

	// Create .gitignore
	gitignorePath := filepath.Join(sandboxDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignore := `# Large/binary files
*.log
*.tmp
**/node_modules/
**/.cache/
**/__pycache__/
**/*.pyc
**/.venv/
**/venv/
**/.npm/
**/.yarn/
`
		if err := os.WriteFile(gitignorePath, []byte(gitignore), 0644); err != nil {
			return fmt.Errorf("create .gitignore: %w", err)
		}
	}

	// Initialize git repo if not already initialized
	if !IsInitialized() {
		if err := runGit(sandboxDir, "init"); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
	}

	// Configure git user for this repo (local config)
	if err := runGit(sandboxDir, "config", "user.email", "ayo@local"); err != nil {
		return fmt.Errorf("git config user.email: %w", err)
	}
	if err := runGit(sandboxDir, "config", "user.name", "ayo"); err != nil {
		return fmt.Errorf("git config user.name: %w", err)
	}

	// Check if there are any commits
	if err := runGit(sandboxDir, "rev-parse", "HEAD"); err != nil {
		// No commits yet - create initial commit
		if err := runGit(sandboxDir, "add", "-A"); err != nil {
			return fmt.Errorf("git add: %w", err)
		}
		if err := runGit(sandboxDir, "commit", "-m", "Initial sandbox state"); err != nil {
			return fmt.Errorf("git commit: %w", err)
		}
	}

	return nil
}

// Commit creates a new commit with all changes in the sandbox directory.
func Commit(message string) error {
	if !IsInitialized() {
		return fmt.Errorf("sandbox not initialized as git repo")
	}

	sandboxDir := SandboxDir()

	// Stage all changes
	if err := runGit(sandboxDir, "add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Check if there are staged changes
	output, err := runGitOutput(sandboxDir, "diff", "--cached", "--quiet")
	if err == nil {
		// No changes to commit
		return nil
	}
	_ = output // Suppress unused variable warning

	// Commit changes
	if err := runGit(sandboxDir, "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}

// GetBranch returns the current branch name.
func GetBranch() (string, error) {
	if !IsInitialized() {
		return "", fmt.Errorf("sandbox not initialized as git repo")
	}

	output, err := runGitOutput(SandboxDir(), "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("get current branch: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// CreateMachineBranch creates a branch for the current machine.
// Branch name: machines/{hostname}
func CreateMachineBranch() (string, error) {
	if !IsInitialized() {
		return "", fmt.Errorf("sandbox not initialized as git repo")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("get hostname: %w", err)
	}

	// Sanitize hostname for branch name
	hostname = strings.ReplaceAll(hostname, " ", "-")
	hostname = strings.ToLower(hostname)

	branchName := fmt.Sprintf("machines/%s", hostname)

	// Check if branch already exists
	if err := runGit(SandboxDir(), "rev-parse", "--verify", branchName); err == nil {
		// Branch exists, just checkout
		if err := runGit(SandboxDir(), "checkout", branchName); err != nil {
			return "", fmt.Errorf("checkout branch %s: %w", branchName, err)
		}
		return branchName, nil
	}

	// Create and checkout new branch
	if err := runGit(SandboxDir(), "checkout", "-b", branchName); err != nil {
		return "", fmt.Errorf("create branch %s: %w", branchName, err)
	}

	return branchName, nil
}

// HasChanges returns true if there are uncommitted changes in the sandbox.
func HasChanges() (bool, error) {
	if !IsInitialized() {
		return false, fmt.Errorf("sandbox not initialized as git repo")
	}

	// Check for any changes (staged or unstaged)
	if err := runGit(SandboxDir(), "diff", "--quiet"); err != nil {
		return true, nil // Has unstaged changes
	}
	if err := runGit(SandboxDir(), "diff", "--cached", "--quiet"); err != nil {
		return true, nil // Has staged changes
	}

	// Check for untracked files
	output, err := runGitOutput(SandboxDir(), "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}

	return strings.TrimSpace(output) != "", nil
}

// runGit executes a git command in the specified directory.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// runGitOutput executes a git command and returns its output.
func runGitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	return string(output), err
}

// SyncStatus represents the current sync state.
type SyncStatus struct {
	LocalBranch      string
	RemoteConfigured bool
	RemoteName       string
	RemoteURL        string
	Ahead            int
	Behind           int
	HasChanges       bool
	LastCommit       string
	LastCommitTime   string
}

// HasRemote returns true if a git remote is configured.
func HasRemote() bool {
	if !IsInitialized() {
		return false
	}

	output, err := runGitOutput(SandboxDir(), "remote")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) != ""
}

// GetRemote returns the name and URL of the first configured remote.
func GetRemote() (name, url string, err error) {
	if !IsInitialized() {
		return "", "", fmt.Errorf("sandbox not initialized as git repo")
	}

	output, err := runGitOutput(SandboxDir(), "remote")
	if err != nil {
		return "", "", fmt.Errorf("list remotes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "", "", nil // No remote configured
	}

	name = lines[0]
	url, err = runGitOutput(SandboxDir(), "remote", "get-url", name)
	if err != nil {
		return name, "", nil // Remote exists but no URL
	}

	return name, strings.TrimSpace(url), nil
}

// AddRemote adds a git remote to the sandbox repository.
func AddRemote(name, url string) error {
	if !IsInitialized() {
		return fmt.Errorf("sandbox not initialized as git repo")
	}

	return runGit(SandboxDir(), "remote", "add", name, url)
}

// SetRemoteURL updates the URL of an existing remote.
func SetRemoteURL(name, url string) error {
	if !IsInitialized() {
		return fmt.Errorf("sandbox not initialized as git repo")
	}

	return runGit(SandboxDir(), "remote", "set-url", name, url)
}

// Fetch fetches from the remote without merging.
func Fetch() error {
	if !IsInitialized() {
		return fmt.Errorf("sandbox not initialized as git repo")
	}

	if !HasRemote() {
		return fmt.Errorf("no remote configured")
	}

	return runGit(SandboxDir(), "fetch", "--all")
}

// Push commits and pushes to the remote.
// If message is empty, uses a default message.
func Push(message string) error {
	if !IsInitialized() {
		return fmt.Errorf("sandbox not initialized as git repo")
	}

	if !HasRemote() {
		return fmt.Errorf("no remote configured")
	}

	sandboxDir := SandboxDir()

	// Get current branch
	branch, err := GetBranch()
	if err != nil {
		return fmt.Errorf("get branch: %w", err)
	}

	// Commit any changes
	hasChanges, err := HasChanges()
	if err != nil {
		return fmt.Errorf("check changes: %w", err)
	}

	if hasChanges {
		if message == "" {
			message = fmt.Sprintf("Sync from %s", getHostname())
		}
		if err := Commit(message); err != nil {
			return fmt.Errorf("commit: %w", err)
		}
	}

	// Get remote name
	remoteName, _, err := GetRemote()
	if err != nil || remoteName == "" {
		return fmt.Errorf("get remote: %w", err)
	}

	// Push with set-upstream if needed
	if err := runGit(sandboxDir, "push", "-u", remoteName, branch); err != nil {
		return fmt.Errorf("push: %w", err)
	}

	return nil
}

// Pull fetches and merges from the remote.
func Pull() error {
	if !IsInitialized() {
		return fmt.Errorf("sandbox not initialized as git repo")
	}

	if !HasRemote() {
		return fmt.Errorf("no remote configured")
	}

	sandboxDir := SandboxDir()

	// First commit any local changes to avoid losing work
	hasChanges, err := HasChanges()
	if err != nil {
		return fmt.Errorf("check changes: %w", err)
	}

	if hasChanges {
		if err := Commit(fmt.Sprintf("Auto-save before pull from %s", getHostname())); err != nil {
			return fmt.Errorf("commit before pull: %w", err)
		}
	}

	// Fetch first
	if err := Fetch(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	// Get remote and branch
	remoteName, _, err := GetRemote()
	if err != nil || remoteName == "" {
		return fmt.Errorf("get remote: %w", err)
	}

	branch, err := GetBranch()
	if err != nil {
		return fmt.Errorf("get branch: %w", err)
	}

	// Check if remote branch exists
	remoteBranch := fmt.Sprintf("%s/%s", remoteName, branch)
	if err := runGit(sandboxDir, "rev-parse", "--verify", remoteBranch); err != nil {
		// Remote branch doesn't exist, nothing to pull
		return nil
	}

	// Pull with rebase to keep history clean
	if err := runGit(sandboxDir, "pull", "--rebase", remoteName, branch); err != nil {
		// If rebase fails, try regular merge
		if err := runGit(sandboxDir, "rebase", "--abort"); err != nil {
			// Ignore abort errors
		}
		if err := runGit(sandboxDir, "pull", remoteName, branch); err != nil {
			return fmt.Errorf("pull: %w", err)
		}
	}

	return nil
}

// Status returns the current sync status.
func Status() (*SyncStatus, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("sandbox not initialized as git repo")
	}

	status := &SyncStatus{}

	// Get branch
	branch, err := GetBranch()
	if err == nil {
		status.LocalBranch = branch
	}

	// Check for remote
	remoteName, remoteURL, err := GetRemote()
	if err == nil && remoteName != "" {
		status.RemoteConfigured = true
		status.RemoteName = remoteName
		status.RemoteURL = remoteURL
	}

	// Check for changes
	hasChanges, err := HasChanges()
	if err == nil {
		status.HasChanges = hasChanges
	}

	// Get last commit info
	output, err := runGitOutput(SandboxDir(), "log", "-1", "--format=%H|%ar")
	if err == nil {
		parts := strings.SplitN(strings.TrimSpace(output), "|", 2)
		if len(parts) >= 1 {
			status.LastCommit = parts[0][:min(8, len(parts[0]))]
		}
		if len(parts) >= 2 {
			status.LastCommitTime = parts[1]
		}
	}

	// Calculate ahead/behind if remote is configured
	if status.RemoteConfigured && status.LocalBranch != "" {
		remoteBranch := fmt.Sprintf("%s/%s", remoteName, status.LocalBranch)
		
		// First fetch to get latest
		_ = runGit(SandboxDir(), "fetch", remoteName)

		// Check if remote branch exists
		if err := runGit(SandboxDir(), "rev-parse", "--verify", remoteBranch); err == nil {
			// Get ahead/behind counts
			output, err := runGitOutput(SandboxDir(), "rev-list", "--left-right", "--count",
				fmt.Sprintf("%s...%s", status.LocalBranch, remoteBranch))
			if err == nil {
				parts := strings.Fields(strings.TrimSpace(output))
				if len(parts) >= 2 {
					fmt.Sscanf(parts[0], "%d", &status.Ahead)
					fmt.Sscanf(parts[1], "%d", &status.Behind)
				}
			}
		}
	}

	return status, nil
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
