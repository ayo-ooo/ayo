// Package approval provides session-scoped approval caching for file modifications.
package approval

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

// Cache stores approval decisions for the current session.
// It is NOT persisted to disk for security reasons.
type Cache struct {
	mu        sync.RWMutex
	patterns  []Pattern   // Pattern-based approvals
	allFiles  bool        // "Always for session" was selected
	sessionID string      // Session this cache belongs to
}

// Pattern represents a glob pattern approval.
type Pattern struct {
	Pattern   string    // glob pattern like "*.md" or "src/**/*.go"
	Directory string    // scoped to this directory
	CreatedAt time.Time
}

// NewCache creates a new approval cache for a session.
func NewCache(sessionID string) *Cache {
	return &Cache{
		sessionID: sessionID,
		patterns:  make([]Pattern, 0),
	}
}

// IsApproved checks if a path is pre-approved based on cached decisions.
func (c *Cache) IsApproved(path string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check "always approve" flag
	if c.allFiles {
		return true
	}

	// Check pattern-based approvals
	for _, p := range c.patterns {
		if c.matchesPattern(path, p) {
			return true
		}
	}

	return false
}

// ApproveAll sets the cache to auto-approve all future requests.
func (c *Cache) ApproveAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allFiles = true
}

// AddPattern adds a glob pattern approval scoped to a directory.
func (c *Cache) AddPattern(pattern, directory string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.patterns = append(c.patterns, Pattern{
		Pattern:   pattern,
		Directory: directory,
		CreatedAt: time.Now(),
	})
}

// Clear resets the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allFiles = false
	c.patterns = make([]Pattern, 0)
}

// matchesPattern checks if a path matches an approval pattern.
func (c *Cache) matchesPattern(path string, p Pattern) bool {
	// If directory is specified, check path is under it
	if p.Directory != "" {
		relPath, err := filepath.Rel(p.Directory, path)
		if err != nil || len(relPath) > 0 && relPath[0] == '.' {
			return false
		}
		// Match pattern against relative path
		matched, _ := doublestar.Match(p.Pattern, relPath)
		return matched
	}

	// Match against full path
	matched, _ := doublestar.Match(p.Pattern, filepath.Base(path))
	return matched
}

// ApprovalType indicates how a request was approved.
type ApprovalType string

const (
	ApprovalNone         ApprovalType = ""              // Not approved
	ApprovalSessionCache ApprovalType = "session_cache" // From cache
	ApprovalNoJodas      ApprovalType = "no_jodas"      // CLI flag
	ApprovalAgentConfig  ApprovalType = "agent_config"  // Agent auto_approve
	ApprovalGlobalConfig ApprovalType = "global_config" // Global setting
	ApprovalUserApproved ApprovalType = "user_approved" // User clicked yes
)

// Manager manages approval caches for multiple sessions.
type Manager struct {
	mu     sync.RWMutex
	caches map[string]*Cache
}

// NewManager creates a new approval manager.
func NewManager() *Manager {
	return &Manager{
		caches: make(map[string]*Cache),
	}
}

// GetCache returns or creates a cache for a session.
func (m *Manager) GetCache(sessionID string) *Cache {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cache, ok := m.caches[sessionID]; ok {
		return cache
	}

	cache := NewCache(sessionID)
	m.caches[sessionID] = cache
	return cache
}

// ClearSession removes a session's cache.
func (m *Manager) ClearSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.caches, sessionID)
}

// Global manager instance
var (
	globalManager *Manager
	managerOnce   sync.Once
)

// GetManager returns the global approval manager.
func GetManager() *Manager {
	managerOnce.Do(func() {
		globalManager = NewManager()
	})
	return globalManager
}

// IsApproved is a convenience function to check if a path is approved for a session.
func IsApproved(sessionID, path string) bool {
	return GetManager().GetCache(sessionID).IsApproved(path)
}
