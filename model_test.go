package main

import (
	"testing"
	"time"
)

// newTestCache returns a DataCache pre-populated with a badge status (ID=1)
// and badge type (ID=1), sufficient for constructing badges in tests.
func newTestCache() *DataCache {
	c := NewDataCache(nil)
	c.statuses[1] = &LnlBadgeStatus{ID: 1, Name: "Active"}
	c.badgeTypes[1] = &LnlBadgeType{ID: 1, Name: "Standard"}
	return c
}

// ---- LnlBadge tests ----

func TestLnlBadge_fromProps_shouldParseId(t *testing.T) {
	props := map[string]any{
		"ID":       float64(42),
		"BADGEKEY": float64(1),
		"STATUS":   float64(1),
		"TYPE":     float64(1),
	}
	badge, err := NewLnlBadge(props, newTestCache())
	if err != nil {
		t.Fatal(err)
	}
	if badge.ID != 42 {
		t.Errorf("expected ID 42, got %d", badge.ID)
	}
}

func TestLnlBadge_fromProps_shouldErrorWhenIdAbsent(t *testing.T) {
	props := map[string]any{
		"BADGEKEY": float64(1),
	}
	_, err := NewLnlBadge(props, NewDataCache(nil))
	if err != ErrBadgeMissingID {
		t.Errorf("expected ErrBadgeMissingID, got %v", err)
	}
}

func TestLnlBadge_toJSON_shouldCreateCorrectJsonStructure(t *testing.T) {
	props := map[string]any{
		"ID":         float64(7),
		"BADGEKEY":   float64(1001),
		"STATUS":     float64(1),
		"TYPE":       float64(1),
		"ACTIVATE":   "2025-01-01",
		"DEACTIVATE": "2026-12-31",
	}
	badge, err := NewLnlBadge(props, newTestCache())
	if err != nil {
		t.Fatal(err)
	}
	j := badge.ToJSON()

	if j["type_name"] != "Lnl_Badge" {
		t.Errorf("expected type_name 'Lnl_Badge', got %v", j["type_name"])
	}
	pvm, ok := j["property_value_map"].(map[string]any)
	if !ok {
		t.Fatal("expected property_value_map to be a map")
	}
	if pvm["badgeKey"] != 1001 {
		t.Errorf("expected badgeKey 1001, got %v", pvm["badgeKey"])
	}
	if pvm["activate"] != "2025-01-01" {
		t.Errorf("expected activate '2025-01-01', got %v", pvm["activate"])
	}
	if pvm["deactivate"] != "2026-12-31" {
		t.Errorf("expected deactivate '2026-12-31', got %v", pvm["deactivate"])
	}
}

func TestLnlBadge_toJSON_shouldPutNilForAbsentDates(t *testing.T) {
	props := map[string]any{
		"ID":       float64(3),
		"BADGEKEY": float64(42),
		"STATUS":   float64(1),
		"TYPE":     float64(1),
	}
	badge, err := NewLnlBadge(props, newTestCache())
	if err != nil {
		t.Fatal(err)
	}
	j := badge.ToJSON()

	pvm, ok := j["property_value_map"].(map[string]any)
	if !ok {
		t.Fatal("expected property_value_map to be a map")
	}
	if pvm["activate"] != nil {
		t.Errorf("expected activate nil, got %v", pvm["activate"])
	}
	if pvm["deactivate"] != nil {
		t.Errorf("expected deactivate nil, got %v", pvm["deactivate"])
	}
}

// ---- LnlAccessLevel tests ----

func TestLnlAccessLevel_fromProps_shouldParseId(t *testing.T) {
	props := map[string]any{
		"ID":   float64(5),
		"Name": "Main Entrance",
	}
	al, err := NewLnlAccessLevel(props)
	if err != nil {
		t.Fatal(err)
	}
	if al.ID != 5 {
		t.Errorf("expected ID 5, got %d", al.ID)
	}
}

func TestLnlAccessLevel_fromProps_shouldParseNameAndID(t *testing.T) {
	props := map[string]any{
		"ID":   float64(10),
		"Name": "Conference Room",
	}
	al, err := NewLnlAccessLevel(props)
	if err != nil {
		t.Fatal(err)
	}
	if al.ID != 10 {
		t.Errorf("expected ID 10, got %d", al.ID)
	}
	if al.Name != "Conference Room" {
		t.Errorf("expected Name 'Conference Room', got %q", al.Name)
	}
}


// ---- propDate / dateStr symmetry ----

func TestPropDate_shouldParseISODate(t *testing.T) {
	props := map[string]any{"ACTIVATE": "2018-09-12"}
	d := propDate(props, "ACTIVATE")
	if d == nil {
		t.Fatal("expected non-nil date")
	}
	want := time.Date(2018, 9, 12, 0, 0, 0, 0, time.UTC)
	if !d.Equal(want) {
		t.Errorf("expected %v, got %v", want, *d)
	}
}

func TestPropDate_shouldReturnNilForMissingKey(t *testing.T) {
	props := map[string]any{}
	d := propDate(props, "ACTIVATE")
	if d != nil {
		t.Errorf("expected nil, got %v", d)
	}
}
