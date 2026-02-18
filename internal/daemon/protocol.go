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

	// Squad management methods
	MethodSquadCreate         = "squads.create"
	MethodSquadDestroy        = "squads.destroy"
	MethodSquadList           = "squads.list"
	MethodSquadGet            = "squads.get"
	MethodSquadStart          = "squads.start"
	MethodSquadStop           = "squads.stop"
	MethodSquadAddAgent       = "squads.add_agent"
	MethodSquadRemoveAgent    = "squads.remove_agent"
	MethodSquadTicketsReady   = "squads.tickets_ready"
	MethodSquadNotifyAgents   = "squads.notify_agents"
	MethodSquadWaitCompletion = "squads.wait_completion"
	MethodSquadSyncOutput     = "squads.sync_output"
	MethodSquadCleanup        = "squads.cleanup"
	MethodSquadDispatch       = "squads.dispatch"

	// Agent invocation methods
	MethodAgentInvoke = "agent.invoke"
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

// Ticket method names
const (
	MethodTicketCreate  = "tickets.create"
	MethodTicketGet     = "tickets.get"
	MethodTicketList    = "tickets.list"
	MethodTicketUpdate  = "tickets.update"
	MethodTicketDelete  = "tickets.delete"
	MethodTicketStart   = "tickets.start"
	MethodTicketClose   = "tickets.close"
	MethodTicketReopen  = "tickets.reopen"
	MethodTicketBlock   = "tickets.block"
	MethodTicketAssign  = "tickets.assign"
	MethodTicketAddNote = "tickets.add_note"
	MethodTicketReady   = "tickets.ready"
	MethodTicketBlocked = "tickets.blocked"
	MethodTicketAddDep  = "tickets.add_dep"
	MethodTicketRemDep  = "tickets.remove_dep"
)

// TicketCreateParams is the request for tickets.create.
type TicketCreateParams struct {
	SessionID   string   `json:"session_id"`
	SquadName   string   `json:"squad_name,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	Assignee    string   `json:"assignee,omitempty"`
	Deps        []string `json:"deps,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ExternalRef string   `json:"external_ref,omitempty"`
}

// TicketCreateResult is the response to tickets.create.
type TicketCreateResult struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

// TicketGetParams is the request for tickets.get.
type TicketGetParams struct {
	SessionID string `json:"session_id"`
	SquadName string `json:"squad_name,omitempty"`
	TicketID  string `json:"ticket_id"`
}

// TicketInfo represents a ticket in RPC responses.
type TicketInfo struct {
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	Type        string   `json:"type"`
	Priority    int      `json:"priority"`
	Assignee    string   `json:"assignee,omitempty"`
	Deps        []string `json:"deps"`
	Links       []string `json:"links"`
	Parent      string   `json:"parent,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Created     int64    `json:"created"`
	Started     int64    `json:"started,omitempty"`
	Closed      int64    `json:"closed,omitempty"`
	Session     string   `json:"session,omitempty"`
	ExternalRef string   `json:"external_ref,omitempty"`
	FilePath    string   `json:"file_path"`
}

// TicketGetResult is the response to tickets.get.
type TicketGetResult struct {
	Ticket TicketInfo `json:"ticket"`
}

// TicketListParams is the request for tickets.list.
type TicketListParams struct {
	SessionID string   `json:"session_id"`
	SquadName string   `json:"squad_name,omitempty"`
	Status    string   `json:"status,omitempty"`
	Assignee  string   `json:"assignee,omitempty"`
	Type      string   `json:"type,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Parent    string   `json:"parent,omitempty"`
}

// TicketListResult is the response to tickets.list.
type TicketListResult struct {
	Tickets []TicketInfo `json:"tickets"`
}

// TicketUpdateParams is the request for tickets.update.
type TicketUpdateParams struct {
	SessionID   string   `json:"session_id"`
	TicketID    string   `json:"ticket_id"`
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Type        *string  `json:"type,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	Assignee    *string  `json:"assignee,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ExternalRef *string  `json:"external_ref,omitempty"`
}

// TicketDeleteParams is the request for tickets.delete.
type TicketDeleteParams struct {
	SessionID string `json:"session_id"`
	TicketID  string `json:"ticket_id"`
}

// TicketStatusParams is the request for tickets.start/close/reopen/block.
type TicketStatusParams struct {
	SessionID string `json:"session_id"`
	SquadName string `json:"squad_name,omitempty"`
	TicketID  string `json:"ticket_id"`
}

