# Output Formatting Behavior

Generated agents handle output differently based on whether structured output is configured.

## Output Modes

### Structured Output Mode

When `output.jsonschema` is present in the agent configuration:

- Output is JSON-encoded matching the schema
- Validates against the schema before output
- Enables type-safe consumption by downstream tools

```bash
# Example: agent with output.jsonschema
$ my-agent --input "data"
{"result": "processed", "count": 42}
```

### Text Output Mode

When no `output.jsonschema` is defined:

- Output is raw text from the model response
- NO JSON encoding applied
- Preserves formatting, newlines, and structure from model output

```bash
# Example: text-only agent
$ my-agent "tell me a joke"
Why did the chicken cross the road?
To get to the other side!
```

## File Output

The `--output` or `-o` flag writes output to a file:

```bash
# Write to file
$ my-agent "task" --output result.txt

# Short form
$ my-agent "task" -o result.json
```

File output uses the same formatting as stdout based on the output mode.

## Current Issue

**Bug**: The generated CLI currently JSON-encodes ALL output, including text mode.

**Location**: `internal/generate/cli.go:82-88`

```go
// Current (broken):
output, err := json.Marshal(result)  // Always JSON-encodes!
fmt.Println(string(output))
```

**Expected Behavior**:
- Structured mode: `json.Marshal(result)` then output
- Text mode: Output `result` directly without encoding

## Implementation Requirements

1. Check if `proj.Output` is nil in `GenerateCLI`
2. Generate conditional output logic:
   ```go
   if hasOutputSchema {
       output, err := json.Marshal(result)
       fmt.Println(string(output))
   } else {
       fmt.Println(result)
   }
   ```

3. File output must follow the same logic

## Error Handling

| Error | Cause | Output |
|-------|-------|--------|
| Schema validation failure | Output doesn't match schema | Error to stderr, exit 1 |
| JSON marshal failure | Type not serializable | Error to stderr, exit 1 |
| File write failure | Permission denied, disk full | Error to stderr, exit 1 |
