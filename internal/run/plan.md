Creates and manages a structured task plan for tracking progress on complex, multi-step tasks.

<structure>
Plans support a hierarchical structure with three levels:

1. **Phases** (optional): High-level groupings for major work stages
   - If used, must have at least 2 phases
   - Each phase must contain at least 1 task
   
2. **Tasks** (required): Units of work that can be completed
   - Can exist at the top level (no phases) or within phases
   - Each task requires `content` and `active_form`
   
3. **Todos** (optional): Atomic sub-items within tasks
   - Use for granular tracking within a task
   - Each todo requires `content` and `active_form`

**Valid structures:**
- Just tasks: `{"tasks": [...]}`
- Phases with tasks: `{"phases": [{"name": "...", "tasks": [...]}]}`
- Tasks with todos: `{"tasks": [{"content": "...", "todos": [...]}]}`
- Full hierarchy: phases containing tasks containing todos
</structure>

<when_to_use>
Use this tool proactively in these scenarios:

- Complex multi-step tasks requiring 3+ distinct steps or actions
- Non-trivial tasks requiring careful planning or multiple operations
- User explicitly requests task planning or tracking
- User provides multiple tasks (numbered or comma-separated list)
- After receiving new instructions to capture requirements
- When starting work on a task (mark as in_progress BEFORE beginning)
- After completing a task (mark completed and add new follow-up tasks)
</when_to_use>

<when_not_to_use>
Skip this tool when:

- Single, straightforward task
- Trivial task with no organizational benefit
- Task completable in less than 3 trivial steps
- Purely conversational or informational request
</when_not_to_use>

<task_states>
- **pending**: Not yet started
- **in_progress**: Currently working on (limit to ONE item at a time across all levels)
- **completed**: Finished successfully

**IMPORTANT**: Each task and todo requires two forms:
- **content**: Imperative form describing what needs to be done (e.g., "Run tests", "Build the project")
- **active_form**: Present continuous form shown during execution (e.g., "Running tests", "Building the project")
</task_states>

<planning_guidelines>
**When to use phases:**
- Large projects with distinct stages (e.g., "Setup", "Implementation", "Testing")
- Work that spans multiple sessions or days
- Projects where progress reporting matters

**When to use just tasks:**
- Medium complexity work that can be listed linearly
- Single-session work with clear steps
- Most common case

**When to use todos within tasks:**
- A task has multiple atomic steps that should be tracked
- You want granular progress visibility
- Steps are small enough they don't warrant separate tasks
</planning_guidelines>

<task_management>
- Update status in real-time as you work
- Mark items complete IMMEDIATELY after finishing (don't batch completions)
- Exactly ONE item should be in_progress at any time (at any level)
- Complete current items before starting new ones
- Remove items that are no longer relevant from the list entirely
- When a task's todos are all completed, mark the task as completed too
</task_management>

<completion_requirements>
ONLY mark an item as completed when you have FULLY accomplished it.

Never mark completed if:
- Tests are failing
- Implementation is partial
- You encountered unresolved errors
- You couldn't find necessary files or dependencies

If blocked:
- Keep item as in_progress
- Add new task/todo describing what needs to be resolved
</completion_requirements>

<examples>
**Simple tasks (flat list):**
```json
{
  "tasks": [
    {"content": "Run tests", "active_form": "Running tests", "status": "completed"},
    {"content": "Fix failing test", "active_form": "Fixing failing test", "status": "in_progress"},
    {"content": "Update documentation", "active_form": "Updating documentation", "status": "pending"}
  ]
}
```

**Tasks with todos:**
```json
{
  "tasks": [
    {
      "content": "Implement user authentication",
      "active_form": "Implementing user authentication",
      "status": "in_progress",
      "todos": [
        {"content": "Create user model", "active_form": "Creating user model", "status": "completed"},
        {"content": "Add password hashing", "active_form": "Adding password hashing", "status": "in_progress"},
        {"content": "Create login endpoint", "active_form": "Creating login endpoint", "status": "pending"}
      ]
    }
  ]
}
```

**Full hierarchy with phases:**
```json
{
  "phases": [
    {
      "name": "Phase 1: Setup",
      "status": "completed",
      "tasks": [
        {"content": "Initialize project", "active_form": "Initializing project", "status": "completed"},
        {"content": "Configure dependencies", "active_form": "Configuring dependencies", "status": "completed"}
      ]
    },
    {
      "name": "Phase 2: Implementation", 
      "status": "in_progress",
      "tasks": [
        {
          "content": "Build core features",
          "active_form": "Building core features",
          "status": "in_progress",
          "todos": [
            {"content": "Create data models", "active_form": "Creating data models", "status": "completed"},
            {"content": "Implement API endpoints", "active_form": "Implementing API endpoints", "status": "in_progress"}
          ]
        }
      ]
    }
  ]
}
```
</examples>

<output_behavior>
**NEVER** print or list the plan in your response text. The user sees the plan in real-time in the UI.
</output_behavior>

<tips>
- When in doubt, use this tool - being proactive demonstrates attentiveness
- One item in_progress at a time keeps work focused
- Update immediately after state changes for accurate tracking
- Start simple (just tasks) and add hierarchy only when needed
- Phase names should describe the stage of work, not just "Phase 1"
</tips>
