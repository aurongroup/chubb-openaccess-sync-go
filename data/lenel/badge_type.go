package lenel

import (
	"openaccess-sync/client"
	"openaccess-sync/data/model"
)

type BadgeTypeCache struct {
	list   []*model.BadgeType
	byID   map[int32]*model.BadgeType
	byName map[string]*model.BadgeType
}

func NewBadgeTypeCache() BadgeTypeCache {
	return BadgeTypeCache{
		byID:   make(map[int32]*model.BadgeType),
		byName: make(map[string]*model.BadgeType),
	}
}

func (c *BadgeTypeCache) Fill(cl *client.Client) error {
	list, byID, byKey, err := fetchAndIndex(cl, "Lnl_BadgeType",
		model.NewBadgeTypeFromJSON,
		func(t *model.BadgeType) int32 { return t.ID },
		func(t *model.BadgeType) string { return t.Name },
	)
	if err != nil {
		return err
	}
	c.list, c.byID, c.byName = list, byID, byKey
	return nil
}
