package shared

// BashRenderer renders bash tool calls.
type BashRenderer struct{}

// Name returns "bash".
func (b *BashRenderer) Name() string { return "bash" }

// Render produces render output for bash tool calls.
func (b *BashRenderer) Render(input ToolRenderInput) ToolRenderOutput {
	out := ToolRenderOutput{
		Label:    "Bash",
		State:    input.State,
		IsNested: input.ParentID != "",
	}

	// Parse parameters
	var params BashParams
	if err := ParseJSON(input.RawInput, &params); err == nil {
		input.Params = params
	} else {
		params = BashParams{}
	}

	// Parse metadata
	var meta BashResponseMetadata
	if err := ParseJSON(input.RawMetadata, &meta); err == nil {
		input.Metadata = meta
	}

	// Build header params: sanitized command, optional flags
	cmd := SanitizeCommand(params.Command)
	out.HeaderParams = []string{cmd}
	if params.RunInBackground {
		out.HeaderParams = append(out.HeaderParams, "background", "true")
	}

	// Build body sections based on state
	if input.State == ToolStateSuccess || input.State == ToolStateError {
		// Get output from metadata or raw output
		output := meta.Output
		if output == "" && input.RawOutput != "" {
			output = input.RawOutput
		}

		if output != "" {
			// Detect if JSON and render appropriately
			sectionType := SectionPlain
			trimmed := trimSpaceAndCheck(output)
			if (len(trimmed) > 0 && trimmed[0] == '{') || (len(trimmed) > 0 && trimmed[0] == '[') {
				sectionType = SectionJSON
			}

			out.Sections = append(out.Sections, RenderSection{
				Type:     sectionType,
				Content:  output,
				MaxLines: 50,
			})
		}
	}

	return out
}

// trimSpaceAndCheck returns trimmed string for prefix checking
func trimSpaceAndCheck(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' && s[i] != '\n' && s[i] != '\r' {
			return s[i:]
		}
	}
	return ""
}

func init() {
	RegisterToolRenderer(&BashRenderer{})
}
