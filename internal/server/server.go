// Package server provides the HTTP API for ayo agents.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/session"
)

// Server is the HTTP server for the ayo API.
type Server struct {
	config   config.Config
	services *session.Services
	auth     *Auth
	logger   *slog.Logger

	// Handlers
	agentHandler *AgentHandler
	chatHandler  *ChatHandler

	// Managers
	sessionManager *SessionManager
	sandboxManager *SandboxManager

	httpServer *http.Server
	mux        *http.ServeMux

	mu       sync.RWMutex
	started  bool
	addr     string
	shutdown chan struct{}

	onReady func(addr string)
}

// Options configures the server.
type Options struct {
	// Addr is the address to listen on (default: "127.0.0.1:0" for random port).
	Addr string

	// Services provides session persistence.
	Services *session.Services

	// Logger for request logging. If nil, uses slog.Default().
	Logger *slog.Logger

	// AllowRemote allows connections from non-localhost addresses.
	// When false (default), only localhost connections are allowed without auth.
	AllowRemote bool

	// OnReady is called after the server starts listening, with the actual address.
	// This is useful for displaying connection info when using a random port.
	OnReady func(addr string)

	// IdleTimeout is how long before idle agent sessions auto-stop.
	// Zero means no auto-stop.
	IdleTimeout time.Duration

	// EnableSandbox enables the sandbox manager.
	EnableSandbox bool
}

// New creates a new server instance.
func New(cfg config.Config, opts Options) *Server {
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:0"
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	s := &Server{
		config:       cfg,
		services:     opts.Services,
		auth:         NewAuth(!opts.AllowRemote),
		logger:       opts.Logger,
		mux:          http.NewServeMux(),
		shutdown:     make(chan struct{}),
		addr:         opts.Addr,
		agentHandler: NewAgentHandler(cfg),
		chatHandler:  NewChatHandler(cfg, opts.Services),
		onReady:      opts.OnReady,
	}

	// Create session manager
	s.sessionManager = NewSessionManager(SessionManagerConfig{
		Config:      cfg,
		Services:    opts.Services,
		IdleTimeout: opts.IdleTimeout,
		Logger:      opts.Logger,
	})

	// Optionally create sandbox manager
	if opts.EnableSandbox {
		s.sandboxManager = NewSandboxManager(SandboxManagerConfig{
			KeepOnStop: true,
			Logger:     opts.Logger,
		})
	}

	s.registerRoutes()
	return s
}

// registerRoutes sets up all HTTP routes.
func (s *Server) registerRoutes() {
	// Web client at root
	s.mux.HandleFunc("GET /", s.handleWebClient)

	// Health check (with CORS for cross-origin requests)
	s.mux.HandleFunc("GET /health", s.withCORS(s.handleHealth))
	s.mux.HandleFunc("OPTIONS /health", s.handleCORS)

	// CORS preflight for all paths
	s.mux.HandleFunc("OPTIONS /", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /connect", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /agents", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /agents/{handle}", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /agents/{handle}/chat", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /agents/{handle}/sessions/{id}", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /agents/{handle}/wake", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /agents/{handle}/sleep", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /agents/{handle}/status", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /sessions", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /daemon/sessions", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /daemon/sessions/{id}", s.handleCORS)
	s.mux.HandleFunc("OPTIONS /daemon/sandbox", s.handleCORS)

	// Agent endpoints
	s.mux.HandleFunc("GET /agents", s.withMiddleware(s.handleListAgents))
	s.mux.HandleFunc("GET /agents/{handle}", s.withMiddleware(s.handleGetAgent))

	// Agent lifecycle (wake/sleep)
	s.mux.HandleFunc("POST /agents/{handle}/wake", s.withMiddleware(s.handleAgentWake))
	s.mux.HandleFunc("POST /agents/{handle}/sleep", s.withMiddleware(s.handleAgentSleep))
	s.mux.HandleFunc("GET /agents/{handle}/status", s.withMiddleware(s.handleAgentStatus))

	// Chat endpoints
	s.mux.HandleFunc("POST /agents/{handle}/chat", s.withMiddleware(s.handleChat))
	s.mux.HandleFunc("POST /agents/{handle}/sessions/{id}", s.withMiddleware(s.handleContinueSession))
	s.mux.HandleFunc("GET /agents/{handle}/sessions/{id}", s.withMiddleware(s.handleGetSession))
	s.mux.HandleFunc("DELETE /agents/{handle}/sessions/{id}", s.withMiddleware(s.handleDeleteSession))

	// Session list (persisted sessions)
	s.mux.HandleFunc("GET /sessions", s.withMiddleware(s.handleListSessions))

	// Daemon-managed sessions
	s.mux.HandleFunc("GET /daemon/sessions", s.withMiddleware(s.handleListDaemonSessions))
	s.mux.HandleFunc("GET /daemon/sessions/{id}", s.withMiddleware(s.handleGetDaemonSession))
	s.mux.HandleFunc("DELETE /daemon/sessions/{id}", s.withMiddleware(s.handleStopDaemonSession))
	s.mux.HandleFunc("POST /daemon/sessions/{id}/inject", s.withMiddleware(s.handleInjectMessage))

	// Sandbox status
	s.mux.HandleFunc("GET /daemon/sandbox", s.withMiddleware(s.handleSandboxStatus))

	// Connection info (for QR code generation)
	s.mux.HandleFunc("GET /connect", s.withCORS(s.handleConnect))
}

// withMiddleware wraps a handler with logging, auth, and CORS.
func (s *Server) withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Auth check
		if !s.auth.Validate(r) {
			s.logger.Warn("unauthorized request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Call handler
		next(w, r)

		// Log request
		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).Round(time.Millisecond),
		)
	}
}

