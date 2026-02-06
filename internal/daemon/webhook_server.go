package daemon

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

// WebhookServer provides an HTTP server for receiving webhooks.
type WebhookServer struct {
	mu            sync.RWMutex
	server        *http.Server
	listener      net.Listener
	triggers      map[string]*WebhookTrigger // path -> trigger
	callback      TriggerCallback
	logger        *slog.Logger
	bindAddr      string
	secret        string // optional HMAC secret
	running       bool
}

// WebhookTrigger represents a webhook trigger configuration.
type WebhookTrigger struct {
	ID      string `json:"id"`
	Path    string `json:"path"` // e.g., /hooks/my-trigger
	Agent   string `json:"agent"`
	Prompt  string `json:"prompt,omitempty"`
	Secret  string `json:"secret,omitempty"` // per-trigger secret
	Format  string `json:"format,omitempty"` // github, gitlab, generic
}

// WebhookServerConfig configures the webhook server.
type WebhookServerConfig struct {
	BindAddr string // default: "127.0.0.1:0" (random port, localhost only)
	Secret   string // global HMAC secret
	Logger   *slog.Logger
	Callback TriggerCallback
}

// NewWebhookServer creates a new webhook server.
func NewWebhookServer(cfg WebhookServerConfig) *WebhookServer {
	if cfg.BindAddr == "" {
		cfg.BindAddr = "127.0.0.1:0"
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &WebhookServer{
		triggers: make(map[string]*WebhookTrigger),
		callback: cfg.Callback,
		logger:   cfg.Logger,
		bindAddr: cfg.BindAddr,
		secret:   cfg.Secret,
	}
}

// Start starts the webhook server.
func (s *WebhookServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("webhook server already running")
	}
	s.running = true
	s.mu.Unlock()

	// Create listener
	var err error
	s.listener, err = net.Listen("tcp", s.bindAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/hooks/", s.handleHook)
	mux.HandleFunc("/hooks/github", s.handleGitHub)
	mux.HandleFunc("/hooks/gitlab", s.handleGitLab)
	mux.HandleFunc("/hooks/generic", s.handleGeneric)
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start serving
	go func() {
		if err := s.server.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			s.logger.Error("webhook server error", "error", err)
		}
	}()

	s.logger.Info("webhook server started", "addr", s.listener.Addr().String())
	return nil
}

// Stop stops the webhook server.
func (s *WebhookServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
	}

	s.logger.Info("webhook server stopped")
	return nil
}

// Addr returns the server's address.
func (s *WebhookServer) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// Port returns the server's port.
func (s *WebhookServer) Port() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.listener == nil {
		return 0
	}
	if tcpAddr, ok := s.listener.Addr().(*net.TCPAddr); ok {
		return tcpAddr.Port
	}
	return 0
}

// Register registers a webhook trigger.
func (s *WebhookServer) Register(trigger *WebhookTrigger) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if trigger.Path == "" {
		return fmt.Errorf("path is required")
	}

	// Normalize path
	if !strings.HasPrefix(trigger.Path, "/") {
		trigger.Path = "/" + trigger.Path
	}

	s.triggers[trigger.Path] = trigger
	s.logger.Info("registered webhook trigger", "id", trigger.ID, "path", trigger.Path, "agent", trigger.Agent)
	return nil
}

// Unregister removes a webhook trigger.
func (s *WebhookServer) Unregister(triggerPath string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Normalize path
	if !strings.HasPrefix(triggerPath, "/") {
		triggerPath = "/" + triggerPath
	}

	delete(s.triggers, triggerPath)
	s.logger.Info("unregistered webhook trigger", "path", triggerPath)
}

// handleRoot returns a list of available endpoints.
func (s *WebhookServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"endpoints": []string{
			"/health",
			"/hooks/{trigger-path}",
			"/hooks/github",
			"/hooks/gitlab",
			"/hooks/generic",
		},
	})
}

