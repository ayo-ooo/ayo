// Package session provides stub implementation for build system compatibility.
// This package is maintained for backward compatibility but has no functionality
// in the build system architecture.

package session

// Session represents a stub session for compatibility.
type Session struct {
	ID string
}

// New creates a new stub session.
func New(id string) *Session {
	return &Session{ID: id}
}

// GetID returns the session ID.
func (s *Session) GetID() string {
	return s.ID
}

// Close is a no-op for compatibility.
func (s *Session) Close() error {
	return nil
}
