package shared

import (
	"testing"
)

func TestPrettifyToolName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "Tool"},
		{"bash", "Bash"},
		{"todo", "Todo"},
		{"agent_call", "Agent Call"},
		{"my_tool_name", "My Tool Name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := PrettifyToolName(tt.input)
			if result != tt.expected {
				t.Errorf("PrettifyToolName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetToolRenderer_Registered(t *testing.T) {
	// Bash renderer should be registered
	r := GetToolRenderer("bash")
	if r.Name() != "bash" {
		t.Errorf("GetToolRenderer(\"bash\").Name() = %q, want \"bash\"", r.Name())
	}

	// Todo renderer should be registered
	r = GetToolRenderer("todo")
	if r.Name() != "todo" {
		t.Errorf("GetToolRenderer(\"todo\").Name() = %q, want \"todo\"", r.Name())
	}

	// Todos renderer should be registered
	r = GetToolRenderer("todos")
	if r.Name() != "todos" {
		t.Errorf("GetToolRenderer(\"todos\").Name() = %q, want \"todos\"", r.Name())
	}
}

func TestGetToolRenderer_Fallback(t *testing.T) {
	// Unknown tool should return generic renderer
	r := GetToolRenderer("unknown_tool")
	if r.Name() != "" {
		t.Errorf("GetToolRenderer(\"unknown_tool\").Name() = %q, want \"\"", r.Name())
	}
}

func TestBashRenderer_Render(t *testing.T) {
	input := ToolRenderInput{
		Name:     "bash",
		RawInput: `{"command":"ls -la","description":"listing files"}`,
		State:    ToolStateSuccess,
	}

	output := RenderTool(input)

	if output.Label != "Bash" {
		t.Errorf("output.Label = %q, want \"Bash\"", output.Label)
	}

	if output.State != ToolStateSuccess {
		t.Errorf("output.State = %v, want %v", output.State, ToolStateSuccess)
	}

	if len(output.HeaderParams) == 0 {
		t.Error("expected HeaderParams to contain command")
	} else if output.HeaderParams[0] != "ls -la" {
		t.Errorf("HeaderParams[0] = %q, want \"ls -la\"", output.HeaderParams[0])
	}
}

func TestBashRenderer_WithBackground(t *testing.T) {
	input := ToolRenderInput{
		Name:     "bash",
		RawInput: `{"command":"npm start","description":"starting server","run_in_background":true}`,
		State:    ToolStateRunning,
	}

	output := RenderTool(input)

	// Should have background flag in header params
	if len(output.HeaderParams) < 3 {
		t.Errorf("expected HeaderParams to include background flag, got %v", output.HeaderParams)
	}
}

func TestTodoRenderer_Render(t *testing.T) {
	input := ToolRenderInput{
		Name: "todo",
		RawInput: `{"todos":[
			{"content":"Task 1","status":"completed"},
			{"content":"Task 2","status":"in_progress","active_form":"Working on task 2"},
			{"content":"Task 3","status":"pending"}
		]}`,
		State: ToolStateSuccess,
	}

	output := RenderTool(input)

	if output.Label != "Todo" {
		t.Errorf("output.Label = %q, want \"Todo\"", output.Label)
	}

	// Header should show progress like "1/3"
	if len(output.HeaderParams) == 0 {
		t.Error("expected HeaderParams to contain progress")
	}
}

func TestGenericRenderer_Render(t *testing.T) {
	input := ToolRenderInput{
		Name:      "custom_tool",
		RawInput:  `{"foo":"bar"}`,
		RawOutput: "Some output text",
		State:     ToolStateSuccess,
	}

	output := RenderTool(input)

	if output.Label != "Custom Tool" {
		t.Errorf("output.Label = %q, want \"Custom Tool\"", output.Label)
	}

	if len(output.Sections) == 0 {
		t.Error("expected Sections to contain output")
	} else if output.Sections[0].Content != "Some output text" {
		t.Errorf("Sections[0].Content = %q, want \"Some output text\"", output.Sections[0].Content)
	}
}

func TestToolState(t *testing.T) {
	// Verify state constants are distinct
	states := []ToolState{
		ToolStatePending,
		ToolStateRunning,
		ToolStateSuccess,
		ToolStateError,
		ToolStateCancelled,
	}

	seen := make(map[ToolState]bool)
	for _, s := range states {
		if seen[s] {
			t.Errorf("duplicate ToolState value: %v", s)
		}
		seen[s] = true
	}
}
