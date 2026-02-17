package daemon

import (
	"testing"
)

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		params  any
		id      int64
		wantErr bool
	}{
		{
			name:   "simple request",
			method: MethodPing,
			params: nil,
			id:     1,
		},
		{
			name:   "request with params",
			method: MethodSandboxAcquire,
			params: SandboxAcquireParams{Agent: "test", Timeout: 30},
			id:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewRequest(tt.method, tt.params, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if req.JSONRPC != "2.0" {
				t.Errorf("JSONRPC = %v, want 2.0", req.JSONRPC)
			}
			if req.Method != tt.method {
				t.Errorf("Method = %v, want %v", req.Method, tt.method)
			}
			if req.ID != tt.id {
				t.Errorf("ID = %v, want %v", req.ID, tt.id)
			}
		})
	}
}

func TestNewResponse(t *testing.T) {
	tests := []struct {
		name    string
		result  any
		id      int64
		wantErr bool
	}{
		{
			name:   "ping response",
			result: PingResult{Pong: true},
			id:     1,
		},
		{
			name:   "nil result",
			result: nil,
			id:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := NewResponse(tt.result, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if resp.JSONRPC != "2.0" {
				t.Errorf("JSONRPC = %v, want 2.0", resp.JSONRPC)
			}
			if resp.ID != tt.id {
				t.Errorf("ID = %v, want %v", resp.ID, tt.id)
			}
			if resp.Error != nil {
				t.Errorf("Error should be nil")
			}
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	err := NewError(ErrCodeMethodNotFound, "method not found")
	resp := NewErrorResponse(err, 42)

	if resp.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want 2.0", resp.JSONRPC)
	}
	if resp.ID != 42 {
		t.Errorf("ID = %v, want 42", resp.ID)
	}
	if resp.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if resp.Error.Code != ErrCodeMethodNotFound {
		t.Errorf("Error.Code = %v, want %v", resp.Error.Code, ErrCodeMethodNotFound)
	}
	if resp.Error.Message != "method not found" {
		t.Errorf("Error.Message = %v, want 'method not found'", resp.Error.Message)
	}
}
