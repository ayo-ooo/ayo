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
	MethodSandboxJoin    = "sandbox.join"
	MethodSandboxAgents  = "sandbox.agents"

	// Session management methods
	MethodSessionList  = "session.list"
	MethodSessionStart = "session.start"
	MethodSessionStop  = "session.stop"
	MethodAgentWake    = "agent.wake"
	MethodAgentSleep   = "agent.sleep"
	MethodAgentStatus  = "agent.status"

	// Trigger management methods
	MethodTriggerList       = "trigger.list"
	MethodTriggerGet        = "trigger.get"
	MethodTriggerRegister   = "trigger.register"
	MethodTriggerRemove     = "trigger.remove"
	MethodTriggerTest       = "trigger.test"
	MethodTriggerSetEnabled = "trigger.set_enabled"
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

// SandboxJoinParams is the request for sandbox.join.
type SandboxJoinParams struct {
	SandboxID string `json:"sandbox_id"`
	Agent     string `json:"agent"`
}

// SandboxAgentsParams is the request for sandbox.agents.
type SandboxAgentsParams struct {
	SandboxID string `json:"sandbox_id"`
}

// SandboxAgentsResult is the response to sandbox.agents.
type SandboxAgentsResult struct {
	Agents []string `json:"agents"`
}

// Session management types

// SessionListResult is the response to session.list.
type SessionListResult struct {
	Sessions []SessionInfo `json:"sessions"`
}

// SessionInfo represents an active agent session.
type SessionInfo struct {
	ID          string `json:"id"`
	AgentHandle string `json:"agent_handle"`
	StartedAt   int64  `json:"started_at"`
	TriggerID   string `json:"trigger_id,omitempty"`
	Status      string `json:"status"` // running, idle, stopped
	LastActive  int64  `json:"last_active"`
	SessionID   string `json:"session_id,omitempty"`
}

// SessionStartParams is the request for session.start.
type SessionStartParams struct {
	AgentHandle string `json:"agent_handle"`
	TriggerID   string `json:"trigger_id,omitempty"`
	SessionID   string `json:"session_id,omitempty"` // Resume existing session
}

// SessionStartResult is the response to session.start.
type SessionStartResult struct {
	Session SessionInfo `json:"session"`
}

// SessionStopParams is the request for session.stop.
type SessionStopParams struct {
	SessionID string `json:"session_id"`
}

// AgentWakeParams is the request for agent.wake.
type AgentWakeParams struct {
	Handle    string `json:"handle"`
	TriggerID string `json:"trigger_id,omitempty"`
	SessionID string `json:"session_id,omitempty"` // Resume existing session
}

// AgentWakeResult is the response to agent.wake.
type AgentWakeResult struct {
	Session SessionInfo `json:"session"`
}

// AgentSleepParams is the request for agent.sleep.
type AgentSleepParams struct {
	Handle string `json:"handle"`
}

// AgentStatusParams is the request for agent.status.
type AgentStatusParams struct {
	Handle string `json:"handle"`
}

// AgentStatusResult is the response to agent.status.
type AgentStatusResult struct {
	Active     bool         `json:"active"`
	Handle     string       `json:"handle"`
	Session    *SessionInfo `json:"session,omitempty"`
	StartedAt  int64        `json:"started_at,omitempty"`
	LastActive int64        `json:"last_active,omitempty"`
}

// Trigger management types

// TriggerInfo represents a trigger for RPC responses.
type TriggerInfo struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // "cron", "watch", or "webhook"
	Agent   string `json:"agent"`
	Prompt  string `json:"prompt,omitempty"`
	Source  string `json:"source,omitempty"`
	Enabled bool   `json:"enabled"`

	// Cron-specific
	Schedule string `json:"schedule,omitempty"`

	// Watch-specific
	Path      string   `json:"path,omitempty"`
	Patterns  []string `json:"patterns,omitempty"`
	Recursive bool     `json:"recursive,omitempty"`
	Events    []string `json:"events,omitempty"`

	// Webhook-specific
	WebhookPath   string `json:"webhook_path,omitempty"`
	WebhookSecret string `json:"webhook_secret,omitempty"`
	WebhookFormat string `json:"webhook_format,omitempty"` // github, gitlab, generic
}

