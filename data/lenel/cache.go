package lenel

import (
	"log"
	"openaccess-sync/client"
	"openaccess-sync/data/model"
)

// DataCache holds all data fetched from the OpenAccess API.
type DataCache struct {
	client *client.Client

	// indexed maps for O(1) lookup
	accessLevels           map[int32]*model.AccessLevel
	accessLevelsByBadgeKey map[int64][]*model.AccessLevel
	badges                 map[int32]*model.Badge
	badgeByKey             map[int64]*model.Badge
	statuses               map[int32]*model.BadgeStatus
	badgeTypes             map[int32]*model.BadgeType
	cardholders            map[int32]*model.Cardholder

	// ordered slices for iteration (matches Java insertion order)
	accessLevelList []*model.AccessLevel
	badgeList       []*model.Badge
	cardholderList  []*model.Cardholder
	badgeStatusList []*model.BadgeStatus
	badgeTypeList   []*model.BadgeType
	assignments     []*model.AccessLevelAssignment
}

// NewDataCache constructs an empty DataCache backed by the given client.
func NewDataCache(client *client.Client) *DataCache {
	return &DataCache{
		client:                 client,
		accessLevels:           make(map[int32]*model.AccessLevel),
		accessLevelsByBadgeKey: make(map[int64][]*model.AccessLevel),
		badges:                 make(map[int32]*model.Badge),
		badgeByKey:             make(map[int64]*model.Badge),
		statuses:               make(map[int32]*model.BadgeStatus),
		badgeTypes:             make(map[int32]*model.BadgeType),
		cardholders:            make(map[int32]*model.Cardholder),
	}
}

func (c *DataCache) GetAccessLevel(id int32) *model.AccessLevel {
	return c.accessLevels[id]
}

func (c *DataCache) GetAccessLevels() []*model.AccessLevel {
	return c.accessLevelList
}

func (c *DataCache) GetBadge(id int32) *model.Badge {
	return c.badges[id]
}

func (c *DataCache) GetBadges() []*model.Badge {
	return c.badgeList
}

func (c *DataCache) GetBadgeByKey(key int64) *model.Badge {
	return c.badgeByKey[key]
}

func (c *DataCache) GetBadgeStatus(id int32) *model.BadgeStatus {
	return c.statuses[id]
}

func (c *DataCache) GetBadgeStatuses() []*model.BadgeStatus {
	return c.badgeStatusList
}

func (c *DataCache) GetBadgeType(id int32) *model.BadgeType {
	return c.badgeTypes[id]
}

func (c *DataCache) GetBadgeTypes() []*model.BadgeType {
	return c.badgeTypeList
}

func (c *DataCache) GetCardholder(id int32) *model.Cardholder {
	return c.cardholders[id]
}

func (c *DataCache) GetCardholders() []*model.Cardholder {
	return c.cardholderList
}

// Fill fetches all data from the API in the required order.
// References (status, type, cardholder on badges; badge and access level on
// assignments) are resolved against already-populated maps.
func (c *DataCache) Fill() error {
	if err := c.fillAccessLevels(); err != nil {
		return err
	}

	if err := c.fillBadgeStatuses(); err != nil {
		return err
	}

	if err := c.fillBadgeTypes(); err != nil {
		return err
	}

	if err := c.fillCardholders(); err != nil {
		return err
	}

	// FIXME
	//if err := c.fillBadges(); err != nil {
	//	return err
	//}
	//
	//if err := c.fillAssignments(); err != nil {
	//	return err
	//}

	return nil
}

func (c *DataCache) fillAccessLevels() error {
	items, err := c.client.GetInstancesWithProgress("Lnl_AccessLevel", "")
	if err != nil {
		return err
	}

	for _, props := range items {
		al, err := model.NewAccessLevelFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_AccessLevel: %v", err)
			continue
		}

		c.accessLevels[al.ID] = al
		c.accessLevelList = append(c.accessLevelList, al)
	}

	log.Printf("Retrieved %d Lnl_AccessLevel records", len(c.accessLevelList))
	return nil
}

func (c *DataCache) fillBadgeStatuses() error {
	items, err := c.client.GetInstancesWithProgress("Lnl_BadgeStatus", "")
	if err != nil {
		return err
	}

	for _, props := range items {
		s, err := model.NewBadgeStatusFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_BadgeStatus: %v", err)
			continue
		}

		c.statuses[s.ID] = s
		c.badgeStatusList = append(c.badgeStatusList, s)
	}

	log.Printf("Retrieved %d Lnl_BadgeStatus records", len(c.badgeStatusList))
	return nil
}

func (c *DataCache) fillBadgeTypes() error {
	items, err := c.client.GetInstancesWithProgress("Lnl_BadgeType", "")
	if err != nil {
		return err
	}

	for _, props := range items {
		t, err := model.NewBadgeTypeFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_BadgeType: %v", err)
			continue
		}

		c.badgeTypes[t.ID] = t
		c.badgeTypeList = append(c.badgeTypeList, t)
	}

	log.Printf("Retrieved %d Lnl_BadgeType records", len(c.badgeTypeList))
	return nil
}

func (c *DataCache) fillCardholders() error {
	items, err := c.client.GetInstancesWithProgress("Lnl_Cardholder", "")
	if err != nil {
		return err
	}

	for _, props := range items {
		ch, err := model.NewCardholderFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_Cardholder: %v", err)
			continue
		}

		c.cardholders[ch.ID] = ch
		c.cardholderList = append(c.cardholderList, ch)
	}

	log.Printf("Retrieved %d Lnl_Cardholder records", len(c.cardholderList))
	return nil
}

// FIXME
//func (c *DataCache) fillBadges() error {
//	items, err := c.client.GetInstancesWithProgress("Lnl_Badge", "")
//	if err != nil {
//		return err
//	}
//
//	for _, props := range items {
//		b, err := model.NewBadgeFromJSON(props, c)
//		if err != nil {
//			log.Printf("skipping Lnl_Badge: %v", err)
//			continue
//		}
//
//		c.badges[b.ID] = b
//		c.badgeByKey[b.Key] = b
//		c.badgeList = append(c.badgeList, b)
//	}
//	log.Printf("Retrieved %d Lnl_Badge records", len(c.badgeList))
//	return nil
//}

// FIXME
//func (c *DataCache) fillAssignments() error {
//	items, err := c.client.GetInstancesWithProgress("Lnl_AccessLevelAssignment", "")
//	if err != nil {
//		return err
//	}
//
//	for _, props := range items {
//		a, err := model.NewAccessLevelAssignmentFromJSON(props, c)
//		if err != nil {
//			log.Printf("skipping Lnl_AccessLevelAssignment: %v", err)
//			continue
//		}
//
//		c.assignments = append(c.assignments, a)
//
//		if _, ok := c.accessLevelsByBadge[a.Badge.ID]; !ok {
//			c.accessLevelsByBadge[a.Badge.ID] = make([]*model.AccessLevel, 0)
//		}
//		c.accessLevelsByBadge[a.Badge.ID] = append(c.accessLevelsByBadge[a.Badge.ID], a.AccessLevel)
//	}
//	log.Printf("Retrieved %d Lnl_AccessLevelAssignment records", len(c.assignments))
//	return nil
//}

func (c *DataCache) GetAccessLevelsByBadge(badgeID int64) []*model.AccessLevel {
	a := make([]*model.AccessLevel, 0)

	r, ok := c.accessLevelsByBadgeKey[badgeID]
	if ok {
		a = append(a, r...)
	}

	return a
}
