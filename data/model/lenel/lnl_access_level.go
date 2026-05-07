package lenel

import (
	"openaccess-sync/data"
	"openaccess-sync/util/json"
)

// LnlAccessLevel represents an access level from the OpenAccess API.
type LnlAccessLevel struct {
	ID   int
	Name string
}

func NewLnlAccessLevel(props map[string]any) (*LnlAccessLevel, error) {
	id := json.PropToInt(props, "ID")
	if id == 0 {
		return nil, data.ErrAccessLevelMissingID
	}

	name := json.PropToStr(props, "Name")
	if name == "" {
		return nil, data.ErrAccessLevelMissingName
	}

	return &LnlAccessLevel{ID: id, Name: name}, nil
}
