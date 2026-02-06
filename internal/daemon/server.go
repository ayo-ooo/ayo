package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
	"github.com/alexcabrera/ayo/internal/version"
)

// Server is the daemon server that manages ayo resources.
type Server struct {
	listener    net.Listener
	pool        *sandbox.Pool
	provider    providers.SandboxProvider
	startTime   time.Time
	shutdownCh  chan struct{}
	wg          sync.WaitGroup
	connections atomic.Int32
	mu          sync.RWMutex
	running     bool
}

// ServerConfig configures the daemon server.
type ServerConfig struct {
	SocketPath   string
	PoolConfig   sandbox.PoolConfig
	IdleTimeout  time.Duration
}

// DefaultServerConfig returns the default server configuration.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		SocketPath: DefaultSocketPath(),
		PoolConfig: sandbox.PoolConfig{
			Name:    "daemon",
			MinSize: 1,
			MaxSize: 4,
		},
		IdleTimeout: 30 * time.Minute,
	}
}

// DefaultSocketPath returns the default socket path for the platform.
func DefaultSocketPath() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\ayo-daemon`
	}
	return filepath.Join(paths.DataDir(), "daemon.sock")
}

// DefaultPIDPath returns the default PID file path.
func DefaultPIDPath() string {
	return filepath.Join(paths.DataDir(), "daemon.pid")
}

// NewServer creates a new daemon server.
func NewServer(cfg ServerConfig) (*Server, error) {
	// Select the best available sandbox provider
	provider := selectSandboxProvider()

	// Create pool
	pool := sandbox.NewPool(cfg.PoolConfig, provider)

	return &Server{
		provider:   provider,
		pool:       pool,
		shutdownCh: make(chan struct{}),
	}, nil
}

// Start starts the daemon server.
func (s *Server) Start(ctx context.Context, socketPath string) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.startTime = time.Now()
	s.mu.Unlock()

	// Remove stale socket file
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove stale socket: %w", err)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}

	// Start listening
	var err error
	if runtime.GOOS == "windows" {
		// Windows named pipe - for now just use TCP localhost
		s.listener, err = net.Listen("tcp", "127.0.0.1:0")
	} else {
		s.listener, err = net.Listen("unix", socketPath)
	}
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	// Start sandbox pool
	if err := s.pool.Start(ctx); err != nil {
		s.listener.Close()
		return fmt.Errorf("start pool: %w", err)
	}

	// Write PID file
	if err := s.writePIDFile(); err != nil {
		s.listener.Close()
		return fmt.Errorf("write pid file: %w", err)
	}

	// Accept connections
	s.wg.Add(1)
	go s.acceptLoop(ctx)

	return nil
}

// Stop gracefully stops the daemon server.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	// Signal shutdown
	close(s.shutdownCh)

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Stop sandbox pool
	if s.pool != nil {
		s.pool.Stop(ctx)
	}

	// Wait for connections to close
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Remove PID file
	os.Remove(DefaultPIDPath())

	return nil
}

// Addr returns the listener address.
func (s *Server) Addr() net.Addr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

func (s *Server) acceptLoop(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdownCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.shutdownCh:
				return
			default:
				continue
			}
		}

		s.connections.Add(1)
		s.wg.Add(1)
		go s.handleConnection(ctx, conn)
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		conn.Close()
		s.connections.Add(-1)
		s.wg.Done()
	}()

	reader := bufio.NewReader(conn)
	encoder := json.NewEncoder(conn)

	for {
		select {
		case <-s.shutdownCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		// Read request
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			resp := NewErrorResponse(NewError(ErrCodeParse, "parse error"), 0)
			encoder.Encode(resp)
			continue
		}

		// Handle request
		resp := s.handleRequest(ctx, &req)
		if err := encoder.Encode(resp); err != nil {
			return
		}
	}
}

func (s *Server) handleRequest(ctx context.Context, req *Request) *Response {
	switch req.Method {
	case MethodPing:
		return s.handlePing(req)
	case MethodStatus:
		return s.handleStatus(req)
	case MethodShutdown:
		return s.handleShutdown(req)
	case MethodSandboxAcquire:
		return s.handleSandboxAcquire(ctx, req)
	case MethodSandboxRelease:
		return s.handleSandboxRelease(ctx, req)
	case MethodSandboxExec:
		return s.handleSandboxExec(ctx, req)
	case MethodSandboxStatus:
		return s.handleSandboxStatus(req)
	default:
		return NewErrorResponse(NewError(ErrCodeMethodNotFound, "method not found: "+req.Method), req.ID)
	}
}

func (s *Server) handlePing(req *Request) *Response {
	resp, _ := NewResponse(PingResult{Pong: true}, req.ID)
	return resp
}

func (s *Server) handleStatus(req *Request) *Response {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	poolStatus := s.pool.Status()

	result := StatusResult{
		Running:     true,
		Uptime:      int64(time.Since(s.startTime).Seconds()),
		PID:         os.Getpid(),
		Version:     version.Version,
		MemoryUsage: int64(mem.Alloc),
		Sandboxes: SandboxStatusResult{
			Total: poolStatus.Total,
			Idle:  poolStatus.Idle,
			InUse: poolStatus.InUse,
		},
	}

	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleShutdown(req *Request) *Response {
	var params ShutdownParams
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	// Respond first, then shutdown
	resp, _ := NewResponse(struct{}{}, req.ID)

	// Trigger shutdown in background
	go func() {
		time.Sleep(100 * time.Millisecond) // Let response be sent
		s.Stop(context.Background())
	}()

	return resp
}

func (s *Server) handleSandboxAcquire(ctx context.Context, req *Request) *Response {
	var params SandboxAcquireParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if params.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(params.Timeout)*time.Second)
		defer cancel()
	}

	sb, err := s.pool.Acquire(ctx, params.Agent)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeSandboxExhausted, err.Error()), req.ID)
	}

	result := SandboxAcquireResult{
		SandboxID:  sb.ID,
		WorkingDir: "/",
	}

	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSandboxRelease(ctx context.Context, req *Request) *Response {
	var params SandboxReleaseParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if err := s.pool.Release(ctx, params.SandboxID); err != nil {
		return NewErrorResponse(NewError(ErrCodeSandboxNotFound, err.Error()), req.ID)
	}

	resp, _ := NewResponse(struct{}{}, req.ID)
	return resp
}

func (s *Server) handleSandboxExec(ctx context.Context, req *Request) *Response {
	var params SandboxExecParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	timeout := 30 * time.Second
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Second
	}

	execResult, err := s.pool.Exec(ctx, params.SandboxID, providers.ExecOptions{
		Command:    params.Command,
		WorkingDir: params.WorkingDir,
		Timeout:    timeout,
	})
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	result := SandboxExecResult{
		Stdout:   execResult.Stdout,
		Stderr:   execResult.Stderr,
		ExitCode: execResult.ExitCode,
		TimedOut: execResult.TimedOut,
	}

	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSandboxStatus(req *Request) *Response {
	poolStatus := s.pool.Status()
	result := SandboxStatusResult{
		Total: poolStatus.Total,
		Idle:  poolStatus.Idle,
		InUse: poolStatus.InUse,
	}

	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) writePIDFile() error {
	pidPath := DefaultPIDPath()
	return os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

// selectSandboxProvider returns the best available sandbox provider for the current platform.
// Priority:
// 1. Apple Container (macOS 26+ on Apple Silicon)
// 2. systemd-nspawn (Linux with systemd)
// 3. None (host execution, no isolation)
func selectSandboxProvider() providers.SandboxProvider {
	// Try Apple Container on macOS
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		appleProvider := sandbox.NewAppleProvider()
		if appleProvider.IsAvailable() {
			return appleProvider
		}
	}

	// Try systemd-nspawn on Linux
	if runtime.GOOS == "linux" {
		linuxProvider := sandbox.NewLinuxProvider()
		if linuxProvider.IsAvailable() {
			return linuxProvider
		}
	}

	// Fall back to none provider (no isolation)
	return sandbox.NewNoneProvider()
}
