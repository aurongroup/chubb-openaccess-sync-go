package main

import (
	"testing"
	"time"

	"openaccess-sync/models"
)

// ---- ContentEquals ----

func TestContentEquals_shouldReturnTrueForIdenticalRecords(t *testing.T) {
	r, err := models.NewAccessRecord("A", "Bob", "Smith", "L1", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	if !ContentEquals(r, r, nil) {
		t.Error("expected ContentEquals to return true for the same record")
	}
}

func TestContentEquals_shouldReturnFalseWhenFieldDiffers(t *testing.T) {
	base := func(first string) *models.AccessRecord {
		r, err := models.NewAccessRecord("A", first, "Smith", "L1", "", "", "", "", "", "100", nil, nil, "active", "Employee")
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	if ContentEquals(base("Bob"), base("Alice"), nil) {
		t.Error("expected false when First differs")
	}
}

func TestContentEquals_shouldCompareDatesCorrectly(t *testing.T) {
	d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	r1, err := models.NewAccessRecord("A", "", "Smith", "", "", "", "", "", "", "100", &d, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := models.NewAccessRecord("A", "", "Smith", "", "", "", "", "", "", "100", &d, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	if !ContentEquals(r1, r2, nil) {
		t.Error("expected true for records with equal date values")
	}
}

func TestContentEquals_shouldReturnFalseWhenOneDateNil(t *testing.T) {
	d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	r1, err := models.NewAccessRecord("A", "", "Smith", "", "", "", "", "", "", "100", &d, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := models.NewAccessRecord("A", "", "Smith", "", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	if ContentEquals(r1, r2, nil) {
		t.Error("expected false when one Activate is nil and the other is set")
	}
}

// ---- SyncStatus.String ----

func TestSyncStatus_String_shouldReturnAllLabels(t *testing.T) {
	cases := []struct {
		s    models.SyncStatus
		want string
	}{
		{models.SyncNew, "new"},
		{models.SyncExisting, "existing"},
		{models.SyncUpdate, "update"},
		{models.SyncDelete, "delete"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("SyncStatus(%d).String(): expected %q, got %q", c.s, c.want, got)
		}
	}
}

func TestSyncStatus_String_shouldReturnUnknownForInvalidValue(t *testing.T) {
	if got := models.SyncStatus(99).String(); got != "unknown" {
		t.Errorf("expected %q, got %q", "unknown", got)
	}
}
