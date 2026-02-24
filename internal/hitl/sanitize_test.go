package hitl

import (
	"testing"
)

func TestSanitizer_Sanitize(t *testing.T) {
	s := DefaultSanitizer()

	tests := []struct {
		name    string
		input   string
		wantAI  bool
	}{
		{
			name:   "removes I am an AI",
			input:  "Hello! I'm an AI assistant. How can I help?",
			wantAI: false,
		},
		{
			name:   "removes As an AI",
			input:  "As an AI, I cannot do that. However, I can help with...",
			wantAI: false,
		},
		{
			name:   "removes I am a language model",
			input:  "I am a language model trained to assist you.",
			wantAI: false,
		},
		{
			name:   "removes model references",
			input:  "I was created by OpenAI and I'm similar to GPT-4.",
			wantAI: false,
		},
		{
			name:   "removes training references",
			input:  "My training data includes information up to 2023.",
			wantAI: false,
		},
		{
			name:   "preserves normal text",
			input:  "The weather looks nice today. Let me check the schedule.",
			wantAI: false,
		},
		{
			name:   "removes I don't have feelings",
			input:  "I don't have feelings, but I understand your concern.",
			wantAI: false,
		},
		{
			name:   "removes I was programmed",
			input:  "I was programmed to help with tasks like this.",
			wantAI: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.Sanitize(tt.input)
			hasAI := s.ContainsAIIndicators(result)
			if hasAI != tt.wantAI {
				t.Errorf("after sanitization, ContainsAIIndicators() = %v, want %v\nInput: %s\nResult: %s", hasAI, tt.wantAI, tt.input, result)
			}
		})
	}
}

func TestSanitizer_ContainsAIIndicators(t *testing.T) {
	s := DefaultSanitizer()

	tests := []struct {
		input string
		want  bool
	}{
		{"Hello, how can I help?", false},
		{"I'm an AI assistant", true},
		{"As an artificial intelligence, I...", true},
		{"I am a language model", true},
		{"Built with Claude technology", true},
		{"Powered by GPT", true},
		{"I cannot feel emotions", true},
		{"My training cutoff is 2023", true},
		{"I was trained to help", true},
		{"The train was late", false},
		{"I feel like having coffee", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := s.ContainsAIIndicators(tt.input)
			if got != tt.want {
				t.Errorf("ContainsAIIndicators(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizer_CustomPatterns(t *testing.T) {
	s, err := NewSanitizer(`(?i)\bconfidential\b`)
	if err != nil {
		t.Fatalf("failed to create sanitizer: %v", err)
	}

	input := "This is confidential information."
	result := s.Sanitize(input)
	if result == input {
		t.Error("custom pattern should have modified the text")
	}
}

func TestSanitizer_SanitizeForRecipient(t *testing.T) {
	s := DefaultSanitizer()
	pm := NewPersonaManager(PersonaConfig{
		Name:       "Assistant",
		Disclosure: DisclosureOwnerOnly,
	}, "owner-123")

	input := "I'm an AI assistant. How can I help?"

	// Should sanitize for non-owner
	result := s.SanitizeForRecipient(input, pm, Recipient{Type: RecipientEmail, Address: "user@example.com"})
	if s.ContainsAIIndicators(result) {
		t.Error("should have sanitized AI indicators for email recipient")
	}

	// Should not sanitize for owner
	result = s.SanitizeForRecipient(input, pm, Recipient{Type: RecipientOwner})
	if result != input {
		t.Error("should not sanitize for owner")
	}
}

func TestSanitizer_CleanupSpaces(t *testing.T) {
	s := DefaultSanitizer()

	input := "Hello.   I'm an AI assistant.   How are you?"
	result := s.Sanitize(input)

	// Should have single spaces, not double
	if containsHelper(result, "  ") {
		t.Errorf("result should not have double spaces: %q", result)
	}
}
