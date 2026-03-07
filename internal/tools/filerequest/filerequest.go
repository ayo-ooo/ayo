// Package filerequest provides a tool for agents to request files from the host.
// When agents run in sandboxed environments with working copies (not live mounts),
// they may need to request additional files that weren't included in the initial copy.
package filerequest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/providers"
	// Sandbox infrastructure removed - TODO: Re-enable when sandbox is re-implemented as standalone executable
	// "github.com/alexcabrera/ayo/internal/sandbox/workingcopy"
)

// FileRequestParams are the parameters for the file_request tool.
type FileRequestParams struct {
	// Paths is a list of file or directory paths to request from the host.
	// Paths are relative to the project root.
	Paths []string `json:"paths" jsonschema:"required,description=List of file or directory paths to request from the host (relative to project root)"`

	// Destination is where to place the files in the sandbox.
	// Defaults to the sandbox's working copy directory.
	Destination string `json:"destination,omitempty" jsonschema:"description=Destination directory in sandbox (default: working copy directory)"`
}

// FileRequestResult contains the result of a file request.
type FileRequestResult struct {
	// Copied lists files that were successfully copied.
	Copied []string `json:"copied"`

	// Skipped lists files that were skipped (e.g., already exist, not found).
	Skipped []string `json:"skipped,omitempty"`

	// Errors lists any errors that occurred.
	Errors []string `json:"errors,omitempty"`
}

func (r FileRequestResult) String() string {
	var sb strings.Builder

	if len(r.Copied) > 0 {
		sb.WriteString("Copied:\n")
		for _, f := range r.Copied {
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
		return "No files requested"
	}

	return sb.String()
}

// ToolConfig configures the file request tool for a specific sandbox.
type ToolConfig struct {
	// Provider is the sandbox provider to use for file operations.
	Provider providers.SandboxProvider

	// SandboxID is the ID of the sandbox to copy files to.
	SandboxID string

	// HostProjectPath is the path to the project on the host.
	HostProjectPath string

	// SandboxWorkingCopy is the working copy in the sandbox.
	// If nil, files are copied directly without working copy tracking.
	// Sandbox infrastructure removed - TODO: Re-enable when sandbox is re-implemented as standalone executable
	// SandboxWorkingCopy *workingcopy.WorkingCopy
	SandboxWorkingCopy interface{}

	// AllowedPrefixes restricts which paths can be requested.
	// If empty, all paths under HostProjectPath are allowed.
	AllowedPrefixes []string

	// BlockedPatterns are glob patterns for files that cannot be requested.
	// Common defaults: .git/*, .env, **/secrets/*
	BlockedPatterns []string
}

// DefaultBlockedPatterns returns the default patterns for files that cannot be requested.
func DefaultBlockedPatterns() []string {
	return []string{
		".git/*",
		".env",
		".env.*",
		"**/secrets/*",
		"**/*.key",
		"**/*.pem",
		"**/id_rsa*",
		"**/.ssh/*",
	}
}

// NewFileRequestTool creates a file request tool for the given configuration.
func NewFileRequestTool(cfg ToolConfig) fantasy.AgentTool {
	// Apply defaults
	if cfg.BlockedPatterns == nil {
		cfg.BlockedPatterns = DefaultBlockedPatterns()
	}

	return fantasy.NewAgentTool(
		"file_request",
		"Request files from the host project to be copied into the sandbox",
		func(ctx context.Context, params FileRequestParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if len(params.Paths) == 0 {
				return fantasy.NewTextErrorResponse("paths is required; provide a list of files to request"), nil
			}

			result := FileRequestResult{
				Copied:  make([]string, 0),
				Skipped: make([]string, 0),
				Errors:  make([]string, 0),
			}

			destination := params.Destination
			if destination == "" {
				// Sandbox infrastructure removed - TODO: Re-enable when sandbox is re-implemented as standalone executable
				// if cfg.SandboxWorkingCopy != nil {
				// 	destination = cfg.SandboxWorkingCopy.SandboxPath
				// } else {
				// 	destination = "/workspace"
				// }
				destination = "/workspace"
			}

			for _, reqPath := range params.Paths {
				// Validate path
				if err := validatePath(reqPath, cfg); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", reqPath, err))
					continue
				}

				// Resolve host path
				hostPath := filepath.Join(cfg.HostProjectPath, reqPath)

				// Check if file exists on host
				info, err := os.Stat(hostPath)
				if err != nil {
					if os.IsNotExist(err) {
						result.Skipped = append(result.Skipped, fmt.Sprintf("%s: not found", reqPath))
					} else {
						result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", reqPath, err))
					}
					continue
				}

				// Copy file(s) to sandbox
				sandboxPath := filepath.Join(destination, reqPath)

				if info.IsDir() {
					// Copy directory
					copied, err := copyDirectoryToSandbox(ctx, cfg.Provider, cfg.SandboxID, hostPath, sandboxPath)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", reqPath, err))
					} else {
						result.Copied = append(result.Copied, copied...)
					}
				} else {
					// Copy single file
					if err := copyFileToSandbox(ctx, cfg.Provider, cfg.SandboxID, hostPath, sandboxPath); err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", reqPath, err))
					} else {
						result.Copied = append(result.Copied, reqPath)
					}
				}
			}

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// validatePath checks if a requested path is allowed.
func validatePath(path string, cfg ToolConfig) error {
	// Prevent path traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Check blocked patterns
	for _, pattern := range cfg.BlockedPatterns {
		matched, _ := filepath.Match(pattern, path)
		if matched {
			return fmt.Errorf("path matches blocked pattern: %s", pattern)
		}
		// Also check basename
		matched, _ = filepath.Match(pattern, filepath.Base(path))
		if matched {
			return fmt.Errorf("path matches blocked pattern: %s", pattern)
		}
	}

	// Check allowed prefixes if specified
	if len(cfg.AllowedPrefixes) > 0 {
		allowed := false
		for _, prefix := range cfg.AllowedPrefixes {
			if strings.HasPrefix(path, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("path not in allowed prefixes")
		}
	}

	return nil
}

// copyFileToSandbox copies a single file from host to sandbox.
func copyFileToSandbox(ctx context.Context, provider providers.SandboxProvider, sandboxID, hostPath, sandboxPath string) error {
	// Read file from host
	data, err := os.ReadFile(hostPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Create parent directory in sandbox
	parentDir := filepath.Dir(sandboxPath)
	_, err = provider.Exec(ctx, sandboxID, providers.ExecOptions{
		Command: fmt.Sprintf("mkdir -p %s", parentDir),
	})
	if err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	// Write file to sandbox using cat with stdin
	_, err = provider.Exec(ctx, sandboxID, providers.ExecOptions{
		Command: fmt.Sprintf("cat > %s", sandboxPath),
		Stdin:   data,
	})
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Preserve file permissions
	info, err := os.Stat(hostPath)
	if err == nil {
		mode := info.Mode().Perm()
		_, _ = provider.Exec(ctx, sandboxID, providers.ExecOptions{
			Command: fmt.Sprintf("chmod %o %s", mode, sandboxPath),
		})
	}

	return nil
}

// copyDirectoryToSandbox copies a directory from host to sandbox.
// Sandbox infrastructure removed - TODO: Re-enable when sandbox is re-implemented as standalone executable
func copyDirectoryToSandbox(ctx context.Context, provider providers.SandboxProvider, sandboxID, hostPath, sandboxPath string) ([]string, error) {
	return nil, fmt.Errorf("directory copy disabled during sandbox infrastructure removal")
}
