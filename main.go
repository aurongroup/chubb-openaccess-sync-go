package main

import (
	"errors"
	"log"
	"openaccess-sync/client"
	"openaccess-sync/config"
	"openaccess-sync/data/csv"
	"openaccess-sync/data/lenel"
	"openaccess-sync/data/model"
	"os"
	"strconv"

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
		items, err := cl.GetInstancesWithProgress("Lnl_BadgeStatus", "")
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}

		statusList := make([]*model.BadgeStatus, 0)
		statusListByID := make(map[int32]*model.BadgeStatus)
		statusListByName := make(map[string]*model.BadgeStatus)
		for _, props := range items {
			s, err := model.NewBadgeStatusFromJSON(props)
			if err != nil {
				log.Printf("skipping Lnl_BadgeStatus: %v", err)
				continue
			}

			statusList = append(statusList, s)
			statusListByID[s.ID] = s
			statusListByName[s.Name] = s
		}

		log.Printf("Retrieved %d Lnl_BadgeStatus records", len(statusList))

		items, err = cl.GetInstancesWithProgress("Lnl_BadgeType", "")
		if err != nil {
			log.Printf("skipping Lnl_BadgeType: %v", err)
		}

		typeList := make([]*model.BadgeType, 0)
		typeListByID := make(map[int32]*model.BadgeType)
		typeListByName := make(map[string]*model.BadgeType)
		for _, props := range items {
			t, err := model.NewBadgeTypeFromJSON(props)
			if err != nil {
				log.Printf("skipping Lnl_BadgeType: %v", err)
				continue
			}

			typeList = append(typeList, t)
			typeListByID[t.ID] = t
			typeListByName[t.Name] = t
		}

		log.Printf("Retrieved %d Lnl_BadgeType records", len(typeList))

		csvRecords, err := ParseCSV(cfg.File)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}

		csvBadges := make([]*model.Badge, 0, len(csvRecords))
		csvBadgesByID := make(map[int64]*model.Badge)
		for _, r := range csvRecords {
			id, err := strconv.ParseInt(r.BadgeID, 10, 64)
			if err != nil {
				continue
			}

			badgeStatus, ok := statusListByName[r.Status]
			if !ok {
				log.Printf("skipping CSV badge: unknown status %s", r.Status)
				continue
			}

			badgeType, ok := typeListByName[r.BadgeType]
			if !ok {
				log.Printf("skipping CSV badge: unknown type %s", r.BadgeType)
				continue
			}

			//log.Printf("Processing CSV badge: ID=%d, Status=%s, Type=%s", id, r.Status, r.BadgeType)

			b, err := model.NewBadge(
				int64(id),
				0,
				nil,
				nil,
				badgeStatus.ID,
				badgeType.ID,
				0,
			)
			if err != nil {
				log.Printf("skipping CSV badge: %v", err)
				continue
			}

			csvBadges = append(csvBadges, b)
			csvBadgesByID[b.ID] = b
		}

		log.Printf("Retrieved %d CSV Badge records", len(csvBadges))

		// Get badges
		badges, err := cl.GetInstancesWithProgress("Lnl_Badge", "")
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}

		badgeList := make([]*model.Badge, 0, len(badges))
		badgesByID := make(map[int64]*model.Badge)
		badgesByKey := make(map[int32]*model.Badge)

		for _, props := range badges {
			b, err := model.NewBadgeFromJSON(props, cache)
			if err != nil {
				log.Printf("skipping Lnl_Badge: %v", err)
				continue
			}

			badgeList = append(badgeList, b)
			badgesByID[b.ID] = b
			badgesByKey[b.Key] = b

			//log.Printf("Processing Lenel badge: ID=%d, Status=%s, Type=%s", b.ID, statusListByID[b.Status].Name, typeListByID[b.Type].Name)
		}

		log.Printf("Retrieved %d Lnl_Badge records", len(badgeList))

		for _, b := range badgeList {
			if _, ok := csvBadgesByID[b.ID]; !ok {
				log.Printf("deleting Lenel badge: ID=%d, Status=%s, Type=%s", b.ID, b.Status, b.Type)

				// Remove from cache
				continue
			} else {
				// Compare if the same, update if necessary
			}
		}

		// Remove all badges that are not in the CSV
		// Remove all cardholders that are not in the CSV
		// Update all cardholders with SSNOs that are in the CSV
		// Add all cardholders that are in the CSV but not in Lenel (squash duplicates)
		// Add all badges that are in the CSV but not in Lenel
		// Update all access level assignments for badges

		// 1. Pull data from Lenel
		// 2. Load data from CSV
		// 3. Update cardholders
		// - 3.1 Add cardholders to Lenel that are in CSV but not in Lenel
		// - 3.2 Update cardholders in Lenel using CSV data
		// - 3.3 Delete cardholders from Lenel that are not in CSV
		// 4. Update badges in Lenel using CSV data
		// - 7.1 Update activate/deactivate, type, status
		// - 7.2 Update access levels
		// - 7.3 Delete badges from Lenel that are not in CSV

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

	case config.ModeCleanup:
		log.Println("cleanup not yet implemented")

	case config.ModeFullExport:
		err = ExportXLSX(cache, cfg.File)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}
	}
}
