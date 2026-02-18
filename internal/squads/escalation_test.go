package squads

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/tickets"
)

func TestNewEscalateTool(t *testing.T) {
	t.Run("creates tool with correct name and description", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := EscalationToolConfig{
			SquadName:   "test-squad",
			AgentHandle: "@test-agent",
			TicketsDir:  tmpDir,
		}

		tool := NewEscalateTool(cfg)
		info := tool.Info()
		if info.Name != "escalate" {
			t.Errorf("expected name 'escalate', got %q", info.Name)
		}
		if info.Description == "" {
			t.Error("expected non-empty description")
		}
	})

	t.Run("creates escalation ticket on success", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := EscalationToolConfig{
			SquadName:   "alpha-squad",
			AgentHandle: "@worker",
			TicketsDir:  tmpDir,
		}

		tool := NewEscalateTool(cfg)

		// Create params as JSON (like actual tool calls)
		params := EscalationParams{
			Reason:   "Cannot access external API",
			Context:  "Tried 3 retries, all failed with 403",
			Priority: 1,
		}
		paramsJSON, _ := json.Marshal(params)

		call := fantasy.ToolCall{
			ID:    "test-1",
			Name:  "escalate",
			Input: string(paramsJSON),
		}

		resp, err := tool.Run(context.Background(), call)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check response indicates success
		if resp.Content == "" {
			t.Error("expected non-empty response content")
		}

		// Verify ticket was created
		files, err := os.ReadDir(tmpDir)
		if err != nil {
			t.Fatalf("failed to read tmpDir: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 ticket file, got %d", len(files))
		}

		// Load and verify ticket
		svc := tickets.NewDirectService(tmpDir)
		all, err := svc.List("", tickets.Filter{})
		if err != nil {
			t.Fatalf("failed to list tickets: %v", err)
		}
		if len(all) != 1 {
			t.Fatalf("expected 1 ticket, got %d", len(all))
		}

		ticket := all[0]
		if ticket.Type != tickets.TypeEscalation {
			t.Errorf("expected type escalation, got %s", ticket.Type)
		}
		if ticket.Priority != 1 {
			t.Errorf("expected priority 1, got %d", ticket.Priority)
		}
		if ticket.Assignee != "squad-lead" {
			t.Errorf("expected assignee 'squad-lead', got %q", ticket.Assignee)
		}
	})

	t.Run("returns error when reason is empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := EscalationToolConfig{
			SquadName:   "test-squad",
			AgentHandle: "@test",
			TicketsDir:  tmpDir,
		}

		tool := NewEscalateTool(cfg)

		params := EscalationParams{
			Reason: "",
		}
		paramsJSON, _ := json.Marshal(params)

		call := fantasy.ToolCall{
			ID:    "test-2",
			Name:  "escalate",
			Input: string(paramsJSON),
		}

		resp, err := tool.Run(context.Background(), call)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return an error response
		if !resp.IsError {
			t.Error("expected error response when reason is empty")
		}
	})

	t.Run("uses default priority when invalid", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := EscalationToolConfig{
			SquadName:   "test-squad",
			AgentHandle: "@test",
			TicketsDir:  tmpDir,
		}

		tool := NewEscalateTool(cfg)

		params := EscalationParams{
			Reason:   "Some reason",
			Priority: 99, // Invalid
		}
		paramsJSON, _ := json.Marshal(params)

		call := fantasy.ToolCall{
			ID:    "test-3",
			Name:  "escalate",
			Input: string(paramsJSON),
		}

		_, err := tool.Run(context.Background(), call)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify ticket has default priority
		svc := tickets.NewDirectService(tmpDir)
		all, _ := svc.List("", tickets.Filter{})
		if len(all) != 1 {
			t.Fatal("expected 1 ticket")
		}
		if all[0].Priority != tickets.DefaultPriority {
			t.Errorf("expected default priority %d, got %d", tickets.DefaultPriority, all[0].Priority)
		}
	})
}

