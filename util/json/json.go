package json

import (
	"openaccess-sync/util/date"
	"time"
)

// PropToInt extracts an integer value from a JSON object map.
// JSON numbers decode as float64; ToJSON returns native int — handle both.
func PropToInt(m map[string]any, key string) int {
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

// PropToStr extracts a string value from a JSON object map.
func PropToStr(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}

	s, _ := v.(string)

	return s
}

// PropToDate extracts a date (ISO 8601 YYYY-MM-DD) from a JSON object map.
func PropToDate(m map[string]any, key string) *time.Time {
	return date.Parse(PropToStr(m, key))
}
