package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTemplate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "simple template",
			input:   "Hello, {{.name}}!",
			wantErr: false,
		},
		{
			name:    "template with function",
			input:   "Upper: {{upper .name}}",
			wantErr: false,
		},
		{
			name:    "template with json function",
			input:   "Data: {{json .}}",
			wantErr: false,
		},
		{
			name:    "invalid template syntax",
			input:   "Hello, {{.name",
			wantErr: true,
		},
		{
			name:    "undefined function",
			input:   "Value: {{undefined .name}}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tmpl == nil {
				t.Error("ParseTemplate() returned nil template without error")
			}
		})
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    string
		data    map[string]any
		want    string
		wantErr bool
	}{
		{
			name: "simple substitution",
			tmpl: "Hello, {{.name}}!",
			data: map[string]any{"name": "World"},
			want: "Hello, World!",
		},
		{
			name: "multiple fields",
			tmpl: "{{.greeting}}, {{.name}}!",
			data: map[string]any{"greeting": "Hello", "name": "Alice"},
			want: "Hello, Alice!",
		},
		{
			name: "nested data",
			tmpl: "User: {{.user.name}}, Age: {{.user.age}}",
			data: map[string]any{
				"user": map[string]any{
					"name": "Bob",
					"age":  30,
				},
			},
			want: "User: Bob, Age: 30",
		},
		{
			name: "missing field",
			tmpl: "Value: {{.missing}}",
			data: map[string]any{"other": "value"},
			want: "Value: <no value>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tt.tmpl)
			if err != nil {
				t.Fatalf("ParseTemplate() error = %v", err)
			}

			got, err := Render(tmpl, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderString(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    string
		data    map[string]any
		want    string
		wantErr bool
	}{
		{
			name: "simple render",
			tmpl: "Hello, {{.name}}!",
			data: map[string]any{"name": "World"},
			want: "Hello, World!",
		},
		{
			name:    "invalid template",
			tmpl:    "{{.invalid",
			data:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderString(tt.tmpl, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("RenderString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpperFunc(t *testing.T) {
	tmpl, err := ParseTemplate("{{upper .value}}")
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	got, err := Render(tmpl, map[string]any{"value": "hello world"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := "HELLO WORLD"
	if got != want {
		t.Errorf("upper function = %q, want %q", got, want)
	}
}

func TestLowerFunc(t *testing.T) {
	tmpl, err := ParseTemplate("{{lower .value}}")
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	got, err := Render(tmpl, map[string]any{"value": "HELLO WORLD"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := "hello world"
	if got != want {
		t.Errorf("lower function = %q, want %q", got, want)
	}
}

func TestTitleFunc(t *testing.T) {
	tmpl, err := ParseTemplate("{{title .value}}")
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	got, err := Render(tmpl, map[string]any{"value": "hello world"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := "Hello World"
	if got != want {
		t.Errorf("title function = %q, want %q", got, want)
	}
}

func TestTrimFunc(t *testing.T) {
	tmpl, err := ParseTemplate("{{trim .value}}")
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	got, err := Render(tmpl, map[string]any{"value": "  hello world  "})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := "hello world"
	if got != want {
		t.Errorf("trim function = %q, want %q", got, want)
	}
}

func TestEnvFunc(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	tmpl, err := ParseTemplate("{{env \"TEST_VAR\"}}")
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	got, err := Render(tmpl, map[string]any{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := "test_value"
	if got != want {
		t.Errorf("env function = %q, want %q", got, want)
	}
}

func TestEnvFuncMissing(t *testing.T) {
	tmpl, err := ParseTemplate("{{env \"NONEXISTENT_VAR\"}}")
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	got, err := Render(tmpl, map[string]any{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := ""
	if got != want {
		t.Errorf("env function for missing var = %q, want %q", got, want)
	}
}

func TestFileFunc(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := "file content here"

	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tmpl, err := ParseTemplate("{{file .path}}")
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	got, err := Render(tmpl, map[string]any{"path": tmpFile})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if got != content {
		t.Errorf("file function = %q, want %q", got, content)
	}
}

func TestFileFuncMissing(t *testing.T) {
	tmpl, err := ParseTemplate("{{file \"/nonexistent/path/file.txt\"}}")
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	_, err = Render(tmpl, map[string]any{})
	if err == nil {
		t.Error("file function should return error for missing file")
	}
}

func TestJsonFunc(t *testing.T) {
	tests := []struct {
		name         string
		data         map[string]any
		tmpl         string
		wantContains []string
	}{
		{
			name:         "simple object",
			data:         map[string]any{"name": "Alice", "age": 30},
			tmpl:         "{{json .}}",
			wantContains: []string{`"name": "Alice"`, `"age": 30`},
		},
		{
			name:         "nested object",
			data:         map[string]any{"user": map[string]any{"name": "Bob"}},
			tmpl:         "{{json .user}}",
			wantContains: []string{`"name": "Bob"`},
		},
		{
			name:         "array",
			data:         map[string]any{"items": []any{"a", "b", "c"}},
			tmpl:         "{{json .items}}",
			wantContains: []string{`"a"`, `"b"`, `"c"`},
		},
		{
			name:         "string value",
			data:         map[string]any{"value": "test"},
			tmpl:         "{{json .value}}",
			wantContains: []string{`"test"`},
		},
		{
			name:         "number value",
			data:         map[string]any{"value": 42},
			tmpl:         "{{json .value}}",
			wantContains: []string{`42`},
		},
		{
			name:         "boolean value",
			data:         map[string]any{"value": true},
			tmpl:         "{{json .value}}",
			wantContains: []string{`true`},
		},
		{
			name:         "null value",
			data:         map[string]any{"value": nil},
			tmpl:         "{{json .value}}",
			wantContains: []string{`null`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tt.tmpl)
			if err != nil {
				t.Fatalf("ParseTemplate() error = %v", err)
			}

			got, err := Render(tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("json function output %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestCombinedFunctions(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "upper and trim",
			tmpl: "{{upper (trim .value)}}",
			data: map[string]any{"value": "  hello  "},
			want: "HELLO",
		},
		{
			name: "lower and title",
			tmpl: "{{title (lower .value)}}",
			data: map[string]any{"value": "HELLO WORLD"},
			want: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tt.tmpl)
			if err != nil {
				t.Fatalf("ParseTemplate() error = %v", err)
			}

			got, err := Render(tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("combined functions = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConditionalRendering(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "if true",
			tmpl: "{{if .enabled}}yes{{else}}no{{end}}",
			data: map[string]any{"enabled": true},
			want: "yes",
		},
		{
			name: "if false",
			tmpl: "{{if .enabled}}yes{{else}}no{{end}}",
			data: map[string]any{"enabled": false},
			want: "no",
		},
		{
			name: "if string non-empty",
			tmpl: "{{if .name}}has name{{else}}no name{{end}}",
			data: map[string]any{"name": "Alice"},
			want: "has name",
		},
		{
			name: "if string empty",
			tmpl: "{{if .name}}has name{{else}}no name{{end}}",
			data: map[string]any{"name": ""},
			want: "no name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderString(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("RenderString() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("conditional rendering = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRangeRendering(t *testing.T) {
	tmpl := "{{range .items}}{{.}},{{end}}"
	data := map[string]any{
		"items": []any{"a", "b", "c"},
	}

	got, err := RenderString(tmpl, data)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	want := "a,b,c,"
	if got != want {
		t.Errorf("range rendering = %q, want %q", got, want)
	}
}
