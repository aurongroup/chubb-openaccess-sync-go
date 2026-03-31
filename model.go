package main

import (
	"fmt"
	"time"
)

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
func (b *LnlBadge) ToJSON() map[string]any {
	return map[string]any{
		"type_name": "Lnl_Badge",
		"property_value_map": map[string]any{
			"badgeKey":   b.BadgeKey,
			"activate":   dateStr(b.Activate),
			"deactivate": dateStr(b.Deactivate),
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

// AccessRecord represents a single row in the pipe-delimited access control CSV.
type AccessRecord struct {
	SSNO, First, Last         string
	AccLvl1, AccLvl2, AccLvl3 string
	AccLvl4, AccLvl5, AccLvl6 string
	BadgeID                   string
	Activate, Deactivate      *time.Time
	Status, BadgeType         string
	Badge                     *LnlBadge
	AccessLevels              []*LnlAccessLevel
	SyncStatus                SyncStatus
}

// ToRow converts an AccessRecord to a slice of strings for CSV output.
func (r *AccessRecord) ToRow() []string {
	return []string{
		r.SSNO, r.First, r.Last,
		r.AccLvl1, r.AccLvl2, r.AccLvl3, r.AccLvl4, r.AccLvl5, r.AccLvl6,
		r.BadgeID,
		formatDate(r.Activate),
		formatDate(r.Deactivate),
		r.Status, r.BadgeType,
	}
}
