// Package share provides stub implementation for build system compatibility.
// This package is maintained for backward compatibility but has no functionality
// in the build system architecture.

package share

// Share represents a stub share for compatibility.
type Share struct {
	ID string
}

// New creates a new stub share.
func New(id string) *Share {
	return &Share{ID: id}
}

// GetID returns the share ID.
func (s *Share) GetID() string {
	return s.ID
}

// Close is a no-op for compatibility.
func (s *Share) Close() error {
	return nil
}
