package lenel

import (
	"errors"
	"openaccess-sync/util/json"
)

var (
	ErrCardholderMissingIdentifier = errors.New("cardholder: missing required identifier (ID or SSNO)")
	ErrCardholderMissingLastName   = errors.New("cardholder: missing required LastName")
)

// Cardholder represents a cardholder from the OpenAccess API.
type Cardholder struct {
	ID        int
	FirstName string
	LastName  string
	SSNO      string
}

func NewCardholder(props map[string]any) (*Cardholder, error) {
	id := json.PropToInt(props, "ID")
	ssno := json.PropToStr(props, "SSNO")
	if id == 0 && ssno == "" {
		return nil, ErrCardholderMissingIdentifier
	}

	lastName := json.PropToStr(props, "LASTNAME")
	if lastName == "" {
		return nil, ErrCardholderMissingLastName
	}

	return &Cardholder{
		ID:        id,
		FirstName: json.PropToStr(props, "FIRSTNAME"),
		LastName:  lastName,
		SSNO:      ssno,
	}, nil
}
