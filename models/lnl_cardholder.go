package models

// LnlCardholder represents a cardholder from the OpenAccess API.
type LnlCardholder struct {
	ID        int
	FirstName string
	LastName  string
	SSNO      string
}

func NewLnlCardholder(props map[string]any) (*LnlCardholder, error) {
	id := propInt(props, "ID")
	ssno := propStr(props, "SSNO")
	if id == 0 && ssno == "" {
		return nil, ErrCardholderMissingIdentifier
	}

	lastName := propStr(props, "LASTNAME")
	if lastName == "" {
		return nil, ErrCardholderMissingLastName
	}

	return &LnlCardholder{
		ID:        id,
		FirstName: propStr(props, "FIRSTNAME"),
		LastName:  lastName,
		SSNO:      ssno,
	}, nil
}
