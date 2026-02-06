package server

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/session"
)

// AgentSessionStatus represents the status of an agent session.
type AgentSessionStatus string

const (
	// SessionStatusRunning indicates the session is actively processing.
	SessionStatusRunning AgentSessionStatus = "running"

	// SessionStatusIdle indicates the session is waiting for input.
	SessionStatusIdle AgentSessionStatus = "idle"

	// SessionStatusStopped indicates the session has been stopped.
	SessionStatusStopped AgentSessionStatus = "stopped"
)

// AgentSession represents an active agent session tracked by the daemon.
type AgentSession struct {
	ID          string             `json:"id"`
	AgentHandle string             `json:"agent_handle"`
	StartedAt   time.Time          `json:"started_at"`
	TriggerID   string             `json:"trigger_id,omitempty"` // Set if started by trigger
	Status      AgentSessionStatus `json:"status"`
	LastActive  time.Time          `json:"last_active"`
	SessionID   string             `json:"session_id,omitempty"` // Database session ID if persisted
}

// SessionManagerConfig configures the session manager.
type SessionManagerConfig struct {
	// IdleTimeout is how long before an idle session auto-stops.
	// Zero means no auto-stop.
	IdleTimeout time.Duration

	// Logger for session manager operations.
	Logger *slog.Logger

	// Config for loading agents.
	Config config.Config

	// Services for session persistence.
	Services *session.Services
}

// SessionManager tracks active agent sessions spawned by the daemon.
type SessionManager struct {
	config   SessionManagerConfig
	logger   *slog.Logger
	sessions map[string]*managedSession
	mu       sync.RWMutex

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// managedSession wraps an agent session with its runner.
type managedSession struct {
	AgentSession
	runner *run.Runner
	agent  agent.Agent
	cancel context.CancelFunc
}

// NewSessionManager creates a new session manager.
func NewSessionManager(cfg SessionManagerConfig) *SessionManager {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &SessionManager{
		config:   cfg,
		logger:   cfg.Logger,
		sessions: make(map[string]*managedSession),
		stopCh:   make(chan struct{}),
	}
}

// Start begins the session manager, including idle timeout monitoring.
func (m *SessionManager) Start(ctx context.Context) error {
	m.logger.Info("session manager started")

	if m.config.IdleTimeout > 0 {
		m.wg.Add(1)
		go m.idleTimeoutLoop()
	}

	return nil
}

// Stop stops all managed sessions and the manager itself.
func (m *SessionManager) Stop(ctx context.Context) error {
	close(m.stopCh)
	m.wg.Wait()

	// Stop all sessions
	m.mu.Lock()
	for id, sess := range m.sessions {
		m.logger.Info("stopping session on shutdown", "session", id, "agent", sess.AgentHandle)
		if sess.cancel != nil {
			sess.cancel()
		}
	}
	m.sessions = make(map[string]*managedSession)
	m.mu.Unlock()

	m.logger.Info("session manager stopped")
	return nil
}

// WakeOptions configures how an agent is woken.
type WakeOptions struct {
	// TriggerID is set if woken by a trigger.
	TriggerID string

	// SessionID to resume (optional).
	SessionID string
}

// Wake starts or resumes a session for an agent.
// Returns the session ID if the agent was woken successfully.
func (m *SessionManager) Wake(ctx context.Context, handle string, opts WakeOptions) (*AgentSession, error) {
	// Normalize handle
	if len(handle) > 0 && handle[0] != '@' {
		handle = "@" + handle
	}

	// Check if already running
	m.mu.RLock()
	for _, sess := range m.sessions {
		if sess.AgentHandle == handle && sess.Status != SessionStatusStopped {
			m.mu.RUnlock()
			return &sess.AgentSession, nil
		}
	}
	m.mu.RUnlock()

	// Load agent
	ag, err := agent.Load(m.config.Config, handle)
	if err != nil {
		return nil, fmt.Errorf("load agent: %w", err)
	}

	// Create runner
	runner, err := run.NewRunner(m.config.Config, false, run.RunnerOptions{
		Services: m.config.Services,
	})
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	// Generate session ID
	id := generateSessionID()
	now := time.Now()

	// Create cancellable context
	sessionCtx, cancel := context.WithCancel(context.Background())

	sess := &managedSession{
		AgentSession: AgentSession{
			ID:          id,
			AgentHandle: handle,
			StartedAt:   now,
			TriggerID:   opts.TriggerID,
			Status:      SessionStatusIdle,
			LastActive:  now,
			SessionID:   opts.SessionID,
		},
		runner: runner,
		agent:  ag,
		cancel: cancel,
	}

	// If resuming a session, load the messages
	if opts.SessionID != "" && m.config.Services != nil {
		messages, err := m.config.Services.Messages.List(ctx, opts.SessionID)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("load session messages: %w", err)
		}
		if err := runner.ResumeSession(ctx, ag, opts.SessionID, messages); err != nil {
			cancel()
			return nil, fmt.Errorf("resume session: %w", err)
		}
	}

	m.mu.Lock()
	m.sessions[id] = sess
	m.mu.Unlock()

	m.logger.Info("agent woken", "agent", handle, "session", id)

	// Keep session context alive until cancelled
	go func() {
		<-sessionCtx.Done()
		m.mu.Lock()
		if s, ok := m.sessions[id]; ok {
			s.Status = SessionStatusStopped
		}
		m.mu.Unlock()
	}()

	return &sess.AgentSession, nil
}

