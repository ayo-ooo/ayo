package agent

import (
	"sort"
	"strings"

	"github.com/alexcabrera/ayo/internal/skills"
)

func buildSkillsPrompt(metas []skills.Metadata) string {
	if len(metas) == 0 {
		return ""
	}
	sorted := make([]skills.Metadata, len(metas))
	copy(sorted, metas)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	var b strings.Builder
	b.WriteString("<available_skills>\n")
	b.WriteString("When a user's request matches a skill's description, read the skill file to get detailed instructions.\n")
	b.WriteString("Use: cat <location> to read the full skill instructions.\n\n")

	for _, m := range sorted {
		b.WriteString("  <skill>\n")
		b.WriteString("    <name>" + escapeXML(m.Name) + "</name>\n")
		b.WriteString("    <description>" + escapeXML(m.Description) + "</description>\n")
		if m.Path != "" {
			b.WriteString("    <location>" + escapeXML(m.Path) + "</location>\n")
		}
		
		// Add resource hints if any optional directories exist
		var resources []string
		if m.HasScripts {
			resources = append(resources, "scripts/")
		}
		if m.HasRefs {
			resources = append(resources, "references/")
		}
		if m.HasAssets {
			resources = append(resources, "assets/")
		}
		if len(resources) > 0 {
			b.WriteString("    <resources>" + escapeXML(strings.Join(resources, ", ")) + "</resources>\n")
		}
		
		b.WriteString("  </skill>\n")
	}
	b.WriteString("</available_skills>")
	return b.String()
}

func escapeXML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(s)
}
