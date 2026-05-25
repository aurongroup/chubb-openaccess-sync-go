package model

import (
	"errors"
	"openaccess-sync/pkg/util/json"
)

var (
	ErrAssignmentNilCache       = errors.New("assignment: cache is nil")
	ErrAssignmentNilAccessLevel = errors.New("assignment: AccessLevel is nil")
	ErrAssignmentNilBadge       = errors.New("assignment: Badge is nil")
)

// AccessLevelAssignment links an access level to a badge.
type AccessLevelAssignment struct {
	AccessLevel int32
	BadgeKey    int32
}

func NewAccessLevelAssignment(al, b int32) (*AccessLevelAssignment, error) {
	if al == 0 {
		return nil, ErrAssignmentNilAccessLevel
	}

	if b == 0 {
		return nil, ErrAssignmentNilBadge
	}

	a := &AccessLevelAssignment{
		AccessLevel: al,
		BadgeKey:    b,
	}

	return a, nil
}

func NewAccessLevelAssignmentFromJSON(props map[string]any) (*AccessLevelAssignment, error) {
	alID := json.PropToInt32(props, "AccessLevelID")
	if alID == 0 {
		return nil, ErrAssignmentNilAccessLevel
	}

	badgeKey := json.PropToInt32(props, "BadgeKey")
	if badgeKey == 0 {
		return nil, ErrAssignmentNilBadge
	}

	return NewAccessLevelAssignment(
		alID,
		badgeKey, // Lnl_AccessLevel instances use BadgeKey rather than ID as a link
	)
}

// TODO
//func NewAccessLevelAssignmentFromKeys(alk, bk string, cache KeyCache) (*AccessLevelAssignment, error) {
//	if cache == nil {
//		return nil, ErrAssignmentNilCache
//	}
//
//	return NewAccessLevelAssignment(
//		cache.GetAccessLevelByKey(alk),
//		cache.GetBadgeByKey(bk),
//	)
//}
