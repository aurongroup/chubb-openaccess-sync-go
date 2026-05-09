package data

import (
	"openaccess-sync/data/model"
)

type Cache interface {
	GetAccessLevelsByBadge() map[int][]*model.AccessLevel
	GetAccessLevel(id int) *model.AccessLevel
	GetAccessLevels() []*model.AccessLevel
	GetBadge(id int) *model.Badge
	GetBadges() []*model.Badge
	GetBadgeByKey(key int) *model.Badge
	GetBadgeStatus(id int) *model.BadgeStatus
	GetBadgeStatuses() []*model.BadgeStatus
	GetBadgeType(id int) *model.BadgeType
	GetBadgeTypes() []*model.BadgeType
	GetCardholder(id int) *model.Cardholder
	GetCardholders() []*model.Cardholder
	Fill() error
}
