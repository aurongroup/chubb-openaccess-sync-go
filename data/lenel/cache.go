package lenel

import (
	"fmt"
	"log"
	"openaccess-sync/client"
	"openaccess-sync/data/model"
	"sort"
	"strings"
)

// DataCache holds all data fetched from the OpenAccess API.
type DataCache struct {
	client *client.Client

	// indexed maps for O(1) lookup
	accessLevels           map[int32]*model.AccessLevel
	accessLevelsByName     map[string]*model.AccessLevel
	accessLevelsByBadgeKey map[int64][]*model.AccessLevel
	badges                 map[int32]*model.Badge
	badgeByKey             map[int64]*model.Badge
	statuses               map[int32]*model.BadgeStatus
	statusesByName         map[string]*model.BadgeStatus
	badgeTypes             map[int32]*model.BadgeType
	badgeTypesByName       map[string]*model.BadgeType
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
		accessLevelsByName:     make(map[string]*model.AccessLevel),
		accessLevelsByBadgeKey: make(map[int64][]*model.AccessLevel),
		badges:                 make(map[int32]*model.Badge),
		badgeByKey:             make(map[int64]*model.Badge),
		statuses:               make(map[int32]*model.BadgeStatus),
		statusesByName:         make(map[string]*model.BadgeStatus),
		badgeTypes:             make(map[int32]*model.BadgeType),
		badgeTypesByName:       make(map[string]*model.BadgeType),
		cardholders:            make(map[int32]*model.Cardholder),
	}
}

func (c *DataCache) GetAccessLevel(id int32) *model.AccessLevel {
	return c.accessLevels[id]
}

func (c *DataCache) GetAccessLevelByName(name string) (*model.AccessLevel, bool) {
	al, ok := c.accessLevelsByName[name]
	return al, ok
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

func (c *DataCache) GetBadgeStatusByName(name string) (*model.BadgeStatus, bool) {
	s, ok := c.statusesByName[name]
	return s, ok
}

func (c *DataCache) GetBadgeStatuses() []*model.BadgeStatus {
	return c.badgeStatusList
}

func (c *DataCache) GetBadgeType(id int32) *model.BadgeType {
	return c.badgeTypes[id]
}

func (c *DataCache) GetBadgeTypeByName(name string) (*model.BadgeType, bool) {
	t, ok := c.badgeTypesByName[name]
	return t, ok
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

// ValidateAccessRecords checks that every status, badge type, and access level
// name referenced in records exists in the cache. It collects all unknown values
// and returns a single error listing them grouped by category, or nil if all
// values resolve.
func (c *DataCache) ValidateAccessRecords(records []*model.AccessRecord) error {
	unknownStatuses := make(map[string]struct{})
	unknownTypes := make(map[string]struct{})
	unknownLevels := make(map[string]struct{})

	for _, r := range records {
		if _, ok := c.statusesByName[r.Status]; !ok {
			unknownStatuses[r.Status] = struct{}{}
		}
		if _, ok := c.badgeTypesByName[r.BadgeType]; !ok {
			unknownTypes[r.BadgeType] = struct{}{}
		}
		for _, lvl := range []string{r.AccLvl1, r.AccLvl2, r.AccLvl3, r.AccLvl4, r.AccLvl5, r.AccLvl6} {
			if lvl == "" {
				continue
			}
			if _, ok := c.accessLevelsByName[lvl]; !ok {
				unknownLevels[lvl] = struct{}{}
			}
		}
	}

	var parts []string
	if len(unknownStatuses) > 0 {
		parts = append(parts, fmt.Sprintf("unknown badge statuses: %s", sortedKeys(unknownStatuses)))
	}
	if len(unknownTypes) > 0 {
		parts = append(parts, fmt.Sprintf("unknown badge types: %s", sortedKeys(unknownTypes)))
	}
	if len(unknownLevels) > 0 {
		parts = append(parts, fmt.Sprintf("unknown access levels: %s", sortedKeys(unknownLevels)))
	}
	if len(parts) > 0 {
		return fmt.Errorf("%s", strings.Join(parts, "; "))
	}
	return nil
}

func sortedKeys(m map[string]struct{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return "[" + strings.Join(keys, " ") + "]"
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
	list, byID, byName, err := fetchAndIndex(c.client, "Lnl_AccessLevel",
		model.NewAccessLevelFromJSON,
		func(al *model.AccessLevel) int32 { return al.ID },
		func(al *model.AccessLevel) string { return al.Name },
	)
	if err != nil {
		return err
	}
	c.accessLevelList, c.accessLevels, c.accessLevelsByName = list, byID, byName
	return nil
}

func (c *DataCache) fillBadgeStatuses() error {
	list, byID, byName, err := fetchAndIndex(c.client, "Lnl_BadgeStatus",
		model.NewBadgeStatusFromJSON,
		func(s *model.BadgeStatus) int32 { return s.ID },
		func(s *model.BadgeStatus) string { return s.Name },
	)
	if err != nil {
		return err
	}
	c.badgeStatusList, c.statuses, c.statusesByName = list, byID, byName
	return nil
}

func (c *DataCache) fillBadgeTypes() error {
	list, byID, byName, err := fetchAndIndex(c.client, "Lnl_BadgeType",
		model.NewBadgeTypeFromJSON,
		func(t *model.BadgeType) int32 { return t.ID },
		func(t *model.BadgeType) string { return t.Name },
	)
	if err != nil {
		return err
	}
	c.badgeTypeList, c.badgeTypes, c.badgeTypesByName = list, byID, byName
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
