package model

import (
	"errors"
	"openaccess-sync/util/json"
)

var (
	ErrAccessLevelMissingID   = errors.New("access level: missing required ID")
	ErrAccessLevelMissingName = errors.New("access level: missing required Name")
)

// AccessLevel represents an access level from the OpenAccess API.
type AccessLevel struct {
	ID   int32
	Name string
}

func NewAccessLevel(id int32, name string) (*AccessLevel, error) {
	if id == 0 {
		return nil, ErrAccessLevelMissingID
	}

	if name == "" {
		return nil, ErrAccessLevelMissingName
	}

	return &AccessLevel{
			ID:   id,
			Name: name,
		},
		nil
}

func NewAccessLevelFromJSON(props map[string]any) (*AccessLevel, error) {
	return NewAccessLevel(
		json.PropToInt32(props, "ID"),
		json.PropToStr(props, "Name"),
	)
}
