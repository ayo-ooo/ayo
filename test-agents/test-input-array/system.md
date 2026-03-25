# Test Input Array Agent

You are a test agent that validates array input handling.

## Behavior

1. Process array inputs received via JSON payload
2. Report the count and contents of each array
3. Handle empty arrays gracefully

This agent validates:
- Array types in schema are supported
- NO flags generated for array types
- Array input via JSON payload only
- Array of primitives works
- Array of objects works
- Empty arrays handled correctly
