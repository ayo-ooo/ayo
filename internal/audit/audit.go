// Package audit provides file modification audit logging for security and accountability.
package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
)

// Entry represents a single audit log entry.
type Entry struct {
	Timestamp   time.Time `json:"ts"`
	Agent       string    `json:"agent"`
	Session     string    `json:"session"`
	Action      string    `json:"action"` // "create", "update", "delete"
	Path        string    `json:"path"`
	Approval    string    `json:"approval"` // How approval was obtained
	Size        int64     `json:"size,omitempty"`
	ContentHash string    `json:"hash,omitempty"`
}

// ApprovalType constants for how file modifications were approved.
const (
	ApprovalUserApproved  = "user_approved"   // User pressed Y in prompt
	ApprovalSessionCache  = "session_cache"   // User pressed A earlier in session
	ApprovalNoJodas       = "no_jodas"        // --no-jodas flag was used
	ApprovalAgentConfig   = "agent_config"    // Agent has auto_approve: true
	ApprovalGlobalConfig  = "global_config"   // Global no_jodas setting
)

// ActionType constants for file operations.
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// Logger provides audit logging functionality.
type Logger interface {
	Log(entry Entry) error
	Query(filter Filter) ([]Entry, error)
	Close() error
}

// Filter specifies criteria for querying audit entries.
type Filter struct {
	Agent   string
	Session string
	Action  string
	Path    string
	Since   time.Time
	Until   time.Time
	Limit   int
}

// FileLogger implements Logger using a JSON Lines file.
type FileLogger struct {
	path       string
	file       *os.File
	mu         sync.Mutex
	maxSize    int64 // Max size before rotation
	maxBackups int   // Number of backup files to keep
}

// DefaultMaxSize is the default max log size (10MB).
const DefaultMaxSize = 10 * 1024 * 1024

// DefaultMaxBackups is the default number of backup files to keep.
const DefaultMaxBackups = 3

// NewFileLogger creates a new file-based audit logger.
func NewFileLogger() (*FileLogger, error) {
	return NewFileLoggerWithPath(filepath.Join(paths.DataDir(), "audit.log"))
}

// NewFileLoggerWithPath creates a logger with a specific path.
func NewFileLoggerWithPath(path string) (*FileLogger, error) {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create audit log directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open audit log: %w", err)
	}

	return &FileLogger{
		path:       path,
		file:       file,
		maxSize:    DefaultMaxSize,
		maxBackups: DefaultMaxBackups,
	}, nil
}

// Log writes an audit entry.
func (l *FileLogger) Log(entry Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if rotation needed
	if err := l.rotateIfNeeded(); err != nil {
		return fmt.Errorf("rotate audit log: %w", err)
	}

	// Write JSON line
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}

	if _, err := l.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit entry: %w", err)
	}

	return nil
}

// rotateIfNeeded rotates the log file if it exceeds maxSize.
func (l *FileLogger) rotateIfNeeded() error {
	info, err := l.file.Stat()
	if err != nil {
		return err
	}

	if info.Size() < l.maxSize {
		return nil
	}

	// Close current file
	if err := l.file.Close(); err != nil {
		return err
	}

	// Rotate backup files
	for i := l.maxBackups - 1; i >= 1; i-- {
		old := fmt.Sprintf("%s.%d", l.path, i)
		new := fmt.Sprintf("%s.%d", l.path, i+1)
		os.Rename(old, new) // Ignore errors for missing files
	}

	// Move current to .1
	if err := os.Rename(l.path, l.path+".1"); err != nil {
		return err
	}

	// Remove oldest if over limit
	oldest := fmt.Sprintf("%s.%d", l.path, l.maxBackups)
	os.Remove(oldest) // Ignore errors

	// Reopen log file
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	l.file = file

	return nil
}

// Query reads entries matching the filter.
func (l *FileLogger) Query(filter Filter) ([]Entry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Flush current file
	l.file.Sync()

	// Open for reading
	file, err := os.Open(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open audit log for reading: %w", err)
	}
	defer file.Close()

	return l.queryFile(file, filter)
}

func (l *FileLogger) queryFile(r io.Reader, filter Filter) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed entries
		}

		if l.matchesFilter(entry, filter) {
			entries = append(entries, entry)
			if filter.Limit > 0 && len(entries) >= filter.Limit {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan audit log: %w", err)
	}

	return entries, nil
}

func (l *FileLogger) matchesFilter(entry Entry, filter Filter) bool {
	if filter.Agent != "" && entry.Agent != filter.Agent {
		return false
	}
	if filter.Session != "" && entry.Session != filter.Session {
		return false
	}
	if filter.Action != "" && entry.Action != filter.Action {
		return false
	}
	if filter.Path != "" && entry.Path != filter.Path {
		return false
	}
	if !filter.Since.IsZero() && entry.Timestamp.Before(filter.Since) {
		return false
	}
	if !filter.Until.IsZero() && entry.Timestamp.After(filter.Until) {
		return false
	}
	return true
}

// Close closes the audit log file.
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// Global logger instance
var (
	defaultLogger Logger
	loggerMu      sync.Mutex
)

// GetLogger returns the global audit logger, creating it if needed.
func GetLogger() (Logger, error) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	if defaultLogger == nil {
		logger, err := NewFileLogger()
		if err != nil {
			return nil, err
		}
		defaultLogger = logger
	}
	return defaultLogger, nil
}

// Log writes an entry to the global audit logger.
func Log(entry Entry) error {
	logger, err := GetLogger()
	if err != nil {
		return err
	}
	return logger.Log(entry)
}

// LogFileModification is a convenience function for logging file modifications.
func LogFileModification(agent, session, action, path, approval string, size int64, hash string) error {
	return Log(Entry{
		Timestamp:   time.Now().UTC(),
		Agent:       agent,
		Session:     session,
		Action:      action,
		Path:        path,
		Approval:    approval,
		Size:        size,
		ContentHash: hash,
	})
}
