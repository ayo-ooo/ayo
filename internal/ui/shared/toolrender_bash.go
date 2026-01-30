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

	// Build header params: sanitized command, optional flags
	cmd := SanitizeCommand(params.Command)
	out.HeaderParams = []string{cmd}
	if params.RunInBackground {
		out.HeaderParams = append(out.HeaderParams, "background", "true")
	}

	// Build body sections based on state
	if input.State == ToolStateSuccess || input.State == ToolStateError {
		// Try to parse output as bash tool JSON result
		var bashOutput BashToolOutput
		if err := ParseJSON(input.RawOutput, &bashOutput); err == nil {
			// Successfully parsed bash output JSON
			displayOutput := bashOutput.GetDisplayOutput()
			if displayOutput != "" {
				out.Sections = append(out.Sections, RenderSection{
					Type:     SectionCode,
					Content:  displayOutput,
					MaxLines: 50,
				})
			}
			// Update state based on exit code
			if bashOutput.IsError() {
				out.State = ToolStateError
			}
		} else {
			// Fallback: use raw output directly
			if input.RawOutput != "" {
				sectionType := SectionPlain
				trimmed := trimSpaceAndCheck(input.RawOutput)
				if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
					sectionType = SectionJSON
				}

				out.Sections = append(out.Sections, RenderSection{
					Type:     sectionType,
					Content:  input.RawOutput,
					MaxLines: 50,
				})
			}
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