// TriggerListResult is the response to trigger.list.
type TriggerListResult struct {
	Triggers []TriggerInfo `json:"triggers"`
}

// TriggerGetParams is the request for trigger.get.
type TriggerGetParams struct {
	ID string `json:"id"`
}

// TriggerGetResult is the response to trigger.get.
type TriggerGetResult struct {
	Trigger TriggerInfo `json:"trigger"`
}

// TriggerRegisterParams is the request for trigger.register.
type TriggerRegisterParams struct {
	Type   string `json:"type"` // "cron", "watch", or "webhook"
	Agent  string `json:"agent"`
	Prompt string `json:"prompt,omitempty"`

	// Cron-specific
	Schedule string `json:"schedule,omitempty"`

	// Watch-specific
	Path      string   `json:"path,omitempty"`
	Patterns  []string `json:"patterns,omitempty"`
	Recursive bool     `json:"recursive,omitempty"`
	Events    []string `json:"events,omitempty"`

	// Webhook-specific
	WebhookPath   string `json:"webhook_path,omitempty"`
	WebhookSecret string `json:"webhook_secret,omitempty"`
	WebhookFormat string `json:"webhook_format,omitempty"` // github, gitlab, generic
}

// TriggerRegisterResult is the response to trigger.register.
type TriggerRegisterResult struct {
	Trigger TriggerInfo `json:"trigger"`
}

// TriggerRemoveParams is the request for trigger.remove.
type TriggerRemoveParams struct {
	ID string `json:"id"`
}

// TriggerTestParams is the request for trigger.test.
type TriggerTestParams struct {
	ID string `json:"id"`
}

// TriggerSetEnabledParams is the request for trigger.set_enabled.
type TriggerSetEnabledParams struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

// Matrix method names
const (
	MethodMatrixStatus       = "matrix.status"
	MethodMatrixRoomsList    = "matrix.rooms.list"
	MethodMatrixRoomsCreate  = "matrix.rooms.create"
	MethodMatrixRoomsMembers = "matrix.rooms.members"
	MethodMatrixRoomsInvite  = "matrix.rooms.invite"
	MethodMatrixRoomsJoin    = "matrix.rooms.join"
	MethodMatrixRoomsLeave   = "matrix.rooms.leave"
	MethodMatrixSend         = "matrix.send"
	MethodMatrixRead         = "matrix.read"
	MethodMatrixReadStream   = "matrix.read.stream"
)

// MatrixStatusResult is the response to matrix.status.
type MatrixStatusResult struct {
	Conduit ConduitStatus `json:"conduit"`
	Broker  BrokerStatus  `json:"broker"`
}

// MatrixRoomsListParams is the request for matrix.rooms.list.
type MatrixRoomsListParams struct {
	SessionID string `json:"session_id,omitempty"`
}

// MatrixRoomsListResult is the response to matrix.rooms.list.
type MatrixRoomsListResult struct {
	Rooms []*RoomInfo `json:"rooms"`
}

// MatrixRoomsCreateParams is the request for matrix.rooms.create.
type MatrixRoomsCreateParams struct {
	Name      string   `json:"name"`
	SessionID string   `json:"session_id,omitempty"`
	Invite    []string `json:"invite,omitempty"`
}

// MatrixRoomsCreateResult is the response to matrix.rooms.create.
type MatrixRoomsCreateResult struct {
	RoomID string `json:"room_id"`
}

// MatrixRoomsMembersParams is the request for matrix.rooms.members.
type MatrixRoomsMembersParams struct {
	RoomID string `json:"room_id"`
}

// MatrixRoomsMembersResult is the response to matrix.rooms.members.
type MatrixRoomsMembersResult struct {
	Members []MemberInfo `json:"members"`
}

// MemberInfo represents a room member.
type MemberInfo struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	IsAgent     bool   `json:"is_agent"`
	Handle      string `json:"handle,omitempty"`
}

// MatrixRoomsInviteParams is the request for matrix.rooms.invite.
type MatrixRoomsInviteParams struct {
	RoomID string `json:"room_id"`
	Handle string `json:"handle"`
}

// MatrixRoomsJoinParams is the request for matrix.rooms.join.
type MatrixRoomsJoinParams struct {
	RoomID string `json:"room_id"`
	Handle string `json:"handle"`
}

