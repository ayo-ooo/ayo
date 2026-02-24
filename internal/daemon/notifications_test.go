package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNotificationService_Notify(t *testing.T) {
	// Create temp dir for test
	tmpDir, err := os.MkdirTemp("", "notification-test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	config := NotificationConfig{
		Terminal: false,
		System:   false,
		Store:    true,
	}

	ns, err := NewNotificationService(dbPath, config)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}
	defer ns.Close()

	// Test storing a notification
	err = ns.Notify("test-trigger", NotificationStatusSuccess, "Test completed", "/output/test.txt")
	if err != nil {
		t.Fatalf("notify: %v", err)
	}

	// Check unread count
	count, err := ns.UnreadCount()
	if err != nil {
		t.Fatalf("unread count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 unread, got %d", count)
	}

	// Check get unread
	unread, err := ns.GetUnread()
	if err != nil {
		t.Fatalf("get unread: %v", err)
	}
	if len(unread) != 1 {
		t.Errorf("expected 1 unread, got %d", len(unread))
	}
	if unread[0].TriggerName != "test-trigger" {
		t.Errorf("expected trigger name 'test-trigger', got %q", unread[0].TriggerName)
	}
	if unread[0].Status != NotificationStatusSuccess {
		t.Errorf("expected status success, got %q", unread[0].Status)
	}
}

func TestNotificationService_MarkRead(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "notification-test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ns, err := NewNotificationService(dbPath, DefaultNotificationConfig())
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}
	defer ns.Close()

	// Create notification
	ns.Notify("test", NotificationStatusSuccess, "Test", "")

	// Get the notification
	unread, _ := ns.GetUnread()
	if len(unread) != 1 {
		t.Fatalf("expected 1 unread, got %d", len(unread))
	}

	// Mark as read
	err = ns.MarkRead(unread[0].ID)
	if err != nil {
		t.Fatalf("mark read: %v", err)
	}

	// Check unread count
	count, _ := ns.UnreadCount()
	if count != 0 {
		t.Errorf("expected 0 unread after marking read, got %d", count)
	}
}

func TestNotificationService_MarkAllRead(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "notification-test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ns, err := NewNotificationService(dbPath, DefaultNotificationConfig())
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}
	defer ns.Close()

	// Create multiple notifications
	ns.Notify("test1", NotificationStatusSuccess, "Test 1", "")
	ns.Notify("test2", NotificationStatusFailed, "Test 2", "")
	ns.Notify("test3", NotificationStatusSuccess, "Test 3", "")

	// Check unread count
	count, _ := ns.UnreadCount()
	if count != 3 {
		t.Fatalf("expected 3 unread, got %d", count)
	}

	// Mark all as read
	err = ns.MarkAllRead()
	if err != nil {
		t.Fatalf("mark all read: %v", err)
	}

	// Check unread count
	count, _ = ns.UnreadCount()
	if count != 0 {
		t.Errorf("expected 0 unread after marking all read, got %d", count)
	}
}

func TestNotificationService_Get(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "notification-test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ns, err := NewNotificationService(dbPath, DefaultNotificationConfig())
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}
	defer ns.Close()

	// Create notification
	ns.Notify("test", NotificationStatusFailed, "Error occurred", "/output/error.log")

	// Get all
	all, _ := ns.GetAll(10)
	if len(all) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(all))
	}

	// Get by ID
	n, err := ns.Get(all[0].ID)
	if err != nil {
		t.Fatalf("get notification: %v", err)
	}
	if n.TriggerName != "test" {
		t.Errorf("expected trigger name 'test', got %q", n.TriggerName)
	}
	if n.Summary != "Error occurred" {
		t.Errorf("expected summary 'Error occurred', got %q", n.Summary)
	}
	if n.OutputPath != "/output/error.log" {
		t.Errorf("expected output path '/output/error.log', got %q", n.OutputPath)
	}
}

func TestNotificationService_NotFoundError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "notification-test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ns, err := NewNotificationService(dbPath, DefaultNotificationConfig())
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}
	defer ns.Close()

	// Try to get non-existent notification
	_, err = ns.Get(9999)
	if err == nil {
		t.Error("expected error for non-existent notification")
	}
}
