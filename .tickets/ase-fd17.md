---
id: ase-fd17
status: open
deps: [ase-mwdy]
links: []
created: 2026-02-09T03:14:03Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Functional tests for Matrix message broker

## Background

The Matrix message broker runs in the daemon and routes messages between agents. It maintains a single connection to the Conduit homeserver and spawns agents on demand when messages arrive.

## Why This Matters

The broker is the central nervous system for agent-to-agent communication:
- Message routing errors break multi-agent workflows
- Connection failures could orphan agents
- Race conditions could cause message loss

Functional tests verify the broker works correctly in realistic scenarios.

## Implementation Details

### Test Environment Setup

Tests run against an in-memory Conduit instance or mock Matrix server:

```go
// internal/daemon/broker_test.go
func setupTestBroker(t *testing.T) (*MessageBroker, *MockMatrixServer) {
    server := NewMockMatrixServer()
    broker := NewMessageBroker(server.URL())
    
    t.Cleanup(func() {
        broker.Close()
        server.Close()
    })
    
    return broker, server
}
```

### Test Cases

**broker_test.go:**

```go
func TestBroker_SendMessage(t *testing.T) {
    broker, server := setupTestBroker(t)
    
    err := broker.Send("@code-reviewer", "Please review this code")
    
    assert.NoError(t, err)
    assert.Equal(t, 1, server.MessageCount("@code-reviewer"))
}

func TestBroker_ReceiveMessage(t *testing.T) {
    broker, server := setupTestBroker(t)
    
    // Simulate incoming message
    server.InjectMessage("@ayo", "Help me with a task")
    
    msg, err := broker.Receive(time.Second)
    
    assert.NoError(t, err)
    assert.Equal(t, "Help me with a task", msg.Content)
    assert.Equal(t, "@ayo", msg.To)
}

func TestBroker_AgentSpawnOnMessage(t *testing.T) {
    broker, _ := setupTestBroker(t)
    spawned := make(chan string, 1)
    broker.OnSpawn(func(agent string) {
        spawned <- agent
    })
    
    broker.InjectMessage("@code-reviewer", "Review this")
    
    select {
    case agent := <-spawned:
        assert.Equal(t, "@code-reviewer", agent)
    case <-time.After(time.Second):
        t.Fatal("Agent not spawned")
    }
}

func TestBroker_MessageToNonexistentAgent(t *testing.T) {
    broker, _ := setupTestBroker(t)
    
    err := broker.Send("@nonexistent", "Hello")
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unknown agent")
}

func TestBroker_Reconnection(t *testing.T) {
    broker, server := setupTestBroker(t)
    
    // Simulate disconnect
    server.Disconnect()
    
    // Wait for reconnection
    time.Sleep(2 * time.Second)
    
    // Should reconnect and work
    err := broker.Send("@ayo", "Test after reconnect")
    assert.NoError(t, err)
}

func TestBroker_MessageOrdering(t *testing.T) {
    broker, server := setupTestBroker(t)
    
    // Send multiple messages rapidly
    for i := 0; i < 10; i++ {
        broker.Send("@ayo", fmt.Sprintf("Message %d", i))
    }
    
    // Verify ordering preserved
    messages := server.GetMessages("@ayo")
    for i, msg := range messages {
        assert.Equal(t, fmt.Sprintf("Message %d", i), msg.Content)
    }
}

func TestBroker_RoomCreation(t *testing.T) {
    broker, server := setupTestBroker(t)
    
    // Sending to new room should create it
    err := broker.SendToRoom("#new-project", "@ayo", "Starting project")
    
    assert.NoError(t, err)
    assert.True(t, server.RoomExists("#new-project"))
}

func TestBroker_RoomJoin(t *testing.T) {
    broker, server := setupTestBroker(t)
    
    server.CreateRoom("#team")
    err := broker.JoinRoom("#team", "@code-reviewer")
    
    assert.NoError(t, err)
    assert.True(t, server.IsInRoom("#team", "@code-reviewer"))
}

func TestBroker_BroadcastToRoom(t *testing.T) {
    broker, server := setupTestBroker(t)
    
    server.CreateRoom("#team")
    server.JoinRoom("#team", "@ayo")
    server.JoinRoom("#team", "@code-reviewer")
    server.JoinRoom("#team", "@writer")
    
    broker.BroadcastToRoom("#team", "Team announcement")
    
    // All members should receive
    assert.Equal(t, 1, server.MessageCount("@ayo"))
    assert.Equal(t, 1, server.MessageCount("@code-reviewer"))
    assert.Equal(t, 1, server.MessageCount("@writer"))
}

func TestBroker_MessageHistory(t *testing.T) {
    broker, server := setupTestBroker(t)
    
    // Send some messages
    broker.Send("@ayo", "Message 1")
    broker.Send("@ayo", "Message 2")
    
    // Retrieve history
    history, err := broker.GetHistory("@ayo", 10)
    
    assert.NoError(t, err)
    assert.Len(t, history, 2)
}

func TestBroker_UnrestrictedAgentInvisible(t *testing.T) {
    broker, _ := setupTestBroker(t)
    
    // Unrestricted agents shouldn't appear in agent lists
    agents := broker.ListAgents()
    
    for _, a := range agents {
        if a.TrustLevel == "unrestricted" {
            t.Errorf("Unrestricted agent %s visible in list", a.Name)
        }
    }
}
```

### Mock Matrix Server

```go
// internal/daemon/mock_matrix_test.go
type MockMatrixServer struct {
    server    *httptest.Server
    rooms     map[string]*Room
    messages  map[string][]Message
    mu        sync.Mutex
}

func NewMockMatrixServer() *MockMatrixServer {
    m := &MockMatrixServer{
        rooms:    make(map[string]*Room),
        messages: make(map[string][]Message),
    }
    m.server = httptest.NewServer(m)
    return m
}

func (m *MockMatrixServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Implement Matrix client-server API endpoints
    // /_matrix/client/v3/sync
    // /_matrix/client/v3/rooms/{roomId}/send/{eventType}/{txnId}
    // etc.
}
```

## Acceptance Criteria

- [ ] MockMatrixServer implements key Matrix APIs
- [ ] Send/receive message tests
- [ ] Agent spawn on message test
- [ ] Message ordering preserved
- [ ] Reconnection handling
- [ ] Room creation and join
- [ ] Broadcast to room
- [ ] Message history retrieval
- [ ] Unrestricted agent visibility filtering
- [ ] All tests pass
- [ ] No race conditions (run with -race)

