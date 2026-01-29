package messages

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"

	"github.com/alexcabrera/ayo/internal/ui/shared"
)

// genericRenderer handles unknown tool types with basic parameter display.
type genericRenderer struct {
	baseRenderer
}

// Render displays generic tool output.
func (gr genericRenderer) Render(t *toolCallCmp) string {
	return gr.renderWithParams(t, prettifyToolName(t.call.Name), []string{}, func() string {
		if t.result.Content == "" {
			return ""
		}
		return renderPlainContent(t, t.result.Content, 10)
	})
}

// bashRenderer handles bash command execution display.
type bashRenderer struct {
	baseRenderer
}

// BashParams is an alias for shared.BashParams.
type BashParams = shared.BashParams

// BashResponseMetadata is an alias for shared.BashResponseMetadata.
type BashResponseMetadata = shared.BashResponseMetadata

// Render displays bash command with output.
func (br bashRenderer) Render(t *toolCallCmp) string {
	// Use shared renderer for data extraction
	renderInput := t.ToRenderInput()
	renderOutput := shared.RenderTool(renderInput)

	return br.renderWithParams(t, renderOutput.Label, renderOutput.HeaderParams, func() string {
		// Render body sections
		for _, section := range renderOutput.Sections {
			switch section.Type {
			case shared.SectionJSON:
				return renderPlainContent(t, section.Content, section.MaxLines)
			case shared.SectionPlain:
				return renderPlainContent(t, section.Content, section.MaxLines)
			}
		}
		return ""
	})
}

// todosRenderer handles todo list display.
type todosRenderer struct {
	baseRenderer
}

// Todo is an alias for shared.Todo.
type Todo = shared.Todo

// TodosParams is an alias for shared.TodosParams.
type TodosParams = shared.TodosParams

// TodosResponseMetadata is an alias for shared.TodosResponseMetadata.
type TodosResponseMetadata = shared.TodosResponseMetadata

// Render displays todo list.
func (tr todosRenderer) Render(t *toolCallCmp) string {
	// Use shared renderer for data extraction
	renderInput := t.ToRenderInput()
	renderOutput := shared.RenderTool(renderInput)

	// Parse todos for body rendering
	var params TodosParams
	var meta TodosResponseMetadata
	_ = tr.unmarshalParams(t.call.Input, &params)
	if t.result.Metadata != "" {
		_ = tr.unmarshalParams(t.result.Metadata, &meta)
	}

	return tr.renderWithParams(t, renderOutput.Label, renderOutput.HeaderParams, func() string {
		todos := params.Todos
		if len(meta.Todos) > 0 {
			todos = meta.Todos
		}

		if len(todos) == 0 {
			return ""
		}

		return formatTodosList(todos, t.textWidth()-4)
	})
}

// formatTodosList formats a list of todos for display using shared styling.
func formatTodosList(todos []Todo, width int) string {
	return shared.FormatTodos(todos, width)
}

// agentRenderer handles sub-agent call display.
type agentRenderer struct {
	baseRenderer
}

// AgentParams represents agent tool parameters.
type AgentParams struct {
	Prompt string `json:"prompt"`
}

// Render displays agent call with nested tool calls using lipgloss/tree.
func (ar agentRenderer) Render(t *toolCallCmp) string {
	var params AgentParams
	_ = ar.unmarshalParams(t.call.Input, &params)

	width := t.textWidth()

	// Build header
	header := ar.makeHeader(t, "Agent", width)

	// Format prompt as task description
	prompt := params.Prompt
	prompt = strings.ReplaceAll(prompt, "\n", " ")
	if len(prompt) > 60 {
		prompt = prompt[:57] + "..."
	}

	taskStyle := lipgloss.NewStyle().Foreground(shared.ColorTextDim)
	taskLine := fmt.Sprintf("  Task: %s", taskStyle.Render(prompt))

	// Combine header and task line
	headerWithTask := lipgloss.JoinVertical(lipgloss.Left, header, "", taskLine)

	// Build tree with nested tool calls
	rootTree := tree.Root(headerWithTask)

	// Add nested tool calls as children if expanded
	if t.expanded && len(t.nestedToolCalls) > 0 {
		for _, nested := range t.nestedToolCalls {
			childView := nested.View()
			rootTree.Child(childView)
		}
	}

	// Apply enumerator and render tree
	var result string
	if len(t.nestedToolCalls) > 0 {
		result = rootTree.Enumerator(roundedEnumerator(2, 3)).String()
	} else {
		result = headerWithTask
	}

	// Add collapse indicator if there are nested calls
	if len(t.nestedToolCalls) > 0 && !t.expanded {
		collapseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
		result = lipgloss.JoinVertical(lipgloss.Left, result, "",
			collapseStyle.Render(fmt.Sprintf("  [%d nested tool calls collapsed]", len(t.nestedToolCalls))))
	}

	// Add result content when completed
	if t.result.ToolCallID != "" && t.result.Content != "" {
		body := renderMarkdownContent(t, t.result.Content, 10)
		result = joinHeaderBody(result, body)
	}

	return result
}

// init registers all built-in renderers.
func init() {
	registry.register("bash", func() renderer { return bashRenderer{} })
	registry.register("todo", func() renderer { return todosRenderer{} })
	registry.register("todos", func() renderer { return todosRenderer{} })
	registry.register("agent", func() renderer { return agentRenderer{} })
}

// RegisterRenderer allows external packages to register custom renderers.
func RegisterRenderer(name string, factory func() renderer) {
	registry.register(name, func() renderer { return factory() })
}

// LookupRenderer retrieves a renderer by name.
func LookupRenderer(name string) renderer {
	return registry.lookup(name)
}

// GetRegisteredRenderers returns all registered renderer names.
func GetRegisteredRenderers() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// IsRendererRegistered checks if a renderer is registered for the given name.
func IsRendererRegistered(name string) bool {
	_, ok := registry[name]
	return ok
}

// RenderToolCallJSON provides a fallback for rendering unknown JSON.
func RenderToolCallJSON(t *toolCallCmp, data string) string {
	var parsed interface{}
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		return renderPlainContent(t, data, 10)
	}

	pretty, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return renderPlainContent(t, data, 10)
	}

	return renderPlainContent(t, string(pretty), 10)
}
