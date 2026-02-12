package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/alexcabrera/ayo/internal/util"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/charmbracelet/lipgloss/table"
)

// JSONToRenderedOutput converts JSON to a styled terminal output using lipgloss components.
func JSONToRenderedOutput(jsonStr string) (string, bool) {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", false
	}

	var b strings.Builder
	renderAny(&b, data, 0)
	return strings.TrimSpace(b.String()), true
}

func renderAny(b *strings.Builder, v any, depth int) {
	if depth > 3 {
		b.WriteString("(...)\n")
		return
	}

	switch val := v.(type) {
	case map[string]any:
		renderObject(b, val, depth)
	case []any:
		renderArray(b, val, depth)
	default:
		fmt.Fprintf(b, "%v", val)
	}
}

func renderObject(b *strings.Builder, obj map[string]any, depth int) {
	// Check for SearXNG response pattern
	if isSearXNGResponse(obj) {
		renderSearXNGResponse(b, obj)
		return
	}

	// Check if this looks like a search result
	if isSearchResult(obj) {
		renderSearchResultItem(b, obj)
		return
	}

	// Render as key-value table
	renderKeyValueTable(b, obj)
}

func renderKeyValueTable(b *strings.Builder, obj map[string]any) {
	// Collect non-null, non-empty values
	var rows [][]string
	keys := getSortedKeys(obj)

	for _, k := range keys {
		v := obj[k]
		if v == nil {
			continue
		}

		valStr := formatValue(v)
		if valStr == "" {
			continue
		}

		rows = append(rows, []string{formatKeyName(k), valStr})
	}

	if len(rows) == 0 {
		return
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 0 {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("212")).
					Bold(true).
					Padding(0, 1)
			}
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(0, 1).
				Width(60)
		})

	for _, row := range rows {
		t.Row(row[0], row[1])
	}

	b.WriteString(t.Render())
	b.WriteString("\n")
}

func renderArray(b *strings.Builder, arr []any, depth int) {
	if len(arr) == 0 {
		return
	}

	// Check if array contains search results
	if len(arr) > 0 {
		if obj, ok := arr[0].(map[string]any); ok {
			if isSearchResult(obj) {
				renderSearchResults(b, arr)
				return
			}
		}
	}

	// Render as a list
	items := make([]any, 0, len(arr))
	maxItems := 15
	for i, item := range arr {
		if i >= maxItems {
			items = append(items, fmt.Sprintf("... and %d more", len(arr)-maxItems))
			break
		}
		items = append(items, formatValue(item))
	}

	l := list.New(items...).
		Enumerator(list.Bullet).
		EnumeratorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("212")))

	b.WriteString(l.String())
	b.WriteString("\n")
}

// SearXNG-specific rendering
func isSearXNGResponse(obj map[string]any) bool {
	_, hasQuery := obj["query"]
	_, hasResults := obj["results"]
	return hasQuery && hasResults
}

func renderSearXNGResponse(b *strings.Builder, obj map[string]any) {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212"))

	// Query header
	if query, ok := obj["query"].(string); ok {
		b.WriteString(headerStyle.Render("Search: "+query) + "\n\n")
	}

	// Results
	if results, ok := obj["results"].([]any); ok && len(results) > 0 {
		b.WriteString(headerStyle.Render("Results") + "\n")
		renderSearchResults(b, results)
	}

	// Answers (direct answers)
	if answers, ok := obj["answers"].([]any); ok && len(answers) > 0 {
		b.WriteString("\n" + headerStyle.Render("Answers") + "\n")
		items := make([]any, len(answers))
		for i, a := range answers {
			items[i] = fmt.Sprintf("%v", a)
		}
		l := list.New(items...).Enumerator(list.Bullet)
		b.WriteString(l.String() + "\n")
	}

	// Suggestions
	if suggestions, ok := obj["suggestions"].([]any); ok && len(suggestions) > 0 {
		b.WriteString("\n" + headerStyle.Render("Related") + " ")
		strs := make([]string, 0, len(suggestions))
		for _, s := range suggestions {
			if str, ok := s.(string); ok {
				strs = append(strs, str)
			}
		}
		mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		b.WriteString(mutedStyle.Render(strings.Join(strs, ", ")) + "\n")
	}
}

func isSearchResult(obj map[string]any) bool {
	_, hasURL := obj["url"]
	_, hasTitle := obj["title"]
	return hasURL && hasTitle
}

func renderSearchResults(b *strings.Builder, results []any) {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	urlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Italic(true)
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	metaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	maxResults := 8
	for i, r := range results {
		if i >= maxResults {
			fmt.Fprintf(b, "\n  ... and %d more results\n", len(results)-maxResults)
			break
		}

		result, ok := r.(map[string]any)
		if !ok {
			continue
		}

		title, _ := result["title"].(string)
		url, _ := result["url"].(string)
		content, _ := result["content"].(string)
		metadata, _ := result["metadata"].(string)

		if title == "" {
			continue
		}

		// Title
		b.WriteString("\n  " + titleStyle.Render(util.Truncate(title, 70)) + "\n")

		// URL
		if url != "" {
			b.WriteString("  " + urlStyle.Render(util.Truncate(url, 80)) + "\n")
		}

		// Content snippet
		if content != "" {
			b.WriteString("  " + contentStyle.Render(util.Truncate(content, 100)) + "\n")
		}

		// Metadata (date, source)
		if metadata != "" {
			b.WriteString("  " + metaStyle.Render(metadata) + "\n")
		}
	}
}

func renderSearchResultItem(b *strings.Builder, obj map[string]any) {
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	title, _ := obj["title"].(string)
	url, _ := obj["url"].(string)

	if title != "" {
		b.WriteString(titleStyle.Render(title) + "\n")
	}
	if url != "" {
		b.WriteString(urlStyle.Render(url) + "\n")
	}
}

// Helper functions

func formatKeyName(key string) string {
	// Convert snake_case to Title Case
	key = strings.ReplaceAll(key, "_", " ")
	words := strings.Fields(key)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return util.Truncate(strings.TrimSpace(val), 60)
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case bool:
		return fmt.Sprintf("%v", val)
	case []any:
		if len(val) == 0 {
			return ""
		}
		if len(val) <= 3 {
			strs := make([]string, len(val))
			for i, item := range val {
				strs[i] = fmt.Sprintf("%v", item)
			}
			return strings.Join(strs, ", ")
		}
		return fmt.Sprintf("[%d items]", len(val))
	case map[string]any:
		return "{...}"
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func getSortedKeys(obj map[string]any) []string {
	// Priority keys first
	priority := []string{"title", "name", "query", "status", "message", "error", "url", "content", "description"}
	
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Reorder with priority keys first
	result := make([]string, 0, len(keys))
	seen := make(map[string]bool)

	for _, pk := range priority {
		if _, exists := obj[pk]; exists {
			result = append(result, pk)
			seen[pk] = true
		}
	}

	for _, k := range keys {
		if !seen[k] {
			result = append(result, k)
		}
	}

	return result
}

