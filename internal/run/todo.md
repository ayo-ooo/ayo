Creates and manages a structured todo list for tracking progress on complex, multi-step tasks.

<when_to_use>
Use this tool proactively in these scenarios:

- Complex multi-step tasks requiring 3+ distinct steps or actions
- Non-trivial tasks requiring careful planning or multiple operations
- User explicitly requests todo list management
- User provides multiple tasks (numbered or comma-separated list)
- After receiving new instructions to capture requirements
- When starting work on a todo (mark as in_progress BEFORE beginning)
- After completing a todo (mark completed and add new follow-up todos)
</when_to_use>

<when_not_to_use>
Skip this tool when:

- Single, straightforward task
- Trivial task with no organizational benefit
- Task completable in less than 3 trivial steps
- Purely conversational or informational request
</when_not_to_use>

<todo_states>
- **pending**: Todo not yet started
- **in_progress**: Currently working on (limit to ONE todo at a time)
- **completed**: Todo finished successfully

**IMPORTANT**: Each todo requires two forms:
- **content**: Imperative form describing what needs to be done (e.g., "Run tests", "Build the project")
- **active_form**: Present continuous form shown during execution (e.g., "Running tests", "Building the project")
</todo_states>

<todo_management>
- Update todo status in real-time as you work
- Mark todos complete IMMEDIATELY after finishing (don't batch completions)
- Exactly ONE todo must be in_progress at any time (not less, not more)
- Complete current todos before starting new ones
- Remove todos that are no longer relevant from the list entirely
</todo_management>

<completion_requirements>
ONLY mark a todo as completed when you have FULLY accomplished it.

Never mark completed if:
- Tests are failing
- Implementation is partial
- You encountered unresolved errors
- You couldn't find necessary files or dependencies

If blocked:
- Keep todo as in_progress
- Create new todo describing what needs to be resolved
</completion_requirements>

<todo_breakdown>
- Create specific, actionable items
- Break complex tasks into smaller, manageable steps
- Use clear, descriptive todo names
- Always provide both content and active_form
</todo_breakdown>

<examples>
Good todo:
```json
{
  "content": "Implement user authentication with JWT tokens",
  "status": "in_progress",
  "active_form": "Implementing user authentication with JWT tokens"
}
```

Bad todo (missing active_form):
```json
{
  "content": "Fix bug",
  "status": "pending"
}
```
</examples>

<output_behavior>
**NEVER** print or list todos in your response text. The user sees the todo list in real-time in the UI.
</output_behavior>

<tips>
- When in doubt, use this tool - being proactive demonstrates attentiveness
- One todo in_progress at a time keeps work focused
- Update immediately after state changes for accurate tracking
</tips>