// MatrixRoomsLeaveParams is the request for matrix.rooms.leave.
type MatrixRoomsLeaveParams struct {
	RoomID string `json:"room_id"`
	Handle string `json:"handle"`
}

// MatrixSendParams is the request for matrix.send.
type MatrixSendParams struct {
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
	AsAgent string `json:"as_agent,omitempty"`
	ReplyTo string `json:"reply_to,omitempty"`
	Format  string `json:"format,omitempty"` // "plain" or "markdown"
}

// MatrixSendResult is the response to matrix.send.
type MatrixSendResult struct {
	EventID string `json:"event_id"`
}

// MatrixReadParams is the request for matrix.read.
type MatrixReadParams struct {
	RoomID string `json:"room_id"`
	Limit  int    `json:"limit,omitempty"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

// MatrixReadResult is the response to matrix.read.
type MatrixReadResult struct {
	Messages []*QueuedMessage `json:"messages"`
	HasMore  bool             `json:"has_more"`
	Start    string           `json:"start,omitempty"`
	End      string           `json:"end,omitempty"`
}

// MatrixReadStreamParams is the request for matrix.read.stream.
type MatrixReadStreamParams struct {
	RoomID  string `json:"room_id"`
	Timeout int    `json:"timeout,omitempty"`
	Since   string `json:"since,omitempty"`
}

// Flow method names
const (
	MethodFlowRun     = "flow.run"
	MethodFlowList    = "flow.list"
	MethodFlowGet     = "flow.get"
	MethodFlowHistory = "flow.history"
)

// FlowRunParams is the request for flow.run.
type FlowRunParams struct {
	FlowName   string         `json:"flow_name"`
	Params     map[string]any `json:"params,omitempty"`
	Timeout    int            `json:"timeout,omitempty"` // seconds
	Async      bool           `json:"async,omitempty"`   // return immediately with run ID
	SessionID  string         `json:"session_id,omitempty"`
}

// FlowRunResult is the response to flow.run.
type FlowRunResult struct {
	RunID     string                    `json:"run_id"`
	FlowName  string                    `json:"flow_name"`
	Status    string                    `json:"status"`
	Steps     map[string]*FlowStepResult `json:"steps,omitempty"`
	StartTime int64                     `json:"start_time"`
	EndTime   int64                     `json:"end_time,omitempty"`
	Duration  int64                     `json:"duration_ms,omitempty"`
	Error     string                    `json:"error,omitempty"`
}

// FlowStepResult represents a single step's execution result.
type FlowStepResult struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Stdout    string `json:"stdout,omitempty"`
	Stderr    string `json:"stderr,omitempty"`
	Output    string `json:"output,omitempty"`
	ExitCode  int    `json:"exit_code,omitempty"`
	Error     string `json:"error,omitempty"`
	Skipped   bool   `json:"skipped,omitempty"`
	Duration  int64  `json:"duration_ms"`
}

// FlowListParams is the request for flow.list.
type FlowListParams struct {
	Source string `json:"source,omitempty"` // filter by source: built-in, user, project
}

// FlowListResult is the response to flow.list.
type FlowListResult struct {
	Flows []FlowInfo `json:"flows"`
}

// FlowInfo represents flow metadata.
type FlowInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Path        string `json:"path"`
	Version     int    `json:"version,omitempty"`
	IsYAML      bool   `json:"is_yaml"`
	StepCount   int    `json:"step_count,omitempty"`
}

// FlowGetParams is the request for flow.get.
type FlowGetParams struct {
	Name string `json:"name"`
}

// FlowGetResult is the response to flow.get.
type FlowGetResult struct {
	Flow FlowInfo `json:"flow"`
}

// FlowHistoryParams is the request for flow.history.
type FlowHistoryParams struct {
	FlowName string `json:"flow_name,omitempty"`
	Status   string `json:"status,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// FlowHistoryResult is the response to flow.history.
type FlowHistoryResult struct {
	Runs []FlowRunSummary `json:"runs"`
}

// FlowRunSummary is a summary of a flow run for history.
type FlowRunSummary struct {
	RunID     string `json:"run_id"`
	FlowName  string `json:"flow_name"`
	Status    string `json:"status"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time,omitempty"`
	Duration  int64  `json:"duration_ms,omitempty"`
	Error     string `json:"error,omitempty"`
}
