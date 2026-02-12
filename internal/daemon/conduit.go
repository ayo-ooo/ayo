package daemon

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
)

// ConduitProcess manages the Conduit Matrix homeserver subprocess.
type ConduitProcess struct {
	binaryPath string
	dataDir    string
	socketPath string
	configPath string
	logPath    string

	cmd     *exec.Cmd
	logFile *os.File

	mu           sync.RWMutex
	running      bool
	pid          int
	restarts     int
	lastRestart  time.Time
	startTime    time.Time
	healthCancel context.CancelFunc
}

// ConduitStatus represents the status of the Conduit process.
type ConduitStatus struct {
	Running     bool      `json:"running"`
	Pid         int       `json:"pid,omitempty"`
	Restarts    int       `json:"restarts"`
	LastRestart time.Time `json:"last_restart,omitempty"`
	Healthy     bool      `json:"healthy"`
	SocketPath  string    `json:"socket_path"`
	DataDir     string    `json:"data_dir"`
	Version     string    `json:"version,omitempty"`
	Uptime      int64     `json:"uptime_seconds,omitempty"`
}

// NewConduitProcess creates a new Conduit process manager.
func NewConduitProcess() *ConduitProcess {
	return &ConduitProcess{
		binaryPath: paths.ConduitBinary(),
		dataDir:    paths.MatrixDataDir(),
		socketPath: paths.MatrixSocket(),
		configPath: paths.ConduitConfig(),
		logPath:    paths.ConduitLogPath(),
	}
}

// Start starts the Conduit process.
func (c *ConduitProcess) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("conduit already running")
	}

	// Locate binary
	binaryPath, err := c.locateBinary()
	if err != nil {
		return fmt.Errorf("locate conduit binary: %w", err)
	}
	c.binaryPath = binaryPath

	// Ensure directories exist
	if err := os.MkdirAll(c.dataDir, 0700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(c.socketPath), 0755); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(c.logPath), 0755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	// Generate config if needed
	if err := c.ensureConfig(); err != nil {
		return fmt.Errorf("ensure config: %w", err)
	}

	// Remove stale socket
	os.Remove(c.socketPath)

	// Open log file
	logFile, err := os.OpenFile(c.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	c.logFile = logFile

	// Start process
	c.cmd = exec.CommandContext(ctx, c.binaryPath)
	c.cmd.Dir = c.dataDir
	c.cmd.Env = append(os.Environ(),
		"CONDUIT_CONFIG="+c.configPath,
	)
	c.cmd.Stdout = logFile
	c.cmd.Stderr = logFile

	if err := c.cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("start conduit: %w", err)
	}

	c.running = true
	c.pid = c.cmd.Process.Pid
	c.startTime = time.Now()

	// Wait for socket to appear
	if err := c.waitForSocket(ctx, 30*time.Second); err != nil {
		c.stopLocked()
		return fmt.Errorf("conduit failed to start: %w", err)
	}

	// Start health check goroutine
	healthCtx, cancel := context.WithCancel(ctx)
	c.healthCancel = cancel
	go c.healthCheckLoop(healthCtx)

	// Start process monitor
	go c.monitorProcess(ctx)

	return nil
}

// Stop stops the Conduit process.
func (c *ConduitProcess) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stopLocked()
}

func (c *ConduitProcess) stopLocked() error {
	if !c.running {
		return nil
	}

	// Cancel health check
	if c.healthCancel != nil {
		c.healthCancel()
	}

	// Send SIGTERM
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Signal(syscall.SIGTERM)

		// Wait up to 5 seconds
		done := make(chan error, 1)
		go func() {
			done <- c.cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited
		case <-time.After(5 * time.Second):
			// Force kill
			c.cmd.Process.Kill()
			<-done
		}
	}

	// Clean up
	if c.logFile != nil {
		c.logFile.Close()
		c.logFile = nil
	}
	os.Remove(c.socketPath)

	c.running = false
	c.pid = 0
	c.cmd = nil

	return nil
}

