package json

import (
	"openaccess-sync/pkg/util/date"
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
	case int32:
		return int(n)
	case int64:
		return int(n)
	}

	return 0
}

// PropToInt32 extracts an integer value from a JSON object map.
// JSON numbers decode as float64; ToJSON returns native int — handle both.
func PropToInt32(m map[string]any, key string) int32 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}

	switch n := v.(type) {
	case float64:
		return int32(n)
	case int:
		return int32(n)
	case int32:
		return n
	case int64:
		return int32(n)
	}

	return 0
}

// PropToInt64 extracts an integer value from a JSON object map.
// JSON numbers decode as float64; ToJSON returns native int — handle both.
func PropToInt64(m map[string]any, key string) int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}

	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	}

	return 0
}

// PropToStr extracts a string's value from a JSON object map.
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
