package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	stdsync "sync"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
	"github.com/alexcabrera/ayo/internal/sandbox/mounts"
	"github.com/alexcabrera/ayo/internal/sync"
)

const (
	// PersistentSandboxName is the fixed name for the persistent sandbox.
	PersistentSandboxName = "ayo-sandbox"

	// HealthCheckInterval is how often to check sandbox health.
	HealthCheckInterval = 30 * time.Second

	// HealthCheckTimeout is the maximum time to wait for a health check.
	HealthCheckTimeout = 10 * time.Second
)

// SandboxManager manages the persistent sandbox lifecycle.
type SandboxManager struct {
	provider providers.SandboxProvider
	logger   *slog.Logger

	mu        stdsync.RWMutex
	sandboxID string
	running   bool
	stopCh    chan struct{}
	wg        stdsync.WaitGroup

	// Configuration
	keepOnStop bool // Whether to keep sandbox running when daemon stops
}

// SandboxManagerConfig configures the sandbox manager.
type SandboxManagerConfig struct {
	// KeepOnStop keeps the sandbox running when the daemon stops.
	// Default: true (sandbox persists across daemon restarts)
	KeepOnStop bool

	// Logger for sandbox manager operations.
	Logger *slog.Logger
}

// NewSandboxManager creates a new sandbox manager.
func NewSandboxManager(cfg SandboxManagerConfig) *SandboxManager {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	// Use Apple Container provider by default, fall back to none
	var provider providers.SandboxProvider
	apple := sandbox.NewAppleProvider()
	if apple.IsAvailable() {
		provider = apple
	} else {
		provider = sandbox.NewNoneProvider()
	}

	return &SandboxManager{
		provider:   provider,
		logger:     cfg.Logger,
		keepOnStop: cfg.KeepOnStop,
		stopCh:     make(chan struct{}),
	}
}

// Start initializes the persistent sandbox and begins health monitoring.
func (m *SandboxManager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("sandbox manager already running")
	}
	m.running = true
	m.mu.Unlock()

	// Ensure sandbox state directories exist on host
	if err := m.ensureHostDirectories(); err != nil {
		m.logger.Warn("failed to create host directories", "error", err)
	}

	// Ensure persistent sandbox exists and is running
	if err := m.ensurePersistentSandbox(ctx); err != nil {
		return fmt.Errorf("ensure persistent sandbox: %w", err)
	}

	// Start health check goroutine
	m.wg.Add(1)
	go m.healthCheckLoop()

	m.logger.Info("sandbox manager started", "sandbox", m.sandboxID)
	return nil
}

// Stop stops the sandbox manager and optionally the sandbox.
func (m *SandboxManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = false
	close(m.stopCh)
	m.mu.Unlock()

	// Wait for health check to stop
	m.wg.Wait()

	// Optionally stop the sandbox
	if !m.keepOnStop && m.sandboxID != "" {
		m.logger.Info("stopping persistent sandbox", "sandbox", m.sandboxID)
		if err := m.provider.Stop(ctx, m.sandboxID, providers.SandboxStopOptions{
			Timeout: 30 * time.Second,
		}); err != nil {
			m.logger.Warn("failed to stop sandbox", "error", err)
		}
	}

	m.logger.Info("sandbox manager stopped")
	return nil
}

// SandboxID returns the ID of the persistent sandbox.
func (m *SandboxManager) SandboxID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sandboxID
}

// Provider returns the sandbox provider.
func (m *SandboxManager) Provider() providers.SandboxProvider {
	return m.provider
}

// IsHealthy returns true if the sandbox is running and responsive.
func (m *SandboxManager) IsHealthy(ctx context.Context) bool {
	m.mu.RLock()
	id := m.sandboxID
	m.mu.RUnlock()

	if id == "" {
		return false
	}

	status, err := m.provider.Status(ctx, id)
	if err != nil {
		return false
	}

	return status == providers.SandboxStatusRunning
}

// ensureHostDirectories creates the host-side directories for sandbox mounts.
func (m *SandboxManager) ensureHostDirectories() error {
	dirs := []string{
		sync.HomesDir(),
		sync.SharedDir(),
		sync.SandboxDir() + "/workspaces",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}

	return nil
}

// ensurePersistentSandbox ensures the persistent sandbox exists and is running.
func (m *SandboxManager) ensurePersistentSandbox(ctx context.Context) error {
	// Check if sandbox already exists
	existing, err := m.findExistingSandbox(ctx)
	if err != nil {
		m.logger.Debug("no existing sandbox found", "error", err)
	}

	if existing != "" {
		// Check if it's running
		status, err := m.provider.Status(ctx, existing)
		if err != nil {
			m.logger.Warn("failed to get sandbox status", "sandbox", existing, "error", err)
		} else if status == providers.SandboxStatusRunning {
			m.mu.Lock()
			m.sandboxID = existing
			m.mu.Unlock()
			m.logger.Info("using existing sandbox", "sandbox", existing)
			return nil
		} else if status == providers.SandboxStatusStopped {
			// Try to start it
			m.logger.Info("starting stopped sandbox", "sandbox", existing)
			if err := m.provider.Start(ctx, existing); err != nil {
				m.logger.Warn("failed to start existing sandbox, will create new", "error", err)
			} else {
				m.mu.Lock()
				m.sandboxID = existing
				m.mu.Unlock()
				return nil
			}
		}
	}

	// Create new persistent sandbox
	return m.createPersistentSandbox(ctx)
}

