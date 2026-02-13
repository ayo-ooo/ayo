package daemon

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/tickets"
	"github.com/fsnotify/fsnotify"
)

// AgentRunner is the interface for spawning agents with tickets.
type AgentRunner interface {
	// RunWithTicket starts an agent to work on a specific ticket.
	// The agent receives the ticket context and should work autonomously.
	RunWithTicket(ctx context.Context, agentHandle, sessionID, ticketID string) error
}

// TicketClosedHandler is called when a ticket is closed.
type TicketClosedHandler func(contextID, ticketID string, isSquad bool)

// TicketWatcher watches ticket directories and spawns agents when tickets are assigned.
type TicketWatcher struct {
	watcher *fsnotify.Watcher
	service *tickets.Service
	runner  AgentRunner

	// Callback for ticket closures
	onTicketClosed TicketClosedHandler

	// Track watched sessions and running agents
	mu              sync.RWMutex
	watchedSessions map[string]bool             // sessionID -> watched
	watchedSquads   map[string]bool             // squadName -> watched
	runningAgents   map[string]context.CancelFunc // ticketID -> cancel func

	// Track previous ticket states for change detection
	ticketStates    map[string]tickets.Status     // ticketID -> last known status

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// TicketWatcherConfig configures the ticket watcher.
type TicketWatcherConfig struct {
	Runner         AgentRunner
	OnTicketClosed TicketClosedHandler
}

// NewTicketWatcher creates a new ticket watcher.
func NewTicketWatcher(cfg TicketWatcherConfig) (*TicketWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &TicketWatcher{
		watcher:         watcher,
		service:         tickets.NewService(paths.SessionsDir()),
		runner:          cfg.Runner,
		onTicketClosed:  cfg.OnTicketClosed,
		watchedSessions: make(map[string]bool),
		watchedSquads:   make(map[string]bool),
		runningAgents:   make(map[string]context.CancelFunc),
		ticketStates:    make(map[string]tickets.Status),
		stopCh:          make(chan struct{}),
	}, nil
}

// Start starts watching for ticket changes.
func (tw *TicketWatcher) Start(ctx context.Context) error {
	// Watch the sessions directory for new session directories
	sessionsDir := paths.SessionsDir()
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return err
	}

	if err := tw.watcher.Add(sessionsDir); err != nil {
		return err
	}

	// Watch existing session ticket directories
	if err := tw.watchExistingSessions(); err != nil {
		return err
	}

	tw.wg.Add(1)
	go tw.eventLoop(ctx)

	return nil
}

// Stop stops the ticket watcher.
func (tw *TicketWatcher) Stop(ctx context.Context) error {
	close(tw.stopCh)

	// Cancel all running agents
	tw.mu.Lock()
	for _, cancel := range tw.runningAgents {
		cancel()
	}
	tw.runningAgents = make(map[string]context.CancelFunc)
	tw.mu.Unlock()

	tw.wg.Wait()
	return tw.watcher.Close()
}

// watchExistingSessions watches ticket directories for existing sessions.
func (tw *TicketWatcher) watchExistingSessions() error {
	sessionsDir := paths.SessionsDir()
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			sessionID := entry.Name()
			tw.watchSession(sessionID)
		}
	}

	return nil
}

// watchSession adds a watch on a session's ticket directory.
func (tw *TicketWatcher) watchSession(sessionID string) error {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.watchedSessions[sessionID] {
		return nil // Already watching
	}

	ticketsDir := paths.SessionTicketsDir(sessionID)
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		return err
	}

	if err := tw.watcher.Add(ticketsDir); err != nil {
		return err
	}

	tw.watchedSessions[sessionID] = true
	return nil
}

