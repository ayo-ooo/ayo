---
id: ase-bj67
status: closed
deps: [ase-syl9]
links: []
created: 2026-02-09T03:27:25Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Unit tests for chat CLI commands

## Background

The chat CLI (ase-syl9) provides Matrix communication for agents. These commands need unit tests.

## Test Cases

### chat rooms Tests

```go
func TestChatRooms_Empty(t *testing.T) {
    output, err := runCmd("chat", "rooms")
    assert.NoError(t, err)
    assert.Contains(t, output, "No rooms")
}

func TestChatRooms_WithRooms(t *testing.T) {
    setupTestRooms(t)
    output, err := runCmd("chat", "rooms")
    assert.NoError(t, err)
    assert.Contains(t, output, "#session-")
}

func TestChatRooms_FilterBySession(t *testing.T) {
    setupTestRooms(t)
    output, err := runCmd("chat", "rooms", "--session", "abc123")
    assert.NoError(t, err)
    // Only rooms for session abc123
}

func TestChatRooms_JSON(t *testing.T) {
    setupTestRooms(t)
    output, err := runCmd("chat", "rooms", "--json")
    assert.NoError(t, err)
    var rooms []Room
    err = json.Unmarshal([]byte(output), &rooms)
    assert.NoError(t, err)
}
```

### chat send Tests

```go
func TestChatSend_Basic(t *testing.T) {
    room := setupTestRoom(t)
    output, err := runCmd("chat", "send", room, "Hello world")
    assert.NoError(t, err)
    assert.Contains(t, output, "Sent")
}

func TestChatSend_AsAgent(t *testing.T) {
    room := setupTestRoom(t)
    output, err := runCmd("chat", "send", room, "Hello", "--as", "@researcher")
    assert.NoError(t, err)
    // Verify message sent as @researcher
}

func TestChatSend_FromFile(t *testing.T) {
    room := setupTestRoom(t)
    tmpFile := createTempFile(t, "file content")
    output, err := runCmd("chat", "send", room, "--file", tmpFile)
    assert.NoError(t, err)
}

func TestChatSend_InvalidRoom(t *testing.T) {
    _, err := runCmd("chat", "send", "#nonexistent", "Hello")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "room not found")
}

func TestChatSend_EmptyMessage(t *testing.T) {
    room := setupTestRoom(t)
    _, err := runCmd("chat", "send", room, "")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "message required")
}
```

### chat read Tests

```go
func TestChatRead_Basic(t *testing.T) {
    room := setupTestRoomWithMessages(t, 5)
    output, err := runCmd("chat", "read", room)
    assert.NoError(t, err)
    // Should show messages
}

func TestChatRead_Limit(t *testing.T) {
    room := setupTestRoomWithMessages(t, 50)
    output, err := runCmd("chat", "read", room, "--limit", "10")
    assert.NoError(t, err)
    // Count messages in output
    lines := strings.Split(output, "\n")
    messageLines := filterMessageLines(lines)
    assert.LessOrEqual(t, len(messageLines), 10)
}

func TestChatRead_Since(t *testing.T) {
    room := setupTestRoomWithMessages(t, 10)
    output, err := runCmd("chat", "read", room, "--since", "1h")
    assert.NoError(t, err)
}

func TestChatRead_JSON(t *testing.T) {
    room := setupTestRoomWithMessages(t, 5)
    output, err := runCmd("chat", "read", room, "--json")
    assert.NoError(t, err)
    var messages []Message
    err = json.Unmarshal([]byte(output), &messages)
    assert.NoError(t, err)
}

func TestChatRead_EmptyRoom(t *testing.T) {
    room := setupTestRoom(t)
    output, err := runCmd("chat", "read", room)
    assert.NoError(t, err)
    assert.Contains(t, output, "No messages")
}
```

### chat who Tests

```go
func TestChatWho_Basic(t *testing.T) {
    room := setupTestRoomWithMembers(t, "@ayo", "@researcher")
    output, err := runCmd("chat", "who", room)
    assert.NoError(t, err)
    assert.Contains(t, output, "@ayo")
    assert.Contains(t, output, "@researcher")
}

func TestChatWho_EmptyRoom(t *testing.T) {
    room := setupTestRoom(t)
    output, err := runCmd("chat", "who", room)
    assert.NoError(t, err)
    // Just creator
}
```

### chat invite/join/leave Tests

```go
func TestChatInvite(t *testing.T) {
    room := setupTestRoom(t)
    output, err := runCmd("chat", "invite", room, "@researcher")
    assert.NoError(t, err)
    assert.Contains(t, output, "Invited")
}

func TestChatJoin(t *testing.T) {
    room := setupTestRoom(t)
    output, err := runCmd("chat", "join", room)
    assert.NoError(t, err)
}

func TestChatLeave(t *testing.T) {
    room := setupTestRoomWithMember(t, "@researcher")
    // Set agent context
    os.Setenv("AYO_AGENT_HANDLE", "researcher")
    output, err := runCmd("chat", "leave", room)
    assert.NoError(t, err)
}
```

### chat create Tests

```go
func TestChatCreate(t *testing.T) {
    output, err := runCmd("chat", "create", "--name", "#test-room")
    assert.NoError(t, err)
    assert.Contains(t, output, "Created")
}

func TestChatCreate_InvalidName(t *testing.T) {
    _, err := runCmd("chat", "create", "--name", "no-hash")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "must start with #")
}
```

### Files to Create

1. `cmd/ayo/chat_test.go` - All chat CLI tests

## Acceptance Criteria

- [ ] rooms subcommand tests (empty, with rooms, filters, JSON)
- [ ] send subcommand tests (basic, as agent, from file, errors)
- [ ] read subcommand tests (basic, limit, since, JSON, empty)
- [ ] who subcommand tests
- [ ] invite/join/leave tests
- [ ] create subcommand tests
- [ ] Error message verification
- [ ] JSON output parsing
- [ ] All tests pass

