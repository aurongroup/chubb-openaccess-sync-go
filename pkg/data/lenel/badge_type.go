package lenel

import (
	"fmt"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/data/model"
	"strings"
)

type BadgeTypeCache struct {
	list   []*model.BadgeType
	byID   map[int32]*model.BadgeType
	byName map[string]*model.BadgeType
}

func NewBadgeTypeCache() *BadgeTypeCache {
	return &BadgeTypeCache{
		byID:   make(map[int32]*model.BadgeType),
		byName: make(map[string]*model.BadgeType),
	}
}

func (c *BadgeTypeCache) GetItems() []*model.BadgeType {
	return c.list
}

func (c *BadgeTypeCache) GetRowItems() []model.RowObject {
	result := make([]model.RowObject, len(c.list))
	for i, v := range c.list {
		result[i] = v
	}
	return result
}

func (c *BadgeTypeCache) GetByID(id int32) *model.BadgeType {
	bt, ok := c.byID[id]

	if ok {
		return bt
	}

	return nil
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

func (c *BadgeTypeCache) Validate(values []string) error {
	failures := make([]string, 0, len(values))

	for _, name := range values {
		if _, ok := c.byName[name]; !ok {
			failures = append(failures, name)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("badge types not found in Lenel: %s", strings.Join(failures, ","))
	}

	return nil
}

func (c *BadgeTypeCache) RowHeader() []string {
	return []string{"ID", "Name"}
}
