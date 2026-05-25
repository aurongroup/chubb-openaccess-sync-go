package csv

import (
	model2 "openaccess-sync/pkg/data/model"
	"sort"
)

// AccessRecordCache holds access records and indexes of unique values seen
// across all records.
type AccessRecordCache struct {
	records                []*model2.AccessRecord
	recordsByCardholderKey map[string][]*model2.AccessRecord
	badgeTypeNames         map[string]struct{}
	badgeStatusNames       map[string]struct{}
	accessLevelNames       map[string]struct{}
	byBadgeID              map[string]*model2.AccessRecord
}

// NewAccessRecordCache constructs an empty AccessRecordCache.
func NewAccessRecordCache() *AccessRecordCache {
	return &AccessRecordCache{
		recordsByCardholderKey: make(map[string][]*model2.AccessRecord),
		badgeTypeNames:         make(map[string]struct{}),
		badgeStatusNames:       make(map[string]struct{}),
		accessLevelNames:       make(map[string]struct{}),
		byBadgeID:              make(map[string]*model2.AccessRecord),
	}
}

// Add appends r to the cache and updates all indexes.
func (c *AccessRecordCache) Add(r *model2.AccessRecord) {
	c.records = append(c.records, r)

	if r.CardholderKey != "" {
		c.recordsByCardholderKey[r.CardholderKey] = append(c.recordsByCardholderKey[r.CardholderKey], r)
	}

	c.badgeTypeNames[r.BadgeType] = struct{}{}
	c.badgeStatusNames[r.Status] = struct{}{}

	for _, lvl := range []string{r.AccLvl1, r.AccLvl2, r.AccLvl3, r.AccLvl4, r.AccLvl5, r.AccLvl6} {
		if lvl != "" {
			c.accessLevelNames[lvl] = struct{}{}
		}
	}

	if r.BadgeID != "" {
		c.byBadgeID[r.BadgeID] = r
	}
}

func BuildAccessRecordCache(c model2.IDCache) *AccessRecordCache {
	arc := NewAccessRecordCache()

	// TODO FIXME
	//for _, badge := range c.GetBadges() {
	//	r, err := badge.ToAccessRecord(c.GetAccessLevelsByBadge(badge.ID))
	//	if err != nil {
	//		log.Printf("skipping access record for badge %d: %v", badge.ID, err)
	//		continue
	//	}
	//
	//	arc.Add(r)
	//}

	return arc
}

func (c *AccessRecordCache) Records() []*model2.AccessRecord {
	return c.records
}

func (c *AccessRecordCache) RecordsByCardholderKey(key string) []*model2.AccessRecord {
	return c.recordsByCardholderKey[key]
}

func (c *AccessRecordCache) GetByBadgeID(id string) (*model2.AccessRecord, bool) {
	r, ok := c.byBadgeID[id]
	return r, ok
}

func (c *AccessRecordCache) BadgeTypeNames() []string {
	return sortedSetKeys(c.badgeTypeNames)
}

func (c *AccessRecordCache) BadgeStatusNames() []string {
	return sortedSetKeys(c.badgeStatusNames)
}

func (c *AccessRecordCache) AccessLevelNames() []string {
	return sortedSetKeys(c.accessLevelNames)
}

func sortedSetKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
