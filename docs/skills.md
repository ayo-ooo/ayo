# Skills Discovery

Skills are reusable capabilities that can be embedded into agents. They are discovered at build time and made available to the model at runtime through the system message.

## Overview

A skill is a directory containing a `SKILL.md` file that describes the capability:

```
skills/
├── search-web/
│   ├── SKILL.md          # Required: Skill description and usage
│   └── scripts/
│       └── search.sh     # Optional: Supporting scripts
└── analyze-code/
    ├── SKILL.md
    └── patterns.json
```

## Build-Time Embedding

Skills are embedded into the binary during `ayo build`:

```go
// Generated in embed.go
//go:embed skills/search-web/*
var skillSearchWeb embed.FS

//go:embed skills/analyze-code/*
var skillAnalyzeCode embed.FS
```

### Embedding Code Generation

Location: `internal/generate/embed.go:38-45`

```go
if len(proj.Skills) > 0 {
    b.WriteString("// Skill files embedded in the binary.\n")
    for _, skill := range proj.Skills {
        safeName := toSafeIdentifier(skill.Name)
        b.WriteString(fmt.Sprintf("//go:embed skills/%s/*\n", skill.Name))
        b.WriteString(fmt.Sprintf("var skill%s embed.FS\n\n", safeName))
    }
}
```

## Runtime Discovery

At runtime, skills must be made discoverable to the model by including them in the system message.

### Skills Catalog Format

A skills catalog is generated and appended to the system message:

```markdown
## Available Skills

### search-web
**Location**: `embedded://skills/search-web/SKILL.md`

Web search capability for finding current information.

**Usage**: When you need to search the web for current information,
read the skill file at `embedded://skills/search-web/SKILL.md` to
understand how to use this capability.

### analyze-code
**Location**: `embedded://skills/analyze-code/SKILL.md`

Code analysis patterns and best practices.

**Usage**: Read `embedded://skills/analyze-code/SKILL.md` for patterns.
```

### Catalog Generation

```go
func generateSkillsCatalog(skills []Skill) string {
    var b strings.Builder
    b.WriteString("## Available Skills\n\n")

    for _, skill := range skills {
        b.WriteString(fmt.Sprintf("### %s\n", skill.Name))
        b.WriteString(fmt.Sprintf("**Location**: `embedded://skills/%s/SKILL.md`\n\n", skill.Name))
        b.WriteString(skill.Description)
        b.WriteString("\n\n")
    }

    return b.String()
}
```

### System Message Injection

The `getSystemMessage()` function must be updated to include the skills catalog:

```go
// Current (broken):
func getSystemMessage() string {
    return systemMessage
}

// Required:
func getSystemMessage() string {
    if len(skillsCatalog) == 0 {
        return systemMessage
    }
    return systemMessage + "\n\n" + skillsCatalog
}
```

## SKILL.md Format

Each skill's `SKILL.md` should follow this structure:

```markdown
# Skill Name

Brief description of what this skill enables.

## When to Use

Describe scenarios where this skill should be invoked.

## How to Use

Step-by-step instructions for the model to follow.

## Examples

Example usage patterns.

## Files

- `scripts/helper.sh` - Supporting script
- `data/patterns.json` - Configuration data
```

## File Access Protocol

Skills use an `embedded://` protocol for file references:

| Path Pattern | Meaning |
|--------------|---------|
| `embedded://skills/{name}/SKILL.md` | Main skill file |
| `embedded://skills/{name}/scripts/*` | Script files |
| `embedded://skills/{name}/data/*` | Data files |

### Reading Embedded Files

Generated code should provide a function to read skill files:

```go
func readSkillFile(skillName, relativePath string) ([]byte, error) {
    switch skillName {
    case "search-web":
        return skillSearchWeb.ReadFile(filepath.Join("skills", skillName, relativePath))
    case "analyze-code":
        return skillAnalyzeCode.ReadFile(filepath.Join("skills", skillName, relativePath))
    default:
        return nil, fmt.Errorf("unknown skill: %s", skillName)
    }
}
```

## Implementation Requirements

### 1. Parse Skills at Build Time

Location: `internal/project/project.go`

```go
type Skill struct {
    Name        string
    Path        string
    Description string // Parsed from SKILL.md first paragraph
}

func (p *Parser) parseSkills() ([]Skill, error) {
    skillsDir := filepath.Join(p.path, "skills")
    entries, _ := os.ReadDir(skillsDir)

    var skills []Skill
    for _, entry := range entries {
        if entry.IsDir() {
            skillPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
            if _, err := os.Stat(skillPath); err == nil {
                desc := extractDescription(skillPath)
                skills = append(skills, Skill{
                    Name:        entry.Name(),
                    Path:        skillPath,
                    Description: desc,
                })
            }
        }
    }
    return skills, nil
}
```

### 2. Generate Skills Catalog

Location: `internal/generate/embed.go`

Add catalog generation to `GenerateEmbeds`:

```go
if len(proj.Skills) > 0 {
    b.WriteString("var skillsCatalog = `\n")
    b.WriteString(generateSkillsCatalog(proj.Skills))
    b.WriteString("`\n\n")
}
```

### 3. Update getSystemMessage

```go
b.WriteString("func getSystemMessage() string {\n")
b.WriteString("\tif len(skillsCatalog) == 0 {\n")
b.WriteString("\t\treturn systemMessage\n")
b.WriteString("\t}\n")
b.WriteString("\treturn systemMessage + \"\\n\\n\" + skillsCatalog\n")
b.WriteString("}\n")
```

## Testing

### Unit Tests

- Parse skills directory correctly
- Generate valid catalog format
- Extract descriptions from SKILL.md

### Integration Tests

- Build agent with skills
- Run agent and verify skills in system message
- Verify embedded files are accessible

## Current Status

Skills are embedded but NOT discoverable:
- ✅ Skills embedded in binary (`embed.go`)
- ❌ Skills catalog not generated
- ❌ System message doesn't include skills
- ❌ File access protocol not implemented
