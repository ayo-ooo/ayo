package tickets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Service provides ticket operations for a base directory.
type Service struct {
	baseDir    string // e.g., ~/.local/share/ayo/sessions
	directMode bool   // If true, baseDir IS the tickets directory (no sessionID nesting)
}

// NewService creates a new ticket service that organizes tickets under:
// {baseDir}/{sessionID}/.tickets/{ticketID}.md
func NewService(baseDir string) *Service {
	return &Service{baseDir: baseDir, directMode: false}
}

// NewDirectService creates a ticket service that stores tickets directly in:
// {baseDir}/{ticketID}.md
// This is useful for planner state directories where no session nesting is needed.
func NewDirectService(baseDir string) *Service {
	return &Service{baseDir: baseDir, directMode: true}
}

// ticketsDir returns the .tickets directory for a session.
func (s *Service) ticketsDir(sessionID string) string {
	if s.directMode {
		// In direct mode, baseDir IS the tickets directory
		return s.baseDir
	}
	return filepath.Join(s.baseDir, sessionID, ".tickets")
}

// BaseDir returns the base directory path used by this service.
func (s *Service) BaseDir() string {
	return s.baseDir
}

// IsDirectMode returns true if the service operates in direct mode.
func (s *Service) IsDirectMode() bool {
	return s.directMode
}

// ticketPath returns the file path for a ticket.
func (s *Service) ticketPath(sessionID, ticketID string) string {
	return filepath.Join(s.ticketsDir(sessionID), ticketID+".md")
}

// Create creates a new ticket and returns it.
func (s *Service) Create(sessionID string, opts CreateOptions) (*Ticket, error) {
	ticketsDir := s.ticketsDir(sessionID)

	// Ensure directory exists
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		return nil, fmt.Errorf("create tickets directory: %w", err)
	}

	// Generate unique ID
	id, err := GenerateUniqueID(ticketsDir)
	if err != nil {
		return nil, fmt.Errorf("generate ticket ID: %w", err)
	}

	// Set defaults
	ticketType := opts.Type
	if ticketType == "" {
		ticketType = TypeTask
	}
	priority := opts.Priority
	if !ValidatePriority(priority) {
		priority = DefaultPriority
	}

	now := time.Now().UTC()
	ticket := &Ticket{
		ID:          id,
		Status:      StatusOpen,
		Type:        ticketType,
		Priority:    priority,
		Assignee:    opts.Assignee,
		Deps:        opts.Deps,
		Links:       []string{},
		Parent:      opts.Parent,
		Tags:        opts.Tags,
		Created:     now,
		Session:     sessionID,
		ExternalRef: opts.ExternalRef,
		Title:       opts.Title,
		Description: opts.Description,
		FilePath:    s.ticketPath(sessionID, id),
	}

	// Ensure slices are not nil
	if ticket.Deps == nil {
		ticket.Deps = []string{}
	}
	if ticket.Tags == nil {
		ticket.Tags = []string{}
	}

	// Write to file
	if err := s.write(ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

// Get retrieves a ticket by ID.
func (s *Service) Get(sessionID, ticketID string) (*Ticket, error) {
	// Support partial ID matching
	fullID, err := s.resolveID(sessionID, ticketID)
	if err != nil {
		return nil, err
	}

	path := s.ticketPath(sessionID, fullID)
	ticket, err := Parse(path)
	if err != nil {
		return nil, fmt.Errorf("parse ticket %s: %w", fullID, err)
	}

	return ticket, nil
}

// Update writes changes to an existing ticket.
func (s *Service) Update(ticket *Ticket) error {
	if ticket.FilePath == "" {
		return fmt.Errorf("ticket has no file path")
	}

	return s.write(ticket)
}

// Delete removes a ticket file.
func (s *Service) Delete(sessionID, ticketID string) error {
	fullID, err := s.resolveID(sessionID, ticketID)
	if err != nil {
		return err
	}

	path := s.ticketPath(sessionID, fullID)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete ticket %s: %w", fullID, err)
	}

	return nil
}

