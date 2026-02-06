package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/sync"
)

// IRCMessage represents a parsed IRC log message.
type IRCMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Channel   string    `json:"channel"`
	Sender    string    `json:"sender"`
	Text      string    `json:"text"`
}

// IRC log format regex: [2025-02-05 10:30:45] <sender> message text
var ircLogRegex = regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\] <([^>]+)> (.*)$`)

func newMessagesCmd() *cobra.Command {
	var follow bool
	var channel string
	var search string
	var fromUser string
	var toUser string
	var limit int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "messages",
		Short: "View IRC messages from sandbox",
		Long: `View IRC messages from sandbox inter-agent communication.

Examples:
  ayo messages                    # Recent from #general
  ayo messages -f                 # Live tail mode
  ayo messages -c project         # Messages from #project
  ayo messages -s "error"         # Search for "error"
  ayo messages --from @ayo        # Messages from @ayo
  ayo messages --json             # JSON output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine log file path
			ircLogsDir := sync.IRCLogsDir()
			if channel == "" {
				channel = "general"
			}
			channel = strings.TrimPrefix(channel, "#")
			logFile := filepath.Join(ircLogsDir, channel+".log")

			// Check if log file exists
			if _, err := os.Stat(logFile); os.IsNotExist(err) {
				return fmt.Errorf("no logs found for channel #%s", channel)
			}

			if follow {
				return followLogs(logFile, channel, search, fromUser, toUser, jsonOutput)
			}

			return showLogs(logFile, channel, search, fromUser, toUser, limit, jsonOutput)
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow mode (like tail -f)")
	cmd.Flags().StringVarP(&channel, "channel", "c", "", "channel to view (default: general)")
	cmd.Flags().StringVarP(&search, "search", "s", "", "search for messages containing text")
	cmd.Flags().StringVar(&fromUser, "from", "", "filter by sender (e.g., @ayo)")
	cmd.Flags().StringVar(&toUser, "to", "", "filter by recipient")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "number of messages to show")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}

func showLogs(logFile, channel, search, fromUser, toUser string, limit int, jsonOutput bool) error {
	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read all lines
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	// Get last N lines
	startIdx := 0
	if len(lines) > limit {
		startIdx = len(lines) - limit
	}
	lines = lines[startIdx:]

	// Parse and filter messages
	var messages []IRCMessage
	for _, line := range lines {
		msg, ok := parseLine(line, channel)
		if !ok {
			continue
		}
		if !matchesFilters(msg, search, fromUser, toUser) {
			continue
		}
		messages = append(messages, msg)
	}

	// Output
	if jsonOutput {
		return outputMessagesJSON(messages)
	}
	return outputFormatted(messages)
}

func followLogs(logFile, channel, search, fromUser, toUser string, jsonOutput bool) error {
	// Initial read of existing content
	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Seek to end to only show new messages
	_, err = file.Seek(0, 2) // Seek to end
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to seek: %w", err)
	}

	// Set up watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(logFile)
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to watch file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Watching #%s for new messages (Ctrl+C to exit)...\n", channel)

	scanner := bufio.NewScanner(file)

	// Process events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				file.Close()
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Read new lines
				for scanner.Scan() {
					line := scanner.Text()
					msg, ok := parseLine(line, channel)
					if !ok {
						continue
					}
					if !matchesFilters(msg, search, fromUser, toUser) {
						continue
					}
					if jsonOutput {
						outputMessagesJSON([]IRCMessage{msg})
					} else {
						printMessage(msg)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				file.Close()
				return nil
			}
			fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
		}
	}
}

func parseLine(line, channel string) (IRCMessage, bool) {
	matches := ircLogRegex.FindStringSubmatch(line)
	if matches == nil {
		return IRCMessage{}, false
	}

	timestamp, err := time.Parse("2006-01-02 15:04:05", matches[1])
	if err != nil {
		return IRCMessage{}, false
	}

	return IRCMessage{
		Timestamp: timestamp,
		Channel:   channel,
		Sender:    matches[2],
		Text:      matches[3],
	}, true
}

func matchesFilters(msg IRCMessage, search, fromUser, toUser string) bool {
	// Search filter
	if search != "" && !strings.Contains(strings.ToLower(msg.Text), strings.ToLower(search)) {
		return false
	}

	// From user filter
	if fromUser != "" {
		fromUser = strings.TrimPrefix(fromUser, "@")
		if !strings.EqualFold(msg.Sender, fromUser) {
			return false
		}
	}

	// To user filter (check if message mentions the user)
	if toUser != "" {
		toUser = strings.TrimPrefix(toUser, "@")
		if !strings.Contains(strings.ToLower(msg.Text), "@"+strings.ToLower(toUser)) &&
			!strings.Contains(strings.ToLower(msg.Text), strings.ToLower(toUser)+":") {
			return false
		}
	}

	return true
}

func outputMessagesJSON(messages []IRCMessage) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(messages)
}

func outputFormatted(messages []IRCMessage) error {
	for _, msg := range messages {
		printMessage(msg)
	}
	return nil
}

func printMessage(msg IRCMessage) {
	// Styles for formatted output
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	senderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	textStyle := lipgloss.NewStyle()

	timestamp := timeStyle.Render(msg.Timestamp.Format("15:04:05"))
	sender := senderStyle.Render("<" + msg.Sender + ">")
	text := textStyle.Render(msg.Text)

	fmt.Printf("%s %s %s\n", timestamp, sender, text)
}