// TicketAssignParams is the request for tickets.assign.
type TicketAssignParams struct {
	SessionID string `json:"session_id"`
	SquadName string `json:"squad_name,omitempty"`
	TicketID  string `json:"ticket_id"`
	Assignee  string `json:"assignee"`
}

// TicketAddNoteParams is the request for tickets.add_note.
type TicketAddNoteParams struct {
	SessionID string `json:"session_id"`
	SquadName string `json:"squad_name,omitempty"`
	TicketID  string `json:"ticket_id"`
	Content   string `json:"content"`
}

// TicketReadyParams is the request for tickets.ready.
type TicketReadyParams struct {
	SessionID string `json:"session_id"`
	SquadName string `json:"squad_name,omitempty"`
	Assignee  string `json:"assignee,omitempty"`
}

// TicketReadyResult is the response to tickets.ready.
type TicketReadyResult struct {
	Tickets []TicketInfo `json:"tickets"`
}

// TicketBlockedParams is the request for tickets.blocked.
type TicketBlockedParams struct {
	SessionID string `json:"session_id"`
	SquadName string `json:"squad_name,omitempty"`
	Assignee  string `json:"assignee,omitempty"`
}

// TicketBlockedResult is the response to tickets.blocked.
type TicketBlockedResult struct {
	Tickets []TicketInfo `json:"tickets"`
}

// TicketDepParams is the request for tickets.add_dep and tickets.remove_dep.
type TicketDepParams struct {
	SessionID string `json:"session_id"`
	SquadName string `json:"squad_name,omitempty"`
	TicketID  string `json:"ticket_id"`
	DepID     string `json:"dep_id"`
}

// Squad management request/response types

