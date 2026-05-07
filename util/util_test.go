package util

import "testing"

// ---- IsValidURL ----

func TestIsValidURL_shouldReturnTrueForHTTPS(t *testing.T) {
	if !IsValidURL("https://example.com") {
		t.Error("expected true for https://example.com")
	}
}

func TestIsValidURL_shouldReturnFalseForMissingHost(t *testing.T) {
	if IsValidURL("not-a-url") {
		t.Error("expected false for 'not-a-url'")
	}
}
