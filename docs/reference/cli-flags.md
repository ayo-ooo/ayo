# CLI Flags

Customize command-line flag generation from input schemas.

## Default Behavior

Without customization, Ayo generates flags based on property names:

```json
{
  "source_language": {
    "type": "string"
  }
}
```

Generates: `--source-language string`

## x-cli-flag

Customize the flag name:

```json
{
  "from": {
    "type": "string",
    "description": "Source language",
    "x-cli-flag": "source-language"
  }
}
```

Usage: `./translate "Hello" --source-language english`

### Use Cases

- More descriptive names (`from` → `--source-language`)
- Kebab-case convention
- Avoiding reserved words

## x-cli-short

Add a short flag:

```json
{
  "format": {
    "type": "string",
    "x-cli-short": "-f"
  }
}
```

Usage: `./agent -f json` or `./agent --format json`

### Combined with x-cli-flag

```json
{
  "to": {
    "type": "string",
    "description": "Target language",
    "x-cli-flag": "target-language",
    "x-cli-short": "-t"
  }
}
```

Usage: `./translate "Hello" -t spanish`

## x-cli-position

Make a field a positional argument:

```json
{
  "text": {
    "type": "string",
    "description": "Text to process",
    "x-cli-position": 1
  },
  "output": {
    "type": "string",
    "description": "Output file",
    "x-cli-position": 2
  }
}
```

Usage: `./agent "input text" output.json`

### Position Ordering

- Positions are 1-indexed
- Arguments are processed in position order
- Only one field per position

### Best Practices

- Position 1: Primary input (text, file, query)
- Position 2: Secondary input (output file, target)
- Use flags for options and modifiers

## x-cli-file

Load file contents into a field:

```json
{
  "config": {
    "type": "string",
    "description": "Configuration file",
    "x-cli-file": true
  }
}
```

When the user provides a path, the file contents are loaded:

```bash
./agent --config settings.yaml
```

The LLM receives the file contents, not the path.

### Use Cases

- Configuration files
- Data files for processing
- Context documents

## Complete Example

From the translate example:

```json
{
  "type": "object",
  "properties": {
    "text": {
      "type": "string",
      "description": "Text to translate",
      "x-cli-position": 1
    },
    "from": {
      "type": "string",
      "description": "Source language",
      "x-cli-flag": "source-language",
      "x-cli-short": "-s",
      "default": "auto"
    },
    "to": {
      "type": "string",
      "description": "Target language",
      "x-cli-short": "-t"
    },
    "formal": {
      "type": "boolean",
      "description": "Use formal tone",
      "default": false
    }
  },
  "required": ["to"]
}
```

Generates CLI:

```
Usage:
  translate [text] [flags]

Flags:
      --formal                  Use formal tone (default false)
  -h, --help                    help for translate
  -o, --output string           Write output to file
      --provider string         AI provider to use
  -s, --source-language string  Source language (default "auto")
  -t, --to string               Target language
```

Usage examples:

```bash
# Basic
translate "Hello" --to spanish

# Short flags
translate "Hello" -t spanish

# With options
translate "Hello" -s english -t german --formal

# With output file
translate "Hello" --to french -o result.json
```

## Type-Specific Flags

| Type | Flag Type | Example |
|------|-----------|---------|
| `string` | StringVar | `--name string` |
| `integer` | IntVar | `--count 10` |
| `number` | Float64Var | `--ratio 0.5` |
| `boolean` | BoolVar | `--verbose` |

## Required vs Optional

Fields in the `required` array must be provided:

```json
{
  "required": ["text", "to"]
}
```

Optional fields can be omitted. Use defaults for optional fields:

```json
{
  "format": {
    "type": "string",
    "default": "json"
  }
}
```

## Next Steps

- [Input Schema](input-schema.md) - Full schema reference
- [Examples](../examples/README.md) - See complete examples