// SquadInfo contains information about a squad.
type SquadInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`
	Agents      []string `json:"agents,omitempty"`
	Ephemeral   bool     `json:"ephemeral,omitempty"`
	TicketsDir  string   `json:"tickets_dir,omitempty"`
	ContextDir  string   `json:"context_dir,omitempty"`
	WorkspaceDir string  `json:"workspace_dir,omitempty"`
}

// SquadCreateParams is the request for squads.create.
type SquadCreateParams struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	Image          string   `json:"image,omitempty"`
	Ephemeral      bool     `json:"ephemeral,omitempty"`
	Agents         []string `json:"agents,omitempty"`
	WorkspaceMount string   `json:"workspace_mount,omitempty"`
	Packages       []string `json:"packages,omitempty"`
	OutputPath     string   `json:"output_path,omitempty"`
}

// SquadCreateResult is the response to squads.create.
type SquadCreateResult struct {
	Squad SquadInfo `json:"squad"`
}

// SquadDestroyParams is the request for squads.destroy.
type SquadDestroyParams struct {
	Name       string `json:"name"`
	DeleteData bool   `json:"delete_data,omitempty"`
}

// SquadDestroyResult is the response to squads.destroy.
type SquadDestroyResult struct {
	Success bool `json:"success"`
}

// SquadListParams is the request for squads.list.
type SquadListParams struct{}

// SquadListResult is the response to squads.list.
type SquadListResult struct {
	Squads []SquadInfo `json:"squads"`
}

// SquadGetParams is the request for squads.get.
type SquadGetParams struct {
	Name string `json:"name"`
}

// SquadGetResult is the response to squads.get.
type SquadGetResult struct {
	Squad SquadInfo `json:"squad"`
}

// SquadStartParams is the request for squads.start.
type SquadStartParams struct {
	Name string `json:"name"`
}

// SquadStartResult is the response to squads.start.
type SquadStartResult struct {
	Success bool `json:"success"`
}

// SquadStopParams is the request for squads.stop.
type SquadStopParams struct {
	Name string `json:"name"`
}

// SquadStopResult is the response to squads.stop.
type SquadStopResult struct {
	Success bool `json:"success"`
}

// SquadAddAgentParams is the request for squads.add_agent.
type SquadAddAgentParams struct {
	Name        string `json:"name"`
	AgentHandle string `json:"agent_handle"`
}

// SquadAddAgentResult is the response to squads.add_agent.
type SquadAddAgentResult struct {
	Success bool `json:"success"`
}

// SquadRemoveAgentParams is the request for squads.remove_agent.
type SquadRemoveAgentParams struct {
	Name        string `json:"name"`
	AgentHandle string `json:"agent_handle"`
}

// SquadRemoveAgentResult is the response to squads.remove_agent.
type SquadRemoveAgentResult struct {
	Success bool `json:"success"`
}

// SquadTicketsReadyParams is the request for squads.tickets_ready.
// Called by @ayo after creating tickets in a squad to trigger agent spawning.
type SquadTicketsReadyParams struct {
	Name string `json:"name"` // Squad name
}

// SquadTicketsReadyResult is the response to squads.tickets_ready.
type SquadTicketsReadyResult struct {
	TicketsFound  int      `json:"tickets_found"`
	AgentsSpawned []string `json:"agents_spawned,omitempty"`
}

// SquadNotifyAgentsParams is the request for squads.notify_agents.
type SquadNotifyAgentsParams struct {
	Name string `json:"name"` // Squad name
}

// SquadNotifyAgentsResult is the response to squads.notify_agents.
type SquadNotifyAgentsResult struct {
	SessionsSpawned []string `json:"sessions_spawned,omitempty"`
	TicketsAssigned int      `json:"tickets_assigned"`
}

// SquadWaitCompletionParams is the request for squads.wait_completion.
type SquadWaitCompletionParams struct {
	Name    string `json:"name"`    // Squad name
	Timeout int    `json:"timeout"` // Timeout in seconds (0 = no timeout)
}

// SquadWaitCompletionResult is the response to squads.wait_completion.
type SquadWaitCompletionResult struct {
	Completed     bool   `json:"completed"`
	TicketsClosed int    `json:"tickets_closed"`
	TicketsOpen   int    `json:"tickets_open"`
	TimedOut      bool   `json:"timed_out,omitempty"`
}

// SquadSyncOutputParams is the request for squads.sync_output.
type SquadSyncOutputParams struct {
	Name       string `json:"name"`        // Squad name
	TargetPath string `json:"target_path"` // Target directory on host
}

// SquadSyncOutputResult is the response to squads.sync_output.
type SquadSyncOutputResult struct {
	FilesCopied int `json:"files_copied"`
	BytesCopied int `json:"bytes_copied"`
}

// SquadCleanupParams is the request for squads.cleanup.
type SquadCleanupParams struct {
	Name string `json:"name"` // Squad name
}

// SquadCleanupResult is the response to squads.cleanup.
type SquadCleanupResult struct {
	Success bool `json:"success"`
}

// SquadDispatchParams is the request for squads.dispatch.
type SquadDispatchParams struct {
	// Name is the squad name (without # prefix).
	Name string `json:"name"`

	// Prompt is a free-form text prompt for the squad.
	Prompt string `json:"prompt,omitempty"`

	// Data contains structured input data.
	// If the squad has an input schema, this data is validated against it.
	Data map[string]any `json:"data,omitempty"`

	// StartIfStopped starts the squad if it's not running.
	StartIfStopped bool `json:"start_if_stopped,omitempty"`

	// Timeout is the maximum time to wait for a result in seconds.
	// If zero, a default timeout is used.
	Timeout int `json:"timeout,omitempty"`
}

// SquadDispatchResult is the response to squads.dispatch.
type SquadDispatchResult struct {
	// Output contains structured output data.
	Output map[string]any `json:"output,omitempty"`

	// Raw is the raw text output if not structured.
	Raw string `json:"raw,omitempty"`

	// Error contains any error message from the squad.
	Error string `json:"error,omitempty"`
}

// AgentInvokeParams is the request for agent.invoke.
type AgentInvokeParams struct {
	// Agent is the agent handle (e.g., "@ayo", "@backend").
	Agent string `json:"agent"`

	// Prompt is the task or question for the agent.
	Prompt string `json:"prompt"`

	// SessionID optionally specifies an existing session to continue.
	SessionID string `json:"session_id,omitempty"`

	// Skills are additional skills to enable for this invocation.
	Skills []string `json:"skills,omitempty"`
}

// AgentInvokeResult is the response to agent.invoke.
type AgentInvokeResult struct {
	// SessionID is the session ID used for this invocation.
	SessionID string `json:"session_id"`

	// Response is the agent's response text.
	Response string `json:"response"`

	// Error contains any error message from the agent.
	Error string `json:"error,omitempty"`
}
