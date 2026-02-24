package squads

import "testing"

func TestLeadTools(t *testing.T) {
	tools := LeadTools()
	if len(tools) == 0 {
		t.Error("LeadTools should return at least one tool")
	}

	// Lead should have ticket tools
	hasTicketCreate := false
	for _, tool := range tools {
		if tool == "ticket_create" {
			hasTicketCreate = true
			break
		}
	}
	if !hasTicketCreate {
		t.Error("LeadTools should include ticket_create")
	}

	// Lead should NOT have bash
	for _, tool := range tools {
		if tool == "bash" {
			t.Error("LeadTools should NOT include bash")
		}
	}
}

func TestWorkerTools(t *testing.T) {
	tools := WorkerTools()
	if len(tools) == 0 {
		t.Error("WorkerTools should return at least one tool")
	}

	// Worker should have bash
	hasBash := false
	for _, tool := range tools {
		if tool == "bash" {
			hasBash = true
			break
		}
	}
	if !hasBash {
		t.Error("WorkerTools should include bash")
	}

	// Worker should NOT have ticket_create
	for _, tool := range tools {
		if tool == "ticket_create" {
			t.Error("WorkerTools should NOT include ticket_create")
		}
	}
}

func TestIsLeadRole(t *testing.T) {
	tests := []struct {
		role     string
		expected bool
	}{
		{"lead", true},
		{"architect", true},
		{"planner", true},
		{"pm", true},
		{"manager", true},
		{"developer", false},
		{"worker", false},
		{"frontend", false},
		{"backend", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			got := IsLeadRole(tt.role)
			if got != tt.expected {
				t.Errorf("IsLeadRole(%q) = %v, want %v", tt.role, got, tt.expected)
			}
		})
	}
}

func TestGetToolsForRole(t *testing.T) {
	// Lead role should get lead tools
	leadTools := GetToolsForRole("architect", false)
	hasTicketCreate := false
	for _, tool := range leadTools {
		if tool == "ticket_create" {
			hasTicketCreate = true
			break
		}
	}
	if !hasTicketCreate {
		t.Error("Lead role should get ticket_create tool")
	}

	// Explicit isLead should work
	leadTools2 := GetToolsForRole("developer", true)
	hasTicketCreate = false
	for _, tool := range leadTools2 {
		if tool == "ticket_create" {
			hasTicketCreate = true
			break
		}
	}
	if !hasTicketCreate {
		t.Error("isLead=true should get ticket_create tool")
	}

	// Worker should get worker tools
	workerTools := GetToolsForRole("developer", false)
	hasBash := false
	for _, tool := range workerTools {
		if tool == "bash" {
			hasBash = true
			break
		}
	}
	if !hasBash {
		t.Error("Worker role should get bash tool")
	}
}