// Running returns true if Conduit is running.
func (c *ConduitProcess) Running() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// Pid returns the process ID.
func (c *ConduitProcess) Pid() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pid
}

// SocketPath returns the Unix socket path.
func (c *ConduitProcess) SocketPath() string {
	return c.socketPath
}

// Status returns the current status.
func (c *ConduitProcess) Status() ConduitStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := ConduitStatus{
		Running:     c.running,
		Pid:         c.pid,
		Restarts:    c.restarts,
		LastRestart: c.lastRestart,
		Healthy:     c.running && c.isHealthyLocked(),
		SocketPath:  c.socketPath,
		DataDir:     c.dataDir,
	}

	if c.running && !c.startTime.IsZero() {
		status.Uptime = int64(time.Since(c.startTime).Seconds())
	}

	return status
}

// locateBinary finds the Conduit binary.
func (c *ConduitProcess) locateBinary() (string, error) {
	// Check configured path first
	if c.binaryPath != "" {
		if _, err := os.Stat(c.binaryPath); err == nil {
			return c.binaryPath, nil
		}
	}

	// Check default path
	defaultPath := paths.ConduitBinary()
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, nil
	}

	// Check PATH
	if path, err := exec.LookPath("conduit"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("conduit binary not found; install with: cargo install conduit")
}

// ensureConfig generates the Conduit config file if it doesn't exist.
func (c *ConduitProcess) ensureConfig() error {
	if _, err := os.Stat(c.configPath); err == nil {
		// Config already exists
		return nil
	}

	config := fmt.Sprintf(`[global]
server_name = "ayo.local"
database_backend = "rocksdb"
database_path = "%s"

# Use Unix socket for local-only access
port = 0
unix_socket_path = "%s"

# Allow registration for agents
allow_registration = true
allow_encryption = true

# Disable federation (local only)
allow_federation = false

# Logging
log = "info"

# Allow guests for temporary connections
allow_guest_registration = false

# Room settings
max_request_size = 20000000
`, c.dataDir, c.socketPath)

	if err := os.WriteFile(c.configPath, []byte(config), 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// waitForSocket waits for the Unix socket to appear.
func (c *ConduitProcess) waitForSocket(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if _, err := os.Stat(c.socketPath); err == nil {
			// Socket exists, try to connect
			conn, err := net.DialTimeout("unix", c.socketPath, time.Second)
			if err == nil {
				conn.Close()
				return nil
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for conduit socket")
}

// healthCheckLoop periodically checks Conduit health.
func (c *ConduitProcess) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !c.isHealthy() {
				c.mu.Lock()
				if c.running {
					// Process died, attempt restart
					c.restarts++
					c.lastRestart = time.Now()

					// Limit restarts
					if c.restarts > 3 {
						// Too many restarts, give up
						c.mu.Unlock()
						return
					}

					// Restart
					c.stopLocked()
					c.mu.Unlock()

					time.Sleep(time.Second)

					if err := c.Start(ctx); err != nil {
						// Failed to restart
						return
					}
				} else {
					c.mu.Unlock()
				}
			}
		}
	}
}

// monitorProcess waits for the process to exit.
func (c *ConduitProcess) monitorProcess(_ context.Context) {
	if c.cmd == nil {
		return
	}

	err := c.cmd.Wait()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		c.running = false
		c.pid = 0
		if err != nil {
			// Process exited unexpectedly
			// The health check loop will handle restart
		}
	}
}

// isHealthy checks if Conduit is healthy.
func (c *ConduitProcess) isHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isHealthyLocked()
}

func (c *ConduitProcess) isHealthyLocked() bool {
	if !c.running {
		return false
	}

	// Check socket exists
	if _, err := os.Stat(c.socketPath); err != nil {
		return false
	}

	// HTTP health check
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", c.socketPath)
			},
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://localhost/_matrix/client/versions")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}
