// Package requestaccess provides a tool for agents to request access to host files/directories.
// When an agent needs to access a path on the host that isn't currently shared,
// it can use this tool to request that the user approve mounting the path.
package requestaccess

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/huh"


)

// RequestAccessParams are the parameters for the request_access tool.
type RequestAccessParams struct {
	// Path is the host filesystem path to request access to.
	Path string `json:"path" jsonschema:"required,description=The host filesystem path to request access to (file or directory)"`

	// Reason explains why the agent needs access to this path.
	Reason string `json:"reason,omitempty" jsonschema:"description=Explanation of why access is needed (shown to user)"`

	// Name is an optional friendly name for the mount in /workspace/.
	// If not provided, the basename of the path will be used.
	Name string `json:"name,omitempty" jsonschema:"description=Friendly name for the mount (default: basename of path)"`

	// ReadOnly requests read-only access when true (default: true).
	// Note: This is advisory; the share service doesn't currently enforce read-only.
	ReadOnly *bool `json:"read_only,omitempty" jsonschema:"description=Request read-only access (default: true)"`
}

// RequestAccessResult contains the result of an access request.
type RequestAccessResult struct {
	// Granted is true if the user approved the request.
	Granted bool `json:"granted"`

	// MountPath is the path where the share is mounted in the sandbox.
	// Only set if Granted is true.
	MountPath string `json:"mount_path,omitempty"`

	// Message provides additional context about the result.
	Message string `json:"message,omitempty"`
}

func (r RequestAccessResult) String() string {
	if r.Granted {
		return fmt.Sprintf("Access granted. Path mounted at: %s", r.MountPath)
	}
	if r.Message != "" {
		return fmt.Sprintf("Access denied: %s", r.Message)
	}
	return "Access denied by user"
}

// ToolConfig configures the request_access tool.
type ToolConfig struct {
	// ShareService has been removed as part of framework cleanup.
	// File access is now handled directly by the build system.
	// ShareService *share.Service

	// SessionID is the current session ID (for session-scoped shares).
	SessionID string

	// SessionScoped determines if shares are removed when the session ends.
	// Default: true
	SessionScoped *bool

	// BlockedPaths are paths that cannot be requested (e.g., sensitive directories).
	BlockedPaths []string

	// AllowedPrefixes restricts which paths can be requested.
	// If empty, any path is allowed (subject to BlockedPaths).
	AllowedPrefixes []string
}

// DefaultBlockedPaths returns paths that should never be shared.
func DefaultBlockedPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		"/",
		"/etc",
		"/var",
		"/usr",
		"/bin",
		"/sbin",
		"/System",
		"/Library",
		filepath.Join(home, ".ssh"),
		filepath.Join(home, ".gnupg"),
		filepath.Join(home, ".aws"),
		filepath.Join(home, ".kube"),
		filepath.Join(home, ".config"),
	}
}

