package lenel

import (
	"openaccess-sync/client"
	"openaccess-sync/data/model"
)

type BadgeStatusCache struct {
	list   []*model.BadgeStatus
	byID   map[int32]*model.BadgeStatus
	byName map[string]*model.BadgeStatus
}

func NewBadgeStatusCache() BadgeStatusCache {
	return BadgeStatusCache{
		byID:   make(map[int32]*model.BadgeStatus),
		byName: make(map[string]*model.BadgeStatus),
	}
}

func (c *BadgeStatusCache) Fill(cl *client.Client) error {
	list, byID, byKey, err := fetchAndIndex(cl, "Lnl_BadgeStatus",
		model.NewBadgeStatusFromJSON,
		func(s *model.BadgeStatus) int32 { return s.ID },
		func(s *model.BadgeStatus) string { return s.Name },
	)
	if err != nil {
		return err
	}
	c.list, c.byID, c.byName = list, byID, byKey
	return nil
}
