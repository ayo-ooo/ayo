package jsonl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
)

// Structure manages the session directory structure.
type Structure struct {
	Root    string // Root sessions directory
	IndexDB string // Path to index database
}

// NewStructure creates a new session structure manager.
// If root is empty, uses the default sessions directory.
func NewStructure(root string) *Structure {
	if root == "" {
		root = DefaultSessionsDir()
	}
	return &Structure{
		Root:    root,
		IndexDB: filepath.Join(root, "index.sqlite"),
	}
}

// DefaultSessionsDir returns the default sessions directory.
func DefaultSessionsDir() string {
	return filepath.Join(paths.DataDir(), "sessions")
}

// Initialize creates the directory structure if it doesn't exist.
func (s *Structure) Initialize() error {
	// Create root directory
	if err := os.MkdirAll(s.Root, 0755); err != nil {
		return fmt.Errorf("create root: %w", err)
	}
	return nil
}

// Exists returns true if the structure has been initialized.
func (s *Structure) Exists() bool {
	info, err := os.Stat(s.Root)
	return err == nil && info.IsDir()
}

// AgentDir returns the directory for a specific agent.
func (s *Structure) AgentDir(agentHandle string) string {
	// Sanitize agent handle for filesystem
	safe := sanitizeFilename(agentHandle)
	return filepath.Join(s.Root, safe)
}

// MonthDir returns the year-month subdirectory for a given time.
func (s *Structure) MonthDir(agentHandle string, t time.Time) string {
	yearMonth := t.UTC().Format("2006-01")
	return filepath.Join(s.AgentDir(agentHandle), yearMonth)
}

// SessionPath returns the full path for a session file.
func (s *Structure) SessionPath(agentHandle, sessionID string, createdAt time.Time) string {
	filename := sessionID + ".jsonl"
	return filepath.Join(s.MonthDir(agentHandle, createdAt), filename)
}

// SessionPathByID finds a session file by ID (searches all agent dirs).
// Returns the path and agent handle if found.
func (s *Structure) SessionPathByID(sessionID string) (path, agentHandle string, err error) {
	filename := sessionID + ".jsonl"

	entries, err := os.ReadDir(s.Root)
	if err != nil {
		return "", "", fmt.Errorf("read root: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "index.sqlite" {
			continue
		}

		agentDir := filepath.Join(s.Root, entry.Name())
		monthDirs, err := os.ReadDir(agentDir)
		if err != nil {
			continue
		}

		for _, monthEntry := range monthDirs {
			if !monthEntry.IsDir() {
				continue
			}

			fullPath := filepath.Join(agentDir, monthEntry.Name(), filename)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath, entry.Name(), nil
			}
		}
	}

	return "", "", ErrSessionNotFound
}

// ListAgents returns all agent handles that have sessions.
func (s *Structure) ListAgents() ([]string, error) {
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var agents []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			agents = append(agents, entry.Name())
		}
	}
	return agents, nil
}

// ListSessions returns all session files for an agent.
func (s *Structure) ListSessions(agentHandle string) ([]string, error) {
	agentDir := s.AgentDir(agentHandle)

	entries, err := os.ReadDir(agentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sessions []string
	for _, monthEntry := range entries {
		if !monthEntry.IsDir() {
			continue
		}

		monthDir := filepath.Join(agentDir, monthEntry.Name())
		files, err := os.ReadDir(monthDir)
		if err != nil {
			continue
		}

		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".jsonl") {
				sessions = append(sessions, filepath.Join(monthDir, f.Name()))
			}
		}
	}

	return sessions, nil
}

// ListAllSessions returns all session files across all agents.
func (s *Structure) ListAllSessions() ([]string, error) {
	agents, err := s.ListAgents()
	if err != nil {
		return nil, err
	}

	var all []string
	for _, agent := range agents {
		sessions, err := s.ListSessions(agent)
		if err != nil {
			continue
		}
		all = append(all, sessions...)
	}
	return all, nil
}

// EnsureDir creates the directory for a session file.
func (s *Structure) EnsureDir(agentHandle string, createdAt time.Time) error {
	dir := s.MonthDir(agentHandle, createdAt)
	return os.MkdirAll(dir, 0755)
}

// sanitizeFilename makes a string safe for use as a filename.
func sanitizeFilename(name string) string {
	// Remove leading @ from agent handles
	name = strings.TrimPrefix(name, "@")

	// Replace unsafe characters
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(name)
}
