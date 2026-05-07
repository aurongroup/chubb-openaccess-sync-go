package models

// LnlAccessLevel represents an access level from the OpenAccess API.
type LnlAccessLevel struct {
	ID   int
	Name string
}

func NewLnlAccessLevel(props map[string]any) (*LnlAccessLevel, error) {
	id := propInt(props, "ID")
	if id == 0 {
		return nil, ErrAccessLevelMissingID
	}

	name := propStr(props, "Name")
	if name == "" {
		return nil, ErrAccessLevelMissingName
	}

	return &LnlAccessLevel{ID: id, Name: name}, nil
}
