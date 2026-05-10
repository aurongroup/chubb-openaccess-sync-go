package model

import (
	"errors"
	"fmt"
	"openaccess-sync/util/date"
	"openaccess-sync/util/json"
	"strconv"
	"time"
)

var (
	ErrBadgeNilCache             = errors.New("badge: cache is nil")
	ErrBadgeMissingID            = errors.New("badge: missing required ID")
	ErrBadgeMissingBadgeKey      = errors.New("badge: missing required BadgeKey")
	ErrBadgeUnresolvedStatus     = errors.New("badge: STATUS not found in cache")
	ErrBadgeUnresolvedType       = errors.New("badge: TYPE not found in cache")
	ErrBadgeUnresolvedCardholder = errors.New("badge: CARDHOLDER not found in cache")
)

type IDCache interface {
	GetAccessLevel(int) *AccessLevel
	GetBadge(int) *Badge
	GetBadgeByKey(int) *Badge
	GetBadges() []*Badge
	GetBadgeStatus(int) *BadgeStatus
	GetBadgeType(int) *BadgeType
	GetCardholder(int) *Cardholder
	GetAccessLevelsByBadge(int) []*AccessLevel
}

type KeyCache interface {
	GetAccessLevelByKey(string) *AccessLevel
	GetBadgeByKey(int) *Badge
	GetBadgeStatusByKey(string) *BadgeStatus
	GetBadgeTypeByKey(string) *BadgeType
	GetCardholderByKey(string) *Cardholder
}

// Badge represents a badge from the OpenAccess API.
type Badge struct {
	ID         int
	Key        int
	Activate   *time.Time
	Deactivate *time.Time
	Status     *BadgeStatus
	Type       *BadgeType
	Cardholder *Cardholder
}

func NewBadge(id, key int, activate, deactivate *time.Time, badgeStatus *BadgeStatus, badgeType *BadgeType, cardholder *Cardholder) (*Badge, error) {

	if key == 0 {
		return nil, ErrBadgeMissingBadgeKey
	}

	b := &Badge{
		ID:         id,
		Key:        key,
		Activate:   activate,
		Deactivate: deactivate,
	}

	if badgeStatus == nil {
		return nil, ErrBadgeUnresolvedStatus
	}
	b.Status = badgeStatus

	if badgeType == nil {
		return nil, ErrBadgeUnresolvedType
	}
	b.Type = badgeType

	b.Cardholder = cardholder

	return b, nil
}

func NewBadgeFromJSON(props map[string]any, cache IDCache) (*Badge, error) {
	if cache == nil {
		return nil, ErrBadgeNilCache
	}

	id := json.PropToInt(props, "ID")
	if id == 0 {
		return nil, ErrBadgeMissingID
	}

	key := json.PropToInt(props, "BADGEKEY")
	if key == 0 {
		return nil, ErrBadgeMissingBadgeKey
	}

	activate := json.PropToDate(props, "ACTIVATE")
	deactivate := json.PropToDate(props, "DEACTIVATE")

	statusID := json.PropToInt(props, "STATUS")
	if statusID == 0 {
		return nil, ErrBadgeUnresolvedStatus
	}

	badgeStatus := cache.GetBadgeStatus(statusID)
	if badgeStatus == nil {
		return nil, ErrBadgeUnresolvedStatus
	}

	typeID := json.PropToInt(props, "TYPE")
	if typeID == 0 {
		return nil, ErrBadgeUnresolvedType
	}

	badgeType := cache.GetBadgeType(typeID)
	if badgeType == nil {
		return nil, ErrBadgeUnresolvedType
	}

	var cardholder *Cardholder = nil
	if personID := json.PropToInt(props, "PERSONID"); personID != 0 {
		cardholder = cache.GetCardholder(personID)
	}

	return NewBadge(id, key, activate, deactivate, badgeStatus, badgeType, cardholder)
}

func NewBadgeFromKeys(badgeID string, activate, deactivate *time.Time, statusKey, typeKey, cardholderKey string, cache KeyCache) (*Badge, error) {
	if cache == nil {
		return nil, ErrBadgeNilCache
	}

	key, err := strconv.Atoi(badgeID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse badge ID: %w", err)
	}

	if key == 0 {
		return nil, ErrBadgeMissingBadgeKey
	}

	badgeStatus := cache.GetBadgeStatusByKey(statusKey)
	if badgeStatus == nil {
		return nil, ErrBadgeUnresolvedStatus
	}

	badgeType := cache.GetBadgeTypeByKey(typeKey)
	if badgeType == nil {
		return nil, ErrBadgeUnresolvedType
	}

	var cardholder *Cardholder = nil
	cardholder = cache.GetCardholderByKey(cardholderKey)
	if cardholder == nil {
		return nil, ErrBadgeUnresolvedCardholder
	}

	return NewBadge(0, key, activate, deactivate, badgeStatus, badgeType, cardholder)
}

// ToAccessRecord builds an AccessRecord from a Badge and its access levels.
func (badge *Badge) ToAccessRecord(accessLevels []*AccessLevel) (*AccessRecord, error) {
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

	return NewAccessRecord(
		ssno, first, last,
		lvl[0], lvl[1], lvl[2], lvl[3], lvl[4], lvl[5],
		fmt.Sprintf("%d", badge.ID),
		badge.Activate, badge.Deactivate,
		status, badgeType,
	)
}

// ToJSON returns the API wire format map for a b.
func (b *Badge) ToJSON() map[string]any {
	return map[string]any{
		"type_name": "Lnl_Badge",
		"property_value_map": map[string]any{
			"badgeKey":   b.Key,
			"activate":   date.ISO8601Str(b.Activate),
			"deactivate": date.ISO8601Str(b.Deactivate),
		},
	}
}
