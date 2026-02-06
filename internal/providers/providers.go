// Package providers defines interfaces for pluggable subsystems in ayo.
// Providers enable swappable implementations for memory, sandbox, embedding, and observers.
package providers

import (
	"context"
	"time"
)

// ProviderType identifies the category of provider.
type ProviderType string

const (
	ProviderTypeMemory    ProviderType = "memory"
	ProviderTypeSandbox   ProviderType = "sandbox"
	ProviderTypeEmbedding ProviderType = "embedding"
	ProviderTypeObserver  ProviderType = "observer"
)

// Provider is the base interface all providers must implement.
type Provider interface {
	// Name returns the unique identifier for this provider.
	Name() string

	// Type returns the provider category.
	Type() ProviderType

	// Init initializes the provider with the given configuration.
	// Called once when the provider is loaded.
	Init(ctx context.Context, config map[string]any) error

	// Close releases any resources held by the provider.
	Close() error
}

// MemoryProvider defines the interface for memory storage backends.
type MemoryProvider interface {
	Provider

	// Create stores a new memory and returns it with generated fields (ID, timestamps).
	Create(ctx context.Context, m Memory) (Memory, error)

	// Get retrieves a memory by ID.
	Get(ctx context.Context, id string) (Memory, error)

	// Search finds memories matching the query with semantic and/or text search.
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)

	// List returns all active memories, optionally filtered.
	List(ctx context.Context, opts ListOptions) ([]Memory, error)

	// Update modifies an existing memory.
	Update(ctx context.Context, m Memory) error

	// Forget soft-deletes a memory by ID.
	Forget(ctx context.Context, id string) error

	// Supersede replaces an old memory with a new one, tracking the relationship.
	Supersede(ctx context.Context, oldID string, newMemory Memory, reason string) (Memory, error)

	// Topics returns the list of known topics.
	Topics(ctx context.Context) ([]string, error)

	// Link creates a bidirectional link between two memories.
	Link(ctx context.Context, id1, id2 string) error

	// Unlink removes a link between two memories.
	Unlink(ctx context.Context, id1, id2 string) error

	// Reindex rebuilds any derived indexes from source files.
	Reindex(ctx context.Context) error
}

// Memory represents a stored memory unit.
type Memory struct {
	ID                 string
	Content            string
	Category           MemoryCategory
	Topics             []string
	AgentHandle        string    // Empty for global memories
	PathScope          string    // Empty for non-path-scoped memories
	SourceSessionID    string
	SourceMessageID    string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Confidence         float64
	LastAccessedAt     time.Time
	AccessCount        int64
	SupersedesID       string
	SupersededByID     string
	SupersessionReason string
	Status             MemoryStatus
	Unclear            bool   // True if memory needs clarification from user
	UnclearReason      string // Why the memory is unclear
	LinkedIDs          []string
	Embedding          []float32 // Optional, populated by embedding provider
}

// MemoryCategory classifies the type of memory.
type MemoryCategory string

const (
	MemoryCategoryPreference  MemoryCategory = "preference"
	MemoryCategoryFact        MemoryCategory = "fact"
	MemoryCategoryCorrection  MemoryCategory = "correction"
	MemoryCategoryPattern     MemoryCategory = "pattern"
)

// MemoryStatus represents the lifecycle state of a memory.
type MemoryStatus string

const (
	MemoryStatusActive     MemoryStatus = "active"
	MemoryStatusSuperseded MemoryStatus = "superseded"
	MemoryStatusArchived   MemoryStatus = "archived"
	MemoryStatusForgotten  MemoryStatus = "forgotten"
)

// SearchOptions configures memory search.
type SearchOptions struct {
	AgentHandle string
	PathScope   string
	Threshold   float32        // Minimum similarity (0-1)
	Limit       int
	Categories  []MemoryCategory
	Topics      []string
	Status      []MemoryStatus // Default: active only
}

// ListOptions configures memory listing.
type ListOptions struct {
	AgentHandle string
	PathScope   string
	Categories  []MemoryCategory
	Topics      []string
	Status      []MemoryStatus
	Limit       int
	Offset      int
}

// SearchResult contains a memory with its search relevance score.
type SearchResult struct {
	Memory     Memory
	Similarity float32 // 0-1, higher is more similar
	MatchType  string  // "semantic", "text", or "hybrid"
}

