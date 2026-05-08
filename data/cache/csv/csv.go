package csv

import (
	"log"
	"openaccess-sync/data/model/csv"
)

// AccessRecordCache holds the computed access records derived from a DataCache.
type AccessRecordCache struct {
	records                []*csv.AccessRecord
	recordsByCardholderKey map[string][]*csv.AccessRecord
}

func BuildAccessRecordCache(c *DataCache) *AccessRecordCache {
	arc := AccessRecordCache{
		recordsByCardholderKey: make(map[string][]*csv.AccessRecord),
	}

	levelsByBadgeID := c.GetAccessLevelsByBadge()
	for _, badge := range c.GetBadges() {
		r, err := badge.ToAccessRecord(levelsByBadgeID[badge.ID])
		if err != nil {
			log.Printf("skipping access record for badge %d: %v", badge.ID, err)
			continue
		}

		arc.records = append(arc.records, r)

		if r.CardholderKey != "" {
			arc.recordsByCardholderKey[r.CardholderKey] = append(arc.recordsByCardholderKey[r.CardholderKey], r)
		}
	}

	return &arc
}

func (c *AccessRecordCache) Records() []*csv.AccessRecord {
	return c.records
}

func (c *AccessRecordCache) RecordsByCardholderKey(key string) []*csv.AccessRecord {
	return c.recordsByCardholderKey[key]
}
