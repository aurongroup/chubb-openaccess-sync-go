package lenel

import (
	"fmt"
	"log"
	"openaccess-sync/pkg/client"
	model2 "openaccess-sync/pkg/data/model"
)

// AssignmentCache holds access level assignments and a reverse index from
// badge key to the access levels assigned to that badge.
type AssignmentCache struct {
	list          []*model2.AccessLevelAssignment
	byAccessLevel map[int32][]*model2.AccessLevelAssignment
	byBadgeKey    map[int32][]*model2.AccessLevelAssignment
}

func (c *AssignmentCache) GetItems() []*model2.AccessLevelAssignment {
	return c.list
}

func (c *AssignmentCache) FillForBadge(cl *client.Client, b *model2.Badge) error {
	items, err := cl.GetInstancesWithProgress("Lnl_AccessLevelAssignment", fmt.Sprintf("BADGEKEY=%d", b.Key))
	if err != nil {
		return err
	}

	for _, props := range items {
		b, err := model2.NewAccessLevelAssignmentFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_AccessLevelAssignment: %v", err)
			continue
		}

		c.list = append(c.list, b)

		if _, ok := c.byAccessLevel[b.AccessLevel]; !ok {
			c.byAccessLevel[b.AccessLevel] = []*model2.AccessLevelAssignment{}
		}
		c.byAccessLevel[b.AccessLevel] = append(c.byAccessLevel[b.AccessLevel], b)

		if _, ok := c.byBadgeKey[b.BadgeKey]; !ok {
			c.byBadgeKey[b.BadgeKey] = []*model2.AccessLevelAssignment{}
		}
		c.byBadgeKey[b.BadgeKey] = append(c.byBadgeKey[b.BadgeKey], b)
	}
	log.Printf("Retrieved %d Lnl_AccessLevelAssignment records", len(c.list))
	return nil
}

func NewAssignmentCache() AssignmentCache {
	return AssignmentCache{
		byAccessLevel: make(map[int32][]*model2.AccessLevelAssignment),
		byBadgeKey:    make(map[int32][]*model2.AccessLevelAssignment),
	}
}

// FIXME
//func (c *AssignmentCache) Fill(cl *client.Client, cache model.IDCache) error {
//	items, err := cl.GetInstancesWithProgress("Lnl_AccessLevelAssignment", "")
//	if err != nil {
//		return err
//	}
//
//	for _, props := range items {
//		a, err := model.NewAccessLevelAssignmentFromJSON(props, cache)
//		if err != nil {
//			log.Printf("skipping Lnl_AccessLevelAssignment: %v", err)
//			continue
//		}
//
//		c.list = append(c.list, a)
//		c.byBadgeKey[a.Badge.Key] = append(c.byBadgeKey[a.Badge.Key], a.AccessLevel)
//	}
//	log.Printf("Retrieved %d Lnl_AccessLevelAssignment records", len(c.list))
//	return nil
//}
