# Implementation Plan

## Progress
- [ ] Step 1: Create example projects directory structure
- [ ] Step 2: Implement core examples (echo, status-check, summarize, translate, code-review)
- [ ] Step 3: Write getting-started documentation
- [ ] Step 4: Implement advanced examples (research, task-runner, notifier, data-pipeline)
- [ ] Step 5: Write reference documentation
- [ ] Step 6: Write examples gallery
- [ ] Step 7: Write guides
- [ ] Step 8: Add build verification tests

---

## Error Handling Protocol

**CRITICAL**: If any error or unexpected behavior is encountered during example implementation:

1. **STOP** - Do not continue to next example
2. **RESEARCH** - Examine error, read source code, find root cause
3. **PLAN** - Design minimal fix, consider edge cases
4. **VALIDATE** - Review against codebase patterns, check side effects
5. **EXECUTE** - Implement fix with minimal changes
6. **TEST** - Run affected tests, run full suite
7. **VERIFY** - Re-run failing example, confirm resolution
8. **RESUME** - Return to example that failed, continue from there

Loop until all 9 examples build and run successfully.

---

## Step 1: Create Directory Structure

**Objective**: Set up the examples and docs directory structure.

**Implementation**:
```bash
mkdir -p examples/{echo,status-check,summarize,translate,code-review,research,task-runner,notifier,data-pipeline}
mkdir -p docs/{getting-started,reference,guides,examples,api}
```

**Tests Required**:
- Verify all directories exist

**Demo**:
- Show directory tree structure

---

## Step 2: Implement Core Examples

**Objective**: Create the first 5 examples that demonstrate fundamental features.

### Step 2.1: echo (Minimal)

**Files**:
- `examples/echo/config.toml`
- `examples/echo/system.md`
- `examples/echo/.gitignore`

### Step 2.2: status-check (No Input / Autonomous)

**Files**:
- `examples/status-check/config.toml`
- `examples/status-check/system.md`
- `examples/status-check/output.jsonschema`
- `examples/status-check/.gitignore`

### Step 2.3: summarize (Structured Output)

**Files**:
- `examples/summarize/config.toml`
- `examples/summarize/system.md`
- `examples/summarize/input.jsonschema`
- `examples/summarize/output.jsonschema`
- `examples/summarize/.gitignore`

### Step 2.4: translate (CLI Flags)

**Files**:
- `examples/translate/config.toml`
- `examples/translate/system.md`
- `examples/translate/input.jsonschema`
- `examples/translate/.gitignore`

### Step 2.5: code-review (Complex Types + File Input)

**Files**:
- `examples/code-review/config.toml`
- `examples/code-review/system.md`
- `examples/code-review/input.jsonschema`
- `examples/code-review/output.jsonschema`
- `examples/code-review/.gitignore`

**Tests Required**:
- All examples build with `ayo build`
- All JSON schemas are valid

**Demo**:
- Build each example
- Show generated CLI help

---

## Step 3: Write Getting-Started Documentation

**Objective**: Create documentation for new users.

**Files**:
- `docs/README.md` - Landing page
- `docs/getting-started/installation.md`
- `docs/getting-started/quickstart.md`
- `docs/getting-started/first-agent.md`

**Content Requirements**:
- Installation instructions
- 5-minute quickstart using echo example
- Step-by-step first agent creation

**Tests Required**:
- All code blocks are copy-pasteable
- All links are valid

---

## Step 4: Implement Advanced Examples

**Objective**: Create examples demonstrating advanced features.

### Step 4.1: research (Prompt Templates)

**Files**:
- `examples/research/config.toml`
- `examples/research/system.md`
- `examples/research/prompt.tmpl`
- `examples/research/input.jsonschema`
- `examples/research/output.jsonschema`
- `examples/research/.gitignore`

### Step 4.2: task-runner (Skills)

**Files**:
- `examples/task-runner/config.toml`
- `examples/task-runner/system.md`
- `examples/task-runner/input.jsonschema`
- `examples/task-runner/output.jsonschema`
- `examples/task-runner/skills/plan/SKILL.md`
- `examples/task-runner/skills/execute/SKILL.md`
- `examples/task-runner/skills/review/SKILL.md`
- `examples/task-runner/.gitignore`

### Step 4.3: notifier (Hooks)

**Files**:
- `examples/notifier/config.toml`
- `examples/notifier/system.md`
- `examples/notifier/input.jsonschema`
- `examples/notifier/hooks/agent-start`
- `examples/notifier/hooks/agent-finish`
- `examples/notifier/hooks/agent-error`
- `examples/notifier/.gitignore`

### Step 4.4: data-pipeline (Full-Featured)

**Files**:
- `examples/data-pipeline/config.toml`
- `examples/data-pipeline/system.md`
- `examples/data-pipeline/prompt.tmpl`
- `examples/data-pipeline/input.jsonschema`
- `examples/data-pipeline/output.jsonschema`
- `examples/data-pipeline/skills/extract/SKILL.md`
- `examples/data-pipeline/skills/transform/SKILL.md`
- `examples/data-pipeline/skills/validate/SKILL.md`
- `examples/data-pipeline/hooks/agent-start`
- `examples/data-pipeline/hooks/step-start`
- `examples/data-pipeline/hooks/step-finish`
- `examples/data-pipeline/hooks/agent-finish`
- `examples/data-pipeline/.gitignore`

**Tests Required**:
- All examples build successfully
- Hook scripts are executable
- Templates parse without errors

---

## Step 5: Write Reference Documentation

**Objective**: Create comprehensive reference docs.

**Files**:
- `docs/reference/project-structure.md`
- `docs/reference/config.md`
- `docs/reference/input-schema.md`
- `docs/reference/output-schema.md`
- `docs/reference/cli-flags.md`
- `docs/reference/prompt-templates.md`
- `docs/reference/skills.md`
- `docs/reference/hooks.md`
- `docs/reference/generated-code.md`

**Content Requirements**:
- Complete coverage of all options
- Examples from example projects
- Tables for quick reference
- Code snippets for all features

---

## Step 6: Write Examples Gallery

**Objective**: Create browsable documentation for all examples.

**Files**:
- `docs/examples/README.md`
- `docs/examples/echo.md`
- `docs/examples/status-check.md`
- `docs/examples/summarize.md`
- `docs/examples/translate.md`
- `docs/examples/code-review.md`
- `docs/examples/research.md`
- `docs/examples/task-runner.md`
- `docs/examples/notifier.md`
- `docs/examples/data-pipeline.md`

**Content Requirements**:
- Feature matrix
- Use case categorization
- Individual example pages with:
  - Overview
  - Files breakdown
  - Usage examples
  - Key features demonstrated

---

## Step 7: Write Guides

**Objective**: Create how-to guides for common tasks.

**Files**:
- `docs/guides/building-agents.md`
- `docs/guides/structured-outputs.md`
- `docs/guides/file-processing.md`
- `docs/guides/integrations.md`
- `docs/guides/best-practices.md`

**Content Requirements**:
- Step-by-step instructions
- Decision guidance
- Common patterns
- Anti-patterns to avoid

---

## Step 8: Add Build Verification Tests

**Objective**: Ensure examples remain buildable.

**Implementation**:
- Create `examples/test_examples.sh` script
- Add to CI pipeline
- Verify all examples:
  - Parse successfully
  - Build without errors
  - Generate expected files

**Tests Required**:
- Script exits 0 on success
- Catches build failures
- Reports which example failed
