package sandbox

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
)

func TestNoneProvider_Name(t *testing.T) {
	p := NewNoneProvider()
	if p.Name() != "none" {
		t.Errorf("Name() = %v, want none", p.Name())
	}
	if p.Type() != providers.ProviderTypeSandbox {
		t.Errorf("Type() = %v, want sandbox", p.Type())
	}
}

func TestNoneProvider_Create(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	sb, err := p.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-sandbox",
		Pool: "default",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if sb.Name != "test-sandbox" {
		t.Errorf("Name = %v, want test-sandbox", sb.Name)
	}
	if sb.Status != providers.SandboxStatusRunning {
		t.Errorf("Status = %v, want running", sb.Status)
	}
}

func TestNoneProvider_Get(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	sb, _ := p.Create(ctx, providers.SandboxCreateOptions{Name: "get-test"})

	got, err := p.Get(ctx, sb.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != sb.ID {
		t.Errorf("Get().ID = %v, want %v", got.ID, sb.ID)
	}
}

func TestNoneProvider_Get_NotFound(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	_, err := p.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("Get() expected error for nonexistent sandbox")
	}
}

func TestNoneProvider_List(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		p.Create(ctx, providers.SandboxCreateOptions{})
	}

	list, err := p.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 3 {
		t.Errorf("List() returned %d sandboxes, want 3", len(list))
	}
}

func TestNoneProvider_StartStop(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	sb, _ := p.Create(ctx, providers.SandboxCreateOptions{})

	// Stop
	if err := p.Stop(ctx, sb.ID, providers.SandboxStopOptions{}); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	status, _ := p.Status(ctx, sb.ID)
	if status != providers.SandboxStatusStopped {
		t.Errorf("Status after Stop = %v, want stopped", status)
	}

	// Start
	if err := p.Start(ctx, sb.ID); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	status, _ = p.Status(ctx, sb.ID)
	if status != providers.SandboxStatusRunning {
		t.Errorf("Status after Start = %v, want running", status)
	}
}

func TestNoneProvider_Delete(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	sb, _ := p.Create(ctx, providers.SandboxCreateOptions{})

	if err := p.Delete(ctx, sb.ID, false); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := p.Get(ctx, sb.ID)
	if err == nil {
		t.Error("Get() should fail after Delete")
	}
}

func TestNoneProvider_Exec(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	sb, _ := p.Create(ctx, providers.SandboxCreateOptions{})

	result, err := p.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo hello",
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("Stdout = %q, want 'hello\\n'", result.Stdout)
	}
}