// Sleep gracefully stops an agent session.
func (m *SessionManager) Sleep(ctx context.Context, handle string) error {
	// Normalize handle
	if len(handle) > 0 && handle[0] != '@' {
		handle = "@" + handle
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, sess := range m.sessions {
		if sess.AgentHandle == handle && sess.Status != SessionStatusStopped {
			m.logger.Info("putting agent to sleep", "agent", handle, "session", id)
			sess.Status = SessionStatusStopped
			if sess.cancel != nil {
				sess.cancel()
			}
			return nil
		}
	}

	return fmt.Errorf("no active session for agent: %s", handle)
}

// List returns all active sessions.
func (m *SessionManager) List() []AgentSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]AgentSession, 0, len(m.sessions))
	for _, sess := range m.sessions {
		result = append(result, sess.AgentSession)
	}
	return result
}

// Get returns a specific session by ID.
func (m *SessionManager) Get(id string) (*AgentSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if sess, ok := m.sessions[id]; ok {
		return &sess.AgentSession, nil
	}
	return nil, fmt.Errorf("session not found: %s", id)
}

// GetByAgent returns the active session for an agent handle.
func (m *SessionManager) GetByAgent(handle string) (*AgentSession, error) {
	// Normalize handle
	if len(handle) > 0 && handle[0] != '@' {
		handle = "@" + handle
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sess := range m.sessions {
		if sess.AgentHandle == handle && sess.Status != SessionStatusStopped {
			return &sess.AgentSession, nil
		}
	}
	return nil, fmt.Errorf("no active session for agent: %s", handle)
}

// StopSession stops a specific session by ID.
func (m *SessionManager) StopSession(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}

	m.logger.Info("stopping session", "session", id, "agent", sess.AgentHandle)
	sess.Status = SessionStatusStopped
	if sess.cancel != nil {
		sess.cancel()
	}

	return nil
}

// Inject sends a message to an active agent session.
func (m *SessionManager) Inject(ctx context.Context, id string, message string) (string, error) {
	m.mu.Lock()
	sess, ok := m.sessions[id]
	if !ok {
		m.mu.Unlock()
		return "", fmt.Errorf("session not found: %s", id)
	}

	if sess.Status == SessionStatusStopped {
		m.mu.Unlock()
		return "", fmt.Errorf("session is stopped: %s", id)
	}

	// Mark as running
	sess.Status = SessionStatusRunning
	sess.LastActive = time.Now()
	m.mu.Unlock()

	// Execute chat
	response, err := sess.runner.Chat(ctx, sess.agent, message)

	// Update status
	m.mu.Lock()
	if s, ok := m.sessions[id]; ok {
		s.Status = SessionStatusIdle
		s.LastActive = time.Now()
		if sess.runner.GetSessionID(sess.agent.Handle) != "" {
			s.SessionID = sess.runner.GetSessionID(sess.agent.Handle)
		}
	}
	m.mu.Unlock()

	if err != nil {
		return "", fmt.Errorf("chat: %w", err)
	}

	return response, nil
}

// MarkActive updates the last active time for a session.
func (m *SessionManager) MarkActive(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sess, ok := m.sessions[id]; ok {
		sess.LastActive = time.Now()
	}
}

// idleTimeoutLoop checks for and stops idle sessions.
func (m *SessionManager) idleTimeoutLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.IdleTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkIdleSessions()
		}
	}
}

// checkIdleSessions stops sessions that have been idle too long.
func (m *SessionManager) checkIdleSessions() {
	threshold := time.Now().Add(-m.config.IdleTimeout)

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, sess := range m.sessions {
		if sess.Status == SessionStatusIdle && sess.LastActive.Before(threshold) {
			m.logger.Info("stopping idle session", "session", id, "agent", sess.AgentHandle,
				"idle_since", sess.LastActive)
			sess.Status = SessionStatusStopped
			if sess.cancel != nil {
				sess.cancel()
			}
		}
	}
}

// CleanupStopped removes stopped sessions from tracking.
func (m *SessionManager) CleanupStopped() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, sess := range m.sessions {
		if sess.Status == SessionStatusStopped {
			delete(m.sessions, id)
		}
	}
}

// generateSessionID generates a unique session ID.
func generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
