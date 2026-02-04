// Package daemon provides a background process for managing ayo resources.
package daemon

import (
	"encoding/json"
	"fmt"
)

// JSON-RPC 2.0 protocol types

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      int64           `json:"id"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	ID      int64           `json:"id"`
}

// Error represents a JSON-RPC 2.0 error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

// Standard JSON-RPC error codes
const (
	ErrCodeParse          = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternal       = -32603
)

// Application-specific error codes
const (
	ErrCodeSandboxNotFound   = -1001
	ErrCodeSandboxExhausted  = -1002
	ErrCodeSandboxTimeout    = -1003
	ErrCodeDaemonShuttingDown = -1004
)

// NewError creates a new RPC error.
func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// NewRequest creates a new JSON-RPC request.
func NewRequest(method string, params any, id int64) (*Request, error) {
	var rawParams json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params: %w", err)
		}
		rawParams = data
	}
	return &Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  rawParams,
		ID:      id,
	}, nil
}

// NewResponse creates a successful JSON-RPC response.
func NewResponse(result any, id int64) (*Response, error) {
	var rawResult json.RawMessage
	if result != nil {
		data, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("marshal result: %w", err)
		}
		rawResult = data
	}
	return &Response{
		JSONRPC: "2.0",
		Result:  rawResult,
		ID:      id,
	}, nil
}

// NewErrorResponse creates an error JSON-RPC response.
func NewErrorResponse(err *Error, id int64) *Response {
	return &Response{
		JSONRPC: "2.0",
		Error:   err,
		ID:      id,
	}
}

// Method names
const (
	MethodPing           = "daemon.ping"
	MethodStatus         = "daemon.status"
	MethodShutdown       = "daemon.shutdown"
	MethodSandboxAcquire = "sandbox.acquire"
	MethodSandboxRelease = "sandbox.release"
	MethodSandboxExec    = "sandbox.exec"
	MethodSandboxStatus  = "sandbox.status"
)

// Request/Response types for each method

// PingResult is the response to daemon.ping.
type PingResult struct {
	Pong bool `json:"pong"`
}

// StatusResult is the response to daemon.status.
type StatusResult struct {
	Running     bool   `json:"running"`
	Uptime      int64  `json:"uptime_seconds"`
	PID         int    `json:"pid"`
	Version     string `json:"version"`
	Sandboxes   SandboxStatusResult `json:"sandboxes"`
	MemoryUsage int64  `json:"memory_usage_bytes"`
}

// ShutdownParams is the request for daemon.shutdown.
type ShutdownParams struct {
	Graceful bool `json:"graceful"`
}

// SandboxAcquireParams is the request for sandbox.acquire.
type SandboxAcquireParams struct {
	Agent   string `json:"agent"`
	Timeout int    `json:"timeout,omitempty"` // seconds
}

// SandboxAcquireResult is the response to sandbox.acquire.
type SandboxAcquireResult struct {
	SandboxID  string `json:"sandbox_id"`
	WorkingDir string `json:"working_dir"`
}

// SandboxReleaseParams is the request for sandbox.release.
type SandboxReleaseParams struct {
	SandboxID string `json:"sandbox_id"`
}

// SandboxExecParams is the request for sandbox.exec.
type SandboxExecParams struct {
	SandboxID  string `json:"sandbox_id"`
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir,omitempty"`
	Timeout    int    `json:"timeout,omitempty"` // seconds
}

// SandboxExecResult is the response to sandbox.exec.
type SandboxExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	TimedOut bool   `json:"timed_out,omitempty"`
}

// SandboxStatusResult is the response to sandbox.status.
type SandboxStatusResult struct {
	Total int `json:"total"`
	Idle  int `json:"idle"`
	InUse int `json:"in_use"`
}
