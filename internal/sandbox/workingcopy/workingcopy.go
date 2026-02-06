// Package workingcopy provides a copy-based project synchronization model for sandboxes.
// Instead of mounting the host project directory directly (which allows real-time changes),
// the working copy model creates a copy inside the sandbox that the agent works on.
// Changes can then be synced back to the host on demand.
package workingcopy

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
)

// WorkingCopy represents a project copy inside a sandbox.
type WorkingCopy struct {
	// HostPath is the original project path on the host.
	HostPath string

	// SandboxPath is the path inside the sandbox where the copy lives.
	SandboxPath string

	// SandboxID is the ID of the sandbox containing this working copy.
	SandboxID string

	// CreatedAt is when the working copy was created.
	CreatedAt time.Time

	// LastSyncedAt is when the working copy was last synced to host.
	LastSyncedAt time.Time

	// IgnorePatterns are patterns for files to exclude from sync.
	IgnorePatterns []string
}

// Manager handles working copy operations.
type Manager struct {
	provider providers.SandboxProvider
}

// NewManager creates a new working copy manager.
func NewManager(provider providers.SandboxProvider) *Manager {
	return &Manager{provider: provider}
}

// Create creates a new working copy in the sandbox by copying the host project.
func (m *Manager) Create(ctx context.Context, sandboxID, hostPath, sandboxPath string, ignorePatterns []string) (*WorkingCopy, error) {
	// Resolve absolute path
	absHostPath, err := filepath.Abs(hostPath)
	if err != nil {
		return nil, fmt.Errorf("resolve host path: %w", err)
	}

	// Verify host path exists
	info, err := os.Stat(absHostPath)
	if err != nil {
		return nil, fmt.Errorf("host path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("host path must be a directory: %s", absHostPath)
	}

	// Default sandbox path
	if sandboxPath == "" {
		sandboxPath = "/workspace"
	}

	// Default ignore patterns
	if ignorePatterns == nil {
		ignorePatterns = DefaultIgnorePatterns()
	}

	// Create the directory in sandbox
	_, err = m.provider.Exec(ctx, sandboxID, providers.ExecOptions{
		Command: fmt.Sprintf("mkdir -p %s", sandboxPath),
	})
	if err != nil {
		return nil, fmt.Errorf("create sandbox directory: %w", err)
	}

	// Create tar of host directory
	tarData, err := createTar(absHostPath, ignorePatterns)
	if err != nil {
		return nil, fmt.Errorf("create tar: %w", err)
	}

	// Extract tar in sandbox
	_, err = m.provider.Exec(ctx, sandboxID, providers.ExecOptions{
		Command:    "tar",
		Args:       []string{"-xf", "-", "-C", sandboxPath},
		Stdin:      tarData,
		WorkingDir: sandboxPath,
	})
	if err != nil {
		return nil, fmt.Errorf("extract in sandbox: %w", err)
	}

	return &WorkingCopy{
		HostPath:       absHostPath,
		SandboxPath:    sandboxPath,
		SandboxID:      sandboxID,
		CreatedAt:      time.Now(),
		IgnorePatterns: ignorePatterns,
	}, nil
}

// Sync synchronizes changes from the sandbox working copy back to the host.
// It returns a list of files that were changed.
func (m *Manager) Sync(ctx context.Context, wc *WorkingCopy) ([]string, error) {
	// Get list of files in sandbox
	result, err := m.provider.Exec(ctx, wc.SandboxID, providers.ExecOptions{
		Command:    "find",
		Args:       []string{".", "-type", "f"},
		WorkingDir: wc.SandboxPath,
	})
	if err != nil {
		return nil, fmt.Errorf("list sandbox files: %w", err)
	}

	sandboxFiles := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	if len(sandboxFiles) == 1 && sandboxFiles[0] == "" {
		sandboxFiles = nil
	}

	// Create tar of sandbox directory
	tarResult, err := m.provider.Exec(ctx, wc.SandboxID, providers.ExecOptions{
		Command:    "tar",
		Args:       []string{"-cf", "-", "."},
		WorkingDir: wc.SandboxPath,
	})
	if err != nil {
		return nil, fmt.Errorf("create sandbox tar: %w", err)
	}

	// Extract to host
	changedFiles, err := extractTar([]byte(tarResult.Stdout), wc.HostPath, wc.IgnorePatterns)
	if err != nil {
		return nil, fmt.Errorf("extract to host: %w", err)
	}

	wc.LastSyncedAt = time.Now()

	return changedFiles, nil
}

// Diff returns a list of files that differ between host and sandbox.
func (m *Manager) Diff(ctx context.Context, wc *WorkingCopy) ([]FileDiff, error) {
	var diffs []FileDiff

	// Get file list and checksums from sandbox
	result, err := m.provider.Exec(ctx, wc.SandboxID, providers.ExecOptions{
		Command:    "sh",
		Args:       []string{"-c", "find . -type f -exec md5sum {} \\;"},
		WorkingDir: wc.SandboxPath,
	})
	if err != nil {
		return nil, fmt.Errorf("get sandbox checksums: %w", err)
	}

	sandboxChecksums := parseChecksums(result.Stdout)

	// Compare with host
	err = filepath.Walk(wc.HostPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(wc.HostPath, path)
		if shouldIgnore(relPath, wc.IgnorePatterns) {
			return nil
		}

		hostChecksum, err := fileChecksum(path)
		if err != nil {
			return nil
		}

		sandboxChecksum, exists := sandboxChecksums["./"+relPath]
		if !exists {
			diffs = append(diffs, FileDiff{
				Path:   relPath,
				Status: DiffStatusDeleted,
			})
		} else if hostChecksum != sandboxChecksum {
			diffs = append(diffs, FileDiff{
				Path:   relPath,
				Status: DiffStatusModified,
			})
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk host directory: %w", err)
	}

	// Check for new files in sandbox
	for path := range sandboxChecksums {
		relPath := strings.TrimPrefix(path, "./")
		if shouldIgnore(relPath, wc.IgnorePatterns) {
			continue
		}

		hostPath := filepath.Join(wc.HostPath, relPath)
		if _, err := os.Stat(hostPath); os.IsNotExist(err) {
			diffs = append(diffs, FileDiff{
				Path:   relPath,
				Status: DiffStatusAdded,
			})
		}
	}

	return diffs, nil
}

// Reset discards changes in the sandbox and restores from host.
func (m *Manager) Reset(ctx context.Context, wc *WorkingCopy) error {
	// Remove sandbox directory
	_, err := m.provider.Exec(ctx, wc.SandboxID, providers.ExecOptions{
		Command: fmt.Sprintf("rm -rf %s/*", wc.SandboxPath),
	})
	if err != nil {
		return fmt.Errorf("clean sandbox: %w", err)
	}

	// Re-copy from host
	tarData, err := createTar(wc.HostPath, wc.IgnorePatterns)
	if err != nil {
		return fmt.Errorf("create tar: %w", err)
	}

	_, err = m.provider.Exec(ctx, wc.SandboxID, providers.ExecOptions{
		Command:    "tar",
		Args:       []string{"-xf", "-", "-C", wc.SandboxPath},
		Stdin:      tarData,
		WorkingDir: wc.SandboxPath,
	})
	if err != nil {
		return fmt.Errorf("extract in sandbox: %w", err)
	}

	return nil
}

// DiffStatus represents the type of difference for a file.
type DiffStatus string

const (
	DiffStatusAdded    DiffStatus = "added"
	DiffStatusModified DiffStatus = "modified"
	DiffStatusDeleted  DiffStatus = "deleted"
)

// FileDiff represents a difference between host and sandbox.
type FileDiff struct {
	Path   string
	Status DiffStatus
}

// DefaultIgnorePatterns returns default patterns for files to ignore.
func DefaultIgnorePatterns() []string {
	return []string{
		".git",
		".git/**",
		"node_modules",
		"node_modules/**",
		".venv",
		".venv/**",
		"__pycache__",
		"__pycache__/**",
		"*.pyc",
		".DS_Store",
		"*.swp",
		"*.swo",
		"*~",
	}
}

// CreateTarFromDir creates a tar archive of the source directory.
// This is an exported version for use by other packages.
func CreateTarFromDir(source string, ignorePatterns []string) ([]byte, error) {
	return createTar(source, ignorePatterns)
}

// createTar creates a tar archive of the source directory.
func createTar(source string, ignorePatterns []string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		relPath, _ := filepath.Rel(source, path)
		if relPath == "." {
			return nil
		}

		if shouldIgnore(relPath, ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// extractTar extracts a tar archive to the destination.
func extractTar(tarData []byte, dest string, ignorePatterns []string) ([]string, error) {
	var changedFiles []string
	tr := tar.NewReader(bytes.NewReader(tarData))

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if shouldIgnore(header.Name, ignorePatterns) {
			continue
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return nil, err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return nil, err
			}
			file, err := os.Create(target)
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(file, tr); err != nil {
				file.Close()
				return nil, err
			}
			file.Close()
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return nil, err
			}
			changedFiles = append(changedFiles, header.Name)
		}
	}

	return changedFiles, nil
}

// shouldIgnore checks if a path matches any ignore pattern.
func shouldIgnore(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple matching - handle exact matches and glob patterns
		if pattern == path || strings.HasPrefix(path, pattern+"/") {
			return true
		}
		if strings.Contains(pattern, "*") {
			matched, _ := filepath.Match(pattern, path)
			if matched {
				return true
			}
			matched, _ = filepath.Match(pattern, filepath.Base(path))
			if matched {
				return true
			}
		}
	}
	return false
}

// parseChecksums parses md5sum output into a map of path -> checksum.
func parseChecksums(output string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result[parts[1]] = parts[0]
		}
	}
	return result
}

// fileChecksum calculates the MD5 checksum of a file.
func fileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
