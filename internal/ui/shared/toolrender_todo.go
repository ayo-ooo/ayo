package shared

import (
	"fmt"

	"github.com/alexcabrera/ayo/internal/util"
)

// TodoRenderer renders todo/todos tool calls.
type TodoRenderer struct{}

// Name returns "todo".
func (t *TodoRenderer) Name() string { return "todo" }

// Render produces render output for todo tool calls.
func (t *TodoRenderer) Render(input ToolRenderInput) ToolRenderOutput {
	out := ToolRenderOutput{
		Label:    "Todo",
		State:    input.State,
		IsNested: input.ParentID != "",
	}

	// Parse parameters
	var params TodosParams
	if err := ParseJSON(input.RawInput, &params); err == nil {
		input.Params = params
	} else {
		params = TodosParams{}
	}

	// Parse metadata
	var meta TodosResponseMetadata
	if err := ParseJSON(input.RawMetadata, &meta); err == nil {
		input.Metadata = meta
	}

	// Build header text: "X/Y" or "X/Y - current task"
	todos := params.Todos
	if len(meta.Todos) > 0 {
		todos = meta.Todos
	}

	completed := 0
	inProgressTask := ""
	for _, todo := range todos {
		if todo.Status == "completed" {
			completed++
		}
		if todo.Status == "in_progress" {
			if todo.ActiveForm != "" {
				inProgressTask = todo.ActiveForm
			} else {
				inProgressTask = todo.Content
			}
		}
	}

	// Use metadata counts if available
	if meta.Total > 0 {
		completed = meta.Completed
	}
	total := len(todos)
	if meta.Total > 0 {
		total = meta.Total
	}

	headerText := fmt.Sprintf("%d/%d", completed, total)
	if meta.JustStarted != "" {
		inProgressTask = meta.JustStarted
	}
	if inProgressTask != "" {
		headerText = fmt.Sprintf("%d/%d - %s", completed, total, util.Truncate(inProgressTask, 40))
	}

	out.HeaderParams = []string{headerText}

	// Add todo list as body section
	if len(todos) > 0 {
		out.Sections = append(out.Sections, RenderSection{
			Type:    SectionTodos,
			Content: "", // Content is not used; todos are in input.Params
		})
	}

	return out
}

// TodosRenderer is an alias for TodoRenderer (handles both "todo" and "todos").
type TodosRenderer struct {
	TodoRenderer
}

// Name returns "todos".
func (t *TodosRenderer) Name() string { return "todos" }

func init() {
	RegisterToolRenderer(&TodoRenderer{})
	RegisterToolRenderer(&TodosRenderer{})
}