// write serializes and writes a ticket to disk.
func (s *Service) write(ticket *Ticket) error {
	content, err := Serialize(ticket)
	if err != nil {
		return fmt.Errorf("serialize ticket: %w", err)
	}

	if err := os.WriteFile(ticket.FilePath, content, 0644); err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}

// Start sets a ticket's status to in_progress.
func (s *Service) Start(sessionID, ticketID string) error {
	ticket, err := s.Get(sessionID, ticketID)
	if err != nil {
		return err
	}

	if ticket.Status != StatusOpen && ticket.Status != StatusBlocked {
		return fmt.Errorf("cannot start ticket with status %s (must be open or blocked)", ticket.Status)
	}

	now := time.Now().UTC()
	ticket.Status = StatusInProgress
	ticket.Started = &now

	return s.Update(ticket)
}

// Close sets a ticket's status to closed.
func (s *Service) Close(sessionID, ticketID string) error {
	ticket, err := s.Get(sessionID, ticketID)
	if err != nil {
		return err
	}

	if ticket.Status == StatusClosed {
		return fmt.Errorf("ticket is already closed")
	}

	now := time.Now().UTC()
	ticket.Status = StatusClosed
	ticket.Closed = &now

	return s.Update(ticket)
}

// Reopen sets a closed ticket's status back to open.
func (s *Service) Reopen(sessionID, ticketID string) error {
	ticket, err := s.Get(sessionID, ticketID)
	if err != nil {
		return err
	}

	if ticket.Status != StatusClosed {
		return fmt.Errorf("cannot reopen ticket with status %s (must be closed)", ticket.Status)
	}

	ticket.Status = StatusOpen
	ticket.Closed = nil

	return s.Update(ticket)
}

// Block sets a ticket's status to blocked.
func (s *Service) Block(sessionID, ticketID string) error {
	ticket, err := s.Get(sessionID, ticketID)
	if err != nil {
		return err
	}

	if ticket.Status == StatusClosed {
		return fmt.Errorf("cannot block a closed ticket")
	}

	ticket.Status = StatusBlocked

	return s.Update(ticket)
}

// List returns all tickets matching the filter.
func (s *Service) List(sessionID string, filter Filter) ([]*Ticket, error) {
	ticketsDir := s.ticketsDir(sessionID)

	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Ticket{}, nil
		}
		return nil, fmt.Errorf("read tickets directory: %w", err)
	}

	var tickets []*Ticket
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(ticketsDir, entry.Name())
		ticket, err := Parse(path)
		if err != nil {
			continue // Skip unparseable files
		}

		if matchesFilter(ticket, filter) {
			tickets = append(tickets, ticket)
		}
	}

	return tickets, nil
}

// Ready returns tickets that are ready to work on (deps resolved).
func (s *Service) Ready(sessionID, assignee string) ([]*Ticket, error) {
	// Get all tickets to build status map
	allTickets, err := s.List(sessionID, Filter{})
	if err != nil {
		return nil, err
	}

	// Build map of ticket ID -> status
	statusMap := make(map[string]Status)
	for _, t := range allTickets {
		statusMap[t.ID] = t.Status
	}

	var ready []*Ticket
	for _, ticket := range allTickets {
		// Only consider open or in_progress tickets
		if ticket.Status != StatusOpen && ticket.Status != StatusInProgress {
			continue
		}

		// Filter by assignee if specified
		if assignee != "" && ticket.Assignee != assignee {
			continue
		}

		// Check if all deps are closed
		allDepsResolved := true
		for _, depID := range ticket.Deps {
			if status, ok := statusMap[depID]; !ok || status != StatusClosed {
				allDepsResolved = false
				break
			}
		}

		if allDepsResolved {
			ready = append(ready, ticket)
		}
	}

	return ready, nil
}

