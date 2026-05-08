package lenel

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

func NewBadgeType(props map[string]any) (*BadgeType, error) {
	id := json.PropToInt(props, "ID")
	if id == 0 {
		return nil, ErrBadgeTypeMissingID
	}

	name := json.PropToStr(props, "Name")
	if name == "" {
		return nil, ErrBadgeTypeMissingName
	}

	return &BadgeType{ID: id, Name: name}, nil
}
