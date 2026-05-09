package model

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

func NewCardholder(id int, ssno, firstName, lastName string) (*Cardholder, error) {
	if lastName == "" {
		return nil, ErrCardholderMissingLastName
	}

	return &Cardholder{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		SSNO:      ssno,
	}, nil
}

func NewCardholderFromJSON(props map[string]any) (*Cardholder, error) {
	return NewCardholder(
		json.PropToInt(props, "ID"),
		json.PropToStr(props, "FIRSTNAME"),
		json.PropToStr(props, "LASTNAME"),
		json.PropToStr(props, "SSNO"),
	)
}
