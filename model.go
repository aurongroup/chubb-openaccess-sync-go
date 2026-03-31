package main

import (
	"fmt"
	"time"
)

// AccessRecord represents a single row in the pipe-delimited access control CSV.
type AccessRecord struct {
	SSNO         string
	First        string
	Last         string
	AccLvl1      string
	AccLvl2      string
	AccLvl3      string
	AccLvl4      string
	AccLvl5      string
	AccLvl6      string
	BadgeID      string
	Activate     *time.Time
	Deactivate   *time.Time
	Status       string
	BadgeType    string
	Badge        *LnlBadge
	AccessLevels []*LnlAccessLevel
	SyncStatus   SyncStatus
}

// ToRow converts an AccessRecord to a slice of strings for CSV output.
func (r *AccessRecord) ToRow() []string {
	return []string{
		r.SSNO,
		r.First,
		r.Last,
		r.AccLvl1,
		r.AccLvl2,
		r.AccLvl3,
		r.AccLvl4,
		r.AccLvl5,
		r.AccLvl6,
		r.BadgeID,
		formatDate(r.Activate),
		formatDate(r.Deactivate),
		r.Status,
		r.BadgeType,
	}
}

// LnlBadgeStatus represents a badge status from the OpenAccess API.
type LnlBadgeStatus struct {
	ID   int
	Name string
}

func NewLnlBadgeStatus(props map[string]any) *LnlBadgeStatus {
	return &LnlBadgeStatus{
		ID:   propInt(props, "ID"),
		Name: propStr(props, "Name"),
	}
}

// LnlBadgeType represents a badge type from the OpenAccess API.
type LnlBadgeType struct {
	ID   int
	Name string
}

func NewLnlBadgeType(props map[string]any) *LnlBadgeType {
	return &LnlBadgeType{
		ID:   propInt(props, "ID"),
		Name: propStr(props, "Name"),
	}
}

// FromBadge builds an AccessRecord from an API badge and its access levels.
func (badge *LnlBadge) FromBadge(accessLevels []*LnlAccessLevel) AccessRecord {
	var ssno, first, last string
	if badge.Cardholder != nil {
		ssno = badge.Cardholder.SSNO
		first = badge.Cardholder.FirstName
		last = badge.Cardholder.LastName
	}

	lvl := [6]string{}
	for i := 0; i < len(accessLevels) && i < 6; i++ {
		lvl[i] = accessLevels[i].Name
	}

	var status, badgeType string
	if badge.Status != nil {
		status = badge.Status.Name
	}
	if badge.Type != nil {
		badgeType = badge.Type.Name
	}

	return AccessRecord{
		SSNO:         ssno,
		First:        first,
		Last:         last,
		AccLvl1:      lvl[0],
		AccLvl2:      lvl[1],
		AccLvl3:      lvl[2],
		AccLvl4:      lvl[3],
		AccLvl5:      lvl[4],
		AccLvl6:      lvl[5],
		BadgeID:      fmt.Sprintf("%d", badge.ID),
		Activate:     badge.Activate,
		Deactivate:   badge.Deactivate,
		Status:       status,
		BadgeType:    badgeType,
		Badge:        badge,
		AccessLevels: accessLevels,
	}
}

// LnlAccessLevel represents an access level from the OpenAccess API.
type LnlAccessLevel struct {
	ID   int
	Name string
}

func NewLnlAccessLevel(props map[string]any) (*LnlAccessLevel, error) {
	id := propInt(props, "ID")
	if id == 0 {
		return nil, fmt.Errorf("JSON node does not contain ID: %v", props)
	}
	name := propStr(props, "Name")
	if name == "" {
		return nil, fmt.Errorf("JSON node does not contain name: %v", props)
	}
	return &LnlAccessLevel{ID: id, Name: name}, nil
}

// ToJSON returns the API wire format map for an access level.
func (a *LnlAccessLevel) ToJSON() map[string]any {
	return map[string]any{
		"property_value_map": map[string]any{
			"ID":   a.ID,
			"Name": a.Name,
		},
	}
}

// LnlCardholder represents a cardholder from the OpenAccess API.
type LnlCardholder struct {
	ID        int
	FirstName string
	LastName  string
	SSNO      string
}

func NewLnlCardholder(props map[string]any) *LnlCardholder {
	return &LnlCardholder{
		ID:        propInt(props, "ID"),
		FirstName: propStr(props, "FIRSTNAME"),
		LastName:  propStr(props, "LASTNAME"),
		SSNO:      propStr(props, "SSNO"),
	}
}

// LnlBadge represents a badge from the OpenAccess API.
type LnlBadge struct {
	ID         int
	BadgeKey   int
	Activate   *time.Time
	Deactivate *time.Time
	Status     *LnlBadgeStatus
	Type       *LnlBadgeType
	Cardholder *LnlCardholder
}

func NewLnlBadge(props map[string]any, cache *DataCache) *LnlBadge {
	b := &LnlBadge{
		ID:         propInt(props, "ID"),
		BadgeKey:   propInt(props, "BADGEKEY"),
		Activate:   propDate(props, "ACTIVATE"),
		Deactivate: propDate(props, "DEACTIVATE"),
	}
	if cache != nil {
		if statusID := propInt(props, "STATUS"); statusID != 0 {
			b.Status = cache.GetBadgeStatus(statusID)
		}
		if typeID := propInt(props, "TYPE"); typeID != 0 {
			b.Type = cache.GetBadgeType(typeID)
		}
		if personID := propInt(props, "PERSONID"); personID != 0 {
			b.Cardholder = cache.GetCardholder(personID)
		}
	}
	return b
}

// ToJSON returns the API wire format map for a badge.
func (badge *LnlBadge) ToJSON() map[string]any {
	return map[string]any{
		"type_name": "Lnl_Badge",
		"property_value_map": map[string]any{
			"badgeKey":   badge.BadgeKey,
			"activate":   dateStr(badge.Activate),
			"deactivate": dateStr(badge.Deactivate),
		},
	}
}

// LnlAccessLevelAssignment links an access level to a badge.
type LnlAccessLevelAssignment struct {
	AccessLevel *LnlAccessLevel
	Badge       *LnlBadge
}

func NewLnlAccessLevelAssignment(props map[string]any, cache *DataCache) *LnlAccessLevelAssignment {
	a := &LnlAccessLevelAssignment{}
	if cache != nil {
		if alID := propInt(props, "AccessLevelID"); alID != 0 {
			a.AccessLevel = cache.GetAccessLevel(alID)
		}
		if badgeKey := propInt(props, "BadgeKey"); badgeKey != 0 {
			a.Badge = cache.GetBadgeByKey(badgeKey)
		}
	}
	return a
}
