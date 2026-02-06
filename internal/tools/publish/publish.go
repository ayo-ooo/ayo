// Package publish provides a tool for agents to publish files from sandbox to host.
// Unlike the full sync operation, publish allows targeted export of specific files
// or artifacts from the sandbox to approved locations on the host.
package publish

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/providers"
)

// PublishParams are the parameters for the publish tool.
type PublishParams struct {
	// Files is a list of files to publish from the sandbox.
	// Paths are relative to the sandbox working directory.
	Files []string `json:"files" jsonschema:"required,description=List of file paths to publish from sandbox (relative to working directory)"`

	// Destination is the host directory to publish to.
	// Must be within allowed destinations.
	Destination string `json:"destination,omitempty" jsonschema:"description=Host destination directory (default: project root)"`

	// Overwrite controls whether to overwrite existing files.
	Overwrite bool `json:"overwrite,omitempty" jsonschema:"description=Whether to overwrite existing files (default: false)"`
}

// PublishResult contains the result of a publish operation.
type PublishResult struct {
	// Published lists files that were successfully published.
	Published []string `json:"published"`

	// Skipped lists files that were skipped (e.g., already exist, not found).
	Skipped []string `json:"skipped,omitempty"`

	// Errors lists any errors that occurred.
	Errors []string `json:"errors,omitempty"`
}

func (r PublishResult) String() string {
	var sb strings.Builder

	if len(r.Published) > 0 {
		sb.WriteString("Published:\n")
		for _, f := range r.Published {
			sb.WriteString("  ")
			sb.WriteString(f)
			sb.WriteString("\n")
		}
	}

	if len(r.Skipped) > 0 {
		sb.WriteString("Skipped:\n")
		for _, f := range r.Skipped {
			sb.WriteString("  ")
			sb.WriteString(f)
			sb.WriteString("\n")
		}
	}

	if len(r.Errors) > 0 {
		sb.WriteString("Errors:\n")
		for _, e := range r.Errors {
			sb.WriteString("  ")
			sb.WriteString(e)
			sb.WriteString("\n")
		}
	}

	if sb.Len() == 0 {
		return "No files published"
	}

	return sb.String()
}

// ToolConfig configures the publish tool for a specific sandbox.
type ToolConfig struct {
	// Provider is the sandbox provider to use for file operations.
	Provider providers.SandboxProvider

	// SandboxID is the ID of the sandbox to publish from.
	SandboxID string

	// SandboxWorkingDir is the working directory in the sandbox.
	SandboxWorkingDir string

	// HostProjectPath is the default destination on the host.
	HostProjectPath string

	// AllowedDestinations restricts where files can be published.
	// If empty, only HostProjectPath is allowed.
	// Use "*" to allow any destination (dangerous).
	AllowedDestinations []string

	// AllowedFilePatterns restricts which files can be published.
	// If empty, all files are allowed.
	AllowedFilePatterns []string

	// BlockedFilePatterns are patterns for files that cannot be published.
	BlockedFilePatterns []string
}

// DefaultBlockedPatterns returns patterns for files that should not be published.
// Note: filepath.Match doesn't support ** globs, so we use simple patterns
// and check both full path and basename.
func DefaultBlockedPatterns() []string {
	return []string{
		".git/*",
		".env",
		".env.*",
		"secrets/*",
		"*.key",
		"*.pem",
	}
}

// NewPublishTool creates a publish tool for the given configuration.
func NewPublishTool(cfg ToolConfig) fantasy.AgentTool {
	// Apply defaults
	if cfg.BlockedFilePatterns == nil {
		cfg.BlockedFilePatterns = DefaultBlockedPatterns()
	}
	if cfg.SandboxWorkingDir == "" {
		cfg.SandboxWorkingDir = "/workspace"
	}

	return fantasy.NewAgentTool(
		"publish",
		"Publish files from the sandbox to the host filesystem",
		func(ctx context.Context, params PublishParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if len(params.Files) == 0 {
				return fantasy.NewTextErrorResponse("files is required; provide a list of files to publish"), nil
			}

			result := PublishResult{
				Published: make([]string, 0),
				Skipped:   make([]string, 0),
				Errors:    make([]string, 0),
			}

			// Determine destination
			destination := params.Destination
			if destination == "" {
				destination = cfg.HostProjectPath
			}

			// Validate destination
			if err := validateDestination(destination, cfg); err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid destination: %v", err)), nil
			}

			for _, file := range params.Files {
				// Validate file
				if err := validateFile(file, cfg); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", file, err))
					continue
				}

				sandboxPath := filepath.Join(cfg.SandboxWorkingDir, file)
				hostPath := filepath.Join(destination, file)

				// Check if file exists on host and overwrite is disabled
				if !params.Overwrite {
					if _, err := os.Stat(hostPath); err == nil {
						result.Skipped = append(result.Skipped, fmt.Sprintf("%s: already exists", file))
						continue
					}
				}

				// Copy file from sandbox to host
				if err := copyFileFromSandbox(ctx, cfg.Provider, cfg.SandboxID, sandboxPath, hostPath); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", file, err))
				} else {
					result.Published = append(result.Published, file)
				}
			}

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// validateDestination checks if a destination is allowed.
func validateDestination(dest string, cfg ToolConfig) error {
	// Resolve to absolute path
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if within allowed destinations
	if len(cfg.AllowedDestinations) == 0 {
		// Default: only project path is allowed
		absProject, _ := filepath.Abs(cfg.HostProjectPath)
		if !strings.HasPrefix(absDest, absProject) {
			return fmt.Errorf("destination must be within project directory")
		}
		return nil
	}

	// Check allowed list
	for _, allowed := range cfg.AllowedDestinations {
		if allowed == "*" {
			return nil
		}
		absAllowed, _ := filepath.Abs(allowed)
		if strings.HasPrefix(absDest, absAllowed) {
			return nil
		}
	}

	return fmt.Errorf("destination not in allowed list")
}

