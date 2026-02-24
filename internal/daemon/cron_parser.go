package daemon

import (
	"fmt"
	"regexp"
	"strings"
)

// cronAliases maps friendly aliases to cron expressions.
var cronAliases = map[string]string{
	"@hourly":   "0 * * * *",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@weekly":   "0 0 * * 0",
	"@monthly":  "0 0 1 * *",
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@weekdays": "0 9 * * 1-5",
	"@weekends": "0 9 * * 0,6",
}

// ExpandCronAlias expands a cron alias to its expression, or returns the input unchanged.
func ExpandCronAlias(expr string) string {
	if expanded, ok := cronAliases[strings.ToLower(expr)]; ok {
		return expanded
	}
	return expr
}

// ValidateCronExpression validates a cron expression.
// Returns nil if valid, error with helpful message if invalid.
func ValidateCronExpression(expr string) error {
	// Handle @every syntax for intervals
	if strings.HasPrefix(expr, "@every ") {
		return nil // gocron handles this
	}

	// Check for alias
	if strings.HasPrefix(expr, "@") {
		lower := strings.ToLower(expr)
		if _, ok := cronAliases[lower]; !ok {
			return &CronError{
				Expression: expr,
				Message:    fmt.Sprintf("unknown cron alias: %s", expr),
				Suggestion: "Valid aliases: @hourly, @daily, @weekly, @monthly, @yearly, @weekdays, @weekends",
			}
		}
		return nil
	}

	// Parse cron expression - validate by checking field count and basic structure
	fields := strings.Fields(expr)
	switch len(fields) {
	case 5:
		// Standard cron (minute hour day month weekday)
		// gocron.CronJob doesn't return error directly - validation happens at NewJob
		// We do basic field validation here
		if err := validateCronFields(fields); err != nil {
			return &CronError{
				Expression: expr,
				Message:    err.Error(),
				Suggestion: "Example: '0 9 * * *' runs at 9:00 AM every day",
			}
		}
	case 6:
		// Extended cron with seconds (second minute hour day month weekday)
		if err := validateCronFields(fields); err != nil {
			return &CronError{
				Expression: expr,
				Message:    err.Error(),
				Suggestion: "Example: '0 0 9 * * *' runs at 9:00:00 AM every day",
			}
		}
	default:
		return &CronError{
			Expression: expr,
			Message:    fmt.Sprintf("expected 5 fields (minute hour day month weekday) or 6 fields (with seconds), got %d", len(fields)),
			Suggestion: "Example: '0 9 * * *' runs at 9:00 AM every day\nSee 'ayo help cron' for syntax reference",
		}
	}

	return nil
}

// CronError provides helpful error messages for cron expression issues.
type CronError struct {
	Expression string
	Message    string
	Suggestion string
}

func (e *CronError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Invalid cron expression '%s'\n", e.Expression))
	sb.WriteString(fmt.Sprintf("       %s\n", e.Message))
	if e.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("\n       %s", e.Suggestion))
	}
	return sb.String()
}

// ParseCronSchedule parses a schedule string, expanding aliases and validating.
// Returns the expanded cron expression and any error.
func ParseCronSchedule(schedule string) (string, error) {
	// First check if it's valid
	if err := ValidateCronExpression(schedule); err != nil {
		return "", err
	}

	// Expand alias if present
	return ExpandCronAlias(schedule), nil
}

// GetCronAliases returns all available aliases for help display.
func GetCronAliases() map[string]string {
	// Return a copy to prevent modification
	result := make(map[string]string, len(cronAliases))
	for k, v := range cronAliases {
		result[k] = v
	}
	return result
}

// CronHelp returns formatted help text for cron syntax.
func CronHelp() string {
	return `Cron Expression Syntax
======================

Format: minute hour day month weekday

Field     Allowed Values
-----     --------------
minute    0-59
hour      0-23
day       1-31
month     1-12 or JAN-DEC
weekday   0-6 or SUN-SAT (0=Sunday)

Special Characters
------------------
*   Any value
,   List separator (1,3,5)
-   Range (1-5)
/   Step (/15 = every 15)

Aliases
-------
@hourly    0 * * * *      Every hour at minute 0
@daily     0 0 * * *      Every day at midnight
@weekly    0 0 * * 0      Every Sunday at midnight
@monthly   0 0 1 * *      First of month at midnight
@yearly    0 0 1 1 *      January 1st at midnight
@weekdays  0 9 * * 1-5    9am Monday through Friday
@weekends  0 9 * * 0,6    9am Saturday and Sunday

Extended Syntax (6 fields)
--------------------------
second minute hour day month weekday

Example: '0 0 9 * * *' runs at exactly 9:00:00 AM

Examples
--------
0 9 * * *       Every day at 9:00 AM
0 */2 * * *     Every 2 hours
0 9 * * 1-5     Weekdays at 9:00 AM
0 0 1,15 * *    1st and 15th of month at midnight
30 8 * * 1      Every Monday at 8:30 AM
`
}

// cronFieldPattern matches valid cron field values.
var cronFieldPattern = regexp.MustCompile(`^(\*|[0-9]+(-[0-9]+)?)(\/[0-9]+)?(,(\*|[0-9]+(-[0-9]+)?)(\/[0-9]+)?)*$|^[A-Za-z]{3}(-[A-Za-z]{3})?$`)

// validateCronFields validates the structure of cron fields.
func validateCronFields(fields []string) error {
	for i, field := range fields {
		if !cronFieldPattern.MatchString(field) {
			return fmt.Errorf("invalid cron field %d: %s", i+1, field)
		}
	}
	return nil
}
