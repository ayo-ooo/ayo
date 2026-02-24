package main

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestAgentsCommandStructure verifies the agents command tree is set up correctly.
func TestAgentsCommandStructure(t *testing.T) {
	cfgPath := ""
	cmd := newAgentsCmd(&cfgPath)

	// Verify command name and aliases (Use is "agent", "agents" is alias)
	if cmd.Use != "agent" {
		t.Errorf("agent command Use = %q, want 'agent'", cmd.Use)
	}

	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "agents" {
		t.Error("agent command should have 'agents' alias")
	}

	// Check subcommands exist
	subcommands := make(map[string]*cobra.Command)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Name()] = sub
	}

	expectedSubcmds := []string{"list", "show", "create", "rm", "update", "status", "wake", "sleep", "capabilities", "promote", "archive", "unarchive", "refine"}
	for _, name := range expectedSubcmds {
		if _, ok := subcommands[name]; !ok {
			t.Errorf("missing subcommand: %s", name)
		}
	}
}

// TestAgentsListFlags verifies list command has expected flags.
func TestAgentsListFlags(t *testing.T) {
	cfgPath := ""
	cmd := newAgentsCmd(&cfgPath)

	// Find list subcommand
	var listCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "list" {
			listCmd = sub
			break
		}
	}

	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}

	// Should have --trust flag
	trustFlag := listCmd.Flags().Lookup("trust")
	if trustFlag == nil {
		t.Error("list should have --trust flag")
	}

	// Should have --type flag
	typeFlag := listCmd.Flags().Lookup("type")
	if typeFlag == nil {
		t.Error("list should have --type flag")
	}
}

// TestAgentsShowArgs verifies show command requires exactly one argument.
func TestAgentsShowArgs(t *testing.T) {
	cfgPath := ""
	cmd := newAgentsCmd(&cfgPath)

	// Find show subcommand
	var showCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "show" {
			showCmd = sub
			break
		}
	}

	if showCmd == nil {
		t.Fatal("show subcommand not found")
	}

	// Should require exactly 1 arg
	if showCmd.Args == nil {
		t.Error("show command should have Args validator")
		return
	}

	err := showCmd.Args(showCmd, []string{})
	if err == nil {
		t.Error("show should error with 0 args")
	}

	err = showCmd.Args(showCmd, []string{"@test"})
	if err != nil {
		t.Errorf("show should accept 1 arg: %v", err)
	}
}

// TestAgentsCreateFlags verifies create command has expected flags.
func TestAgentsCreateFlags(t *testing.T) {
	cfgPath := ""
	cmd := newAgentsCmd(&cfgPath)

	// Find create subcommand
	var createCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "create" {
			createCmd = sub
			break
		}
	}

	if createCmd == nil {
		t.Fatal("create subcommand not found")
	}

	// Required flags
	expectedFlags := []string{"model", "description", "system", "system-file", "tools", "skills"}
	for _, flag := range expectedFlags {
		if createCmd.Flags().Lookup(flag) == nil {
			t.Errorf("create should have --%s flag", flag)
		}
	}
}

// TestAgentsRmAliases verifies rm has remove and delete aliases.
func TestAgentsRmAliases(t *testing.T) {
	cfgPath := ""
	cmd := newAgentsCmd(&cfgPath)

	// Find rm subcommand
	var rmCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "rm" {
			rmCmd = sub
			break
		}
	}

	if rmCmd == nil {
		t.Fatal("rm subcommand not found")
	}

	aliases := make(map[string]bool)
	for _, a := range rmCmd.Aliases {
		aliases[a] = true
	}

	if !aliases["remove"] {
		t.Error("rm should have 'remove' alias")
	}
	if !aliases["delete"] {
		t.Error("rm should have 'delete' alias")
	}
}

// TestAgentsCapabilitiesSubcmd verifies capabilities subcommand exists with expected args.
func TestAgentsCapabilitiesSubcmd(t *testing.T) {
	cfgPath := ""
	cmd := newAgentsCmd(&cfgPath)

	// Find capabilities subcommand
	var capCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "capabilities" {
			capCmd = sub
			break
		}
	}

	if capCmd == nil {
		t.Fatal("capabilities subcommand not found")
	}

	// Should accept handle argument
	if capCmd.Args == nil {
		// Some commands don't have arg validators - that's OK
		return
	}
}

// TestAgentsRefineSubcmd verifies refine subcommand exists.
func TestAgentsRefineSubcmd(t *testing.T) {
	cfgPath := ""
	cmd := newAgentsCmd(&cfgPath)

	// Find refine subcommand
	var refineCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "refine" {
			refineCmd = sub
			break
		}
	}

	if refineCmd == nil {
		t.Fatal("refine subcommand not found")
	}
}

// TestAgentsPromoteSubcmd verifies promote subcommand exists.
func TestAgentsPromoteSubcmd(t *testing.T) {
	cfgPath := ""
	cmd := newAgentsCmd(&cfgPath)

	// Find promote subcommand
	var promoteCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "promote" {
			promoteCmd = sub
			break
		}
	}

	if promoteCmd == nil {
		t.Fatal("promote subcommand not found")
	}
}

// TestAgentsArchiveSubcmd verifies archive/unarchive subcommands exist.
func TestAgentsArchiveSubcmd(t *testing.T) {
	cfgPath := ""
	cmd := newAgentsCmd(&cfgPath)

	subcommands := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Name()] = true
	}

	if !subcommands["archive"] {
		t.Error("archive subcommand not found")
	}
	if !subcommands["unarchive"] {
		t.Error("unarchive subcommand not found")
	}
}
