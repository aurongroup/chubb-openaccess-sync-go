package model

import (
	"errors"
	"fmt"
	"openaccess-sync/pkg/util/date"
	"openaccess-sync/pkg/util/json"
	"strconv"
	"time"
)

var (
	ErrBadgeNilCache             = errors.New("badge: cache is nil")
	ErrBadgeMissingID            = errors.New("badge: missing required ID")
	ErrBadgeUnresolvedStatus     = errors.New("badge: STATUS not found in cache")
	ErrBadgeUnresolvedType       = errors.New("badge: TYPE not found in cache")
	ErrBadgeUnresolvedCardholder = errors.New("badge: CARDHOLDER not found in cache")
)

type IDCache interface {
	GetAccessLevel(int32) *AccessLevel
	GetBadge(int32) *Badge
	GetBadgeByKey(int64) *Badge
	GetBadges() []*Badge
	GetBadgeStatus(int32) *BadgeStatus
	GetBadgeType(int32) *BadgeType
	//GetCardholder(int) *Cardholder // FIXME
	//GetAccessLevelsByBadge(int32) []*AccessLevel // FIXME
}

type KeyCache interface {
	GetAccessLevelByKey(string) *AccessLevel
	GetBadgeByKey(int64) *Badge
	GetBadgeStatusByKey(string) *BadgeStatus
	GetBadgeTypeByKey(string) *BadgeType
	GetCardholderByKey(string) *Cardholder
}

// Badge represents a badge from the OpenAccess API.
// For Lnl_Badge objects, 'key' (BADGEKEY) is actually the internal database ID which is an int32,
// while 'id' (ID) is the user-specified identifier which is an int64
type Badge struct {
	ID         int64
	Key        int32
	Activate   *time.Time
	Deactivate *time.Time
	Status     int32
	Type       int32
	Cardholder int32
}

func NewBadge(id int64, key int32, activate, deactivate *time.Time, badgeStatus, badgeType, cardholder int32) (*Badge, error) {

	if id == 0 {
		return nil, ErrBadgeMissingID
	}

	b := &Badge{
		ID:         id,
		Key:        key,
		Activate:   activate,
		Deactivate: deactivate,
		Status:     badgeStatus,
		Type:       badgeType,
		Cardholder: cardholder,
	}

	return b, nil
}

func NewBadgeFromJSON(props map[string]any) (*Badge, error) {
	id := json.PropToInt64(props, "ID")
	if id == 0 {
		return nil, ErrBadgeMissingID
	}

	key := json.PropToInt32(props, "BADGEKEY")

	activate := json.PropToDate(props, "ACTIVATE")
	deactivate := json.PropToDate(props, "DEACTIVATE")

	statusID := json.PropToInt32(props, "STATUS")
	if statusID == 0 {
		return nil, ErrBadgeUnresolvedStatus
	}

	typeID := json.PropToInt32(props, "TYPE")
	if typeID == 0 {
		return nil, ErrBadgeUnresolvedType
	}

	personID := json.PropToInt32(props, "PERSONID")
	if personID == 0 {
		return nil, ErrBadgeUnresolvedCardholder
	}

	return NewBadge(id, key, activate, deactivate, statusID, typeID, personID)
}

// FIXME
//func NewBadgeFromAccessRecord(a *AccessRecord) (*Badge, error) {
//	return NewBadge(
//		0,
//		a.Key,
//		a.Activate,
//		a.Deactivate,
//		a.Status,
//		a.Type,
//		a.Cardholder,
//	)
//}

// FIXME
//func NewBadgeFromKeys(badgeID string, activate, deactivate *time.Time, statusKey, typeKey, cardholderKey string, cache KeyCache) (*Badge, error) {
//	if cache == nil {
//		return nil, ErrBadgeNilCache
//	}
//
//	key, err := strconv.Atoi(badgeID)
//	if err != nil {
//		return nil, fmt.Errorf("failed to parse badge ID: %w", err)
//	}
//
//	if key == 0 {
//		return nil, ErrBadgeMissingBadgeKey
//	}
//
//	badgeStatus := cache.GetBadgeStatusByKey(statusKey)
//	if badgeStatus == nil {
//		return nil, ErrBadgeUnresolvedStatus
//	}
//
//	badgeType := cache.GetBadgeTypeByKey(typeKey)
//	if badgeType == nil {
//		return nil, ErrBadgeUnresolvedType
//	}
//
//	var cardholder *Cardholder = nil
//	cardholder = cache.GetCardholderByKey(cardholderKey)
//	if cardholder == nil {
//		return nil, ErrBadgeUnresolvedCardholder
//	}
//
//	return NewBadge(0, key, activate, deactivate, badgeStatus, badgeType, cardholder)
//}

// ToAccessRecord builds an AccessRecord from a Badge and its access levels.
func (b *Badge) ToAccessRecord(accessLevels []*AccessLevel) (*AccessRecord, error) {
	var ssno, first, last, status, badgeType string

	//if b.Cardholder != nil { // FIXME
	//	ssno = b.Cardholder.SSNO
	//	first = b.Cardholder.FirstName
	//	last = b.Cardholder.LastName
	//}

	lvl := [6]string{}
	for i := 0; i < len(accessLevels) && i < 6; i++ {
		lvl[i] = accessLevels[i].Name
	}

	//if b.Status != nil { // FIXME
	//	status = b.Status.Name
	//}
	//
	//if b.Type != nil {
	//	badgeType = b.Type.Name
	//}

	return NewAccessRecord(
		ssno, first, last,
		lvl[0], lvl[1], lvl[2], lvl[3], lvl[4], lvl[5],
		fmt.Sprintf("%d", b.ID),
		b.Activate, b.Deactivate,
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

func (b *Badge) ToRow() []string {
	return []string{
		strconv.FormatInt(int64(b.ID), 10),
		strconv.FormatInt(int64(b.Key), 10),
		"ACTIVATE",
		"DEACTIVATE",
		"STATUS",
		"TYPE",
		"CARDHOLDER",
	}
}
