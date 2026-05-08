package lenel

import (
	"errors"
	"fmt"
	"openaccess-sync/data/cache"
	"openaccess-sync/data/model/csv"
	"openaccess-sync/util/date"
	"openaccess-sync/util/json"
	"time"
)

var (
	ErrBadgeNilCache         = errors.New("badge: cache is nil")
	ErrBadgeMissingID        = errors.New("badge: missing required ID")
	ErrBadgeMissingBadgeKey  = errors.New("badge: missing required BadgeKey")
	ErrBadgeUnresolvedStatus = errors.New("badge: STATUS not found in cache")
	ErrBadgeUnresolvedType   = errors.New("badge: TYPE not found in cache")
)

// Badge represents a badge from the OpenAccess API.
type Badge struct {
	ID         int
	BadgeKey   int
	Activate   *time.Time
	Deactivate *time.Time
	Status     *BadgeStatus
	Type       *BadgeType
	Cardholder *Cardholder
}

func NewBadge(props map[string]any, cache cache.Cache) (*Badge, error) {
	if cache == nil {
		return nil, ErrBadgeNilCache
	}

	id := json.PropToInt(props, "ID")
	if id == 0 {
		return nil, ErrBadgeMissingID
	}

	badgeKey := json.PropToInt(props, "BADGEKEY")
	if badgeKey == 0 {
		return nil, ErrBadgeMissingBadgeKey
	}

	b := &Badge{
		ID:         id,
		BadgeKey:   badgeKey,
		Activate:   json.PropToDate(props, "ACTIVATE"),
		Deactivate: json.PropToDate(props, "DEACTIVATE"),
	}

	statusID := json.PropToInt(props, "STATUS")
	if statusID == 0 {
		return nil, ErrBadgeUnresolvedStatus
	}

	b.Status = cache.GetBadgeStatus(statusID)
	if b.Status == nil {
		return nil, ErrBadgeUnresolvedStatus
	}

	typeID := json.PropToInt(props, "TYPE")
	if typeID == 0 {
		return nil, ErrBadgeUnresolvedType
	}

	b.Type = cache.GetBadgeType(typeID)
	if b.Type == nil {
		return nil, ErrBadgeUnresolvedType
	}

	if personID := json.PropToInt(props, "PERSONID"); personID != 0 {
		b.Cardholder = cache.GetCardholder(personID)
	}

	return b, nil
}

// ToAccessRecord builds an AccessRecord from an Badge and its access levels.
func (badge *Badge) ToAccessRecord(accessLevels []*AccessLevel) (*csv.AccessRecord, error) {
	var ssno, first, last, status, badgeType string

	if badge.Cardholder != nil {
		ssno = badge.Cardholder.SSNO
		first = badge.Cardholder.FirstName
		last = badge.Cardholder.LastName
	}

	lvl := [6]string{}
	for i := 0; i < len(accessLevels) && i < 6; i++ {
		lvl[i] = accessLevels[i].Name
	}

	if badge.Status != nil {
		status = badge.Status.Name
	}

	if badge.Type != nil {
		badgeType = badge.Type.Name
	}

	return csv.NewAccessRecord(
		ssno, first, last,
		lvl[0], lvl[1], lvl[2], lvl[3], lvl[4], lvl[5],
		fmt.Sprintf("%d", badge.ID),
		badge.Activate, badge.Deactivate,
		status, badgeType,
	)
}

// ToJSON returns the API wire format map for a badge.
func (badge *Badge) ToJSON() map[string]any {
	return map[string]any{
		"type_name": "Lnl_Badge",
		"property_value_map": map[string]any{
			"badgeKey":   badge.BadgeKey,
			"activate":   date.ISO8601Str(badge.Activate),
			"deactivate": date.ISO8601Str(badge.Deactivate),
		},
	}
}
