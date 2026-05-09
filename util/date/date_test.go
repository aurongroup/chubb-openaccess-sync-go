package date

import (
	"testing"
	"time"
)

// ---- Format ----

func TestFormat_shouldReturnFormattedDate(t *testing.T) {
	d := time.Date(2018, 9, 12, 0, 0, 0, 0, time.UTC)
	got := Format(&d)
	if got != "2018-09-12" {
		t.Errorf("expected %q, got %q", "2018-09-12", got)
	}
}

func TestFormat_shouldReturnEmptyStringForNil(t *testing.T) {
	got := Format(nil)
	if got != "" {
		t.Errorf("expected empty strings, got %q", got)
	}
}

// ---- Parse ----

func TestParse_shouldParseValidDate(t *testing.T) {
	d := Parse("2018-09-12")
	if d == nil {
		t.Fatal("expected non-nil date")
	}
	want := time.Date(2018, 9, 12, 0, 0, 0, 0, time.UTC)
	if !d.Equal(want) {
		t.Errorf("expected %v, got %v", want, *d)
	}
}

func TestParse_shouldReturnNilForBlankInput(t *testing.T) {
	for _, s := range []string{"", "  "} {
		if got := Parse(s); got != nil {
			t.Errorf("ParseDate(%q): expected nil, got %v", s, got)
		}
	}
}

func TestParse_shouldReturnNilForInvalidFormat(t *testing.T) {
	got := Parse("not-a-date")
	if got != nil {
		t.Errorf("expected nil for invalid input, got %v", got)
	}
}
