---
id: ase-3iuw
status: closed
deps: [ase-lbg7]
links: []
created: 2026-02-09T03:07:18Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-k48b
---
# Implement flow parser and validator

Parse flow YAML files and validate against the spec.

## Background

Flow files are YAML with a defined schema. The parser must:
- Load and parse YAML
- Validate against spec (required fields, types, valid references)
- Resolve template expressions to validate they reference valid steps
- Return structured Flow object

## Implementation

1. Create Flow struct matching spec:
   ```go
   type Flow struct {
       Version     int                    `yaml:"version"`
       Name        string                 `yaml:"name"`
       Description string                 `yaml:"description"`
       CreatedBy   string                 `yaml:"created_by"`
       CreatedAt   time.Time              `yaml:"created_at"`
       Input       *jsonschema.Schema     `yaml:"input,omitempty"`
       Output      *jsonschema.Schema     `yaml:"output,omitempty"`
       Steps       []Step                 `yaml:"steps"`
       Triggers    []FlowTrigger          `yaml:"triggers,omitempty"`
   }
   
   type Step struct {
       ID      string `yaml:"id"`
       Type    string `yaml:"type"`  // 'shell' or 'agent'
       // Shell fields
       Run     string `yaml:"run,omitempty"`
       // Agent fields
       Agent   string `yaml:"agent,omitempty"`
       Prompt  string `yaml:"prompt,omitempty"`
       Context string `yaml:"context,omitempty"`
       Input   string `yaml:"input,omitempty"`
       // Common
       When    string `yaml:"when,omitempty"`
   }
   ```

2. Create parser:
   ```go
   func ParseFlow(path string) (*Flow, error)
   func ParseFlowBytes(data []byte) (*Flow, error)
   ```

3. Create validator:
   ```go
   func (f *Flow) Validate() error
   ```
   
   Validations:
   - All required fields present
   - Step IDs unique
   - Step references ({{ steps.X }}) point to earlier steps
   - Agent handles are valid format
   - Trigger schedules are valid cron expressions
   - Input/output schemas are valid JSON Schema

4. Template expression parser (for validation only, not execution):
   ```go
   func ValidateTemplate(expr string, availableSteps []string) error
   ```

## Files to create

- internal/flows/types.go (structs)
- internal/flows/parse.go (parser)
- internal/flows/validate.go (validator)
- internal/flows/template.go (template expression handling)

## Acceptance Criteria

- Valid flows parse successfully
- Invalid flows return clear error messages
- Step references validated
- Cron expressions validated
- Template expressions validated

