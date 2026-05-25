package lenel

import (
	model2 "openaccess-sync/pkg/data/model"
	"strings"
	"testing"
)

// newTestCache builds a DataCache with empty sub-caches and no client,
// suitable for unit-testing methods that only access the in-memory maps.
func newTestCache() *DataCache {
	return &DataCache{
		statuses:     NewBadgeStatusCache(),
		badgeTypes:   NewBadgeTypeCache(),
		accessLevels: NewAccessLevelCache(),
	}
}

// newRecord constructs an AccessRecord with the given status, badge type, and
// up to 6 optional access level names. last and badgeID are fixed to satisfy
// validation requirements.
func newRecord(status, badgeType string, levels ...string) *model2.AccessRecord {
	lvl := make([]string, 6)
	copy(lvl, levels)
	r, err := model2.NewAccessRecord("", "", "Smith", lvl[0], lvl[1], lvl[2], lvl[3], lvl[4], lvl[5], "1", nil, nil, status, badgeType)
	if err != nil {
		panic(err)
	}
	return r
}

func TestValidateAccessRecords_shouldReturnNilForEmptyInput(t *testing.T) {
	c := newTestCache()
	if err := c.ValidateAccessRecords(nil); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidateAccessRecords_shouldReturnNilWhenAllValuesKnown(t *testing.T) {
	c := newTestCache()
	c.statuses.byName["Active"] = &model2.BadgeStatus{ID: 1, Name: "Active"}
	c.badgeTypes.byName["Employee"] = &model2.BadgeType{ID: 1, Name: "Employee"}
	c.accessLevels.byName["Floor 1"] = &model2.AccessLevel{ID: 1, Name: "Floor 1"}

	records := []*model2.AccessRecord{newRecord("Active", "Employee", "Floor 1")}
	if err := c.ValidateAccessRecords(records); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidateAccessRecords_shouldReturnErrorForUnknownStatus(t *testing.T) {
	c := newTestCache()
	c.badgeTypes.byName["Employee"] = &model2.BadgeType{ID: 1, Name: "Employee"}

	err := c.ValidateAccessRecords([]*model2.AccessRecord{newRecord("Unknown", "Employee")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown badge statuses") {
		t.Errorf("expected 'unknown badge statuses' in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "Unknown") {
		t.Errorf("expected value 'Unknown' in error, got %q", err.Error())
	}
}

func TestValidateAccessRecords_shouldReturnErrorForUnknownBadgeType(t *testing.T) {
	c := newTestCache()
	c.statuses.byName["Active"] = &model2.BadgeStatus{ID: 1, Name: "Active"}

	err := c.ValidateAccessRecords([]*model2.AccessRecord{newRecord("Active", "Contractor")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown badge types") {
		t.Errorf("expected 'unknown badge types' in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "Contractor") {
		t.Errorf("expected value 'Contractor' in error, got %q", err.Error())
	}
}

func TestValidateAccessRecords_shouldReturnErrorForUnknownAccessLevel(t *testing.T) {
	c := newTestCache()
	c.statuses.byName["Active"] = &model2.BadgeStatus{ID: 1, Name: "Active"}
	c.badgeTypes.byName["Employee"] = &model2.BadgeType{ID: 1, Name: "Employee"}

	err := c.ValidateAccessRecords([]*model2.AccessRecord{newRecord("Active", "Employee", "NoSuchLevel")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown access levels") {
		t.Errorf("expected 'unknown access levels' in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "NoSuchLevel") {
		t.Errorf("expected value 'NoSuchLevel' in error, got %q", err.Error())
	}
}

func TestValidateAccessRecords_shouldCheckAllSixAccessLevelSlots(t *testing.T) {
	c := newTestCache()
	c.statuses.byName["Active"] = &model2.BadgeStatus{ID: 1, Name: "Active"}
	c.badgeTypes.byName["Employee"] = &model2.BadgeType{ID: 1, Name: "Employee"}
	c.accessLevels.byName["Known"] = &model2.AccessLevel{ID: 1, Name: "Known"}

	err := c.ValidateAccessRecords([]*model2.AccessRecord{
		newRecord("Active", "Employee", "Known", "Known", "Known", "Known", "Known", "Missing6"),
	})
	if err == nil {
		t.Fatal("expected error for unknown level in slot 6, got nil")
	}
	if !strings.Contains(err.Error(), "Missing6") {
		t.Errorf("expected 'Missing6' in error, got %q", err.Error())
	}
}

func TestValidateAccessRecords_shouldSkipEmptyAccessLevelSlots(t *testing.T) {
	c := newTestCache()
	c.statuses.byName["Active"] = &model2.BadgeStatus{ID: 1, Name: "Active"}
	c.badgeTypes.byName["Employee"] = &model2.BadgeType{ID: 1, Name: "Employee"}

	// no levels provided — all six slots are empty strings
	if err := c.ValidateAccessRecords([]*model2.AccessRecord{newRecord("Active", "Employee")}); err != nil {
		t.Errorf("expected nil for empty access level slots, got %v", err)
	}
}

func TestValidateAccessRecords_shouldDeduplicateUnknownValues(t *testing.T) {
	c := newTestCache()
	c.badgeTypes.byName["Employee"] = &model2.BadgeType{ID: 1, Name: "Employee"}

	records := []*model2.AccessRecord{
		newRecord("Unknown", "Employee"),
		newRecord("Unknown", "Employee"),
		newRecord("Unknown", "Employee"),
	}
	err := c.ValidateAccessRecords(records)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if count := strings.Count(err.Error(), "Unknown"); count != 1 {
		t.Errorf("expected 'Unknown' to appear once, got %d occurrences in %q", count, err.Error())
	}
}

func TestValidateAccessRecords_shouldReportAllCategoriesInOneError(t *testing.T) {
	c := newTestCache() // empty — nothing resolves

	err := c.ValidateAccessRecords([]*model2.AccessRecord{
		newRecord("BadStatus", "BadType", "BadLevel"),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	for _, want := range []string{"unknown badge statuses", "unknown badge types", "unknown access levels"} {
		if !strings.Contains(msg, want) {
			t.Errorf("expected %q in error, got %q", want, msg)
		}
	}
}
