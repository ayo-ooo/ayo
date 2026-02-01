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
	s.mux.HandleFunc("OPTIONS /sessions", s.handleCORS)

	// Agent endpoints
	s.mux.HandleFunc("GET /agents", s.withMiddleware(s.handleListAgents))
	s.mux.HandleFunc("GET /agents/{handle}", s.withMiddleware(s.handleGetAgent))

	// Chat endpoints
	s.mux.HandleFunc("POST /agents/{handle}/chat", s.withMiddleware(s.handleChat))
	s.mux.HandleFunc("POST /agents/{handle}/sessions/{id}", s.withMiddleware(s.handleContinueSession))
	s.mux.HandleFunc("GET /agents/{handle}/sessions/{id}", s.withMiddleware(s.handleGetSession))
	s.mux.HandleFunc("DELETE /agents/{handle}/sessions/{id}", s.withMiddleware(s.handleDeleteSession))

	// Session list
	s.mux.HandleFunc("GET /sessions", s.withMiddleware(s.handleListSessions))

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
