package sandbox

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexcabrera/ayo/internal/providers"
)

// PoolConfig configures a sandbox pool.
type PoolConfig struct {
	Name       string
	Provider   string            // Provider name ("none", "applecontainer", etc.)
	Image      string            // Base image to use
	MinSize    int               // Minimum number of warm sandboxes
	MaxSize    int               // Maximum number of sandboxes
	IdleAfter  int               // Seconds before idle sandbox is recycled
	Resources  providers.Resources
	Network    providers.NetworkConfig
	Mounts     []providers.Mount
}

// Pool manages a set of sandboxes.
type Pool struct {
	config    PoolConfig
	provider  providers.SandboxProvider
	sandboxes map[string]*poolEntry
	mu        sync.RWMutex
}

type poolEntry struct {
	sandbox   providers.Sandbox
	inUse     bool
	agents    []string
	group     string // Collaboration group name (agents in same group share sandbox)
}

// NewPool creates a new sandbox pool.
func NewPool(config PoolConfig, provider providers.SandboxProvider) *Pool {
	return &Pool{
		config:    config,
		provider:  provider,
		sandboxes: make(map[string]*poolEntry),
	}
}

// Start initializes the pool with warm sandboxes.
func (p *Pool) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Create minimum number of sandboxes
	for i := 0; i < p.config.MinSize; i++ {
		if err := p.createSandbox(ctx); err != nil {
			return fmt.Errorf("create sandbox %d: %w", i, err)
		}
	}

	return nil
}

// Stop shuts down all sandboxes in the pool.
func (p *Pool) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for id, entry := range p.sandboxes {
		if err := p.provider.Stop(ctx, entry.sandbox.ID, providers.SandboxStopOptions{}); err != nil {
			errs = append(errs, fmt.Errorf("stop %s: %w", id, err))
		}
		if err := p.provider.Delete(ctx, entry.sandbox.ID, true); err != nil {
			errs = append(errs, fmt.Errorf("delete %s: %w", id, err))
		}
	}
	p.sandboxes = make(map[string]*poolEntry)

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}
	return nil
}

// Acquire gets an available sandbox from the pool.
// If no sandbox is available and max not reached, creates a new one.
func (p *Pool) Acquire(ctx context.Context, agentHandle string) (providers.Sandbox, error) {
	return p.AcquireWithOptions(ctx, AcquireOptions{Agent: agentHandle})
}

// AcquireWithOptions gets a sandbox with advanced options for collaboration.
func (p *Pool) AcquireWithOptions(ctx context.Context, opts AcquireOptions) (providers.Sandbox, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If joining a specific sandbox, find it
	if opts.JoinSandbox != "" {
		entry, ok := p.sandboxes[opts.JoinSandbox]
		if !ok {
			return providers.Sandbox{}, fmt.Errorf("sandbox not found: %s", opts.JoinSandbox)
		}
		// Add this agent to the sandbox
		entry.agents = append(entry.agents, opts.Agent)
		return entry.sandbox, nil
	}

	// If in a collaboration group, look for existing sandbox in that group
	if opts.Group != "" {
		for _, entry := range p.sandboxes {
			if entry.group == opts.Group {
				// Add this agent to the group's sandbox
				entry.agents = append(entry.agents, opts.Agent)
				return entry.sandbox, nil
			}
		}
	}

	// First, try to find an existing sandbox for this agent
	for _, entry := range p.sandboxes {
		for _, a := range entry.agents {
			if a == opts.Agent {
				return entry.sandbox, nil
			}
		}
	}

	// Next, try to find an idle sandbox
	for _, entry := range p.sandboxes {
		if !entry.inUse {
			entry.inUse = true
			entry.agents = append(entry.agents, opts.Agent)
			if opts.Group != "" {
				entry.group = opts.Group
			}
			return entry.sandbox, nil
		}
	}

	// Check if we can create a new one
	if p.config.MaxSize > 0 && len(p.sandboxes) >= p.config.MaxSize {
		return providers.Sandbox{}, fmt.Errorf("pool exhausted: max %d sandboxes", p.config.MaxSize)
	}

	// Create a new sandbox
	if err := p.createSandbox(ctx); err != nil {
		return providers.Sandbox{}, err
	}

	// Find the newly created sandbox
	for _, entry := range p.sandboxes {
		if !entry.inUse {
			entry.inUse = true
			entry.agents = append(entry.agents, opts.Agent)
			if opts.Group != "" {
				entry.group = opts.Group
			}
			return entry.sandbox, nil
		}
	}

	return providers.Sandbox{}, fmt.Errorf("failed to acquire sandbox")
}

