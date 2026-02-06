// Package sandbox provides container-based agent execution environments.
package sandbox

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
)

// ErrMockSandboxNotFound is returned when a mock sandbox is not found.
var ErrMockSandboxNotFound = errors.New("mock sandbox not found")

// MockProvider is a mock sandbox provider for testing.
// It simulates container behavior without actual containers.
type MockProvider struct {
	mu        sync.Mutex
	sandboxes map[string]*mockSandbox
	
	// Configurable behavior
	ExecFunc   func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error)
	CreateFunc func(ctx context.Context, opts providers.SandboxCreateOptions) (providers.Sandbox, error)
	StatsFunc  func(ctx context.Context, id string) (providers.SandboxStats, error)
	FailCreate bool
	FailExec   bool
	
	// Recording
	ExecCalls   []ExecCall
	CreateCalls []providers.SandboxCreateOptions
}

// ExecCall records a call to Exec.
type ExecCall struct {
	SandboxID string
	Options   providers.ExecOptions
}

type mockSandbox struct {
	id        string
	name      string
	status    providers.SandboxStatus
	createdAt time.Time
	agents    []string
	image     string
}

// NewMockProvider creates a new mock sandbox provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		sandboxes:   make(map[string]*mockSandbox),
		ExecCalls:   make([]ExecCall, 0),
		CreateCalls: make([]providers.SandboxCreateOptions, 0),
	}
}

func (p *MockProvider) Name() string                  { return "mock" }
func (p *MockProvider) Type() providers.ProviderType { return providers.ProviderTypeSandbox }
func (p *MockProvider) Init(_ context.Context, _ map[string]any) error { return nil }
func (p *MockProvider) Close() error { return nil }

// IsAvailable always returns true for mocks.
func (p *MockProvider) IsAvailable() bool { return true }

// Create creates a mock sandbox.
func (p *MockProvider) Create(ctx context.Context, opts providers.SandboxCreateOptions) (providers.Sandbox, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.CreateCalls = append(p.CreateCalls, opts)
	
	if p.FailCreate {
		return providers.Sandbox{}, context.Canceled
	}
	
	if p.CreateFunc != nil {
		return p.CreateFunc(ctx, opts)
	}
	
	id := generateMockID()
	name := opts.Name
	if name == "" {
		name = "mock-sandbox-" + id
	}
	
	sb := &mockSandbox{
		id:        id,
		name:      name,
		status:    providers.SandboxStatusRunning,
		createdAt: time.Now(),
		agents:    make([]string, 0),
		image:     opts.Image,
	}
	p.sandboxes[id] = sb
	
	return p.toProvidersSandbox(sb), nil
}

// Get retrieves a mock sandbox.
func (p *MockProvider) Get(ctx context.Context, id string) (providers.Sandbox, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	sb, ok := p.sandboxes[id]
	if !ok {
		return providers.Sandbox{}, ErrMockSandboxNotFound
	}
	return p.toProvidersSandbox(sb), nil
}

// List returns all mock sandboxes.
func (p *MockProvider) List(ctx context.Context) ([]providers.Sandbox, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	result := make([]providers.Sandbox, 0, len(p.sandboxes))
	for _, sb := range p.sandboxes {
		result = append(result, p.toProvidersSandbox(sb))
	}
	return result, nil
}

// Start starts a mock sandbox.
func (p *MockProvider) Start(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	sb, ok := p.sandboxes[id]
	if !ok {
		return ErrMockSandboxNotFound
	}
	sb.status = providers.SandboxStatusRunning
	return nil
}

// Stop stops a mock sandbox.
func (p *MockProvider) Stop(ctx context.Context, id string, opts providers.SandboxStopOptions) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	sb, ok := p.sandboxes[id]
	if !ok {
		return ErrMockSandboxNotFound
	}
	sb.status = providers.SandboxStatusStopped
	return nil
}

