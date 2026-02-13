package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

// Client connects to the daemon server.
type Client struct {
	conn      net.Conn
	reader    *bufio.Reader
	encoder   *json.Encoder
	nextID    atomic.Int64
	connected bool
}

// NewClient creates a new daemon client.
func NewClient() *Client {
	return &Client{}
}

// Connect connects to the daemon server at the default socket path.
func (c *Client) Connect(ctx context.Context) error {
	return c.ConnectTo(ctx, DefaultSocketPath())
}

// ConnectTo connects to the daemon server at the specified socket path.
func (c *Client) ConnectTo(ctx context.Context, socketPath string) error {
	if c.connected {
		return nil
	}

	var conn net.Conn
	var err error

	if runtime.GOOS == "windows" {
		// Windows: try to read port from socket path
		conn, err = net.DialTimeout("tcp", "127.0.0.1:0", 5*time.Second)
	} else {
		conn, err = net.DialTimeout("unix", socketPath, 5*time.Second)
	}
	if err != nil {
		return fmt.Errorf("connect to daemon: %w", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.encoder = json.NewEncoder(conn)
	c.connected = true

	return nil
}

// Close closes the connection to the daemon.
func (c *Client) Close() error {
	if !c.connected {
		return nil
	}
	c.connected = false
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected returns true if connected to the daemon.
func (c *Client) IsConnected() bool {
	return c.connected
}

// call sends a request and waits for a response.
func (c *Client) call(_ context.Context, method string, params any, result any) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}

	id := c.nextID.Add(1)

	req, err := NewRequest(method, params, id)
	if err != nil {
		return err
	}

	// Send request
	if err := c.encoder.Encode(req); err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	// Read response
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if resp.Error != nil {
		return resp.Error
	}

	if result != nil && resp.Result != nil {
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("parse result: %w", err)
		}
	}

	return nil
}

// Ping pings the daemon.
func (c *Client) Ping(ctx context.Context) error {
	var result PingResult
	if err := c.call(ctx, MethodPing, nil, &result); err != nil {
		return err
	}
	if !result.Pong {
		return fmt.Errorf("invalid ping response")
	}
	return nil
}

