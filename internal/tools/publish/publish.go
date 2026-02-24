// Package publish provides a tool for agents to publish files from sandbox to host.
// Files must be in /output/ (the safe write zone) to be published.
// Publishing requires user approval via the file_request flow.
package publish

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/providers"
)

// FileMapping represents a source-to-destination file mapping.
type FileMapping struct {
	// Source is the path in /output/ to publish.
	Source string `json:"source" jsonschema:"required,description=Path in /output/ to publish"`

	// Destination is the host filesystem path.
	Destination string `json:"destination" jsonschema:"required,description=Host filesystem destination path"`
}

// PublishParams are the parameters for the publish tool.
type PublishParams struct {
	// Files is a list of source-to-destination file mappings.
	Files []FileMapping `json:"files" jsonschema:"required,description=Files to publish from /output/ to host"`

	// Message is an optional message explaining what's being published.
	Message string `json:"message,omitempty" jsonschema:"description=Optional message explaining the publish operation"`

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

// ApprovalRequester is called to request user approval for publish operations.
// Returns true if approved, false if denied.
type ApprovalRequester func(ctx context.Context, files []FileMapping, message string) (bool, error)

// ToolConfig configures the publish tool for a specific sandbox.
type ToolConfig struct {
	// Provider is the sandbox provider to use for file operations.
	Provider providers.SandboxProvider

	// SandboxID is the ID of the sandbox to publish from.
	SandboxID string

	// SessionID is the current session ID (for output directory).
	SessionID string

	// OutputDir is the path to the output directory in sandbox (default: /output).
	OutputDir string

	// HostOutputDir is the host-side path to the output directory.
	// Files are synced via VirtioFS so we read from host directly.
	HostOutputDir string

	// ApprovalRequester is called to request user approval.
	// If nil, all requests are approved (for testing).
	ApprovalRequester ApprovalRequester

	// BlockedDestinations are paths that cannot be published to.
	BlockedDestinations []string
}

// DefaultBlockedDestinations returns paths that should not be written to.
func DefaultBlockedDestinations() []string {
	return []string{
		"/",
		"/etc",
		"/usr",
		"/bin",
		"/sbin",
		"/var",
		"/System",
	}
}

// NewPublishTool creates a publish tool for the given configuration.
func NewPublishTool(cfg ToolConfig) fantasy.AgentTool {
	// Apply defaults
	if cfg.BlockedDestinations == nil {
		cfg.BlockedDestinations = DefaultBlockedDestinations()
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "/output"
	}

	return fantasy.NewAgentTool(
		"publish",
		"Publish files from /output/ to the host filesystem. Files must be in /output/ (the safe write zone). Requires user approval.",
		func(ctx context.Context, params PublishParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if len(params.Files) == 0 {
				return fantasy.NewTextErrorResponse("files is required; provide source-destination pairs to publish"), nil
			}

			result := PublishResult{
				Published: make([]string, 0),
				Skipped:   make([]string, 0),
				Errors:    make([]string, 0),
			}

			// Validate all sources are in /output/
			var validFiles []FileMapping
			for _, f := range params.Files {
				// Normalize source path
				source := f.Source
				if !strings.HasPrefix(source, "/") {
					source = "/" + source
				}

				// Must be in /output/
				if !strings.HasPrefix(source, cfg.OutputDir) && !strings.HasPrefix(source, "/output/") && !strings.HasPrefix(source, "/output") {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: source must be in /output/", f.Source))
					continue
				}

				// Validate destination
				if err := validateDestination(f.Destination, cfg); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", f.Destination, err))
					continue
				}

				validFiles = append(validFiles, FileMapping{
					Source:      source,
					Destination: f.Destination,
				})
			}

			if len(validFiles) == 0 {
				if len(result.Errors) > 0 {
					return fantasy.NewTextResponse(result.String()), nil
				}
				return fantasy.NewTextErrorResponse("no valid files to publish"), nil
			}

			// Request approval if requester is configured
			if cfg.ApprovalRequester != nil {
				approved, err := cfg.ApprovalRequester(ctx, validFiles, params.Message)
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("approval error: %v", err)), nil
				}
				if !approved {
					return fantasy.NewTextResponse("Publish request denied by user"), nil
				}
			}

			// Copy files
			for _, f := range validFiles {
				// Expand ~ in destination
				dest := expandHome(f.Destination)

				// Check if file exists and overwrite is disabled
				if !params.Overwrite {
					if _, err := os.Stat(dest); err == nil {
						result.Skipped = append(result.Skipped, fmt.Sprintf("%s: already exists", f.Destination))
						continue
					}
				}

				// Copy file from sandbox output to host
				if err := copyOutputFile(ctx, cfg, f.Source, dest); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", f.Source, err))
				} else {
					result.Published = append(result.Published, fmt.Sprintf("%s -> %s", f.Source, f.Destination))
				}
			}

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// validateDestination checks if a destination is allowed.
func validateDestination(dest string, cfg ToolConfig) error {
	// Expand home directory
	expanded := expandHome(dest)

	// Resolve to absolute path
	absDest, err := filepath.Abs(expanded)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check blocked destinations
	for _, blocked := range cfg.BlockedDestinations {
		if absDest == blocked || strings.HasPrefix(absDest, blocked+"/") {
			return fmt.Errorf("destination %s is not allowed", blocked)
		}
	}

	return nil
}

// expandHome expands ~ to the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}

// copyOutputFile copies a file from the sandbox output directory to the host.
// Since /output is mounted via VirtioFS, we can read directly from the host-side mount.
func copyOutputFile(ctx context.Context, cfg ToolConfig, source, dest string) error {
	// If we have a host output dir, read directly from it
	var sourceContent []byte
	var err error

	if cfg.HostOutputDir != "" {
		// Translate sandbox path to host path
		relPath := strings.TrimPrefix(source, cfg.OutputDir)
		relPath = strings.TrimPrefix(relPath, "/output")
		relPath = strings.TrimPrefix(relPath, "/")
		hostSource := filepath.Join(cfg.HostOutputDir, relPath)

		sourceContent, err = os.ReadFile(hostSource)
		if err != nil {
			return fmt.Errorf("read source: %w", err)
		}
	} else {
		// Fall back to reading via sandbox exec
		result, execErr := cfg.Provider.Exec(ctx, cfg.SandboxID, providers.ExecOptions{
			Command: fmt.Sprintf("cat %s", source),
		})
		if execErr != nil {
			return fmt.Errorf("exec failed: %w", execErr)
		}
		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "No such file") {
				return fmt.Errorf("file not found in sandbox")
			}
			return fmt.Errorf("cat failed: %s", result.Stderr)
		}
		sourceContent = []byte(result.Stdout)
	}

	// Create parent directory on host
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	// Write file to host
	if err := os.WriteFile(dest, sourceContent, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
