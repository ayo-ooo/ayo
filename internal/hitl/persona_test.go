package hitl

import (
	"testing"
)

func TestPersonaManager_ShouldDisclose(t *testing.T) {
	tests := []struct {
		name       string
		disclosure DisclosureLevel
		recipient  Recipient
		want       bool
	}{
		{
			name:       "always discloses to anyone",
			disclosure: DisclosureAlways,
			recipient:  Recipient{Type: RecipientEmail, Address: "stranger@example.com"},
			want:       true,
		},
		{
			name:       "never discloses to anyone",
			disclosure: DisclosureNever,
			recipient:  Recipient{Type: RecipientOwner},
			want:       false,
		},
		{
			name:       "owner_only discloses to owner",
			disclosure: DisclosureOwnerOnly,
			recipient:  Recipient{Type: RecipientOwner},
			want:       true,
		},
		{
			name:       "owner_only hides from email",
			disclosure: DisclosureOwnerOnly,
			recipient:  Recipient{Type: RecipientEmail, Address: "stranger@example.com"},
			want:       false,
		},
		{
			name:       "owner_only hides from chat",
			disclosure: DisclosureOwnerOnly,
			recipient:  Recipient{Type: RecipientChat},
			want:       false,
		},
		{
			name:       "default is owner_only",
			disclosure: "",
			recipient:  Recipient{Type: RecipientOwner},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPersonaManager(PersonaConfig{
				Name:       "Test Agent",
				Disclosure: tt.disclosure,
			}, "owner-123")

			got := pm.ShouldDisclose(tt.recipient)
			if got != tt.want {
				t.Errorf("ShouldDisclose() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPersonaManager_GetSignature(t *testing.T) {
	tests := []struct {
		name      string
		config    PersonaConfig
		recipient Recipient
		wantAI    bool
	}{
		{
			name: "adds AI suffix for owner",
			config: PersonaConfig{
				Name:       "Assistant",
				Signature:  "Best,\nAssistant",
				Disclosure: DisclosureOwnerOnly,
			},
			recipient: Recipient{Type: RecipientOwner},
			wantAI:    true,
		},
		{
			name: "no AI suffix for third party",
			config: PersonaConfig{
				Name:       "Assistant",
				Signature:  "Best,\nAssistant",
				Disclosure: DisclosureOwnerOnly,
			},
			recipient: Recipient{Type: RecipientEmail, Address: "user@example.com"},
			wantAI:    false,
		},
		{
			name: "uses default signature",
			config: PersonaConfig{
				Name:       "Finance Bot",
				Disclosure: DisclosureNever,
			},
			recipient: Recipient{Type: RecipientOwner},
			wantAI:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPersonaManager(tt.config, "owner-123")
			sig := pm.GetSignature(tt.recipient)

			hasAI := contains(sig, "(AI Assistant)")
			if hasAI != tt.wantAI {
				t.Errorf("signature AI indicator = %v, want %v\nSignature: %s", hasAI, tt.wantAI, sig)
			}
		})
	}
}

func TestPersonaManager_GetDisplayName(t *testing.T) {
	tests := []struct {
		name   string
		config PersonaConfig
		want   string
	}{
		{
			name: "name only",
			config: PersonaConfig{
				Name: "Finance Assistant",
			},
			want: "Finance Assistant",
		},
		{
			name: "name with title",
			config: PersonaConfig{
				Name:  "Sarah",
				Title: "Accounts Payable",
			},
			want: "Sarah - Accounts Payable",
		},
		{
			name:   "default name",
			config: PersonaConfig{},
			want:   "Assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPersonaManager(tt.config, "owner-123")
			got := pm.GetDisplayName(Recipient{Type: RecipientOwner})
			if got != tt.want {
				t.Errorf("GetDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
