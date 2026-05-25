package model

import (
	"errors"
	"fmt"
	"openaccess-sync/pkg/util/date"
	stru "openaccess-sync/pkg/util/strings"
	"strings"
	"time"
)

var (
	ErrAccessRecordMissingLast      = errors.New("access record: missing required Last")
	ErrAccessRecordMissingBadgeID   = errors.New("access record: missing required BadgeID")
	ErrAccessRecordMissingStatus    = errors.New("access record: missing required Status")
	ErrAccessRecordMissingBadgeType = errors.New("access record: missing required BadgeType")
)

// AccessRecord represents a single row in the pipe-delimited access control CSV.
type AccessRecord struct {
	SSNO          string
	First         string
	Last          string
	AccLvl1       string
	AccLvl2       string
	AccLvl3       string
	AccLvl4       string
	AccLvl5       string
	AccLvl6       string
	BadgeID       string
	Activate      *time.Time
	Deactivate    *time.Time
	Status        string
	BadgeType     string
	CardholderKey string
	RecordKey     string
}

func (a *AccessRecord) GetKey() string {
	str := fmt.Sprintf("%s%s%s", a.SSNO, a.First, a.Last)
	return strings.ToUpper(stru.Clean(str))
}

func generateCardholderKey(ssno, first, last string) string {
	str := fmt.Sprintf("%s%s%s", ssno, first, last)
	return strings.ToUpper(stru.Clean(str))
}

// NewAccessRecord constructs and validates an AccessRecord.
// last, badgeID, status, and badgeType are required.
func NewAccessRecord(
	ssno, first, last string,
	accLvl1, accLvl2, accLvl3, accLvl4, accLvl5, accLvl6 string,
	badgeID string,
	activate, deactivate *time.Time,
	status, badgeType string,
) (*AccessRecord, error) {
	if last == "" {
		return nil, ErrAccessRecordMissingLast
	}

	if badgeID == "" {
		return nil, ErrAccessRecordMissingBadgeID
	}

	if status == "" {
		return nil, ErrAccessRecordMissingStatus
	}

	if badgeType == "" {
		return nil, ErrAccessRecordMissingBadgeType
	}

	return &AccessRecord{
		SSNO:          ssno,
		First:         first,
		Last:          last,
		AccLvl1:       accLvl1,
		AccLvl2:       accLvl2,
		AccLvl3:       accLvl3,
		AccLvl4:       accLvl4,
		AccLvl5:       accLvl5,
		AccLvl6:       accLvl6,
		BadgeID:       badgeID,
		Activate:      activate,
		Deactivate:    deactivate,
		Status:        status,
		BadgeType:     badgeType,
		CardholderKey: generateCardholderKey(ssno, first, last),
	}, nil
}

// ToRow converts an AccessRecord to a slice of strings for CSV output.
func (a *AccessRecord) ToRow() []string {
	return []string{
		a.SSNO,
		a.First,
		a.Last,
		a.AccLvl1,
		a.AccLvl2,
		a.AccLvl3,
		a.AccLvl4,
		a.AccLvl5,
		a.AccLvl6,
		a.BadgeID,
		date.Format(a.Activate),
		date.Format(a.Deactivate),
		a.Status,
		a.BadgeType,
	}
}
