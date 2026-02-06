package daemon

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	gosync "sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/alexcabrera/ayo/internal/sync"
)

// IRCMessage represents a parsed IRC message from logs.
type IRCMessage struct {
	Timestamp time.Time
	Channel   string
	Sender    string
	Text      string
	Mentions  []string // Agent handles mentioned
}

// IRCBridge monitors IRC logs and routes messages.
type IRCBridge struct {
	mu             gosync.RWMutex
	watcher        *fsnotify.Watcher
	running        bool
	stopCh         chan struct{}
	pendingMsgs    map[string][]IRCMessage // Agent handle -> pending messages
	onMention      func(agent string, msg IRCMessage)
	onMessage      func(msg IRCMessage)
}

// IRC log format: [2025-02-05 10:30:45] <sender> message text
var ircLogPattern = regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\] <([^>]+)> (.*)$`)

// Mention pattern: @agent or @agent:
var mentionPattern = regexp.MustCompile(`@(\w+)`)

// NewIRCBridge creates a new IRC bridge.
func NewIRCBridge() *IRCBridge {
	return &IRCBridge{
		stopCh:      make(chan struct{}),
		pendingMsgs: make(map[string][]IRCMessage),
	}
}

// OnMention sets a callback for when an agent is mentioned.
func (b *IRCBridge) OnMention(fn func(agent string, msg IRCMessage)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onMention = fn
}

// OnMessage sets a callback for all messages.
func (b *IRCBridge) OnMessage(fn func(msg IRCMessage)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onMessage = fn
}

// Start begins monitoring IRC logs.
func (b *IRCBridge) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}
	b.running = true
	b.mu.Unlock()

	// Create file watcher
	var err error
	b.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Get IRC logs directory
	ircLogsDir := sync.IRCLogsDir()

	// Ensure directory exists
	if err := os.MkdirAll(ircLogsDir, 0755); err != nil {
		b.watcher.Close()
		return err
	}

	// Watch the directory for new files
	if err := b.watcher.Add(ircLogsDir); err != nil {
		b.watcher.Close()
		return err
	}

	// Start watching in background
	go b.watchLoop(ctx, ircLogsDir)

	return nil
}

// Stop stops the IRC bridge.
func (b *IRCBridge) Stop() {
	b.mu.Lock()
	if !b.running {
		b.mu.Unlock()
		return
	}
	b.running = false
	b.mu.Unlock()

	close(b.stopCh)
	if b.watcher != nil {
		b.watcher.Close()
	}
}

// GetPendingMessages returns pending messages for an agent and clears them.
func (b *IRCBridge) GetPendingMessages(agent string) []IRCMessage {
	b.mu.Lock()
	defer b.mu.Unlock()

	agent = strings.TrimPrefix(agent, "@")
	msgs := b.pendingMsgs[agent]
	delete(b.pendingMsgs, agent)
	return msgs
}

// HasPendingMessages returns true if there are pending messages for an agent.
func (b *IRCBridge) HasPendingMessages(agent string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	agent = strings.TrimPrefix(agent, "@")
	return len(b.pendingMsgs[agent]) > 0
}

// FormatPendingContext formats pending messages as context for injection.
func (b *IRCBridge) FormatPendingContext(agent string) string {
	msgs := b.GetPendingMessages(agent)
	if len(msgs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<pending_irc_messages>\n")
	sb.WriteString("You have messages from other agents:\n\n")
	for _, msg := range msgs {
		sb.WriteString("- ")
		sb.WriteString(msg.Timestamp.Format("15:04:05"))
		sb.WriteString(" <")
		sb.WriteString(msg.Sender)
		sb.WriteString("> ")
		sb.WriteString(msg.Text)
		sb.WriteString("\n")
	}
	sb.WriteString("</pending_irc_messages>")
	return sb.String()
}

func (b *IRCBridge) watchLoop(ctx context.Context, ircLogsDir string) {
	// Track file positions for tailing
	filePositions := make(map[string]int64)

	// Do initial scan of existing files
	entries, err := os.ReadDir(ircLogsDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
				path := filepath.Join(ircLogsDir, entry.Name())
				info, err := entry.Info()
				if err == nil {
					filePositions[path] = info.Size()
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-b.stopCh:
			return
		case event, ok := <-b.watcher.Events:
			if !ok {
				return
			}

			// Only interested in writes to .log files
			if !strings.HasSuffix(event.Name, ".log") {
				continue
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				b.processNewLines(event.Name, filePositions)
			} else if event.Op&fsnotify.Create == fsnotify.Create {
				// New log file - start from beginning
				filePositions[event.Name] = 0
				b.processNewLines(event.Name, filePositions)
			}
		case _, ok := <-b.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue
		}
	}
}

func (b *IRCBridge) processNewLines(path string, positions map[string]int64) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	// Get current position
	pos, ok := positions[path]
	if !ok {
		pos = 0
	}

	// Seek to position
	if pos > 0 {
		if _, err := file.Seek(pos, 0); err != nil {
			return
		}
	}

	// Extract channel from filename
	channel := strings.TrimSuffix(filepath.Base(path), ".log")

	// Read new lines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if msg, ok := b.parseLine(line, channel); ok {
			b.handleMessage(msg)
		}
	}

	// Update position
	newPos, err := file.Seek(0, 1) // Get current position
	if err == nil {
		positions[path] = newPos
	}
}

func (b *IRCBridge) parseLine(line, channel string) (IRCMessage, bool) {
	matches := ircLogPattern.FindStringSubmatch(line)
	if matches == nil {
		return IRCMessage{}, false
	}

	timestamp, err := time.Parse("2006-01-02 15:04:05", matches[1])
	if err != nil {
		return IRCMessage{}, false
	}

	msg := IRCMessage{
		Timestamp: timestamp,
		Channel:   channel,
		Sender:    matches[2],
		Text:      matches[3],
	}

	// Extract mentions
	mentionMatches := mentionPattern.FindAllStringSubmatch(msg.Text, -1)
	for _, m := range mentionMatches {
		msg.Mentions = append(msg.Mentions, m[1])
	}

	return msg, true
}

func (b *IRCBridge) handleMessage(msg IRCMessage) {
	b.mu.RLock()
	onMention := b.onMention
	onMessage := b.onMessage
	b.mu.RUnlock()

	// Invoke message callback
	if onMessage != nil {
		onMessage(msg)
	}

	// Store pending messages for mentioned agents
	for _, agent := range msg.Mentions {
		b.mu.Lock()
		b.pendingMsgs[agent] = append(b.pendingMsgs[agent], msg)
		b.mu.Unlock()

		// Invoke mention callback
		if onMention != nil {
			onMention(agent, msg)
		}
	}
}
