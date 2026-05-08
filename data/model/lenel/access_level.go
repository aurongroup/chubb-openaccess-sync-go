package lenel

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
	ID   int
	Name string
}

func NewAccessLevel(props map[string]any) (*AccessLevel, error) {
	id := json.PropToInt(props, "ID")
	if id == 0 {
		return nil, ErrAccessLevelMissingID
	}

	name := json.PropToStr(props, "Name")
	if name == "" {
		return nil, ErrAccessLevelMissingName
	}

	return &AccessLevel{ID: id, Name: name}, nil
}