func TestTruncateTitle(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "short string unchanged",
			input:  "Short reason",
			expect: "Short reason",
		},
		{
			name:   "exactly 60 chars unchanged",
			input:  "123456789012345678901234567890123456789012345678901234567890",
			expect: "123456789012345678901234567890123456789012345678901234567890",
		},
		{
			name:   "longer than 60 chars truncated",
			input:  "This is a very long reason that exceeds the sixty character limit and needs truncation",
			expect: "This is a very long reason that exceeds the sixty charact...",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := truncateTitle(tc.input)
			if got != tc.expect {
				t.Errorf("truncateTitle(%q) = %q, want %q", tc.input, got, tc.expect)
			}
		})
	}
}

func TestIsEscalation(t *testing.T) {
	tests := []struct {
		name   string
		ticket *tickets.Ticket
		expect bool
	}{
		{
			name:   "nil ticket",
			ticket: nil,
			expect: false,
		},
		{
			name:   "escalation type",
			ticket: &tickets.Ticket{Type: tickets.TypeEscalation},
			expect: true,
		},
		{
			name:   "escalation tag",
			ticket: &tickets.Ticket{Type: tickets.TypeTask, Tags: []string{"escalation"}},
			expect: true,
		},
		{
			name:   "regular task",
			ticket: &tickets.Ticket{Type: tickets.TypeTask, Tags: []string{"bug"}},
			expect: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsEscalation(tc.ticket)
			if got != tc.expect {
				t.Errorf("IsEscalation() = %v, want %v", got, tc.expect)
			}
		})
	}
}

func TestListEscalations(t *testing.T) {
	tmpDir := t.TempDir()
	svc := tickets.NewDirectService(tmpDir)

	// Create mixed tickets
	_, _ = svc.Create("", tickets.CreateOptions{
		Title: "Regular task",
		Type:  tickets.TypeTask,
	})
	_, _ = svc.Create("", tickets.CreateOptions{
		Title: "Escalation 1",
		Type:  tickets.TypeEscalation,
	})
	_, _ = svc.Create("", tickets.CreateOptions{
		Title: "Bug fix",
		Type:  tickets.TypeBug,
	})
	_, _ = svc.Create("", tickets.CreateOptions{
		Title: "Tagged escalation",
		Type:  tickets.TypeTask,
		Tags:  []string{EscalationTag},
	})

	escalations, err := ListEscalations(svc, "")
	if err != nil {
		t.Fatalf("ListEscalations failed: %v", err)
	}

	if len(escalations) != 2 {
		t.Errorf("expected 2 escalations, got %d", len(escalations))
	}
}

func TestEscalationTypeConstant(t *testing.T) {
	// Verify EscalationType equals tickets.TypeEscalation
	if EscalationType != tickets.TypeEscalation {
		t.Errorf("EscalationType (%q) != tickets.TypeEscalation (%q)", EscalationType, tickets.TypeEscalation)
	}

	// Verify it's valid
	if !tickets.TypeEscalation.Valid() {
		t.Error("TypeEscalation should be valid")
	}
}

func TestEscalationTicketContent(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := EscalationToolConfig{
		SquadName:   "dev-squad",
		AgentHandle: "@coder",
		TicketsDir:  tmpDir,
	}

	tool := NewEscalateTool(cfg)

	params := EscalationParams{
		Reason:  "Need database access",
		Context: "Working on feature X",
	}
	paramsJSON, _ := json.Marshal(params)

	call := fantasy.ToolCall{
		ID:    "test",
		Name:  "escalate",
		Input: string(paramsJSON),
	}

	_, _ = tool.Run(context.Background(), call)

	// Read the ticket file directly
	files, _ := os.ReadDir(tmpDir)
	if len(files) != 1 {
		t.Fatal("expected 1 ticket file")
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	if err != nil {
		t.Fatalf("failed to read ticket: %v", err)
	}

	// Verify content includes expected sections
	contentStr := string(content)
	checks := []string{
		"Need database access",
		"Working on feature X",
		"@coder",
		"dev-squad",
		"## Reason",
		"## Context",
		"## Source",
	}

	for _, check := range checks {
		if !strings.Contains(contentStr, check) {
			t.Errorf("ticket content missing %q", check)
		}
	}
}
