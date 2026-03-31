package main

import (
	"fmt"
	"time"
)

// SyncStatus represents the result of comparing a CSV record against the API.
type SyncStatus int

const (
	SyncNew      SyncStatus = iota
	SyncExisting SyncStatus = iota
	SyncUpdate   SyncStatus = iota
	SyncDelete   SyncStatus = iota
)

func (s SyncStatus) String() string {
	switch s {
	case SyncNew:
		return "new"
	case SyncExisting:
		return "existing"
	case SyncUpdate:
		return "update"
	case SyncDelete:
		return "delete"
	}
	return "unknown"
}

// accessRecordFromBadge builds an AccessRecord from an API badge and its access levels.
func accessRecordFromBadge(badge *LnlBadge, accessLevels []*LnlAccessLevel) AccessRecord {
	var ssno, first, last string
	if badge.Cardholder != nil {
		ssno = badge.Cardholder.SSNO
		first = badge.Cardholder.FirstName
		last = badge.Cardholder.LastName
	}

	lvl := [6]string{}
	for i := 0; i < len(accessLevels) && i < 6; i++ {
		lvl[i] = accessLevels[i].Name
	}

	var status, badgeType string
	if badge.Status != nil {
		status = badge.Status.Name
	}
	if badge.Type != nil {
		badgeType = badge.Type.Name
	}

	return AccessRecord{
		SSNO:         ssno,
		First:        first,
		Last:         last,
		AccLvl1:      lvl[0],
		AccLvl2:      lvl[1],
		AccLvl3:      lvl[2],
		AccLvl4:      lvl[3],
		AccLvl5:      lvl[4],
		AccLvl6:      lvl[5],
		BadgeID:      fmt.Sprintf("%d", badge.ID),
		Activate:     badge.Activate,
		Deactivate:   badge.Deactivate,
		Status:       status,
		BadgeType:    badgeType,
		Badge:        badge,
		AccessLevels: accessLevels,
	}
}

// ContentEquals returns true if two AccessRecords have identical content across all 14 fields.
func ContentEquals(a, b AccessRecord) bool {
	return a.SSNO == b.SSNO &&
		a.First == b.First &&
		a.Last == b.Last &&
		a.AccLvl1 == b.AccLvl1 &&
		a.AccLvl2 == b.AccLvl2 &&
		a.AccLvl3 == b.AccLvl3 &&
		a.AccLvl4 == b.AccLvl4 &&
		a.AccLvl5 == b.AccLvl5 &&
		a.AccLvl6 == b.AccLvl6 &&
		a.BadgeID == b.BadgeID &&
		dateEqual(a.Activate, b.Activate) &&
		dateEqual(a.Deactivate, b.Deactivate) &&
		a.Status == b.Status &&
		a.BadgeType == b.BadgeType
}

func dateEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

// CompareRecords classifies each record in first against second:
// NEW if absent from second, EXISTING if content matches, UPDATE if different.
// If a record in second is not in first it is marked DELETE.
func CompareRecords(first, second []AccessRecord) []AccessRecord {
	secondByID := make(map[string]AccessRecord, len(second))
	for _, r := range second {
		secondByID[r.BadgeID] = r
	}

	firstIDs := make(map[string]struct{}, len(first))
	var result []AccessRecord

	for _, r := range first {
		firstIDs[r.BadgeID] = struct{}{}
		if s, ok := secondByID[r.BadgeID]; !ok {
			r.SyncStatus = SyncNew
		} else if ContentEquals(r, s) {
			r.SyncStatus = SyncExisting
		} else {
			r.SyncStatus = SyncUpdate
		}
		result = append(result, r)
	}

	for _, r := range second {
		if _, ok := firstIDs[r.BadgeID]; !ok {
			r.SyncStatus = SyncDelete
			result = append(result, r)
		}
	}

	return result
}
