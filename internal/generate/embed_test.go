package generate

import (
	"strings"
	"testing"

	"github.com/charmbracelet/ayo/internal/project"
)

func TestGenerateEmbeds_SkillsEmbedding(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
		Skills: []project.Skill{
			{Name: "search-web", Path: "skills/search-web", Description: "Web search capability"},
			{Name: "analyze-code", Path: "skills/analyze-code", Description: "Code analysis"},
		},
	}

	code, err := GenerateEmbeds(proj, "main")
	if err != nil {
		t.Fatalf("GenerateEmbeds() error = %v", err)
	}

	for _, skill := range proj.Skills {
		if !strings.Contains(code, `//go:embed skills/`+skill.Name+`/*`) {
			t.Errorf("Missing embed directive for skill %q", skill.Name)
		}

		safeName := toSafeIdentifier(skill.Name)
		if !strings.Contains(code, "var skill"+safeName+" embed.FS") {
			t.Errorf("Missing embed.FS variable for skill %q", skill.Name)
		}
	}
}

func TestGenerateEmbeds_SkillsCatalogGenerated(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
		Skills: []project.Skill{
			{Name: "search-web", Path: "skills/search-web", Description: "Web search capability for finding current information."},
		},
	}

	code, err := GenerateEmbeds(proj, "main")
	if err != nil {
		t.Fatalf("GenerateEmbeds() error = %v", err)
	}

	if !strings.Contains(code, "skillsCatalog") {
		t.Error("Skills catalog variable should be generated when skills exist")
	}

	if !strings.Contains(code, "## Available Skills") {
		t.Error("Skills catalog should include header")
	}

	if !strings.Contains(code, "### search-web") {
		t.Error("Skills catalog should list skill name")
	}

	if !strings.Contains(code, "embedded://skills/search-web/SKILL.md") {
		t.Error("Skills catalog should include embedded:// path to skill")
	}
}

func TestGenerateEmbeds_SkillsCatalogWithDescription(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
		Skills: []project.Skill{
			{Name: "search-web", Path: "skills/search-web", Description: "Web search capability for finding current information."},
		},
	}

	code, err := GenerateEmbeds(proj, "main")
	if err != nil {
		t.Fatalf("GenerateEmbeds() error = %v", err)
	}

	if !strings.Contains(code, "Web search capability for finding current information.") {
		t.Error("Skills catalog should include skill description")
	}
}

func TestGenerateEmbeds_SystemMessageWithSkills(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
		Skills: []project.Skill{
			{Name: "search-web", Path: "skills/search-web", Description: "Web search capability"},
		},
	}

	code, err := GenerateEmbeds(proj, "main")
	if err != nil {
		t.Fatalf("GenerateEmbeds() error = %v", err)
	}

	if !strings.Contains(code, "getSystemMessage()") {
		t.Error("getSystemMessage function should be generated")
	}

	if !strings.Contains(code, "skillsCatalog") {
		t.Error("getSystemMessage should reference skillsCatalog")
	}

	if !strings.Contains(code, "systemMessage +") {
		t.Error("getSystemMessage should concatenate skills catalog with system message")
	}
}

func TestGenerateEmbeds_NoSkillsNoCatalog(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
		Skills: nil,
	}

	code, err := GenerateEmbeds(proj, "main")
	if err != nil {
		t.Fatalf("GenerateEmbeds() error = %v", err)
	}

	if strings.Contains(code, "skillsCatalog") {
		t.Error("Skills catalog should NOT be generated when no skills exist")
	}

	if strings.Contains(code, "## Available Skills") {
		t.Error("Skills catalog header should NOT be present when no skills")
	}
}

func TestGenerateEmbeds_EmbedImport(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
		Skills: []project.Skill{
			{Name: "test-skill", Path: "skills/test-skill", Description: "Test"},
		},
	}

	code, err := GenerateEmbeds(proj, "main")
	if err != nil {
		t.Fatalf("GenerateEmbeds() error = %v", err)
	}

	if !strings.Contains(code, `"embed"`) {
		t.Error("Should import embed package when skills exist")
	}
}

func TestGenerateEmbeds_MultipleSkillsInCatalog(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
		Skills: []project.Skill{
			{Name: "search-web", Path: "skills/search-web", Description: "Web search"},
			{Name: "analyze-code", Path: "skills/analyze-code", Description: "Code analysis"},
			{Name: "write-tests", Path: "skills/write-tests", Description: "Test generation"},
		},
	}

	code, err := GenerateEmbeds(proj, "main")
	if err != nil {
		t.Fatalf("GenerateEmbeds() error = %v", err)
	}

	for _, skill := range proj.Skills {
		if !strings.Contains(code, "### "+skill.Name) {
			t.Errorf("Skills catalog should include %q", skill.Name)
		}
		if !strings.Contains(code, skill.Description) {
			t.Errorf("Skills catalog should include description for %q", skill.Name)
		}
	}
}

func TestGenerateEmbeds_SafeIdentifierForSkills(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
		Skills: []project.Skill{
			{Name: "search-web", Path: "skills/search-web", Description: "Search"},
			{Name: "code.analysis", Path: "skills/code.analysis", Description: "Analysis"},
		},
	}

	code, err := GenerateEmbeds(proj, "main")
	if err != nil {
		t.Fatalf("GenerateEmbeds() error = %v", err)
	}

	if !strings.Contains(code, "skillSearchWeb") {
		t.Error("Should convert hyphenated skill name to CamelCase")
	}

	if !strings.Contains(code, "skillCodeAnalysis") {
		t.Error("Should convert dotted skill name to CamelCase")
	}
}
