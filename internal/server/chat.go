package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/smallmodel"
)

// ChatRequest represents a chat API request.
type ChatRequest struct {
	Message string `json:"message"`
	Model   string `json:"model,omitempty"` // Optional model override
}

// ChatHandler handles chat-related API requests.
type ChatHandler struct {
	config           config.Config
	services         *session.Services
	memoryService    *memory.Service
	formationService *memory.FormationService
	smallModel       *smallmodel.Service

	// Runners keyed by session ID for session continuity
	mu      sync.RWMutex
	runners map[string]*run.Runner
}

// NewChatHandler creates a new chat handler.
func NewChatHandler(cfg config.Config, services *session.Services) *ChatHandler {
	return &ChatHandler{
		config:   cfg,
		services: services,
		runners:  make(map[string]*run.Runner),
	}
}

// SetMemoryService sets the memory service for the chat handler.
func (h *ChatHandler) SetMemoryService(svc *memory.Service) {
	h.memoryService = svc
}

// SetFormationService sets the formation service for the chat handler.
func (h *ChatHandler) SetFormationService(svc *memory.FormationService) {
	h.formationService = svc
}

// SetSmallModel sets the small model service for the chat handler.
func (h *ChatHandler) SetSmallModel(svc *smallmodel.Service) {
	h.smallModel = svc
}

// Chat handles a new chat session request.
func (h *ChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	if handle == "" {
		http.Error(w, "handle required", http.StatusBadRequest)
		return
	}

	// Normalize handle
	if handle[0] != '@' {
		handle = "@" + handle
	}

	// Parse request body
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		http.Error(w, "message required", http.StatusBadRequest)
		return
	}

	// Load agent
	ag, err := agent.Load(h.config, handle)
	if err != nil {
		http.Error(w, "agent not found: "+handle, http.StatusNotFound)
		return
	}

	// Apply model override if specified
	if req.Model != "" {
		ag.Model = req.Model
	}

	// Create SSE writer
	sse := NewSSEWriter(w)
	if sse == nil {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	defer sse.Close()

	// Create runner
	runner, err := run.NewRunner(h.config, false, run.RunnerOptions{
		Services:         h.services,
		MemoryService:    h.memoryService,
		FormationService: h.formationService,
		SmallModel:       h.smallModel,
		StreamWriter:     sse,
	})
	if err != nil {
		sse.WriteError(err)
		return
	}

	// Start heartbeat goroutine
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go h.sendHeartbeats(ctx, sse)

	// Execute chat
	response, err := runner.Chat(ctx, ag, req.Message)
	if err != nil {
		sse.WriteError(err)
		return
	}

	// Get session ID from runner for response
	sessionID := runner.GetSessionID(ag.Handle)

	// Send done event with session info
	sse.writeEvent("done", map[string]string{
		"response":   response,
		"session_id": sessionID,
	})
}

// ContinueSession handles continuing an existing chat session.
func (h *ChatHandler) ContinueSession(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	sessionID := r.PathValue("id")

	if handle == "" || sessionID == "" {
		http.Error(w, "handle and session id required", http.StatusBadRequest)
		return
	}

	// Normalize handle
	if handle[0] != '@' {
		handle = "@" + handle
	}

	// Parse request
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		http.Error(w, "message required", http.StatusBadRequest)
		return
	}

	// Load agent
	ag, err := agent.Load(h.config, handle)
	if err != nil {
		http.Error(w, "agent not found: "+handle, http.StatusNotFound)
		return
	}

	// Check if session exists
	if h.services != nil {
		_, err := h.services.Sessions.Get(r.Context(), sessionID)
		if err != nil {
			http.Error(w, "session not found: "+sessionID, http.StatusNotFound)
			return
		}
	}

	// Create SSE writer
	sse := NewSSEWriter(w)
	if sse == nil {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	defer sse.Close()

	// Get or create runner for this session
	runner := h.getOrCreateRunner(sessionID)
	runner.SetStreamWriter(sse)

	// Resume session if not already loaded
	if h.services != nil && runner.GetSessionID(ag.Handle) == "" {
		// Load messages from database
		messages, err := h.services.Messages.List(r.Context(), sessionID)
		if err != nil {
			sse.WriteError(err)
			return
		}
		if err := runner.ResumeSession(r.Context(), ag, sessionID, messages); err != nil {
			sse.WriteError(err)
			return
		}
	}

	// Start heartbeat
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go h.sendHeartbeats(ctx, sse)

	// Execute chat
	response, err := runner.Chat(ctx, ag, req.Message)
	if err != nil {
		sse.WriteError(err)
		return
	}

	sse.writeEvent("done", map[string]string{
		"response":   response,
		"session_id": sessionID,
	})
}

