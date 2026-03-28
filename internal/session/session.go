package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Session represents a multi-turn conversation session.
type Session struct {
	ID        string    `json:"id"`
	AgentName string    `json:"agent_name"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store manages session persistence on disk.
type Store struct {
	// baseDir overrides the default storage directory for testing.
	baseDir string
}

// NewStore creates a new Store using the default storage directory.
func NewStore() *Store {
	return &Store{}
}

// NewStoreWithDir creates a new Store using the given base directory.
// This is useful for testing.
func NewStoreWithDir(dir string) *Store {
	return &Store{baseDir: dir}
}

// sessionsDir returns the directory where sessions are stored for the given agent.
func (s *Store) sessionsDir(agentName string) string {
	base := s.baseDir
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		base = filepath.Join(home, ".local", "share", "agents")
	}
	return filepath.Join(base, agentName, "sessions")
}

// sessionPath returns the file path for a given session.
func (s *Store) sessionPath(agentName, sessionID string) string {
	return filepath.Join(s.sessionsDir(agentName), sessionID+".json")
}

// New creates a new session with a random ID.
func (s *Store) New(agentName string) *Session {
	now := time.Now()
	return &Session{
		ID:        generateID(),
		AgentName: agentName,
		Messages:  []Message{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Save persists a session to disk as a JSON file.
func (s *Store) Save(sess *Session) error {
	dir := s.sessionsDir(sess.AgentName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating sessions directory: %w", err)
	}

	sess.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}

	path := s.sessionPath(sess.AgentName, sess.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing session file: %w", err)
	}

	return nil
}

// Load reads a session from disk by agent name and session ID.
func (s *Store) Load(agentName, sessionID string) (*Session, error) {
	path := s.sessionPath(agentName, sessionID)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session %q not found", sessionID)
		}
		return nil, fmt.Errorf("reading session file: %w", err)
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("parsing session file: %w", err)
	}

	return &sess, nil
}

// List returns the IDs of all sessions for the given agent.
func (s *Store) List(agentName string) ([]string, error) {
	dir := s.sessionsDir(agentName)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading sessions directory: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			ids = append(ids, strings.TrimSuffix(name, ".json"))
		}
	}

	return ids, nil
}

// generateID returns an 8-character random hex string.
func generateID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails.
		return fmt.Sprintf("%08x", time.Now().UnixNano()&0xFFFFFFFF)
	}
	return hex.EncodeToString(b)
}
