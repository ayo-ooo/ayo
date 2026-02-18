// Package tickets provides a file-based ticket system for agent coordination.
package tickets

import "time"

// Status represents the current state of a ticket.
type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusClosed     Status = "closed"
)

// Valid returns true if the status is a known value.
func (s Status) Valid() bool {
	switch s {
	case StatusOpen, StatusInProgress, StatusBlocked, StatusClosed:
		return true
	}
	return false
}

// Type represents the category of work a ticket represents.
type Type string

const (
	TypeEpic       Type = "epic"
	TypeFeature    Type = "feature"
	TypeTask       Type = "task"
	TypeBug        Type = "bug"
	TypeChore      Type = "chore"
	TypeEscalation Type = "escalation"
)

// Valid returns true if the type is a known value.
func (t Type) Valid() bool {
	switch t {
	case TypeEpic, TypeFeature, TypeTask, TypeBug, TypeChore, TypeEscalation:
		return true
	}
	return false
}

// Ticket represents a unit of work with metadata, description, and notes.
type Ticket struct {
	// Frontmatter fields (stored in YAML)
	ID         string    `yaml:"id"`
	Status     Status    `yaml:"status"`
	Type       Type      `yaml:"type"`
	Priority   int       `yaml:"priority"`
	Assignee   string    `yaml:"assignee,omitempty"`
	Deps       []string  `yaml:"deps"`
	Links      []string  `yaml:"links"`
	Parent     string    `yaml:"parent,omitempty"`
	Tags       []string  `yaml:"tags,omitempty"`
	Created    time.Time `yaml:"created"`
	Started    *time.Time `yaml:"started,omitempty"`
	Closed     *time.Time `yaml:"closed,omitempty"`
	Session    string    `yaml:"session,omitempty"`
	ExternalRef string   `yaml:"external_ref,omitempty"`

	// Parsed from markdown body
	Title       string `yaml:"-"`
	Description string `yaml:"-"`
	Notes       []Note `yaml:"-"`

	// Metadata (not serialized)
	FilePath string `yaml:"-"`
}

// Note represents a timestamped comment on a ticket.
type Note struct {
	Timestamp time.Time
	Content   string
}

// Filter specifies criteria for listing tickets.
type Filter struct {
	Status   Status
	Assignee string
	Type     Type
	Tags     []string
	Parent   string
}

// CreateOptions specifies parameters for creating a new ticket.
type CreateOptions struct {
	Title       string
	Description string
	Type        Type
	Priority    int
	Assignee    string
	Deps        []string
	Parent      string
	Tags        []string
	ExternalRef string
}

// DefaultPriority is the default priority for new tickets (middle of 0-4 range).
const DefaultPriority = 2

// ValidatePriority returns true if priority is in valid range 0-4.
func ValidatePriority(p int) bool {
	return p >= 0 && p <= 4
}
