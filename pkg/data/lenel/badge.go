package lenel

import (
	"errors"
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/data/model"
	json "openaccess-sync/pkg/util/json"
)

var ErrBadgeCreateMissingKey = errors.New("badge: create response missing BADGEKEY")

type BadgeCache struct {
	list         []*model.Badge
	byID         map[int64]*model.Badge
	byKey        map[int32]*model.Badge
	byCardholder map[int32][]*model.Badge
}

func NewBadgeCache() BadgeCache {
	return BadgeCache{
		byID:         make(map[int64]*model.Badge),
		byKey:        make(map[int32]*model.Badge),
		byCardholder: make(map[int32][]*model.Badge),
	}
}

func (c *BadgeCache) GetItems() []*model.Badge {
	return c.list
}

func (c *BadgeCache) GetByID(id int64) *model.Badge {
	b, ok := c.byID[id]

	if ok {
		return b
	}

	return nil
}

func (c *BadgeCache) GetByKey(key int32) *model.Badge {
	b, ok := c.byKey[key]

	if ok {
		return b
	}

	return nil
}

func (c *BadgeCache) GetByCardholder(key int32) []*model.Badge {
	b, ok := c.byCardholder[key]

	if ok {
		return b
	}

	return []*model.Badge{}
}

func (c *BadgeCache) Fill(cl *client.Client) error {
	return c.FillWithFilter(cl, "")
}

func (c *BadgeCache) FillWithFilter(cl *client.Client, filter string) error {
	items, err := cl.GetInstancesWithProgress("Lnl_Badge", filter)
	if err != nil {
		return err
	}

	for _, props := range items {
		b, err := model.NewBadgeFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_Badge: %v", err)
			continue
		}

		c.list = append(c.list, b)
		c.byID[b.ID] = b
		c.byKey[b.Key] = b

		if _, ok := c.byCardholder[b.Cardholder]; !ok {
			c.byCardholder[b.Cardholder] = []*model.Badge{}
		}
		c.byCardholder[b.Cardholder] = append(c.byCardholder[b.Cardholder], b)
	}
	log.Printf("Retrieved %d Lnl_Badge records", len(c.list))
	return nil
}

func (c *BadgeCache) Create(cl *client.Client, b *model.Badge) (int32, error) {
	props, err := cl.CreateInstance(
		"Lnl_Badge",
		map[string]interface{}{
			"ID":         b.ID,
			"ACTIVATE":   b.Activate,
			"DEACTIVATE": b.Deactivate,
			"STATUS":     b.Status,
			"TYPE":       b.Type,
			"CARDHOLDER": b.Cardholder,
		},
	)

	if err != nil {
		return 0, err
	}

	key := json.PropToInt32(props, "BADGEKEY")
	if key == 0 {
		return 0, ErrBadgeCreateMissingKey
	}
	b.Key = key // Lnl_Badge uses BADGEKEY as the identifier rather than ID

	return key, nil
}

func (c *BadgeCache) Update(cl *client.Client, b *model.Badge) error {

	return cl.UpdateInstance(
		"Lnl_Badge",
		map[string]interface{}{
			"ID":         b.ID,
			"BADGEKEY":   b.Key,
			"ACTIVATE":   b.Activate,
			"DEACTIVATE": b.Deactivate,
			"STATUS":     b.Status,
			"TYPE":       b.Type,
		},
	)
}

func (c *BadgeCache) Delete(cl *client.Client, b *model.Badge) error {
	return cl.DeleteInstance(
		"Lnl_Badge",
		map[string]interface{}{
			"BADGEKEY": b.Key,
		},
	)
}
