package ui

// AsyncOperation identifies the type of async operation
type AsyncOperation int

const (
	// AsyncOpMemoryStore is for storing a memory
	AsyncOpMemoryStore AsyncOperation = iota
	// AsyncOpMemoryEmbed is for generating embeddings
	AsyncOpMemoryEmbed
)

func (op AsyncOperation) String() string {
	switch op {
	case AsyncOpMemoryStore:
		return "memory store"
	case AsyncOpMemoryEmbed:
		return "memory embed"
	default:
		return "unknown"
	}
}

// AsyncStatus represents the status of an async operation
type AsyncStatus int

const (
	// AsyncStatusPending means the operation is queued
	AsyncStatusPending AsyncStatus = iota
	// AsyncStatusInProgress means the operation is running
	AsyncStatusInProgress
	// AsyncStatusCompleted means the operation succeeded
	AsyncStatusCompleted
	// AsyncStatusFailed means the operation failed
	AsyncStatusFailed
)

func (s AsyncStatus) String() string {
	switch s {
	case AsyncStatusPending:
		return "pending"
	case AsyncStatusInProgress:
		return "in progress"
	case AsyncStatusCompleted:
		return "completed"
	case AsyncStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// AsyncStatusMsg is sent to the UI when an async operation status changes
type AsyncStatusMsg struct {
	ID        string         // Unique ID for this operation
	Operation AsyncOperation // Type of operation
	Status    AsyncStatus    // Current status
	Message   string         // Human-readable message (e.g., "Storing memory..." or error text)
}
