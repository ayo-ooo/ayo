package shared

// SpinnerType identifies different kinds of async operations.
type SpinnerType int

const (
	// SpinnerAgent is for LLM thinking/generation (default)
	SpinnerAgent SpinnerType = iota
	// SpinnerTool is for tool execution
	SpinnerTool
	// SpinnerMemory is for memory operations (embed, store, search)
	SpinnerMemory
	// SpinnerReasoning is for extended thinking/reasoning
	SpinnerReasoning
	// SpinnerSystem is for system operations (setup, install)
	SpinnerSystem
)

// Spinner animation frames by type.
var (
	// Default braille spinner for agent thinking
	FramesAgent = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	// Dots for tool execution
	FramesTool = []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●"}
	// Pulse for memory operations
	FramesMemory = []string{"◇", "◈", "◆", "◈"}
	// Meter for reasoning
	FramesReasoning = []string{"▱▱▱", "▰▱▱", "▰▰▱", "▰▰▰", "▰▰▱", "▰▱▱"}
	// Line for system ops
	FramesSystem = []string{"|", "/", "-", "\\"}
)

// SpinnerFrames is the default animation frames (agent/thinking).
var SpinnerFrames = FramesAgent

// GetSpinnerFrames returns the animation frames for a given spinner type.
func GetSpinnerFrames(t SpinnerType) []string {
	switch t {
	case SpinnerTool:
		return FramesTool
	case SpinnerMemory:
		return FramesMemory
	case SpinnerReasoning:
		return FramesReasoning
	case SpinnerSystem:
		return FramesSystem
	default:
		return FramesAgent
	}
}
