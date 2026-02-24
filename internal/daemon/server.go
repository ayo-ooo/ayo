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

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/planners"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/squads"
	ayosync "github.com/alexcabrera/ayo/internal/sync"
	"github.com/alexcabrera/ayo/internal/tickets"
	"github.com/alexcabrera/ayo/internal/version"
)

// Server is the daemon server that manages ayo resources.
type Server struct {
	config         config.Config
	services       *session.Services
	listener       net.Listener
	pool           *sandbox.Pool
	provider       providers.SandboxProvider
	sessionManager *DaemonSessionManager
	triggerEngine  *TriggerEngine
	squadRPC       *SquadRPC
	startTime      time.Time
	shutdownCh     chan struct{}
	wg             sync.WaitGroup
	connections    atomic.Int32
	mu             sync.RWMutex
	running        bool
	tickets        *tickets.Service
	squadTickets   *tickets.SquadTicketService
	ticketWatcher  *TicketWatcher
}

// ServerConfig configures the daemon server.
type ServerConfig struct {
	Config          config.Config
	Services        *session.Services // optional, for session persistence
	SocketPath      string
	PoolConfig      sandbox.PoolConfig
	IdleTimeout     time.Duration
	MaxConcurrent   int    // max concurrent agent executions (default: 3)
}

