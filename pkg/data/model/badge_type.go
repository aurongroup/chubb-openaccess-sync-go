package model

import (
	"errors"
	"openaccess-sync/pkg/util/json"
	"strconv"
)

// BadgeType represents a badge type from the OpenAccess API.
type BadgeType struct {
	ID   int32
	Name string
}

var (
	ErrBadgeTypeMissingID   = errors.New("badge type: missing required ID")
	ErrBadgeTypeMissingName = errors.New("badge type: missing required Name")
)

func NewBadgeType(id int32, name string) (*BadgeType, error) {
	if id == 0 {
		return nil, ErrBadgeTypeMissingID
	}

	if name == "" {
		return nil, ErrBadgeTypeMissingName
	}

	return &BadgeType{
			ID:   id,
			Name: name,
		},
		nil
}

func NewBadgeTypeFromJSON(props map[string]any) (*BadgeType, error) {
	return NewBadgeType(
		json.PropToInt32(props, "ID"),
		json.PropToStr(props, "Name"),
	)
}

func (bt *BadgeType) ToRow() []string {
	return []string{strconv.FormatInt(int64(bt.ID), 10), bt.Name}
}
