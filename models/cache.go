package models

import (
	"log"
	"openaccess-sync/client"
)

// Cache is the lookup interface that model constructors require to resolve
// referenced entities. *DataCache in the main package satisfies this implicitly.
type Cache interface {
	GetBadgeStatus(id int) *LnlBadgeStatus
	GetBadgeType(id int) *LnlBadgeType
	GetCardholder(id int) *LnlCardholder
	GetAccessLevel(id int) *LnlAccessLevel
	GetBadgeByKey(key int) *LnlBadge
}

var _ Cache = (*DataCache)(nil)

// AccessRecordCache holds the computed access records derived from a DataCache.
type AccessRecordCache struct {
	records                []*AccessRecord
	recordsByCardholderKey map[string][]*AccessRecord
}

func BuildAccessRecordCache(c *DataCache) *AccessRecordCache {
	arc := AccessRecordCache{
		recordsByCardholderKey: make(map[string][]*AccessRecord),
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

func (c *AccessRecordCache) Records() []*AccessRecord {
	return c.records
}

func (c *AccessRecordCache) GetRecordsByCardholderKey(key string) []*AccessRecord {
	return c.recordsByCardholderKey[key]
}

// DataCache holds all data fetched from the OpenAccess API.
type DataCache struct {
	client *client.Client

	// indexed maps for O(1) lookup
	accessLevels map[int]*LnlAccessLevel
	badges       map[int]*LnlBadge
	badgeByKey   map[int]*LnlBadge
	statuses     map[int]*LnlBadgeStatus
	badgeTypes   map[int]*LnlBadgeType
	cardholders  map[int]*LnlCardholder

	// ordered slices for iteration (matches Java insertion order)
	accessLevelList []*LnlAccessLevel
	badgeList       []*LnlBadge
	cardholderList  []*LnlCardholder
	badgeStatusList []*LnlBadgeStatus
	badgeTypeList   []*LnlBadgeType
	assignments     []*LnlAccessLevelAssignment
}

// NewDataCache constructs an empty DataCache backed by the given client.
func NewDataCache(client *client.Client) *DataCache {
	return &DataCache{
		client:       client,
		accessLevels: make(map[int]*LnlAccessLevel),
		badges:       make(map[int]*LnlBadge),
		badgeByKey:   make(map[int]*LnlBadge),
		statuses:     make(map[int]*LnlBadgeStatus),
		badgeTypes:   make(map[int]*LnlBadgeType),
		cardholders:  make(map[int]*LnlCardholder),
	}
}

func (c *DataCache) GetAccessLevel(id int) *LnlAccessLevel {
	return c.accessLevels[id]
}

func (c *DataCache) GetBadge(id int) *LnlBadge {
	return c.badges[id]
}

func (c *DataCache) GetBadges() []*LnlBadge {
	return c.badgeList
}

func (c *DataCache) GetBadgeByKey(key int) *LnlBadge {
	return c.badgeByKey[key]
}

func (c *DataCache) GetBadgeStatus(id int) *LnlBadgeStatus {
	return c.statuses[id]
}

func (c *DataCache) GetBadgeType(id int) *LnlBadgeType {
	return c.badgeTypes[id]
}

func (c *DataCache) GetCardholder(id int) *LnlCardholder {
	return c.cardholders[id]
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

	if err := c.fillBadges(); err != nil {
		return err
	}

	if err := c.fillAssignments(); err != nil {
		return err
	}

	return nil
}

func (c *DataCache) fillAccessLevels() error {
	items, err := c.client.GetInstancesWithProgress("Lnl_AccessLevel", "")
	if err != nil {
		return err
	}

	for _, props := range items {
		al, err := NewLnlAccessLevel(props)
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
		s, err := NewLnlBadgeStatus(props)
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
		t, err := NewLnlBadgeType(props)
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
		ch, err := NewLnlCardholder(props)
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

func (c *DataCache) fillBadges() error {
	items, err := c.client.GetInstancesWithProgress("Lnl_Badge", "")
	if err != nil {
		return err
	}

	for _, props := range items {
		b, err := NewLnlBadge(props, c)
		if err != nil {
			log.Printf("skipping Lnl_Badge: %v", err)
			continue
		}

		c.badges[b.ID] = b
		c.badgeByKey[b.BadgeKey] = b
		c.badgeList = append(c.badgeList, b)
	}
	log.Printf("Retrieved %d Lnl_Badge records", len(c.badgeList))
	return nil
}

func (c *DataCache) fillAssignments() error {
	items, err := c.client.GetInstancesWithProgress("Lnl_AccessLevelAssignment", "")
	if err != nil {
		return err
	}

	for _, props := range items {
		a, err := NewLnlAccessLevelAssignment(props, c)
		if err != nil {
			log.Printf("skipping Lnl_AccessLevelAssignment: %v", err)
			continue
		}

		c.assignments = append(c.assignments, a)
	}
	log.Printf("Retrieved %d Lnl_AccessLevelAssignment records", len(c.assignments))
	return nil
}

// GetAccessLevelsByBadge returns a map from badge ID to its assigned access levels,
// in assignment order.
func (c *DataCache) GetAccessLevelsByBadge() map[int][]*LnlAccessLevel {
	m := make(map[int][]*LnlAccessLevel)

	for _, a := range c.assignments {
		m[a.Badge.ID] = append(m[a.Badge.ID], a.AccessLevel)
	}

	return m
}
