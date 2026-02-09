package daemon

import (
	"encoding/json"
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
		{
			name:   "matrix status request",
			method: MethodMatrixStatus,
			params: nil,
			id:     3,
		},
		{
			name:   "matrix send request",
			method: MethodMatrixSend,
			params: MatrixSendParams{RoomID: "!test:ayo.local", Content: "hello"},
			id:     4,
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
			name:   "matrix status response",
			result: MatrixStatusResult{},
			id:     2,
		},
		{
			name:   "nil result",
			result: nil,
			id:     3,
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

func TestMatrixTypes(t *testing.T) {
	// Test that Matrix types can be serialized/deserialized
	t.Run("MatrixSendParams", func(t *testing.T) {
		params := MatrixSendParams{
			AsAgent: "test-agent",
			RoomID:  "!session-123:ayo.local",
			Content: "Hello world",
		}
		data, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		var decoded MatrixSendParams
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if decoded.AsAgent != params.AsAgent {
			t.Errorf("AsAgent mismatch")
		}
		if decoded.RoomID != params.RoomID {
			t.Errorf("RoomID mismatch")
		}
		if decoded.Content != params.Content {
			t.Errorf("Content mismatch")
		}
	})

	t.Run("MatrixRoomsMembersResult", func(t *testing.T) {
		result := MatrixRoomsMembersResult{
			Members: []MemberInfo{
				{UserID: "@agent1:ayo.local", DisplayName: "agent1", IsAgent: true, Handle: "agent1"},
				{UserID: "@agent2:ayo.local", DisplayName: "agent2", IsAgent: true, Handle: "agent2"},
			},
		}
		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		var decoded MatrixRoomsMembersResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if len(decoded.Members) != 2 {
			t.Errorf("Expected 2 members, got %d", len(decoded.Members))
		}
		if decoded.Members[0].Handle != "agent1" {
			t.Errorf("First member handle mismatch")
		}
	})

	t.Run("MatrixReadResult", func(t *testing.T) {
		result := MatrixReadResult{
			Messages: []*QueuedMessage{
				{
					EventID: "$abc123",
					Sender:  "@agent1:ayo.local",
					Content: "Hello",
				},
			},
			HasMore: true,
		}
		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		var decoded MatrixReadResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if len(decoded.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(decoded.Messages))
		}
		if !decoded.HasMore {
			t.Errorf("Expected HasMore to be true")
		}
	})
}

func TestMatrixMethodConstants(t *testing.T) {
	// Verify Matrix method constants are defined
	methods := []string{
		MethodMatrixStatus,
		MethodMatrixRoomsList,
		MethodMatrixRoomsCreate,
		MethodMatrixRoomsMembers,
		MethodMatrixRoomsInvite,
		MethodMatrixSend,
		MethodMatrixRead,
	}

	for _, method := range methods {
		if method == "" {
			t.Errorf("Method constant is empty")
		}
		// All matrix methods should have "matrix." prefix
		if len(method) < 7 || method[:7] != "matrix." {
			t.Errorf("Method %q should start with 'matrix.'", method)
		}
	}
}
