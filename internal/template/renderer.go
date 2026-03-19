package template

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

var funcMap = template.FuncMap{
	"json": jsonFunc,
	"file": fileFunc,
	"env":  envFunc,
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	"title": strings.Title,
	"trim":  strings.TrimSpace,
}

func jsonFunc(v any) (string, error) {
	data, err := marshalJSON(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func fileFunc(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", path, err)
	}
	return string(data), nil
}

func envFunc(key string) string {
	return os.Getenv(key)
}

func marshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := jsonEncoder{buf: &buf}
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type jsonEncoder struct {
	buf *bytes.Buffer
}

func (e *jsonEncoder) Encode(v any) error {
	return e.encodeValue(v, 0)
}

func (e *jsonEncoder) encodeValue(v any, indent int) error {
	switch val := v.(type) {
	case string:
		e.buf.WriteString(fmt.Sprintf("%q", val))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		e.buf.WriteString(fmt.Sprintf("%v", val))
	case nil:
		e.buf.WriteString("null")
	case map[string]any:
		e.encodeMap(val, indent)
	case []any:
		e.encodeSlice(val, indent)
	default:
		e.buf.WriteString(fmt.Sprintf("%q", fmt.Sprintf("%v", val)))
	}
	return nil
}

func (e *jsonEncoder) encodeMap(m map[string]any, indent int) {
	e.buf.WriteString("{\n")
	first := true
	for k, v := range m {
		if !first {
			e.buf.WriteString(",\n")
		}
		first = false
		e.buf.WriteString(strings.Repeat("  ", indent+1))
		e.buf.WriteString(fmt.Sprintf("%q: ", k))
		e.encodeValue(v, indent+1)
	}
	e.buf.WriteString("\n")
	e.buf.WriteString(strings.Repeat("  ", indent))
	e.buf.WriteString("}")
}

func (e *jsonEncoder) encodeSlice(s []any, indent int) {
	if len(s) == 0 {
		e.buf.WriteString("[]")
		return
	}
	e.buf.WriteString("[\n")
	for i, v := range s {
		if i > 0 {
			e.buf.WriteString(",\n")
		}
		e.buf.WriteString(strings.Repeat("  ", indent+1))
		e.encodeValue(v, indent+1)
	}
	e.buf.WriteString("\n")
	e.buf.WriteString(strings.Repeat("  ", indent))
	e.buf.WriteString("]")
}

func ParseTemplate(data string) (*template.Template, error) {
	return template.New("prompt").Funcs(funcMap).Parse(data)
}

func Render(tmpl *template.Template, data map[string]any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering template: %w", err)
	}
	return buf.String(), nil
}

func RenderString(templateStr string, data map[string]any) (string, error) {
	tmpl, err := ParseTemplate(templateStr)
	if err != nil {
		return "", err
	}
	return Render(tmpl, data)
}
