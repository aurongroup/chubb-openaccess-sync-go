package lenel

import (
	"openaccess-sync/client"
	"openaccess-sync/data/model"
)

type AccessLevelCache struct {
	list   []*model.AccessLevel
	byID   map[int32]*model.AccessLevel
	byName map[string]*model.AccessLevel
}

func newAccessLevelCache() AccessLevelCache {
	return AccessLevelCache{
		byID:   make(map[int32]*model.AccessLevel),
		byName: make(map[string]*model.AccessLevel),
	}
}

func (c *AccessLevelCache) fill(cl *client.Client) error {
	list, byID, byName, err := fetchAndIndex(cl, "Lnl_AccessLevel",
		model.NewAccessLevelFromJSON,
		func(al *model.AccessLevel) int32 { return al.ID },
		func(al *model.AccessLevel) string { return al.Name },
	)
	if err != nil {
		return err
	}
	c.list, c.byID, c.byName = list, byID, byName
	return nil
}
