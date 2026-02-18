package daemon

import (
	"encoding/json"
	"testing"
)

func TestAgentInvokeParams(t *testing.T) {
	// Test marshaling/unmarshaling
	params := AgentInvokeParams{
		Agent:     "@ayo",
		Prompt:    "Hello, world!",
		SessionID: "session-123",
		Skills:    []string{"coding", "testing"},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded AgentInvokeParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Agent != params.Agent {
		t.Errorf("Agent mismatch: got %q, want %q", decoded.Agent, params.Agent)
	}
	if decoded.Prompt != params.Prompt {
		t.Errorf("Prompt mismatch: got %q, want %q", decoded.Prompt, params.Prompt)
	}
	if decoded.SessionID != params.SessionID {
		t.Errorf("SessionID mismatch: got %q, want %q", decoded.SessionID, params.SessionID)
	}
	if len(decoded.Skills) != len(params.Skills) {
		t.Errorf("Skills length mismatch: got %d, want %d", len(decoded.Skills), len(params.Skills))
	}
}

func TestAgentInvokeResult(t *testing.T) {
	// Test marshaling/unmarshaling
	result := AgentInvokeResult{
		SessionID: "session-456",
		Response:  "Hello! I'm @ayo.",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded AgentInvokeResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.SessionID != result.SessionID {
		t.Errorf("SessionID mismatch: got %q, want %q", decoded.SessionID, result.SessionID)
	}
	if decoded.Response != result.Response {
		t.Errorf("Response mismatch: got %q, want %q", decoded.Response, result.Response)
	}
}