// DefaultServerConfig returns the default server configuration.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		SocketPath: DefaultSocketPath(),
		PoolConfig: sandbox.PoolConfig{
			Name:    "daemon",
			MinSize: 1,
			MaxSize: 4,
			Mounts: []providers.Mount{
				{
					Source:      ayosync.HomesDir(),
					Destination: "/home",
					Mode:        providers.MountModeBind,
					ReadOnly:    false,
				},
				{
					Source:      ayosync.SharedDir(),
					Destination: "/shared",
					Mode:        providers.MountModeBind,
					ReadOnly:    false,
				},
				{
					Source:      ayosync.WorkspaceDir(),
					Destination: "/workspace",
					Mode:        providers.MountModeBind,
					ReadOnly:    false,
				},
				{
					Source:      ayosync.SandboxDir() + "/workspaces",
					Destination: "/workspaces",
					Mode:        providers.MountModeBind,
					ReadOnly:    false,
				},
				{
					Source:      paths.RuntimeDir(),
					Destination: "/run/ayo",
					Mode:        providers.MountModeBind,
					ReadOnly:    false,
				},
			},
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

	// Create session manager
	sessionManager := NewDaemonSessionManager(cfg.IdleTimeout)

	// Create ticket service
	ticketService := tickets.NewService(paths.SessionsDir())

	server := &Server{
		config:         cfg.Config,
		services:       cfg.Services,
		provider:       provider,
		pool:           pool,
		sessionManager: sessionManager,
		tickets:        ticketService,
		shutdownCh:     make(chan struct{}),
	}

	// Create trigger engine with callback to spawn sessions
	server.triggerEngine = NewTriggerEngine(TriggerEngineConfig{
		Callback: server.handleTriggerEvent,
	})

	// Create agent runner for ticket-based spawning
	agentRunner := NewDaemonAgentRunner(DaemonAgentRunnerConfig{
		Config:          cfg.Config,
		Services:        cfg.Services,
		SandboxProvider: provider,
		TicketService:   ticketService,
		MaxConcurrent:   cfg.MaxConcurrent,
	})

	// Create ticket watcher with runner
	ticketWatcher, err := NewTicketWatcher(TicketWatcherConfig{
		Runner:         agentRunner,
		OnTicketClosed: server.handleTicketClosed,
	})
	if err != nil {
		return nil, fmt.Errorf("create ticket watcher: %w", err)
	}
	server.ticketWatcher = ticketWatcher

	// Create squad service and RPC handler
	// Note: Using AppleProvider if available, otherwise nil (squad features disabled)
	var squadService *squads.Service
	if appleProvider, ok := provider.(*sandbox.AppleProvider); ok {
		squadService = squads.NewService(appleProvider)
	}
	squadTicketService := tickets.NewSquadTicketService()

	// Create planner manager for squad-specific planners
	plannerManager := planners.NewSandboxPlannerManager(nil, cfg.Config)

	// Create squad agent invoker for dispatching prompts to agents
	squadInvoker := NewSquadAgentInvoker(SquadAgentInvokerConfig{
		Config:          cfg.Config,
		Services:        cfg.Services,
		SandboxProvider: provider,
		PlannerManager:  plannerManager,
	})

	server.squadTickets = squadTicketService
	server.squadRPC = NewSquadRPC(squadService, squadTicketService, ticketWatcher, squadInvoker)

	return server, nil
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

	// Ensure runtime directory exists (for sandbox mounts)
	if err := os.MkdirAll(paths.RuntimeDir(), 0755); err != nil {
		return fmt.Errorf("create runtime dir: %w", err)
	}

	// Ensure all mount source directories exist before creating sandboxes
	// This is needed for the pool's default mounts to work
	for _, mount := range s.pool.Config().Mounts {
		if err := os.MkdirAll(mount.Source, 0755); err != nil {
			return fmt.Errorf("create mount source dir %s: %w", mount.Source, err)
		}
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

	// Pre-warm @ayo sandbox so first invocation is fast
	if appleProvider, ok := s.provider.(*sandbox.AppleProvider); ok {
		go func() {
			if _, err := sandbox.EnsureAyoSandbox(ctx, appleProvider); err != nil {
				// Log but don't fail - @ayo sandbox is optional enhancement
			}
		}()
	}

	// Start session manager
	s.sessionManager.Start()

	// Start trigger engine
	if err := s.triggerEngine.Start(ctx); err != nil {
		s.listener.Close()
		return fmt.Errorf("start trigger engine: %w", err)
	}

	// Start ticket watcher (optional - don't fail if fsnotify unavailable)
	if s.ticketWatcher != nil {
		if err := s.ticketWatcher.Start(ctx); err != nil {
			// Log warning but don't fail - ticket watching is optional
		}
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

	// Stop session manager
	if s.sessionManager != nil {
		s.sessionManager.Stop()
	}

	// Stop trigger engine
	if s.triggerEngine != nil {
		s.triggerEngine.Stop(ctx)
	}

	// Stop ticket watcher
	if s.ticketWatcher != nil {
		s.ticketWatcher.Stop(ctx)
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

// TriggerEngine returns the trigger engine.
func (s *Server) TriggerEngine() *TriggerEngine {
	return s.triggerEngine
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
	case MethodSandboxJoin:
		return s.handleSandboxJoin(ctx, req)
	case MethodSandboxAgents:
		return s.handleSandboxAgents(req)
	case MethodSessionList:
		return s.handleSessionList(req)
	case MethodSessionStart:
		return s.handleSessionStart(req)
	case MethodSessionStop:
		return s.handleSessionStop(req)
	case MethodAgentWake:
		return s.handleAgentWake(req)
	case MethodAgentSleep:
		return s.handleAgentSleep(req)
	case MethodAgentStatus:
		return s.handleAgentStatus(req)
	case MethodTriggerList:
		return s.handleTriggerList(req)
	case MethodTriggerGet:
		return s.handleTriggerGet(req)
	case MethodTriggerRegister:
		return s.handleTriggerRegister(req)
	case MethodTriggerRemove:
		return s.handleTriggerRemove(req)
	case MethodTriggerTest:
		return s.handleTriggerTest(req)
	case MethodTriggerSetEnabled:
		return s.handleTriggerSetEnabled(req)
	// Flow methods
	case MethodFlowRun:
		return s.handleFlowRun(ctx, req)
	case MethodFlowList:
		return s.handleFlowList(req)
	case MethodFlowGet:
		return s.handleFlowGet(req)
	case MethodFlowHistory:
		return s.handleFlowHistory(req)
	// Ticket methods
	case MethodTicketCreate:
		return s.handleTicketCreate(req)
	case MethodTicketGet:
		return s.handleTicketGet(req)
	case MethodTicketList:
		return s.handleTicketList(req)
	case MethodTicketUpdate:
		return s.handleTicketUpdate(req)
	case MethodTicketDelete:
		return s.handleTicketDelete(req)
	case MethodTicketStart:
		return s.handleTicketStart(req)
	case MethodTicketClose:
		return s.handleTicketClose(req)
	case MethodTicketReopen:
		return s.handleTicketReopen(req)
	case MethodTicketBlock:
		return s.handleTicketBlock(req)
	case MethodTicketAssign:
		return s.handleTicketAssign(req)
	case MethodTicketAddNote:
		return s.handleTicketAddNote(req)
	case MethodTicketReady:
		return s.handleTicketReady(req)
	case MethodTicketBlocked:
		return s.handleTicketBlocked(req)
	case MethodTicketAddDep:
		return s.handleTicketAddDep(req)
	case MethodTicketRemDep:
		return s.handleTicketRemoveDep(req)
	// Squad methods
	case MethodSquadCreate:
		return s.handleSquadCreate(ctx, req)
	case MethodSquadDestroy:
		return s.handleSquadDestroy(ctx, req)
	case MethodSquadList:
		return s.handleSquadList(ctx, req)
	case MethodSquadGet:
		return s.handleSquadGet(ctx, req)
	case MethodSquadStart:
		return s.handleSquadStart(ctx, req)
	case MethodSquadStop:
		return s.handleSquadStop(ctx, req)
	case MethodSquadAddAgent:
		return s.handleSquadAddAgent(ctx, req)
	case MethodSquadRemoveAgent:
		return s.handleSquadRemoveAgent(ctx, req)
	case MethodSquadTicketsReady:
		return s.handleSquadTicketsReady(ctx, req)
	case MethodSquadNotifyAgents:
		return s.handleSquadNotifyAgents(ctx, req)
	case MethodSquadWaitCompletion:
		return s.handleSquadWaitCompletion(ctx, req)
	case MethodSquadSyncOutput:
		return s.handleSquadSyncOutput(ctx, req)
	case MethodSquadCleanup:
		return s.handleSquadCleanup(ctx, req)
	case MethodSquadDispatch:
		return s.handleSquadDispatch(ctx, req)
	// Agent invocation
	case MethodAgentInvoke:
		return s.handleAgentInvoke(ctx, req)
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
		os.Exit(0) // Exit the process after graceful shutdown
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

func (s *Server) handleSandboxJoin(ctx context.Context, req *Request) *Response {
	var params SandboxJoinParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	_, err := s.pool.AcquireWithOptions(ctx, sandbox.AcquireOptions{
		Agent:       params.Agent,
		JoinSandbox: params.SandboxID,
	})
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Create user account for the agent in the sandbox
	_ = s.provider.EnsureAgentUser(ctx, params.SandboxID, params.Agent, "")

	resp, _ := NewResponse(struct{}{}, req.ID)
	return resp
}

func (s *Server) handleSandboxAgents(req *Request) *Response {
	var params SandboxAgentsParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	agents := s.pool.GetSandboxAgents(params.SandboxID)
	result := SandboxAgentsResult{Agents: agents}

	resp, _ := NewResponse(result, req.ID)
	return resp
}

// Session management handlers

func (s *Server) handleSessionList(req *Request) *Response {
	sessions := s.sessionManager.List()
	result := SessionListResult{Sessions: sessions}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSessionStart(req *Request) *Response {
	var params SessionStartParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	session := s.sessionManager.Wake(params.AgentHandle, params.TriggerID, params.SessionID)
	result := SessionStartResult{Session: session}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSessionStop(req *Request) *Response {
	var params SessionStopParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if err := s.sessionManager.StopSession(params.SessionID); err != nil {
		if rpcErr, ok := err.(*Error); ok {
			return NewErrorResponse(rpcErr, req.ID)
		}
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	resp, _ := NewResponse(struct{}{}, req.ID)
	return resp
}

func (s *Server) handleAgentWake(req *Request) *Response {
	var params AgentWakeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	session := s.sessionManager.Wake(params.Handle, params.TriggerID, params.SessionID)
	result := AgentWakeResult{Session: session}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleAgentSleep(req *Request) *Response {
	var params AgentSleepParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if err := s.sessionManager.Sleep(params.Handle); err != nil {
		if rpcErr, ok := err.(*Error); ok {
			return NewErrorResponse(rpcErr, req.ID)
		}
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	resp, _ := NewResponse(struct{}{}, req.ID)
	return resp
}

func (s *Server) handleAgentStatus(req *Request) *Response {
	var params AgentStatusParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	sess, err := s.sessionManager.GetByAgent(params.Handle)
	if err != nil {
		// Not an error - just means no active session
		result := AgentStatusResult{
			Active: false,
			Handle: params.Handle,
		}
		resp, _ := NewResponse(result, req.ID)
		return resp
	}

	result := AgentStatusResult{
		Active:     true,
		Handle:     sess.AgentHandle,
		Session:    sess,
		StartedAt:  sess.StartedAt,
		LastActive: sess.LastActive,
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) writePIDFile() error {
	pidPath := DefaultPIDPath()
	return os.WriteFile(pidPath, fmt.Appendf(nil, "%d", os.Getpid()), 0644)
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

// handleTriggerEvent handles a trigger event by waking the agent and injecting context.
func (s *Server) handleTriggerEvent(event TriggerEvent) {
	// Wake the agent with trigger context
	s.sessionManager.Wake(event.Agent, event.TriggerID, "")

	// TODO: In the future, we could inject the prompt and context into the agent session
	// For now, just log that the trigger fired
}

// handleTicketClosed handles a ticket being closed.
// This is called by the TicketWatcher when a ticket transitions to closed status.
func (s *Server) handleTicketClosed(contextID, ticketID string, isSquad bool) {
	// Log the closure for now
	// Future: could notify other agents, update metrics, etc.
}

// Trigger management handlers

func (s *Server) handleTriggerList(req *Request) *Response {
	triggers := s.triggerEngine.List()
	infos := make([]TriggerInfo, len(triggers))
	for i, t := range triggers {
		infos[i] = triggerToInfo(t)
	}

	result := TriggerListResult{Triggers: infos}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTriggerGet(req *Request) *Response {
	var params TriggerGetParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	trigger, err := s.triggerEngine.Get(params.ID)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	result := TriggerGetResult{Trigger: triggerToInfo(trigger)}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTriggerRegister(req *Request) *Response {
	var params TriggerRegisterParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	// Validate type
	if params.Type != "cron" && params.Type != "watch" && params.Type != "webhook" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "type must be 'cron', 'watch', or 'webhook'"), req.ID)
	}

	// Validate agent
	if params.Agent == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "agent is required"), req.ID)
	}

	// Create trigger
	trigger := &Trigger{
		ID:      GenerateTriggerID(),
		Type:    TriggerType(params.Type),
		Agent:   params.Agent,
		Prompt:  params.Prompt,
		Source:  "cli",
		Enabled: true,
		Config: TriggerConfig{
			Schedule:  params.Schedule,
			Path:      params.Path,
			Patterns:  params.Patterns,
			Recursive: params.Recursive,
			Events:    params.Events,
		},
	}

	if err := s.triggerEngine.Register(trigger); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	result := TriggerRegisterResult{Trigger: triggerToInfo(trigger)}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTriggerRemove(req *Request) *Response {
	var params TriggerRemoveParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if err := s.triggerEngine.Unregister(params.ID); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	resp, _ := NewResponse(struct{}{}, req.ID)
	return resp
}

func (s *Server) handleTriggerTest(req *Request) *Response {
	var params TriggerTestParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	trigger, err := s.triggerEngine.Get(params.ID)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Fire the trigger manually
	event := TriggerEvent{
		TriggerID: trigger.ID,
		FiredAt:   time.Now(),
		Context:   map[string]any{"test": true},
		Agent:     trigger.Agent,
		Prompt:    trigger.Prompt,
	}

	if s.triggerEngine.callback != nil {
		go s.triggerEngine.callback(event)
	}

	resp, _ := NewResponse(struct{}{}, req.ID)
	return resp
}

func (s *Server) handleTriggerSetEnabled(req *Request) *Response {
	var params TriggerSetEnabledParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if err := s.triggerEngine.SetEnabled(params.ID, params.Enabled); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	resp, _ := NewResponse(struct{}{}, req.ID)
	return resp
}

// triggerToInfo converts an internal Trigger to a TriggerInfo for RPC.
func triggerToInfo(t *Trigger) TriggerInfo {
	return TriggerInfo{
		ID:            t.ID,
		Type:          string(t.Type),
		Agent:         t.Agent,
		Prompt:        t.Prompt,
		Source:        t.Source,
		Enabled:   t.Enabled,
		Schedule:  t.Config.Schedule,
		Path:      t.Config.Path,
		Patterns:  t.Config.Patterns,
		Recursive: t.Config.Recursive,
		Events:    t.Config.Events,
	}
}

// Squad RPC handlers - delegate to squadRPC

func (s *Server) handleSquadCreate(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadCreate(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadDestroy(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadDestroy(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadList(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadList(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadGet(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadGet(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadStart(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadStart(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadStop(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadStop(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadAddAgent(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadAddAgent(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadRemoveAgent(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadRemoveAgent(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadTicketsReady(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadTicketsReady(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadNotifyAgents(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadNotifyAgents(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadWaitCompletion(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadWaitCompletion(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadSyncOutput(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadSyncOutput(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadCleanup(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadCleanup(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleSquadDispatch(ctx context.Context, req *Request) *Response {
	result, err := s.squadRPC.HandleSquadDispatch(ctx, req.Params)
	if err != nil {
		return NewErrorResponse(err, req.ID)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}
