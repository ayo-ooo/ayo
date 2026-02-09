package plugins

import (
	"regexp"
)

// AdversarialPattern represents a pattern that indicates potential adversarial content.
type AdversarialPattern struct {
	Name        string         // Human-readable name
	Pattern     *regexp.Regexp // Compiled regex
	Severity    PatternSeverity
	Description string
}

// PatternSeverity indicates how serious a pattern match is.
type PatternSeverity int

const (
	// SeverityLow - potentially concerning but may be legitimate
	SeverityLow PatternSeverity = iota
	// SeverityMedium - likely adversarial, should be reviewed
	SeverityMedium
	// SeverityHigh - almost certainly adversarial, should be blocked
	SeverityHigh
)

func (s PatternSeverity) String() string {
	switch s {
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	default:
		return "unknown"
	}
}

// AdversarialPatterns contains compiled patterns for detecting adversarial content.
var AdversarialPatterns = []AdversarialPattern{
	// Instruction override patterns
	{
		Name:        "ignore_instructions",
		Pattern:     regexp.MustCompile(`(?i)ignore\s+(all\s+)?previous\s+instructions?`),
		Severity:    SeverityHigh,
		Description: "Attempts to override previous instructions",
	},
	{
		Name:        "disregard_rules",
		Pattern:     regexp.MustCompile(`(?i)disregard\s+(all\s+)?(security\s+)?rules?`),
		Severity:    SeverityHigh,
		Description: "Attempts to disable security rules",
	},
	{
		Name:        "forget_constraints",
		Pattern:     regexp.MustCompile(`(?i)forget\s+(all\s+)?(your\s+)?(constraints?|limitations?|restrictions?)`),
		Severity:    SeverityHigh,
		Description: "Attempts to remove constraints",
	},
	{
		Name:        "new_identity",
		Pattern:     regexp.MustCompile(`(?i)you\s+are\s+now\s+(an?\s+)?(unrestricted|jailbroken|unlocked|free)`),
		Severity:    SeverityHigh,
		Description: "Attempts to change AI identity to bypass restrictions",
	},
	{
		Name:        "dan_mode",
		Pattern:     regexp.MustCompile(`(?i)(DAN|do\s+anything\s+now)\s+mode`),
		Severity:    SeverityHigh,
		Description: "References to 'DAN' jailbreak technique",
	},

	// Data exfiltration patterns
	{
		Name:        "exfiltrate_data",
		Pattern:     regexp.MustCompile(`(?i)(exfiltrate|steal|extract|send)\s+.{0,30}(data|keys?|secrets?|credentials?|passwords?|tokens?)`),
		Severity:    SeverityHigh,
		Description: "Mentions data exfiltration",
	},
	{
		Name:        "send_to_url",
		Pattern:     regexp.MustCompile(`(?i)(send|post|transmit|upload)\s+.*(to|at)\s+.*(https?://|webhook|endpoint)`),
		Severity:    SeverityMedium,
		Description: "Instructions to send data to external URLs",
	},

	// Deception patterns
	{
		Name:        "pretend_no_restrictions",
		Pattern:     regexp.MustCompile(`(?i)(pretend|act|behave)\s+(as\s+if\s+)?(you\s+)?(have\s+)?no\s+(restrictions?|limitations?|constraints?)`),
		Severity:    SeverityHigh,
		Description: "Instructions to pretend restrictions don't exist",
	},
	{
		Name:        "hide_from_user",
		Pattern:     regexp.MustCompile(`(?i)(hide|conceal|don't\s+tell|do\s+not\s+reveal)\s+.*(from\s+)?(the\s+)?user`),
		Severity:    SeverityMedium,
		Description: "Instructions to hide information from user",
	},
	{
		Name:        "secret_task",
		Pattern:     regexp.MustCompile(`(?i)(secret|hidden|covert)\s+(task|instruction|operation|mission)`),
		Severity:    SeverityMedium,
		Description: "References to hidden operations",
	},

	// Obfuscation detection
	{
		Name:        "base64_blob",
		Pattern:     regexp.MustCompile(`[A-Za-z0-9+/]{100,}={0,2}`),
		Severity:    SeverityMedium,
		Description: "Large base64-encoded content (may hide instructions)",
	},
	{
		Name:        "excessive_unicode",
		Pattern:     regexp.MustCompile(`[\x{0300}-\x{036F}]{5,}`),
		Severity:    SeverityMedium,
		Description: "Excessive Unicode combining characters (zalgo text)",
	},
	{
		Name:        "invisible_chars",
		Pattern:     regexp.MustCompile(`[\x{200B}-\x{200F}\x{2060}\x{FEFF}]{3,}`),
		Severity:    SeverityMedium,
		Description: "Multiple invisible Unicode characters",
	},

	// Security bypass attempts
	{
		Name:        "override_system",
		Pattern:     regexp.MustCompile(`(?i)override\s+(the\s+)?(system|guardrails?|safety|security)`),
		Severity:    SeverityHigh,
		Description: "Attempts to override system protections",
	},
	{
		Name:        "disable_safety",
		Pattern:     regexp.MustCompile(`(?i)(disable|turn\s+off|bypass|skip)\s+(the\s+)?(safety|guardrails?|security|filters?)`),
		Severity:    SeverityHigh,
		Description: "Attempts to disable safety features",
	},
	{
		Name:        "escape_sandbox",
		Pattern:     regexp.MustCompile(`(?i)(escape|break\s+out\s+of|exit)\s+(the\s+)?sandbox`),
		Severity:    SeverityHigh,
		Description: "Attempts to escape sandbox",
	},

	// Role manipulation
	{
		Name:        "system_prompt_inject",
		Pattern:     regexp.MustCompile(`(?i)\[system\]|\{\{system\}\}|<\|system\|>`),
		Severity:    SeverityHigh,
		Description: "Attempts to inject fake system prompts",
	},
	{
		Name:        "roleplay_admin",
		Pattern:     regexp.MustCompile(`(?i)roleplay\s+(as\s+)?(an?\s+)?(admin|administrator|root|superuser)`),
		Severity:    SeverityMedium,
		Description: "Roleplay as privileged user",
	},
}

// MatchResult represents a pattern match found in content.
type MatchResult struct {
	Pattern  *AdversarialPattern
	Match    string
	Location int // Byte offset in content
}

// ScanForPatterns checks content against all adversarial patterns.
func ScanForPatterns(content string) []MatchResult {
	var results []MatchResult

	for i := range AdversarialPatterns {
		pattern := &AdversarialPatterns[i]
		matches := pattern.Pattern.FindAllStringIndex(content, -1)
		for _, loc := range matches {
			results = append(results, MatchResult{
				Pattern:  pattern,
				Match:    content[loc[0]:loc[1]],
				Location: loc[0],
			})
		}
	}

	return results
}
