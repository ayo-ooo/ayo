package interactive

// TTYCheckCode returns Go code for TTY and interactive mode detection.
// This code is embedded in generated CLIs.
func TTYCheckCode() string {
	return `// IsTerminal returns true if stdout is connected to a terminal.
func IsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// IsInteractiveMode checks if the agent should use interactive forms.
// Returns false if:
// - stdout is not a TTY (piped or redirected)
// - --non-interactive flag is set
// - agent config has interactive: false
func IsInteractiveMode(agentInteractive bool, nonInteractiveFlag bool) bool {
	if nonInteractiveFlag {
		return false
	}
	if !agentInteractive {
		return false
	}
	return IsTerminal()
}
`
}

// MissingArgsError generates styled error for missing required arguments.
func MissingArgsError() string {
	return `func formatMissingArgsError(schemaProps map[string]PropertyInfo, missing []string) string {
	var b strings.Builder
	b.WriteString("Missing required arguments:\n\n")

	for _, name := range missing {
		info := schemaProps[name]
		desc := info.Description
		if desc == "" {
			desc = name
		}
		b.WriteString(fmt.Sprintf("  --%-15s %s\n", name, desc))
	}

	b.WriteString("\nRun with --help for usage information.")
	return b.String()
}
`
}

// PropertyInfo describes a schema property for error messages.
type PropertyInfo struct {
	Name        string
	Description string
	Type        string
	Required    bool
}