// findExistingSandbox looks for an existing persistent sandbox.
func (m *SandboxManager) findExistingSandbox(ctx context.Context) (string, error) {
	sandboxes, err := m.provider.List(ctx)
	if err != nil {
		return "", err
	}

	for _, sb := range sandboxes {
		if sb.Name == PersistentSandboxName {
			return sb.ID, nil
		}
	}

	return "", fmt.Errorf("sandbox not found")
}

// createPersistentSandbox creates a new persistent sandbox with standard configuration.
func (m *SandboxManager) createPersistentSandbox(ctx context.Context) error {
	m.logger.Info("creating persistent sandbox")

	// Build mounts for host directories
	mountList := []providers.Mount{
		{
			Source:      sync.HomesDir(),
			Destination: "/home",
			Mode:        providers.MountModeBind,
			ReadOnly:    false,
		},
		{
			Source:      sync.SharedDir(),
			Destination: "/shared",
			Mode:        providers.MountModeBind,
			ReadOnly:    false,
		},
		{
			Source:      sync.SandboxDir() + "/workspaces",
			Destination: "/workspaces",
			Mode:        providers.MountModeBind,
			ReadOnly:    false,
		},
		// Matrix/daemon runtime directory (socket access)
		{
			Source:      paths.RuntimeDir(),
			Destination: "/run/ayo",
			Mode:        providers.MountModeBind,
			ReadOnly:    false,
		},
	}

	// Add mounts from persistent grants (ayo mount add)
	grants, grantsErr := mounts.LoadGrants()
	if grantsErr != nil {
		m.logger.Warn("could not load grants", "error", grantsErr)
	} else {
		grantMounts := grants.ToProviderMounts()
		m.logger.Info("loaded grants", "count", len(grantMounts))
		for _, gm := range grantMounts {
			m.logger.Info("adding grant mount", "source", gm.Source, "dest", gm.Destination, "readonly", gm.ReadOnly)
			mountList = append(mountList, providers.Mount{
				Source:      gm.Source,
				Destination: gm.Destination,
				Mode:        providers.MountModeBind,
				ReadOnly:    gm.ReadOnly,
			})
		}
	}

	sb, err := m.provider.Create(ctx, providers.SandboxCreateOptions{
		Name:   PersistentSandboxName,
		Image:  "docker.io/library/alpine:3.21",
		Mounts: mountList,
		Network: providers.NetworkConfig{
			Enabled: true,
		},
	})
	if err != nil {
		return fmt.Errorf("create sandbox: %w", err)
	}

	m.mu.Lock()
	m.sandboxID = sb.ID
	m.mu.Unlock()

	m.logger.Info("persistent sandbox created", "sandbox", sb.ID)
	return nil
}

// healthCheckLoop periodically checks sandbox health and restarts if needed.
func (m *SandboxManager) healthCheckLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck checks sandbox health and restarts if unhealthy.
func (m *SandboxManager) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), HealthCheckTimeout)
	defer cancel()

	m.mu.RLock()
	id := m.sandboxID
	m.mu.RUnlock()

	if id == "" {
		m.logger.Debug("no sandbox to health check")
		return
	}

	// Check status
	status, err := m.provider.Status(ctx, id)
	if err != nil {
		m.logger.Warn("health check failed to get status", "sandbox", id, "error", err)
		m.attemptRestart(ctx)
		return
	}

	if status != providers.SandboxStatusRunning {
		m.logger.Warn("sandbox not running", "sandbox", id, "status", status)
		m.attemptRestart(ctx)
		return
	}

	// Try to execute a simple command to verify responsiveness
	result, err := m.provider.Exec(ctx, id, providers.ExecOptions{
		Command: "echo ok",
		Timeout: 5 * time.Second,
	})
	if err != nil || result.ExitCode != 0 {
		m.logger.Warn("sandbox unresponsive", "sandbox", id, "error", err)
		m.attemptRestart(ctx)
		return
	}

	m.logger.Debug("sandbox healthy", "sandbox", id)
}

// attemptRestart tries to restart the sandbox.
func (m *SandboxManager) attemptRestart(ctx context.Context) {
	m.mu.RLock()
	id := m.sandboxID
	m.mu.RUnlock()

	m.logger.Info("attempting sandbox restart", "sandbox", id)

	// Try to start if stopped
	if err := m.provider.Start(ctx, id); err != nil {
		m.logger.Warn("failed to start sandbox", "error", err)

		// If start fails, try to recreate
		if err := m.createPersistentSandbox(ctx); err != nil {
			m.logger.Error("failed to recreate sandbox", "error", err)
		}
	} else {
		m.logger.Info("sandbox restarted", "sandbox", id)
	}
}

// Status returns the current sandbox status.
type SandboxStatus struct {
	ID        string                   `json:"id"`
	Name      string                   `json:"name"`
	Running   bool                     `json:"running"`
	Status    providers.SandboxStatus  `json:"status"`
	Provider  string                   `json:"provider"`
	Healthy   bool                     `json:"healthy"`
}

// GetStatus returns the current status of the persistent sandbox.
func (m *SandboxManager) GetStatus(ctx context.Context) SandboxStatus {
	m.mu.RLock()
	id := m.sandboxID
	m.mu.RUnlock()

	result := SandboxStatus{
		ID:       id,
		Name:     PersistentSandboxName,
		Provider: m.provider.Name(),
	}

	if id == "" {
		return result
	}

	status, err := m.provider.Status(ctx, id)
	if err == nil {
		result.Status = status
		result.Running = status == providers.SandboxStatusRunning
	}

	result.Healthy = m.IsHealthy(ctx)
	return result
}
