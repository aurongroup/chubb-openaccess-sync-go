package model

import (
	"openaccess-sync/data"
	"openaccess-sync/data/model/csv"
	"openaccess-sync/data/model/lenel"
	"testing"
	"time"
)

// testCache is a minimal Cache implementation for model tests.
type testCache struct {
	statuses     map[int]*lenel.BadgeStatus
	badgeTypes   map[int]*lenel.BadgeType
	cardholders  map[int]*lenel.Cardholder
	accessLevels map[int]*lenel.AccessLevel
	badgeByKey   map[int]*lenel.Badge
}

func (c *testCache) GetBadgeStatus(id int) *lenel.BadgeStatus { return c.statuses[id] }
func (c *testCache) GetBadgeType(id int) *lenel.BadgeType     { return c.badgeTypes[id] }
func (c *testCache) GetCardholder(id int) *lenel.Cardholder   { return c.cardholders[id] }
func (c *testCache) GetAccessLevel(id int) *lenel.AccessLevel { return c.accessLevels[id] }
func (c *testCache) GetBadgeByKey(key int) *lenel.Badge       { return c.badgeByKey[key] }

func newTestCache() *testCache {
	return &testCache{
		statuses:     map[int]*lenel.BadgeStatus{1: {ID: 1, Name: "Active"}},
		badgeTypes:   map[int]*lenel.BadgeType{1: {ID: 1, Name: "Standard"}},
		cardholders:  map[int]*lenel.Cardholder{},
		accessLevels: map[int]*lenel.AccessLevel{},
		badgeByKey:   map[int]*lenel.Badge{},
	}
}

