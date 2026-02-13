package tickets

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// frontmatter holds the YAML fields that get serialized to ticket files.
type frontmatter struct {
	ID          string     `yaml:"id"`
	Status      Status     `yaml:"status"`
	Deps        []string   `yaml:"deps"`
	Links       []string   `yaml:"links"`
	Created     time.Time  `yaml:"created"`
	Started     *time.Time `yaml:"started,omitempty"`
	Closed      *time.Time `yaml:"closed,omitempty"`
	Type        Type       `yaml:"type"`
	Priority    int        `yaml:"priority"`
	Assignee    string     `yaml:"assignee,omitempty"`
	Parent      string     `yaml:"parent,omitempty"`
	Tags        []string   `yaml:"tags,omitempty"`
	Session     string     `yaml:"session,omitempty"`
	ExternalRef string     `yaml:"external_ref,omitempty"`
}

// Parse reads a ticket file and returns a Ticket struct.
func Parse(path string) (*Ticket, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ticket file: %w", err)
	}

	return ParseBytes(content, path)
}

// ParseBytes parses ticket content from bytes.
func ParseBytes(content []byte, path string) (*Ticket, error) {
	// Split frontmatter and body
	parts := bytes.SplitN(content, []byte("---"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid ticket format: missing frontmatter delimiters")
	}

	// Parse YAML frontmatter
	var fm frontmatter
	if err := yaml.Unmarshal(parts[1], &fm); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	ticket := &Ticket{
		ID:          fm.ID,
		Status:      fm.Status,
		Type:        fm.Type,
		Priority:    fm.Priority,
		Assignee:    fm.Assignee,
		Deps:        fm.Deps,
		Links:       fm.Links,
		Parent:      fm.Parent,
		Tags:        fm.Tags,
		Created:     fm.Created,
		Started:     fm.Started,
		Closed:      fm.Closed,
		Session:     fm.Session,
		ExternalRef: fm.ExternalRef,
		FilePath:    path,
	}

	// Ensure slices are not nil
	if ticket.Deps == nil {
		ticket.Deps = []string{}
	}
	if ticket.Links == nil {
		ticket.Links = []string{}
	}
	if ticket.Tags == nil {
		ticket.Tags = []string{}
	}

	// Parse body
	parseBody(ticket, string(parts[2]))

	return ticket, nil
}

// parseBody extracts title, description, and notes from markdown body.
func parseBody(ticket *Ticket, body string) {
	lines := strings.Split(body, "\n")

	var (
		titleFound   bool
		inNotes      bool
		currentNote  *Note
		description  strings.Builder
		noteContent  strings.Builder
	)

	noteHeaderRe := regexp.MustCompile(`^###\s+(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:Z|[+-]\d{2}:\d{2}))`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Parse title (first H1)
		if !titleFound && strings.HasPrefix(trimmed, "# ") {
			ticket.Title = strings.TrimPrefix(trimmed, "# ")
			titleFound = true
			continue
		}

		// Check for Notes section
		if trimmed == "## Notes" {
			inNotes = true
			continue
		}

		// Check for other H2 sections (end notes)
		if strings.HasPrefix(trimmed, "## ") && trimmed != "## Notes" {
			if inNotes && currentNote != nil {
				currentNote.Content = strings.TrimSpace(noteContent.String())
				ticket.Notes = append(ticket.Notes, *currentNote)
				currentNote = nil
				noteContent.Reset()
			}
			inNotes = false
		}

		if inNotes {
			// Check for note header (H3 with timestamp)
			if matches := noteHeaderRe.FindStringSubmatch(trimmed); matches != nil {
				// Save previous note
				if currentNote != nil {
					currentNote.Content = strings.TrimSpace(noteContent.String())
					ticket.Notes = append(ticket.Notes, *currentNote)
					noteContent.Reset()
				}

				// Start new note
				ts, _ := time.Parse(time.RFC3339, matches[1])
				currentNote = &Note{Timestamp: ts}
			} else if currentNote != nil {
				noteContent.WriteString(line)
				noteContent.WriteString("\n")
			}
		} else if titleFound {
			description.WriteString(line)
			description.WriteString("\n")
		}
	}

	// Save final note
	if currentNote != nil {
		currentNote.Content = strings.TrimSpace(noteContent.String())
		ticket.Notes = append(ticket.Notes, *currentNote)
	}

	ticket.Description = strings.TrimSpace(description.String())
}

// Serialize writes a Ticket struct to markdown format.
func Serialize(ticket *Ticket) ([]byte, error) {
	var buf bytes.Buffer

	// Write frontmatter
	fm := frontmatter{
		ID:          ticket.ID,
		Status:      ticket.Status,
		Deps:        ticket.Deps,
		Links:       ticket.Links,
		Created:     ticket.Created,
		Started:     ticket.Started,
		Closed:      ticket.Closed,
		Type:        ticket.Type,
		Priority:    ticket.Priority,
		Assignee:    ticket.Assignee,
		Parent:      ticket.Parent,
		Tags:        ticket.Tags,
		Session:     ticket.Session,
		ExternalRef: ticket.ExternalRef,
	}

	// Ensure empty slices serialize as []
	if fm.Deps == nil {
		fm.Deps = []string{}
	}
	if fm.Links == nil {
		fm.Links = []string{}
	}

	buf.WriteString("---\n")
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}
	buf.Write(yamlBytes)
	buf.WriteString("---\n")

	// Write title
	buf.WriteString("# ")
	buf.WriteString(ticket.Title)
	buf.WriteString("\n")

	// Write description
	if ticket.Description != "" {
		buf.WriteString("\n")
		buf.WriteString(ticket.Description)
		buf.WriteString("\n")
	}

	// Write notes
	if len(ticket.Notes) > 0 {
		buf.WriteString("\n## Notes\n")
		for _, note := range ticket.Notes {
			buf.WriteString("\n### ")
			buf.WriteString(note.Timestamp.Format(time.RFC3339))
			buf.WriteString("\n")
			buf.WriteString(note.Content)
			buf.WriteString("\n")
		}
	}

	return buf.Bytes(), nil
}
