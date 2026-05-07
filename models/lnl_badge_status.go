package models

// LnlBadgeStatus represents a badge status from the OpenAccess API.
type LnlBadgeStatus struct {
	ID   int
	Name string
}

func NewLnlBadgeStatus(props map[string]any) (*LnlBadgeStatus, error) {
	id := propInt(props, "ID")
	if id == 0 {
		return nil, ErrBadgeStatusMissingID
	}

	name := propStr(props, "Name")
	if name == "" {
		return nil, ErrBadgeStatusMissingName
	}

	return &LnlBadgeStatus{ID: id, Name: name}, nil
}
