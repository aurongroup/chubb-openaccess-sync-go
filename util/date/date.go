package date

import (
	"strings"
	"time"
)

// Format formats a date as ISO 8601 (yyyy-MM-dd).
func Format(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.Format("2006-01-02")
}

// Parse parses a date in ISO 8601 (yyyy-MM-dd format). Returns nil for blank input.
func Parse(s string) *time.Time {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}

	return &t
}
