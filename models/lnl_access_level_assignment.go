package models

// LnlAccessLevelAssignment links an access level to a badge.
type LnlAccessLevelAssignment struct {
	AccessLevel *LnlAccessLevel
	Badge       *LnlBadge
}

func NewLnlAccessLevelAssignment(props map[string]any, cache Cache) (*LnlAccessLevelAssignment, error) {
	if cache == nil {
		return nil, ErrAssignmentNilCache
	}

	a := &LnlAccessLevelAssignment{}

	if alID := propInt(props, "AccessLevelID"); alID != 0 {
		a.AccessLevel = cache.GetAccessLevel(alID)
	}

	if a.AccessLevel == nil {
		return nil, ErrAssignmentUnresolvedAccessLevel
	}

	if badgeKey := propInt(props, "BadgeKey"); badgeKey != 0 {
		a.Badge = cache.GetBadgeByKey(badgeKey)
	}

	if a.Badge == nil {
		return nil, ErrAssignmentUnresolvedBadge
	}

	return a, nil
}
