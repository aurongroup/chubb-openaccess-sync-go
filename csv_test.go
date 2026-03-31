package main

import (
	"testing"
	"time"
)

func TestParseCSV_shouldParseAllRecords(t *testing.T) {
	records, err := ParseCSV("testdata/access.csv")
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}

func TestParseCSV_shouldParseFirstRecord(t *testing.T) {
	records, err := ParseCSV("testdata/access.csv")
	if err != nil {
		t.Fatal(err)
	}

	r := records[0]

	assertStr(t, "SSNO", "8274", r.SSNO)
	assertStr(t, "First", "BOB", r.First)
	assertStr(t, "Last", "BROWN", r.Last)
	assertStr(t, "AccLvl1", "Coffee Fresh", r.AccLvl1)
	assertStr(t, "AccLvl2", "OTIS (ALL LEVELS)", r.AccLvl2)
	assertStr(t, "AccLvl3", "COMMS ROOM L2", r.AccLvl3)
	assertStr(t, "AccLvl4", "CDI Super", r.AccLvl4)
	assertStr(t, "AccLvl5", "OCS Palace", r.AccLvl5)
	assertStr(t, "AccLvl6", "DALKIA", r.AccLvl6)
	assertStr(t, "BadgeID", "9017", r.BadgeID)
	assertDate(t, "Activate", time.Date(2018, 9, 12, 0, 0, 0, 0, time.UTC), r.Activate)
	assertDate(t, "Deactivate", time.Date(2020, 9, 12, 0, 0, 0, 0, time.UTC), r.Deactivate)
	assertStr(t, "Status", "active", r.Status)
	assertStr(t, "BadgeType", "Employee", r.BadgeType)
}

func TestParseCSV_shouldHandleEmptyAccessLevels(t *testing.T) {
	records, err := ParseCSV("testdata/access.csv")
	if err != nil {
		t.Fatal(err)
	}

	r := records[1]

	assertStr(t, "SSNO", "1234", r.SSNO)
	assertStr(t, "First", "Tim", r.First)
	assertStr(t, "Last", "Smith", r.Last)
	assertStr(t, "AccLvl1", "DALKIA", r.AccLvl1)
	assertStr(t, "AccLvl2", "", r.AccLvl2)
	assertStr(t, "AccLvl3", "", r.AccLvl3)
	assertStr(t, "AccLvl4", "", r.AccLvl4)
	assertStr(t, "AccLvl5", "", r.AccLvl5)
	assertStr(t, "AccLvl6", "", r.AccLvl6)
	assertStr(t, "BadgeID", "1923", r.BadgeID)
	assertDate(t, "Activate", time.Date(2016, 9, 11, 0, 0, 0, 0, time.UTC), r.Activate)
	assertDate(t, "Deactivate", time.Date(2022, 9, 12, 0, 0, 0, 0, time.UTC), r.Deactivate)
	assertStr(t, "Status", "active", r.Status)
	assertStr(t, "BadgeType", "Employee", r.BadgeType)
}

// ---- CompareRecords tests ----

func TestCompareRecords_shouldMarkNewRecord(t *testing.T) {
	r, err := NewAccessRecord("A", "", "Smith", "", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}

	result := CompareRecords([]*AccessRecord{r}, []*AccessRecord{})
	if len(result) != 1 || result[0].SyncStatus != SyncNew {
		t.Errorf("expected 1 NEW record, got %v", result)
	}
}

func TestCompareRecords_shouldMarkExistingRecord(t *testing.T) {
	r, err := NewAccessRecord("A", "Bob", "Smith", "", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}

	result := CompareRecords([]*AccessRecord{r}, []*AccessRecord{r})
	if len(result) != 1 || result[0].SyncStatus != SyncExisting {
		t.Errorf("expected 1 EXISTING record, got %v", result)
	}
}

func TestCompareRecords_shouldMarkUpdatedRecord(t *testing.T) {
	csvRec, err := NewAccessRecord("A", "New", "Smith", "", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}

	apiRec, err := NewAccessRecord("A", "Old", "Smith", "", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}

	result := CompareRecords([]*AccessRecord{csvRec}, []*AccessRecord{apiRec})
	if len(result) != 1 || result[0].SyncStatus != SyncUpdate {
		t.Errorf("expected 1 UPDATE record, got %v", result)
	}
}

func TestCompareRecords_shouldMarkDeletedRecord(t *testing.T) {
	r, err := NewAccessRecord("A", "", "Smith", "", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}

	result := CompareRecords([]*AccessRecord{}, []*AccessRecord{r})
	if len(result) != 1 || result[0].SyncStatus != SyncDelete {
		t.Errorf("expected 1 DELETE record, got %v", result)
	}
}

// helpers

func assertStr(t *testing.T, field, want, got string) {
	t.Helper()
	if want != got {
		t.Errorf("%s: expected %q, got %q", field, want, got)
	}
}

func assertDate(t *testing.T, field string, want time.Time, got *time.Time) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: expected %v, got nil", field, want)
		return
	}

	if !got.Equal(want) {
		t.Errorf("%s: expected %v, got %v", field, want, *got)
	}
}
