package lenel

import (
	"fmt"
	"log"
	"openaccess-sync/pkg/client"
	model2 "openaccess-sync/pkg/data/model"
)

type BadgeCache struct {
	list  []*model2.Badge
	byID  map[int32]*model2.Badge
	byKey map[int64]*model2.Badge
}

func NewBadgeCache() BadgeCache {
	return BadgeCache{
		byID:  make(map[int32]*model2.Badge),
		byKey: make(map[int64]*model2.Badge),
	}
}

func (c *BadgeCache) GetItems() []*model2.Badge {
	return c.list
}

func (c *BadgeCache) FillActiveForCardholder(cl *client.Client, ch *model2.Cardholder) error {
	items, err := cl.GetInstancesWithProgress("Lnl_Badge", fmt.Sprintf("PERSONID=%d AND STATUS=1", ch.ID))
	if err != nil {
		return err
	}

	for _, props := range items {
		b, err := model2.NewBadgeFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_Badge: %v", err)
			continue
		}

		c.list = append(c.list, b)
		c.byID[b.Key] = b
		c.byKey[b.ID] = b
	}
	log.Printf("Retrieved %d Lnl_Badge records", len(c.list))
	return nil
}

// FIXME
//func (c *BadgeCache) Fill(cl *client.Client, cache model.IDCache) error {
//	items, err := cl.GetInstancesWithProgress("Lnl_Badge", "")
//	if err != nil {
//		return err
//	}
//
//	for _, props := range items {
//		b, err := model.NewBadgeFromJSON(props, cache)
//		if err != nil {
//			log.Printf("skipping Lnl_Badge: %v", err)
//			continue
//		}
//
//		c.list = append(c.list, b)
//		c.byID[b.ID] = b  // note: Badge.ID is int64, map key is int32 — needs review
//		c.byKey[b.Key] = b
//	}
//	log.Printf("Retrieved %d Lnl_Badge records", len(c.list))
//	return nil
//}
