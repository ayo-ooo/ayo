package sandbox

import (
	"context"
	"runtime"
	"testing"

	"github.com/alexcabrera/ayo/internal/providers"
)

// Provider interface compliance tests
func TestNoneProvider_ImplementsInterface(t *testing.T) {
	var _ providers.SandboxProvider = (*NoneProvider)(nil)
}

func TestAppleProvider_ImplementsInterface(t *testing.T) {
	var _ providers.SandboxProvider = (*AppleProvider)(nil)
}

func TestLinuxProvider_ImplementsInterface(t *testing.T) {
	var _ providers.SandboxProvider = (*LinuxProvider)(nil)
}

func TestMockProvider_ImplementsInterface(t *testing.T) {
	var _ providers.SandboxProvider = (*MockProvider)(nil)
}

// Provider selection tests - we can't directly test selectSandboxProvider
// since it's in daemon package, but we can test the individual providers

func TestProviderSelection_NoneAlwaysAvailable(t *testing.T) {
	p := NewNoneProvider()
	if err := p.Init(context.Background(), nil); err != nil {
		t.Errorf("NoneProvider.Init failed: %v", err)
	}
	// NoneProvider is always available (fallback provider)
	// Verify it can create sandboxes successfully
	sb, err := p.Create(context.Background(), providers.SandboxCreateOptions{})
	if err != nil {
		t.Errorf("NoneProvider.Create failed: %v", err)
	}
	if sb.ID == "" {
		t.Error("NoneProvider should create sandboxes with valid IDs")
	}
}

func TestProviderSelection_AppleRequiresDarwin(t *testing.T) {
	p := NewAppleProvider()
	available := p.IsAvailable()

	if runtime.GOOS != "darwin" && available {
		t.Error("AppleProvider should not be available on non-Darwin")
	}
	if runtime.GOOS == "darwin" && runtime.GOARCH != "arm64" && available {
		t.Error("AppleProvider should not be available on non-ARM64 Darwin")
	}
}

func TestProviderSelection_LinuxRequiresLinux(t *testing.T) {
	p := NewLinuxProvider()
	available := p.IsAvailable()

	if runtime.GOOS != "linux" && available {
		t.Error("LinuxProvider should not be available on non-Linux")
	}
}

// Provider unavailability tests
func TestAppleProvider_UnavailableOnNonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("This test is for non-Darwin platforms")
	}

	p := NewAppleProvider()
	if p.IsAvailable() {
		t.Error("AppleProvider should not be available on non-Darwin")
	}
}

func TestLinuxProvider_UnavailableOnNonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("This test is for non-Linux platforms")
	}

	p := NewLinuxProvider()
	if p.IsAvailable() {
		t.Error("LinuxProvider should not be available on non-Linux")
	}
}

// Mock provider comprehensive tests
func TestMockProvider_ExecSuccess(t *testing.T) {
	p := NewMockProvider()
	p.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		return providers.ExecResult{
			Stdout:   "hello\n",
			ExitCode: 0,
		}, nil
	}

	ctx := context.Background()
	sb, err := p.Create(ctx, providers.SandboxCreateOptions{})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	result, err := p.Exec(ctx, sb.ID, providers.ExecOptions{Command: "echo hello"})
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("got %q, want %q", result.Stdout, "hello\n")
	}
	if result.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", result.ExitCode)
	}
}

func TestMockProvider_ExecFailure(t *testing.T) {
	p := NewMockProvider()
	p.FailExec = true

	ctx := context.Background()
	sb, err := p.Create(ctx, providers.SandboxCreateOptions{})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	result, err := p.Exec(ctx, sb.ID, providers.ExecOptions{Command: "fail"})
	if err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}
	// FailExec returns non-zero exit code, not an error
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code from FailExec")
	}
	if result.Stderr != "mock error" {
		t.Errorf("got stderr %q, want %q", result.Stderr, "mock error")
	}
}

func TestMockProvider_CreateFailure(t *testing.T) {
	p := NewMockProvider()
	p.FailCreate = true

	ctx := context.Background()
	_, err := p.Create(ctx, providers.SandboxCreateOptions{})
	if err == nil {
		t.Error("expected error from FailCreate")
	}
}

// Pool tests with mock provider
func TestPool_WithMockProvider(t *testing.T) {
	mock := NewMockProvider()
	pool := NewPool(PoolConfig{
		Name:    "test-pool",
		MinSize: 1,
		MaxSize: 4,
	}, mock)

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer pool.Stop(ctx)

	sb, err := pool.Acquire(ctx, "@ayo")
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}
	if sb.ID == "" {
		t.Error("acquired sandbox should have ID")
	}

	// Release the sandbox
	if err := pool.Release(ctx, sb.ID); err != nil {
		t.Errorf("Release failed: %v", err)
	}
}

func TestPool_AcquireReleaseCycle(t *testing.T) {
	mock := NewMockProvider()
	pool := NewPool(PoolConfig{
		Name:    "test-pool",
		MinSize: 2,
		MaxSize: 4,
	}, mock)

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer pool.Stop(ctx)

	// Acquire multiple sandboxes
	var sandboxes []providers.Sandbox
	for i := 0; i < 3; i++ {
		sb, err := pool.Acquire(ctx, "@ayo")
		if err != nil {
			t.Fatalf("Acquire %d failed: %v", i, err)
		}
		sandboxes = append(sandboxes, sb)
	}

	// Release all
	for _, sb := range sandboxes {
		if err := pool.Release(ctx, sb.ID); err != nil {
			t.Errorf("Release %s failed: %v", sb.ID, err)
		}
	}

	// Check pool status
	status := pool.Status()
	if status.Total == 0 {
		t.Error("pool should have sandboxes after release")
	}
}

func TestPool_ExecViaMock(t *testing.T) {
	mock := NewMockProvider()
	mock.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		return providers.ExecResult{
			Stdout:   "mock output",
			ExitCode: 0,
		}, nil
	}

	pool := NewPool(PoolConfig{
		Name:    "test-pool",
		MinSize: 1,
		MaxSize: 2,
	}, mock)

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer pool.Stop(ctx)

	sb, err := pool.Acquire(ctx, "@ayo")
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}
	defer pool.Release(ctx, sb.ID)

	result, err := pool.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo test",
	})
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	if result.Stdout != "mock output" {
		t.Errorf("got %q, want %q", result.Stdout, "mock output")
	}
}

// Provider name tests
func TestProviderNames(t *testing.T) {
	tests := []struct {
		provider providers.SandboxProvider
		wantName string
	}{
		{NewNoneProvider(), "none"},
		{NewAppleProvider(), "apple-container"},
		{NewLinuxProvider(), "systemd-nspawn"},
		{NewMockProvider(), "mock"},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			if tt.provider.Name() != tt.wantName {
				t.Errorf("Name() = %v, want %v", tt.provider.Name(), tt.wantName)
			}
			if tt.provider.Type() != providers.ProviderTypeSandbox {
				t.Errorf("Type() = %v, want sandbox", tt.provider.Type())
			}
		})
	}
}

// Provider close tests
func TestProviders_Close(t *testing.T) {
	providers := []providers.SandboxProvider{
		NewNoneProvider(),
		NewAppleProvider(),
		NewLinuxProvider(),
		NewMockProvider(),
	}

	for _, p := range providers {
		if err := p.Close(); err != nil {
			t.Errorf("%s.Close() error = %v", p.Name(), err)
		}
	}
}
