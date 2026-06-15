package lenel

import (
	"fmt"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/data/model"
	"strings"
)

type BadgeStatusCache struct {
	list   []*model.BadgeStatus
	byID   map[int32]*model.BadgeStatus
	byName map[string]*model.BadgeStatus
}

func NewBadgeStatusCache() *BadgeStatusCache {
	return &BadgeStatusCache{
		byID:   make(map[int32]*model.BadgeStatus),
		byName: make(map[string]*model.BadgeStatus),
	}
}

func (c *BadgeStatusCache) GetItems() []*model.BadgeStatus {
	return c.list
}

func (c *BadgeStatusCache) GetRowItems() []model.RowObject {
	result := make([]model.RowObject, len(c.list))
	for i, v := range c.list {
		result[i] = v
	}
	return result
}

func (c *BadgeStatusCache) GetByID(id int32) *model.BadgeStatus {
	bs, ok := c.byID[id]

	if ok {
		return bs
	}

	return nil
}

func (c *BadgeStatusCache) GetByName(name string) *model.BadgeStatus {
	bs, ok := c.byName[name]

	if ok {
		return bs
	}

	return nil
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

func (c *BadgeStatusCache) Validate(values []string) error {
	failures := make([]string, 0, len(values))

	for _, name := range values {
		if _, ok := c.byName[name]; !ok {
			failures = append(failures, name)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("badge statuses not found in Lenel: %s", strings.Join(failures, ","))
	}

	return nil
}

func (c *BadgeStatusCache) RowHeader() []string {
	return []string{"ID", "Name"}
}
