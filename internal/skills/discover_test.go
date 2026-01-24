package skills

import (
	"path/filepath"
	"testing"
)

func TestDiscoverAll(t *testing.T) {
	root := t.TempDir()

	// Set up directory structure
	agentDir := filepath.Join(root, "agent", "skills")
	userDir := filepath.Join(root, "user")
	builtinDir := filepath.Join(root, "builtin")

	// Create skills in each location
	mustWriteSkill(t, filepath.Join(agentDir, "agent-skill"), "agent-skill", "agent specific skill")
	mustWriteSkill(t, filepath.Join(userDir, "user-skill"), "user-skill", "user shared skill")
	mustWriteSkill(t, filepath.Join(builtinDir, "builtin-skill"), "builtin-skill", "built-in skill")

	// Duplicate skill in user and builtin - user should win
	mustWriteSkill(t, filepath.Join(userDir, "common-skill"), "common-skill", "user version")
	mustWriteSkill(t, filepath.Join(builtinDir, "common-skill"), "common-skill", "builtin version")

	result := DiscoverAll(DiscoveryOptions{
		AgentSkillsDir: agentDir,
		UserSharedDir:  userDir,
		BuiltinDir:     builtinDir,
		IgnorePlugins:  true,
	})

	// Should have 4 skills: agent-skill, user-skill, builtin-skill, common-skill
	if len(result.Skills) != 4 {
		t.Fatalf("expected 4 skills, got %d: %v", len(result.Skills), skillNames(result.Skills))
	}

	// Verify common-skill has user version
	for _, s := range result.Skills {
		if s.Name == "common-skill" {
			if s.Description != "user version" {
				t.Errorf("common-skill should be user version, got %s", s.Description)
			}
			if s.Source != SourceUserShared {
				t.Errorf("common-skill source should be user, got %s", s.Source)
			}
		}
	}
}

func TestDiscoverAllWithIncludeFilter(t *testing.T) {
	root := t.TempDir()
	mustWriteSkill(t, filepath.Join(root, "skill-a"), "skill-a", "A")
	mustWriteSkill(t, filepath.Join(root, "skill-b"), "skill-b", "B")
	mustWriteSkill(t, filepath.Join(root, "skill-c"), "skill-c", "C")

	result := DiscoverAll(DiscoveryOptions{
		UserSharedDir: root,
		IncludeSkills: []string{"skill-a", "skill-c"},
		IgnorePlugins: true,
	})

	if len(result.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(result.Skills))
	}

	names := skillNames(result.Skills)
	if !sliceContains(names, "skill-a") || !sliceContains(names, "skill-c") {
		t.Errorf("expected skill-a and skill-c, got %v", names)
	}
	if sliceContains(names, "skill-b") {
		t.Error("skill-b should be excluded")
	}
}

func TestDiscoverAllWithExcludeFilter(t *testing.T) {
	root := t.TempDir()
	mustWriteSkill(t, filepath.Join(root, "skill-a"), "skill-a", "A")
	mustWriteSkill(t, filepath.Join(root, "skill-b"), "skill-b", "B")
	mustWriteSkill(t, filepath.Join(root, "skill-c"), "skill-c", "C")

	result := DiscoverAll(DiscoveryOptions{
		UserSharedDir: root,
		ExcludeSkills: []string{"skill-b"},
		IgnorePlugins: true,
	})

	if len(result.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(result.Skills))
	}

	names := skillNames(result.Skills)
	if sliceContains(names, "skill-b") {
		t.Error("skill-b should be excluded")
	}
}

func TestDiscoverAllIgnoreBuiltin(t *testing.T) {
	root := t.TempDir()
	userDir := filepath.Join(root, "user")
	builtinDir := filepath.Join(root, "builtin")

	mustWriteSkill(t, filepath.Join(userDir, "user-skill"), "user-skill", "user")
	mustWriteSkill(t, filepath.Join(builtinDir, "builtin-skill"), "builtin-skill", "builtin")

	result := DiscoverAll(DiscoveryOptions{
		UserSharedDir: userDir,
		BuiltinDir:    builtinDir,
		IgnoreBuiltin: true,
		IgnorePlugins: true,
	})

	if len(result.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(result.Skills))
	}

	if result.Skills[0].Name != "user-skill" {
		t.Errorf("expected user-skill, got %s", result.Skills[0].Name)
	}
}

func TestDiscoverAllIgnoreShared(t *testing.T) {
	root := t.TempDir()
	userDir := filepath.Join(root, "user")
	builtinDir := filepath.Join(root, "builtin")

	mustWriteSkill(t, filepath.Join(userDir, "user-skill"), "user-skill", "user")
	mustWriteSkill(t, filepath.Join(builtinDir, "builtin-skill"), "builtin-skill", "builtin")

	result := DiscoverAll(DiscoveryOptions{
		UserSharedDir: userDir,
		BuiltinDir:    builtinDir,
		IgnoreShared:  true,
		IgnorePlugins: true,
	})

	if len(result.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(result.Skills))
	}

	if result.Skills[0].Name != "builtin-skill" {
		t.Errorf("expected builtin-skill, got %s", result.Skills[0].Name)
	}
}

func TestDiscoverAllSortsByName(t *testing.T) {
	root := t.TempDir()
	mustWriteSkill(t, filepath.Join(root, "zebra"), "zebra", "Z")
	mustWriteSkill(t, filepath.Join(root, "alpha"), "alpha", "A")
	mustWriteSkill(t, filepath.Join(root, "beta"), "beta", "B")

	result := DiscoverAll(DiscoveryOptions{
		UserSharedDir: root,
		IgnorePlugins: true,
	})

	if len(result.Skills) != 3 {
		t.Fatalf("expected 3 skills, got %d", len(result.Skills))
	}

	expected := []string{"alpha", "beta", "zebra"}
	for i, s := range result.Skills {
		if s.Name != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], s.Name)
		}
	}
}

func TestDiscoverForAgent(t *testing.T) {
	root := t.TempDir()
	agentDir := filepath.Join(root, "agent")
	userDir := filepath.Join(root, "user")

	mustWriteSkill(t, filepath.Join(agentDir, "agent-skill"), "agent-skill", "agent")
	mustWriteSkill(t, filepath.Join(userDir, "user-skill"), "user-skill", "user")

	result := DiscoverForAgent(agentDir, []string{userDir}, DiscoveryFilterConfig{
		IgnorePlugins: true,
	})

	if len(result.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(result.Skills))
	}
}

func TestFilterSkillsEmpty(t *testing.T) {
	skills := []Metadata{
		{Name: "a"},
		{Name: "b"},
	}

	// No filters - should return all
	filtered := filterSkills(skills, nil, nil)
	if len(filtered) != 2 {
		t.Errorf("expected 2 skills, got %d", len(filtered))
	}
}

func TestFilterSkillsBothFilters(t *testing.T) {
	skills := []Metadata{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
	}

	// Include a, b but exclude b
	filtered := filterSkills(skills, []string{"a", "b"}, []string{"b"})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(filtered))
	}
	if filtered[0].Name != "a" {
		t.Errorf("expected a, got %s", filtered[0].Name)
	}
}

func skillNames(skills []Metadata) []string {
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	return names
}

func sliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
