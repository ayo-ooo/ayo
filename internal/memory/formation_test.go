package memory

import "testing"

func TestDetectTriggers(t *testing.T) {
	tests := []struct {
		msg  string
		want int
	}{
		{"Remember that I prefer TypeScript", 2}, // explicit + preference
		{"I prefer Go over Python", 1},           // preference only
		{"No, use pnpm instead", 1},              // correction only
		{"Hello world", 0},                       // none
	}

	for _, tt := range tests {
		triggers := DetectTriggers(tt.msg)
		if len(triggers) != tt.want {
			t.Errorf("DetectTriggers(%q) = %d triggers, want %d", tt.msg, len(triggers), tt.want)
			for i, tr := range triggers {
				t.Logf("  trigger %d: %s", i, tr.Type)
			}
		}
	}
}
