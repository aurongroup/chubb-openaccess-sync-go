package lenel

import "openaccess-sync/data/model"

// AssignmentCache holds access level assignments and a reverse index from
// badge key to the access levels assigned to that badge.
type AssignmentCache struct {
	list       []*model.AccessLevelAssignment
	byBadgeKey map[int64][]*model.AccessLevel
}

func NewAssignmentCache() AssignmentCache {
	return AssignmentCache{
		byBadgeKey: make(map[int64][]*model.AccessLevel),
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
