package model

import (
	"errors"
	"openaccess-sync/util/json"
)

var (
	ErrAssignmentNilCache       = errors.New("assignment: cache is nil")
	ErrAssignmentNilAccessLevel = errors.New("assignment: AccessLevel is nil")
	ErrAssignmentNilBadge       = errors.New("assignment: Badge is nil")
)

// AccessLevelAssignment links an access level to a badge.
type AccessLevelAssignment struct {
	AccessLevel *AccessLevel
	Badge       *Badge
}

func NewAccessLevelAssignment(al *AccessLevel, b *Badge) (*AccessLevelAssignment, error) {
	if al == nil {
		return nil, ErrAssignmentNilAccessLevel
	}

	if b == nil {
		return nil, ErrAssignmentNilBadge
	}

	a := &AccessLevelAssignment{
		AccessLevel: al,
		Badge:       b,
	}

	return a, nil
}

func NewAccessLevelAssignmentFromJSON(props map[string]any, cache IDCache) (*AccessLevelAssignment, error) {
	if cache == nil {
		return nil, ErrAssignmentNilCache
	}

	alID := json.PropToInt(props, "AccessLevelID")
	if alID == 0 {
		return nil, ErrAssignmentNilAccessLevel
	}

	badgeKey := json.PropToInt(props, "BadgeKey")
	if badgeKey == 0 {
		return nil, ErrAssignmentNilBadge
	}

	return NewAccessLevelAssignment(
		cache.GetAccessLevel(alID),
		cache.GetBadge(badgeKey),
	)
}

func NewAccessLevelAssignmentFromKeys(alk, bk string, cache KeyCache) (*AccessLevelAssignment, error) {
	if cache == nil {
		return nil, ErrAssignmentNilCache
	}

	return NewAccessLevelAssignment(
		cache.GetAccessLevelByKey(alk),
		cache.GetBadgeByKey(bk),
	)
}
