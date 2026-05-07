package models

import (
	"openaccess-sync/util/date"
	"time"
)

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
	return date.Parse(propStr(m, key))
}

// dateStr returns the ISO 8601 string for a date, or nil if the date is nil.
func dateStr(t *time.Time) any {
	if t == nil {
		return nil
	}

	return t.Format("2006-01-02")
}
