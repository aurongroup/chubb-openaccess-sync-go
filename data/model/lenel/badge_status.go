package lenel

import (
	"errors"
	"openaccess-sync/util/json"
)

var (
	ErrBadgeStatusMissingID   = errors.New("badge status: missing required ID")
	ErrBadgeStatusMissingName = errors.New("badge status: missing required Name")
)

// BadgeStatus represents a badge status from the OpenAccess API.
type BadgeStatus struct {
	ID   int
	Name string
}

func NewBadgeStatus(props map[string]any) (*BadgeStatus, error) {
	id := json.PropToInt(props, "ID")
	if id == 0 {
		return nil, ErrBadgeStatusMissingID
	}

	name := json.PropToStr(props, "Name")
	if name == "" {
		return nil, ErrBadgeStatusMissingName
	}

	return &BadgeStatus{ID: id, Name: name}, nil
}