// NewRequestAccessTool creates a request_access tool for the given configuration.
func NewRequestAccessTool(cfg ToolConfig) fantasy.AgentTool {
	// Apply defaults
	if cfg.BlockedPaths == nil {
		cfg.BlockedPaths = DefaultBlockedPaths()
	}
	sessionScoped := true
	if cfg.SessionScoped != nil {
		sessionScoped = *cfg.SessionScoped
	}

	return fantasy.NewAgentTool(
		"request_access",
		"Request access to a file or directory on the host filesystem. The user will be prompted to approve or deny the request. If approved, the path will be mounted in the sandbox's /workspace/ directory.",
		func(ctx context.Context, params RequestAccessParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Path == "" {
				return fantasy.NewTextErrorResponse("path is required"), nil
			}

			// Expand ~ to home directory
			path := params.Path
			if strings.HasPrefix(path, "~/") {
				home, err := os.UserHomeDir()
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("cannot expand ~: %v", err)), nil
				}
				path = strings.Replace(path, "~", home, 1)
			}

			// Resolve to absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid path: %v", err)), nil
			}

			// Validate path exists
			info, err := os.Stat(absPath)
			if err != nil {
				if os.IsNotExist(err) {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("path does not exist: %s", absPath)), nil
				}
				return fantasy.NewTextErrorResponse(fmt.Sprintf("cannot access path: %v", err)), nil
			}

			// Check blocked paths
			if err := validatePath(absPath, cfg); err != nil {
				result := RequestAccessResult{
					Granted: false,
					Message: err.Error(),
				}
				return fantasy.NewTextResponse(result.String()), nil
			}

			// Check if already shared
			// ShareService has been removed as part of framework cleanup
			// if cfg.ShareService != nil {
			// 	if existing := cfg.ShareService.GetByPath(absPath); existing != nil {
			// 		result := RequestAccessResult{
			// 			Granted:   true,
			// 			MountPath: fmt.Sprintf("/workspace/%s", existing.Name),
			// 			Message:   "Path is already shared",
			// 		}
			// 		return fantasy.NewTextResponse(result.String()), nil
			// 	}
			// }

			// Determine mount name
			mountName := params.Name
			if mountName == "" {
				mountName = filepath.Base(absPath)
			}

			// Build prompt message
			pathType := "directory"
			if !info.IsDir() {
				pathType = "file"
			}

			promptTitle := fmt.Sprintf("🔒 Access Request: %s", absPath)
			promptDesc := fmt.Sprintf("An AI agent is requesting access to this %s.", pathType)
			if params.Reason != "" {
				promptDesc += fmt.Sprintf("\n\nReason: %s", params.Reason)
			}
			promptDesc += fmt.Sprintf("\n\nIf approved, it will be mounted at: /workspace/%s", mountName)

			// Prompt user for approval using huh
			var approved bool
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(promptTitle).
						Description(promptDesc).
						Affirmative("Allow").
						Negative("Deny").
						Value(&approved),
				),
			).WithTheme(huh.ThemeCharm())

			if err := form.RunWithContext(ctx); err != nil {
				// Context cancelled or form error
				result := RequestAccessResult{
					Granted: false,
					Message: "Request cancelled",
				}
				return fantasy.NewTextResponse(result.String()), nil
			}

			if !approved {
				result := RequestAccessResult{
					Granted: false,
					Message: "User denied the request",
				}
				return fantasy.NewTextResponse(result.String()), nil
			}

			// Add share
			// ShareService has been removed as part of framework cleanup
			// if cfg.ShareService != nil {
			// 	if err := cfg.ShareService.Add(absPath, mountName, sessionScoped, cfg.SessionID); err != nil {
			// 		result := RequestAccessResult{
			// 			Granted: false,
			// 			Message: fmt.Sprintf("Failed to create share: %v", err),
					}
					return fantasy.NewTextResponse(result.String()), nil
				// }
			// }

			result := RequestAccessResult{
				Granted:   true,
				MountPath: fmt.Sprintf("/workspace/%s", mountName),
			}
			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// validatePath checks if a path is allowed to be requested.
func validatePath(absPath string, cfg ToolConfig) error {
	// Check blocked paths
	for _, blocked := range cfg.BlockedPaths {
		if absPath == blocked {
			return fmt.Errorf("access to %s is not allowed", absPath)
		}
		// Also check if path is a subdirectory of blocked path for system dirs
		if blocked == "/" || blocked == "/etc" || blocked == "/var" ||
			blocked == "/usr" || blocked == "/bin" || blocked == "/sbin" ||
			blocked == "/System" || blocked == "/Library" {
			if strings.HasPrefix(absPath, blocked+"/") {
				return fmt.Errorf("access to %s is not allowed", absPath)
			}
		}
	}

	// Check allowed prefixes if specified
	if len(cfg.AllowedPrefixes) > 0 {
		allowed := false
		for _, prefix := range cfg.AllowedPrefixes {
			if strings.HasPrefix(absPath, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("path is not in allowed locations")
		}
	}

	return nil
}
