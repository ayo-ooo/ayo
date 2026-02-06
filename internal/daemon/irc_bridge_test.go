package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIRCBridge_ParseLine(t *testing.T) {
	bridge := NewIRCBridge()

	tests := []struct {
		name      string
		line      string
		channel   string
		wantOk    bool
		wantSender string
		wantText   string
		wantMentions int
	}{
		{
			name:    "valid message",
			line:    "[2025-02-05 10:30:45] <ayo> Hello world",
			channel: "general",
			wantOk:  true,
			wantSender: "ayo",
			wantText: "Hello world",
			wantMentions: 0,
		},
		{
			name:    "message with mention",
			line:    "[2025-02-05 10:30:45] <crush> Hey @ayo check this",
			channel: "general",
			wantOk:  true,
			wantSender: "crush",
			wantText: "Hey @ayo check this",
			wantMentions: 1,
		},
		{
			name:    "message with multiple mentions",
			line:    "[2025-02-05 10:30:45] <user> @ayo @crush please help",
			channel: "project",
			wantOk:  true,
			wantSender: "user",
			wantText: "@ayo @crush please help",
			wantMentions: 2,
		},
		{
			name:    "invalid format",
			line:    "not a valid line",
			channel: "general",
			wantOk:  false,
		},
		{
			name:    "empty line",
			line:    "",
			channel: "general",
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, ok := bridge.parseLine(tt.line, tt.channel)
			if ok != tt.wantOk {
				t.Fatalf("parseLine() ok = %v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if msg.Sender != tt.wantSender {
				t.Errorf("Sender = %q, want %q", msg.Sender, tt.wantSender)
			}
			if msg.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", msg.Text, tt.wantText)
			}
			if len(msg.Mentions) != tt.wantMentions {
				t.Errorf("Mentions count = %d, want %d", len(msg.Mentions), tt.wantMentions)
			}
			if msg.Channel != tt.channel {
				t.Errorf("Channel = %q, want %q", msg.Channel, tt.channel)
			}
		})
	}
}

func TestIRCBridge_PendingMessages(t *testing.T) {
	bridge := NewIRCBridge()

	// Simulate receiving a message with mention
	msg := IRCMessage{
		Timestamp: time.Now(),
		Channel:   "general",
		Sender:    "crush",
		Text:      "Hey @ayo check this",
		Mentions:  []string{"ayo"},
	}

	// Handle the message
	bridge.handleMessage(msg)

	// Check pending messages
	if !bridge.HasPendingMessages("ayo") {
		t.Error("expected pending messages for ayo")
	}
	if !bridge.HasPendingMessages("@ayo") { // Should work with @ prefix too
		t.Error("expected pending messages for @ayo")
	}
	if bridge.HasPendingMessages("crush") {
		t.Error("did not expect pending messages for crush (sender)")
	}

	// Get and clear pending messages
	msgs := bridge.GetPendingMessages("ayo")
	if len(msgs) != 1 {
		t.Fatalf("expected 1 pending message, got %d", len(msgs))
	}
	if msgs[0].Text != msg.Text {
		t.Errorf("message text = %q, want %q", msgs[0].Text, msg.Text)
	}

	// Should be cleared now
	if bridge.HasPendingMessages("ayo") {
		t.Error("expected pending messages to be cleared")
	}
}

func TestIRCBridge_FormatPendingContext(t *testing.T) {
	bridge := NewIRCBridge()

	// No pending messages
	ctx := bridge.FormatPendingContext("ayo")
	if ctx != "" {
		t.Error("expected empty context for no pending messages")
	}

	// Add a pending message
	msg := IRCMessage{
		Timestamp: time.Date(2025, 2, 5, 10, 30, 45, 0, time.UTC),
		Channel:   "general",
		Sender:    "crush",
		Text:      "Hey @ayo check this",
		Mentions:  []string{"ayo"},
	}
	bridge.handleMessage(msg)

	// Get formatted context
	ctx = bridge.FormatPendingContext("ayo")
	if ctx == "" {
		t.Fatal("expected non-empty context")
	}
	if !contains(ctx, "pending_irc_messages") {
		t.Error("expected context to contain pending_irc_messages tag")
	}
	if !contains(ctx, "crush") {
		t.Error("expected context to contain sender name")
	}
	if !contains(ctx, "Hey @ayo check this") {
		t.Error("expected context to contain message text")
	}
}

func TestIRCBridge_Callbacks(t *testing.T) {
	bridge := NewIRCBridge()

	var receivedMessages []IRCMessage
	var mentionedAgents []string

	bridge.OnMessage(func(msg IRCMessage) {
		receivedMessages = append(receivedMessages, msg)
	})

	bridge.OnMention(func(agent string, msg IRCMessage) {
		mentionedAgents = append(mentionedAgents, agent)
	})

	// Handle a message with mentions
	msg := IRCMessage{
		Timestamp: time.Now(),
		Channel:   "general",
		Sender:    "user",
		Text:      "@ayo @crush please help",
		Mentions:  []string{"ayo", "crush"},
	}
	bridge.handleMessage(msg)

	// Check message callback was called
	if len(receivedMessages) != 1 {
		t.Fatalf("expected 1 received message, got %d", len(receivedMessages))
	}

	// Check mention callback was called for each mention
	if len(mentionedAgents) != 2 {
		t.Fatalf("expected 2 mentioned agents, got %d", len(mentionedAgents))
	}
}

func TestIRCBridge_ProcessNewLines(t *testing.T) {
	bridge := NewIRCBridge()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "irc-bridge-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a log file
	logPath := filepath.Join(tmpDir, "general.log")
	content := `[2025-02-05 10:30:45] <ayo> Hello
[2025-02-05 10:30:50] <crush> Hey @ayo
`
	if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}

	// Process the file
	positions := make(map[string]int64)
	bridge.processNewLines(logPath, positions)

	// Should have pending message for ayo
	if !bridge.HasPendingMessages("ayo") {
		t.Error("expected pending messages for ayo")
	}

	// Position should be updated
	if positions[logPath] == 0 {
		t.Error("expected position to be updated")
	}

	// Process again - should not add more pending messages
	msgsBefore := len(bridge.pendingMsgs["ayo"])
	bridge.processNewLines(logPath, positions)
	msgsAfter := len(bridge.pendingMsgs["ayo"])
	if msgsAfter != msgsBefore {
		t.Errorf("expected no new messages, got %d new", msgsAfter-msgsBefore)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(s == substr || len(s) > len(substr) && 
			(s[:len(substr)] == substr || contains(s[1:], substr)))
}
