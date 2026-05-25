package lenel

import (
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/data/model"
)

type CardholderCache struct {
	list       []*model.Cardholder
	byID       map[int32]*model.Cardholder
	byKey      map[string]*model.Cardholder
	duplicates map[string][]*model.Cardholder
}

func NewCardholderCache() CardholderCache {
	return CardholderCache{
		byID:       make(map[int32]*model.Cardholder),
		byKey:      make(map[string]*model.Cardholder),
		duplicates: make(map[string][]*model.Cardholder),
	}
}

func (c *CardholderCache) GetItems() []*model.Cardholder {
	return c.list
}

func (c *CardholderCache) GetDuplicates() map[string][]*model.Cardholder {
	return c.duplicates
}

func (c *CardholderCache) Fill(cl *client.Client) error {
	items, err := cl.GetInstancesWithProgress("Lnl_Cardholder", "")
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
		c.byKey[ch.GetKey()] = ch
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
		c.byKey[ch.GetKey()] = ch
	}
	log.Printf("Retrieved %d Lnl_Cardholder detached records", len(c.list))
	return nil
}

func (c *CardholderCache) FillNoSSNO(cl *client.Client) error {
	items, err := cl.GetInstancesWithProgress("Lnl_Cardholder", "SSNO=null")
	if err != nil {
		return err
	}

	for _, props := range items {
		ch, err := model.NewCardholderFromJSON(props)
		if err != nil {
			log.Printf("skipping Lnl_Cardholder: %v", err)
			continue
		}

		// For cardholders without an SSNO, we might not have a unique key
		if original, ok := c.byKey[ch.GetKey()]; ok {
			if _, ok := c.duplicates[ch.GetKey()]; !ok {
				c.duplicates[ch.GetKey()] = []*model.Cardholder{}
				c.duplicates[ch.GetKey()] = append(c.duplicates[ch.GetKey()], original)
			}

			c.duplicates[ch.GetKey()] = append(c.duplicates[ch.GetKey()], ch)
			//log.Printf("skipping Lnl_Cardholder: duplicate key %s: %d", ch.GetKey(), ch.ID)
			continue
		}
		c.list = append(c.list, ch)
		c.byID[ch.ID] = ch
		c.byKey[ch.GetKey()] = ch
	}
	log.Printf("Retrieved %d Lnl_Cardholder records", len(c.list))
	return nil
}
