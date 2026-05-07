package url

import "testing"

// ---- IsValidURL ----

func TestIsValid_shouldReturnTrueForHTTPS(t *testing.T) {
	if !IsValid("https://example.com") {
		t.Error("expected true for https://example.com")
	}
}

func TestIsValid_shouldReturnFalseForMissingHost(t *testing.T) {
	if IsValid("not-a-url") {
		t.Error("expected false for 'not-a-url'")
	}
}
