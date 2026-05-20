package lenel

import (
	"fmt"
	"openaccess-sync/client"
	"openaccess-sync/data/model"
	"strings"
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
