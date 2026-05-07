package lenel

import "openaccess-sync/data"

// LnlBadgeType represents a badge type from the OpenAccess API.
type LnlBadgeType struct {
	ID   int
	Name string
}

func NewLnlBadgeType(props map[string]any) (*LnlBadgeType, error) {
	id := json.PropToInt(props, "ID")
	if id == 0 {
		return nil, data.ErrBadgeTypeMissingID
	}

	name := json.PropToStr(props, "Name")
	if name == "" {
		return nil, data.ErrBadgeTypeMissingName
	}

	return &LnlBadgeType{ID: id, Name: name}, nil
}
