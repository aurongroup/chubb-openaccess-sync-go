package main

import (
	"openaccess-sync/pkg/data/model"
	"testing"
	"time"
)

// ---- ContentEquals ----

func TestContentEquals_shouldReturnTrueForIdenticalRecords(t *testing.T) {
	r, err := model.NewAccessRecord("A", "Bob", "Smith", "L1", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	if !ContentEquals(r, r, nil) {
		t.Error("expected ContentEquals to return true for the same record")
	}
}

func TestContentEquals_shouldReturnFalseWhenFieldDiffers(t *testing.T) {
	base := func(first string) *model.AccessRecord {
		r, err := model.NewAccessRecord("A", first, "Smith", "L1", "", "", "", "", "", "100", nil, nil, "active", "Employee")
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
	r1, err := model.NewAccessRecord("A", "", "Smith", "", "", "", "", "", "", "100", &d, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := model.NewAccessRecord("A", "", "Smith", "", "", "", "", "", "", "100", &d, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	if !ContentEquals(r1, r2, nil) {
		t.Error("expected true for records with equal date values")
	}
}

func TestContentEquals_shouldReturnFalseWhenOneDateNil(t *testing.T) {
	d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	r1, err := model.NewAccessRecord("A", "", "Smith", "", "", "", "", "", "", "100", &d, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := model.NewAccessRecord("A", "", "Smith", "", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	if ContentEquals(r1, r2, nil) {
		t.Error("expected false when one Activate is nil and the other is set")
	}
}