// eventLoop handles fsnotify events.
func (tw *TicketWatcher) eventLoop(ctx context.Context) {
	defer tw.wg.Done()

	// Debounce timer for batch processing
	var debounceTimer *time.Timer
	pendingEvents := make(map[string]fsnotify.Event)
	pendingMu := sync.Mutex{}

	processEvents := func() {
		pendingMu.Lock()
		events := pendingEvents
		pendingEvents = make(map[string]fsnotify.Event)
		pendingMu.Unlock()

		for path, event := range events {
			tw.handleEvent(ctx, path, event)
		}
	}

	for {
		select {
		case <-tw.stopCh:
			return
		case <-ctx.Done():
			return

		case event, ok := <-tw.watcher.Events:
			if !ok {
				return
			}

			// Debounce events to the same file
			pendingMu.Lock()
			pendingEvents[event.Name] = event
			pendingMu.Unlock()

			// Reset debounce timer
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(100*time.Millisecond, processEvents)

		case err, ok := <-tw.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue
			_ = err
		}
	}
}

// handleEvent processes a single file system event.
func (tw *TicketWatcher) handleEvent(ctx context.Context, path string, event fsnotify.Event) {
	// Check if this is a new session directory
	sessionsDir := paths.SessionsDir()
	if filepath.Dir(path) == sessionsDir && event.Has(fsnotify.Create) {
		// New session directory created - watch its tickets directory
		sessionID := filepath.Base(path)
		tw.watchSession(sessionID)
		return
	}

	// Check if this is a ticket file
	if !strings.HasSuffix(path, ".md") {
		return
	}

	// Determine if this is a session or squad ticket
	// Session: .../sessions/{sessionID}/.tickets/{ticketID}.md
	// Squad:   .../squads/{squadName}/tickets/{ticketID}.md
	dir := filepath.Dir(path) // .tickets or tickets dir
	parentDir := filepath.Dir(dir)
	grandParentDir := filepath.Dir(parentDir)

	var contextID string
	var isSquad bool

	squadsDir := paths.SquadsDir()
	if grandParentDir == squadsDir {
		// This is a squad ticket
		contextID = filepath.Base(parentDir)
		isSquad = true
	} else {
		// This is a session ticket
		contextID = filepath.Base(parentDir)
		isSquad = false
	}

	switch {
	case event.Has(fsnotify.Create), event.Has(fsnotify.Write):
		tw.handleTicketChange(ctx, contextID, path, isSquad)

	case event.Has(fsnotify.Remove):
		// Ticket was deleted - cancel any running agent
		ticketID := strings.TrimSuffix(filepath.Base(path), ".md")
		tw.cancelAgent(ticketID)
	}
}

// handleTicketChange processes a ticket create/update event.
func (tw *TicketWatcher) handleTicketChange(ctx context.Context, contextID, path string, isSquad bool) {
	// Parse the ticket
	ticket, err := tickets.Parse(path)
	if err != nil {
		return // Unparseable file, skip
	}

	// Check for status change to closed
	tw.mu.Lock()
	previousStatus := tw.ticketStates[ticket.ID]
	tw.ticketStates[ticket.ID] = ticket.Status
	tw.mu.Unlock()

	// If ticket is now closed and wasn't before, notify
	if ticket.Status == tickets.StatusClosed && previousStatus != tickets.StatusClosed {
		// Cancel any running agent
		tw.cancelAgent(ticket.ID)

		// Notify via callback
		if tw.onTicketClosed != nil {
			tw.onTicketClosed(contextID, ticket.ID, isSquad)
		}

		// Check dependents
		tw.CheckDependents(ctx, contextID, ticket.ID)
		return
	}

	// Check if ticket is assigned and ready to work
	if ticket.Assignee == "" {
		return // No assignee
	}

	if ticket.Status != tickets.StatusOpen && ticket.Status != tickets.StatusInProgress {
		// Ticket is blocked or closed - cancel any running agent
		tw.cancelAgent(ticket.ID)
		return
	}

	// Check if dependencies are resolved
	if !tw.areDepsResolved(contextID, ticket) {
		return // Blocked by dependencies
	}

	// Ensure agent is running for this ticket
	tw.ensureAgentRunning(ctx, contextID, ticket)
}

// areDepsResolved checks if all ticket dependencies are closed.
func (tw *TicketWatcher) areDepsResolved(sessionID string, ticket *tickets.Ticket) bool {
	if len(ticket.Deps) == 0 {
		return true
	}

	for _, depID := range ticket.Deps {
		dep, err := tw.service.Get(sessionID, depID)
		if err != nil {
			return false // Can't find dependency - treat as unresolved
		}
		if dep.Status != tickets.StatusClosed {
			return false
		}
	}

	return true
}

