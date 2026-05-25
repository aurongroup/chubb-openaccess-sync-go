package main

import (
	"errors"
	"log"
	"openaccess-sync/client"
	"openaccess-sync/config"
	"openaccess-sync/data/csv"
	"openaccess-sync/data/lenel"
	"os"

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

	cache := lenel.NewDataCache(cl)

	//if err := cache.Fill(); err != nil {
	//	log.Fatalf("Failed to load API data: %v", err)
	//}

	switch cfg.Mode {
	case config.ModeExport:
		arc := csv.BuildAccessRecordCache(cache)

		err = PrintCSVReport(arc.Records(), cfg.File)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}

	case config.ModeSync:
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
		csvCache, err := ParseCSV(cfg.File)
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

	case config.ModeCleanup:
		log.Println("cleanup not yet implemented")
		break

		// 3. Cleanup Cardholders that don't have badges
		chc := lenel.NewCardholderCache()
		if err := chc.FillDetached(cl); err != nil {
			log.Fatalf("Detached Cardholder cache fill failed: %v", err)
		}

		for _, ch := range chc.GetItems() {
			log.Printf("Deleting detached cardholder with ID %d", ch.ID)
			err := cl.DeleteInstance("Lnl_Cardholder", map[string]interface{}{"ID": ch.ID})
			if err != nil {
				log.Printf("-> failed to delete cardholder with ID %d: %v", ch.ID, err)
			}
		}

		// 4. Remove Cardholders that don't have an SSNO and are duplicates.
		//    WARNING - this is a DESTRUCTIVE action because we are unable to transfer badges
		//              from duplicate cardholders
		chc = lenel.NewCardholderCache()
		if err := chc.FillNoSSNO(cl); err != nil {
			log.Fatalf("No SSNO Cardholder cache fill failed: %v", err)
		}

		// Cardholders can only have a maximum of two active badges......

		duplicates := chc.GetDuplicates()
		for k, chs := range duplicates {
			log.Printf("Processing duplicates for %s (%d)", k, chs[0].ID)
			for _, ch := range chs[1:] {
				// Grab assigned badges for duplicate
				bc := lenel.NewBadgeCache()
				if err := bc.FillActiveForCardholder(cl, ch); err != nil {
					log.Fatalf("Badge cache fill for Cardholder %d failed: %v", ch.ID, err)
				}

				for _, badge := range bc.GetItems() {
					log.Printf("-> Processing Badge %d for duplicate Cardholder %d", badge.ID, ch.ID)

					ac := lenel.NewAssignmentCache()
					if err := ac.FillForBadge(cl, badge); err != nil {
						log.Fatalf("Assignment cache fill for Badge %d (%d) failed: %v", badge.ID, badge.Key, err)
					}

					for _, assignment := range ac.GetItems() {
						log.Printf("--> Processing Assignment %d for Badge %d (%d)", assignment.AccessLevel, badge.ID, badge.Key)
					}

					// Delete the original badge
					// Create a new badge
					// Create assignments
					//cl.CreateInstance()
				}
				// For each badge, grab access levels
				// Delete badge
				// Create new badge against primary
				// Add access levels to new badge
				// Delete duplicate

				log.Printf("-> Deleting duplicate cardholder with ID %d", ch.ID)
				//err := cl.DeleteInstance("Lnl_Cardholder", map[string]interface{}{"ID": ch.ID})
				//if err != nil {
				//	log.Printf("-> failed to delete cardholder with ID %d: %v", ch.ID, err)
				//}
			}
		}

		// 3. Retrieve Cardholders from Lenel that have a NULL SSNO so we can clean up duplicates
		//chc := lenel.NewCardholderCache()
		//if err := chc.FillSSNO(cl); err != nil {
		//	log.Fatalf("Cardholder cache fill failed: %v", err)
		//}

		//if err := cache.ValidateAccessRecords(csvCache.Records()); err != nil {
		//	log.Fatalf("CSV validation failed: %v", err)
		//}

		// -- SNIP --
		//csvBadges := make([]*model.Badge, 0, len(csvRecords))
		//csvBadgesByID := make(map[int64]*model.Badge)
		//for _, r := range csvRecords {
		//	id, err := strconv.ParseInt(r.BadgeID, 10, 64)
		//	if err != nil {
		//		continue
		//	}
		//
		//	badgeStatus, ok := cache.GetBadgeStatusByName(r.Status)
		//	if !ok {
		//		log.Printf("skipping CSV badge: unknown status %s", r.Status)
		//		continue
		//	}
		//
		//	badgeType, ok := cache.GetBadgeTypeByName(r.BadgeType)
		//	if !ok {
		//		log.Printf("skipping CSV badge: unknown type %s", r.BadgeType)
		//		continue
		//	}
		//
		//	b, err := model.NewBadge(
		//		int64(id),
		//		0,
		//		nil,
		//		nil,
		//		badgeStatus.ID,
		//		badgeType.ID,
		//		0,
		//	)
		//	if err != nil {
		//		log.Printf("skipping CSV badge: %v", err)
		//		continue
		//	}
		//
		//	csvBadges = append(csvBadges, b)
		//	csvBadgesByID[b.ID] = b
		//}
		//
		//log.Printf("Retrieved %d CSV Badge records", len(csvBadges))
		//
		//// Get badges
		//badges, err := cl.GetInstancesWithProgress("Lnl_Badge", "")
		//if err != nil {
		//	log.Fatalf("Operation failed: %v", err)
		//}
		//
		//badgeList := make([]*model.Badge, 0, len(badges))
		//badgesByID := make(map[int64]*model.Badge)
		//badgesByKey := make(map[int32]*model.Badge)
		//
		//for _, props := range badges {
		//	b, err := model.NewBadgeFromJSON(props, cache)
		//	if err != nil {
		//		log.Printf("skipping Lnl_Badge: %v", err)
		//		continue
		//	}
		//
		//	badgeList = append(badgeList, b)
		//	badgesByID[b.ID] = b
		//	badgesByKey[b.Key] = b
		//
		//	//log.Printf("Processing Lenel badge: ID=%d, Status=%s, Type=%s", b.ID, statusListByID[b.Status].Name, typeListByID[b.Type].Name)
		//}
		//
		//log.Printf("Retrieved %d Lnl_Badge records", len(badgeList))
		//
		//for _, b := range badgeList {
		//	if _, ok := csvBadgesByID[b.ID]; !ok {
		//		log.Printf("deleting Lenel badge: ID=%d, Status=%s, Type=%s", b.ID, b.Status, b.Type)
		//
		//		// Remove from cache
		//		continue
		//	} else {
		//		// Compare if the same, update if necessary
		//	}
		//}
		// -- SNIP --

		//arc := csv.BuildAccessRecordCache(cache)

		//csvRecords, err := ParseCSV(cfg.File)
		//if err != nil {
		//	log.Fatalf("Operation failed: %v", err)
		//}

		//var diffWriter io.Writer
		//if cfg.DiffFile != "" {
		//	f, err := os.Create(cfg.DiffFile)
		//	if err != nil {
		//		log.Fatalf("Failed to open diff file: %v", err)
		//	}
		//	defer f.Close()
		//	diffWriter = f
		//}
		//result := CompareRecords(csvRecords, arc.Records(), diffWriter)

		//log.Printf(
		//	"Total records: %d, Existing %d, Update %d, Delete %d, New %d",
		//	len(result.All),
		//	len(result.Existing),
		//	len(result.Update),
		//	len(result.Delete),
		//	len(result.New),
		//)
		//
		//if cfg.Verbose {
		//	for _, r := range result.All {
		//		//log.Printf("status=%s ssno=%s badgeId=%s", r.SyncStatus.String(), r.SSNO, r.BadgeID) // TODO
		//		log.Printf("ssno=%s badgeId=%s", r.SSNO, r.BadgeID)
		//	}
		//}

	case config.ModeFullExport:
		err = ExportXLSX(cache, cfg.File)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}
	}
}
