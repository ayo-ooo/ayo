package interactive

import (
	"fmt"
	"strings"

	"github.com/ayo-ooo/ayo/internal/schema"
)

// GenerateInteractiveCode generates Go code for interactive form handling.
// This code is embedded in generated CLIs when the agent has input.jsonschema.
func GenerateInteractiveCode(inputSchema *schema.ParsedSchema, agentInteractive bool, inputOrder []string) string {
	var b strings.Builder

	// Generate form imports
	b.WriteString("// Form imports (for interactive mode)\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"os\"\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"strconv\"\n")
	b.WriteString("\t\"strings\"\n")
	b.WriteString("\n")
	b.WriteString("\t\"charm.land/huh/v2\"\n")
	b.WriteString(")\n\n")

	// Generate form configuration
	b.WriteString("// Form configuration\n")
	b.WriteString(fmt.Sprintf("var agentInteractive = %v\n", agentInteractive))
	b.WriteString("var nonInteractive bool\n\n")

	// Generate property info map for error messages
	b.WriteString("// Property descriptions for error messages\n")
	b.WriteString("var propertyInfo = map[string]struct {\n")
	b.WriteString("\tdescription string\n")
	b.WriteString("\trequired    bool\n")
	b.WriteString("}{\n")

	for name, prop := range inputSchema.Properties {
		required := false
		for _, r := range inputSchema.Required {
			if r == name {
				required = true
				break
			}
		}
		b.WriteString(fmt.Sprintf("\t%q: {%q, %v},\n", name, escapeStr(prop.Description), required))
	}
	b.WriteString("}\n\n")

	// Generate IsTerminal function
	b.WriteString(`// isTerminal returns true if stdout is connected to a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

`)

	// Generate interactive mode check
	b.WriteString(`// shouldShowForm checks if we should display the interactive form.
func shouldShowForm() bool {
	if nonInteractive {
		return false
	}
	if !agentInteractive {
		return false
	}
	return isTerminal()
}

`)

	// Generate missing args error
	b.WriteString(`// formatMissingArgs creates a styled error for missing arguments.
func formatMissingArgs(missing []string) string {
	var b strings.Builder
	b.WriteString("Missing required arguments:\n\n")

	for _, name := range missing {
		info := propertyInfo[name]
		desc := info.description
		if desc == "" {
			desc = name
		}
		b.WriteString(fmt.Sprintf("  --%-15s %s\n", name, desc))
	}

	b.WriteString("\nRun with --help for usage information.")
	return b.String()
}

`)

	// Generate form creation function
	b.WriteString("// createForm creates the interactive form from the schema.\n")
	b.WriteString("func createForm(prefill map[string]any) *huh.Form {\n")
	b.WriteString("\tvar fields []huh.Field\n\n")

	// Generate fields in order
	if len(inputOrder) > 0 {
		b.WriteString("\t// Field order from config\n")
		b.WriteString("\torder := []string{")
		for i, name := range inputOrder {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%q", name))
		}
		b.WriteString("}\n\n")
	} else {
		b.WriteString("\t// Default field order\n")
		b.WriteString("\torder := []string{")
		i := 0
		for name := range inputSchema.Properties {
			if schema.IsPrimitiveType(inputSchema.Properties[name].Type) || len(inputSchema.Properties[name].Enum) > 0 {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(fmt.Sprintf("%q", name))
				i++
			}
		}
		b.WriteString("}\n\n")
	}

	// Track which fields we've added
	b.WriteString("\tadded := make(map[string]bool)\n")
	b.WriteString("\tfor _, name := range order {\n")
	b.WriteString("\t\tadded[name] = true\n")

	// Generate field switch
	b.WriteString("\t\tswitch name {\n")
	for name, prop := range inputSchema.Properties {
		b.WriteString(generateFieldCase(name, prop, inputSchema))
	}
	b.WriteString("\t\t}\n")
	b.WriteString("\t}\n\n")

	// Add any fields not in order
	b.WriteString("\t// Add remaining fields not in order\n")
	b.WriteString("\tfor name := range propertyInfo {\n")
	b.WriteString("\t\tif !added[name] {\n")
	b.WriteString("\t\t\tswitch name {\n")
	for name, prop := range inputSchema.Properties {
		b.WriteString(generateFieldCase(name, prop, inputSchema))
	}
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t}\n\n")

	b.WriteString("\treturn huh.NewForm(huh.NewGroup(fields...))\n")
	b.WriteString("}\n\n")

	// Generate collect results function
	b.WriteString(generateCollectResults(inputSchema))

	return b.String()
}

