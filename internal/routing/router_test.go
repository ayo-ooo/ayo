package routing

import (
	"context"
	"testing"
)

func TestIsTrivial(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected bool
	}{
		{"empty", "", true},
		{"greeting", "hi", true},
		{"short question", "what time is it?", true},
		{"longer question under threshold", "can you help me with this?", true},
		{"long prompt", "I need you to build a comprehensive authentication system with JWT tokens, refresh tokens, rate limiting, and integration with our existing user database. The system should support multiple auth providers.", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isTrivial(tc.prompt)
			if got != tc.expected {
				t.Errorf("isTrivial(%q) = %v, want %v", tc.prompt, got, tc.expected)
			}
		})
	}
}

func TestParseExplicitTarget(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{"agent", "@agent write some code", "@agent"},
		{"agent with dots", "@agent.sub do something", "@agent.sub"},
		{"squad", "#devteam build auth feature", "#devteam"},
		{"no target", "build auth feature", ""},
		{"at sign mid prompt", "please ask @agent to help", ""},
		{"hash mid prompt", "use #hashtag in tweet", ""},
		{"just at sign", "@", ""},
		{"just hash", "#", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseExplicitTarget(tc.prompt)
			if got != tc.expected {
				t.Errorf("parseExplicitTarget(%q) = %q, want %q", tc.prompt, got, tc.expected)
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty", "", 0},
		{"single", "hello", 1},
		{"multiple", "hello world", 2},
		{"with extra spaces", "  hello   world  ", 2},
		{"many words", "one two three four five", 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := countWords(tc.input)
			if got != tc.expected {
				t.Errorf("countWords(%q) = %d, want %d", tc.input, got, tc.expected)
			}
		})
	}
}

func TestRouter_Decide_NoSearcher(t *testing.T) {
	router := NewRouter(nil)
	decision, err := router.Decide(context.Background(), "build a complex feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.TargetType != Self {
		t.Errorf("expected Self, got %v", decision.TargetType)
	}
}

func TestRouter_Decide_Trivial(t *testing.T) {
	router := NewRouter(nil)
	decision, err := router.Decide(context.Background(), "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.TargetType != Self {
		t.Errorf("expected Self, got %v", decision.TargetType)
	}
	if decision.Reason != "trivial input" {
		t.Errorf("expected reason 'trivial input', got %q", decision.Reason)
	}
}

func TestRouter_Decide_ExplicitAgent(t *testing.T) {
	router := NewRouter(nil)
	decision, err := router.Decide(context.Background(), "@coder write a function")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.TargetType != Agent {
		t.Errorf("expected Agent, got %v", decision.TargetType)
	}
	if decision.Target != "@coder" {
		t.Errorf("expected target @coder, got %q", decision.Target)
	}
	if decision.Reason != "explicit target" {
		t.Errorf("expected reason 'explicit target', got %q", decision.Reason)
	}
}

func TestRouter_Decide_ExplicitSquad(t *testing.T) {
	router := NewRouter(nil)
	decision, err := router.Decide(context.Background(), "#devteam build the auth system")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.TargetType != Squad {
		t.Errorf("expected Squad, got %v", decision.TargetType)
	}
	if decision.Target != "#devteam" {
		t.Errorf("expected target #devteam, got %q", decision.Target)
	}
}

func TestTargetType_String(t *testing.T) {
	tests := []struct {
		targetType TargetType
		expected   string
	}{
		{Self, "self"},
		{Agent, "agent"},
		{Squad, "squad"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			got := tc.targetType.String()
			if got != tc.expected {
				t.Errorf("TargetType(%d).String() = %q, want %q", tc.targetType, got, tc.expected)
			}
		})
	}
}
