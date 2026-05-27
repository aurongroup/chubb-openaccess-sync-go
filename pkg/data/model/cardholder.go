package model

import (
	"errors"
	"openaccess-sync/pkg/util/json"
)

var (
	ErrCardholderMissingLastName = errors.New("cardholder: missing required LastName")
)

// Cardholder represents a cardholder from the OpenAccess API.
type Cardholder struct {
	ID          int32
	FirstName   string
	LastName    string
	SSNO        string
	OfficePhone string
}

func NewCardholder(id int32, ssno, firstName, lastName string) (*Cardholder, error) {
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
		json.PropToInt32(props, "ID"),
		json.PropToStr(props, "SSNO"),
		json.PropToStr(props, "FIRSTNAME"),
		json.PropToStr(props, "LASTNAME"),
	)
}

func NewCardholderFromAccessRecord(a *AccessRecord) (*Cardholder, error) {
	return NewCardholder(
		0,
		a.SSNO,
		a.First,
		a.Last,
	)
}
