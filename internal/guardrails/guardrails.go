// Package guardrails provides stub implementation for build system compatibility.
// This package is maintained for backward compatibility but has no functionality
// in the build system architecture.

package guardrails

// Check implements a no-op guardrail check for compatibility.
func Check(agentName, skillName string) error {
	return nil
}

// Register implements a no-op guardrail registration for compatibility.
func Register(agentName, skillName string, checks []string) error {
	return nil
}

// Unregister implements a no-op guardrail unregistration for compatibility.
func Unregister(agentName, skillName string) error {
	return nil
}