// Release returns a sandbox to the pool.
func (p *Pool) Release(ctx context.Context, sandboxID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.sandboxes[sandboxID]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	entry.inUse = false
	entry.agents = nil
	entry.group = ""

	// Check if we have too many idle sandboxes
	if p.config.MinSize > 0 {
		idleCount := 0
		for _, e := range p.sandboxes {
			if !e.inUse {
				idleCount++
			}
		}
		if idleCount > p.config.MinSize {
			// Delete this sandbox
			if err := p.provider.Delete(ctx, sandboxID, false); err != nil {
				return err
			}
			delete(p.sandboxes, sandboxID)
		}
	}

	return nil
}

// ReleaseAgent removes an agent from a sandbox without releasing the sandbox.
// The sandbox is only released when all agents have left.
func (p *Pool) ReleaseAgent(ctx context.Context, sandboxID, agentHandle string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.sandboxes[sandboxID]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	// Remove the agent from the list
	newAgents := make([]string, 0, len(entry.agents))
	for _, a := range entry.agents {
		if a != agentHandle {
			newAgents = append(newAgents, a)
		}
	}
	entry.agents = newAgents

	// If no agents left, release the sandbox
	if len(entry.agents) == 0 {
		entry.inUse = false
		entry.group = ""
	}

	return nil
}

// GetSandboxAgents returns the list of agents currently using a sandbox.
func (p *Pool) GetSandboxAgents(sandboxID string) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entry, ok := p.sandboxes[sandboxID]
	if !ok {
		return nil
	}

	result := make([]string, len(entry.agents))
	copy(result, entry.agents)
	return result
}

// Exec executes a command in a sandbox.
func (p *Pool) Exec(ctx context.Context, sandboxID string, opts providers.ExecOptions) (providers.ExecResult, error) {
	p.mu.RLock()
	_, ok := p.sandboxes[sandboxID]
	p.mu.RUnlock()

	if !ok {
		return providers.ExecResult{}, fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	return p.provider.Exec(ctx, sandboxID, opts)
}

// Status returns pool status.
func (p *Pool) Status() PoolStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := PoolStatus{
		Name:     p.config.Name,
		Provider: p.config.Provider,
		Total:    len(p.sandboxes),
	}

	for _, entry := range p.sandboxes {
		if entry.inUse {
			status.InUse++
		} else {
			status.Idle++
		}
	}

	return status
}

// PoolStatus contains pool status information.
type PoolStatus struct {
	Name     string
	Provider string
	Total    int
	InUse    int
	Idle     int
}

// AcquireOptions configures sandbox acquisition.
type AcquireOptions struct {
	// Agent is the agent handle acquiring the sandbox.
	Agent string

	// Group specifies a collaboration group. Agents in the same group share sandboxes.
	// If empty, the agent gets its own sandbox.
	Group string

	// JoinSandbox specifies an existing sandbox ID to join.
	// Takes precedence over Group if specified.
	JoinSandbox string
}

func (p *Pool) createSandbox(ctx context.Context) error {
	sb, err := p.provider.Create(ctx, providers.SandboxCreateOptions{
		Image:     p.config.Image,
		Pool:      p.config.Name,
		Resources: p.config.Resources,
		Network:   p.config.Network,
		Mounts:    p.config.Mounts,
	})
	if err != nil {
		return err
	}

	p.sandboxes[sb.ID] = &poolEntry{
		sandbox: sb,
		inUse:   false,
	}

	return nil
}

// List returns all sandboxes in the pool.
func (p *Pool) List() []providers.Sandbox {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]providers.Sandbox, 0, len(p.sandboxes))
	for _, entry := range p.sandboxes {
		result = append(result, entry.sandbox)
	}
	return result
}
