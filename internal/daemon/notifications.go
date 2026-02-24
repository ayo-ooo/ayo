package daemon

import (
	"database/sql"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// NotificationStatus represents the status of a triggered job.
type NotificationStatus string

const (
	NotificationStatusSuccess NotificationStatus = "success"
	NotificationStatusFailed  NotificationStatus = "failed"
)

// Notification represents a stored notification.
type Notification struct {
	ID          int64              `json:"id"`
	TriggerName string             `json:"trigger_name"`
	CreatedAt   time.Time          `json:"created_at"`
	ReadAt      *time.Time         `json:"read_at,omitempty"`
	Status      NotificationStatus `json:"status"`
	Summary     string             `json:"summary,omitempty"`
	OutputPath  string             `json:"output_path,omitempty"`
}

// NotificationConfig configures the notification service.
type NotificationConfig struct {
	Terminal bool `json:"terminal"` // Show in-terminal notifications
	System   bool `json:"system"`   // Use system notifications (macOS)
	Store    bool `json:"store"`    // Store for later viewing
}

// DefaultNotificationConfig returns the default configuration.
func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{
		Terminal: true,
		System:   false,
		Store:    true,
	}
}

// NotificationService handles trigger completion notifications.
type NotificationService struct {
	db     *sql.DB
	config NotificationConfig
	mu     sync.RWMutex
	dbPath string
}

// DefaultNotificationDBPath returns the default path for the notifications database.
func DefaultNotificationDBPath() string {
	return filepath.Join(paths.DataDir(), "ayo.db")
}

// NewNotificationService creates a new notification service.
func NewNotificationService(dbPath string, config NotificationConfig) (*NotificationService, error) {
	if dbPath == "" {
		dbPath = DefaultNotificationDBPath()
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	ns := &NotificationService{
		db:     db,
		config: config,
		dbPath: dbPath,
	}

	if err := ns.init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	return ns, nil
}

// init creates the database schema.
func (ns *NotificationService) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		trigger_name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		read_at TIMESTAMP,
		status TEXT NOT NULL,
		summary TEXT,
		output_path TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
	CREATE INDEX IF NOT EXISTS idx_notifications_read_at ON notifications(read_at);
	`

	_, err := ns.db.Exec(schema)
	return err
}

// Notify handles a trigger completion notification.
func (ns *NotificationService) Notify(triggerName string, status NotificationStatus, summary string, outputPath string) error {
	// Store in DB
	if ns.config.Store {
		if err := ns.store(triggerName, status, summary, outputPath); err != nil {
			return fmt.Errorf("store notification: %w", err)
		}
	}

	// System notification (macOS)
	if ns.config.System {
		ns.systemNotify(triggerName, status, summary)
	}

	return nil
}

// store saves a notification to the database.
func (ns *NotificationService) store(triggerName string, status NotificationStatus, summary string, outputPath string) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	_, err := ns.db.Exec(`
		INSERT INTO notifications (trigger_name, status, summary, output_path)
		VALUES (?, ?, ?, ?)
	`, triggerName, status, summary, outputPath)

	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}

	// Keep only last 100 notifications
	_, _ = ns.db.Exec(`
		DELETE FROM notifications WHERE id NOT IN (
			SELECT id FROM notifications ORDER BY created_at DESC LIMIT 100
		)
	`)

	return nil
}

// systemNotify sends a system notification (macOS).
func (ns *NotificationService) systemNotify(triggerName string, status NotificationStatus, summary string) {
	if runtime.GOOS != "darwin" {
		return
	}

	statusIcon := "✓"
	if status == NotificationStatusFailed {
		statusIcon = "✗"
	}

	title := fmt.Sprintf("%s %s", statusIcon, triggerName)
	message := summary
	if message == "" {
		if status == NotificationStatusSuccess {
			message = "Completed successfully"
		} else {
			message = "Failed"
		}
	}

	// Use osascript for macOS notifications
	script := fmt.Sprintf(`display notification %q with title "ayo" subtitle %q`, message, title)
	exec.Command("osascript", "-e", script).Run()
}

// GetUnread returns all unread notifications.
func (ns *NotificationService) GetUnread() ([]*Notification, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	rows, err := ns.db.Query(`
		SELECT id, trigger_name, created_at, read_at, status, summary, output_path
		FROM notifications WHERE read_at IS NULL ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		n := &Notification{}
		if err := rows.Scan(&n.ID, &n.TriggerName, &n.CreatedAt, &n.ReadAt, &n.Status, &n.Summary, &n.OutputPath); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

// GetAll returns all notifications.
func (ns *NotificationService) GetAll(limit int) ([]*Notification, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	rows, err := ns.db.Query(`
		SELECT id, trigger_name, created_at, read_at, status, summary, output_path
		FROM notifications ORDER BY created_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		n := &Notification{}
		if err := rows.Scan(&n.ID, &n.TriggerName, &n.CreatedAt, &n.ReadAt, &n.Status, &n.Summary, &n.OutputPath); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

// Get returns a specific notification by ID.
func (ns *NotificationService) Get(id int64) (*Notification, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	n := &Notification{}
	err := ns.db.QueryRow(`
		SELECT id, trigger_name, created_at, read_at, status, summary, output_path
		FROM notifications WHERE id = ?
	`, id).Scan(&n.ID, &n.TriggerName, &n.CreatedAt, &n.ReadAt, &n.Status, &n.Summary, &n.OutputPath)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("notification not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("query notification: %w", err)
	}

	return n, nil
}

// MarkRead marks a notification as read.
func (ns *NotificationService) MarkRead(id int64) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	_, err := ns.db.Exec("UPDATE notifications SET read_at = ? WHERE id = ?", time.Now(), id)
	return err
}

// MarkAllRead marks all notifications as read.
func (ns *NotificationService) MarkAllRead() error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	_, err := ns.db.Exec("UPDATE notifications SET read_at = ? WHERE read_at IS NULL", time.Now())
	return err
}

// UnreadCount returns the number of unread notifications.
func (ns *NotificationService) UnreadCount() (int, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	var count int
	err := ns.db.QueryRow("SELECT COUNT(*) FROM notifications WHERE read_at IS NULL").Scan(&count)
	return count, err
}

// Close closes the database connection.
func (ns *NotificationService) Close() error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if ns.db == nil {
		return nil
	}

	err := ns.db.Close()
	ns.db = nil
	return err
}