// withCORS wraps a handler with CORS headers only (no auth).
func (s *Server) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		next(w, r)
	}
}

// Start begins listening for requests.
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return errors.New("server already started")
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.addr = listener.Addr().String()
	s.started = true

	s.httpServer = &http.Server{
		Handler:      s.mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0, // Disabled for SSE streaming
		IdleTimeout:  120 * time.Second,
	}

	s.mu.Unlock()

	// Start managers
	if err := s.sessionManager.Start(ctx); err != nil {
		return fmt.Errorf("start session manager: %w", err)
	}

	if s.sandboxManager != nil {
		if err := s.sandboxManager.Start(ctx); err != nil {
			return fmt.Errorf("start sandbox manager: %w", err)
		}
	}

	// Call OnReady callback with actual address
	if s.onReady != nil {
		s.onReady(s.addr)
	}

	s.logger.Info("server starting", "addr", s.addr)

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	err = s.httpServer.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	close(s.shutdown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop managers
	if s.sessionManager != nil {
		s.sessionManager.Stop(ctx)
	}

	if s.sandboxManager != nil {
		s.sandboxManager.Stop(ctx)
	}

	s.started = false
	return s.httpServer.Shutdown(ctx)
}

// Addr returns the address the server is listening on.
// Only valid after Start() is called.
func (s *Server) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.addr
}

// Token returns the authentication token for this server instance.
func (s *Server) Token() string {
	return s.auth.Token()
}

// SessionManager returns the session manager.
func (s *Server) SessionManager() *SessionManager {
	return s.sessionManager
}

// SandboxManager returns the sandbox manager (may be nil if not enabled).
func (s *Server) SandboxManager() *SandboxManager {
	return s.sandboxManager
}

// handleCORS handles CORS preflight requests.
func (s *Server) handleCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.WriteHeader(http.StatusOK)
}

// handleHealth returns server health status.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleConnect returns connection info for clients.
func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"version": 1,
		"token":   s.auth.Token(),
	})
}

// Placeholder handlers - will be implemented in subsequent tickets

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	s.agentHandler.ListAgents(w, r)
}

func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	s.agentHandler.GetAgent(w, r)
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	s.chatHandler.Chat(w, r)
}

func (s *Server) handleContinueSession(w http.ResponseWriter, r *http.Request) {
	s.chatHandler.ContinueSession(w, r)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	s.chatHandler.GetSession(w, r)
}

func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	s.chatHandler.DeleteSession(w, r)
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	s.chatHandler.ListSessions(w, r)
}

// Agent lifecycle handlers

// handleAgentWake wakes up (starts a session for) an agent.
func (s *Server) handleAgentWake(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	if handle == "" {
		http.Error(w, "handle required", http.StatusBadRequest)
		return
	}

	sess, err := s.sessionManager.Wake(r.Context(), handle, WakeOptions{})
	if err != nil {
		http.Error(w, "failed to wake agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}

// handleAgentSleep puts an agent to sleep (stops its session).
func (s *Server) handleAgentSleep(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	if handle == "" {
		http.Error(w, "handle required", http.StatusBadRequest)
		return
	}

	if err := s.sessionManager.Sleep(r.Context(), handle); err != nil {
		http.Error(w, "failed to sleep agent: "+err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleAgentStatus returns the status of an agent's session.
func (s *Server) handleAgentStatus(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	if handle == "" {
		http.Error(w, "handle required", http.StatusBadRequest)
		return
	}

	sess, err := s.sessionManager.GetByAgent(handle)
	if err != nil {
		// Return empty status if no session
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"active": false,
			"handle": handle,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"active":     true,
		"handle":     sess.AgentHandle,
		"session_id": sess.ID,
		"status":     sess.Status,
		"started_at": sess.StartedAt,
		"last_active": sess.LastActive,
	})
}

// Daemon session handlers

// handleListDaemonSessions lists all active agent sessions managed by the daemon.
func (s *Server) handleListDaemonSessions(w http.ResponseWriter, r *http.Request) {
	sessions := s.sessionManager.List()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// handleGetDaemonSession returns details of a specific daemon session.
func (s *Server) handleGetDaemonSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	sess, err := s.sessionManager.Get(id)
	if err != nil {
		http.Error(w, "session not found: "+id, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}

// handleStopDaemonSession stops a specific daemon session.
func (s *Server) handleStopDaemonSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	if err := s.sessionManager.StopSession(r.Context(), id); err != nil {
		http.Error(w, "failed to stop session: "+err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// InjectRequest is the request body for injecting a message into a session.
type InjectRequest struct {
	Message string `json:"message"`
}

// handleInjectMessage injects a message into an active session.
func (s *Server) handleInjectMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	var req InjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "message required", http.StatusBadRequest)
		return
	}

	response, err := s.sessionManager.Inject(r.Context(), id, req.Message)
	if err != nil {
		http.Error(w, "failed to inject message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"response": response,
	})
}

// Sandbox status handler

// handleSandboxStatus returns the status of the persistent sandbox.
func (s *Server) handleSandboxStatus(w http.ResponseWriter, r *http.Request) {
	if s.sandboxManager == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"enabled": false,
		})
		return
	}

	status := s.sandboxManager.GetStatus(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"enabled":  true,
		"id":       status.ID,
		"name":     status.Name,
		"running":  status.Running,
		"status":   status.Status,
		"provider": status.Provider,
		"healthy":  status.Healthy,
	})
}
