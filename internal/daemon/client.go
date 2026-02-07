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
func (c *Client) call(ctx context.Context, method string, params any, result any) error {
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
