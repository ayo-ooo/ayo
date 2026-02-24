// Package ayod provides the in-sandbox daemon for ayo.
// ayod runs as PID 1 inside sandboxes and handles all sandbox operations
// via a Unix socket, providing consistent behavior across providers.
package ayod

import "os"

// SocketPath is the default path for the ayod Unix socket inside sandboxes.
const SocketPath = "/run/ayod.sock"

// UserAddRequest is the request to create a new user in the sandbox.
type UserAddRequest struct {
	// Username is the name for the new user.
	Username string `json:"username"`

	// Shell is the login shell (default: /bin/sh).
	Shell string `json:"shell,omitempty"`

	// Dotfiles is an optional tar archive of dotfiles to extract to home.
	Dotfiles []byte `json:"dotfiles,omitempty"`
}

// UserAddResponse is the response from creating a user.
type UserAddResponse struct {
	// UID is the numeric user ID.
	UID int `json:"uid"`

	// GID is the numeric group ID.
	GID int `json:"gid"`

	// Home is the home directory path.
	Home string `json:"home"`
}

// ExecRequest is the request to execute a command.
type ExecRequest struct {
	// User is the username to run as.
	User string `json:"user"`

	// Command is the command and arguments to execute.
	Command []string `json:"command"`

	// Env is additional environment variables.
	Env map[string]string `json:"env,omitempty"`

	// Cwd is the working directory. Empty uses the user's home.
	Cwd string `json:"cwd,omitempty"`

	// Timeout in seconds. 0 means no timeout.
	Timeout int `json:"timeout,omitempty"`
}

// ExecResponse is the result of command execution.
type ExecResponse struct {
	// ExitCode is the process exit code.
	ExitCode int `json:"exit_code"`

	// Stdout is the standard output.
	Stdout string `json:"stdout"`

	// Stderr is the standard error.
	Stderr string `json:"stderr"`
}

// ReadFileRequest is the request to read a file.
type ReadFileRequest struct {
	// Path is the file path to read.
	Path string `json:"path"`
}

// ReadFileResponse is the response containing file contents.
type ReadFileResponse struct {
	// Content is the file content.
	Content []byte `json:"content"`

	// Mode is the file mode.
	Mode os.FileMode `json:"mode"`
}

// WriteFileRequest is the request to write a file.
type WriteFileRequest struct {
	// Path is the file path to write.
	Path string `json:"path"`

	// Content is the file content.
	Content []byte `json:"content"`

	// Mode is the file mode.
	Mode os.FileMode `json:"mode"`
}

// HealthResponse is the health check response.
type HealthResponse struct {
	// Status is "ok" if healthy.
	Status string `json:"status"`

	// Users is the list of created users.
	Users []string `json:"users"`

	// Uptime is seconds since ayod started.
	Uptime float64 `json:"uptime"`
}

// RPCRequest is the JSON-RPC request envelope.
type RPCRequest struct {
	// Method is the RPC method name.
	Method string `json:"method"`

	// Params is the method parameters (JSON-encoded).
	Params []byte `json:"params,omitempty"`
}

// RPCResponse is the JSON-RPC response envelope.
type RPCResponse struct {
	// Result is the method result (JSON-encoded).
	Result []byte `json:"result,omitempty"`

	// Error is the error message if failed.
	Error string `json:"error,omitempty"`
}
