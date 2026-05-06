package main

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatalf("Error parsing command line arguments: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	client, err := NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close client session: %v", err)
		}
	}()

	cache := NewDataCache(client)
	if err := cache.Fill(); err != nil {
		log.Fatalf("Failed to load API data: %v", err)
	}

	switch cfg.Mode {
	case ModeExport:
		arc := BuildAccessRecordCache(cache)

		err = PrintCSVReport(arc.records, cfg.File)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}

	case ModeSync:
		arc := BuildAccessRecordCache(cache)

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
		result := CompareRecords(csvRecords, arc.records, diffWriter)

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
				log.Printf("status=%s ssno=%s badgeId=%s", r.SyncStatus.String(), r.SSNO, r.BadgeID)
			}
		}

	case ModeCleanup:
		log.Println("cleanup not yet implemented")

	case ModeFullExport:
		err = ExportXLSX(cache, cfg.File)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}
	}
}
