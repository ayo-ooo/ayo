// Package sync provides utilities for synchronizing work products
// between sandboxes and the user's filesystem.
package sync

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
)

// SyncResult contains the results of a sync operation.
type SyncResult struct {
	// FilesCopied is the number of files copied.
	FilesCopied int

	// BytesCopied is the total bytes copied.
	BytesCopied int64

	// Errors contains any errors encountered during sync.
	Errors []error
}

// SyncOptions configures sync behavior.
type SyncOptions struct {
	// DryRun if true, reports what would be copied without copying.
	DryRun bool

	// Overwrite if true, overwrites existing files.
	Overwrite bool

	// Verbose if true, logs each file operation.
	Verbose bool

	// Include is a list of glob patterns to include.
	// If empty, all files are included.
	Include []string

	// Exclude is a list of glob patterns to exclude.
	Exclude []string
}

// SyncOutput copies files from a source directory (typically @ayo's /output/ mount)
// to a target directory (typically the user's CWD).
// Returns the sync result and any fatal error.
func SyncOutput(srcDir, dstDir string, opts SyncOptions) (SyncResult, error) {
	result := SyncResult{}

	// Ensure source exists
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			debug.Log("sync source does not exist", "src", srcDir)
			return result, nil // Nothing to sync
		}
		return result, fmt.Errorf("stat source: %w", err)
	}
	if !srcInfo.IsDir() {
		return result, fmt.Errorf("source is not a directory: %s", srcDir)
	}

	// Ensure destination exists
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return result, fmt.Errorf("create destination: %w", err)
	}

	// Walk source directory
	err = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			result.Errors = append(result.Errors, fmt.Errorf("walk %s: %w", path, walkErr))
			return nil // Continue walking
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("rel path %s: %w", path, err))
			return nil
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		// Apply include/exclude filters
		if !shouldInclude(relPath, opts.Include, opts.Exclude) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dstPath := filepath.Join(dstDir, relPath)

		if d.IsDir() {
			// Create directory
			if !opts.DryRun {
				if err := os.MkdirAll(dstPath, 0755); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("mkdir %s: %w", dstPath, err))
				}
			}
			if opts.Verbose {
				debug.Log("sync mkdir", "path", relPath)
			}
			return nil
		}

		// Copy file
		copied, size, err := copyFile(path, dstPath, opts)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("copy %s: %w", relPath, err))
			return nil
		}
		if copied {
			result.FilesCopied++
			result.BytesCopied += size
			if opts.Verbose {
				debug.Log("sync copied", "path", relPath, "size", size)
			}
		}

		return nil
	})

	if err != nil {
		return result, fmt.Errorf("walk source: %w", err)
	}

	return result, nil
}

// SyncAyoOutput syncs the @ayo sandbox output directory to the target.
// If targetDir is empty, uses the current working directory.
func SyncAyoOutput(targetDir string, opts SyncOptions) (SyncResult, error) {
	srcDir := paths.AyoSandboxOutputDir()

	if targetDir == "" {
		var err error
		targetDir, err = os.Getwd()
		if err != nil {
			return SyncResult{}, fmt.Errorf("get working directory: %w", err)
		}
	}

	return SyncOutput(srcDir, targetDir, opts)
}

// ClearOutput removes all files from the @ayo sandbox output directory.
func ClearOutput() error {
	outputDir := paths.AyoSandboxOutputDir()

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read output directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(outputDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// shouldInclude checks if a path should be included based on patterns.
func shouldInclude(relPath string, include, exclude []string) bool {
	// Check excludes first
	for _, pattern := range exclude {
		matched, _ := filepath.Match(pattern, filepath.Base(relPath))
		if matched {
			return false
		}
		// Also check full path
		matched, _ = matchPath(pattern, relPath)
		if matched {
			return false
		}
	}

	// If no includes specified, include everything
	if len(include) == 0 {
		return true
	}

	// Check includes
	for _, pattern := range include {
		matched, _ := filepath.Match(pattern, filepath.Base(relPath))
		if matched {
			return true
		}
		matched, _ = matchPath(pattern, relPath)
		if matched {
			return true
		}
	}

	return false
}

// matchPath matches a pattern against a path, handling ** for recursive matching.
func matchPath(pattern, path string) (bool, error) {
	// Simple implementation - use standard filepath.Match
	// For more complex patterns, we could use doublestar package
	return filepath.Match(pattern, path)
}

// copyFile copies a file from src to dst.
// Returns whether the file was copied, the size, and any error.
func copyFile(src, dst string, opts SyncOptions) (bool, int64, error) {
	if opts.DryRun {
		info, err := os.Stat(src)
		if err != nil {
			return false, 0, err
		}
		return true, info.Size(), nil
	}

	// Check if destination exists
	if !opts.Overwrite {
		if _, err := os.Stat(dst); err == nil {
			debug.Log("sync skip existing", "path", dst)
			return false, 0, nil
		}
	}

	// Open source
	srcFile, err := os.Open(src)
	if err != nil {
		return false, 0, err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return false, 0, err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return false, 0, err
	}

	// Create destination
	dstFile, err := os.Create(dst)
	if err != nil {
		return false, 0, err
	}
	defer dstFile.Close()

	// Copy content
	size, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return false, 0, err
	}

	// Preserve permissions
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		// Not fatal, just log
		debug.Log("sync chmod failed", "path", dst, "error", err)
	}

	return true, size, nil
}

// FormatSyncResult formats a sync result for display.
func FormatSyncResult(result SyncResult) string {
	var parts []string

	if result.FilesCopied > 0 {
		parts = append(parts, fmt.Sprintf("%d files copied", result.FilesCopied))
		if result.BytesCopied > 0 {
			parts = append(parts, fmt.Sprintf("(%s)", formatBytes(result.BytesCopied)))
		}
	} else {
		parts = append(parts, "no files copied")
	}

	if len(result.Errors) > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", len(result.Errors)))
	}

	return strings.Join(parts, ", ")
}

// formatBytes formats bytes as a human-readable string.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
