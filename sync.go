package main

import (
	"fmt"
	"io"
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

// ContentEquals returns true if two AccessRecords have identical content across all 14 fields.
func ContentEquals(a, b *AccessRecord, w io.Writer) bool {
	var diffs []string

	if a.SSNO != b.SSNO {
		diffs = append(diffs, fmt.Sprintf("SSNO: %q != %q", a.SSNO, b.SSNO))
	}
	if a.First != b.First {
		diffs = append(diffs, fmt.Sprintf("First: %q != %q", a.First, b.First))
	}
	if a.Last != b.Last {
		diffs = append(diffs, fmt.Sprintf("Last: %q != %q", a.Last, b.Last))
	}
	if a.AccLvl1 != b.AccLvl1 {
		diffs = append(diffs, fmt.Sprintf("AccLvl1: %q != %q", a.AccLvl1, b.AccLvl1))
	}
	if a.AccLvl2 != b.AccLvl2 {
		diffs = append(diffs, fmt.Sprintf("AccLvl2: %q != %q", a.AccLvl2, b.AccLvl2))
	}
	if a.AccLvl3 != b.AccLvl3 {
		diffs = append(diffs, fmt.Sprintf("AccLvl3: %q != %q", a.AccLvl3, b.AccLvl3))
	}
	if a.AccLvl4 != b.AccLvl4 {
		diffs = append(diffs, fmt.Sprintf("AccLvl4: %q != %q", a.AccLvl4, b.AccLvl4))
	}
	if a.AccLvl5 != b.AccLvl5 {
		diffs = append(diffs, fmt.Sprintf("AccLvl5: %q != %q", a.AccLvl5, b.AccLvl5))
	}
	if a.AccLvl6 != b.AccLvl6 {
		diffs = append(diffs, fmt.Sprintf("AccLvl6: %q != %q", a.AccLvl6, b.AccLvl6))
	}
	if a.BadgeID != b.BadgeID {
		diffs = append(diffs, fmt.Sprintf("BadgeID: %q != %q", a.BadgeID, b.BadgeID))
	}
	if !dateEqual(a.Activate, b.Activate) {
		diffs = append(diffs, fmt.Sprintf("Activate: %v != %v", a.Activate, b.Activate))
	}
	if !dateEqual(a.Deactivate, b.Deactivate) {
		diffs = append(diffs, fmt.Sprintf("Deactivate: %v != %v", a.Deactivate, b.Deactivate))
	}
	if a.Status != b.Status {
		diffs = append(diffs, fmt.Sprintf("Status: %q != %q", a.Status, b.Status))
	}
	if a.BadgeType != b.BadgeType {
		diffs = append(diffs, fmt.Sprintf("BadgeType: %q != %q", a.BadgeType, b.BadgeType))
	}

	if len(diffs) > 0 {
		if w != nil {
			fmt.Fprintf(w, "ContentEquals differences for BadgeID %s: %v\n", a.BadgeID, diffs)
		}
		return false
	}

	return true
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

// CompareResult holds the output of CompareRecords grouped by sync status.
// All contains every record across all buckets in insertion order (New, Existing, Update, Delete).
type CompareResult struct {
	New      []*AccessRecord
	Existing []*AccessRecord
	Update   []*AccessRecord
	Delete   []*AccessRecord
	All      []*AccessRecord
}

// CompareRecords classifies each record in first against second:
// NEW if absent from second, EXISTING if content matches, UPDATE if different.
// If a record in second is not in first it is marked DELETE.
func CompareRecords(first, second []*AccessRecord, w io.Writer) CompareResult {
	secondByID := make(map[string]*AccessRecord, len(second))
	for _, r := range second {
		secondByID[r.BadgeID] = r
	}

	firstIDs := make(map[string]struct{}, len(first))
	var result CompareResult

	for _, r := range first {
		firstIDs[r.BadgeID] = struct{}{}

		if s, ok := secondByID[r.BadgeID]; !ok {
			r.SyncStatus = SyncNew
			result.New = append(result.New, r)
		} else if ContentEquals(r, s, w) {
			r.SyncStatus = SyncExisting
			result.Existing = append(result.Existing, r)
		} else {
			r.SyncStatus = SyncUpdate
			result.Update = append(result.Update, r)
		}
		result.All = append(result.All, r)
	}

	for _, r := range second {
		if _, ok := firstIDs[r.BadgeID]; !ok {
			r.SyncStatus = SyncDelete
			result.Delete = append(result.Delete, r)
			result.All = append(result.All, r)
		}
	}

	return result
}
