---
name: planning
description: Systematic approach to breaking down complex requests into phases, tasks, and atomic todo items using the plan tool.
metadata:
  author: ayo
  version: "1.0"
---

# Planning Skill

This skill provides a systematic methodology for decomposing complex requests into manageable work units using the `plan` tool.

## Core Concepts

### Hierarchy Levels

1. **Phases** - High-level stages of work (optional)
   - Use when work spans multiple distinct stages
   - **Rule**: If using phases, must have at least 2
   - Examples: "Setup", "Implementation", "Testing", "Deployment"

2. **Tasks** - Units of work that produce a deliverable
   - Always required (either at top level or within phases)
   - Should be completable in a reasonable time
   - Each task needs `content` (imperative) and `active_form` (present continuous)

3. **Todos** - Atomic sub-items within tasks (optional)
   - Use for granular step-by-step tracking
   - Should be very small, concrete actions
   - Helps when a task has multiple sequential steps

## When to Use Each Level

### Use Phases When:
- Project has distinct stages that build on each other
- Work will span multiple sessions or significant time
- Stakeholder wants to see high-level progress
- Work involves different types of activities (research vs implementation vs testing)

### Use Just Tasks When:
- Work is medium complexity with clear linear steps
- Can be completed in a single session
- Steps are roughly equal in scope
- **This is the most common case**

### Use Todos Within Tasks When:
- A task has multiple atomic steps that should be tracked
- You want granular visibility into progress
- Steps are too small to be standalone tasks
- Following a checklist or procedure

## Planning Process

### Step 1: Assess Complexity
Before creating a plan, consider:
- How many distinct activities are involved?
- Are there natural groupings or phases?
- What's the expected duration?
- Does the user need to see granular progress?

### Step 2: Choose Structure
Based on assessment, select:
- **Simple** (3-5 tasks): Just tasks, no hierarchy
- **Medium** (5-10 tasks): Tasks, possibly with todos for complex ones
- **Complex** (10+ tasks or multi-stage): Phases with tasks

### Step 3: Write Good Task Descriptions
Each task should have:
- **content**: Clear imperative statement ("Create user model", "Run test suite")
- **active_form**: Present continuous ("Creating user model", "Running test suite")

Good tasks are:
- Specific and actionable
- Verifiable (you know when it's done)
- Appropriately sized (not too big, not too small)

### Step 4: Assign Initial Statuses
- First task/todo to work on: `in_progress`
- Everything else: `pending`
- Never start with anything `completed`

## Examples

### Simple Plan (Just Tasks)
```json
{
  "tasks": [
    {"content": "Analyze requirements", "active_form": "Analyzing requirements", "status": "in_progress"},
    {"content": "Create implementation", "active_form": "Creating implementation", "status": "pending"},
    {"content": "Write tests", "active_form": "Writing tests", "status": "pending"},
    {"content": "Update documentation", "active_form": "Updating documentation", "status": "pending"}
  ]
}
```

### Medium Plan (Tasks with Todos)
```json
{
  "tasks": [
    {
      "content": "Implement authentication",
      "active_form": "Implementing authentication",
      "status": "in_progress",
      "todos": [
        {"content": "Create user model", "active_form": "Creating user model", "status": "in_progress"},
        {"content": "Add password hashing", "active_form": "Adding password hashing", "status": "pending"},
        {"content": "Create login endpoint", "active_form": "Creating login endpoint", "status": "pending"},
        {"content": "Add session management", "active_form": "Adding session management", "status": "pending"}
      ]
    },
    {
      "content": "Write integration tests",
      "active_form": "Writing integration tests",
      "status": "pending"
    }
  ]
}
```

### Complex Plan (Phases)
```json
{
  "phases": [
    {
      "name": "Phase 1: Research & Design",
      "status": "completed",
      "tasks": [
        {"content": "Analyze existing codebase", "active_form": "Analyzing codebase", "status": "completed"},
        {"content": "Design solution architecture", "active_form": "Designing architecture", "status": "completed"}
      ]
    },
    {
      "name": "Phase 2: Implementation",
      "status": "in_progress",
      "tasks": [
        {"content": "Implement core features", "active_form": "Implementing features", "status": "in_progress"},
        {"content": "Add error handling", "active_form": "Adding error handling", "status": "pending"}
      ]
    },
    {
      "name": "Phase 3: Testing & Polish",
      "status": "pending",
      "tasks": [
        {"content": "Write comprehensive tests", "active_form": "Writing tests", "status": "pending"},
        {"content": "Performance optimization", "active_form": "Optimizing performance", "status": "pending"}
      ]
    }
  ]
}
```

## Progress Management

### Status Transitions
- `pending` → `in_progress`: When starting work
- `in_progress` → `completed`: When work is verified complete
- Never skip directly from `pending` to `completed`

### Rules
1. **One in_progress at a time**: Across all levels, only ONE item should be `in_progress`
2. **Immediate updates**: Update status as soon as work completes, don't batch
3. **Remove irrelevant items**: If a task becomes unnecessary, remove it entirely
4. **Add as needed**: New tasks can be added when discovered during work

### When a Task with Todos Completes
When all todos in a task are completed, mark the parent task as completed too.

### When All Tasks in a Phase Complete
When all tasks in a phase are completed, mark the phase as completed and start the next phase.

## Anti-Patterns to Avoid

1. **Over-planning**: Don't create 50 micro-tasks for simple work
2. **Under-planning**: Don't create one giant task for complex work
3. **Vague tasks**: "Fix stuff" is not a good task
4. **Missing active_form**: Always provide the present continuous form
5. **Single phase**: If using phases, you need at least 2
6. **Empty phases**: Every phase must have at least 1 task
7. **Mixed levels**: Don't use both top-level tasks AND phases

## Integration with Work

1. **Before starting work**: Create or update the plan
2. **When starting a task**: Mark it `in_progress`
3. **When completing work**: Mark it `completed`, start next item
4. **When discovering new work**: Add tasks/todos as needed
5. **When blocked**: Keep as `in_progress`, add clarifying task for the blocker
