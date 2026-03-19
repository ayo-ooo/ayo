# File Processing

Handle file inputs with Ayo agents.

## Overview

The `x-cli-file` extension loads file contents into input fields, enabling agents to process files directly.

## Basic File Input

```json
{
  "properties": {
    "file": {
      "type": "string",
      "description": "Path to file",
      "x-cli-position": 1,
      "x-cli-file": true
    }
  }
}
```

Usage:
```bash
./agent document.txt
```

The file contents are loaded and passed to the LLM, not the path.

## Multiple Files

Process multiple files:

```json
{
  "properties": {
    "source": {
      "type": "string",
      "x-cli-position": 1,
      "x-cli-file": true
    },
    "config": {
      "type": "string",
      "x-cli-flag": "config",
      "x-cli-file": true
    }
  }
}
```

Usage:
```bash
./agent main.py --config settings.yaml
```

## File Types

### Text Files

Process text, markdown, code:

```json
{
  "source_code": {
    "type": "string",
    "x-cli-position": 1,
    "x-cli-file": true
  }
}
```

### Configuration Files

Parse YAML, JSON, TOML:

```json
{
  "config": {
    "type": "string",
    "x-cli-file": true,
    "description": "Configuration file (YAML or JSON)"
  }
}
```

### Data Files

Process CSV, JSON data:

```json
{
  "data_file": {
    "type": "string",
    "x-cli-position": 1,
    "x-cli-file": true,
    "description": "Data file to process"
  }
}
```

## Prompt Templates

Use file contents in templates:

```
Process this file:
{{file .file_path}}

{{if .config_path}}
Configuration:
{{file .config_path}}
{{end}}
```

## Example: Code Review

**input.jsonschema**:
```json
{
  "type": "object",
  "properties": {
    "file": {
      "type": "string",
      "description": "Code file to review",
      "x-cli-position": 1,
      "x-cli-file": true
    },
    "language": {
      "type": "string",
      "description": "Programming language",
      "enum": ["go", "python", "javascript", "typescript", "rust"]
    },
    "severity": {
      "type": "string",
      "enum": ["info", "warning", "error", "critical"],
      "default": "warning"
    }
  },
  "required": ["file"]
}
```

**system.md**:
```markdown
# Code Review Agent

You review source code for quality issues.

## Review Process
1. Parse the code
2. Identify issues by severity
3. Suggest improvements
4. Calculate quality score
```

**Usage**:
```bash
./code-review main.go
./code-review app.py --language python --severity error
./code-review Component.tsx -o review.json
```

## Large Files

For large files, consider:

1. **Chunking**: Split into sections
2. **Summarizing**: Pre-process to extract key content
3. **Selective reading**: Focus on relevant parts

### Prompt Guidance

```markdown
For large files:
1. First, identify the overall structure
2. Focus on the most relevant sections
3. Summarize less important parts
```

## Binary Files

Binary files (images, PDFs) require vision-capable models:

```toml
[model]
requires_vision = true
```

## Best Practices

1. **Validate paths**: Check files exist before processing
2. **Handle encoding**: Expect UTF-8, handle others gracefully
3. **Size limits**: Be aware of token limits
4. **Error messages**: Provide clear feedback for missing files

### Error Handling in Prompt

```markdown
If the file cannot be read:
- Explain what went wrong
- Suggest fixes
- Do not make up content
```

## Next Steps

- [Input Schema](../reference/input-schema.md) - x-cli-file reference
- [Prompt Templates](../reference/prompt-templates.md) - file function
- [Examples](../examples/code-review.md) - Code review example
