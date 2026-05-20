package lenel

import "openaccess-sync/data/model"

type CardholderCache struct {
	list []*model.Cardholder
	byID map[int32]*model.Cardholder
}

func NewCardholderCache() CardholderCache {
	return CardholderCache{
		byID: make(map[int32]*model.Cardholder),
	}
}

// FIXME
//func (c *CardholderCache) Fill(cl *client.Client) error {
//	items, err := cl.GetInstancesWithProgress("Lnl_Cardholder", "")
//	if err != nil {
//		return err
//	}
//
//	for _, props := range items {
//		ch, err := model.NewCardholderFromJSON(props)
//		if err != nil {
//			log.Printf("skipping Lnl_Cardholder: %v", err)
//			continue
//		}
//
//		c.list = append(c.list, ch)
//		c.byID[ch.ID] = ch
//	}
//	log.Printf("Retrieved %d Lnl_Cardholder records", len(c.list))
//	return nil
//}
