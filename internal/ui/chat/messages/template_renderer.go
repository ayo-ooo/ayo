package messages

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/charmbracelet/lipgloss"
)

// templateRenderer renders tool calls using a custom template.
type templateRenderer struct {
	baseRenderer
	name          string
	inlineTempl   *template.Template
	panelTempl    *template.Template
	config        templateConfig
}

// templateConfig holds renderer configuration from render.json.
type templateConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Styles      map[string]string `json:"styles"`
	MaxLines    int               `json:"max_lines"`
}

// PluginRenderDir is the subdirectory where plugin renderers are stored.
const PluginRenderDir = "render"

// LoadPluginRenderers scans plugin tools for custom renderers.
func LoadPluginRenderers(pluginsDir string) error {
	if pluginsDir == "" {
		return nil
	}

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginDir := filepath.Join(pluginsDir, entry.Name())
		toolsDir := filepath.Join(pluginDir, "tools")

		toolEntries, err := os.ReadDir(toolsDir)
		if err != nil {
			continue
		}

		for _, toolEntry := range toolEntries {
			if !toolEntry.IsDir() {
				continue
			}

			toolName := toolEntry.Name()
			renderDir := filepath.Join(toolsDir, toolName, PluginRenderDir)

			if err := loadToolRenderer(toolName, renderDir); err != nil {
				// Log but continue - don't fail on bad renderer
				continue
			}
		}
	}

	return nil
}

// loadToolRenderer loads a renderer from a render/ directory.
func loadToolRenderer(toolName, renderDir string) error {
	info, err := os.Stat(renderDir)
	if err != nil || !info.IsDir() {
		return nil // No render dir, skip
	}

	r := &templateRenderer{
		name: toolName,
		config: templateConfig{
			MaxLines: 10,
		},
	}

	// Load config if present
	configPath := filepath.Join(renderDir, "render.json")
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &r.config)
	}

	// Load inline template if present
	inlinePath := filepath.Join(renderDir, "inline.tmpl")
	if data, err := os.ReadFile(inlinePath); err == nil {
		tmpl, err := template.New("inline").Funcs(templateFuncs()).Parse(string(data))
		if err == nil {
			r.inlineTempl = tmpl
		}
	}

	// Load panel template if present
	panelPath := filepath.Join(renderDir, "panel.tmpl")
	if data, err := os.ReadFile(panelPath); err == nil {
		tmpl, err := template.New("panel").Funcs(templateFuncs()).Parse(string(data))
		if err == nil {
			r.panelTempl = tmpl
		}
	}

	// Only register if we have at least one template
	if r.inlineTempl != nil || r.panelTempl != nil {
		registry.register(toolName, func() renderer { return r })
	}

	return nil
}

// templateFuncs returns the template function map.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"style":      styleFunc,
		"truncate":   truncateFunc,
		"icon":       iconFunc,
		"join":       strings.Join,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"title":      strings.Title,
		"trim":       strings.TrimSpace,
		"jsonPretty": jsonPrettyFunc,
	}
}

// styleFunc applies lipgloss styling.
func styleFunc(color, text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
}

// truncateFunc truncates text to max length.
func truncateFunc(maxLen int, text string) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return "..."
	}
	return text[:maxLen-3] + "..."
}

// iconFunc returns status icons.
func iconFunc(status string) string {
	switch status {
	case "success":
		return ToolSuccess
	case "error":
		return ToolError
	case "pending":
		return ToolPending
	case "running":
		return ToolRunning
	default:
		return ToolPending
	}
}

// jsonPrettyFunc pretty-prints JSON.
func jsonPrettyFunc(s string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return s
	}
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return s
	}
	return string(pretty)
}

// templateData holds data passed to templates.
type templateData struct {
	Name       string
	Input      string
	Params     map[string]interface{}
	Result     string
	IsError    bool
	Duration   string
	IsNested   bool
	IsPending  bool
	IsRunning  bool
	Width      int
	MaxLines   int
}

// Render renders the tool call using the template.
func (r *templateRenderer) Render(t *toolCallCmp) string {
	if r.inlineTempl == nil {
		// Fall back to generic renderer
		return genericRenderer{}.Render(t)
	}

	// Parse input JSON
	var params map[string]interface{}
	json.Unmarshal([]byte(t.call.Input), &params)

	data := templateData{
		Name:      r.name,
		Input:     t.call.Input,
		Params:    params,
		Result:    t.result.Content,
		IsError:   t.result.IsError,
		Duration:  "", // Duration not tracked at component level
		IsNested:  t.isNested,
		IsPending: t.result.ToolCallID == "" && !t.spinning,
		IsRunning: t.spinning,
		Width:     t.textWidth(),
		MaxLines:  r.config.MaxLines,
	}

	var buf bytes.Buffer
	if err := r.inlineTempl.Execute(&buf, data); err != nil {
		return genericRenderer{}.Render(t)
	}

	return buf.String()
}

// RenderPanel renders the tool call for a sidebar panel.
func (r *templateRenderer) RenderPanel(t *toolCallCmp) string {
	if r.panelTempl == nil {
		return ""
	}

	var params map[string]interface{}
	json.Unmarshal([]byte(t.call.Input), &params)

	data := templateData{
		Name:      r.name,
		Input:     t.call.Input,
		Params:    params,
		Result:    t.result.Content,
		IsError:   t.result.IsError,
		Duration:  "", // Duration not tracked at component level
		IsNested:  t.isNested,
		IsPending: t.result.ToolCallID == "" && !t.spinning,
		IsRunning: t.spinning,
		Width:     t.textWidth(),
		MaxLines:  r.config.MaxLines,
	}

	var buf bytes.Buffer
	if err := r.panelTempl.Execute(&buf, data); err != nil {
		return ""
	}

	return buf.String()
}
