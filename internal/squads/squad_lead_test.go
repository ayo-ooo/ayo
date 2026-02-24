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

func TestLeadDisabledTools(t *testing.T) {
	tools := LeadDisabledTools()
	if len(tools) == 0 {
		t.Error("LeadDisabledTools should return at least one tool")
	}

	// Lead should have bash disabled
	hasBash := false
	hasEdit := false
	for _, tool := range tools {
		if tool == "bash" {
			hasBash = true
		}
		if tool == "edit" {
			hasEdit = true
		}
	}
	if !hasBash {
		t.Error("LeadDisabledTools should include bash")
	}
	if !hasEdit {
		t.Error("LeadDisabledTools should include edit")
	}
}

func TestWorkerDisabledTools(t *testing.T) {
	tools := WorkerDisabledTools()
	if len(tools) == 0 {
		t.Error("WorkerDisabledTools should return at least one tool")
	}

	// Worker should have ticket_create disabled
	hasTicketCreate := false
	hasTicketAssign := false
	for _, tool := range tools {
		if tool == "ticket_create" {
			hasTicketCreate = true
		}
		if tool == "ticket_assign" {
			hasTicketAssign = true
		}
	}
	if !hasTicketCreate {
		t.Error("WorkerDisabledTools should include ticket_create")
	}
	if !hasTicketAssign {
		t.Error("WorkerDisabledTools should include ticket_assign")
	}
}

func TestGetDisabledToolsForRole(t *testing.T) {
	tests := []struct {
		name        string
		role        string
		isLead      bool
		wantBash    bool // should have bash in disabled list
		wantTickets bool // should have ticket_create in disabled list
	}{
		{
			name:        "lead role by name",
			role:        "architect",
			isLead:      false,
			wantBash:    true,
			wantTickets: false,
		},
		{
			name:        "worker role by name",
			role:        "developer",
			isLead:      false,
			wantBash:    false,
			wantTickets: true,
		},
		{
			name:        "explicit lead flag overrides",
			role:        "developer",
			isLead:      true,
			wantBash:    true,
			wantTickets: false,
		},
		{
			name:        "pm role (lead)",
			role:        "pm",
			isLead:      false,
			wantBash:    true,
			wantTickets: false,
		},
		{
			name:        "manager role (lead)",
			role:        "manager",
			isLead:      false,
			wantBash:    true,
			wantTickets: false,
		},
		{
			name:        "frontend role (worker)",
			role:        "frontend",
			isLead:      false,
			wantBash:    false,
			wantTickets: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := GetDisabledToolsForRole(tt.role, tt.isLead)

			hasBash := false
			hasTicketCreate := false
			for _, tool := range tools {
				if tool == "bash" {
					hasBash = true
				}
				if tool == "ticket_create" {
					hasTicketCreate = true
				}
			}

			if hasBash != tt.wantBash {
				t.Errorf("bash in disabled list = %v, want %v", hasBash, tt.wantBash)
			}
			if hasTicketCreate != tt.wantTickets {
				t.Errorf("ticket_create in disabled list = %v, want %v", hasTicketCreate, tt.wantTickets)
			}
		})
	}
}
