package sandbox

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
)

func TestLinuxProvider_Name(t *testing.T) {
	p := NewLinuxProvider()
	if p.Name() != "systemd-nspawn" {
		t.Errorf("Name() = %v, want systemd-nspawn", p.Name())
	}
	if p.Type() != providers.ProviderTypeSandbox {
		t.Errorf("Type() = %v, want sandbox", p.Type())
	}
}

func TestLinuxProvider_IsAvailable(t *testing.T) {
	p := NewLinuxProvider()
	// Just verify the method doesn't panic
	_ = p.IsAvailable()
}

func TestLinuxProvider_Init_WhenNotAvailable(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Test only runs on non-Linux platforms")
	}

	p := NewLinuxProvider()
	err := p.Init(context.Background(), nil)
	if err == nil {
		t.Error("Init() should fail when not on Linux")
	}
}

func TestLinuxProvider_Create_WhenNotAvailable(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Test only runs on non-Linux platforms")
	}

	p := NewLinuxProvider()
	_, err := p.Create(context.Background(), providers.SandboxCreateOptions{
		Name: "test",
	})
	if err == nil {
		t.Error("Create() should fail when not on Linux")
	}
}

func TestLinuxProvider_Get_NotFound(t *testing.T) {
	p := NewLinuxProvider()
	_, err := p.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Get() should fail for nonexistent sandbox")
	}
}

func TestLinuxProvider_List_Empty(t *testing.T) {
	p := NewLinuxProvider()
	sandboxes, err := p.List(context.Background())
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(sandboxes) != 0 {
		t.Errorf("List() returned %d sandboxes, want 0", len(sandboxes))
	}
}

func TestLinuxProvider_Delete_NotFound(t *testing.T) {
	p := NewLinuxProvider()
	err := p.Delete(context.Background(), "nonexistent", false)
	if err == nil {
		t.Error("Delete() should fail for nonexistent sandbox")
	}
}

func TestLinuxProvider_Status_NotFound(t *testing.T) {
	p := NewLinuxProvider()
	_, err := p.Status(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Status() should fail for nonexistent sandbox")
	}
}

func TestLinuxProvider_AssignAgent_NotFound(t *testing.T) {
	p := NewLinuxProvider()
	err := p.AssignAgent("nonexistent", "@test")
	if err == nil {
		t.Error("AssignAgent() should fail for nonexistent sandbox")
	}
}

func TestLinuxProvider_Close(t *testing.T) {
	p := NewLinuxProvider()
	err := p.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// Integration tests that only run on Linux with systemd-nspawn available
func TestLinuxProvider_Integration(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux container tests require Linux")
	}

	p := NewLinuxProvider()
	if !p.IsAvailable() {
		t.Skip("systemd-nspawn not available")
	}

	ctx := context.Background()

	// Create sandbox
	sb, err := p.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-integration",
		Network: providers.NetworkConfig{
			Enabled: false,
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer p.Delete(ctx, sb.ID, true)

	// Check status
	status, err := p.Status(ctx, sb.ID)
	if err != nil {
		t.Errorf("Status() error = %v", err)
	}
	if status != providers.SandboxStatusRunning {
		t.Errorf("Status() = %v, want running", status)
	}

	// Execute command
	result, err := p.Exec(ctx, sb.ID, providers.ExecOptions{
		Command: "echo hello",
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Errorf("Exec() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("Exec() stdout = %q, want %q", result.Stdout, "hello\n")
	}

	// Stop sandbox
	err = p.Stop(ctx, sb.ID, providers.SandboxStopOptions{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// List sandboxes
	sandboxes, err := p.List(ctx)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(sandboxes) != 1 {
		t.Errorf("List() returned %d sandboxes, want 1", len(sandboxes))
	}
}
