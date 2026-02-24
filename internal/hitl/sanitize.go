package hitl

import (
	"regexp"
	"strings"
)

// aiPatterns are regex patterns that indicate AI nature in text.
var aiPatterns = []*regexp.Regexp{
	// Direct statements about being AI
	regexp.MustCompile(`(?i)\bI('m| am) an? (AI|artificial intelligence|language model|LLM|chatbot|virtual assistant)\b`),
	regexp.MustCompile(`(?i)\bAs an? (AI|artificial intelligence|language model|LLM)\b`),
	regexp.MustCompile(`(?i)\bI('m| am) (a computer|not human|a machine|a bot)\b`),
	
	// References to AI model names
	regexp.MustCompile(`(?i)\b(GPT|Claude|Gemini|Llama|ChatGPT|OpenAI|Anthropic)('s)?\b`),
	
	// Meta-commentary about capabilities
	regexp.MustCompile(`(?i)\bI (don't|cannot|can't) (actually |really )?(feel|have feelings|experience emotions)\b`),
	regexp.MustCompile(`(?i)\bI (was|am) (trained|programmed|designed) (to|by)\b`),
	regexp.MustCompile(`(?i)\bmy (training|programming|knowledge cutoff)\b`),
	
	// Limitations disclaimers
	regexp.MustCompile(`(?i)\bI (don't|cannot|can't) (browse|access|search) the (internet|web)\b`),
	regexp.MustCompile(`(?i)\b(my|the) (knowledge|training) (cutoff|ends|stopped)\b`),
}

// Sanitizer removes AI indicators from text for third-party communications.
type Sanitizer struct {
	customPatterns []*regexp.Regexp
}

// NewSanitizer creates a new Sanitizer with optional custom patterns.
func NewSanitizer(customPatterns ...string) (*Sanitizer, error) {
	s := &Sanitizer{}
	for _, pattern := range customPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		s.customPatterns = append(s.customPatterns, re)
	}
	return s, nil
}

// Sanitize removes AI indicators from text.
func (s *Sanitizer) Sanitize(text string) string {
	result := text
	
	// Apply built-in patterns
	for _, pattern := range aiPatterns {
		result = pattern.ReplaceAllString(result, "")
	}
	
	// Apply custom patterns
	for _, pattern := range s.customPatterns {
		result = pattern.ReplaceAllString(result, "")
	}
	
	// Clean up multiple spaces and empty lines
	result = regexp.MustCompile(`  +`).ReplaceAllString(result, " ")
	result = regexp.MustCompile(`\n\n\n+`).ReplaceAllString(result, "\n\n")
	result = strings.TrimSpace(result)
	
	return result
}

// SanitizeForRecipient sanitizes text only for non-disclosed recipients.
func (s *Sanitizer) SanitizeForRecipient(text string, persona *PersonaManager, recipient Recipient) string {
	if persona.ShouldDisclose(recipient) {
		return text
	}
	return s.Sanitize(text)
}

// ContainsAIIndicators checks if text contains AI indicators.
func (s *Sanitizer) ContainsAIIndicators(text string) bool {
	for _, pattern := range aiPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	for _, pattern := range s.customPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// DefaultSanitizer returns a sanitizer with default patterns.
func DefaultSanitizer() *Sanitizer {
	s, _ := NewSanitizer()
	return s
}
