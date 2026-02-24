// Package hostwrite provides a tool for agents to request modifications to host files.
// This tool enables agents to create, update, or delete files on the host system
// with an approval flow when required.
package hostwrite

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fantasy"
)

// Action represents the type of file operation.
type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// HostWriteParams are the parameters for the host_write tool.
type HostWriteParams struct {
	// Action is the type of file operation: create, update, or delete.
	Action Action `json:"action" jsonschema:"required,enum=create,enum=update,enum=delete,description=Type of file operation"`

	// Path is the path relative to host home (e.g., 'Projects/app/main.go').
	Path string `json:"path" jsonschema:"required,description=Path relative to host home directory"`

	// Content is the file content (required for create/update).
	Content string `json:"content,omitempty" jsonschema:"description=File content for create/update operations"`

	// Reason explains why this change is needed (shown to user during approval).
	Reason string `json:"reason" jsonschema:"required,description=Why this change is needed (shown to user)"`
}

// HostWriteResult contains the result of a host write request.
type HostWriteResult struct {
	// Status is the result status: approved, denied, or error.
	Status string `json:"status"`

	// Path is the full resolved path on the host.
	Path string `json:"path"`

	// Message provides additional information about the result.
	Message string `json:"message"`
}

// ApprovalHandler handles approval requests for host file modifications.
type ApprovalHandler interface {
	// RequestApproval requests user approval for a file operation.
	// Returns true if approved, false if denied.
	RequestApproval(ctx context.Context, req ApprovalRequest) (bool, error)
}

// ApprovalRequest contains details about a pending file modification.
type ApprovalRequest struct {
	// Action is the operation type.
	Action Action `json:"action"`

	// Path is the full host path.
	Path string `json:"path"`

	// Content is the file content (for create/update).
	Content string `json:"content,omitempty"`

	// Reason is why the change is needed.
	Reason string `json:"reason"`

	// Agent is the agent requesting the modification.
	Agent string `json:"agent,omitempty"`
}

// ToolConfig configures the host write tool.
type ToolConfig struct {
	// BaseDir is the base directory for path resolution (typically host home).
	BaseDir string

	// ApprovalHandler handles approval requests.
	// If nil, all requests are auto-approved.
	ApprovalHandler ApprovalHandler

	// AutoApprove if true, automatically approves all requests.
	AutoApprove bool

	// AllowedPaths restricts which paths can be written.
	// If empty, any path under BaseDir is allowed.
	AllowedPaths []string

	// BlockedPaths are patterns for paths that cannot be modified.
	BlockedPaths []string

	// Agent is the name of the agent using this tool.
	Agent string
}

// DefaultBlockedPaths returns paths that should never be modified.
func DefaultBlockedPaths() []string {
	return []string{
		".ssh/*",
		".gnupg/*",
		".aws/*",
		".kube/*",
		"**/id_rsa*",
		"**/*.pem",
		"**/*.key",
		".bash_history",
		".zsh_history",
	}
}

// NewHostWriteTool creates a host write tool with the given configuration.
func NewHostWriteTool(cfg ToolConfig) fantasy.AgentTool {
	if cfg.BlockedPaths == nil {
		cfg.BlockedPaths = DefaultBlockedPaths()
	}

	return fantasy.NewAgentTool(
		"host_write",
		"Request permission to create, update, or delete a file on the host system",
		func(ctx context.Context, params HostWriteParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			// Validate action
			if params.Action != ActionCreate && params.Action != ActionUpdate && params.Action != ActionDelete {
				return fantasy.NewTextErrorResponse("action must be 'create', 'update', or 'delete'"), nil
			}

			// Validate content for create/update
			if (params.Action == ActionCreate || params.Action == ActionUpdate) && params.Content == "" {
				return fantasy.NewTextErrorResponse("content is required for create/update operations"), nil
			}

			// Validate path
			if params.Path == "" {
				return fantasy.NewTextErrorResponse("path is required"), nil
			}

			// Resolve full path
			fullPath := resolvePath(cfg.BaseDir, params.Path)

			// Validate path is allowed
			if err := validateHostPath(params.Path, fullPath, cfg); err != nil {
				result := HostWriteResult{
					Status:  "denied",
					Path:    fullPath,
					Message: err.Error(),
				}
				return fantasy.NewTextResponse(formatResult(result)), nil
			}

			// Build approval request
			req := ApprovalRequest{
				Action:  params.Action,
				Path:    fullPath,
				Content: params.Content,
				Reason:  params.Reason,
				Agent:   cfg.Agent,
			}

			// Check if approval needed
			approved := cfg.AutoApprove
			if !approved && cfg.ApprovalHandler != nil {
				var err error
				approved, err = cfg.ApprovalHandler.RequestApproval(ctx, req)
				if err != nil {
					result := HostWriteResult{
						Status:  "error",
						Path:    fullPath,
						Message: fmt.Sprintf("approval error: %v", err),
					}
					return fantasy.NewTextResponse(formatResult(result)), nil
				}
			} else if cfg.ApprovalHandler == nil {
				// No handler and not auto-approve - auto-approve by default
				approved = true
			}

			if !approved {
				result := HostWriteResult{
					Status:  "denied",
					Path:    fullPath,
					Message: "User denied the request",
				}
				return fantasy.NewTextResponse(formatResult(result)), nil
			}

			// Execute the operation
			var err error
			var message string

			switch params.Action {
			case ActionCreate, ActionUpdate:
				err = writeFile(fullPath, params.Content)
				if err == nil {
					message = fmt.Sprintf("File %s successfully", params.Action)
				}
			case ActionDelete:
				err = deleteFile(fullPath)
				if err == nil {
					message = "File deleted successfully"
				}
			}

			if err != nil {
				result := HostWriteResult{
					Status:  "error",
					Path:    fullPath,
					Message: err.Error(),
				}
				return fantasy.NewTextResponse(formatResult(result)), nil
			}

			result := HostWriteResult{
				Status:  "approved",
				Path:    fullPath,
				Message: message,
			}
			return fantasy.NewTextResponse(formatResult(result)), nil
		},
	)
}

// resolvePath resolves a relative path against the base directory.
func resolvePath(baseDir, path string) string {
	// Clean the path to prevent traversal
	path = filepath.Clean(path)

	// If path is already absolute, use it (but verify it's under baseDir)
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(baseDir, path)
}

// validateHostPath validates that a path can be written.
func validateHostPath(relPath, fullPath string, cfg ToolConfig) error {
	// Prevent path traversal
	if strings.Contains(relPath, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Check full path is under base dir
	if cfg.BaseDir != "" {
		rel, err := filepath.Rel(cfg.BaseDir, fullPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("path must be under base directory")
		}
	}

	// Check blocked patterns
	for _, pattern := range cfg.BlockedPaths {
		matched, _ := filepath.Match(pattern, relPath)
		if matched {
			return fmt.Errorf("path matches blocked pattern")
		}
		matched, _ = filepath.Match(pattern, filepath.Base(relPath))
		if matched {
			return fmt.Errorf("path matches blocked pattern")
		}
	}

	// Check allowed paths if specified
	if len(cfg.AllowedPaths) > 0 {
		allowed := false
		for _, prefix := range cfg.AllowedPaths {
			if strings.HasPrefix(relPath, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("path not in allowed paths")
		}
	}

	return nil
}

// writeFile writes content to a file, creating parent directories as needed.
func writeFile(path, content string) error {
	// Create parent directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// deleteFile deletes a file.
func deleteFile(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

// formatResult formats a result for display.
func formatResult(r HostWriteResult) string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}
