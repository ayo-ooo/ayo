package daemon

import (
	"sync"
	"time"
)

// DaemonSessionManager manages agent sessions within the daemon.
// This is a simpler version than the HTTP server's SessionManager since the daemon
// uses RPC calls instead of direct agent execution.
type DaemonSessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*DaemonSession
	idleTimeout time.Duration
	stopCh      chan struct{}
	wg          sync.WaitGroup
	running     bool
}

// DaemonSession represents an active agent session in the daemon.
type DaemonSession struct {
	ID          string    `json:"id"`
	AgentHandle string    `json:"agent_handle"`
	StartedAt   time.Time `json:"started_at"`
	TriggerID   string    `json:"trigger_id,omitempty"`
	Status      string    `json:"status"` // running, idle, stopped
	LastActive  time.Time `json:"last_active"`
	SessionID   string    `json:"session_id,omitempty"` // Database session ID
}

// NewDaemonSessionManager creates a new daemon session manager.
func NewDaemonSessionManager(idleTimeout time.Duration) *DaemonSessionManager {
	return &DaemonSessionManager{
		sessions:    make(map[string]*DaemonSession),
		idleTimeout: idleTimeout,
		stopCh:      make(chan struct{}),
	}
}

// Start starts the session manager.
func (m *DaemonSessionManager) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	if m.idleTimeout > 0 {
		m.wg.Add(1)
		go m.idleCheckLoop()
	}
}

// Stop stops the session manager.
func (m *DaemonSessionManager) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	close(m.stopCh)
	m.mu.Unlock()

	m.wg.Wait()
}

// List returns all active sessions.
func (m *DaemonSessionManager) List() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]SessionInfo, 0, len(m.sessions))
	for _, sess := range m.sessions {
		if sess.Status != "stopped" {
			result = append(result, sessionToInfo(sess))
		}
	}
	return result
}

// Wake starts or returns an existing session for an agent.
func (m *DaemonSessionManager) Wake(handle, triggerID, resumeSessionID string) SessionInfo {
	// Normalize handle
	if len(handle) > 0 && handle[0] != '@' {
		handle = "@" + handle
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for existing active session
	for _, sess := range m.sessions {
		if sess.AgentHandle == handle && sess.Status != "stopped" {
			sess.LastActive = time.Now()
			return sessionToInfo(sess)
		}
	}

	// Create new session
	now := time.Now()
	id := generateDaemonSessionID()

	sess := &DaemonSession{
		ID:          id,
		AgentHandle: handle,
		StartedAt:   now,
		TriggerID:   triggerID,
		Status:      "idle",
		LastActive:  now,
		SessionID:   resumeSessionID,
	}

	m.sessions[id] = sess
	return sessionToInfo(sess)
}

// Sleep stops an agent's session.
func (m *DaemonSessionManager) Sleep(handle string) error {
	// Normalize handle
	if len(handle) > 0 && handle[0] != '@' {
		handle = "@" + handle
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sess := range m.sessions {
		if sess.AgentHandle == handle && sess.Status != "stopped" {
			sess.Status = "stopped"
			return nil
		}
	}

	return NewError(ErrCodeSessionNotFound, "no active session for agent: "+handle)
}

// GetByAgent returns the session for an agent.
func (m *DaemonSessionManager) GetByAgent(handle string) (*SessionInfo, error) {
	// Normalize handle
	if len(handle) > 0 && handle[0] != '@' {
		handle = "@" + handle
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sess := range m.sessions {
		if sess.AgentHandle == handle && sess.Status != "stopped" {
			info := sessionToInfo(sess)
			return &info, nil
		}
	}

	return nil, NewError(ErrCodeSessionNotFound, "no active session for agent: "+handle)
}

// GetByID returns a session by ID.
func (m *DaemonSessionManager) GetByID(id string) (*SessionInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if sess, ok := m.sessions[id]; ok {
		info := sessionToInfo(sess)
		return &info, nil
	}

	return nil, NewError(ErrCodeSessionNotFound, "session not found: "+id)
}

// StopSession stops a session by ID.
func (m *DaemonSessionManager) StopSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[id]
	if !ok {
		return NewError(ErrCodeSessionNotFound, "session not found: "+id)
	}

	sess.Status = "stopped"
	return nil
}

// idleCheckLoop checks for idle sessions periodically.
func (m *DaemonSessionManager) idleCheckLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.idleTimeout / 2)
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
func (m *DaemonSessionManager) checkIdleSessions() {
	threshold := time.Now().Add(-m.idleTimeout)

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sess := range m.sessions {
		if sess.Status == "idle" && sess.LastActive.Before(threshold) {
			sess.Status = "stopped"
		}
	}
}

// sessionToInfo converts a DaemonSession to SessionInfo.
func sessionToInfo(sess *DaemonSession) SessionInfo {
	return SessionInfo{
		ID:          sess.ID,
		AgentHandle: sess.AgentHandle,
		StartedAt:   sess.StartedAt.Unix(),
		TriggerID:   sess.TriggerID,
		Status:      sess.Status,
		LastActive:  sess.LastActive.Unix(),
		SessionID:   sess.SessionID,
	}
}

// generateDaemonSessionID generates a unique session ID.
func generateDaemonSessionID() string {
	return "dsess_" + time.Now().Format("20060102150405.000000")
}

// Application error codes for sessions
const (
	ErrCodeSessionNotFound = -2001
)
