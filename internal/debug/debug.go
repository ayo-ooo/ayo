// Package debug provides centralized debug logging for ayo.
//
// Debug logging is controlled by:
//   - AYO_DEBUG environment variable (set to "1" or "true" to enable)
//   - --debug flag on CLI commands
//   - SetEnabled(true) in code
//
// Usage:
//
//	debug.Log("message", "key", value, "key2", value2)
//	debug.Logf("formatted %s", arg)
//
// Debug logs are written to stderr in a structured format.
package debug

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	enabled   bool
	output    io.Writer = os.Stderr
	mu        sync.RWMutex
	component string // Current component for context
)

func init() {
	// Check environment variable on init
	env := os.Getenv("AYO_DEBUG")
	enabled = env == "1" || strings.EqualFold(env, "true")
}

// SetEnabled enables or disables debug logging globally.
func SetEnabled(e bool) {
	mu.Lock()
	defer mu.Unlock()
	enabled = e
}

// IsEnabled returns true if debug logging is enabled.
func IsEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

// SetOutput sets the output writer for debug logs.
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	output = w
}

// SetComponent sets the current component name for context.
func SetComponent(c string) {
	mu.Lock()
	defer mu.Unlock()
	component = c
}

// Log writes a structured debug message with key-value pairs.
// Arguments after msg should be pairs of key, value.
//
//	debug.Log("request received", "method", "GET", "path", "/api")
func Log(msg string, args ...any) {
	mu.RLock()
	if !enabled {
		mu.RUnlock()
		return
	}
	w := output
	comp := component
	mu.RUnlock()

	ts := time.Now().Format("15:04:05.000")

	var sb strings.Builder
	sb.WriteString(ts)
	sb.WriteString(" [DEBUG]")

	if comp != "" {
		sb.WriteString(" [")
		sb.WriteString(comp)
		sb.WriteString("]")
	}

	sb.WriteString(" ")
	sb.WriteString(msg)

	// Format key-value pairs
	for i := 0; i+1 < len(args); i += 2 {
		key := fmt.Sprint(args[i])
		value := args[i+1]
		sb.WriteString(" ")
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(formatValue(value))
	}

	sb.WriteString("\n")
	fmt.Fprint(w, sb.String())
}

// Logf writes a formatted debug message.
//
//	debug.Logf("processing %d items", count)
func Logf(format string, args ...any) {
	mu.RLock()
	if !enabled {
		mu.RUnlock()
		return
	}
	w := output
	comp := component
	mu.RUnlock()

	ts := time.Now().Format("15:04:05.000")

	var sb strings.Builder
	sb.WriteString(ts)
	sb.WriteString(" [DEBUG]")

	if comp != "" {
		sb.WriteString(" [")
		sb.WriteString(comp)
		sb.WriteString("]")
	}

	sb.WriteString(" ")
	fmt.Fprintf(&sb, format, args...)
	sb.WriteString("\n")

	fmt.Fprint(w, sb.String())
}

// Dump writes a detailed debug dump of a value.
// Useful for inspecting complex data structures.
func Dump(label string, value any) {
	mu.RLock()
	if !enabled {
		mu.RUnlock()
		return
	}
	w := output
	mu.RUnlock()

	ts := time.Now().Format("15:04:05.000")
	fmt.Fprintf(w, "%s [DEBUG] %s:\n%+v\n", ts, label, value)
}

// formatValue formats a value for logging.
func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		if strings.ContainsAny(val, " \t\n\"") {
			return fmt.Sprintf("%q", val)
		}
		return val
	case []byte:
		return fmt.Sprintf("%q", string(val))
	case error:
		return fmt.Sprintf("%q", val.Error())
	case time.Duration:
		return val.String()
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		return fmt.Sprint(val)
	}
}

// WithComponent returns a logger that includes the component name in all logs.
func WithComponent(name string) *Logger {
	return &Logger{component: name}
}

// Logger is a component-scoped debug logger.
type Logger struct {
	component string
}

// Log writes a structured debug message for this component.
func (l *Logger) Log(msg string, args ...any) {
	mu.RLock()
	if !enabled {
		mu.RUnlock()
		return
	}
	w := output
	mu.RUnlock()

	ts := time.Now().Format("15:04:05.000")

	var sb strings.Builder
	sb.WriteString(ts)
	sb.WriteString(" [DEBUG] [")
	sb.WriteString(l.component)
	sb.WriteString("] ")
	sb.WriteString(msg)

	for i := 0; i+1 < len(args); i += 2 {
		key := fmt.Sprint(args[i])
		value := args[i+1]
		sb.WriteString(" ")
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(formatValue(value))
	}

	sb.WriteString("\n")
	fmt.Fprint(w, sb.String())
}

// Logf writes a formatted debug message for this component.
func (l *Logger) Logf(format string, args ...any) {
	mu.RLock()
	if !enabled {
		mu.RUnlock()
		return
	}
	w := output
	mu.RUnlock()

	ts := time.Now().Format("15:04:05.000")

	var sb strings.Builder
	sb.WriteString(ts)
	sb.WriteString(" [DEBUG] [")
	sb.WriteString(l.component)
	sb.WriteString("] ")
	fmt.Fprintf(&sb, format, args...)
	sb.WriteString("\n")

	fmt.Fprint(w, sb.String())
}
