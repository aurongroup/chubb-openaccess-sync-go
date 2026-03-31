package main

import (
	"net/url"
	"strings"
	"time"
)

// formatDate formats a date as M/d/yyyy (matching the Java DateTimeFormatter pattern).
func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.Format("1/2/2006")
}

// parseDate parses a date in M/d/yyyy format. Returns nil for blank input.
func parseDate(s string) *time.Time {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	t, err := time.Parse("1/2/2006", s)
	if err != nil {
		return nil
	}

	return &t
}

func isValidURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	return u.Host != ""
}

// propInt extracts an integer value from a property_value_map.
// JSON numbers decode as float64; ToJSON returns native int — handle both.
func propInt(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}

	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}

	return 0
}

// propStr extracts a string value from a property_value_map.
func propStr(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}

	s, _ := v.(string)

	return s
}

// propDate extracts a date (ISO 8601 YYYY-MM-DD) from a property_value_map.
func propDate(m map[string]any, key string) *time.Time {
	s := propStr(m, key)
	if s == "" {
		return nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}

	return &t
}

// dateStr returns the ISO 8601 string for a date, or nil if the date is nil.
// Used by ToJSON methods.
func dateStr(t *time.Time) any {
	if t == nil {
		return nil
	}

	return t.Format("2006-01-02")
}
