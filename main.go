package main

import (
	"errors"
	"io"
	"log"
	client2 "openaccess-sync/client"
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

	client, err := client2.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close client session: %v", err)
		}
	}()

	cache := lenel.NewDataCache(client)
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
			len(csvRecords),
			len(result.Existing),
			len(result.Update),
			len(result.New),
			len(result.Delete),
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
