package memory

import (
	"sync"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/ui"
)

func TestQueue_Enqueue(t *testing.T) {
	// Create queue with nil service (we won't process)
	q := NewQueue(nil, QueueConfig{BufferSize: 10})

	id := q.Enqueue("test content", CategoryFact, "", "")

	if id == "" {
		t.Error("expected non-empty ID")
	}

	if q.Pending() != 1 {
		t.Errorf("expected 1 pending, got %d", q.Pending())
	}
}

func TestQueue_StatusCallbacks(t *testing.T) {
	var mu sync.Mutex
	var statuses []ui.AsyncStatusMsg

	onStatus := func(msg ui.AsyncStatusMsg) {
		mu.Lock()
		statuses = append(statuses, msg)
		mu.Unlock()
	}

	q := NewQueue(nil, QueueConfig{
		BufferSize: 10,
		OnStatus:   onStatus,
	})

	q.Enqueue("test", CategoryFact, "", "")

	mu.Lock()
	if len(statuses) != 1 {
		t.Errorf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].Status != ui.AsyncStatusPending {
		t.Errorf("expected pending status, got %v", statuses[0].Status)
	}
	mu.Unlock()
}

func TestQueue_BufferFull(t *testing.T) {
	var mu sync.Mutex
	var statuses []ui.AsyncStatusMsg

	onStatus := func(msg ui.AsyncStatusMsg) {
		mu.Lock()
		statuses = append(statuses, msg)
		mu.Unlock()
	}

	// Tiny buffer
	q := NewQueue(nil, QueueConfig{
		BufferSize: 1,
		OnStatus:   onStatus,
	})

	// First should succeed
	q.Enqueue("test1", CategoryFact, "", "")

	// Second should fail (buffer full, no worker draining)
	q.Enqueue("test2", CategoryFact, "", "")

	mu.Lock()
	defer mu.Unlock()

	// Should have: pending, pending, failed
	if len(statuses) < 2 {
		t.Fatalf("expected at least 2 statuses, got %d", len(statuses))
	}

	// Last status should be failed
	lastStatus := statuses[len(statuses)-1]
	if lastStatus.Status != ui.AsyncStatusFailed {
		t.Errorf("expected failed status for overflow, got %v", lastStatus.Status)
	}
}

func TestQueue_StartStop(t *testing.T) {
	q := NewQueue(nil, QueueConfig{BufferSize: 10})

	q.Start()

	// Give worker time to start
	time.Sleep(10 * time.Millisecond)

	// Stop should complete quickly
	done := make(chan struct{})
	go func() {
		q.Stop(1 * time.Second)
		close(done)
	}()

	select {
	case <-done:
		// Good
	case <-time.After(2 * time.Second):
		t.Error("Stop took too long")
	}
}

func TestQueue_DrainOnStop(t *testing.T) {
	// We can't use a real service in unit tests, so we'll test the mechanics
	// by checking that stop waits for the worker to exit

	q := NewQueue(nil, QueueConfig{BufferSize: 10})

	// Add items before starting
	q.Enqueue("test1", CategoryFact, "", "")
	q.Enqueue("test2", CategoryFact, "", "")

	if q.Pending() != 2 {
		t.Errorf("expected 2 pending, got %d", q.Pending())
	}

	// Start will begin draining (but fail since service is nil)
	q.Start()

	// Stop should drain the queue
	q.Stop(1 * time.Second)

	// Queue should be empty after stop
	if q.Pending() != 0 {
		t.Errorf("expected 0 pending after stop, got %d", q.Pending())
	}
}
