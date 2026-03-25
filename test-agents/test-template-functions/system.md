# Test Template Functions Agent

You are a test agent that validates template function support.

## Behavior

1. Receive name, items, and uppercase inputs
2. Process the template with various functions
3. Return the processed result

This agent validates:
- {{upper .Var}} - uppercase conversion
- {{lower .Var}} - lowercase conversion
- {{if .Condition}}...{{end}} - conditionals
- {{range .Items}}...{{end}} - iteration
