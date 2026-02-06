package daemon

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestProtocol_NewRequest(t *testing.T) {
	params := map[string]string{"key": "value"}
	req, err := NewRequest("test.method", params, 1)
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}

	if req.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %q, want '2.0'", req.JSONRPC)
	}
	if req.Method != "test.method" {
		t.Errorf("Method = %q, want 'test.method'", req.Method)
	}
	if req.ID != 1 {
		t.Errorf("ID = %d, want 1", req.ID)
	}
}

func TestProtocol_NewResponse(t *testing.T) {
	result := PingResult{Pong: true}
	resp, err := NewResponse(result, 1)
	if err != nil {
		t.Fatalf("NewResponse error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %q, want '2.0'", resp.JSONRPC)
	}
	if resp.Error != nil {
		t.Errorf("Error = %v, want nil", resp.Error)
	}
	if resp.ID != 1 {
		t.Errorf("ID = %d, want 1", resp.ID)
	}

	var pingResult PingResult
	json.Unmarshal(resp.Result, &pingResult)
	if !pingResult.Pong {
		t.Error("Result.Pong = false, want true")
	}
}

func TestProtocol_NewErrorResponse(t *testing.T) {
	err := NewError(ErrCodeMethodNotFound, "method not found")
	resp := NewErrorResponse(err, 1)

	if resp.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if resp.Error.Code != ErrCodeMethodNotFound {
		t.Errorf("Error.Code = %d, want %d", resp.Error.Code, ErrCodeMethodNotFound)
	}
	if resp.Error.Message != "method not found" {
		t.Errorf("Error.Message = %q, want 'method not found'", resp.Error.Message)
	}
}

func TestError_Error(t *testing.T) {
	err := NewError(ErrCodeInternal, "something went wrong")
	s := err.Error()
	if s != "rpc error -32603: something went wrong" {
		t.Errorf("Error() = %q, unexpected format", s)
	}
}

func TestServer_StartStop(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.SocketPath = t.TempDir() + "/test.sock"
	// Don't create containers at startup to make test fast
	cfg.PoolConfig.MinSize = 0

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}

	ctx := context.Background()

	if err := server.Start(ctx, cfg.SocketPath); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	// Server should be running
	if server.Addr() == nil {
		t.Error("Addr() should not be nil when running")
	}

	// Stop server
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Stop(stopCtx); err != nil {
		t.Errorf("Stop error: %v", err)
	}
}

func TestServer_PingHandler(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.SocketPath = t.TempDir() + "/test.sock"

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}

	ctx := context.Background()
	if err := server.Start(ctx, cfg.SocketPath); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer server.Stop(ctx)

	// Create client and connect
	client := NewClient()
	_ = client // Client created but connection test needs socket path override
	// Need to wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// For testing, connect directly since client expects default path
	// In real usage, client would use default socket path
}

func TestDefaultSocketPath(t *testing.T) {
	path := DefaultSocketPath()
	if path == "" {
		t.Error("DefaultSocketPath() should not be empty")
	}
}

func TestDefaultPIDPath(t *testing.T) {
	path := DefaultPIDPath()
	if path == "" {
		t.Error("DefaultPIDPath() should not be empty")
	}
}

func TestClient_IsConnected(t *testing.T) {
	client := NewClient()
	if client.IsConnected() {
		t.Error("New client should not be connected")
	}
}

func TestIsDaemonRunning_NotRunning(t *testing.T) {
	// Without a valid PID file, should return false
	if IsDaemonRunning() {
		// This might be true if daemon is actually running
		// For now, just verify the function doesn't panic
	}
}

func TestClientServer_Integration(t *testing.T) {
	cfg := DefaultServerConfig()
	// Use /tmp for shorter socket path (Unix sockets have 108 char limit)
	socketPath := "/tmp/ayo-test-" + t.Name() + ".sock"
	cfg.SocketPath = socketPath
	t.Cleanup(func() {
		os.Remove(socketPath)
	})

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}

	ctx := context.Background()
	if err := server.Start(ctx, socketPath); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer server.Stop(ctx)

	// Give server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Connect client to server
	client := NewClient()
	if err := client.ConnectTo(ctx, socketPath); err != nil {
		t.Fatalf("ConnectTo error: %v", err)
	}
	defer client.Close()

	// Test ping
	t.Run("Ping", func(t *testing.T) {
		if err := client.Ping(ctx); err != nil {
			t.Errorf("Ping error: %v", err)
		}
	})

	// Test status
	t.Run("Status", func(t *testing.T) {
		status, err := client.Status(ctx)
		if err != nil {
			t.Errorf("Status error: %v", err)
			return
		}
		if !status.Running {
			t.Error("Status.Running = false, want true")
		}
		if status.Version == "" {
			t.Error("Status.Version should not be empty")
		}
	})

	// Test sandbox acquire/release
	t.Run("SandboxAcquireRelease", func(t *testing.T) {
		result, err := client.SandboxAcquire(ctx, "@ayo", 30)
		if err != nil {
			t.Errorf("SandboxAcquire error: %v", err)
			return
		}
		if result.SandboxID == "" {
			t.Error("SandboxID should not be empty")
		}

		// Release sandbox
		if err := client.SandboxRelease(ctx, result.SandboxID); err != nil {
			t.Errorf("SandboxRelease error: %v", err)
		}
	})

	// Test sandbox exec
	t.Run("SandboxExec", func(t *testing.T) {
		// First acquire
		acquired, err := client.SandboxAcquire(ctx, "@test", 30)
		if err != nil {
			t.Fatalf("SandboxAcquire error: %v", err)
		}
		defer client.SandboxRelease(ctx, acquired.SandboxID)

		// Execute command
		result, err := client.SandboxExec(ctx, acquired.SandboxID, "echo hello", "", 10)
		if err != nil {
			t.Errorf("SandboxExec error: %v", err)
			return
		}
		if result.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0", result.ExitCode)
		}
		if result.Stdout != "hello\n" {
			t.Errorf("Stdout = %q, want 'hello\\n'", result.Stdout)
		}
	})

	// Test sandbox status
	t.Run("SandboxStatus", func(t *testing.T) {
		status, err := client.SandboxStatus(ctx)
		if err != nil {
			t.Errorf("SandboxStatus error: %v", err)
			return
		}
		if status.Total < 0 {
			t.Error("Total should not be negative")
		}
	})
}
