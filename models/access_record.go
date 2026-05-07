package models

import (
	"fmt"
	"strings"
	"time"
	"unicode"
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
	SyncStatus    SyncStatus
	CardholderKey string
}

func generateCardholderKey(ssno, first, last string) string {
	str := fmt.Sprintf("%s%s%s", ssno, first, last)
	return strings.ToUpper(strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str))
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
func (r *AccessRecord) ToRow() []string {
	return []string{
		r.SSNO,
		r.First,
		r.Last,
		r.AccLvl1,
		r.AccLvl2,
		r.AccLvl3,
		r.AccLvl4,
		r.AccLvl5,
		r.AccLvl6,
		r.BadgeID,
		FormatDate(r.Activate),
		FormatDate(r.Deactivate),
		r.Status,
		r.BadgeType,
	}
}