// Status returns the daemon status.
func (c *Client) Status(ctx context.Context) (*StatusResult, error) {
	var result StatusResult
	if err := c.call(ctx, MethodStatus, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Shutdown requests the daemon to shut down.
func (c *Client) Shutdown(ctx context.Context, graceful bool) error {
	params := ShutdownParams{Graceful: graceful}
	return c.call(ctx, MethodShutdown, params, nil)
}

// SandboxAcquire acquires a sandbox for an agent.
func (c *Client) SandboxAcquire(ctx context.Context, agent string, timeout int) (*SandboxAcquireResult, error) {
	params := SandboxAcquireParams{Agent: agent, Timeout: timeout}
	var result SandboxAcquireResult
	if err := c.call(ctx, MethodSandboxAcquire, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SandboxRelease releases a sandbox.
func (c *Client) SandboxRelease(ctx context.Context, sandboxID string) error {
	params := SandboxReleaseParams{SandboxID: sandboxID}
	return c.call(ctx, MethodSandboxRelease, params, nil)
}

// SandboxExec executes a command in a sandbox.
func (c *Client) SandboxExec(ctx context.Context, sandboxID, command, workingDir string, timeout int) (*SandboxExecResult, error) {
	params := SandboxExecParams{
		SandboxID:  sandboxID,
		Command:    command,
		WorkingDir: workingDir,
		Timeout:    timeout,
	}
	var result SandboxExecResult
	if err := c.call(ctx, MethodSandboxExec, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SandboxStatus returns the sandbox pool status.
func (c *Client) SandboxStatus(ctx context.Context) (*SandboxStatusResult, error) {
	var result SandboxStatusResult
	if err := c.call(ctx, MethodSandboxStatus, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SessionList returns all active agent sessions.
func (c *Client) SessionList(ctx context.Context) (*SessionListResult, error) {
	var result SessionListResult
	if err := c.call(ctx, MethodSessionList, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SessionStart starts a new session for an agent.
func (c *Client) SessionStart(ctx context.Context, agentHandle string, triggerID, sessionID string) (*SessionStartResult, error) {
	params := SessionStartParams{
		AgentHandle: agentHandle,
		TriggerID:   triggerID,
		SessionID:   sessionID,
	}
	var result SessionStartResult
	if err := c.call(ctx, MethodSessionStart, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SessionStop stops a session.
func (c *Client) SessionStop(ctx context.Context, sessionID string) error {
	params := SessionStopParams{SessionID: sessionID}
	return c.call(ctx, MethodSessionStop, params, nil)
}

// AgentWake wakes up (starts a session for) an agent.
func (c *Client) AgentWake(ctx context.Context, handle string) (*AgentWakeResult, error) {
	params := AgentWakeParams{Handle: handle}
	var result AgentWakeResult
	if err := c.call(ctx, MethodAgentWake, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AgentSleep puts an agent to sleep (stops its session).
func (c *Client) AgentSleep(ctx context.Context, handle string) error {
	params := AgentSleepParams{Handle: handle}
	return c.call(ctx, MethodAgentSleep, params, nil)
}

// AgentStatus returns the status of an agent's session.
func (c *Client) AgentStatus(ctx context.Context, handle string) (*AgentStatusResult, error) {
	params := AgentStatusParams{Handle: handle}
	var result AgentStatusResult
	if err := c.call(ctx, MethodAgentStatus, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TriggerList returns all registered triggers.
func (c *Client) TriggerList(ctx context.Context) (*TriggerListResult, error) {
	var result TriggerListResult
	if err := c.call(ctx, MethodTriggerList, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TriggerGet returns a trigger by ID.
func (c *Client) TriggerGet(ctx context.Context, id string) (*TriggerGetResult, error) {
	params := TriggerGetParams{ID: id}
	var result TriggerGetResult
	if err := c.call(ctx, MethodTriggerGet, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TriggerRegister registers a new trigger.
func (c *Client) TriggerRegister(ctx context.Context, params TriggerRegisterParams) (*TriggerRegisterResult, error) {
	var result TriggerRegisterResult
	if err := c.call(ctx, MethodTriggerRegister, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TriggerRemove removes a trigger by ID.
func (c *Client) TriggerRemove(ctx context.Context, id string) error {
	params := TriggerRemoveParams{ID: id}
	return c.call(ctx, MethodTriggerRemove, params, nil)
}

// TriggerTest fires a trigger manually for testing.
func (c *Client) TriggerTest(ctx context.Context, id string) error {
	params := TriggerTestParams{ID: id}
	return c.call(ctx, MethodTriggerTest, params, nil)
}

// TriggerSetEnabled enables or disables a trigger.
func (c *Client) TriggerSetEnabled(ctx context.Context, id string, enabled bool) error {
	params := TriggerSetEnabledParams{ID: id, Enabled: enabled}
	return c.call(ctx, MethodTriggerSetEnabled, params, nil)
}

// SandboxJoin adds an agent to an existing sandbox.
func (c *Client) SandboxJoin(ctx context.Context, sandboxID, agent string) error {
	params := SandboxJoinParams{SandboxID: sandboxID, Agent: agent}
	return c.call(ctx, MethodSandboxJoin, params, nil)
}

// SandboxAgents returns the list of agents in a sandbox.
func (c *Client) SandboxAgents(ctx context.Context, sandboxID string) ([]string, error) {
	params := SandboxAgentsParams{SandboxID: sandboxID}
	var result SandboxAgentsResult
	if err := c.call(ctx, MethodSandboxAgents, params, &result); err != nil {
		return nil, err
	}
	return result.Agents, nil
}

// IsDaemonRunning checks if the daemon is running.
func IsDaemonRunning() bool {
	pidPath := DefaultPIDPath()
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds - need to send signal 0 to check
	if runtime.GOOS != "windows" {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			return false
		}
	}

	return true
}

// ConnectOrStart connects to the daemon, starting it if necessary.
func ConnectOrStart(ctx context.Context) (*Client, error) {
	client := NewClient()

	// Try to connect first
	if err := client.Connect(ctx); err == nil {
		// Verify connection with ping
		if err := client.Ping(ctx); err == nil {
			return client, nil
		}
		client.Close()
	}

	// Daemon not running - try to start it
	if err := StartDaemonBackground(); err != nil {
		return nil, fmt.Errorf("start daemon: %w", err)
	}

	// Wait for daemon to be ready (up to 10 seconds)
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)

		if err := client.Connect(ctx); err == nil {
			if err := client.Ping(ctx); err == nil {
				return client, nil
			}
			client.Close()
		}
	}

	return nil, fmt.Errorf("daemon started but not responding")
}

// StartDaemonBackground starts the daemon in the background.
func StartDaemonBackground() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}

	// Start in background using nohup-style approach
	procAttr := &os.ProcAttr{
		Dir:   "/",
		Env:   os.Environ(),
		Files: []*os.File{nil, nil, nil}, // Detach from terminal
	}

	process, err := os.StartProcess(exe, []string{exe, "daemon", "start", "--foreground"}, procAttr)
	if err != nil {
		return fmt.Errorf("start daemon process: %w", err)
	}

	// Release the process so it continues after we exit
	if err := process.Release(); err != nil {
		return fmt.Errorf("release process: %w", err)
	}

	return nil
}

// Call makes a generic RPC call. For use when specific client methods are not available.
func (c *Client) Call(ctx context.Context, method string, params any, result any) error {
	return c.call(ctx, method, params, result)
}

// MatrixStatus returns the Matrix subsystem status.
func (c *Client) MatrixStatus(ctx context.Context) (*MatrixStatusResult, error) {
	var result MatrixStatusResult
	if err := c.call(ctx, MethodMatrixStatus, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MatrixRoomsList lists Matrix rooms.
func (c *Client) MatrixRoomsList(ctx context.Context, params MatrixRoomsListParams) (*MatrixRoomsListResult, error) {
	var result MatrixRoomsListResult
	if err := c.call(ctx, MethodMatrixRoomsList, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MatrixRoomsCreate creates a Matrix room.
func (c *Client) MatrixRoomsCreate(ctx context.Context, params MatrixRoomsCreateParams) (*MatrixRoomsCreateResult, error) {
	var result MatrixRoomsCreateResult
	if err := c.call(ctx, MethodMatrixRoomsCreate, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MatrixRoomsMembers gets members of a Matrix room.
func (c *Client) MatrixRoomsMembers(ctx context.Context, params MatrixRoomsMembersParams) (*MatrixRoomsMembersResult, error) {
	var result MatrixRoomsMembersResult
	if err := c.call(ctx, MethodMatrixRoomsMembers, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MatrixRoomsInvite invites an agent to a room.
func (c *Client) MatrixRoomsInvite(ctx context.Context, params MatrixRoomsInviteParams) error {
	return c.call(ctx, MethodMatrixRoomsInvite, params, nil)
}

// MatrixSend sends a message to a Matrix room.
func (c *Client) MatrixSend(ctx context.Context, params MatrixSendParams) (*MatrixSendResult, error) {
	var result MatrixSendResult
	if err := c.call(ctx, MethodMatrixSend, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MatrixRead reads messages from a Matrix room.
func (c *Client) MatrixRead(ctx context.Context, params MatrixReadParams) (*MatrixReadResult, error) {
	var result MatrixReadResult
	if err := c.call(ctx, MethodMatrixRead, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TicketCreate creates a new ticket.
func (c *Client) TicketCreate(ctx context.Context, params TicketCreateParams) (*TicketCreateResult, error) {
	var result TicketCreateResult
	if err := c.call(ctx, MethodTicketCreate, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TicketGet retrieves a ticket by ID.
func (c *Client) TicketGet(ctx context.Context, sessionID, ticketID string) (*TicketGetResult, error) {
	params := TicketGetParams{SessionID: sessionID, TicketID: ticketID}
	var result TicketGetResult
	if err := c.call(ctx, MethodTicketGet, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TicketList lists tickets with optional filters.
func (c *Client) TicketList(ctx context.Context, params TicketListParams) (*TicketListResult, error) {
	var result TicketListResult
	if err := c.call(ctx, MethodTicketList, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TicketStart sets a ticket status to in_progress.
func (c *Client) TicketStart(ctx context.Context, sessionID, ticketID string) error {
	params := TicketStatusParams{SessionID: sessionID, TicketID: ticketID}
	return c.call(ctx, MethodTicketStart, params, nil)
}

// TicketClose sets a ticket status to closed.
func (c *Client) TicketClose(ctx context.Context, sessionID, ticketID string) error {
	params := TicketStatusParams{SessionID: sessionID, TicketID: ticketID}
	return c.call(ctx, MethodTicketClose, params, nil)
}

// TicketReopen reopens a closed ticket.
func (c *Client) TicketReopen(ctx context.Context, sessionID, ticketID string) error {
	params := TicketStatusParams{SessionID: sessionID, TicketID: ticketID}
	return c.call(ctx, MethodTicketReopen, params, nil)
}

// TicketBlock sets a ticket status to blocked.
func (c *Client) TicketBlock(ctx context.Context, sessionID, ticketID string) error {
	params := TicketStatusParams{SessionID: sessionID, TicketID: ticketID}
	return c.call(ctx, MethodTicketBlock, params, nil)
}

// TicketAssign assigns a ticket to an agent.
func (c *Client) TicketAssign(ctx context.Context, sessionID, ticketID, assignee string) error {
	params := TicketAssignParams{SessionID: sessionID, TicketID: ticketID, Assignee: assignee}
	return c.call(ctx, MethodTicketAssign, params, nil)
}

// TicketAddNote adds a note to a ticket.
func (c *Client) TicketAddNote(ctx context.Context, sessionID, ticketID, content string) error {
	params := TicketAddNoteParams{SessionID: sessionID, TicketID: ticketID, Content: content}
	return c.call(ctx, MethodTicketAddNote, params, nil)
}

// TicketReady returns tickets ready to work on (deps resolved).
func (c *Client) TicketReady(ctx context.Context, sessionID, assignee string) (*TicketReadyResult, error) {
	params := TicketReadyParams{SessionID: sessionID, Assignee: assignee}
	var result TicketReadyResult
	if err := c.call(ctx, MethodTicketReady, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TicketBlocked returns tickets blocked on dependencies.
func (c *Client) TicketBlocked(ctx context.Context, sessionID, assignee string) (*TicketBlockedResult, error) {
	params := TicketBlockedParams{SessionID: sessionID, Assignee: assignee}
	var result TicketBlockedResult
	if err := c.call(ctx, MethodTicketBlocked, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TicketAddDep adds a dependency to a ticket.
func (c *Client) TicketAddDep(ctx context.Context, sessionID, ticketID, depID string) error {
	params := TicketDepParams{SessionID: sessionID, TicketID: ticketID, DepID: depID}
	return c.call(ctx, MethodTicketAddDep, params, nil)
}

// TicketRemoveDep removes a dependency from a ticket.
func (c *Client) TicketRemoveDep(ctx context.Context, sessionID, ticketID, depID string) error {
	params := TicketDepParams{SessionID: sessionID, TicketID: ticketID, DepID: depID}
	return c.call(ctx, MethodTicketRemDep, params, nil)
}

// TicketDelete deletes a ticket.
func (c *Client) TicketDelete(ctx context.Context, sessionID, ticketID string) error {
	params := TicketDeleteParams{SessionID: sessionID, TicketID: ticketID}
	return c.call(ctx, MethodTicketDelete, params, nil)
}

// Squad management client methods

// SquadCreate creates a new squad.
func (c *Client) SquadCreate(ctx context.Context, params SquadCreateParams) (*SquadCreateResult, error) {
	var result SquadCreateResult
	if err := c.call(ctx, MethodSquadCreate, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SquadDestroy destroys a squad.
func (c *Client) SquadDestroy(ctx context.Context, name string, deleteData bool) error {
	params := SquadDestroyParams{Name: name, DeleteData: deleteData}
	return c.call(ctx, MethodSquadDestroy, params, nil)
}

// SquadList lists all squads.
func (c *Client) SquadList(ctx context.Context) (*SquadListResult, error) {
	var result SquadListResult
	if err := c.call(ctx, MethodSquadList, SquadListParams{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SquadGet gets a squad by name.
func (c *Client) SquadGet(ctx context.Context, name string) (*SquadGetResult, error) {
	params := SquadGetParams{Name: name}
	var result SquadGetResult
	if err := c.call(ctx, MethodSquadGet, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SquadStart starts a squad sandbox.
func (c *Client) SquadStart(ctx context.Context, name string) error {
	params := SquadStartParams{Name: name}
	return c.call(ctx, MethodSquadStart, params, nil)
}

// SquadStop stops a squad sandbox.
func (c *Client) SquadStop(ctx context.Context, name string) error {
	params := SquadStopParams{Name: name}
	return c.call(ctx, MethodSquadStop, params, nil)
}

// SquadAddAgent adds an agent to a squad.
func (c *Client) SquadAddAgent(ctx context.Context, name, agentHandle string) error {
	params := SquadAddAgentParams{Name: name, AgentHandle: agentHandle}
	return c.call(ctx, MethodSquadAddAgent, params, nil)
}

// SquadRemoveAgent removes an agent from a squad.
func (c *Client) SquadRemoveAgent(ctx context.Context, name, agentHandle string) error {
	params := SquadRemoveAgentParams{Name: name, AgentHandle: agentHandle}
	return c.call(ctx, MethodSquadRemoveAgent, params, nil)
}