// GetSession returns session history.
func (h *ChatHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	if h.services == nil {
		http.Error(w, "session persistence not configured", http.StatusServiceUnavailable)
		return
	}

	// Get session
	sess, err := h.services.Sessions.Get(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "session not found: "+sessionID, http.StatusNotFound)
		return
	}

	// Get messages
	messages, err := h.services.Messages.List(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "failed to get messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to API format
	apiMessages := make([]MessageInfo, 0, len(messages))
	for _, msg := range messages {
		apiMessages = append(apiMessages, messageToInfo(msg))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SessionDetail{
		ID:           sess.ID,
		AgentHandle:  sess.AgentHandle,
		Title:        sess.Title,
		MessageCount: sess.MessageCount,
		CreatedAt:    sess.CreatedAt,
		UpdatedAt:    sess.UpdatedAt,
		Messages:     apiMessages,
	})
}

// DeleteSession deletes a session.
func (h *ChatHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	if h.services == nil {
		http.Error(w, "session persistence not configured", http.StatusServiceUnavailable)
		return
	}

	if err := h.services.Sessions.Delete(r.Context(), sessionID); err != nil {
		http.Error(w, "failed to delete session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Remove runner from cache
	h.mu.Lock()
	delete(h.runners, sessionID)
	h.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// ListSessions returns all sessions, optionally filtered by agent.
func (h *ChatHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	if h.services == nil {
		http.Error(w, "session persistence not configured", http.StatusServiceUnavailable)
		return
	}

	agentFilter := r.URL.Query().Get("agent")

	var sessions []session.Session
	var err error

	if agentFilter != "" {
		sessions, err = h.services.Sessions.ListByAgent(r.Context(), agentFilter, 100)
	} else {
		sessions, err = h.services.Sessions.List(r.Context(), 100)
	}

	if err != nil {
		http.Error(w, "failed to list sessions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	apiSessions := make([]SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		apiSessions = append(apiSessions, SessionInfo{
			ID:           sess.ID,
			AgentHandle:  sess.AgentHandle,
			Title:        sess.Title,
			MessageCount: sess.MessageCount,
			CreatedAt:    sess.CreatedAt,
			UpdatedAt:    sess.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiSessions)
}

// getOrCreateRunner gets or creates a runner for a session.
func (h *ChatHandler) getOrCreateRunner(sessionID string) *run.Runner {
	h.mu.Lock()
	defer h.mu.Unlock()

	if runner, ok := h.runners[sessionID]; ok {
		return runner
	}

	runner, _ := run.NewRunner(h.config, false, run.RunnerOptions{
		Services:         h.services,
		MemoryService:    h.memoryService,
		FormationService: h.formationService,
		SmallModel:       h.smallModel,
	})

	h.runners[sessionID] = runner
	return runner
}

// sendHeartbeats sends periodic heartbeat comments to keep the connection alive.
func (h *ChatHandler) sendHeartbeats(ctx context.Context, sse *SSEWriter) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sse.SendHeartbeat()
		}
	}
}

// SessionInfo represents a session summary for API responses.
type SessionInfo struct {
	ID           string `json:"id"`
	AgentHandle  string `json:"agent_handle"`
	Title        string `json:"title"`
	MessageCount int64  `json:"message_count"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// SessionDetail represents a full session with messages.
type SessionDetail struct {
	ID           string        `json:"id"`
	AgentHandle  string        `json:"agent_handle"`
	Title        string        `json:"title"`
	MessageCount int64         `json:"message_count"`
	CreatedAt    int64         `json:"created_at"`
	UpdatedAt    int64         `json:"updated_at"`
	Messages     []MessageInfo `json:"messages"`
}

// MessageInfo represents a message for API responses.
type MessageInfo struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

// messageToInfo converts a message to API format.
func messageToInfo(msg session.Message) MessageInfo {
	// Extract text content from parts
	content := ""
	for _, part := range msg.Parts {
		if text, ok := part.(session.TextContent); ok {
			content += text.Text
		}
	}

	return MessageInfo{
		ID:        msg.ID,
		Role:      string(msg.Role),
		Content:   content,
		CreatedAt: msg.CreatedAt,
	}
}
