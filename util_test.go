package main

import (
	"testing"
	"time"
)

// ---- formatDate ----

func TestFormatDate_shouldReturnFormattedDate(t *testing.T) {
	d := time.Date(2018, 9, 12, 0, 0, 0, 0, time.UTC)
	got := formatDate(&d)
	if got != "12/09/2018" {
		t.Errorf("expected %q, got %q", "12/09/2018", got)
	}
}

func TestFormatDate_shouldReturnEmptyStringForNil(t *testing.T) {
	got := formatDate(nil)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// ---- parseDate ----

func TestParseDate_shouldParseValidDate(t *testing.T) {
	d := parseDate("12/09/2018")
	if d == nil {
		t.Fatal("expected non-nil date")
	}
	want := time.Date(2018, 9, 12, 0, 0, 0, 0, time.UTC)
	if !d.Equal(want) {
		t.Errorf("expected %v, got %v", want, *d)
	}
}

func TestParseDate_shouldReturnNilForBlankInput(t *testing.T) {
	for _, s := range []string{"", "  "} {
		if got := parseDate(s); got != nil {
			t.Errorf("parseDate(%q): expected nil, got %v", s, got)
		}
	}
}

func TestParseDate_shouldReturnNilForInvalidFormat(t *testing.T) {
	got := parseDate("2018-09-12")
	if got != nil {
		t.Errorf("expected nil for ISO format input, got %v", got)
	}
}

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

// ---- isValidURL ----

func TestIsValidURL_shouldReturnTrueForHTTPS(t *testing.T) {
	if !isValidURL("https://example.com") {
		t.Error("expected true for https://example.com")
	}
}

func TestIsValidURL_shouldReturnFalseForMissingHost(t *testing.T) {
	if isValidURL("not-a-url") {
		t.Error("expected false for 'not-a-url'")
	}
}
