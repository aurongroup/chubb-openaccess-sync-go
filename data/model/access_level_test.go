package model

import "testing"

func TestNewAccessLevel_shouldParseIdAndName(t *testing.T) {
	al, err := NewAccessLevel(5, "Main Entrance")
	if err != nil {
		t.Fatal(err)
	}
	if al.ID != 5 {
		t.Errorf("ID: expected 5, got %d", al.ID)
	}
	if al.Name != "Main Entrance" {
		t.Errorf("Name: expected %q, got %q", "Main Entrance", al.Name)
	}
}

func TestNewAccessLevel_shouldErrorWhenIDMissing(t *testing.T) {
	_, err := NewAccessLevel(0, "Main Entrance")
	if err != ErrAccessLevelMissingID {
		t.Errorf("expected ErrAccessLevelMissingID, got %v", err)
	}
}

func TestNewAccessLevel_shouldErrorWhenNameMissing(t *testing.T) {
	_, err := NewAccessLevel(1, "")
	if err != ErrAccessLevelMissingName {
		t.Errorf("expected ErrAccessLevelMissingName, got %v", err)
	}
}
