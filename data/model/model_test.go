package model

import (
	"testing"
	"time"
)

// testIDCache implements IDCache for model tests.
type testIDCache struct {
	statuses     map[int]*BadgeStatus
	badgeTypes   map[int]*BadgeType
	cardholders  map[int]*Cardholder
	accessLevels map[int]*AccessLevel
	badges       map[int]*Badge
}

func (c *testIDCache) GetBadgeStatus(id int) *BadgeStatus           { return c.statuses[id] }
func (c *testIDCache) GetBadgeType(id int) *BadgeType               { return c.badgeTypes[id] }
func (c *testIDCache) GetCardholder(id int) *Cardholder             { return c.cardholders[id] }
func (c *testIDCache) GetAccessLevel(id int) *AccessLevel           { return c.accessLevels[id] }
func (c *testIDCache) GetBadge(id int) *Badge                       { return c.badges[id] }
func (c *testIDCache) GetBadges() []*Badge                          { return nil }
func (c *testIDCache) GetAccessLevelsByBadge(int) []*AccessLevel    { return nil }

func newTestIDCache() *testIDCache {
	return &testIDCache{
		statuses:     map[int]*BadgeStatus{1: {ID: 1, Name: "Active"}},
		badgeTypes:   map[int]*BadgeType{1: {ID: 1, Name: "Standard"}},
		cardholders:  map[int]*Cardholder{},
		accessLevels: map[int]*AccessLevel{},
		badges:       map[int]*Badge{},
	}
}

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

// ---- NewAccessLevel ----

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

// ---- NewBadgeStatus ----

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

// ---- NewBadgeType ----

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

// ---- NewCardholder ----

func TestNewCardholder_shouldSetAllFields(t *testing.T) {
	ch, err := NewCardholder(10, "1234", "Bob", "Brown")
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != 10 {
		t.Errorf("ID: expected 10, got %d", ch.ID)
	}
	if ch.SSNO != "1234" {
		t.Errorf("SSNO: expected %q, got %q", "1234", ch.SSNO)
	}
	if ch.FirstName != "Bob" {
		t.Errorf("FirstName: expected %q, got %q", "Bob", ch.FirstName)
	}
	if ch.LastName != "Brown" {
		t.Errorf("LastName: expected %q, got %q", "Brown", ch.LastName)
	}
}

func TestNewCardholder_shouldAllowZeroIDAndEmptySsno(t *testing.T) {
	ch, err := NewCardholder(0, "", "Jane", "Doe")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ch.ID != 0 {
		t.Errorf("ID: expected 0, got %d", ch.ID)
	}
	if ch.SSNO != "" {
		t.Errorf("SSNO: expected empty, got %q", ch.SSNO)
	}
}

func TestNewCardholder_shouldErrorWhenLastNameMissing(t *testing.T) {
	_, err := NewCardholder(5, "9999", "Alice", "")
	if err != ErrCardholderMissingLastName {
		t.Errorf("expected ErrCardholderMissingLastName, got %v", err)
	}
}

// ---- NewCardholderFromJSON ----

func TestNewCardholderFromJSON_shouldParseAllFields(t *testing.T) {
	ch, err := NewCardholderFromJSON(map[string]any{
		"ID":        float64(10),
		"SSNO":      "1234",
		"FIRSTNAME": "Bob",
		"LASTNAME":  "Brown",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != 10 {
		t.Errorf("ID: expected 10, got %d", ch.ID)
	}
	if ch.SSNO != "1234" {
		t.Errorf("SSNO: expected %q, got %q", "1234", ch.SSNO)
	}
	if ch.FirstName != "Bob" {
		t.Errorf("FirstName: expected %q, got %q", "Bob", ch.FirstName)
	}
	if ch.LastName != "Brown" {
		t.Errorf("LastName: expected %q, got %q", "Brown", ch.LastName)
	}
}

func TestNewCardholderFromJSON_shouldTolerateMissingOptionalFields(t *testing.T) {
	ch, err := NewCardholderFromJSON(map[string]any{"LASTNAME": "Smith"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ch.ID != 0 {
		t.Errorf("ID: expected 0, got %d", ch.ID)
	}
	if ch.SSNO != "" {
		t.Errorf("SSNO: expected empty, got %q", ch.SSNO)
	}
	if ch.FirstName != "" {
		t.Errorf("FirstName: expected empty, got %q", ch.FirstName)
	}
}

func TestNewCardholderFromJSON_shouldErrorWhenLastNameMissing(t *testing.T) {
	_, err := NewCardholderFromJSON(map[string]any{"ID": float64(5), "SSNO": "9999"})
	if err != ErrCardholderMissingLastName {
		t.Errorf("expected ErrCardholderMissingLastName, got %v", err)
	}
}

// ---- NewBadgeFromJSON ----

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

// ---- Badge.ToJSON ----

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

// ---- Badge.ToAccessRecord ----

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

// ---- NewAccessLevelAssignment ----

func TestNewAccessLevelAssignment_shouldLinkAccessLevelAndBadge(t *testing.T) {
	al := &AccessLevel{ID: 10, Name: "Main Entrance"}
	b := &Badge{ID: 20, Key: 200}
	a, err := NewAccessLevelAssignment(al, b)
	if err != nil {
		t.Fatal(err)
	}
	if a.AccessLevel.ID != 10 {
		t.Errorf("expected AccessLevel.ID 10, got %d", a.AccessLevel.ID)
	}
	if a.Badge.ID != 20 {
		t.Errorf("expected Badge.ID 20, got %d", a.Badge.ID)
	}
}

func TestNewAccessLevelAssignment_shouldErrorWhenAccessLevelNil(t *testing.T) {
	_, err := NewAccessLevelAssignment(nil, &Badge{ID: 20, Key: 200})
	if err != ErrAssignmentNilAccessLevel {
		t.Errorf("expected ErrAssignmentNilAccessLevel, got %v", err)
	}
}

func TestNewAccessLevelAssignment_shouldErrorWhenBadgeNil(t *testing.T) {
	_, err := NewAccessLevelAssignment(&AccessLevel{ID: 10, Name: "Main"}, nil)
	if err != ErrAssignmentNilBadge {
		t.Errorf("expected ErrAssignmentNilBadge, got %v", err)
	}
}

func TestNewAccessLevelAssignmentFromJSON_shouldErrorWhenCacheNil(t *testing.T) {
	_, err := NewAccessLevelAssignmentFromJSON(map[string]any{}, nil)
	if err != ErrAssignmentNilCache {
		t.Errorf("expected ErrAssignmentNilCache, got %v", err)
	}
}

// ---- NewAccessRecord validation ----

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

// ---- AccessRecord.ToRow ----

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