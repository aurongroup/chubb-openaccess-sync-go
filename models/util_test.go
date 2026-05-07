package models

import (
	"testing"
	"time"
)

// ---- dateStr ----

func TestDateStr_shouldReturnISOString(t *testing.T) {
	d := time.Date(2018, 9, 12, 0, 0, 0, 0, time.UTC)
	got := dateStr(&d)
	if got != "2018-09-12" {
		t.Errorf("expected %q, got %v", "2018-09-12", got)
	}
}

func TestDateStr_shouldReturnNilForNilDate(t *testing.T) {
	got := dateStr(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// ---- propInt ----

func TestPropInt_shouldReturnIntFromFloat64(t *testing.T) {
	m := map[string]any{"ID": float64(42)}
	if got := propInt(m, "ID"); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

func TestPropInt_shouldReturnIntFromInt(t *testing.T) {
	m := map[string]any{"ID": 7}
	if got := propInt(m, "ID"); got != 7 {
		t.Errorf("expected 7, got %d", got)
	}
}

func TestPropInt_shouldReturnZeroForMissingKey(t *testing.T) {
	m := map[string]any{}
	if got := propInt(m, "ID"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestPropInt_shouldReturnZeroForNonNumericValue(t *testing.T) {
	m := map[string]any{"ID": "not-a-number"}
	if got := propInt(m, "ID"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

// ---- propStr ----

func TestPropStr_shouldReturnString(t *testing.T) {
	m := map[string]any{"Name": "Main Entrance"}
	if got := propStr(m, "Name"); got != "Main Entrance" {
		t.Errorf("expected %q, got %q", "Main Entrance", got)
	}
}

func TestPropStr_shouldReturnEmptyForMissingKey(t *testing.T) {
	m := map[string]any{}
	if got := propStr(m, "Name"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
