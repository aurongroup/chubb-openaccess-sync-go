package main

import (
	"errors"
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
		formatDate(r.Activate),
		formatDate(r.Deactivate),
		r.Status,
		r.BadgeType,
	}
}

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

// ToAccessRecord builds an AccessRecord from an LnlBadge and its access levels.
func (badge *LnlBadge) ToAccessRecord(accessLevels []*LnlAccessLevel) (*AccessRecord, error) {
	var ssno, first, last, status, badgeType string

	if badge.Cardholder != nil {
		ssno = badge.Cardholder.SSNO
		first = badge.Cardholder.FirstName
		last = badge.Cardholder.LastName
	}

	lvl := [6]string{}
	for i := 0; i < len(accessLevels) && i < 6; i++ {
		lvl[i] = accessLevels[i].Name
	}

	if badge.Status != nil {
		status = badge.Status.Name
	}

	if badge.Type != nil {
		badgeType = badge.Type.Name
	}

	return NewAccessRecord(
		ssno, first, last,
		lvl[0], lvl[1], lvl[2], lvl[3], lvl[4], lvl[5],
		fmt.Sprintf("%d", badge.ID),
		badge.Activate, badge.Deactivate,
		status, badgeType,
	)
}

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

// LnlBadge represents a badge from the OpenAccess API.
type LnlBadge struct {
	ID         int
	BadgeKey   int
	Activate   *time.Time
	Deactivate *time.Time
	Status     *LnlBadgeStatus
	Type       *LnlBadgeType
	Cardholder *LnlCardholder
}

func NewLnlBadge(props map[string]any, cache *DataCache) (*LnlBadge, error) {
	if cache == nil {
		return nil, ErrBadgeNilCache
	}

	id := propInt(props, "ID")
	if id == 0 {
		return nil, ErrBadgeMissingID
	}

	badgeKey := propInt(props, "BADGEKEY")
	if badgeKey == 0 {
		return nil, ErrBadgeMissingBadgeKey
	}

	b := &LnlBadge{
		ID:         id,
		BadgeKey:   badgeKey,
		Activate:   propDate(props, "ACTIVATE"),
		Deactivate: propDate(props, "DEACTIVATE"),
	}

	statusID := propInt(props, "STATUS")
	if statusID == 0 {
		return nil, ErrBadgeUnresolvedStatus
	}

	b.Status = cache.GetBadgeStatus(statusID)
	if b.Status == nil {
		return nil, ErrBadgeUnresolvedStatus
	}

	typeID := propInt(props, "TYPE")
	if typeID == 0 {
		return nil, ErrBadgeUnresolvedType
	}

	b.Type = cache.GetBadgeType(typeID)
	if b.Type == nil {
		return nil, ErrBadgeUnresolvedType
	}

	if personID := propInt(props, "PERSONID"); personID != 0 {
		b.Cardholder = cache.GetCardholder(personID)
	}

	return b, nil
}

// ToJSON returns the API wire format map for a badge.
func (badge *LnlBadge) ToJSON() map[string]any {
	return map[string]any{
		"type_name": "Lnl_Badge",
		"property_value_map": map[string]any{
			"badgeKey":   badge.BadgeKey,
			"activate":   dateStr(badge.Activate),
			"deactivate": dateStr(badge.Deactivate),
		},
	}
}

// Sentinel errors returned by model constructors.
var (
	// AccessRecord
	ErrAccessRecordMissingLast      = errors.New("access record: missing required Last")
	ErrAccessRecordMissingBadgeID   = errors.New("access record: missing required BadgeID")
	ErrAccessRecordMissingStatus    = errors.New("access record: missing required Status")
	ErrAccessRecordMissingBadgeType = errors.New("access record: missing required BadgeType")

	// LnlBadgeStatus
	ErrBadgeStatusMissingID   = errors.New("badge status: missing required ID")
	ErrBadgeStatusMissingName = errors.New("badge status: missing required Name")

	// LnlBadgeType
	ErrBadgeTypeMissingID   = errors.New("badge type: missing required ID")
	ErrBadgeTypeMissingName = errors.New("badge type: missing required Name")

	// LnlAccessLevel
	ErrAccessLevelMissingID   = errors.New("access level: missing required ID")
	ErrAccessLevelMissingName = errors.New("access level: missing required Name")

	// LnlCardholder
	ErrCardholderMissingIdentifier = errors.New("cardholder: must have an ID or SSNO")
	ErrCardholderMissingLastName   = errors.New("cardholder: missing required LastName")

	// LnlBadge
	ErrBadgeNilCache         = errors.New("badge: cache is nil")
	ErrBadgeMissingID        = errors.New("badge: missing required ID")
	ErrBadgeMissingBadgeKey  = errors.New("badge: missing required BadgeKey")
	ErrBadgeUnresolvedStatus = errors.New("badge: STATUS not found in cache")
	ErrBadgeUnresolvedType   = errors.New("badge: TYPE not found in cache")

	// LnlAccessLevelAssignment
	ErrAssignmentNilCache              = errors.New("assignment: cache is nil")
	ErrAssignmentUnresolvedAccessLevel = errors.New("assignment: AccessLevelID not found in cache")
	ErrAssignmentUnresolvedBadge       = errors.New("assignment: BadgeKey not found in cache")
)

// LnlAccessLevelAssignment links an access level to a badge.
type LnlAccessLevelAssignment struct {
	AccessLevel *LnlAccessLevel
	Badge       *LnlBadge
}

func NewLnlAccessLevelAssignment(props map[string]any, cache *DataCache) (*LnlAccessLevelAssignment, error) {
	if cache == nil {
		return nil, ErrAssignmentNilCache
	}

	a := &LnlAccessLevelAssignment{}

	if alID := propInt(props, "AccessLevelID"); alID != 0 {
		a.AccessLevel = cache.GetAccessLevel(alID)
	}

	if a.AccessLevel == nil {
		return nil, ErrAssignmentUnresolvedAccessLevel
	}

	if badgeKey := propInt(props, "BadgeKey"); badgeKey != 0 {
		a.Badge = cache.GetBadgeByKey(badgeKey)
	}

	if a.Badge == nil {
		return nil, ErrAssignmentUnresolvedBadge
	}

	return a, nil
}