// ensureAgentRunning spawns an agent for the ticket if not already running.
func (tw *TicketWatcher) ensureAgentRunning(ctx context.Context, sessionID string, ticket *tickets.Ticket) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	// Check if already running
	if _, running := tw.runningAgents[ticket.ID]; running {
		return
	}

	// No runner configured
	if tw.runner == nil {
		return
	}

	// Start the agent
	agentCtx, cancel := context.WithCancel(ctx)
	tw.runningAgents[ticket.ID] = cancel

	go func() {
		defer func() {
			tw.mu.Lock()
			delete(tw.runningAgents, ticket.ID)
			tw.mu.Unlock()
		}()

		tw.runner.RunWithTicket(agentCtx, ticket.Assignee, sessionID, ticket.ID)
	}()
}

// cancelAgent cancels a running agent for a ticket.
func (tw *TicketWatcher) cancelAgent(ticketID string) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if cancel, ok := tw.runningAgents[ticketID]; ok {
		cancel()
		delete(tw.runningAgents, ticketID)
	}
}

// CheckDependents checks tickets that depend on a closed ticket and spawns agents.
// Called when a ticket is closed to wake up dependent tickets.
func (tw *TicketWatcher) CheckDependents(ctx context.Context, sessionID, closedTicketID string) {
	// List all tickets and find those that depend on the closed ticket
	ticketList, err := tw.service.List(sessionID, tickets.Filter{})
	if err != nil {
		return
	}

	for _, ticket := range ticketList {
		// Check if this ticket depends on the closed one
		dependsOnClosed := false
		for _, dep := range ticket.Deps {
			if dep == closedTicketID {
				dependsOnClosed = true
				break
			}
		}

		if !dependsOnClosed {
			continue
		}

		// Check if ticket is now ready (all deps resolved)
		if ticket.Status == tickets.StatusOpen || ticket.Status == tickets.StatusInProgress {
			if tw.areDepsResolved(sessionID, ticket) {
				tw.ensureAgentRunning(ctx, sessionID, ticket)
			}
		}
	}
}

// RunningAgents returns the number of currently running agents.
func (tw *TicketWatcher) RunningAgents() int {
	tw.mu.RLock()
	defer tw.mu.RUnlock()
	return len(tw.runningAgents)
}

// WatchedSessions returns the number of watched sessions.
func (tw *TicketWatcher) WatchedSessions() int {
	tw.mu.RLock()
	defer tw.mu.RUnlock()
	return len(tw.watchedSessions)
}

// WatchSquad adds a watch on a squad's ticket directory.
func (tw *TicketWatcher) WatchSquad(squadName string) error {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.watchedSquads[squadName] {
		return nil // Already watching
	}

	ticketsDir := paths.SquadTicketsDir(squadName)
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		return err
	}

	if err := tw.watcher.Add(ticketsDir); err != nil {
		return err
	}

	tw.watchedSquads[squadName] = true
	return nil
}

// UnwatchSquad removes a watch from a squad's ticket directory.
func (tw *TicketWatcher) UnwatchSquad(squadName string) error {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if !tw.watchedSquads[squadName] {
		return nil // Not watching
	}

	ticketsDir := paths.SquadTicketsDir(squadName)
	if err := tw.watcher.Remove(ticketsDir); err != nil {
		// Ignore error if directory doesn't exist
		if !os.IsNotExist(err) {
			return err
		}
	}

	delete(tw.watchedSquads, squadName)
	return nil
}

// WatchedSquads returns the number of watched squads.
func (tw *TicketWatcher) WatchedSquads() int {
	tw.mu.RLock()
	defer tw.mu.RUnlock()
	return len(tw.watchedSquads)
}

// IsSquadWatched returns true if the squad is being watched.
func (tw *TicketWatcher) IsSquadWatched(squadName string) bool {
	tw.mu.RLock()
	defer tw.mu.RUnlock()
	return tw.watchedSquads[squadName]
}
