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
	badgesByKey  map[int]*Badge
}

func (c *testIDCache) GetBadgeStatus(id int) *BadgeStatus          { return c.statuses[id] }
func (c *testIDCache) GetBadgeType(id int) *BadgeType              { return c.badgeTypes[id] }
func (c *testIDCache) GetCardholder(id int) *Cardholder            { return c.cardholders[id] }
func (c *testIDCache) GetAccessLevel(id int) *AccessLevel          { return c.accessLevels[id] }
func (c *testIDCache) GetBadge(id int) *Badge                      { return c.badges[id] }
func (c *testIDCache) GetBadgeByKey(key int) *Badge                { return c.badgesByKey[key] }
func (c *testIDCache) GetBadges() []*Badge                         { return nil }
func (c *testIDCache) GetAccessLevelsByBadge(int) []*AccessLevel   { return nil }

func newTestIDCache() *testIDCache {
	return &testIDCache{
		statuses:     map[int]*BadgeStatus{1: {ID: 1, Name: "Active"}},
		badgeTypes:   map[int]*BadgeType{1: {ID: 1, Name: "Standard"}},
		cardholders:  map[int]*Cardholder{},
		accessLevels: map[int]*AccessLevel{},
		badges:       map[int]*Badge{},
		badgesByKey:  map[int]*Badge{},
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