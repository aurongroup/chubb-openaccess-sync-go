package lenel

import (
	"fmt"
	"openaccess-sync/pkg/client"
	model2 "openaccess-sync/pkg/data/model"
	"sort"
	"strings"
)

// DataCache holds all data fetched from the OpenAccess API.
type DataCache struct {
	client       *client.Client
	statuses     BadgeStatusCache
	badgeTypes   BadgeTypeCache
	accessLevels AccessLevelCache
	badges       BadgeCache
	cardholders  CardholderCache
	assignments  AssignmentCache
}

// NewDataCache constructs an empty DataCache backed by the given client.
func NewDataCache(cl *client.Client) *DataCache {
	return &DataCache{
		client:       cl,
		statuses:     NewBadgeStatusCache(),
		badgeTypes:   NewBadgeTypeCache(),
		accessLevels: NewAccessLevelCache(),
		badges:       NewBadgeCache(),
		cardholders:  NewCardholderCache(),
		assignments:  NewAssignmentCache(),
	}
}

func (c *DataCache) GetAccessLevel(id int32) *model2.AccessLevel {
	return c.accessLevels.byID[id]
}

func (c *DataCache) GetAccessLevelByName(name string) (*model2.AccessLevel, bool) {
	al, ok := c.accessLevels.byName[name]
	return al, ok
}

func (c *DataCache) GetAccessLevels() []*model2.AccessLevel {
	return c.accessLevels.list
}

func (c *DataCache) GetAccessLevelsByBadge(badgeKey int32) []*model2.AccessLevel {
	//if levels, ok := c.assignments.byBadgeKey[badgeKey]; ok { // FIXME
	//	return levels
	//}
	return []*model2.AccessLevel{}
}

func (c *DataCache) GetBadge(id int32) *model2.Badge {
	return c.badges.byID[id]
}

func (c *DataCache) GetBadges() []*model2.Badge {
	return c.badges.list
}

func (c *DataCache) GetBadgeByKey(key int64) *model2.Badge {
	return c.badges.byKey[key]
}

func (c *DataCache) GetBadgeStatus(id int32) *model2.BadgeStatus {
	return c.statuses.byID[id]
}

func (c *DataCache) GetBadgeStatusByName(name string) (*model2.BadgeStatus, bool) {
	s, ok := c.statuses.byName[name]
	return s, ok
}

func (c *DataCache) GetBadgeStatuses() []*model2.BadgeStatus {
	return c.statuses.list
}

func (c *DataCache) GetBadgeType(id int32) *model2.BadgeType {
	return c.badgeTypes.byID[id]
}

func (c *DataCache) GetBadgeTypeByName(name string) (*model2.BadgeType, bool) {
	t, ok := c.badgeTypes.byName[name]
	return t, ok
}

func (c *DataCache) GetBadgeTypes() []*model2.BadgeType {
	return c.badgeTypes.list
}

func (c *DataCache) GetCardholder(id int32) *model2.Cardholder {
	return c.cardholders.byID[id]
}

func (c *DataCache) GetCardholders() []*model2.Cardholder {
	return c.cardholders.list
}

// Fill fetches all data from the API in the required order.
// References (status, type, cardholder on badges; badge and access level on
// assignments) are resolved against already-populated maps.
func (c *DataCache) Fill() error {
	if err := c.accessLevels.Fill(c.client); err != nil {
		return err
	}

	if err := c.statuses.Fill(c.client); err != nil {
		return err
	}

	if err := c.badgeTypes.Fill(c.client); err != nil {
		return err
	}

	// FIXME
	//if err := c.cardholders.Fill(c.client); err != nil {
	//	return err
	//}

	// FIXME: badges and assignments require cardholders and access levels to be
	// filled first; badge fill also needs IDCache resolution.
	//if err := c.badges.Fill(c.client, c); err != nil {
	//	return err
	//}
	//
	//if err := c.assignments.Fill(c.client, c); err != nil {
	//	return err
	//}

	return nil
}

// ValidateAccessRecords checks that every status, badge type, and access level
// name referenced in records exists in the cache. It collects all unknown values
// and returns a single error listing them grouped by category, or nil if all
// values resolve.
func (c *DataCache) ValidateAccessRecords(records []*model2.AccessRecord) error {
	unknownStatuses := make(map[string]struct{})
	unknownTypes := make(map[string]struct{})
	unknownLevels := make(map[string]struct{})

	for _, r := range records {
		if _, ok := c.statuses.byName[r.Status]; !ok {
			unknownStatuses[r.Status] = struct{}{}
		}
		if _, ok := c.badgeTypes.byName[r.BadgeType]; !ok {
			unknownTypes[r.BadgeType] = struct{}{}
		}
		for _, lvl := range []string{r.AccLvl1, r.AccLvl2, r.AccLvl3, r.AccLvl4, r.AccLvl5, r.AccLvl6} {
			if lvl == "" {
				continue
			}
			if _, ok := c.accessLevels.byName[lvl]; !ok {
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