func newAssignmentCache() *testCache {
	c := newTestCache()
	c.accessLevels[10] = &lenel.AccessLevel{ID: 10, Name: "Main Entrance"}
	b := &lenel.Badge{ID: 20, BadgeKey: 200, Status: c.statuses[1], Type: c.badgeTypes[1]}
	c.badgeByKey[200] = b
	return c
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

// ---- LnlBadge tests ----

func TestLnlBadge_fromProps_shouldParseId(t *testing.T) {
	props := map[string]any{
		"ID":       float64(42),
		"BADGEKEY": float64(1),
		"STATUS":   float64(1),
		"TYPE":     float64(1),
	}

	badge, err := lenel.NewBadge(props, newTestCache())
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

	_, err := lenel.NewBadge(props, newTestCache())
	if err != data.ErrBadgeMissingID {
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

	badge, err := lenel.NewBadge(props, newTestCache())
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

	badge, err := lenel.NewBadge(props, newTestCache())
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

	al, err := lenel.NewAccessLevel(props)
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

	al, err := lenel.NewAccessLevel(props)
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

	d := data.propDate(props, "ACTIVATE")
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

	d := data.propDate(props, "ACTIVATE")
	if d != nil {
		t.Errorf("expected nil, got %v", d)
	}
}

// ---- NewLnlBadgeStatus ----

func TestNewLnlBadgeStatus_shouldParseIdAndName(t *testing.T) {
	s, err := lenel.NewBadgeStatus(map[string]any{"ID": float64(3), "Name": "Active"})
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != 3 {
		t.Errorf("expected ID 3, got %d", s.ID)
	}
	if s.Name != "Active" {
		t.Errorf("expected Name %q, got %q", "Active", s.Name)
	}
}

func TestNewLnlBadgeStatus_shouldErrorWhenIdAbsent(t *testing.T) {
	_, err := lenel.NewBadgeStatus(map[string]any{"Name": "Active"})
	if err != data.ErrBadgeStatusMissingID {
		t.Errorf("expected ErrBadgeStatusMissingID, got %v", err)
	}
}

func TestNewLnlBadgeStatus_shouldErrorWhenNameAbsent(t *testing.T) {
	_, err := lenel.NewBadgeStatus(map[string]any{"ID": float64(1)})
	if err != data.ErrBadgeStatusMissingName {
		t.Errorf("expected ErrBadgeStatusMissingName, got %v", err)
	}
}

// ---- NewLnlBadgeType ----

func TestNewLnlBadgeType_shouldParseIdAndName(t *testing.T) {
	bt, err := lenel.NewBadgeType(map[string]any{"ID": float64(2), "Name": "Employee"})
	if err != nil {
		t.Fatal(err)
	}
	if bt.ID != 2 {
		t.Errorf("expected ID 2, got %d", bt.ID)
	}
	if bt.Name != "Employee" {
		t.Errorf("expected Name %q, got %q", "Employee", bt.Name)
	}
}

func TestNewLnlBadgeType_shouldErrorWhenIdAbsent(t *testing.T) {
	_, err := lenel.NewBadgeType(map[string]any{"Name": "Employee"})
	if err != data.ErrBadgeTypeMissingID {
		t.Errorf("expected ErrBadgeTypeMissingID, got %v", err)
	}
}

func TestNewLnlBadgeType_shouldErrorWhenNameAbsent(t *testing.T) {
	_, err := lenel.NewBadgeType(map[string]any{"ID": float64(2)})
	if err != data.ErrBadgeTypeMissingName {
		t.Errorf("expected ErrBadgeTypeMissingName, got %v", err)
	}
}

// ---- NewLnlCardholder ----

func TestNewLnlCardholder_shouldParseAllFields(t *testing.T) {
	ch, err := lenel.NewCardholder(map[string]any{
		"ID":        float64(10),
		"SSNO":      "1234",
		"FIRSTNAME": "Bob",
		"LASTNAME":  "Brown",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != 10 {
		t.Errorf("expected ID 10, got %d", ch.ID)
	}
	if ch.SSNO != "1234" {
		t.Errorf("expected SSNO %q, got %q", "1234", ch.SSNO)
	}
	if ch.FirstName != "Bob" {
		t.Errorf("expected FirstName %q, got %q", "Bob", ch.FirstName)
	}
	if ch.LastName != "Brown" {
		t.Errorf("expected LastName %q, got %q", "Brown", ch.LastName)
	}
}

func TestNewLnlCardholder_shouldErrorWhenNeitherIdNorSsno(t *testing.T) {
	_, err := lenel.NewCardholder(map[string]any{"LASTNAME": "Brown"})
	if err != data.ErrCardholderMissingIdentifier {
		t.Errorf("expected ErrCardholderMissingIdentifier, got %v", err)
	}
}

func TestNewLnlCardholder_shouldAcceptIdWithoutSsno(t *testing.T) {
	_, err := lenel.NewCardholder(map[string]any{"ID": float64(5), "LASTNAME": "Brown"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestNewLnlCardholder_shouldAcceptSsnoWithoutId(t *testing.T) {
	_, err := lenel.NewCardholder(map[string]any{"SSNO": "9999", "LASTNAME": "Brown"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestNewLnlCardholder_shouldErrorWhenLastNameAbsent(t *testing.T) {
	_, err := lenel.NewCardholder(map[string]any{"ID": float64(5)})
	if err != data.ErrCardholderMissingLastName {
		t.Errorf("expected ErrCardholderMissingLastName, got %v", err)
	}
}

// ---- NewLnlAccessLevelAssignment ----

func TestNewLnlAccessLevelAssignment_shouldResolveAccessLevelAndBadge(t *testing.T) {
	props := map[string]any{"AccessLevelID": float64(10), "BadgeKey": float64(200)}
	a, err := lenel.NewAccessLevelAssignment(props, newAssignmentCache())
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

func TestNewLnlAccessLevelAssignment_shouldErrorWhenCacheNil(t *testing.T) {
	_, err := lenel.NewAccessLevelAssignment(map[string]any{}, nil)
	if err != data.ErrAssignmentNilCache {
		t.Errorf("expected ErrAssignmentNilCache, got %v", err)
	}
}

func TestNewLnlAccessLevelAssignment_shouldErrorWhenAccessLevelNotFound(t *testing.T) {
	props := map[string]any{"AccessLevelID": float64(999), "BadgeKey": float64(200)}
	_, err := lenel.NewAccessLevelAssignment(props, newAssignmentCache())
	if err != data.ErrAssignmentUnresolvedAccessLevel {
		t.Errorf("expected ErrAssignmentUnresolvedAccessLevel, got %v", err)
	}
}

func TestNewLnlAccessLevelAssignment_shouldErrorWhenBadgeNotFound(t *testing.T) {
	props := map[string]any{"AccessLevelID": float64(10), "BadgeKey": float64(999)}
	_, err := lenel.NewAccessLevelAssignment(props, newAssignmentCache())
	if err != data.ErrAssignmentUnresolvedBadge {
		t.Errorf("expected ErrAssignmentUnresolvedBadge, got %v", err)
	}
}

// ---- NewLnlBadge error cases ----

func TestNewLnlBadge_shouldErrorWhenCacheNil(t *testing.T) {
	props := map[string]any{"ID": float64(1), "BADGEKEY": float64(1)}
	_, err := lenel.NewBadge(props, nil)
	if err != data.ErrBadgeNilCache {
		t.Errorf("expected ErrBadgeNilCache, got %v", err)
	}
}

func TestNewLnlBadge_shouldErrorWhenBadgeKeyAbsent(t *testing.T) {
	props := map[string]any{"ID": float64(1)}
	_, err := lenel.NewBadge(props, newTestCache())
	if err != data.ErrBadgeMissingBadgeKey {
		t.Errorf("expected ErrBadgeMissingBadgeKey, got %v", err)
	}
}

func TestNewLnlBadge_shouldErrorWhenStatusNotInCache(t *testing.T) {
	props := map[string]any{"ID": float64(1), "BADGEKEY": float64(1), "STATUS": float64(999), "TYPE": float64(1)}
	_, err := lenel.NewBadge(props, newTestCache())
	if err != data.ErrBadgeUnresolvedStatus {
		t.Errorf("expected ErrBadgeUnresolvedStatus, got %v", err)
	}
}

func TestNewLnlBadge_shouldErrorWhenTypeNotInCache(t *testing.T) {
	props := map[string]any{"ID": float64(1), "BADGEKEY": float64(1), "STATUS": float64(1), "TYPE": float64(999)}
	_, err := lenel.NewBadge(props, newTestCache())
	if err != data.ErrBadgeUnresolvedType {
		t.Errorf("expected ErrBadgeUnresolvedType, got %v", err)
	}
}

// ---- LnlBadge.ToAccessRecord ----

func TestLnlBadge_ToAccessRecord_shouldMapAllFields(t *testing.T) {
	c := newTestCache()
	activate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	deactivate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	badge := &lenel.Badge{
		ID:         42,
		BadgeKey:   100,
		Activate:   &activate,
		Deactivate: &deactivate,
		Status:     c.statuses[1],
		Type:       c.badgeTypes[1],
		Cardholder: &lenel.Cardholder{ID: 1, FirstName: "Bob", LastName: "Brown", SSNO: "1234"},
	}
	levels := []*lenel.AccessLevel{
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

func TestLnlBadge_ToAccessRecord_shouldErrorWhenCardholderNil(t *testing.T) {
	c := newTestCache()
	badge := &lenel.Badge{ID: 5, BadgeKey: 50, Status: c.statuses[1], Type: c.badgeTypes[1]}
	_, err := badge.ToAccessRecord(nil)
	if err != data.ErrAccessRecordMissingLast {
		t.Errorf("expected ErrAccessRecordMissingLast for nil cardholder, got %v", err)
	}
}

func TestLnlBadge_ToAccessRecord_shouldCapAccessLevelsAtSix(t *testing.T) {
	c := newTestCache()
	badge := &lenel.Badge{
		ID:         7,
		BadgeKey:   70,
		Status:     c.statuses[1],
		Type:       c.badgeTypes[1],
		Cardholder: &lenel.Cardholder{ID: 1, LastName: "Smith"},
	}
	levels := make([]*lenel.AccessLevel, 7)
	for i := range levels {
		levels[i] = &lenel.AccessLevel{ID: i + 1, Name: "Level"}
	}
	r, err := badge.ToAccessRecord(levels)
	if err != nil {
		t.Fatal(err)
	}
	assertStr(t, "AccLvl6", "Level", r.AccLvl6)
}

// ---- NewAccessRecord validation ----

func TestNewAccessRecord_shouldErrorWhenLastMissing(t *testing.T) {
	_, err := csv.NewAccessRecord("", "", "", "", "", "", "", "", "", "100", nil, nil, "active", "Employee")
	if err != data.ErrAccessRecordMissingLast {
		t.Errorf("expected ErrAccessRecordMissingLast, got %v", err)
	}
}

func TestNewAccessRecord_shouldErrorWhenBadgeIDMissing(t *testing.T) {
	_, err := csv.NewAccessRecord("", "", "Smith", "", "", "", "", "", "", "", nil, nil, "active", "Employee")
	if err != data.ErrAccessRecordMissingBadgeID {
		t.Errorf("expected ErrAccessRecordMissingBadgeID, got %v", err)
	}
}

func TestNewAccessRecord_shouldErrorWhenStatusMissing(t *testing.T) {
	_, err := csv.NewAccessRecord("", "", "Smith", "", "", "", "", "", "", "100", nil, nil, "", "Employee")
	if err != data.ErrAccessRecordMissingStatus {
		t.Errorf("expected ErrAccessRecordMissingStatus, got %v", err)
	}
}

func TestNewAccessRecord_shouldErrorWhenBadgeTypeMissing(t *testing.T) {
	_, err := csv.NewAccessRecord("", "", "Smith", "", "", "", "", "", "", "100", nil, nil, "active", "")
	if err != data.ErrAccessRecordMissingBadgeType {
		t.Errorf("expected ErrAccessRecordMissingBadgeType, got %v", err)
	}
}

// ---- AccessRecord.ToRow ----

func TestAccessRecord_ToRow_shouldReturnAllFieldsInOrder(t *testing.T) {
	activate := time.Date(2018, 9, 12, 0, 0, 0, 0, time.UTC)
	deactivate := time.Date(2020, 9, 12, 0, 0, 0, 0, time.UTC)
	r, err := csv.NewAccessRecord("8274", "BOB", "BROWN", "L1", "L2", "L3", "L4", "L5", "L6", "9017", &activate, &deactivate, "active", "Employee")
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
	r, err := csv.NewAccessRecord("", "", "Smith", "", "", "", "", "", "", "1", &d, nil, "active", "Employee")
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
