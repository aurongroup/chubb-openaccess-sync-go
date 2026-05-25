package csv

import (
	"openaccess-sync/pkg/data/model"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseCSV_shouldParseAllRecords(t *testing.T) {
	records, err := Parse("testdata/access.csv")
	if err != nil {
		t.Fatal(err)
	}

	if len(records.Records()) != 2 {
		t.Errorf("expected 2 records, got %d", len(records.Records()))
	}
}

func TestParseCSV_shouldParseFirstRecord(t *testing.T) {
	records, err := Parse("testdata/access.csv")
	if err != nil {
		t.Fatal(err)
	}

	r := records.Records()[0]

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
	records, err := Parse("testdata/access.csv")
	if err != nil {
		t.Fatal(err)
	}

	r := records.Records()[1]

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

// ---- Write ----

func TestPrintCSVReport_shouldWriteReadableCSV(t *testing.T) {
	activate := time.Date(2018, 9, 12, 0, 0, 0, 0, time.UTC)
	deactivate := time.Date(2020, 9, 12, 0, 0, 0, 0, time.UTC)
	r1, err := model.NewAccessRecord("8274", "BOB", "BROWN", "Coffee Fresh", "OTIS", "", "", "", "", "9017", &activate, &deactivate, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := model.NewAccessRecord("1234", "Tim", "Smith", "DALKIA", "", "", "", "", "", "1923", nil, nil, "active", "Employee")
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.CreateTemp("", "report-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	if err := Write([]*model.AccessRecord{r1, r2}, f.Name()); err != nil {
		t.Fatal(err)
	}

	records, err := Parse(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	recs := records.Records()
	if len(recs) != 2 {
		t.Fatalf("expected 2 records, got %d", len(recs))
	}
	assertStr(t, "r1.SSNO", "8274", recs[0].SSNO)
	assertStr(t, "r1.Last", "BROWN", recs[0].Last)
	assertStr(t, "r1.AccLvl1", "Coffee Fresh", recs[0].AccLvl1)
	assertStr(t, "r1.AccLvl2", "OTIS", recs[0].AccLvl2)
	assertDate(t, "r1.Activate", activate, recs[0].Activate)
	assertStr(t, "r2.SSNO", "1234", recs[1].SSNO)
	assertStr(t, "r2.AccLvl1", "DALKIA", recs[1].AccLvl1)
}

func TestPrintCSVReport_shouldWriteCorrectHeader(t *testing.T) {
	f, err := os.CreateTemp("", "report-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	if err := Write([]*model.AccessRecord{}, f.Name()); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	firstLine := strings.Split(string(data), "\n")[0]
	want := "ssno|first|last|acc_lvl1|acc_lvl2|acc_lvl3|acc_lvl4|acc_lvl5|acc_lvl6|badgeid|activate|deactivate|status|badge type"
	if firstLine != want {
		t.Errorf("header: expected %q, got %q", want, firstLine)
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