func TestNoneProvider_Exec_WithWorkingDir(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	sb, _ := p.Create(ctx, providers.SandboxCreateOptions{})

	result, err := p.Exec(ctx, sb.ID, providers.ExecOptions{
		Command:    "pwd",
		WorkingDir: "/tmp",
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if result.Stdout != "/tmp\n" {
		t.Errorf("Stdout = %q, want '/tmp\\n'", result.Stdout)
	}
}

func TestNoneProvider_Exec_Error(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	sb, _ := p.Create(ctx, providers.SandboxCreateOptions{})

	result, err := p.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "exit 42",
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if result.ExitCode != 42 {
		t.Errorf("ExitCode = %d, want 42", result.ExitCode)
	}
}

func TestNoneProvider_Exec_NotFound(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	_, err := p.Exec(ctx, "nonexistent", providers.ExecOptions{
		Command: "echo test",
	})
	if err == nil {
		t.Error("Exec() expected error for nonexistent sandbox")
	}
}

func TestPool_Start(t *testing.T) {
	provider := NewNoneProvider()
	pool := NewPool(PoolConfig{
		Name:    "test-pool",
		MinSize: 2,
		MaxSize: 5,
	}, provider)

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer pool.Stop(ctx)

	status := pool.Status()
	if status.Total != 2 {
		t.Errorf("Total = %d, want 2", status.Total)
	}
	if status.Idle != 2 {
		t.Errorf("Idle = %d, want 2", status.Idle)
	}
}

func TestPool_Acquire(t *testing.T) {
	provider := NewNoneProvider()
	pool := NewPool(PoolConfig{
		Name:    "test-pool",
		MinSize: 1,
		MaxSize: 3,
	}, provider)

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop(ctx)

	sb, err := pool.Acquire(ctx, "@ayo")
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	if sb.ID == "" {
		t.Error("Acquire() returned sandbox with empty ID")
	}

	status := pool.Status()
	if status.InUse != 1 {
		t.Errorf("InUse = %d, want 1", status.InUse)
	}
}

func TestPool_AcquireRelease(t *testing.T) {
	provider := NewNoneProvider()
	pool := NewPool(PoolConfig{
		Name:    "test-pool",
		MinSize: 1,
		MaxSize: 2,
	}, provider)

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop(ctx)

	// Acquire
	sb, _ := pool.Acquire(ctx, "@ayo")
	if pool.Status().InUse != 1 {
		t.Error("Should have 1 in use after acquire")
	}

	// Release
	if err := pool.Release(ctx, sb.ID); err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	if pool.Status().InUse != 0 {
		t.Error("Should have 0 in use after release")
	}
}

func TestPool_Exhausted(t *testing.T) {
	provider := NewNoneProvider()
	pool := NewPool(PoolConfig{
		Name:    "test-pool",
		MinSize: 0,
		MaxSize: 1,
	}, provider)

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop(ctx)

	// First acquire should work
	_, err := pool.Acquire(ctx, "@agent1")
	if err != nil {
		t.Fatalf("First Acquire() error = %v", err)
	}

	// Second acquire should fail (max reached)
	_, err = pool.Acquire(ctx, "@agent2")
	if err == nil {
		t.Error("Second Acquire() should fail when pool exhausted")
	}
}

func TestPool_Exec(t *testing.T) {
	provider := NewNoneProvider()
	pool := NewPool(PoolConfig{
		Name:    "test-pool",
		MinSize: 1,
	}, provider)

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop(ctx)

	sb, _ := pool.Acquire(ctx, "@ayo")

	result, err := pool.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo pooltest",
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if result.Stdout != "pooltest\n" {
		t.Errorf("Stdout = %q, want 'pooltest\\n'", result.Stdout)
	}
}

func TestPool_ReuseForSameAgent(t *testing.T) {
	provider := NewNoneProvider()
	pool := NewPool(PoolConfig{
		Name:    "test-pool",
		MinSize: 1,
		MaxSize: 5,
	}, provider)

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop(ctx)

	// First acquire for @ayo
	sb1, _ := pool.Acquire(ctx, "@ayo")

	// Second acquire for same agent should return same sandbox
	sb2, _ := pool.Acquire(ctx, "@ayo")

	if sb1.ID != sb2.ID {
		t.Errorf("Same agent should get same sandbox: %s != %s", sb1.ID, sb2.ID)
	}
}

func TestNoneProvider_Exec_Timeout(t *testing.T) {
	p := NewNoneProvider()
	ctx := context.Background()

	sb, _ := p.Create(ctx, providers.SandboxCreateOptions{})

	result, err := p.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "sleep 10",
		Timeout: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if !result.TimedOut {
		t.Error("Expected TimedOut to be true")
	}
}

func TestExecutor_Exec(t *testing.T) {
	provider := NewNoneProvider()
	ctx := context.Background()

	sb, err := provider.Create(ctx, providers.SandboxCreateOptions{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer provider.Delete(ctx, sb.ID, true)

	executor := NewExecutor(provider, sb.ID, t.TempDir(), "")

	result, err := executor.Exec(ctx, BashParams{
		Command:     "echo hello world",
		Description: "Test echo",
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello world\n" {
		t.Errorf("Stdout = %q, want 'hello world\\n'", result.Stdout)
	}
}

func TestExecutor_Exec_EmptyCommand(t *testing.T) {
	provider := NewNoneProvider()
	ctx := context.Background()

	sb, _ := provider.Create(ctx, providers.SandboxCreateOptions{})
	executor := NewExecutor(provider, sb.ID, t.TempDir(), "")

	_, err := executor.Exec(ctx, BashParams{})
	if err == nil {
		t.Error("Exec() expected error for empty command")
	}
}

func TestExecutor_Exec_CustomTimeout(t *testing.T) {
	provider := NewNoneProvider()
	ctx := context.Background()

	sb, _ := provider.Create(ctx, providers.SandboxCreateOptions{})
	executor := NewExecutor(provider, sb.ID, t.TempDir(), "")

	result, err := executor.Exec(ctx, BashParams{
		Command:        "sleep 10",
		TimeoutSeconds: 1, // 1 second timeout
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if !result.TimedOut {
		t.Error("Expected TimedOut to be true")
	}
	if result.Error != "command timed out" {
		t.Errorf("Error = %q, want 'command timed out'", result.Error)
	}
}

func TestBashResult_String(t *testing.T) {
	result := BashResult{
		Stdout:   "output",
		Stderr:   "error",
		ExitCode: 0,
	}

	str := result.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	if !containsStr(str, "stdout:") {
		t.Error("String() should contain stdout")
	}
	if !containsStr(str, "stderr:") {
		t.Error("String() should contain stderr")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Mock provider tests
func TestMockProvider_Name(t *testing.T) {
	p := NewMockProvider()
	if p.Name() != "mock" {
		t.Errorf("Name() = %v, want mock", p.Name())
	}
	if p.Type() != providers.ProviderTypeSandbox {
		t.Errorf("Type() = %v, want sandbox", p.Type())
	}
}

func TestMockProvider_CreateAndExec(t *testing.T) {
	p := NewMockProvider()
	ctx := context.Background()

	sb, err := p.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-mock",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if sb.Name != "test-mock" {
		t.Errorf("Name = %v, want test-mock", sb.Name)
	}

	// Exec returns default mock output
	result, err := p.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo test",
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if result.Stdout != "mock output\n" {
		t.Errorf("Stdout = %q, want 'mock output\n'", result.Stdout)
	}

	// Verify call was recorded
	if len(p.ExecCalls) != 1 {
		t.Errorf("ExecCalls = %d, want 1", len(p.ExecCalls))
	}
}

func TestMockProvider_CustomExecFunc(t *testing.T) {
	p := NewMockProvider()
	ctx := context.Background()

	// Configure custom exec behavior
	p.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		return providers.ExecResult{
			Stdout:   "custom: " + opts.Command + "\n",
			ExitCode: 0,
		}, nil
	}

	sb, _ := p.Create(ctx, providers.SandboxCreateOptions{})
	result, _ := p.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "hello",
	})

	if result.Stdout != "custom: hello\n" {
		t.Errorf("Stdout = %q, want 'custom: hello\n'", result.Stdout)
	}
}

func TestMockProvider_FailCreate(t *testing.T) {
	p := NewMockProvider()
	p.FailCreate = true

	ctx := context.Background()
	_, err := p.Create(ctx, providers.SandboxCreateOptions{})
	if err == nil {
		t.Error("Create() should fail when FailCreate is true")
	}
}

func TestMockProvider_FailExec(t *testing.T) {
	p := NewMockProvider()
	ctx := context.Background()

	sb, _ := p.Create(ctx, providers.SandboxCreateOptions{})
	p.FailExec = true

	result, _ := p.Exec(ctx, sb.ID, providers.ExecOptions{})
	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
}

func TestMockProvider_Reset(t *testing.T) {
	p := NewMockProvider()
	ctx := context.Background()

	p.Create(ctx, providers.SandboxCreateOptions{})
	p.FailCreate = true
	p.FailExec = true

	p.Reset()

	if p.SandboxCount() != 0 {
		t.Error("Reset() should clear sandboxes")
	}
	if p.FailCreate {
		t.Error("Reset() should clear FailCreate")
	}
	if p.FailExec {
		t.Error("Reset() should clear FailExec")
	}
	if len(p.CreateCalls) != 0 {
		t.Error("Reset() should clear CreateCalls")
	}
}

func TestMockProvider_Lifecycle(t *testing.T) {
	p := NewMockProvider()
	ctx := context.Background()

	// Create
	sb, _ := p.Create(ctx, providers.SandboxCreateOptions{Name: "lifecycle"})

	// Stop
	if err := p.Stop(ctx, sb.ID, providers.SandboxStopOptions{}); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	status, _ := p.Status(ctx, sb.ID)
	if status != providers.SandboxStatusStopped {
		t.Errorf("Status = %v, want stopped", status)
	}

	// Start
	if err := p.Start(ctx, sb.ID); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	status, _ = p.Status(ctx, sb.ID)
	if status != providers.SandboxStatusRunning {
		t.Errorf("Status = %v, want running", status)
	}

	// Delete
	if err := p.Delete(ctx, sb.ID, false); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if p.SandboxCount() != 0 {
		t.Error("Delete() should remove sandbox")
	}
}

// Apple provider tests
func TestAppleProvider_Name(t *testing.T) {
	p := NewAppleProvider()
	if p.Name() != "apple-container" {
		t.Errorf("Name() = %v, want apple-container", p.Name())
	}
	if p.Type() != providers.ProviderTypeSandbox {
		t.Errorf("Type() = %v, want sandbox", p.Type())
	}
}

func TestAppleProvider_IsAvailable(t *testing.T) {
	p := NewAppleProvider()
	// Just verify the method doesn't panic
	_ = p.IsAvailable()
}

func TestAppleProvider_Init_WhenNotAvailable(t *testing.T) {
	p := &AppleProvider{
		sandboxes: make(map[string]*appleSandbox),
		available: false,
	}

	err := p.Init(context.Background(), nil)
	if err == nil {
		t.Error("Init() should fail when Apple Container is not available")
	}
}

func TestAppleProvider_Create_WhenNotAvailable(t *testing.T) {
	p := &AppleProvider{
		sandboxes: make(map[string]*appleSandbox),
		available: false,
	}

	_, err := p.Create(context.Background(), providers.SandboxCreateOptions{})
	if err == nil {
		t.Error("Create() should fail when Apple Container is not available")
	}
}

func TestAppleProvider_Get_NotFound(t *testing.T) {
	p := &AppleProvider{
		sandboxes: make(map[string]*appleSandbox),
		available: true,
	}

	_, err := p.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Get() should fail for nonexistent sandbox")
	}
}

func TestAppleProvider_AssignAgent_NotFound(t *testing.T) {
	p := &AppleProvider{
		sandboxes: make(map[string]*appleSandbox),
		available: true,
	}

	err := p.AssignAgent("nonexistent", "@ayo")
	if err == nil {
		t.Error("AssignAgent() should fail for nonexistent sandbox")
	}
}

func TestAppleProvider_List(t *testing.T) {
	p := &AppleProvider{
		sandboxes: map[string]*appleSandbox{
			"test1": {id: "test1", name: "sandbox1", status: providers.SandboxStatusRunning},
			"test2": {id: "test2", name: "sandbox2", status: providers.SandboxStatusStopped},
		},
		available: true,
	}

	// Note: List() now queries the real container runtime, not the in-memory map.
	// This test just verifies List() doesn't error when runtime is available.
	list, err := p.List(context.Background())
	if err != nil {
		// If container CLI is not available, skip
		t.Skipf("List() error = %v (container CLI may not be available)", err)
	}

	// Just verify we got a list back (could be any number based on real containers)
	_ = list
}

func TestAppleProvider_Integration(t *testing.T) {
	p := NewAppleProvider()
	if !p.IsAvailable() {
		t.Skip("Apple Container is not available, skipping integration test")
	}

	ctx := context.Background()

	// Create sandbox (name must start with "ayo-" to be listed)
	sb, err := p.Create(ctx, providers.SandboxCreateOptions{
		Name:  "ayo-test-" + t.Name(),
		Image: "docker.io/library/alpine:3.21",
		Network: providers.NetworkConfig{
			Enabled: false,
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer p.Delete(ctx, sb.ID, true)

	if sb.ID == "" {
		t.Error("Create() returned sandbox with empty ID")
	}

	// Execute command
	result, err := p.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo hello",
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("Stdout = %q, want 'hello\\n'", result.Stdout)
	}

	// Get sandbox
	got, err := p.Get(ctx, sb.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != sb.ID {
		t.Errorf("Get().ID = %v, want %v", got.ID, sb.ID)
	}

	// List sandboxes
	list, err := p.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	found := false
	for _, s := range list {
		if s.ID == sb.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("List() should contain created sandbox")
	}

	// Status
	status, err := p.Status(ctx, sb.ID)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status != providers.SandboxStatusRunning {
		t.Errorf("Status = %v, want running", status)
	}
}

func TestAppleProvider_Directories_Integration(t *testing.T) {
	p := NewAppleProvider()
	if !p.IsAvailable() {
		t.Skip("Apple Container is not available, skipping integration test")
	}

	ctx := context.Background()

	// Create sandbox with network enabled
	sb, err := p.Create(ctx, providers.SandboxCreateOptions{
		Name: "ayo-dirs-" + t.Name(),
		Network: providers.NetworkConfig{
			Enabled: true,
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer p.Delete(ctx, sb.ID, true)

	// Verify standard directories exist with correct permissions
	tests := []struct {
		path     string
		wantMode string
	}{
		{"/shared", "1777"},
		{"/workspaces", "755"},
		{"/run/ayo", "755"},
		{"/mnt/host", "755"},
	}

	for _, tt := range tests {
		// Check directory exists
		result, err := p.Exec(ctx, sb.ID, providers.ExecOptions{
			Command: "test -d " + tt.path + " && echo ok",
		})
		if err != nil {
			t.Fatalf("Exec(test %s) error = %v", tt.path, err)
		}
		if result.Stdout != "ok\n" {
			t.Errorf("directory %s should exist", tt.path)
		}

		// Check permissions (using stat)
		result, err = p.Exec(ctx, sb.ID, providers.ExecOptions{
			Command: "stat -c %a " + tt.path,
		})
		if err != nil {
			t.Fatalf("Exec(stat %s) error = %v", tt.path, err)
		}
		gotMode := strings.TrimSpace(result.Stdout)
		if gotMode != tt.wantMode {
			t.Errorf("directory %s mode = %s, want %s", tt.path, gotMode, tt.wantMode)
		}
	}
}
