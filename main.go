package main

import (
	"errors"
	"io"
	"log"
	"openaccess-sync/client"
	"openaccess-sync/config"
	"openaccess-sync/data/csv"
	"openaccess-sync/data/lenel"
	"openaccess-sync/data/model"
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
	if err := cache.Fill(); err != nil {
		log.Fatalf("Failed to load API data: %v", err)
	}

	switch cfg.Mode {
	case config.ModeExport:
		arc := csv.BuildAccessRecordCache(cache)

		err = PrintCSVReport(arc.Records(), cfg.File)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}

	case config.ModeSync:
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
		}

		log.Printf("Retrieved %d Lnl_Badges records", len(badgeList))

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

		arc := csv.BuildAccessRecordCache(cache)

		csvRecords, err := ParseCSV(cfg.File)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}

		var diffWriter io.Writer
		if cfg.DiffFile != "" {
			f, err := os.Create(cfg.DiffFile)
			if err != nil {
				log.Fatalf("Failed to open diff file: %v", err)
			}
			defer f.Close()
			diffWriter = f
		}
		result := CompareRecords(csvRecords, arc.Records(), diffWriter)

		log.Printf(
			"Total records: %d, Existing %d, Update %d, Delete %d, New %d",
			len(result.All),
			len(result.Existing),
			len(result.Update),
			len(result.Delete),
			len(result.New),
		)

		if cfg.Verbose {
			for _, r := range result.All {
				//log.Printf("status=%s ssno=%s badgeId=%s", r.SyncStatus.String(), r.SSNO, r.BadgeID) // TODO
				log.Printf("ssno=%s badgeId=%s", r.SSNO, r.BadgeID)
			}
		}

	case config.ModeCleanup:
		log.Println("cleanup not yet implemented")

	case config.ModeFullExport:
		err = ExportXLSX(cache, cfg.File)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}
	}
}