// SandboxProvider defines the interface for container/sandbox runtimes.
type SandboxProvider interface {
	Provider

	// Create creates a new sandbox container.
	Create(ctx context.Context, opts SandboxCreateOptions) (Sandbox, error)

	// Get retrieves a sandbox by ID.
	Get(ctx context.Context, id string) (Sandbox, error)

	// List returns all sandboxes.
	List(ctx context.Context) ([]Sandbox, error)

	// Start starts a stopped sandbox.
	Start(ctx context.Context, id string) error

	// Stop stops a running sandbox.
	Stop(ctx context.Context, id string, opts SandboxStopOptions) error

	// Delete removes a sandbox.
	Delete(ctx context.Context, id string, force bool) error

	// Exec executes a command inside a sandbox.
	Exec(ctx context.Context, id string, opts ExecOptions) (ExecResult, error)

	// Status returns the current state of a sandbox.
	Status(ctx context.Context, id string) (SandboxStatus, error)

	// Stats returns resource usage statistics for a sandbox.
	Stats(ctx context.Context, id string) (SandboxStats, error)

	// EnsureAgentUser ensures a Unix user exists for the agent in the sandbox.
	// Creates the user and home directory if they don't exist.
	// If dotfilesPath is non-empty, copies dotfiles from that host directory
	// to the user's home directory (e.g., .bashrc, .profile).
	EnsureAgentUser(ctx context.Context, id string, agentHandle string, dotfilesPath string) error
}

// Sandbox represents a container instance.
type Sandbox struct {
	ID        string
	Name      string
	Image     string
	Status    SandboxStatus
	Pool      string   // Pool this sandbox belongs to
	Agents    []string // Agent handles assigned to this sandbox
	User      string   // User account created in the sandbox (empty = root)
	CreatedAt time.Time
	Mounts    []Mount
	Resources Resources
}

// SandboxStatus represents the state of a sandbox.
type SandboxStatus string

const (
	SandboxStatusCreating SandboxStatus = "creating"
	SandboxStatusRunning  SandboxStatus = "running"
	SandboxStatusStopped  SandboxStatus = "stopped"
	SandboxStatusFailed   SandboxStatus = "failed"
)

// SandboxCreateOptions configures sandbox creation.
type SandboxCreateOptions struct {
	Name      string
	Image     string
	Pool      string
	Mounts    []Mount
	Resources Resources
	Network   NetworkConfig
	Labels    map[string]string

	// User specifies the username to create and run as inside the sandbox.
	// If specified, a user account will be created at startup with UID 1000
	// and a home directory at /home/{user}.
	// If empty, commands run as root.
	User string

	// SetupCommands are run as root after container creation but before
	// the sandbox is considered ready. Used for user creation, package
	// installation, etc.
	SetupCommands [][]string
}

// SandboxStopOptions configures sandbox stopping.
type SandboxStopOptions struct {
	Timeout time.Duration // Time to wait before force kill
	Signal  string        // Signal to send (default SIGTERM)
}

// Mount represents a filesystem mount.
type Mount struct {
	Source      string    // Host path
	Destination string    // Container path
	Mode        MountMode
	ReadOnly    bool
}

// MountMode specifies the mount type.
type MountMode string

const (
	MountModeVirtioFS MountMode = "virtiofs"
	MountModeBind     MountMode = "bind"
	MountModeOverlay  MountMode = "overlay"
	MountModeTmpfs    MountMode = "tmpfs"
)

// Resources defines resource limits for a sandbox.
type Resources struct {
	CPUs      int   // Number of CPUs
	MemoryMB  int64 // Memory in megabytes
	DiskMB    int64 // Disk space in megabytes
}

// NetworkConfig configures sandbox networking.
type NetworkConfig struct {
	Enabled bool
	DNS     []string
	Ports   []PortMapping
}

// PortMapping maps a host port to a container port.
type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string // tcp or udp
}

// ExecOptions configures command execution in a sandbox.
type ExecOptions struct {
	Command    string
	Args       []string
	WorkingDir string
	Env        map[string]string
	Timeout    time.Duration
	User       string
	Stdin      []byte
}

// ExecResult contains the output of a command execution.
type ExecResult struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	TimedOut  bool
	Truncated bool
	Duration  time.Duration
}

// SandboxStats contains resource usage statistics for a sandbox.
type SandboxStats struct {
	// CPU usage as a percentage (0-100 per CPU core)
	CPUPercent float64

	// Memory usage in bytes
	MemoryUsageBytes int64

	// Memory limit in bytes (0 if unlimited)
	MemoryLimitBytes int64

	// Disk usage in bytes
	DiskUsageBytes int64

	// Network bytes received
	NetworkRxBytes int64

	// Network bytes transmitted
	NetworkTxBytes int64

	// Number of running processes
	ProcessCount int

	// Uptime since container start
	Uptime time.Duration
}

// EmbeddingProvider defines the interface for vector embedding generation.
type EmbeddingProvider interface {
	Provider

	// Embed generates an embedding vector for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts.
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimensions returns the dimensionality of embeddings produced by this provider.
	Dimensions() int

	// Model returns the name of the embedding model being used.
	Model() string
}

// ObserverProvider defines the interface for session observation and processing.
type ObserverProvider interface {
	Provider

	// Start begins observing sessions.
	Start(ctx context.Context) error

	// Stop stops the observer.
	Stop() error

	// OnMessage is called when a new message is added to a session.
	OnMessage(ctx context.Context, event MessageEvent) error
}

// MessageEvent represents a message in a session that the observer should process.
type MessageEvent struct {
	SessionID   string
	MessageID   string
	Role        string // "user", "assistant", "system"
	Content     string
	AgentHandle string
	Timestamp   time.Time
}