// Blocked returns tickets that are blocked on unresolved dependencies.
func (s *Service) Blocked(sessionID, assignee string) ([]*Ticket, error) {
	// Get all tickets to build status map
	allTickets, err := s.List(sessionID, Filter{})
	if err != nil {
		return nil, err
	}

	// Build map of ticket ID -> status
	statusMap := make(map[string]Status)
	for _, t := range allTickets {
		statusMap[t.ID] = t.Status
	}

	var blocked []*Ticket
	for _, ticket := range allTickets {
		// Only consider open or in_progress tickets
		if ticket.Status != StatusOpen && ticket.Status != StatusInProgress && ticket.Status != StatusBlocked {
			continue
		}

		// Filter by assignee if specified
		if assignee != "" && ticket.Assignee != assignee {
			continue
		}

		// Check if any dep is not closed
		hasUnresolvedDep := false
		for _, depID := range ticket.Deps {
			if status, ok := statusMap[depID]; !ok || status != StatusClosed {
				hasUnresolvedDep = true
				break
			}
		}

		if hasUnresolvedDep {
			blocked = append(blocked, ticket)
		}
	}

	return blocked, nil
}

// matchesFilter checks if a ticket matches the given filter criteria.
func matchesFilter(ticket *Ticket, filter Filter) bool {
	if filter.Status != "" && ticket.Status != filter.Status {
		return false
	}
	if filter.Assignee != "" && ticket.Assignee != filter.Assignee {
		return false
	}
	if filter.Type != "" && ticket.Type != filter.Type {
		return false
	}
	if filter.Priority != nil && ticket.Priority != *filter.Priority {
		return false
	}
	if filter.Parent != "" && ticket.Parent != filter.Parent {
		return false
	}
	if len(filter.Tags) > 0 {
		// Ticket must have all filter tags
		ticketTags := make(map[string]bool)
		for _, tag := range ticket.Tags {
			ticketTags[tag] = true
		}
		for _, tag := range filter.Tags {
			if !ticketTags[tag] {
				return false
			}
		}
	}
	return true
}

// AddDep adds a dependency to a ticket.
func (s *Service) AddDep(sessionID, ticketID, depID string) error {
	ticket, err := s.Get(sessionID, ticketID)
	if err != nil {
		return err
	}

	// Verify dep ticket exists
	depFullID, err := s.resolveID(sessionID, depID)
	if err != nil {
		return fmt.Errorf("dependency %s: %w", depID, err)
	}

	// Check for self-dependency
	if ticket.ID == depFullID {
		return fmt.Errorf("ticket cannot depend on itself")
	}

	// Check if already a dependency
	for _, existing := range ticket.Deps {
		if existing == depFullID {
			return nil // Already exists, no-op
		}
	}

	// Add the dependency
	ticket.Deps = append(ticket.Deps, depFullID)

	// Write to disk first so cycle check can read it
	if err := s.Update(ticket); err != nil {
		return err
	}

	// Check for cycles
	cycles, err := s.FindCycles(sessionID)
	if err != nil {
		return err
	}

	// If there's a cycle involving this ticket, rollback
	for _, cycle := range cycles {
		for _, id := range cycle {
			if id == ticket.ID {
				// Rollback: remove the dep
				ticket.Deps = ticket.Deps[:len(ticket.Deps)-1]
				s.Update(ticket)
				return fmt.Errorf("dependency would create cycle: %v", cycle)
			}
		}
	}

	return nil
}

// RemoveDep removes a dependency from a ticket.
func (s *Service) RemoveDep(sessionID, ticketID, depID string) error {
	ticket, err := s.Get(sessionID, ticketID)
	if err != nil {
		return err
	}

	depFullID, err := s.resolveID(sessionID, depID)
	if err != nil {
		// If dep doesn't exist, just try to remove by partial match
		depFullID = depID
	}

	// Find and remove the dependency
	newDeps := make([]string, 0, len(ticket.Deps))
	found := false
	for _, d := range ticket.Deps {
		if d == depFullID || strings.Contains(d, depID) {
			found = true
		} else {
			newDeps = append(newDeps, d)
		}
	}

	if !found {
		return fmt.Errorf("dependency %s not found", depID)
	}

	ticket.Deps = newDeps
	return s.Update(ticket)
}

