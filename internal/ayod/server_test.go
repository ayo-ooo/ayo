package ayod

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServerClientRoundtrip(t *testing.T) {
	// Create temp socket
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Start server
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	server := NewServer()
	go server.Serve(listener)
	defer server.Shutdown()

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	// Connect client
	client, err := Connect(socketPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer client.Close()

	// Test Health
	health, err := client.Health()
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if health.Status != "ok" {
		t.Errorf("Health.Status = %q, want %q", health.Status, "ok")
	}
}

func TestRPCRequestEncodeDecode(t *testing.T) {
	req := RPCRequest{
		Method: "Exec",
		Params: []byte(`{"user":"test","command":["echo","hello"]}`),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded RPCRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Method != req.Method {
		t.Errorf("Method = %q, want %q", decoded.Method, req.Method)
	}
}

func TestFileHandler(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewFileHandler()

	// Test WriteFile
	testPath := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")
	err := handler.WriteFile(WriteFileRequest{
		Path:    testPath,
		Content: content,
		Mode:    0644,
	})
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testPath); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	// Test ReadFile
	resp, err := handler.ReadFile(ReadFileRequest{Path: testPath})
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(resp.Content) != string(content) {
		t.Errorf("ReadFile content = %q, want %q", resp.Content, content)
	}
}

func TestFileHandler_CreateParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewFileHandler()

	// Write to nested path
	testPath := filepath.Join(tmpDir, "a", "b", "c", "test.txt")
	err := handler.WriteFile(WriteFileRequest{
		Path:    testPath,
		Content: []byte("nested"),
		Mode:    0644,
	})
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testPath); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestFileHandler_ReadDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewFileHandler()

	// Try to read a directory
	_, err := handler.ReadFile(ReadFileRequest{Path: tmpDir})
	if err == nil {
		t.Error("expected error reading directory, got nil")
	}
}
