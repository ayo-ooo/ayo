package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseSchedule converts a natural language schedule to cron syntax.
// If the input is already valid cron syntax, it returns it unchanged.
//
// Supported natural language patterns:
//   - "every hour" -> "0 0 * * * *"
//   - "every minute" -> "0 * * * * *"
//   - "every day" -> "0 0 0 * * *"
//   - "every day at 9am" -> "0 0 9 * * *"
//   - "every day at 9:30am" -> "0 30 9 * * *"
//   - "every monday" -> "0 0 0 * * MON"
//   - "every monday at 3pm" -> "0 0 15 * * MON"
//   - "hourly" -> "0 0 * * * *"
//   - "daily" -> "0 0 0 * * *"
//   - "weekly" -> "0 0 0 * * SUN"
func ParseSchedule(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("schedule cannot be empty")
	}

	// If it looks like cron syntax (contains spaces and numbers), pass through
	if looksLikeCron(input) {
		return input, nil
	}

	// Normalize input
	normalized := strings.ToLower(input)

	// Try simple patterns first
	if schedule, ok := parseSimplePattern(normalized); ok {
		return schedule, nil
	}

	// Try "every X" patterns
	if schedule, ok := parseEveryPattern(normalized); ok {
		return schedule, nil
	}

	// Not recognized - return error with helpful message
	return "", fmt.Errorf("unrecognized schedule format: %q. Use cron syntax (e.g., \"0 0 * * * *\") or natural language (e.g., \"every hour\", \"every day at 9am\")", input)
}

// looksLikeCron checks if the input appears to be cron syntax.
// Cron syntax has 5-6 space-separated fields containing numbers, *, or day names.
func looksLikeCron(input string) bool {
	fields := strings.Fields(input)
	if len(fields) < 5 || len(fields) > 6 {
		return false
	}

	// Check if most fields look like cron (numbers, *, -, /, or day/month names)
	cronFieldPattern := regexp.MustCompile(`^[\d\*\-\/\,]+$|^(SUN|MON|TUE|WED|THU|FRI|SAT|JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)$`)
	cronLikeFields := 0
	for _, field := range fields {
		if cronFieldPattern.MatchString(strings.ToUpper(field)) {
			cronLikeFields++
		}
	}

	// If most fields look like cron, treat it as cron
	return cronLikeFields >= len(fields)-1
}

// parseSimplePattern handles single-word shortcuts.
func parseSimplePattern(input string) (string, bool) {
	patterns := map[string]string{
		"hourly":    "0 0 * * * *",
		"daily":     "0 0 0 * * *",
		"weekly":    "0 0 0 * * SUN",
		"monthly":   "0 0 0 1 * *",
		"yearly":    "0 0 0 1 1 *",
		"annually":  "0 0 0 1 1 *",
		"midnight":  "0 0 0 * * *",
		"noon":      "0 0 12 * * *",
	}

	if schedule, ok := patterns[input]; ok {
		return schedule, true
	}
	return "", false
}

// parseEveryPattern handles "every X" and "every X at Y" patterns.
func parseEveryPattern(input string) (string, bool) {
	// Remove common prefixes
	input = strings.TrimPrefix(input, "every ")
	
	// Check for "at" clause
	atParts := strings.SplitN(input, " at ", 2)
	interval := strings.TrimSpace(atParts[0])
	timeSpec := ""
	if len(atParts) == 2 {
		timeSpec = strings.TrimSpace(atParts[1])
	}

	// Parse the time specification
	hour, minute := 0, 0
	if timeSpec != "" {
		h, m, ok := parseTimeSpec(timeSpec)
		if !ok {
			return "", false
		}
		hour, minute = h, m
	}

	// Handle intervals
	switch interval {
	case "minute":
		return "0 * * * * *", true
	case "hour":
		return fmt.Sprintf("0 %d * * * *", minute), true
	case "day":
		return fmt.Sprintf("0 %d %d * * *", minute, hour), true
	case "week":
		return fmt.Sprintf("0 %d %d * * SUN", minute, hour), true
	case "month":
		return fmt.Sprintf("0 %d %d 1 * *", minute, hour), true
	case "year":
		return fmt.Sprintf("0 %d %d 1 1 *", minute, hour), true
	}

	// Check for day names
	if day, ok := parseDayName(interval); ok {
		return fmt.Sprintf("0 %d %d * * %s", minute, hour, day), true
	}

	// Check for "N hours", "N minutes" patterns
	if schedule, ok := parseIntervalPattern(interval); ok {
		return schedule, true
	}

	return "", false
}

// parseTimeSpec parses time like "9am", "9:30am", "14:00", "3pm".
func parseTimeSpec(input string) (hour, minute int, ok bool) {
	input = strings.ToLower(strings.TrimSpace(input))
	
	// Try 12-hour format with am/pm
	ampmPattern := regexp.MustCompile(`^(\d{1,2})(?::(\d{2}))?\s*(am|pm)$`)
	if matches := ampmPattern.FindStringSubmatch(input); matches != nil {
		hour, _ = strconv.Atoi(matches[1])
		if matches[2] != "" {
			minute, _ = strconv.Atoi(matches[2])
		}
		if matches[3] == "pm" && hour != 12 {
			hour += 12
		} else if matches[3] == "am" && hour == 12 {
			hour = 0
		}
		if hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59 {
			return hour, minute, true
		}
		return 0, 0, false
	}

	// Try 24-hour format
	timePattern := regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
	if matches := timePattern.FindStringSubmatch(input); matches != nil {
		hour, _ = strconv.Atoi(matches[1])
		minute, _ = strconv.Atoi(matches[2])
		if hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59 {
			return hour, minute, true
		}
	}

	// Try just hour
	if h, err := strconv.Atoi(input); err == nil && h >= 0 && h <= 23 {
		return h, 0, true
	}

	return 0, 0, false
}

// parseDayName converts day names to cron day-of-week values.
func parseDayName(input string) (string, bool) {
	days := map[string]string{
		"sunday":    "SUN",
		"monday":    "MON",
		"tuesday":   "TUE",
		"wednesday": "WED",
		"thursday":  "THU",
		"friday":    "FRI",
		"saturday":  "SAT",
		"sun":       "SUN",
		"mon":       "MON",
		"tue":       "TUE",
		"wed":       "WED",
		"thu":       "THU",
		"fri":       "FRI",
		"sat":       "SAT",
	}

	// Handle plurals
	input = strings.TrimSuffix(input, "s")
	
	if day, ok := days[input]; ok {
		return day, true
	}
	return "", false
}

// parseIntervalPattern handles "N hours", "N minutes" etc.
func parseIntervalPattern(input string) (string, bool) {
	pattern := regexp.MustCompile(`^(\d+)\s*(hour|minute|min)s?$`)
	matches := pattern.FindStringSubmatch(input)
	if matches == nil {
		return "", false
	}

	n, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch unit {
	case "minute", "min":
		if n > 0 && n <= 59 && 60%n == 0 {
			return fmt.Sprintf("0 */%d * * * *", n), true
		}
	case "hour":
		if n > 0 && n <= 23 && 24%n == 0 {
			return fmt.Sprintf("0 0 */%d * * *", n), true
		}
	}

	return "", false
}