// handleHealth returns health status.
func (s *WebhookServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// handleHook handles custom webhook paths.
func (s *WebhookServer) handleHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract trigger path from URL
	triggerPath := r.URL.Path

	// Find trigger
	s.mu.RLock()
	trigger, ok := s.triggers[triggerPath]
	s.mu.RUnlock()

	if !ok {
		// Check if it's a pattern match (e.g., /hooks/my-trigger)
		s.mu.RLock()
		for p, t := range s.triggers {
			if matchPath(p, triggerPath) {
				trigger = t
				ok = true
				break
			}
		}
		s.mu.RUnlock()
	}

	if !ok {
		http.Error(w, "Trigger not found", http.StatusNotFound)
		return
	}

	// Validate signature if secret is set
	if trigger.Secret != "" || s.secret != "" {
		secret := trigger.Secret
		if secret == "" {
			secret = s.secret
		}
		if !s.validateSignature(r, secret) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Read body
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB limit
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Parse payload
	var payload map[string]any
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			// Not JSON, store as raw
			payload = map[string]any{"raw": string(body)}
		}
	}

	// Fire trigger
	s.fireTrigger(trigger, payload, "custom")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "accepted",
		"trigger":   trigger.ID,
	})
}

// handleGitHub handles GitHub webhook payloads.
func (s *WebhookServer) handleGitHub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate GitHub signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature != "" && s.secret != "" {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		if !s.validateGitHubSignature(body, signature, s.secret) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		// Parse GitHub payload
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Extract GitHub event
		event := r.Header.Get("X-GitHub-Event")
		delivery := r.Header.Get("X-GitHub-Delivery")

		// Find matching trigger
		s.mu.RLock()
		var trigger *WebhookTrigger
		for _, t := range s.triggers {
			if t.Format == "github" || t.Path == "/hooks/github" {
				trigger = t
				break
			}
		}
		s.mu.RUnlock()

		if trigger != nil {
			s.fireTrigger(trigger, map[string]any{
				"event":    event,
				"delivery": delivery,
				"payload":  payload,
			}, "github")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "accepted",
			"event":  event,
		})
		return
	}

	http.Error(w, "Missing signature", http.StatusBadRequest)
}

// handleGitLab handles GitLab webhook payloads.
func (s *WebhookServer) handleGitLab(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// GitLab uses X-Gitlab-Token header
	token := r.Header.Get("X-Gitlab-Token")
	if s.secret != "" && token != s.secret {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Read body
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Parse GitLab payload
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Extract GitLab event
	event := r.Header.Get("X-Gitlab-Event")
	objectKind := ""
	if ok, _ := payload["object_kind"].(string); ok != "" {
		objectKind = ok
	}

	// Find matching trigger
	s.mu.RLock()
	var trigger *WebhookTrigger
	for _, t := range s.triggers {
		if t.Format == "gitlab" || t.Path == "/hooks/gitlab" {
			trigger = t
			break
		}
	}
	s.mu.RUnlock()

	if trigger != nil {
		s.fireTrigger(trigger, map[string]any{
			"event":       event,
			"object_kind": objectKind,
			"payload":     payload,
		}, "gitlab")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "accepted",
		"event":  event,
	})
}

// handleGeneric handles generic JSON webhook payloads.
func (s *WebhookServer) handleGeneric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Parse payload
	var payload map[string]any
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			payload = map[string]any{"raw": string(body)}
		}
	}

	// Find matching trigger
	agentHandle := r.URL.Query().Get("agent")
	s.mu.RLock()
	var trigger *WebhookTrigger
	for _, t := range s.triggers {
		if t.Format == "generic" || t.Path == "/hooks/generic" {
			if agentHandle == "" || t.Agent == agentHandle {
				trigger = t
				break
			}
		}
	}
	s.mu.RUnlock()

	if trigger != nil {
		s.fireTrigger(trigger, payload, "generic")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "accepted",
	})
}

// fireTrigger fires a webhook trigger.
func (s *WebhookServer) fireTrigger(trigger *WebhookTrigger, payload map[string]any, source string) {
	event := TriggerEvent{
		TriggerID: trigger.ID,
		FiredAt:   time.Now(),
		Context: map[string]any{
			"source":  source,
			"payload": payload,
		},
		Agent:  trigger.Agent,
		Prompt: trigger.Prompt,
	}

	s.logger.Info("webhook trigger fired", "id", trigger.ID, "agent", trigger.Agent, "source", source)

	if s.callback != nil {
		go s.callback(event)
	}
}

// validateSignature validates a simple HMAC signature.
func (s *WebhookServer) validateSignature(r *http.Request, secret string) bool {
	signature := r.Header.Get("X-Signature")
	if signature == "" {
		return false
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

// validateGitHubSignature validates a GitHub webhook signature.
func (s *WebhookServer) validateGitHubSignature(body []byte, signature, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

// matchPath checks if a pattern matches a path.
func matchPath(pattern, urlPath string) bool {
	// Simple prefix matching for now
	matched, _ := path.Match(pattern, urlPath)
	return matched
}
