package model

import "testing"

func TestNewBadgeStatus_shouldParseIdAndName(t *testing.T) {
	s, err := NewBadgeStatus(3, "Active")
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != 3 {
		t.Errorf("ID: expected 3, got %d", s.ID)
	}
	if s.Name != "Active" {
		t.Errorf("Name: expected %q, got %q", "Active", s.Name)
	}
}

func TestNewBadgeStatus_shouldErrorWhenIDMissing(t *testing.T) {
	_, err := NewBadgeStatus(0, "Active")
	if err != ErrBadgeStatusMissingID {
		t.Errorf("expected ErrBadgeStatusMissingID, got %v", err)
	}
}

func TestNewBadgeStatus_shouldErrorWhenNameMissing(t *testing.T) {
	_, err := NewBadgeStatus(1, "")
	if err != ErrBadgeStatusMissingName {
		t.Errorf("expected ErrBadgeStatusMissingName, got %v", err)
	}
}
