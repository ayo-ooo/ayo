package squads

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/tickets"
	"github.com/fsnotify/fsnotify"
)

// ProgressUpdate represents a ticket status change.
type ProgressUpdate struct {
	TicketID  string    `json:"ticket_id"`
	Status    string    `json:"status"`
	Assignee  string    `json:"assignee,omitempty"`
	Title     string    `json:"title,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ProgressCallback is called when a ticket status changes.
type ProgressCallback func(update ProgressUpdate)

// ProgressStreamer watches squad tickets and streams progress updates.
type ProgressStreamer struct {
	squadName string
	callback  ProgressCallback
	watcher   *fsnotify.Watcher

	mu           sync.RWMutex
	ticketStates map[string]string // ticketID -> last known status
	running      bool

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewProgressStreamer creates a new progress streamer for a squad.
func NewProgressStreamer(squadName string, callback ProgressCallback) (*ProgressStreamer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &ProgressStreamer{
		squadName:    squadName,
		callback:     callback,
		watcher:      watcher,
		ticketStates: make(map[string]string),
		stopCh:       make(chan struct{}),
	}, nil
}

// Start starts watching for ticket changes.
func (ps *ProgressStreamer) Start(ctx context.Context) error {
	ps.mu.Lock()
	if ps.running {
		ps.mu.Unlock()
		return nil
	}
	ps.running = true
	ps.mu.Unlock()

	ticketsDir := paths.SquadTicketsDir(ps.squadName)
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		return err
	}

	// Load initial state
	if err := ps.loadInitialState(); err != nil {
		debug.Log("error loading initial state", "error", err)
	}

	// Watch tickets directory
	if err := ps.watcher.Add(ticketsDir); err != nil {
		return err
	}

	ps.wg.Add(1)
	go ps.eventLoop(ctx)

	debug.Log("started progress streamer", "squad", ps.squadName)
	return nil
}

// Stop stops watching.
func (ps *ProgressStreamer) Stop() error {
	ps.mu.Lock()
	if !ps.running {
		ps.mu.Unlock()
		return nil
	}
	ps.running = false
	ps.mu.Unlock()

	close(ps.stopCh)
	ps.wg.Wait()
	return ps.watcher.Close()
}

// loadInitialState loads current ticket states.
func (ps *ProgressStreamer) loadInitialState() error {
	ticketsDir := paths.SquadTicketsDir(ps.squadName)
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			path := filepath.Join(ticketsDir, entry.Name())
			ticket, err := tickets.Parse(path)
			if err != nil {
				continue
			}
			ps.ticketStates[ticket.ID] = string(ticket.Status)
		}
	}

	return nil
}

// eventLoop handles fsnotify events.
func (ps *ProgressStreamer) eventLoop(ctx context.Context) {
	defer ps.wg.Done()

	for {
		select {
		case <-ps.stopCh:
			return
		case <-ctx.Done():
			return

		case event, ok := <-ps.watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
				ps.handleTicketChange(event.Name)
			}

		case err, ok := <-ps.watcher.Errors:
			if !ok {
				return
			}
			debug.Log("watcher error", "error", err)
		}
	}
}

// handleTicketChange processes a ticket file change.
func (ps *ProgressStreamer) handleTicketChange(path string) {
	if !strings.HasSuffix(path, ".md") {
		return
	}

	ticket, err := tickets.Parse(path)
	if err != nil {
		return
	}

	ps.mu.Lock()
	previousStatus := ps.ticketStates[ticket.ID]
	currentStatus := string(ticket.Status)
	ps.ticketStates[ticket.ID] = currentStatus
	ps.mu.Unlock()

	// Only notify on status changes
	if previousStatus != currentStatus {
		update := ProgressUpdate{
			TicketID:  ticket.ID,
			Status:    currentStatus,
			Assignee:  ticket.Assignee,
			Title:     ticket.Title,
			Timestamp: time.Now(),
		}

		if ps.callback != nil {
			ps.callback(update)
		}

		debug.Log("ticket progress update", "ticket", ticket.ID, "status", currentStatus)
	}
}

// GetProgress returns current ticket states for a squad.
func GetProgress(squadName string) (map[string]string, error) {
	ticketsDir := paths.SquadTicketsDir(squadName)
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	states := make(map[string]string)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			path := filepath.Join(ticketsDir, entry.Name())
			ticket, err := tickets.Parse(path)
			if err != nil {
				continue
			}
			states[ticket.ID] = string(ticket.Status)
		}
	}

	return states, nil
}

// StreamProgress is a convenience function to stream progress updates.
func StreamProgress(ctx context.Context, squadName string, callback ProgressCallback) error {
	ps, err := NewProgressStreamer(squadName, callback)
	if err != nil {
		return err
	}

	if err := ps.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return ps.Stop()
}
