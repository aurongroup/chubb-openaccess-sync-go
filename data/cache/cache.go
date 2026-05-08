package cache

import "openaccess-sync/data/model/lenel"

type Cache interface {
	GetAccessLevelsByBadge() map[int][]*lenel.AccessLevel
	GetAccessLevel(id int) *lenel.AccessLevel
	GetAccessLevels() []*lenel.AccessLevel
	GetBadge(id int) *lenel.Badge
	GetBadges() []*lenel.Badge
	GetBadgeByKey(key int) *lenel.Badge
	GetBadgeStatus(id int) *lenel.BadgeStatus
	GetBadgeStatuses() []*lenel.BadgeStatus
	GetBadgeType(id int) *lenel.BadgeType
	GetBadgeTypes() []*lenel.BadgeType
	GetCardholder(id int) *lenel.Cardholder
	GetCardholders() []*lenel.Cardholder
	Fill() error
}
