package lenel

import (
	"openaccess-sync/data"
	"openaccess-sync/util/json"
)

// LnlBadgeStatus represents a badge status from the OpenAccess API.
type LnlBadgeStatus struct {
	ID   int
	Name string
}

func NewLnlBadgeStatus(props map[string]any) (*LnlBadgeStatus, error) {
	id := json.PropToInt(props, "ID")
	if id == 0 {
		return nil, data.ErrBadgeStatusMissingID
	}

	name := json.PropToStr(props, "Name")
	if name == "" {
		return nil, data.ErrBadgeStatusMissingName
	}

	return &LnlBadgeStatus{ID: id, Name: name}, nil
}
