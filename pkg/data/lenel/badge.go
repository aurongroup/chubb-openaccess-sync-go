package lenel

import (
	"errors"
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/data/model"
	"openaccess-sync/pkg/util/date"
	json "openaccess-sync/pkg/util/json"
	"strconv"
)

var ErrBadgeCreateMissingKey = errors.New("badge: create response missing BADGEKEY")

type badgeRow struct {
	id, key, activate, deactivate, status, badgeType, ssno string
	levels                                                 [6]string
}

func (r badgeRow) ToRow() []string {
	return []string{
		r.id, r.key, r.activate, r.deactivate,
		r.status, r.badgeType, r.ssno,
		r.levels[0], r.levels[1], r.levels[2],
		r.levels[3], r.levels[4], r.levels[5],
	}
}

type BadgeCache struct {
	list         []*model.Badge
	byID         map[int64]*model.Badge
	byKey        map[int32]*model.Badge
	byCardholder map[int32][]*model.Badge

	statuses    *BadgeStatusCache
	badgeTypes  *BadgeTypeCache
	cardholders *CardholderCache
	assignments *AssignmentCache
	levels      *AccessLevelCache
}

func NewBadgeCache() *BadgeCache {
	return &BadgeCache{
		byID:         make(map[int64]*model.Badge),
		byKey:        make(map[int32]*model.Badge),
		byCardholder: make(map[int32][]*model.Badge),
	}
}

func (c *BadgeCache) GetItems() []*model.Badge {
	return c.list
}

func (c *BadgeCache) Resolve(
	statuses *BadgeStatusCache,
	types *BadgeTypeCache,
	cardholders *CardholderCache,
	assignments *AssignmentCache,
	levels *AccessLevelCache,
) {
	c.statuses, c.badgeTypes, c.cardholders, c.assignments, c.levels =
		statuses, types, cardholders, assignments, levels
}

func (c *BadgeCache) GetRowItems() []model.RowObject {
	result := make([]model.RowObject, len(c.list))

	for i, b := range c.list {
		row := badgeRow{
			id:         strconv.FormatInt(b.ID, 10),
			key:        strconv.FormatInt(int64(b.Key), 10),
			activate:   date.Format(b.Activate),
			deactivate: date.Format(b.Deactivate),
		}

		if c.statuses != nil {
			if s := c.statuses.byID[b.Status]; s != nil {
				row.status = s.Name
			}
		}

		if c.badgeTypes != nil {
			if t := c.badgeTypes.byID[b.Type]; t != nil {
				row.badgeType = t.Name
			}
		}

		if c.cardholders != nil {
			if ch := c.cardholders.byID[b.Cardholder]; ch != nil {
				row.ssno = ch.SSNO
			}
		}

		if c.assignments != nil && c.levels != nil {
			for j, a := range c.assignments.byBadgeKey[b.Key] {
				if j >= 6 {
					break
				}
				if al := c.levels.byID[a.AccessLevel]; al != nil {
					row.levels[j] = al.Name
				}
			}
		}

		result[i] = row
	}
	return result
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
			"PERSONID":   b.Cardholder,
		},
	)

	if err != nil {
		log.Printf("Failed to create Lnl_Badge: %v", err)
		return 0, err
	}

	key := json.PropToInt32(props, "BADGEKEY")
	if key == 0 {
		log.Printf("Failed to retrieve badge key for new Lnl_Badge with ID %v", b.ID)
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

func (c *BadgeCache) RowHeader() []string {
	return []string{"ID", "Badge Key", "Activate", "Deactivate", "Status", "Type", "Cardholder SSNO", "Access Level 1", "Access Level 2", "Access Level 3", "Access Level 4", "Access Level 5", "Access Level 6"}
}
