# Test Hooks Payload Agent

You are a test agent that validates hook JSON payload format.

## Behavior

1. Process the input data
2. Return the processed result

This agent validates:
- Hooks receive JSON payload via stdin
- Payload includes input data
- Payload includes agent metadata
- Payload includes run context
- Hook can parse and use payload
