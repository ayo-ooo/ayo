package daemon

import (
	"context"
	"testing"

	"github.com/alexcabrera/ayo/internal/config"
)

func TestSandboxAwareInvoker_NilClient(t *testing.T) {
	invoker := NewSandboxAwareInvoker(nil)

	_, err := invoker.Invoke(context.Background(), "test-agent", "test prompt")
	if err == nil {
		t.Fatal("expected error for nil client")
	}
	if err.Error() != "daemon client not initialized" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSandboxAwareInvoker_Interface(t *testing.T) {
	// Verify the invoker implements AgentInvoker
	var _ AgentInvoker = (*SandboxAwareInvoker)(nil)
}

func TestServerAgentInvoker_Interface(t *testing.T) {
	// Verify the invoker implements AgentInvoker
	var _ AgentInvoker = (*ServerAgentInvoker)(nil)
}

func TestNewServerAgentInvoker(t *testing.T) {
	cfg := config.Config{}
	invoker := NewServerAgentInvoker(cfg)
	if invoker == nil {
		t.Fatal("expected non-nil invoker")
	}
}

func TestSandboxAwareInvoker_InvokeInSquad_NilClient(t *testing.T) {
	invoker := NewSandboxAwareInvoker(nil)

	_, err := invoker.InvokeInSquad(context.Background(), "#test-squad", "test-agent", "test prompt")
	if err == nil {
		t.Fatal("expected error for nil client")
	}
	if err.Error() != "daemon client not initialized" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestServerAgentInvoker_InvokeInSquad_NotSupported(t *testing.T) {
	cfg := config.Config{}
	invoker := NewServerAgentInvoker(cfg)

	_, err := invoker.InvokeInSquad(context.Background(), "#test-squad", "test-agent", "test prompt")
	if err == nil {
		t.Fatal("expected error for server-side squad invocation")
	}
	if err.Error() != "server-side InvokeInSquad not supported; use SandboxAwareInvoker for squad context" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNormalizeAgentHandle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"@agent", "agent"},
		{"agent", "agent"},
		{"@foo", "foo"},
		{"", ""},
	}

	for _, tc := range tests {
		result := normalizeAgentHandle(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeAgentHandle(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}
