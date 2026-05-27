package lenel

import (
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/data/model"
)

type CardholderCache struct {
	list []*model.Cardholder
	byID map[int32]*model.Cardholder
}

func NewCardholderCache() CardholderCache {
	return CardholderCache{
		byID: make(map[int32]*model.Cardholder),
	}
}

func (c *CardholderCache) GetItems() []*model.Cardholder {
	return c.list
}

func (c *CardholderCache) GetByID(id int32) *model.Cardholder {
	ch, ok := c.byID[id]

	if ok {
		return ch
	}

	return nil
}

func (c *CardholderCache) Fill(cl *client.Client) error {
	return c.FillWithFilter(cl, "")
}

func (c *CardholderCache) FillWithFilter(cl *client.Client, filter string) error {
	items, err := cl.GetInstancesWithProgress("Lnl_Cardholder", filter)
	if err != nil {
		return err
	}

	for _, props := range items {
		ch, err := model.NewCardholderFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_Cardholder: %v", err)
			continue
		}

		c.list = append(c.list, ch)
		c.byID[ch.ID] = ch
	}
	log.Printf("Retrieved %d Lnl_Cardholder records", len(c.list))
	return nil
}

func (c *CardholderCache) FillDetached(cl *client.Client) error {
	items, err := cl.GetCardholdersWithProgress(true, "")
	if err != nil {
		return err
	}

	for _, props := range items {
		ch, err := model.NewCardholderFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_Cardholder: %v", err)
			continue
		}

		c.list = append(c.list, ch)
		c.byID[ch.ID] = ch
	}
	log.Printf("Retrieved %d Lnl_Cardholder detached records", len(c.list))
	return nil
}

func (c *CardholderCache) Create(cl *client.Client, ch *model.Cardholder) (int32, error) {
	id, err := cl.CreateInstance(
		"Lnl_Cardholder",
		map[string]interface{}{
			"FIRSTNAME": ch.FirstName,
			"LASTNAME":  ch.LastName,
			"SSNO":      ch.SSNO,
		},
	)

	if err != nil {
		return 0, err
	}

	ch.ID = id

	return id, nil
}

func (c *CardholderCache) Update(cl *client.Client, ch *model.Cardholder) error {

	return cl.UpdateInstance(
		"Lnl_Cardholder",
		map[string]interface{}{
			"ID":        ch.ID,
			"FIRSTNAME": ch.FirstName,
			"LASTNAME":  ch.LastName,
			"SSNO":      ch.SSNO,
		},
	)
}

func (c *CardholderCache) Delete(cl *client.Client, ch *model.Cardholder) error {
	return cl.DeleteInstance(
		"Lnl_Cardholder",
		map[string]interface{}{
			"ID": ch.ID,
		},
	)
}
