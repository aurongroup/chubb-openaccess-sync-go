package models

// LnlBadgeType represents a badge type from the OpenAccess API.
type LnlBadgeType struct {
	ID   int
	Name string
}

func NewLnlBadgeType(props map[string]any) (*LnlBadgeType, error) {
	id := propInt(props, "ID")
	if id == 0 {
		return nil, ErrBadgeTypeMissingID
	}

	name := propStr(props, "Name")
	if name == "" {
		return nil, ErrBadgeTypeMissingName
	}

	return &LnlBadgeType{ID: id, Name: name}, nil
}