// Delete removes a mock sandbox.
func (p *MockProvider) Delete(ctx context.Context, id string, force bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if _, ok := p.sandboxes[id]; !ok {
		return ErrMockSandboxNotFound
	}
	delete(p.sandboxes, id)
	return nil
}

// Exec simulates command execution.
func (p *MockProvider) Exec(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
	p.mu.Lock()
	p.ExecCalls = append(p.ExecCalls, ExecCall{SandboxID: id, Options: opts})
	
	if p.FailExec {
		p.mu.Unlock()
		return providers.ExecResult{ExitCode: 1, Stderr: "mock error"}, nil
	}
	
	if p.ExecFunc != nil {
		p.mu.Unlock()
		return p.ExecFunc(ctx, id, opts)
	}
	p.mu.Unlock()
	
	// Default: simulate successful echo
	return providers.ExecResult{
		Stdout:   "mock output\n",
		ExitCode: 0,
		Duration: 10 * time.Millisecond,
	}, nil
}

// Status returns the status of a mock sandbox.
func (p *MockProvider) Status(ctx context.Context, id string) (providers.SandboxStatus, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	sb, ok := p.sandboxes[id]
	if !ok {
		return "", ErrMockSandboxNotFound
	}
	return sb.status, nil
}

// AssignAgent assigns an agent to a mock sandbox.
func (p *MockProvider) AssignAgent(id, agentHandle string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	sb, ok := p.sandboxes[id]
	if !ok {
		return ErrMockSandboxNotFound
	}
	sb.agents = append(sb.agents, agentHandle)
	return nil
}

// Stats returns resource usage statistics for a mock sandbox.
func (p *MockProvider) Stats(ctx context.Context, id string) (providers.SandboxStats, error) {
	p.mu.Lock()
	sb, ok := p.sandboxes[id]
	if !ok {
		p.mu.Unlock()
		return providers.SandboxStats{}, ErrMockSandboxNotFound
	}
	
	if p.StatsFunc != nil {
		p.mu.Unlock()
		return p.StatsFunc(ctx, id)
	}
	
	uptime := time.Since(sb.createdAt)
	p.mu.Unlock()
	
	// Return mock stats
	return providers.SandboxStats{
		CPUPercent:       5.0,
		MemoryUsageBytes: 50 * 1024 * 1024, // 50 MB
		MemoryLimitBytes: 512 * 1024 * 1024, // 512 MB
		ProcessCount:     3,
		Uptime:           uptime,
	}, nil
}

// Reset clears all recorded calls and sandboxes.
func (p *MockProvider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.sandboxes = make(map[string]*mockSandbox)
	p.ExecCalls = make([]ExecCall, 0)
	p.CreateCalls = make([]providers.SandboxCreateOptions, 0)
	p.FailCreate = false
	p.FailExec = false
	p.ExecFunc = nil
	p.CreateFunc = nil
	p.StatsFunc = nil
}

// SandboxCount returns the number of active sandboxes.
func (p *MockProvider) SandboxCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.sandboxes)
}

func (p *MockProvider) toProvidersSandbox(sb *mockSandbox) providers.Sandbox {
	return providers.Sandbox{
		ID:        sb.id,
		Name:      sb.name,
		Status:    sb.status,
		Image:     sb.image,
		Agents:    sb.agents,
		CreatedAt: sb.createdAt,
	}
}

var mockIDCounter int
var mockIDMu sync.Mutex

func generateMockID() string {
	mockIDMu.Lock()
	defer mockIDMu.Unlock()
	mockIDCounter++
	return string(rune('a'-1+mockIDCounter%26)) + "000" + string(rune('0'+mockIDCounter%10))
}

// EnsureAgentUser is a no-op for MockProvider.
func (p *MockProvider) EnsureAgentUser(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

// Verify MockProvider implements SandboxProvider.
var _ providers.SandboxProvider = (*MockProvider)(nil)