// DepTree represents a dependency tree node.
type DepTree struct {
	Ticket   *Ticket
	Children []*DepTree
}

// DepTree returns the dependency tree for a ticket.
func (s *Service) DepTree(sessionID, ticketID string) (*DepTree, error) {
	visited := make(map[string]bool)
	return s.buildDepTree(sessionID, ticketID, visited)
}

func (s *Service) buildDepTree(sessionID, ticketID string, visited map[string]bool) (*DepTree, error) {
	ticket, err := s.Get(sessionID, ticketID)
	if err != nil {
		return nil, err
	}

	tree := &DepTree{Ticket: ticket}

	if visited[ticket.ID] {
		return tree, nil // Already visited, avoid infinite loop
	}
	visited[ticket.ID] = true

	for _, depID := range ticket.Deps {
		childTree, err := s.buildDepTree(sessionID, depID, visited)
		if err != nil {
			continue // Skip missing deps
		}
		tree.Children = append(tree.Children, childTree)
	}

	return tree, nil
}

// FindCycles detects dependency cycles in all tickets.
func (s *Service) FindCycles(sessionID string) ([][]string, error) {
	tickets, err := s.List(sessionID, Filter{})
	if err != nil {
		return nil, err
	}

	// Build adjacency list
	deps := make(map[string][]string)
	for _, t := range tickets {
		deps[t.ID] = t.Deps
	}

	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(id string) bool
	dfs = func(id string) bool {
		visited[id] = true
		recStack[id] = true
		path = append(path, id)

		for _, dep := range deps[id] {
			if !visited[dep] {
				if dfs(dep) {
					return true
				}
			} else if recStack[dep] {
				// Found cycle - extract it from path
				cycleStart := -1
				for i, p := range path {
					if p == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[id] = false
		return false
	}

	for _, t := range tickets {
		if !visited[t.ID] {
			dfs(t.ID)
		}
	}

	return cycles, nil
}

// checkCycle checks if adding deps creates a cycle starting from ticketID.
func (s *Service) checkCycle(sessionID, ticketID string) error {
	cycles, err := s.FindCycles(sessionID)
	if err != nil {
		return err
	}

	for _, cycle := range cycles {
		for _, id := range cycle {
			if id == ticketID {
				return fmt.Errorf("dependency would create cycle: %v", cycle)
			}
		}
	}

	return nil
}

// AddNote appends a timestamped note to a ticket.
func (s *Service) AddNote(sessionID, ticketID, content string) error {
	ticket, err := s.Get(sessionID, ticketID)
	if err != nil {
		return err
	}

	note := Note{
		Timestamp: time.Now().UTC(),
		Content:   strings.TrimSpace(content),
	}

	ticket.Notes = append(ticket.Notes, note)

	return s.Update(ticket)
}

// Assign sets the assignee for a ticket.
func (s *Service) Assign(sessionID, ticketID, assignee string) error {
	ticket, err := s.Get(sessionID, ticketID)
	if err != nil {
		return err
	}

	ticket.Assignee = assignee

	return s.Update(ticket)
}

// Unassign removes the assignee from a ticket.
func (s *Service) Unassign(sessionID, ticketID string) error {
	ticket, err := s.Get(sessionID, ticketID)
	if err != nil {
		return err
	}

	ticket.Assignee = ""

	return s.Update(ticket)
}

// resolveID finds the full ticket ID from a partial match.
func (s *Service) resolveID(sessionID, partialID string) (string, error) {
	ticketsDir := s.ticketsDir(sessionID)

	// Try exact match first
	exactPath := filepath.Join(ticketsDir, partialID+".md")
	if _, err := os.Stat(exactPath); err == nil {
		return partialID, nil
	}

	// Search for partial matches
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		return "", fmt.Errorf("read tickets directory: %w", err)
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".md")
		if strings.Contains(id, partialID) {
			matches = append(matches, id)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("ticket not found: %s", partialID)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous ticket ID %s: matches %v", partialID, matches)
	}
}