func generateFieldCase(name string, prop schema.Property, inputSchema *schema.ParsedSchema) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\t\tcase %q:\n", name))

	// Check if required
	required := false
	for _, r := range inputSchema.Required {
		if r == name {
			required = true
			break
		}
	}

	title := prop.Description
	if title == "" {
		title = name
	}

	switch {
	case len(prop.Enum) > 0 && prop.Type == "string":
		// Select field
		b.WriteString(fmt.Sprintf("\t\t\tvar %sVal string\n", name))
		b.WriteString(fmt.Sprintf("\t\t\tif v, ok := prefill[%q].(string); ok {\n", name))
		b.WriteString(fmt.Sprintf("\t\t\t\t%sVal = v\n", name))
		b.WriteString("\t\t\t}\n")
		b.WriteString("\t\t\toptions := []huh.Option[string]{\n")
		for _, opt := range prop.Enum {
			b.WriteString(fmt.Sprintf("\t\t\t\thuh.NewOption(%q, %q),\n", opt, opt))
		}
		b.WriteString("\t\t\t}\n")
		b.WriteString(fmt.Sprintf("\t\t\tfields = append(fields, huh.NewSelect[string]().Title(%q).Key(%q).Options(options...).Value(&%sVal))\n", title, name, name))

	case len(prop.Enum) > 0 && prop.Type == "array":
		// Multi-select field
		b.WriteString(fmt.Sprintf("\t\t\tvar %sVal []string\n", name))
		b.WriteString(fmt.Sprintf("\t\t\tif v, ok := prefill[%q].([]string); ok {\n", name))
		b.WriteString(fmt.Sprintf("\t\t\t\t%sVal = v\n", name))
		b.WriteString("\t\t\t}\n")
		b.WriteString("\t\t\toptions := []huh.Option[string]{\n")
		for _, opt := range prop.Enum {
			b.WriteString(fmt.Sprintf("\t\t\t\thuh.NewOption(%q, %q),\n", opt, opt))
		}
		b.WriteString("\t\t\t}\n")
		b.WriteString(fmt.Sprintf("\t\t\tfields = append(fields, huh.NewMultiSelect[string]().Title(%q).Key(%q).Options(options...).Value(&%sVal))\n", title, name, name))

	case prop.Type == "boolean":
		// Confirm field
		defVal := "false"
		if b, ok := prop.Default.(bool); ok && b {
			defVal = "true"
		}
		b.WriteString(fmt.Sprintf("\t\t\t%sVal := %s\n", name, defVal))
		b.WriteString(fmt.Sprintf("\t\t\tif v, ok := prefill[%q].(bool); ok {\n", name))
		b.WriteString(fmt.Sprintf("\t\t\t\t%sVal = v\n", name))
		b.WriteString("\t\t\t}\n")
		b.WriteString(fmt.Sprintf("\t\t\tfields = append(fields, huh.NewConfirm().Title(%q).Key(%q).Value(&%sVal))\n", title, name, name))

	case prop.Type == "integer" || prop.Type == "number":
		// Numeric input
		defVal := "0"
		if f, ok := prop.Default.(float64); ok {
			defVal = fmt.Sprintf("%v", f)
		}
		b.WriteString(fmt.Sprintf("\t\t\t%sVal := %q\n", name, defVal))
		b.WriteString(fmt.Sprintf("\t\t\tif v, ok := prefill[%q].(string); ok {\n", name))
		b.WriteString(fmt.Sprintf("\t\t\t\t%sVal = v\n", name))
		b.WriteString(fmt.Sprintf("\t\t\t} else if v, ok := prefill[%q].(int); ok {\n", name))
		b.WriteString(fmt.Sprintf("\t\t\t\t%sVal = strconv.Itoa(v)\n", name))
		b.WriteString(fmt.Sprintf("\t\t\t} else if v, ok := prefill[%q].(float64); ok {\n", name))
		b.WriteString(fmt.Sprintf("\t\t\t\t%sVal = strconv.FormatFloat(v, 'f', -1, 64)\n", name))
		b.WriteString("\t\t\t}\n")

		validation := ""
		if required {
			validation = fmt.Sprintf(".Validate(func(s string) error { if s == \"\" { return fmt.Errorf(\"required\") }; %s; return nil })", prop.Type+"(s)")
		}
		b.WriteString(fmt.Sprintf("\t\t\tfields = append(fields, huh.NewInput().Title(%q).Key(%q).Value(&%sVal)%s)\n", title, name, name, validation))

	case prop.Type == "string":
		// Text input
		defVal := ""
		if s, ok := prop.Default.(string); ok {
			defVal = s
		}
		b.WriteString(fmt.Sprintf("\t\t\t%sVal := %q\n", name, defVal))
		b.WriteString(fmt.Sprintf("\t\t\tif v, ok := prefill[%q].(string); ok {\n", name))
		b.WriteString(fmt.Sprintf("\t\t\t\t%sVal = v\n", name))
		b.WriteString("\t\t\t}\n")

		validation := ""
		if required {
			validation = ".Validate(huh.ValidateNotEmpty())"
		}
		b.WriteString(fmt.Sprintf("\t\t\tfields = append(fields, huh.NewInput().Title(%q).Key(%q).Value(&%sVal)%s)\n", title, name, name, validation))
	}

	return b.String()
}

