package lenel

import (
	"openaccess-sync/data"
	"openaccess-sync/data/cache"
	"openaccess-sync/util/json"
)

// LnlAccessLevelAssignment links an access level to a badge.
type LnlAccessLevelAssignment struct {
	AccessLevel *LnlAccessLevel
	Badge       *LnlBadge
}

func NewLnlAccessLevelAssignment(props map[string]any, cache *cache.DataCache) (*LnlAccessLevelAssignment, error) {
	if cache == nil {
		return nil, data.ErrAssignmentNilCache
	}

	a := &LnlAccessLevelAssignment{}

	if alID := json.PropToInt(props, "AccessLevelID"); alID != 0 {
		a.AccessLevel = cache.GetAccessLevel(alID)
	}

	if a.AccessLevel == nil {
		return nil, data.ErrAssignmentUnresolvedAccessLevel
	}

	if badgeKey := json.PropToInt(props, "BadgeKey"); badgeKey != 0 {
		a.Badge = cache.GetBadgeByKey(badgeKey)
	}

	if a.Badge == nil {
		return nil, data.ErrAssignmentUnresolvedBadge
	}

	return a, nil
}
