package tickets

import (
	"os"

	"github.com/alexcabrera/ayo/internal/paths"
)

// SquadTicketService provides ticket operations for squad directories.
// It wraps the standard Service but uses squad ticket paths.
type SquadTicketService struct {
	*Service
}

// NewSquadTicketService creates a ticket service for squad-based tickets.
// The baseDir should be the squads directory (e.g., ~/.local/share/ayo/sandboxes/squads).
func NewSquadTicketService() *SquadTicketService {
	return &SquadTicketService{
		Service: NewService(paths.SquadsDir()),
	}
}

// SquadTicketsDir returns the .tickets directory for a squad.
func (s *SquadTicketService) SquadTicketsDir(squadName string) string {
	return paths.SquadTicketsDir(squadName)
}

// CreateForSquad creates a ticket in a squad's tickets directory.
func (s *SquadTicketService) CreateForSquad(squadName string, opts CreateOptions) (*Ticket, error) {
	// Override the service's ticketsDir method by using squadName as sessionID
	// The internal path resolution will use paths.SquadTicketsDir
	return s.Create(squadName, opts)
}

// ListForSquad lists tickets in a squad's tickets directory.
func (s *SquadTicketService) ListForSquad(squadName string, filter Filter) ([]*Ticket, error) {
	return s.List(squadName, filter)
}

// GetForSquad gets a ticket from a squad's tickets directory.
func (s *SquadTicketService) GetForSquad(squadName, ticketID string) (*Ticket, error) {
	return s.Get(squadName, ticketID)
}

// ReadyForSquad returns tickets ready to work on in a squad.
func (s *SquadTicketService) ReadyForSquad(squadName, assignee string) ([]*Ticket, error) {
	return s.Ready(squadName, assignee)
}

// EnsureSquadTicketsDir creates the tickets directory for a squad if it doesn't exist.
func EnsureSquadTicketsDir(squadName string) error {
	ticketsDir := paths.SquadTicketsDir(squadName)
	return os.MkdirAll(ticketsDir, 0755)
}

// AllSquadTicketsDirs returns all squad ticket directories for watching.
func AllSquadTicketsDirs() ([]string, error) {
	squads, err := paths.ListSquads()
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, squadName := range squads {
		dir := paths.SquadTicketsDir(squadName)
		// Only include if directory exists
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			dirs = append(dirs, dir)
		}
	}

	return dirs, nil
}
