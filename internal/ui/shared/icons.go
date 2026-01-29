package shared

// Icons for UI elements - colorless Unicode glyphs that respect terminal themes.
// These are carefully selected from box drawing, geometric shapes, and
// miscellaneous symbols to be visually distinct and terminal-compatible.
const (
	// Status icons
	IconSuccess  = "✓" // Check mark - success
	IconError    = "✗" // Ballot X - error
	IconWarning  = "△" // White up-pointing triangle - warning
	IconInfo     = "●" // Black circle - info
	IconCheck    = "✓" // Check mark - completed (alias for Success)
	IconCross    = "✗" // Ballot X - failed (alias for Error)
	IconPending  = "○" // White circle - pending
	IconComplete = "●" // Black circle - complete
	IconSpinner  = "◐" // Circle with left half black - in progress

	// Tool icons
	IconTool     = "▶" // Black right-pointing triangle - tool execution
	IconBash     = "❯" // Heavy right angle bracket - bash/shell prompt
	IconThinking = "◇" // White diamond - thinking/reasoning

	// Navigation/UI icons
	IconArrowRight = "→" // Rightwards arrow - navigation
	IconBullet     = "•" // Bullet - list items
	IconEllipsis   = "⋯" // Midline horizontal ellipsis - loading/truncated
	IconMenu       = "≡" // Identical to - menu (hamburger alternative)

	// Agent icons
	IconAgent    = "◆" // Black diamond - agent
	IconSubAgent = "▹" // White right-pointing small triangle - sub-agent

	// Todo icons
	IconTodoCompleted  = "✓" // Check mark - completed task
	IconTodoInProgress = "▸" // Small right-pointing triangle - in progress
	IconTodoPending    = "○" // White circle - pending task
	IconPlan           = "□" // White square - plan/todo

	// Tool state icons (used in TUI renderers)
	ToolPending = "○" // White circle - pending
	ToolSuccess = "●" // Black circle - success
	ToolError   = "×" // Multiplication sign - error
	ToolRunning = "◐" // Circle with left half black - running
)