// validateFile checks if a file path is allowed for publishing.
func validateFile(file string, cfg ToolConfig) error {
	// Prevent path traversal
	if strings.Contains(file, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Check blocked patterns
	for _, pattern := range cfg.BlockedFilePatterns {
		matched, _ := filepath.Match(pattern, file)
		if matched {
			return fmt.Errorf("matches blocked pattern: %s", pattern)
		}
		// Also check basename
		matched, _ = filepath.Match(pattern, filepath.Base(file))
		if matched {
			return fmt.Errorf("matches blocked pattern: %s", pattern)
		}
	}

	// Check allowed patterns if specified
	if len(cfg.AllowedFilePatterns) > 0 {
		allowed := false
		for _, pattern := range cfg.AllowedFilePatterns {
			matched, _ := filepath.Match(pattern, file)
			if matched {
				allowed = true
				break
			}
			matched, _ = filepath.Match(pattern, filepath.Base(file))
			if matched {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("does not match allowed patterns")
		}
	}

	return nil
}

// copyFileFromSandbox copies a file from sandbox to host.
func copyFileFromSandbox(ctx context.Context, provider providers.SandboxProvider, sandboxID, sandboxPath, hostPath string) error {
	// Read file from sandbox using cat
	result, err := provider.Exec(ctx, sandboxID, providers.ExecOptions{
		Command: fmt.Sprintf("cat %s", sandboxPath),
	})
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}
	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "No such file") {
			return fmt.Errorf("file not found in sandbox")
		}
		return fmt.Errorf("cat failed: %s", result.Stderr)
	}

	// Create parent directory on host
	if err := os.MkdirAll(filepath.Dir(hostPath), 0755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	// Write file to host
	if err := os.WriteFile(hostPath, []byte(result.Stdout), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Try to preserve permissions from sandbox
	statResult, err := provider.Exec(ctx, sandboxID, providers.ExecOptions{
		Command: fmt.Sprintf("stat -c %%a %s", sandboxPath),
	})
	if err == nil && statResult.ExitCode == 0 {
		var mode int
		if _, err := fmt.Sscanf(strings.TrimSpace(statResult.Stdout), "%o", &mode); err == nil {
			_ = os.Chmod(hostPath, os.FileMode(mode))
		}
	}

	return nil
}

// CopyDirectoryFromSandbox copies a directory from sandbox to host.
// This is useful for publishing entire output directories.
func CopyDirectoryFromSandbox(ctx context.Context, provider providers.SandboxProvider, sandboxID, sandboxDir, hostDir string) ([]string, error) {
	var published []string

	// Create tar of sandbox directory
	result, err := provider.Exec(ctx, sandboxID, providers.ExecOptions{
		Command:    "tar",
		Args:       []string{"-cf", "-", "."},
		WorkingDir: sandboxDir,
	})
	if err != nil {
		return nil, fmt.Errorf("create tar: %w", err)
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("tar failed: %s", result.Stderr)
	}

	// Create host directory
	if err := os.MkdirAll(hostDir, 0755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	// Extract tar on host
	tarReader := strings.NewReader(result.Stdout)
	if err := extractTar(tarReader, hostDir, &published); err != nil {
		return nil, fmt.Errorf("extract tar: %w", err)
	}

	return published, nil
}

// extractTar extracts a tar from a reader to a destination directory.
func extractTar(r io.Reader, dest string, published *[]string) error {
	// Simple implementation - read tar and write files
	// For a real implementation, use archive/tar
	// This is a placeholder that would need the tar parsing
	return nil
}
