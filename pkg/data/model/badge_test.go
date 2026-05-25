package model

import (
	"testing"
	"time"
)

func TestNewBadgeFromJSON_shouldParseId(t *testing.T) {
	props := map[string]any{
		"ID":       float64(42),
		"BADGEKEY": float64(1),
		"STATUS":   float64(1),
		"TYPE":     float64(1),
	}

	badge, err := NewBadgeFromJSON(props, newTestIDCache())
	if err != nil {
		t.Fatal(err)
	}

	if badge.ID != 42 {
		t.Errorf("expected ID 42, got %d", badge.ID)
	}
}

func TestNewBadgeFromJSON_shouldErrorWhenCacheNil(t *testing.T) {
	_, err := NewBadgeFromJSON(map[string]any{"ID": float64(1), "BADGEKEY": float64(1)}, nil)
	if err != ErrBadgeNilCache {
		t.Errorf("expected ErrBadgeNilCache, got %v", err)
	}
}

func TestNewBadgeFromJSON_shouldErrorWhenIdAbsent(t *testing.T) {
	props := map[string]any{"BADGEKEY": float64(1)}
	_, err := NewBadgeFromJSON(props, newTestIDCache())
	if err != ErrBadgeMissingID {
		t.Errorf("expected ErrBadgeMissingID, got %v", err)
	}
}

func TestNewBadgeFromJSON_shouldErrorWhenBadgeKeyAbsent(t *testing.T) {
	props := map[string]any{"ID": float64(1)}
	_, err := NewBadgeFromJSON(props, newTestIDCache())
	if err != ErrBadgeMissingBadgeKey {
		t.Errorf("expected ErrBadgeMissingBadgeKey, got %v", err)
	}
}

func TestNewBadgeFromJSON_shouldErrorWhenStatusNotInCache(t *testing.T) {
	props := map[string]any{"ID": float64(1), "BADGEKEY": float64(1), "STATUS": float64(999), "TYPE": float64(1)}
	_, err := NewBadgeFromJSON(props, newTestIDCache())
	if err != ErrBadgeUnresolvedStatus {
		t.Errorf("expected ErrBadgeUnresolvedStatus, got %v", err)
	}
}

func TestNewBadgeFromJSON_shouldErrorWhenTypeNotInCache(t *testing.T) {
	props := map[string]any{"ID": float64(1), "BADGEKEY": float64(1), "STATUS": float64(1), "TYPE": float64(999)}
	_, err := NewBadgeFromJSON(props, newTestIDCache())
	if err != ErrBadgeUnresolvedType {
		t.Errorf("expected ErrBadgeUnresolvedType, got %v", err)
	}
}

func TestBadge_ToJSON_shouldCreateCorrectJsonStructure(t *testing.T) {
	props := map[string]any{
		"ID":         float64(7),
		"BADGEKEY":   float64(1001),
		"STATUS":     float64(1),
		"TYPE":       float64(1),
		"ACTIVATE":   "2025-01-01",
		"DEACTIVATE": "2026-12-31",
	}

	badge, err := NewBadgeFromJSON(props, newTestIDCache())
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

func TestBadge_ToJSON_shouldPutNilForAbsentDates(t *testing.T) {
	props := map[string]any{
		"ID":       float64(3),
		"BADGEKEY": float64(42),
		"STATUS":   float64(1),
		"TYPE":     float64(1),
	}

	badge, err := NewBadgeFromJSON(props, newTestIDCache())
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

func TestBadge_ToAccessRecord_shouldMapAllFields(t *testing.T) {
	c := newTestIDCache()
	activate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	deactivate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	badge := &Badge{
		ID:         42,
		Key:        100,
		Activate:   &activate,
		Deactivate: &deactivate,
		Status:     c.statuses[1],
		Type:       c.badgeTypes[1],
		Cardholder: &Cardholder{ID: 1, FirstName: "Bob", LastName: "Brown", SSNO: "1234"},
	}
	levels := []*AccessLevel{
		{ID: 1, Name: "Main"},
		{ID: 2, Name: "Side"},
	}
	r, err := badge.ToAccessRecord(levels)
	if err != nil {
		t.Fatal(err)
	}
	assertStr(t, "SSNO", "1234", r.SSNO)
	assertStr(t, "First", "Bob", r.First)
	assertStr(t, "Last", "Brown", r.Last)
	assertStr(t, "AccLvl1", "Main", r.AccLvl1)
	assertStr(t, "AccLvl2", "Side", r.AccLvl2)
	assertStr(t, "BadgeID", "42", r.BadgeID)
	assertStr(t, "Status", "Active", r.Status)
	assertStr(t, "BadgeType", "Standard", r.BadgeType)
	assertDate(t, "Activate", activate, r.Activate)
	assertDate(t, "Deactivate", deactivate, r.Deactivate)
}

func TestBadge_ToAccessRecord_shouldErrorWhenCardholderNil(t *testing.T) {
	c := newTestIDCache()
	badge := &Badge{ID: 5, Key: 50, Status: c.statuses[1], Type: c.badgeTypes[1]}
	_, err := badge.ToAccessRecord(nil)
	if err != ErrAccessRecordMissingLast {
		t.Errorf("expected ErrAccessRecordMissingLast for nil cardholder, got %v", err)
	}
}

func TestBadge_ToAccessRecord_shouldCapAccessLevelsAtSix(t *testing.T) {
	c := newTestIDCache()
	badge := &Badge{
		ID:         7,
		Key:        70,
		Status:     c.statuses[1],
		Type:       c.badgeTypes[1],
		Cardholder: &Cardholder{ID: 1, LastName: "Smith"},
	}
	levels := make([]*AccessLevel, 7)
	for i := range levels {
		levels[i] = &AccessLevel{ID: i + 1, Name: "Level"}
	}
	r, err := badge.ToAccessRecord(levels)
	if err != nil {
		t.Fatal(err)
	}
	assertStr(t, "AccLvl6", "Level", r.AccLvl6)
}
