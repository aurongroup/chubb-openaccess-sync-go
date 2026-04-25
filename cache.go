package main

import (
	"fmt"
	"log"

	ansi "github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
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
	assignments     []*LnlAccessLevelAssignment

	records                []*AccessRecord
	recordsByCardholderKey map[string][]*AccessRecord
}

// NewDataCache constructs an empty DataCache backed by the given client.
func NewDataCache(client *Client) *DataCache {
	return &DataCache{
		client:                 client,
		accessLevels:           make(map[int]*LnlAccessLevel),
		badges:                 make(map[int]*LnlBadge),
		badgeByKey:             make(map[int]*LnlBadge),
		statuses:               make(map[int]*LnlBadgeStatus),
		badgeTypes:             make(map[int]*LnlBadgeType),
		cardholders:            make(map[int]*LnlCardholder),
		recordsByCardholderKey: make(map[string][]*AccessRecord),
	}
}

func (c *DataCache) GetAccessLevel(id int) *LnlAccessLevel {
	return c.accessLevels[id]
}

func (c *DataCache) GetBadge(id int) *LnlBadge {
	return c.badges[id]
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

func (c *DataCache) GetRecordsByCardholderKey(key string) []*AccessRecord {
	return c.recordsByCardholderKey[key]
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

	c.records = c.buildAccessRecordList()

	return nil
}

// fetchWithProgress pages through all results for typeName, displaying a progress
// bar as each page arrives. The bar max is totalPages from the first response.
func (c *DataCache) fetchWithProgress(typeName, filter string) ([]map[string]any, error) {
	var all []map[string]any
	var bar *progressbar.ProgressBar

	log.Printf("Fetching %s pages from OpenAccess API...", typeName)

	for page := 1; ; page++ {
		items, totalPages, err := c.client.getInstancesPage(typeName, filter, page)
		if err != nil {
			return nil, err
		}

		if bar == nil {
			bar = progressbar.NewOptions(totalPages,
				progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
				progressbar.OptionEnableColorCodes(true),
				progressbar.OptionSetDescription(typeName),
				progressbar.OptionShowDescriptionAtLineEnd(),
				progressbar.OptionShowCount(),
				progressbar.OptionSetWidth(30),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "[green]=[reset]",
					SaucerHead:    "[green]>[reset]",
					SaucerPadding: " ",
					BarStart:      "[",
					BarEnd:        "]",
				}),
			)
		}
		_ = bar.Add(1)

		all = append(all, items...)

		if page >= totalPages {
			break
		}
	}
	_ = bar.Finish()

	fmt.Println()

	return all, nil
}

func (c *DataCache) fillAccessLevels() error {
	items, err := c.fetchWithProgress("Lnl_AccessLevel", "")
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
	items, err := c.fetchWithProgress("Lnl_BadgeStatus", "")
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
	items, err := c.fetchWithProgress("Lnl_BadgeType", "")
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
	items, err := c.fetchWithProgress("Lnl_Cardholder", "")
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
	items, err := c.fetchWithProgress("Lnl_Badge", "")
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
	items, err := c.fetchWithProgress("Lnl_AccessLevelAssignment", "")
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
func (c *DataCache) buildAccessRecordList() []*AccessRecord {
	levelsByBadgeID := c.accessLevelsByBadge()
	records := make([]*AccessRecord, 0, len(c.badgeList))

	for _, badge := range c.badgeList {
		r, err := badge.ToAccessRecord(levelsByBadgeID[badge.ID])
		if err != nil {
			log.Printf("skipping access record for badge %d: %v", badge.ID, err)
			continue
		}

		records = append(records, r)

		if r.CardholderKey != "" {
			c.recordsByCardholderKey[r.CardholderKey] = append(c.recordsByCardholderKey[r.CardholderKey], r)
		}
	}

	return records
}
