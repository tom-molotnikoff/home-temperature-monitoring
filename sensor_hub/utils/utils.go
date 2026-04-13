package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// NormalizeTimeToSpaceFormat parses a time string in various formats and
// returns it in "YYYY-MM-DD HH:MM:SS" UTC. All timezone-aware inputs are
// converted to UTC so that stored timestamps are always comparable.
func NormalizeTimeToSpaceFormat(s string) string {
	if s == "" {
		return s
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t.UTC().Format("2006-01-02 15:04:05")
		}
	}

	if sec, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(sec, 0).UTC().Format("2006-01-02 15:04:05")
	}
	return s
}

// NormalizeDateTimeParam parses an API date/datetime parameter and returns
// a "YYYY-MM-DD HH:MM:SS" UTC string suitable for SQL BETWEEN queries.
// Accepts: YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS, YYYY-MM-DDTHH:MM:SSZ,
// YYYY-MM-DDTHH:MM:SS±HH:MM. For date-only input, useEndOfDay controls
// whether midnight (false) or 23:59:59 (true) is returned.
func NormalizeDateTimeParam(s string, useEndOfDay bool) (string, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t.UTC().Format("2006-01-02 15:04:05"), nil
		}
	}
	// Date-only: expand to start or end of day
	if t, err := time.Parse("2006-01-02", s); err == nil {
		if useEndOfDay {
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		return t.UTC().Format("2006-01-02 15:04:05"), nil
	}
	return "", fmt.Errorf("unrecognised date/time format: %s", s)
}

var ReadPropertiesFile = func(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open properties file: %w", err)
	}
	defer file.Close()

	props := make(map[string]string)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			props[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read properties file: %w", err)
	}
	return props, nil
}
