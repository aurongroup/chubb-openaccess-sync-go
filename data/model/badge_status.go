package model

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

func NewBadgeStatus(id int, name string) (*BadgeStatus, error) {
	if id == 0 {
		return nil, ErrBadgeStatusMissingID
	}

	if name == "" {
		return nil, ErrBadgeStatusMissingName
	}

	return &BadgeStatus{
			ID:   id,
			Name: name,
		},
		nil
}

func NewBadgeStatusFromJSON(props map[string]any) (*BadgeStatus, error) {
	return NewBadgeStatus(
		json.PropToInt(props, "ID"),
		json.PropToStr(props, "Name"),
	)
}
