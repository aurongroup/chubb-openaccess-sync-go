package csv

import (
	"log"
	"openaccess-sync/data/model"
)

// AccessRecordCache holds the computed access records derived from a DataCache.
type AccessRecordCache struct {
	records                []*model.AccessRecord
	recordsByCardholderKey map[string][]*model.AccessRecord
}

func BuildAccessRecordCache(c model.IDCache) *AccessRecordCache {
	arc := AccessRecordCache{
		recordsByCardholderKey: make(map[string][]*model.AccessRecord),
	}

	// TODO
	for _, badge := range c.GetBadges() {
		r, err := badge.ToAccessRecord(c.GetAccessLevelsByBadge(badge.ID))
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

func (c *AccessRecordCache) Records() []*model.AccessRecord {
	return c.records
}

func (c *AccessRecordCache) RecordsByCardholderKey(key string) []*model.AccessRecord {
	return c.recordsByCardholderKey[key]
}
