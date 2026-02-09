package main

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestTriggerCommandStructure verifies the trigger command tree is set up correctly.
func TestTriggerCommandStructure(t *testing.T) {
	cmd := newTriggerCmd()

	// Verify command name and aliases
	if cmd.Use != "trigger" {
		t.Errorf("trigger command Use = %q, want 'trigger'", cmd.Use)
	}

	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "triggers" {
		t.Errorf("trigger command should have 'triggers' alias")
	}

	// Check subcommands exist
	subcommands := make(map[string]*cobra.Command)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Name()] = sub
	}

	expectedSubcmds := []string{"list", "show", "schedule", "watch", "rm", "test", "enable", "disable"}
	for _, name := range expectedSubcmds {
		if _, ok := subcommands[name]; !ok {
			t.Errorf("missing subcommand: %s", name)
		}
	}

	// Verify 'add' command exists but is hidden
	if addCmd, ok := subcommands["add"]; ok {
		if !addCmd.Hidden {
			t.Error("'add' command should be hidden")
		}
	} else {
		t.Error("'add' command should exist (hidden for backwards compatibility)")
	}
}

// TestScheduleCommandArgs verifies schedule command argument handling.
func TestScheduleCommandArgs(t *testing.T) {
	cmd := scheduleCmd()

	// Should require exactly 2 args
	if cmd.Args == nil {
		t.Error("schedule command should have Args validator")
		return
	}

	// Test that too few args fails
	err := cmd.Args(cmd, []string{"@agent"})
	if err == nil {
		t.Error("schedule should error with only 1 arg")
	}

	// Test that 2 args passes
	err = cmd.Args(cmd, []string{"@agent", "0 * * * * *"})
	if err != nil {
		t.Errorf("schedule should accept 2 args: %v", err)
	}
}

// TestWatchCommandArgs verifies watch command argument handling.
func TestWatchCommandArgs(t *testing.T) {
	cmd := watchCmd()

	// Should require at least 2 args (path and agent)
	if cmd.Args == nil {
		t.Error("watch command should have Args validator")
		return
	}

	// Test that too few args fails
	err := cmd.Args(cmd, []string{"./src"})
	if err == nil {
		t.Error("watch should error with only 1 arg")
	}

	// Test that 2 args passes
	err = cmd.Args(cmd, []string{"./src", "@build"})
	if err != nil {
		t.Errorf("watch should accept 2 args: %v", err)
	}

	// Test that 3+ args passes (patterns)
	err = cmd.Args(cmd, []string{"./src", "@build", "*.go", "*.mod"})
	if err != nil {
		t.Errorf("watch should accept additional pattern args: %v", err)
	}
}

// TestShowCommandArgs verifies show command allows optional ID.
func TestShowCommandArgs(t *testing.T) {
	cmd := showTriggerCmd()

	// Should accept 0 or 1 args
	err := cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("show should accept 0 args: %v", err)
	}

	err = cmd.Args(cmd, []string{"trig_123"})
	if err != nil {
		t.Errorf("show should accept 1 arg: %v", err)
	}

	err = cmd.Args(cmd, []string{"trig_123", "extra"})
	if err == nil {
		t.Error("show should error with 2 args")
	}
}

// TestRemoveCommandArgs verifies rm command allows optional ID.
func TestRemoveCommandArgs(t *testing.T) {
	cmd := removeTriggerCmd()

	// Verify aliases
	if len(cmd.Aliases) < 2 {
		t.Error("rm command should have aliases")
	}

	aliases := make(map[string]bool)
	for _, a := range cmd.Aliases {
		aliases[a] = true
	}
	if !aliases["remove"] || !aliases["delete"] {
		t.Error("rm command should have 'remove' and 'delete' aliases")
	}
}

// TestEnableDisableCommandArgs verifies enable/disable accept optional ID.
func TestEnableDisableCommandArgs(t *testing.T) {
	enableCmd := enableTriggerCmd()
	disableCmd := disableTriggerCmd()

	for _, cmd := range []*cobra.Command{enableCmd, disableCmd} {
		// Should accept 0 or 1 args
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("%s should accept 0 args: %v", cmd.Name(), err)
		}

		err = cmd.Args(cmd, []string{"trig_123"})
		if err != nil {
			t.Errorf("%s should accept 1 arg: %v", cmd.Name(), err)
		}
	}
}

// TestScheduleCommandFlags verifies schedule has expected flags.
func TestScheduleCommandFlags(t *testing.T) {
	cmd := scheduleCmd()

	// Should have --prompt flag
	promptFlag := cmd.Flags().Lookup("prompt")
	if promptFlag == nil {
		t.Error("schedule should have --prompt flag")
	} else {
		if promptFlag.Shorthand != "p" {
			t.Errorf("--prompt shorthand = %q, want 'p'", promptFlag.Shorthand)
		}
	}
}

// TestWatchCommandFlags verifies watch has expected flags.
func TestWatchCommandFlags(t *testing.T) {
	cmd := watchCmd()

	// Should have --prompt flag
	promptFlag := cmd.Flags().Lookup("prompt")
	if promptFlag == nil {
		t.Error("watch should have --prompt flag")
	}

	// Should have --recursive flag
	recursiveFlag := cmd.Flags().Lookup("recursive")
	if recursiveFlag == nil {
		t.Error("watch should have --recursive flag")
	} else {
		if recursiveFlag.Shorthand != "r" {
			t.Errorf("--recursive shorthand = %q, want 'r'", recursiveFlag.Shorthand)
		}
	}

	// Should have --events flag
	eventsFlag := cmd.Flags().Lookup("events")
	if eventsFlag == nil {
		t.Error("watch should have --events flag")
	}
}
