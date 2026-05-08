package lenel

import (
	"errors"
	"openaccess-sync/data/cache"
	"openaccess-sync/util/json"
)

var (
	ErrAssignmentNilCache              = errors.New("assignment: cache is nil")
	ErrAssignmentUnresolvedAccessLevel = errors.New("assignment: AccessLevelID not found in cache")
	ErrAssignmentUnresolvedBadge       = errors.New("assignment: BadgeKey not found in cache")
)

// AccessLevelAssignment links an access level to a badge.
type AccessLevelAssignment struct {
	AccessLevel *AccessLevel
	Badge       *Badge
}

func NewAccessLevelAssignment(props map[string]any, cache cache.Cache) (*AccessLevelAssignment, error) {
	if cache == nil {
		return nil, ErrAssignmentNilCache
	}

	a := &AccessLevelAssignment{}

	if alID := json.PropToInt(props, "AccessLevelID"); alID != 0 {
		a.AccessLevel = cache.GetAccessLevel(alID)
	}

	if a.AccessLevel == nil {
		return nil, ErrAssignmentUnresolvedAccessLevel
	}

	if badgeKey := json.PropToInt(props, "BadgeKey"); badgeKey != 0 {
		a.Badge = cache.GetBadgeByKey(badgeKey)
	}

	if a.Badge == nil {
		return nil, ErrAssignmentUnresolvedBadge
	}

	return a, nil
}
