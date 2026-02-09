---
id: ase-xj8h
status: closed
deps: [ase-u200]
links: []
created: 2026-02-09T03:28:48Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Unit tests for daemon RPC protocol

## Background

The daemon RPC protocol (ase-u200) defines the communication between CLI and daemon. These type definitions and serialization need unit tests.

## Test Cases

### Serialization Tests

```go
func TestRequest_Marshal(t *testing.T) {
    req := Request{
        JSONRPC: "2.0",
        Method:  "matrix.send",
        Params: MatrixSendParams{
            Room:    "#session-123",
            Content: "Hello",
        },
        ID: 1,
    }
    
    data, err := json.Marshal(req)
    assert.NoError(t, err)
    
    var decoded Request
    err = json.Unmarshal(data, &decoded)
    assert.NoError(t, err)
    assert.Equal(t, req.Method, decoded.Method)
}

func TestResponse_Marshal(t *testing.T) {
    resp := Response{
        JSONRPC: "2.0",
        Result: MatrixSendResult{
            EventID: "$abc123",
        },
        ID: 1,
    }
    
    data, err := json.Marshal(resp)
    assert.NoError(t, err)
    
    // Verify shape
    var m map[string]any
    json.Unmarshal(data, &m)
    assert.Equal(t, "2.0", m["jsonrpc"])
    assert.NotNil(t, m["result"])
    assert.Nil(t, m["error"])
}

func TestResponse_ErrorMarshal(t *testing.T) {
    resp := Response{
        JSONRPC: "2.0",
        Error: &Error{
            Code:    ErrCodeAgentNotFound,
            Message: "Agent '@foo' not found",
        },
        ID: 1,
    }
    
    data, err := json.Marshal(resp)
    assert.NoError(t, err)
    
    var m map[string]any
    json.Unmarshal(data, &m)
    assert.Nil(t, m["result"])
    assert.NotNil(t, m["error"])
}
```

### Error Code Tests

```go
func TestErrorCodes_Standard(t *testing.T) {
    assert.Equal(t, -32700, ErrCodeParse)
    assert.Equal(t, -32600, ErrCodeInvalidReq)
    assert.Equal(t, -32601, ErrCodeMethodNotFound)
    assert.Equal(t, -32602, ErrCodeInvalidParams)
    assert.Equal(t, -32603, ErrCodeInternal)
}

func TestErrorCodes_Application(t *testing.T) {
    // Application codes are positive
    assert.Greater(t, ErrCodeAgentNotFound, 0)
    assert.Greater(t, ErrCodeFlowNotFound, 0)
    assert.Greater(t, ErrCodeRoomNotFound, 0)
}
```

### Method Params Tests

```go
func TestMatrixSendParams_Validation(t *testing.T) {
    params := MatrixSendParams{
        Room:    "",  // Invalid
        Content: "Hello",
    }
    
    err := params.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "room")
}

func TestFlowRunParams_Validation(t *testing.T) {
    params := FlowRunParams{
        Name: "",  // Invalid
    }
    
    err := params.Validate()
    assert.Error(t, err)
}

func TestFlowRunParams_DefaultAsync(t *testing.T) {
    params := FlowRunParams{
        Name: "test",
    }
    
    assert.False(t, params.Async)  // Default is false
}
```

### Round-trip Tests

```go
func TestRoundTrip_AllMethods(t *testing.T) {
    methods := []struct {
        name   string
        params any
        result any
    }{
        {"matrix.send", MatrixSendParams{Room: "#r", Content: "c"}, MatrixSendResult{}},
        {"flows.run", FlowRunParams{Name: "f"}, FlowRunResult{}},
        {"agents.create", AgentCreateParams{Name: "a"}, AgentCreateResult{}},
        // ... all methods
    }
    
    for _, m := range methods {
        t.Run(m.name, func(t *testing.T) {
            req := Request{
                JSONRPC: "2.0",
                Method:  m.name,
                Params:  m.params,
                ID:      1,
            }
            
            // Serialize
            data, err := json.Marshal(req)
            assert.NoError(t, err)
            
            // Deserialize
            var decoded Request
            err = json.Unmarshal(data, &decoded)
            assert.NoError(t, err)
            assert.Equal(t, m.name, decoded.Method)
        })
    }
}
```

### Files to Create

1. `internal/daemon/protocol/types_test.go`
2. `internal/daemon/protocol/errors_test.go`

## Acceptance Criteria

- [ ] Request serialization tests
- [ ] Response serialization tests (success and error)
- [ ] Standard error code values verified
- [ ] Application error codes are positive
- [ ] Params validation tests for each method
- [ ] Round-trip serialization for all methods
- [ ] All tests pass

