package model

import (
	"testing"
	"time"
)

func TestNewAccessRecord_shouldErrorWhenLastMissing(t *testing.T) {
	_, err := NewAccessRecord("", "", "", "", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != ErrAccessRecordMissingLast {
		t.Errorf("expected ErrAccessRecordMissingLast, got %v", err)
	}
}

func TestNewAccessRecord_shouldErrorWhenBadgeIDMissing(t *testing.T) {
	_, err := NewAccessRecord("", "", "Smith", "", "", "", "", "", "", "", nil, nil, "active", "Employee")
	if err != ErrAccessRecordMissingBadgeID {
		t.Errorf("expected ErrAccessRecordMissingBadgeID, got %v", err)
	}
}

func TestNewAccessRecord_shouldErrorWhenStatusMissing(t *testing.T) {
	_, err := NewAccessRecord("", "", "Smith", "", "", "", "", "", "", "100", nil, nil, "", "Employee")
	if err != ErrAccessRecordMissingStatus {
		t.Errorf("expected ErrAccessRecordMissingStatus, got %v", err)
	}
}

func TestNewAccessRecord_shouldErrorWhenBadgeTypeMissing(t *testing.T) {
	_, err := NewAccessRecord("", "", "Smith", "", "", "", "", "", "", "100", nil, nil, "active", "")
	if err != ErrAccessRecordMissingBadgeType {
		t.Errorf("expected ErrAccessRecordMissingBadgeType, got %v", err)
	}
}

func TestAccessRecord_ToRow_shouldReturnAllFieldsInOrder(t *testing.T) {
	activate := time.Date(2018, 9, 12, 0, 0, 0, 0, time.UTC)
	deactivate := time.Date(2020, 9, 12, 0, 0, 0, 0, time.UTC)
	r, err := NewAccessRecord("8274", "BOB", "BROWN", "L1", "L2", "L3", "L4", "L5", "L6", "9017", &activate, &deactivate, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	row := r.ToRow()
	if len(row) != 14 {
		t.Fatalf("expected 14 fields, got %d", len(row))
	}
	expected := []string{"8274", "BOB", "BROWN", "L1", "L2", "L3", "L4", "L5", "L6", "9017", "2018-09-12", "2020-09-12", "active", "Employee"}
	for i, want := range expected {
		if row[i] != want {
			t.Errorf("field[%d]: expected %q, got %q", i, want, row[i])
		}
	}
}

func TestAccessRecord_ToRow_shouldFormatDatesCorrectly(t *testing.T) {
	d := time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)
	r, err := NewAccessRecord("", "", "Smith", "", "", "", "", "", "", "1", &d, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	row := r.ToRow()
	if row[10] != "2024-03-05" {
		t.Errorf("activate: expected %q, got %q", "2024-03-05", row[10])
	}
	if row[11] != "" {
		t.Errorf("deactivate: expected empty string, got %q", row[11])
	}
}
