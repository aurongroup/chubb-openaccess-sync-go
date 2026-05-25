package lenel

import (
	"fmt"
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/data/model"
)

// AssignmentCache holds access level assignments and a reverse index from
// badge key to the access levels assigned to that badge.
type AssignmentCache struct {
	list          []*model.AccessLevelAssignment
	byAccessLevel map[int32][]*model.AccessLevelAssignment
	byBadgeKey    map[int32][]*model.AccessLevelAssignment
}

func (c *AssignmentCache) GetItems() []*model.AccessLevelAssignment {
	return c.list
}

func (c *AssignmentCache) GetItemsByBadgeKey(key int32) []*model.AccessLevelAssignment {
	assignments, ok := c.byBadgeKey[key]

	if ok {
		return assignments
	}

	return []*model.AccessLevelAssignment{}
}

func (c *AssignmentCache) FillForBadge(cl *client.Client, b *model.Badge) error {
	items, err := cl.GetInstancesWithProgress("Lnl_AccessLevelAssignment", fmt.Sprintf("BADGEKEY=%d", b.Key))
	if err != nil {
		return err
	}

	for _, props := range items {
		b, err := model.NewAccessLevelAssignmentFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_AccessLevelAssignment: %v", err)
			continue
		}

		c.list = append(c.list, b)

		if _, ok := c.byAccessLevel[b.AccessLevel]; !ok {
			c.byAccessLevel[b.AccessLevel] = []*model.AccessLevelAssignment{}
		}
		c.byAccessLevel[b.AccessLevel] = append(c.byAccessLevel[b.AccessLevel], b)

		if _, ok := c.byBadgeKey[b.BadgeKey]; !ok {
			c.byBadgeKey[b.BadgeKey] = []*model.AccessLevelAssignment{}
		}
		c.byBadgeKey[b.BadgeKey] = append(c.byBadgeKey[b.BadgeKey], b)
	}
	log.Printf("Retrieved %d Lnl_AccessLevelAssignment records", len(c.list))
	return nil
}

func NewAssignmentCache() AssignmentCache {
	return AssignmentCache{
		byAccessLevel: make(map[int32][]*model.AccessLevelAssignment),
		byBadgeKey:    make(map[int32][]*model.AccessLevelAssignment),
	}
}

func (c *AssignmentCache) Fill(cl *client.Client) error {
	items, err := cl.GetInstancesWithProgress("Lnl_AccessLevelAssignment", "")
	if err != nil {
		return err
	}

	for _, props := range items {
		a, err := model.NewAccessLevelAssignmentFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_AccessLevelAssignment: %v", err)
			continue
		}

		c.list = append(c.list, a)

		if _, ok := c.byBadgeKey[a.BadgeKey]; !ok {
			c.byBadgeKey[a.BadgeKey] = []*model.AccessLevelAssignment{}
		}
		c.byBadgeKey[a.BadgeKey] = append(c.byBadgeKey[a.BadgeKey], a)
	}
	log.Printf("Retrieved %d Lnl_AccessLevelAssignment records", len(c.list))
	return nil
}
