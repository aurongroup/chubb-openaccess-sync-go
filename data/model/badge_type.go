package model

import (
	"errors"
	"openaccess-sync/util/json"
)

// BadgeType represents a badge type from the OpenAccess API.
type BadgeType struct {
	ID   int
	Name string
}

var (
	ErrBadgeTypeMissingID   = errors.New("badge type: missing required ID")
	ErrBadgeTypeMissingName = errors.New("badge type: missing required Name")
)

func NewBadgeType(id int, name string) (*BadgeType, error) {
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
		json.PropToInt(props, "ID"),
		json.PropToStr(props, "Name"),
	)
}
