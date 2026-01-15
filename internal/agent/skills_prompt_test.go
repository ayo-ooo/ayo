package agent

import (
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/skills"
)

func TestBuildSkillsPromptEmpty(t *testing.T) {
	result := buildSkillsPrompt(nil)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}

	result = buildSkillsPrompt([]skills.Metadata{})
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestBuildSkillsPromptBasic(t *testing.T) {
	metas := []skills.Metadata{
		{Name: "skill-b", Description: "Description B", Path: "/path/to/b/SKILL.md"},
		{Name: "skill-a", Description: "Description A", Path: "/path/to/a/SKILL.md"},
	}

	result := buildSkillsPrompt(metas)

	// Should contain available_skills tags
	if !strings.Contains(result, "<available_skills>") {
		t.Error("missing <available_skills> tag")
	}
	if !strings.Contains(result, "</available_skills>") {
		t.Error("missing </available_skills> tag")
	}

	// Should contain activation hint
	if !strings.Contains(result, "cat <location>") {
		t.Error("missing activation hint")
	}

	// Should be sorted (skill-a before skill-b)
	aIdx := strings.Index(result, "skill-a")
	bIdx := strings.Index(result, "skill-b")
	if aIdx > bIdx {
		t.Error("skills should be sorted alphabetically")
	}

	// Should contain skill elements
	if !strings.Contains(result, "<name>skill-a</name>") {
		t.Error("missing skill-a name")
	}
	if !strings.Contains(result, "<description>Description A</description>") {
		t.Error("missing skill-a description")
	}
	if !strings.Contains(result, "<location>/path/to/a/SKILL.md</location>") {
		t.Error("missing skill-a location")
	}
}

func TestBuildSkillsPromptWithResources(t *testing.T) {
	metas := []skills.Metadata{
		{
			Name:        "skill-with-resources",
			Description: "Has resources",
			Path:        "/path/SKILL.md",
			HasScripts:  true,
			HasRefs:     true,
			HasAssets:   false,
		},
	}

	result := buildSkillsPrompt(metas)

	if !strings.Contains(result, "<resources>") {
		t.Error("missing resources tag")
	}
	if !strings.Contains(result, "scripts/") {
		t.Error("missing scripts/ in resources")
	}
	if !strings.Contains(result, "references/") {
		t.Error("missing references/ in resources")
	}
	if strings.Contains(result, "assets/") {
		t.Error("assets/ should not be in resources")
	}
}

func TestBuildSkillsPromptNoResources(t *testing.T) {
	metas := []skills.Metadata{
		{
			Name:        "simple-skill",
			Description: "No resources",
			Path:        "/path/SKILL.md",
			HasScripts:  false,
			HasRefs:     false,
			HasAssets:   false,
		},
	}

	result := buildSkillsPrompt(metas)

	if strings.Contains(result, "<resources>") {
		t.Error("should not have resources tag when no resources exist")
	}
}

func TestBuildSkillsPromptXMLEscaping(t *testing.T) {
	metas := []skills.Metadata{
		{
			Name:        "escape-test",
			Description: "Description with <special> & \"chars\"",
			Path:        "/path/to/skill & more/SKILL.md",
		},
	}

	result := buildSkillsPrompt(metas)

	if strings.Contains(result, "<special>") {
		t.Error("< should be escaped")
	}
	if !strings.Contains(result, "&lt;special&gt;") {
		t.Error("should contain escaped special tag")
	}
	if strings.Contains(result, "& ") && !strings.Contains(result, "&amp;") {
		t.Error("& should be escaped")
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"<tag>", "&lt;tag&gt;"},
		{"a & b", "a &amp; b"},
		{"\"quoted\"", "&quot;quoted&quot;"},
		{"it's", "it&apos;s"},
		{"<a & b>", "&lt;a &amp; b&gt;"},
	}

	for _, tt := range tests {
		result := escapeXML(tt.input)
		if result != tt.expected {
			t.Errorf("escapeXML(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
