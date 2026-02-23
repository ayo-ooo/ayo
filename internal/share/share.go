// Package share provides management for user-shared host directories.
package share

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
)

// Share represents a single shared host path.
type Share struct {
	Name      string    `json:"name"`                 // Name in /workspace/
	Path      string    `json:"path"`                 // Absolute host path
	Session   bool      `json:"session"`              // If true, removed when session ends
	SessionID string    `json:"session_id,omitempty"` // Session ID if session share
	SharedAt  time.Time `json:"shared_at"`
}

// SharesFile represents the shares.json file structure.
type SharesFile struct {
	Version int     `json:"version"`
	Shares  []Share `json:"shares"`
}

// Service manages filesystem shares.
type Service struct {
	mu       sync.RWMutex
	filePath string
	shares   *SharesFile
}

// NewService creates a new share service.
func NewService() *Service {
	return &Service{
		filePath: sharesFilePath(),
	}
}

// sharesFilePath returns the path to the shares.json file.
func sharesFilePath() string {
	return filepath.Join(paths.DataDir(), "shares.json")
}

// Load reads the shares file from disk.
func (s *Service) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize with empty shares
			s.shares = &SharesFile{
				Version: 1,
				Shares:  []Share{},
			}
			return nil
		}
		return fmt.Errorf("read shares file: %w", err)
	}

	var shares SharesFile
	if err := json.Unmarshal(data, &shares); err != nil {
		return fmt.Errorf("parse shares file: %w", err)
	}

	s.shares = &shares
	return nil
}

// Save writes the shares file to disk.
func (s *Service) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.saveUnlocked()
}

// saveUnlocked writes shares to disk without acquiring lock.
// Caller must hold the lock.
func (s *Service) saveUnlocked() error {
	if s.shares == nil {
		return fmt.Errorf("no shares loaded")
	}

	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create shares directory: %w", err)
	}

	data, err := json.MarshalIndent(s.shares, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal shares: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("write shares file: %w", err)
	}

	return nil
}

// List returns all shares.
func (s *Service) List() []Share {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shares == nil {
		return nil
	}

	// Return a copy to prevent modification
	result := make([]Share, len(s.shares.Shares))
	copy(result, s.shares.Shares)
	return result
}

// Get returns a share by workspace name, or nil if not found.
func (s *Service) Get(name string) *Share {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shares == nil {
		return nil
	}

	for _, share := range s.shares.Shares {
		if share.Name == name {
			return &share
		}
	}

	return nil
}

// GetByPath returns a share by original host path, or nil if not found.
func (s *Service) GetByPath(path string) *Share {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shares == nil {
		return nil
	}

	// Normalize path for comparison
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil
	}

	for _, share := range s.shares.Shares {
		if share.Path == absPath {
			return &share
		}
	}

	return nil
}

// Add creates a new share with a symlink in the workspace directory.
func (s *Service) Add(path, name string, session bool, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shares == nil {
		s.shares = &SharesFile{Version: 1, Shares: []Share{}}
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Validate path exists
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// Generate name if empty
	if name == "" {
		name = filepath.Base(absPath)
	}

	// Validate name is safe
	if err := validateShareName(name); err != nil {
		return err
	}

	// Check for existing share with same name
	for _, share := range s.shares.Shares {
		if share.Name == name {
			return fmt.Errorf("share '%s' already exists, use --as to specify a different name", name)
		}
	}

	// NOTE: We no longer create symlinks here. Shares are mounted as direct
	// VirtioFS mounts when the sandbox is created (see internal/sandbox/ayo.go).
	// This avoids the issue where symlinks don't work inside containers.

	// Add to shares list
	s.shares.Shares = append(s.shares.Shares, Share{
		Name:      name,
		Path:      absPath,
		Session:   session,
		SessionID: sessionID,
		SharedAt:  time.Now(),
	})

	return s.saveUnlocked()
}

// Remove deletes a share by name or path.
func (s *Service) Remove(nameOrPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shares == nil {
		return nil
	}

	// Find share by name first
	var found *Share
	var foundIndex int
	for i, share := range s.shares.Shares {
		if share.Name == nameOrPath {
			found = &share
			foundIndex = i
			break
		}
	}

	// If not found by name, try by path
	if found == nil {
		absPath, _ := filepath.Abs(nameOrPath)
		for i, share := range s.shares.Shares {
			if share.Path == absPath {
				found = &share
				foundIndex = i
				break
			}
		}
	}

	if found == nil {
		return nil // Not found is not an error (idempotent)
	}

	// NOTE: We no longer remove symlinks here - shares are now direct VirtioFS mounts.
	// The mount will be removed on next sandbox restart.

	// Remove from list
	s.shares.Shares = append(s.shares.Shares[:foundIndex], s.shares.Shares[foundIndex+1:]...)

	return s.saveUnlocked()
}

// RemoveSessionShares removes all shares associated with a session ID.
func (s *Service) RemoveSessionShares(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shares == nil {
		return nil
	}

	// Find and remove session shares
	// NOTE: No symlinks to remove - shares are now direct VirtioFS mounts.
	var remaining []Share
	for _, share := range s.shares.Shares {
		if !(share.Session && share.SessionID == sessionID) {
			remaining = append(remaining, share)
		}
	}

	s.shares.Shares = remaining
	return s.saveUnlocked()
}

// validateShareName checks if a name is safe to use as a symlink name.
func validateShareName(name string) error {
	if name == "" {
		return fmt.Errorf("share name cannot be empty")
	}

	if name == "." || name == ".." {
		return fmt.Errorf("share name cannot be '.' or '..'")
	}

	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("share name cannot contain path separators")
	}

	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("share name cannot start with '-'")
	}

	return nil
}
