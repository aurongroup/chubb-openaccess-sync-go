package lenel

import (
	"openaccess-sync/data"
	"openaccess-sync/util/json"
)

// LnlCardholder represents a cardholder from the OpenAccess API.
type LnlCardholder struct {
	ID        int
	FirstName string
	LastName  string
	SSNO      string
}

func NewLnlCardholder(props map[string]any) (*LnlCardholder, error) {
	id := json.PropToInt(props, "ID")
	ssno := json.PropToStr(props, "SSNO")
	if id == 0 && ssno == "" {
		return nil, data.ErrCardholderMissingIdentifier
	}

	lastName := json.PropToStr(props, "LASTNAME")
	if lastName == "" {
		return nil, data.ErrCardholderMissingLastName
	}

	return &LnlCardholder{
		ID:        id,
		FirstName: json.PropToStr(props, "FIRSTNAME"),
		LastName:  lastName,
		SSNO:      ssno,
	}, nil
}