func generateCollectResults(inputSchema *schema.ParsedSchema) string {
	var b strings.Builder

	b.WriteString(`// collectFormResults extracts values from the form and returns a map.
func collectFormResults(form *huh.Form) map[string]any {
	results := make(map[string]any)

`)

	for name, prop := range inputSchema.Properties {
		switch prop.Type {
		case "boolean":
			b.WriteString(fmt.Sprintf("\tif v := form.Get(%q); v != nil {\n", name))
			b.WriteString(fmt.Sprintf("\t\tresults[%q] = v\n", name))
			b.WriteString("\t}\n")
		case "string":
			b.WriteString(fmt.Sprintf("\tif v := form.GetString(%q); v != \"\" {\n", name))
			b.WriteString(fmt.Sprintf("\t\tresults[%q] = v\n", name))
			b.WriteString("\t}\n")
		case "integer":
			b.WriteString(fmt.Sprintf("\tif v := form.GetString(%q); v != \"\" {\n", name))
			b.WriteString("\t\tif i, err := strconv.Atoi(v); err == nil {\n")
			b.WriteString(fmt.Sprintf("\t\t\tresults[%q] = i\n", name))
			b.WriteString("\t\t}\n")
			b.WriteString("\t}\n")
		case "number":
			b.WriteString(fmt.Sprintf("\tif v := form.GetString(%q); v != \"\" {\n", name))
			b.WriteString("\t\tif f, err := strconv.ParseFloat(v, 64); err == nil {\n")
			b.WriteString(fmt.Sprintf("\t\t\tresults[%q] = f\n", name))
			b.WriteString("\t\t}\n")
			b.WriteString("\t}\n")
		case "array":
			b.WriteString(fmt.Sprintf("\tif v := form.Get(%q); v != nil {\n", name))
			b.WriteString(fmt.Sprintf("\t\tresults[%q] = v\n", name))
			b.WriteString("\t}\n")
		}
	}

	b.WriteString("\treturn results\n")
	b.WriteString("}\n")

	return b.String()
}

func escapeStr(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}
