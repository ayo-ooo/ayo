package hitl

// DisclosureLevel specifies when to reveal AI nature.
type DisclosureLevel string

const (
	DisclosureNever     DisclosureLevel = "never"
	DisclosureOwnerOnly DisclosureLevel = "owner_only"
	DisclosureAlways    DisclosureLevel = "always"
)

// PersonaConfig defines how an agent presents itself.
type PersonaConfig struct {
	Name       string          `json:"name"`
	Title      string          `json:"title,omitempty"`
	Email      string          `json:"email,omitempty"`
	Signature  string          `json:"signature,omitempty"`
	Disclosure DisclosureLevel `json:"disclosure,omitempty"`
}

// PersonaManager handles agent persona and disclosure decisions.
type PersonaManager struct {
	config  PersonaConfig
	ownerID string
}

// NewPersonaManager creates a new PersonaManager with the given config and owner ID.
func NewPersonaManager(config PersonaConfig, ownerID string) *PersonaManager {
	if config.Disclosure == "" {
		config.Disclosure = DisclosureOwnerOnly
	}
	return &PersonaManager{
		config:  config,
		ownerID: ownerID,
	}
}

// ShouldDisclose returns whether AI nature should be disclosed to the recipient.
func (p *PersonaManager) ShouldDisclose(recipient Recipient) bool {
	switch p.config.Disclosure {
	case DisclosureAlways:
		return true
	case DisclosureNever:
		return false
	default: // owner_only
		return recipient.Type == RecipientOwner
	}
}

// GetSignature returns the appropriate signature for the recipient.
func (p *PersonaManager) GetSignature(recipient Recipient) string {
	sig := p.config.Signature
	if sig == "" {
		sig = "Best regards,\n" + p.config.Name
	}
	
	if p.ShouldDisclose(recipient) {
		return sig + "\n\n(AI Assistant)"
	}
	return sig
}

// GetDisplayName returns the appropriate display name for the recipient.
func (p *PersonaManager) GetDisplayName(recipient Recipient) string {
	name := p.config.Name
	if name == "" {
		name = "Assistant"
	}
	
	if p.config.Title != "" {
		return name + " - " + p.config.Title
	}
	return name
}

// Config returns the current persona configuration.
func (p *PersonaManager) Config() PersonaConfig {
	return p.config
}
