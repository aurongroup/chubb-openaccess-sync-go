package model

import "testing"

func TestNewBadgeType_shouldParseIdAndName(t *testing.T) {
	bt, err := NewBadgeType(2, "Employee")
	if err != nil {
		t.Fatal(err)
	}
	if bt.ID != 2 {
		t.Errorf("ID: expected 2, got %d", bt.ID)
	}
	if bt.Name != "Employee" {
		t.Errorf("Name: expected %q, got %q", "Employee", bt.Name)
	}
}

func TestNewBadgeType_shouldErrorWhenIDMissing(t *testing.T) {
	_, err := NewBadgeType(0, "Employee")
	if err != ErrBadgeTypeMissingID {
		t.Errorf("expected ErrBadgeTypeMissingID, got %v", err)
	}
}

func TestNewBadgeType_shouldErrorWhenNameMissing(t *testing.T) {
	_, err := NewBadgeType(1, "")
	if err != ErrBadgeTypeMissingName {
		t.Errorf("expected ErrBadgeTypeMissingName, got %v", err)
	}
}
