package server

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

func newTestSandboxManager(mock *sandbox.MockProvider) *SandboxManager {
	return &SandboxManager{
		provider:   mock,
		logger:     slog.Default(),
		keepOnStop: true,
		stopCh:     make(chan struct{}),
	}
}

func TestSandboxManager_StartStop(t *testing.T) {
	// Use mock provider for testing
	mock := sandbox.NewMockProvider()
	mgr := newTestSandboxManager(mock)

	ctx := context.Background()

	// Start manager
	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Should have created a sandbox
	if len(mock.CreateCalls) != 1 {
		t.Errorf("Expected 1 create call, got %d", len(mock.CreateCalls))
	}

	// Verify sandbox name
	if mock.CreateCalls[0].Name != PersistentSandboxName {
		t.Errorf("Expected sandbox name %q, got %q", PersistentSandboxName, mock.CreateCalls[0].Name)
	}

	// Sandbox ID should be set
	if mgr.SandboxID() == "" {
		t.Error("SandboxID should be set after Start")
	}

	// Stop manager
	if err := mgr.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestSandboxManager_ReusesExistingSandbox(t *testing.T) {
	mock := sandbox.NewMockProvider()
	ctx := context.Background()
	
	// Pre-create a sandbox with the persistent name
	existing, _ := mock.Create(ctx, providers.SandboxCreateOptions{
		Name: PersistentSandboxName,
	})

	// Reset create calls
	mock.CreateCalls = nil

	mgr := newTestSandboxManager(mock)

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer mgr.Stop(ctx)

	// Should not have created a new sandbox (reused existing)
	if len(mock.CreateCalls) != 0 {
		t.Errorf("Expected 0 create calls (reuse existing), got %d", len(mock.CreateCalls))
	}

	// Should use the existing sandbox ID
	if mgr.SandboxID() != existing.ID {
		t.Errorf("Expected sandbox ID %q, got %q", existing.ID, mgr.SandboxID())
	}
}

func TestSandboxManager_IsHealthy(t *testing.T) {
	mock := sandbox.NewMockProvider()
	mgr := newTestSandboxManager(mock)

	ctx := context.Background()

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer mgr.Stop(ctx)

	// Should be healthy
	if !mgr.IsHealthy(ctx) {
		t.Error("Expected sandbox to be healthy")
	}
}

func TestSandboxManager_GetStatus(t *testing.T) {
	mock := sandbox.NewMockProvider()
	mgr := newTestSandboxManager(mock)

	ctx := context.Background()

	// Status before start
	status := mgr.GetStatus(ctx)
	if status.ID != "" {
		t.Errorf("Expected empty ID before start, got %q", status.ID)
	}

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer mgr.Stop(ctx)

	// Status after start
	status = mgr.GetStatus(ctx)
	if status.ID == "" {
		t.Error("Expected non-empty ID after start")
	}
	if status.Name != PersistentSandboxName {
		t.Errorf("Expected name %q, got %q", PersistentSandboxName, status.Name)
	}
	if !status.Running {
		t.Error("Expected sandbox to be running")
	}
	if !status.Healthy {
		t.Error("Expected sandbox to be healthy")
	}
}

func TestNewSandboxManager_DefaultsToAvailableProvider(t *testing.T) {
	mgr := NewSandboxManager(SandboxManagerConfig{
		KeepOnStop: true,
	})

	// Should have a provider
	if mgr.Provider() == nil {
		t.Error("Expected provider to be set")
	}

	// Provider name should be set
	name := mgr.Provider().Name()
	if name != "apple-container" && name != "none" {
		t.Errorf("Expected provider name to be 'apple-container' or 'none', got %q", name)
	}
}

func TestSandboxManager_MountsConfigured(t *testing.T) {
	mock := sandbox.NewMockProvider()
	mgr := newTestSandboxManager(mock)

	ctx := context.Background()

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer mgr.Stop(ctx)

	// Check that mounts were configured
	if len(mock.CreateCalls) != 1 {
		t.Fatalf("Expected 1 create call, got %d", len(mock.CreateCalls))
	}

	mounts := mock.CreateCalls[0].Mounts
	if len(mounts) < 4 {
		t.Errorf("Expected at least 4 mounts, got %d", len(mounts))
	}

	// Check for required mount destinations
	destinations := make(map[string]bool)
	for _, m := range mounts {
		destinations[m.Destination] = true
	}

	required := []string{"/home", "/shared", "/workspaces", "/run/ayo"}
	for _, dest := range required {
		if !destinations[dest] {
			t.Errorf("Expected mount for %q", dest)
		}
	}
}

func TestSandboxManager_ConcurrentAccess(t *testing.T) {
	mock := sandbox.NewMockProvider()
	mgr := newTestSandboxManager(mock)

	ctx := context.Background()

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer mgr.Stop(ctx)

	// Concurrent access should be safe
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = mgr.SandboxID()
			_ = mgr.IsHealthy(ctx)
			_ = mgr.GetStatus(ctx)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent access")
		}
	}
}
