package main

import (
	"log/slog"
)

// DataCache holds all data fetched from the OpenAccess API.
type DataCache struct {
	client *Client

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
	assignments     []LnlAccessLevelAssignment

	records []AccessRecord
}

// NewDataCache constructs an empty DataCache backed by the given client.
func NewDataCache(client *Client) *DataCache {
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

func (c *DataCache) GetAccessLevel(id int) *LnlAccessLevel { return c.accessLevels[id] }
func (c *DataCache) GetBadge(id int) *LnlBadge             { return c.badges[id] }
func (c *DataCache) GetBadgeByKey(key int) *LnlBadge       { return c.badgeByKey[key] }
func (c *DataCache) GetBadgeStatus(id int) *LnlBadgeStatus { return c.statuses[id] }
func (c *DataCache) GetBadgeType(id int) *LnlBadgeType     { return c.badgeTypes[id] }
func (c *DataCache) GetCardholder(id int) *LnlCardholder   { return c.cardholders[id] }

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
	c.records = c.buildAccessRecordList()
	return nil
}

func (c *DataCache) fillAccessLevels() error {
	items, err := c.client.GetInstances("Lnl_AccessLevel", "")
	if err != nil {
		return err
	}
	for _, props := range items {
		al, err := NewLnlAccessLevel(props)
		if err != nil {
			slog.Warn("skipping access level", "err", err)
			continue
		}
		c.accessLevels[al.ID] = al
		c.accessLevelList = append(c.accessLevelList, al)
	}
	slog.Info("Retrieved access levels", "count", len(c.accessLevelList))
	return nil
}

func (c *DataCache) fillBadgeStatuses() error {
	items, err := c.client.GetInstances("Lnl_BadgeStatus", "")
	if err != nil {
		return err
	}
	for _, props := range items {
		s := NewLnlBadgeStatus(props)
		c.statuses[s.ID] = s
		c.badgeStatusList = append(c.badgeStatusList, s)
	}
	slog.Info("Retrieved badge statuses", "count", len(c.badgeStatusList))
	return nil
}

func (c *DataCache) fillBadgeTypes() error {
	items, err := c.client.GetInstances("Lnl_BadgeType", "")
	if err != nil {
		return err
	}
	for _, props := range items {
		t := NewLnlBadgeType(props)
		c.badgeTypes[t.ID] = t
		c.badgeTypeList = append(c.badgeTypeList, t)
	}
	slog.Info("Retrieved badge types", "count", len(c.badgeTypeList))
	return nil
}

func (c *DataCache) fillCardholders() error {
	items, err := c.client.GetInstances("Lnl_Cardholder", "")
	if err != nil {
		return err
	}
	for _, props := range items {
		ch := NewLnlCardholder(props)
		c.cardholders[ch.ID] = ch
		c.cardholderList = append(c.cardholderList, ch)
	}
	slog.Info("Retrieved cardholders", "count", len(c.cardholderList))
	return nil
}

func (c *DataCache) fillBadges() error {
	items, err := c.client.GetInstances("Lnl_Badge", "")
	if err != nil {
		return err
	}
	for _, props := range items {
		b := NewLnlBadge(props, c)
		c.badges[b.ID] = b
		c.badgeByKey[b.BadgeKey] = b
		c.badgeList = append(c.badgeList, b)
	}
	slog.Info("Retrieved badges", "count", len(c.badgeList))
	return nil
}

func (c *DataCache) fillAssignments() error {
	items, err := c.client.GetInstances("Lnl_AccessLevelAssignment", "")
	if err != nil {
		return err
	}
	for _, props := range items {
		a := NewLnlAccessLevelAssignment(props, c)
		if a.AccessLevel != nil && a.Badge != nil {
			c.assignments = append(c.assignments, *a)
		}
	}
	slog.Info("Retrieved assignments", "count", len(c.assignments))
	return nil
}

// accessLevelsByBadge returns a map from badge ID to its assigned access levels,
// in assignment order.
func (c *DataCache) accessLevelsByBadge() map[int][]*LnlAccessLevel {
	m := make(map[int][]*LnlAccessLevel)
	for _, a := range c.assignments {
		m[a.Badge.ID] = append(m[a.Badge.ID], a.AccessLevel)
	}
	return m
}

// buildAccessRecordList groups assignments by badge ID and creates one
// AccessRecord per badge, matching Java's DataCache.buildAccessRecordList().
func (c *DataCache) buildAccessRecordList() []AccessRecord {
	levelsByBadgeID := c.accessLevelsByBadge()
	records := make([]AccessRecord, 0, len(c.badgeList))
	for _, badge := range c.badgeList {
		records = append(records, accessRecordFromBadge(badge, levelsByBadgeID[badge.ID]))
	}
	return records
}
