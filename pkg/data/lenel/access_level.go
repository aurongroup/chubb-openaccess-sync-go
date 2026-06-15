package lenel

import (
	"fmt"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/data/model"
	"strings"
)

type AccessLevelCache struct {
	list   []*model.AccessLevel
	byID   map[int32]*model.AccessLevel
	byName map[string]*model.AccessLevel
}

func NewAccessLevelCache() *AccessLevelCache {
	return &AccessLevelCache{
		byID:   make(map[int32]*model.AccessLevel),
		byName: make(map[string]*model.AccessLevel),
	}
}

func (c *AccessLevelCache) GetItems() []*model.AccessLevel {
	return c.list
}

func (c *AccessLevelCache) GetRowItems() []model.RowObject {
	result := make([]model.RowObject, len(c.list))
	for i, v := range c.list {
		result[i] = v
	}
	return result
}

func (c *AccessLevelCache) GetByID(id int32) *model.AccessLevel {
	al, ok := c.byID[id]

	if ok {
		return al
	}

	return nil
}

func (c *AccessLevelCache) Fill(cl *client.Client) error {
	list, byID, byKey, err := fetchAndIndex(cl, "Lnl_AccessLevel",
		model.NewAccessLevelFromJSON,
		func(al *model.AccessLevel) int32 { return al.ID },
		func(al *model.AccessLevel) string { return al.Name },
	)
	if err != nil {
		return err
	}
	c.list, c.byID, c.byName = list, byID, byKey
	return nil
}

func (c *AccessLevelCache) Validate(values []string) error {
	failures := make([]string, 0, len(values))

	for _, name := range values {
		if _, ok := c.byName[name]; !ok {
			failures = append(failures, name)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("access levels not found in Lenel: %s", strings.Join(failures, ","))
	}

	return nil
}

func (c *AccessLevelCache) RowHeader() []string {
	return []string{"ID", "Name"}
}
