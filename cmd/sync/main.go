package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/config"
	"openaccess-sync/pkg/csv"
	"openaccess-sync/pkg/data/lenel"
	"openaccess-sync/pkg/data/model"
	"os"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatalf("Error parsing command line arguments: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	cl, err := client.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}
	defer func() {
		if err := cl.Close(); err != nil {
			log.Printf("Failed to close client session: %v", err)
		}
	}()

	//cache := lenel.NewDataCache(cl)

	//if err := cache.Fill(); err != nil {
	//	log.Fatalf("Failed to load API data: %v", err)
	//}

	/*
		1. Retrieve status, type, and access levels
		2. Parse CSV and validate status, type, and access levels - fail if missing
		3. Retrieve cardholders
		4. Process cardholders - remove missing, update first/last as required (only for those with SSNO), add new
		5. Retrieve and re-validate (?) cardholders
		6. Retrieve badges
		7. Process badges - remove missing, update activate/deactivate as required, add new
		8. Retrieve and re-validate (?) badges
		9. Retrieve access level assignments
		10. Process access level assignments - remove missing, update as required, add new (sort by access level name)
	*/

	// 1. Retrieve status, type, and access levels from Lenel
	bsc := lenel.NewBadgeStatusCache()
	if err := bsc.Fill(cl); err != nil {
		log.Fatalf("Badge status cache fill failed: %v", err)
	}

	btc := lenel.NewBadgeTypeCache()
	if err := btc.Fill(cl); err != nil {
		log.Fatalf("Badge type cache fill failed: %v", err)
	}

	alc := lenel.NewAccessLevelCache()
	if err := alc.Fill(cl); err != nil {
		log.Fatalf("Access level cache fill failed: %v", err)
	}

	// 2. Parse CSV....
	csvCache, err := csv.Parse(cfg.File)
	if err != nil {
		log.Fatalf("Operation failed: %v", err)
	}

	// ... and ensure that all badge statuses, badge types, and access levels already exist in Lenel
	if err := bsc.Validate(csvCache.BadgeStatusNames()); err != nil {
		log.Fatalf("Fatal data mismatch: %v", err)
	}

	if err := btc.Validate(csvCache.BadgeTypeNames()); err != nil {
		log.Fatalf("Fatal data mismatch: %v", err)
	}

	if err := alc.Validate(csvCache.AccessLevelNames()); err != nil {
		log.Fatalf("Fatal data mismatch: %v", err)
	}
}

// ContentEquals returns true if two AccessRecords have identical content across all 14 fields.
func ContentEquals(a, b *model.AccessRecord, w io.Writer) bool {
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
	New      []*model.AccessRecord
	Existing []*model.AccessRecord
	Update   []*model.AccessRecord
	Delete   []*model.AccessRecord
	All      []*model.AccessRecord
}

// CompareRecords classifies each record in source against target:
// NEW if absent from target, EXISTING if content matches, UPDATE if different.
// If a record in target is not in source it is marked DELETE.
func CompareRecords(source, target []*model.AccessRecord, w io.Writer) CompareResult {
	targetByKey := make(map[string]*model.AccessRecord, len(target))
	for _, r := range target {
		targetByKey[r.GetKey()] = r
	}

	sourceByKey := make(map[string]struct{}, len(source))
	var result CompareResult

	for _, r := range source {
		sourceByKey[r.GetKey()] = struct{}{}

		if s, ok := targetByKey[r.GetKey()]; !ok {
			result.New = append(result.New, r)
		} else if ContentEquals(r, s, w) {
			result.Existing = append(result.Existing, r)
		} else {
			result.Update = append(result.Update, r)
		}
		result.All = append(result.All, r)
	}

	for _, r := range target {
		if _, ok := sourceByKey[r.GetKey()]; !ok {
			result.Delete = append(result.Delete, r)
			result.All = append(result.All, r)
		}
	}

	return result
}
