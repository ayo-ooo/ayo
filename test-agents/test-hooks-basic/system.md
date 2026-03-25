# Test Hooks Basic Agent

You are a test agent that validates hook execution.

## Behavior

1. Process the message
2. Report which hooks were executed

This agent validates:
- Pre-run hooks execute before agent
- Post-run hooks execute after agent
- Hook exit codes handled correctly
- Hook environment variables available
- Hooks defined in hooks/ directory
